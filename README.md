# New Relic Infrastructure Integration for NGINX
New Relic Infrastructure Integration for NGINX captures critical performance metrics and inventory reported by NGINX server. There is an open source and a commercial version of NGINX, both supported by this integration.

Inventory data is obtained from the configuration files and metrics from the status modules.

See our [documentation web site](https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/nginx-monitoring-integration) for more details.

<!---
See [metrics]() or [inventory]() for more details about collected data and review [dashboard]() in order to know how the data is presented.
--->

## Configuration
* Depending on which NGINX edition you use please update your configuration enabling
  * [ngx_http_stub_status_module](http://nginx.org/en/docs/http/ngx_http_stub_status_module.html) for NGINX Open Source
  * [ngx_http_status_module](http://nginx.org/en/docs/http/ngx_http_status_module.html) for NGINX Plus pre 1.13.3
  * [ngx_http_api_module](http://nginx.org/en/docs/http/ngx_http_api_module.html) for NGINX 1.13.3 and higher
  
### ngx_http_api_module configuration
Additions to the `nginx-config.yaml` file:
* status_url: full path to the status url _including_ the api version. For example `http://localhost/api/4`
* status_module: `ngx_http_api_module`
* endpoints: list of `api/4`, NON PARAMETERIZED, endpoints to query
  * Default:"/nginx,/processes,/connections,/ssl,/slabs,/http,/http/requests,/http/server_zones,/http/caches,/http/upstreams,/http/keyvals,/stream,/stream/server_zones,/stream/upstreams,/stream/keyvals,/stream/zone_sync" 

## Installation
* download an archive file for the NGINX Integration
* extract `nginx-definition.yml` and `/bin` directory into `/var/db/newrelic-infra/newrelic-integrations`
* add execute permissions for the binary file `nri-nginx` (if required)
* extract `nginx-config.yml.sample` into `/etc/newrelic-infra/integrations.d`

## Usage
This is the description about how to run the NGINX Integration with New Relic Infrastructure agent, so it is required to have the agent installed (see [agent installation](https://docs.newrelic.com/docs/infrastructure/new-relic-infrastructure/installation/install-infrastructure-linux)).

In order to use the NGINX Integration it is required to configure `nginx-config.yml.sample` file. Firstly, rename the file to `nginx-config.yml`. Then, depending on your needs, specify all instances that you want to monitor. Once this is done, restart the Infrastructure agent.

You can view your data in Insights by creating your own custom NRQL queries. To do so use the **NginxSample** event type.

## Integration development usage
Assuming that you have source code you can build and run the NGINX Integration locally.

* Go to directory of the NGINX Integration and build it
```bash
$ make
```
* The command above will execute tests for the NGINX Integration and build an executable file called `nri-nginx` in `bin` directory.
```bash
$ ./bin/nri-nginx
```
* If you want to know more about usage of `./nri-nginx` check
```bash
$ ./bin/nri-nginx -help
```

For managing external dependencies [govendor tool](https://github.com/kardianos/govendor) is used. It is required to lock all external dependencies to specific version (if possible) into vendor directory.

## Contributing Code

We welcome code contributions (in the form of pull requests) from our user
community. Before submitting a pull request please review [these guidelines](https://github.com/newrelic/nri-nginx/blob/master/CONTRIBUTING.md).

Following these helps us efficiently review and incorporate your contribution
and avoid breaking your code with future changes to the agent.

## Custom Integrations

To extend your monitoring solution with custom metrics, we offer the Integrations
Golang SDK which can be found on [github](https://github.com/newrelic/infra-integrations-sdk).

Refer to [our docs site](https://docs.newrelic.com/docs/infrastructure/integrations-sdk/get-started/intro-infrastructure-integrations-sdk)
to get help on how to build your custom integrations.

## Support

You can find more detailed documentation [on our website](http://newrelic.com/docs),
and specifically in the [Infrastructure category](https://docs.newrelic.com/docs/infrastructure).

If you can't find what you're looking for there, reach out to us on our [support
site](http://support.newrelic.com/) or our [community forum](http://forum.newrelic.com)
and we'll be happy to help you.

Find a bug? Contact us via [support.newrelic.com](http://support.newrelic.com/),
or email support@newrelic.com.

New Relic, Inc.
