package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jeremywohl/flatten"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/log"
)

var metricsStandardDefinition = map[string][]interface{}{
	"software.edition":                 {"edition", metric.ATTRIBUTE},
	"software.version":                 {"version", metric.ATTRIBUTE},
	"net.connectionsActive":            {"active", metric.GAUGE},
	"net.connectionsAcceptedPerSecond": {"accepted", metric.RATE},
	"net.connectionsDroppedPerSecond":  {connectionsDropped, metric.RATE},
	"net.connectionsReading":           {"reading", metric.GAUGE},
	"net.connectionsWaiting":           {"waiting", metric.GAUGE},
	"net.connectionsWriting":           {"writing", metric.GAUGE},
	"net.requestsPerSecond":            {"requests", metric.RATE},
}

var metricsPlusDefinition = map[string][]interface{}{
	"software.edition":                 {"edition", metric.ATTRIBUTE},
	"software.version":                 {"version", metric.ATTRIBUTE},
	"net.connectionsActive":            {"connections.active", metric.GAUGE},
	"net.connectionsIdle":              {"connections.idle", metric.GAUGE},
	"net.connectionsAcceptedPerSecond": {"connections.accepted", metric.RATE},
	"net.connectionsDroppedPerSecond":  {"connections.dropped", metric.RATE},
	"net.requestsPerSecond":            {"requests.total", metric.RATE},
	"processes.respawned":              {"processes.respawned", metric.DELTA},
	"ssl.handshakes":                   {"ssl.handshakes", metric.DELTA},
	"ssl.failedHandshakes":             {"ssl.handshakes_failed", metric.DELTA},
	"ssl.sessionReuses":                {"ssl.session_reuses", metric.DELTA},
}

var metricsPlusAPIDefinition = map[string][]interface{}{
	"software.version":      {"nginx.version", metric.ATTRIBUTE},
	"connections.active":    {"net.connectionsActive", metric.GAUGE},
	"connections.idle":      {"net.connectionsIdle", metric.GAUGE},
	"connections.accepted":  {"net.connectionsAcceptedPerSecond", metric.RATE},
	"connections.dropped":   {"net.connectionsDroppedPerSecond", metric.RATE},
	"processes.respawned":   {"processes.respawned", metric.DELTA},
	"ssl.handshakes":        {"ssl.handshakes", metric.DELTA},
	"ssl.handshakes_failed": {"ssl.failedHandshakes", metric.DELTA},
	"ssl.session_reuses":    {"ssl.sessionReuses", metric.DELTA},
	"http.requests.total":   {"net.requestsPerSecond", metric.RATE},
	"http.requests.current": {"net.requests", metric.GAUGE},
}

// expressions contains the structure of the input data and defines the attributes we want to store
var nginxStatusExpressions = []*regexp.Regexp{
	regexp.MustCompile(`Active connections:\s+(?P<active>\d+)`),
	nil,
	regexp.MustCompile(`\s*(?P<accepted>\d+)\s+(?P<handled>\d+)\s+(?P<requests>\d+)`),
	regexp.MustCompile(`Reading: (?P<reading>\d+)\s+Writing: (?P<writing>\d+)\s+Waiting: (?P<waiting>\d+)`),
}

func connectionsDropped(metrics map[string]interface{}) (int, bool) {
	accepts, ok1 := metrics["accepted"].(int)
	handled, ok2 := metrics["handled"].(int)

	if ok1 && ok2 {
		return accepts - handled, true
	}
	return 0, false
}

// getMetrics reads an NGINX (open edition) status message and transforms its
// contents into a map that can be processed by NR agent.
// It returns a map of metrics with all the keys and values extracted from the
// status endpoint.
func getStandardMetrics(reader *bufio.Reader) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})
	for lineNo, re := range nginxStatusExpressions {
		line, err := reader.ReadString('\n')

		if err == io.EOF {
			return metrics, nil
		}

		if re == nil {
			continue
		}

		match := re.FindStringSubmatch(line)
		if match == nil {
			return nil, fmt.Errorf("Line %d of status doesn't match", lineNo)
		}

		for i, name := range re.SubexpNames() {
			if i != 0 {
				value, err := strconv.Atoi(match[i])
				if err != nil {
					log.Warn("Can't cast value '%s'", match[i])
					continue
				}
				metrics[name] = value
			}
		}
	}
	metrics["version"] = ""
	metrics["edition"] = "open source"

	return metrics, nil
}

// getPlusMetrics reads an NGINX (Plus edition) status message, gets some
// metrics and transforms the contents into a map that can be processed by NR
// agent.
// It returns a map of metrics keys -> values.
func getPlusMetrics(reader *bufio.Reader) (map[string]interface{}, error) {
	jsonMetrics := make(map[string]interface{})
	metrics := make(map[string]interface{})

	dec := json.NewDecoder(reader)
	err := dec.Decode(&jsonMetrics)
	if err != nil {
		return nil, err
	}

	roots := []string{"connections", "requests", "ssl", "processes"}

	for _, rootKey := range roots {
		rootNode, ok := jsonMetrics[rootKey].(map[string]interface{})
		if !ok {
			log.Warn("Can't assert type for %s", rootNode)
			continue
		}
		for key, value := range rootNode {
			metrics[fmt.Sprintf("%s.%s", rootKey, key)] = int(value.(float64))
		}
	}
	metrics["version"] = jsonMetrics["nginx_version"]
	metrics["edition"] = "plus"
	return metrics, nil
}

