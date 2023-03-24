export PATH := $(PATH):$(GOPATH)/bin

INTEGRATION     := nginx
BINARY_NAME      = nri-$(INTEGRATION)
INTEGRATIONS_DIR = /var/db/newrelic-infra/newrelic-integrations/
CONFIG_DIR       = /etc/newrelic-infra/integrations.d
GO_FILES        := ./src/
GOFLAGS          = -mod=readonly # ignore the vendor directory and to report an error if go.mod needs to be updated.

all: build

build: clean compile test

clean:
	@echo "=== $(INTEGRATION) === [ clean ]: cleaning workspace..."
	@rm -rfv bin coverage.xml

bin/$(BINARY_NAME):
	@echo "=== $(INTEGRATION) === [ compile ]: building $(BINARY_NAME)..."
	@go build -v -o bin/$(BINARY_NAME) $(GO_FILES)

compile: bin/$(BINARY_NAME)

test:
	@echo "=== $(INTEGRATION) === [ test ]: running unit tests..."
	@go test -race ./... -count=1

integration-test:
	@echo "=== $(INTEGRATION) === [ test ]: running integration tests..."
	@docker-compose -f tests/integration/docker-compose.yml up -d --build
	@go test -v -tags=integration ./tests/integration/. || (ret=$$?; docker-compose -f tests/integration/docker-compose.yml down && exit $$ret)
	@docker-compose -f tests/integration/docker-compose.yml down

install: compile
	@echo "=== $(INTEGRATION) === [ install ]: installing bin/$(BINARY_NAME)..."
	@sudo install -D --mode=755 --owner=root --strip $(ROOT)bin/$(BINARY_NAME) $(INTEGRATIONS_DIR)/bin/$(BINARY_NAME)
	@sudo install -D --mode=644 --owner=root $(ROOT)$(INTEGRATION)-config.yml.sample $(CONFIG_DIR)/$(INTEGRATION)-config.yml.sample

# Include thematic Makefiles
include $(CURDIR)/build/ci.mk
include $(CURDIR)/build/release.mk

.PHONY: all build clean compile test integration-test install
