package main

import (
	"fmt"
	sdk_args "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/pkg/errors"
	"net/url"
	"os"
)

type argumentList struct {
	sdk_args.DefaultArgumentList
	StatusURL        string `default:"http://127.0.0.1/status" help:"NGINX status URL."`
	ConfigPath       string `default:"/etc/nginx/nginx.conf" help:"NGINX configuration file."`
	RemoteMonitoring bool   `default:"false" help:"Identifies the monitored entity as 'remote'. In doubt: set to true."`
}

const (
	integrationName    = "com.newrelic.nginx"
	integrationVersion = "1.1.0"

	entityRemoteType = "server"

	httpsProtocol    = `https`
	httpProtocol     = `http`
	httpDefaultPort  = `80`
	httpsDefaultPort = `443`
)

var (
	args argumentList
)

func main() {
	i, err := createIntegration()
	fatalIfErr(err)

	e, err := entity(i)
	fatalIfErr(err)

	if args.HasInventory() {
		fatalIfErr(setInventoryData(e.Inventory))
	}

	if args.HasMetrics() {
		hostname, port, err := parseStatusURL(args.StatusURL)
		fatalIfErr(err)

		ms := metricSet(e, "NginxSample", hostname, port)
		fatalIfErr(getMetricsData(ms))
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

func metricSet(e *integration.Entity, eventType, hostname, port string) *metric.Set {
	if args.RemoteMonitoring {
		return e.NewMetricSet(
			eventType,
			metric.Attr("hostname", hostname),
			metric.Attr("port", port),
		)
	}

	return e.NewMetricSet(
		eventType,
		metric.Attr("port", port),
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
