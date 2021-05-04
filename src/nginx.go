//go:generate goversioninfo
package main

import (
	"fmt"
	"runtime"
	"strings"

	"net/url"
	"os"

	sdk_args "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/data/attribute"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/pkg/errors"
)

type argumentList struct {
	sdk_args.DefaultArgumentList
	StatusURL         string `default:"http://127.0.0.1/status" help:"NGINX status URL. If you are using ngx_http_api_module be sure to include the full path ending with the API version number"`
	ConfigPath        string `default:"/etc/nginx/nginx.conf" help:"NGINX configuration file."`
	RemoteMonitoring  bool   `default:"false" help:"Identifies the monitored entity as 'remote'. In doubt: set to true."`
	ConnectionTimeout int    `default:"5" help:"Connection timeout to the Nginx instance in seconds"`
	StatusModule      string `default:"discover" help:"Name of Nginx status module. discover | ngx_http_stub_status_module | ngx_http_status_module | ngx_http_api_module"`
	ValidateCerts     bool   `default:"true" help:"If the status URL is HTTPS with a self-signed certificate, set this to false if you want to avoid certificate validation"`
	ShowVersion       bool   `default:"false" help:"Print build information and exit"`
}

const (
	integrationName = "com.newrelic.nginx"

	entityRemoteType = "server"

	httpsProtocol    = `https`
	httpProtocol     = `http`
	httpDefaultPort  = `80`
	httpsDefaultPort = `443`

	httpStubStatus = "ngx_http_stub_status_module"
	httpStatus     = "ngx_http_status_module"
	httpAPIStatus  = "ngx_http_api_module"
)

var (
	args               argumentList
	integrationVersion = "0.0.0"
	gitCommit          = ""
	buildDate          = ""
)

func main() {
	i, err := createIntegration()
	fatalIfErr(err)

	if args.ShowVersion {
		fmt.Printf(
			"New Relic %s integration Version: %s, Platform: %s, GoVersion: %s, GitCommit: %s, BuildDate: %s\n",
			strings.Title(strings.Replace(integrationName, "com.newrelic.", "", 1)),
			integrationVersion,
			fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			runtime.Version(),
			gitCommit,
			buildDate)
		os.Exit(0)
	}

	e, err := entity(i)
	fatalIfErr(err)

	if args.HasInventory() {
		fatalIfErr(setInventoryData(e.Inventory))
	}

	if args.HasMetrics() {
		ms := metricSet(e, "NginxSample", args.RemoteMonitoring)
		err = getMetricsData(ms)
		fatalIfErr(err)
	}

	fatalIfErr(i.Publish())
}

func entity(i *integration.Integration) (*integration.Entity, error) {
	if args.RemoteMonitoring {
		hostname, port, err := parseStatusURL(args.StatusURL)
		if err != nil {
			return nil, err
		}
		n := fmt.Sprintf("%s:%s", hostname, port)
		return i.Entity(n, entityRemoteType)
	}

	return i.LocalEntity(), nil
}

func metricSet(e *integration.Entity, eventType string, remote bool) *metric.Set {
	hostname, port, err := parseStatusURL(args.StatusURL)
	fatalIfErr(err)
	if remote {
		return e.NewMetricSet(
			eventType,
			attribute.Attr("hostname", hostname),
			attribute.Attr("port", port),
		)
	}

	return e.NewMetricSet(
		eventType,
		attribute.Attr("port", port),
	)
}

func createIntegration() (*integration.Integration, error) {
	cachePath := os.Getenv("NRIA_CACHE_PATH")
	if cachePath == "" {
		return integration.New(integrationName, integrationVersion, integration.Args(&args))
	}

	l := log.NewStdErr(args.Verbose)
	s, err := persist.NewFileStore(cachePath, l, persist.DefaultTTL)
	if err != nil {
		return nil, err
	}

	return integration.New(integrationName, integrationVersion, integration.Args(&args), integration.Storer(s), integration.Logger(l))
}

// parseStatusURL will extract the hostname and the port from the nginx status URL.
func parseStatusURL(statusURL string) (hostname, port string, err error) {
	u, err := url.Parse(statusURL)
	if err != nil {
		return
	}

	if !isHTTP(u) {
		err = errors.New("unsupported protocol scheme")
		return
	}

	hostname = u.Hostname()
	if hostname == "" {
		err = errors.New("http: no Host in request URL")
		return
	}

	if u.Port() != "" {
		port = u.Port()
	} else if u.Scheme == httpsProtocol {
		port = httpsDefaultPort
	} else {
		port = httpDefaultPort
	}
	return
}

// isHTTP is checking if the URL is http/s protocol.
func isHTTP(u *url.URL) bool {
	return u.Scheme == httpProtocol || u.Scheme == httpsProtocol
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
