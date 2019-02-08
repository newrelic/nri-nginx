package main

import (
	sdk_args "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"os"
)

type argumentList struct {
	sdk_args.DefaultArgumentList
	StatusURL  string `default:"http://127.0.0.1/status" help:"NGINX status URL."`
	ConfigPath string `default:"/etc/nginx/nginx.conf" help:"NGINX configuration file."`
}

const (
	integrationName    = "com.newrelic.nginx"
	integrationVersion = "1.1.0"
)

var (
	args argumentList
)

func main() {
	var i *integration.Integration
	var err error
	cachePath := os.Getenv("NRIA_CACHE_PATH")

	if cachePath == "" {
		i, err = integration.New(integrationName, integrationVersion, integration.Args(&args))
	} else {
		var storer persist.Storer

		logger := log.NewStdErr(args.Verbose)
		storer, err = persist.NewFileStore(cachePath, logger, persist.DefaultTTL)
		fatalIfErr(err)

		i, err = integration.New(integrationName, integrationVersion, integration.Args(&args),
			integration.Storer(storer), integration.Logger(logger))
	}

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
