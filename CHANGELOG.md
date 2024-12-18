# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

Unreleased section should follow [Release Toolkit](https://github.com/newrelic/release-toolkit#render-markdown-and-update-markdown)

## Unreleased

## v3.5.1 - 2024-12-18

### ⛓️ Dependencies
- Updated golang patch version to v1.23.4

## v3.5.0 - 2024-10-09

### dependency
- Upgrade go to 1.23.2

### 🚀 Enhancements
- Upgrade integrations SDK so the interval is variable and allows intervals up to 5 minutes

## v3.4.10 - 2024-09-18

### ⛓️ Dependencies
- Updated golang.org/x/tools to v0.25.0

## v3.4.9 - 2024-09-11

### ⛓️ Dependencies
- Updated golang version to v1.23.1

## v3.4.8 - 2024-08-14

### ⛓️ Dependencies
- Updated golang version to v1.22.6

## v3.4.7 - 2024-07-10

### ⛓️ Dependencies
- Updated golang version to v1.22.5

## v3.4.6 - 2024-05-08

### ⛓️ Dependencies
- Updated golang.org/x/tools to v0.21.0
- Updated golang version to v1.22.3

## v3.4.5 - 2024-04-10

### ⛓️ Dependencies
- Updated golang version to v1.22.2

## v3.4.4 - 2024-03-13

### 🐞 Bug fixes
- Updated golang to version v1.21.7 to fix a vulnerability

## v3.4.3 - 2024-02-21

### ⛓️ Dependencies
- Updated github.com/newrelic/infra-integrations-sdk to v3.8.2+incompatible

## v3.4.2 - 2023-10-20

### ⛓️ Dependencies
- Updated golang.org/x/tools to v0.14.0
- Updated golang version to 1.21

## v3.4.2 - 2023-10-18

### ⛓️ Dependencies
- Updated golang.org/x/tools to v0.14.0
- Updated golang version to 1.21

## v3.4.2 - 2023-10-11

### ⛓️ Dependencies
- Updated golang.org/x/tools to v0.14.0
- Updated golang version to 1.21

## v3.4.1 - 2023-08-02

### ⛓️ Dependencies
- Updated golang.org/x/tools to v0.11.1
- Updated golang to v1.20.7

## v3.4.0 - 2023-07-19

### 🚀 Enhancements
- bumped golang version pinning 1.20.6

### ⛓️ Dependencies
- Updated golang.org/x/tools to v0.11.0

## 3.3.0 (2023-06-06)
### Changed
- Update Go version to 1.20

## 3.2.5 (2022-10-18)
## Fixed
- Handle properly the error while parsing the `nginx.conf` file and ignoring the comments #96.

## 3.2.4 (2022-06-21)
### Changed
- Bump dependencies
### Added
Added support for more distributions:
- RHEL(EL) 9
- Ubuntu 22.04

## 3.2.3 (2022-04-21)
### Added
- Config logs examples.
## Changed
- All DELTA and RATE metrics has been changed to PDELTA and PRATE preventing the integration to report negative values whenever the counters are reset. (#90)
- Use Go 1.18 (#89)
- Bump dependencies (#89)

## 3.2.2 (2021-10-20)
### Added
Added support for more distributions:
- Debian 11
- Ubuntu 20.10
- Ubuntu 21.04
- SUSE 12.15
- SUSE 15.1
- SUSE 15.2
- SUSE 15.3
- Oracle Linux 7
- Oracle Linux 8

## 3.2.1 (2021-10-20)
## Changed
Moved default config.sample to [V4](https://docs.newrelic.com/docs/create-integrations/infrastructure-integrations-sdk/specifications/host-integrations-newer-configuration-format/), added a dependency for infra-agent version 1.20.0 https://github.com/newrelic/nri-nginx/pull/83
Please notice that old [V3](https://docs.newrelic.com/docs/create-integrations/infrastructure-integrations-sdk/specifications/host-integrations-standard-configuration-format/) configuration format is deprecated, but still supported.

## 3.1.2 (2021-06-07)
## Changed
- Support for ARM

## 3.1.1 (2021-05-27)
## Changed
- Fixed a bug that preventing the integration from running when the nginx file being parsed included an empty line at the end of the file #80 (#81)

## 3.1.0 (2021-04-30)
## Changed
- Update Go to v1.16.
- Migrate to Go Modules
- Update Infrastracture SDK to v3.6.7.
- Update other dependecies.

## 3.0.1 (2020-08-03)
## Fixed
- Updated the configuration sample to exclude endpoints.
- Take integrationVersion var from the ldflags

## 3.0.0 (2020-07-29)
## Fixed
- Fixed metric types for NGINx Plus metrics.
### Changed
- Removed support for dynamic endpoint metrics. This will be addressed in a future release.

## 2.0.1 (2020-06-12)
## Fixed
- Updated the configuration sample to include the status_url for inventory required
  for entity naming.

## 2.0.0 (2019-01-28)
## Fixed
- Nginx Plus metrics were not being renamed like Nginx standard metrics
## Changed
- Major version change as the fix above breaks compatibility by renaming metrics

## 1.5.1 (2019-12-10)
## Fixed
- Integration version reporting

## 1.5.0 (2019-12-10)
## Added
- Added `validate_certs` configuration option (default: `true`). Set it to `false` if you have self-signed certificates
  and want to avoid the integration to fail.

## 1.4.0 (2019-11-18)
### Changed
- Renamed the integration executable from nr-nginx to nri-nginx in order to be consistent with the package naming. **Important Note:** if you have any security module rules (eg. SELinux), alerts or automation that depends on the name of this binary, these will have to be updated.
## 1.3.1 (2019-09-19)
- Fixed automatic discovery of the `status_module`.

## 1.3.0 (2019-08-09)
### Added
- Support for `ngx_http_api_module`.
- New configuration options:
    - `connection_timeout`: timeout (in seconds) for the connection from the integration to Nginx
    - `status_module` (default: `discover`). Accepted values:
        * `ngx_http_stub_status_module`
        * `ngx_http_status_module`
        * `ngx_http_api_module`
        * `discover` to automatically choose between `ngx_http_stub_status_module` or `ngx_http_status_module`.
    - `endpoints`: if `status_module` is `ngx_http_api_module`, comma separated list, NON PARAMETERIZED, Endpoints   

## 1.2.0 (2019-04-29)
### Added
- Upgraded to SDK v3.1.5. This version implements [the aget/integrations
  protocol v3](https://github.com/newrelic/infra-integrations-sdk/blob/cb45adacda1cd5ff01544a9d2dad3b0fedf13bf1/docs/protocol-v3.md),
  which enables [name local address replacement](https://github.com/newrelic/infra-integrations-sdk/blob/cb45adacda1cd5ff01544a9d2dad3b0fedf13bf1/docs/protocol-v3.md#name-local-address-replacement).
  and could change your entity names and alarms. For more information, refer
  to:

  - https://docs.newrelic.com/docs/integrations/integrations-sdk/file-specifications/integration-executable-file-specifications#h2-loopback-address-replacement-on-entity-names
  - https://docs.newrelic.com/docs/remote-monitoring-host-integration://docs.newrelic.com/docs/remote-monitoring-host-integrations

## 1.1.0 (2019-04-08)
### Added
- Upgraded to SDKv3
- Remote monitoring option. It enables monitoring multiple NGINX instances,
  more information can be found at the [official documentation page](https://docs.newrelic.com/docs/remote-monitoring-host-integrations).

## 1.0.2 (2018-11-05)
### Fixed
- Issue where if 'config_path' provided is a directory the integration generates high CPU load.
- Issue where if nginx status page response code was different than 200

## 1.0.1 (2018-09-07)
### Changed
- Update Makefile

## 1.0.0 (2018-08-02)
### Added
- Initial release, which contains inventory and metrics data
