export PATH := $(PATH):$(GOPATH)/bin

INTEGRATION     := nginx
BINARY_NAME      = nri-$(INTEGRATION)
SRC_DIR          = ./src/
#VALIDATE_DEPS    = github.com/golangci/golangci-lint
#TEST_DEPS        = github.com/axw/gocov github.com/AlekSi/gocov-xml
INTEGRATIONS_DIR = /var/db/newrelic-infra/newrelic-integrations/
CONFIG_DIR       = /etc/newrelic-infra/integrations.d
GO_FILES        := ./src/
GOLANGCI_LINT	 = github.com/golangci/golangci-lint/cmd/golangci-lint
GOCOV            = github.com/axw/gocov/gocov
GOCOV_XML		 = github.com/AlekSi/gocov-xml

all: build

build: clean validate compile test

clean:
	@echo "=== $(INTEGRATION) === [ clean ]: removing binaries and coverage file..."
	@rm -rfv bin coverage.xml

validate:
ifeq ($(strip $(GO_FILES)),)
	@echo "=== $(INTEGRATION) === [ validate ]: no Go files found. Skipping validation."
else
	@printf "=== $(INTEGRATION) === [ validate ]: running golangci-lint... "
	@go run -mod=vendor $(GOLANGCI_LINT) run --verbose
#	@OUTPUT="$(shell gofmt -l $(GO_FILES))" ;\
#	if [ -z "$$OUTPUT" ]; then \
#		echo "passed." ;\
#	else \
#		echo "failed. Incorrect syntax in the following files:" ;\
#		echo "$$OUTPUT" ;\
#		exit 1 ;\
#	fi
#	@printf "=== $(INTEGRATION) === [ validate ]: running golint... "
#	@OUTPUT="$(shell golint $(SRC_DIR)...)" ;\
#	if [ -z "$$OUTPUT" ]; then \
#		echo "passed." ;\
#	else \
#		echo "failed. Issues found:" ;\
#		echo "$$OUTPUT" ;\
#		exit 1 ;\
#	fi
#	@printf "=== $(INTEGRATION) === [ validate ]: running go vet... "
#	@OUTPUT="$(shell go vet $(SRC_DIR)...)" ;\
#	if [ -z "$$OUTPUT" ]; then \
#		echo "passed." ;\
#	else \
#		echo "failed. Issues found:" ;\
#		echo "$$OUTPUT" ;\
#		exit 1;\
#	fi
endif

compile-deps:
	@echo "=== $(INTEGRATION) === [ compile-deps ]: installing build dependencies..."
#	@go get -v -d -t ./...

bin/$(BINARY_NAME):
	@echo "=== $(INTEGRATION) === [ compile ]: building $(BINARY_NAME)..."
	@go build -v -o bin/$(BINARY_NAME) $(GO_FILES)

compile: compile-deps bin/$(BINARY_NAME)

test-deps: compile-deps
	@echo "=== $(INTEGRATION) === [ test-deps ]: installing testing dependencies..."
#	@go get -v $(TEST_DEPS)

test-only:
	@echo "=== $(INTEGRATION) === [ test ]: running unit tests..."
	@go run -mod=vendor $(GOCOV) test ./... | go run -mod=vendor $(GOCOV_XML) > coverage.xml

test: test-deps test-only

integration-test: test-deps
	@echo "=== $(INTEGRATION) === [ test ]: running integration tests..."
	@docker-compose -f tests/integration/docker-compose.yml up -d --build
	@go test -v -tags=integration ./tests/integration/. || (ret=$$?; docker-compose -f tests/integration/docker-compose.yml down && exit $$ret)
	@docker-compose -f tests/integration/docker-compose.yml down

install: bin/$(BINARY_NAME)
	@echo "=== $(INTEGRATION) === [ install ]: installing bin/$(BINARY_NAME)..."
	@sudo install -D --mode=755 --owner=root --strip $(ROOT)bin/$(BINARY_NAME) $(INTEGRATIONS_DIR)/bin/$(BINARY_NAME)
	@sudo install -D --mode=644 --owner=root $(ROOT)$(INTEGRATION)-definition.yml $(INTEGRATIONS_DIR)/$(INTEGRATION)-definition.yml
	@sudo install -D --mode=644 --owner=root $(ROOT)$(INTEGRATION)-config.yml.sample $(CONFIG_DIR)/$(INTEGRATION)-config.yml.sample

# Include thematic Makefiles
include $(CURDIR)/build/ci.mk
include $(CURDIR)/build/release.mk

.PHONY: all build clean validate compile-deps compile test-deps test-only test integration-test install
