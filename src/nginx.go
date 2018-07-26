package main

import (
	sdk_args "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/sdk"
)

type argumentList struct {
	sdk_args.DefaultArgumentList
	StatusURL  string `default:"http://127.0.0.1/status" help:"NGINX status URL."`
	ConfigPath string `default:"/etc/nginx/nginx.conf" help:"NGINX configuration file."`
}

const (
	integrationName    = "com.newrelic.nginx"
	integrationVersion = "1.0.0"
)

var (
	args argumentList
)

func main() {
	integration, err := sdk.NewIntegration(integrationName, integrationVersion, &args)
	fatalIfErr(err)
	log.SetupLogging(args.Verbose)

	if args.All || args.Inventory {
		fatalIfErr(setInventoryData(integration.Inventory))
	}

	if args.All || args.Metrics {
		sample := integration.NewMetricSet("NginxSample")
		fatalIfErr(getMetricsData(sample))
	}

	fatalIfErr(integration.Publish())
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
