# Change Log
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

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
