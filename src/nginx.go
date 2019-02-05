package main

import (
	sdk_args "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
)

type argumentList struct {
	sdk_args.DefaultArgumentList
	StatusURL  string `default:"http://127.0.0.1/status" help:"NGINX status URL."`
	ConfigPath string `default:"/etc/nginx/nginx.conf" help:"NGINX configuration file."`
}

const (
	integrationName    = "com.newrelic.nginx"
	integrationVersion = "1.0.2"
)

var (
	args argumentList
)

func main() {
	i, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	fatalIfErr(err)
	log.SetupLogging(args.Verbose)

	e := i.LocalEntity()

	if args.HasInventory() {
		fatalIfErr(setInventoryData(e.Inventory))
	}

	if args.HasMetrics() {
		ms := e.NewMetricSet("NginxSample", metric.Attr("status_url", args.StatusURL))
		fatalIfErr(getMetricsData(ms))
	}

	fatalIfErr(i.Publish())
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
