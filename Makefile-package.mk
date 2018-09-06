PACKAGE_TYPES     ?= deb rpm
PROJECT_NAME       = nri-$(INTEGRATION)
BINS_DIR           = $(TARGET_DIR)/bin/linux_amd64
SOURCE_DIR         = $(TARGET_DIR)/source
PACKAGES_DIR       = $(TARGET_DIR)/packages
VERSION           ?= 0.0.0
RELEASE           ?= dev
LICENSE            = "https://newrelic.com/terms (also see LICENSE.txt installed with this package)"
VENDOR             = "New Relic, Inc."
PACKAGER           = "New Relic Infrastructure Team <infrastructure-eng@newrelic.com>"
PACKAGE_URL        = "https://www.newrelic.com/infrastructure"
SUMMARY            = "New Relic Infrastructure $(INTEGRATION) Integration"
DESCRIPTION        = "New Relic Infrastructure $(INTEGRATION) Integration extend the core New Relic\nInfrastructure agent's capabilities to allow you to collect metric and\nlive state data from your infrastructure $(INTEGRATION) components."
FPM_COMMON_OPTIONS = --verbose -C $(SOURCE_DIR) -s dir -n $(PROJECT_NAME) -v $(VERSION) --iteration $(RELEASE) --prefix "" --license $(LICENSE) --vendor $(VENDOR) -m $(PACKAGER) --url $(PACKAGE_URL) --config-files /etc/newrelic-infra/ --description "$$(printf $(DESCRIPTION))" --depends "newrelic-infra >= 1.0.726"
FPM_DEB_OPTIONS    = -t deb -p $(PACKAGES_DIR)/deb/  --replaces "newrelic-infra-integrations (<= 1.4.0)"
FPM_RPM_OPTIONS    = -t rpm -p $(PACKAGES_DIR)/rpm/ --epoch 0 --rpm-summary $(SUMMARY)  --replaces "newrelic-infra-integrations <= 1.4.0"

package: create-bins prep-pkg-env $(PACKAGE_TYPES)

create-bins:
	echo "=== Main === [ create-bins ]: creating binary ..."
	go build -v -ldflags '-X main.buildVersion=$(VERSION)' -o $(BINS_DIR)/$(BINARY_NAME) $(GO_FILES) || exit 1
	@echo ""

prep-pkg-env:
	@if [ ! -d $(BINS_DIR) ]; then \
		echo "=== Main === [ prep-pkg-env ]: no built binaries found. Run 'make create-bins'" ;\
		exit 1 ;\
	fi
	@echo "=== Main === [ prep-pkg-env ]: preparing a clean packaging environment..."
	@rm -rf $(SOURCE_DIR)
	@mkdir -p $(SOURCE_DIR)/var/db/newrelic-infra/newrelic-integrations/bin $(SOURCE_DIR)/etc/newrelic-infra/integrations.d
	@echo "=== Main === [ prep-pkg-env ]: adding built binaries and configuration and definition files..."
	@cp $(BINS_DIR)/$(BINARY_NAME) $(SOURCE_DIR)/var/db/newrelic-infra/newrelic-integrations/bin
	@chmod 755 $(SOURCE_DIR)/var/db/newrelic-infra/newrelic-integrations/bin/*
	@cp ./*.yml $(SOURCE_DIR)/var/db/newrelic-infra/newrelic-integrations/
	@chmod 644 $(SOURCE_DIR)/var/db/newrelic-infra/newrelic-integrations/*.yml
	@cp ./*.sample $(SOURCE_DIR)/etc/newrelic-infra/integrations.d/
	@chmod 644 $(SOURCE_DIR)/etc/newrelic-infra/integrations.d/*.sample

deb: prep-pkg-env
	@echo "=== Main === [ deb ]: building DEB package..."
	@mkdir -p $(PACKAGES_DIR)/deb
	@fpm $(FPM_COMMON_OPTIONS) $(FPM_DEB_OPTIONS) .

rpm: prep-pkg-env
	@echo "=== Main === [ rpm ]: building RPM package..."
	@mkdir -p $(PACKAGES_DIR)/rpm
	@fpm $(FPM_COMMON_OPTIONS) $(FPM_RPM_OPTIONS) .

.PHONY: package create-bins prep-pkg-env $(PACKAGE_TYPES)