func populateMetrics(sample *metric.Set, metrics map[string]interface{}, metricsDefinition map[string][]interface{}) error {
	for metricName, metricInfo := range metricsDefinition {
		rawSource := metricInfo[0]
		metricType := metricInfo[1].(metric.SourceType)

		var rawMetric interface{}
		var ok bool

		switch source := rawSource.(type) {
		case string:
			rawMetric, ok = metrics[source]
		case func(map[string]interface{}) (int, bool):
			rawMetric, ok = source(metrics)
		default:
			log.Warn("Invalid raw source metric for %s", metricName)
			continue
		}

		if !ok {
			log.Warn("Can't find raw metrics in results for %s", metricName)
			continue
		}
		err := sample.SetMetric(metricName, rawMetric, metricType)

		if err != nil {
			log.Warn("Error setting value: %s", err)
			continue
		}
	}
	return nil
}

func getMetricsData(sample *metric.Set) error {
	switch args.StatusModule {
	case httpStubStatus:
		resp, err := getStatus("")
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		metricsDefinition := metricsStandardDefinition
		rawMetrics, err := getStandardMetrics(bufio.NewReader(resp.Body))
		if err != nil {
			return err
		}

		rawVersion := strings.Replace(resp.Header.Get("Server"), "nginx/", "", -1)
		rawMetrics["version"] = rawVersion
		return populateMetrics(sample, rawMetrics, metricsDefinition)
	case httpStatus:
		resp, err := getStatus("")
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		metricsDefinition := metricsPlusDefinition
		rawMetrics, err := getPlusMetrics(bufio.NewReader(resp.Body))
		if err != nil {
			return err
		}
		return populateMetrics(sample, rawMetrics, metricsDefinition)
	case httpAPIStatus:
		return pollHttpAPIStatusEndpoints(sample)
	default:
		return getDiscoveredMetricsData(sample)
	}
}

func pollHttpAPIStatusEndpoints(sample *metric.Set) error {
	paths := []string{"/nginx", "/processes", "/connections", "/http/requests", "/ssl"}
	for _, p := range paths {
		resp, err := getStatus(p)
		if err != nil {
			log.Warn("Request to endpoint failed: %s", err)
			continue
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Warn("Unable to close response body: %s", err)
			}
		}()
		getHTTPAPIMetrics(p, sample, bufio.NewReader(resp.Body))
	}
	return nil
}

func getHTTPAPIMetrics(path string, sample *metric.Set, reader *bufio.Reader) {
	jsonMetrics := make(map[string]interface{})
	dec := json.NewDecoder(reader)
	err := dec.Decode(&jsonMetrics)
	if err != nil {
		return
	}
	if jsonMetrics == nil || len(jsonMetrics) <= 0 {
		return
	}

	flat, err := flatten.Flatten(jsonMetrics, "", flatten.DotStyle)
	if err != nil {
		log.Error("Error flattening json: %+v", err)
		return
	}

	for k, v := range flat {
		key := pathToPrefix(path) + k
		realKey, typ := getAttributeType(key, v)

		if err := sample.SetMetric(realKey, v, typ); err != nil {
			log.Error("Unable to set metric: %s", err)
		}
	}
	if err := sample.SetMetric("software.edition", "plus", metric.ATTRIBUTE); err != nil {
		log.Error("Unable to set metric: %s", err)
	}
}

var notJustDots = regexp.MustCompile(`[^.]`)

func pathToPrefix(path string) (prefix string) {
	prefix = strings.TrimPrefix(path, "/")
	prefix = strings.Replace(prefix, "/", ".", -1)
	if !strings.HasSuffix(prefix, ".") {
		prefix = prefix + "."
	}
	if prefix == "." {
		prefix = ""
	}
	if !notJustDots.MatchString(prefix) {
		prefix = ""
	}
	return prefix
}

func getAttributeType(key string, v interface{}) (string, metric.SourceType) {
	if md, ok := metricsPlusAPIDefinition[key]; ok {
		return md[0].(string), md[1].(metric.SourceType)
	}

	// if nothing else matches infer type from the value
	switch v.(type) {
	case string:
		return key, metric.ATTRIBUTE
	default:
		return key, metric.GAUGE
	}
}

func httpClient() *http.Client {
	netClient := http.Client{
		Timeout: time.Duration(args.ConnectionTimeout) * time.Second,
	}
	if !args.ValidateCerts {
		netClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	return &netClient
}

func getStatus(path string) (resp *http.Response, err error) {
	netClient := httpClient()
	resp, err = netClient.Get(args.StatusURL + path)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("failed to get stats from %s. Server returned code %d (%s). Expecting 200", args.StatusURL+path, resp.StatusCode, resp.Status)
	}
	return
}

// For backwards compatibility, the integration tries to discover whether the metrics are standard or nginx plus based
// on their format
func getDiscoveredMetricsData(sample *metric.Set) error {
	netClient := httpClient()
	resp, err := netClient.Get(args.StatusURL)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get stats from nginx. Server returned code %d (%s). Expecting 200",
			resp.StatusCode, resp.Status)
	}
	defer resp.Body.Close()
	var rawMetrics map[string]interface{}
	var metricsDefinition map[string][]interface{}

	if resp.Header.Get("content-type") == "application/json" {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if strings.Contains(string(bodyBytes), nginxPlusApiRootNginxEndpoint) {
			return pollHttpAPIStatusEndpoints(sample)
		}
		metricsDefinition = metricsPlusDefinition
		rawMetrics, err = getPlusMetrics(bufio.NewReader(bytes.NewBuffer(bodyBytes)))
		if err != nil {
			return err
		}
	} else {
		metricsDefinition = metricsStandardDefinition
		rawMetrics, err = getStandardMetrics(bufio.NewReader(resp.Body))
		if err != nil {
			return err
		}
		rawVersion := strings.Replace(resp.Header.Get("Server"), "nginx/", "", -1)
		rawMetrics["version"] = rawVersion
	}
	return populateMetrics(sample, rawMetrics, metricsDefinition)
}
