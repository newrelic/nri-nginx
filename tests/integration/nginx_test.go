// +build integration

package integration

import (
	"flag"
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-nginx/tests/integration/helpers"
	"github.com/newrelic/nri-nginx/tests/integration/jsonschema"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	iName = "nginx"

	defaultContainer = "integration_nri-nginx_1"

	defaultBinPath   = "/nri-nginx"
	defaultStatusURL = "http://nginx:8080/status"

	// cli flags
	container = flag.String("container", defaultContainer, "container where the integration is installed")
	binPath   = flag.String("bin", defaultBinPath, "Integration binary path")

	statusURL = flag.String("status_url", defaultStatusURL, "nginx status url")
)

// Returns the standard output, or fails testing if the command returned an error
func runIntegration(t *testing.T, envVars ...string) (string, string, error) {
	t.Helper()

	command := make([]string, 0)
	command = append(command, *binPath)

	var found bool

	for _, envVar := range envVars {
		if strings.HasPrefix(envVar, "STATUS_URL") {
			found = true
			break
		}
	}

	if !found && statusURL != nil {
		command = append(command, "--status_url", *statusURL)
	}

	stdout, stderr, err := helpers.ExecInContainer(*container, command, envVars...)

	if stderr != "" {
		log.Debug("Integration command Standard Error: ", stderr)
	}

	return stdout, stderr, err
}

func TestMain(m *testing.M) {
	flag.Parse()

	result := m.Run()
	os.Exit(result)
}

func TestNGINXIntegration(t *testing.T) {
	testName := helpers.GetTestName(t)
	stdout, stderr, err := runIntegration(t, fmt.Sprintf("NRIA_CACHE_PATH=/tmp/%v.json", testName))

	assert.NotNil(t, stderr, "unexpected stderr")
	assert.NoError(t, err, "Unexpected error")

	schemaPath := filepath.Join("json-schema-files", "nginx-schema.json")

	err = jsonschema.Validate(schemaPath, stdout)
	assert.NoError(t, err, "The output of NGINX integration doesn't have expected format.")
}

func TestNGINXIntegrationOnlyMetrics(t *testing.T) {
	testName := helpers.GetTestName(t)
	stdout, stderr, err := runIntegration(t, "METRICS=true", fmt.Sprintf("NRIA_CACHE_PATH=/tmp/%v.json", testName))

	assert.NotNil(t, stderr, "unexpected stderr")
	assert.NoError(t, err, "Unexpected error")

	schemaPath := filepath.Join("json-schema-files", "nginx-schema-metrics.json")

	err = jsonschema.Validate(schemaPath, stdout)
	assert.NoError(t, err, "The output of NGINX integration doesn't have expected format.")
}

func TestNGINXIntegrationOnlyInventory(t *testing.T) {
	testName := helpers.GetTestName(t)
	stdout, stderr, err := runIntegration(t, "INVENTORY=true", fmt.Sprintf("NRIA_CACHE_PATH=/tmp/%v.json", testName))

	assert.NotNil(t, stderr, "unexpected stderr")
	assert.NoError(t, err, "Unexpected error")

	schemaPath := filepath.Join("json-schema-files", "nginx-schema-inventory.json")

	err = jsonschema.Validate(schemaPath, stdout)
	assert.NoError(t, err, "The output of NGINX integration doesn't have expected format.")
}

func TestNGINXIntegrationInvalidStatusURL(t *testing.T) {
	testName := helpers.GetTestName(t)

	stdout, stderr, err := runIntegration(t, "STATUS_URL=http://localhost/", fmt.Sprintf("NRIA_CACHE_PATH=%v", testName), "VERBOSE=true")

	expectedErrorMessage := "connection refused"

	errMatch, _ := regexp.MatchString(expectedErrorMessage, stderr)
	assert.Error(t, err, "Expected error")
	assert.Truef(t, errMatch, "Expected error message: '%s', got: '%s'", expectedErrorMessage, stderr)

	assert.NotNil(t, stdout, "unexpected stdout")
}

func TestNGINXIntegrationInvalidStatusURL_NoExistingHost(t *testing.T) {
	testName := helpers.GetTestName(t)

	stdout, stderr, err := runIntegration(t, "STATUS_URL=http://nonExistingHost/status", fmt.Sprintf("NRIA_CACHE_PATH=%v", testName))

	expectedErrorMessage := "no such host"

	errMatch, _ := regexp.MatchString(expectedErrorMessage, stderr)
	assert.Error(t, err, "Expected error")
	assert.Truef(t, errMatch, "Expected error message: '%s', got: '%s'", expectedErrorMessage, stderr)

	assert.NotNil(t, stdout, "unexpected stdout")
}

func TestNGINXIntegrationNotValidURL_NoHttp(t *testing.T) {
	testName := helpers.GetTestName(t)

	stdout, stderr, err := runIntegration(t, "STATUS_URL=localhost/status", fmt.Sprintf("NRIA_CACHE_PATH=/tmp/%v.json", testName))

	expectedErrorMessage := "unsupported protocol scheme"

	errMatch, _ := regexp.MatchString(expectedErrorMessage, stderr)
	assert.Error(t, err, "Expected error")
	assert.Truef(t, errMatch, "Expected error message: '%s', got: '%s'", expectedErrorMessage, stderr)

	assert.NotNil(t, stdout, "unexpected stdout")
}

func TestNGINXIntegrationNotValidURL_OnlyHttp(t *testing.T) {
	testName := helpers.GetTestName(t)

	stdout, stderr, err := runIntegration(t, "STATUS_URL=http://", fmt.Sprintf("NRIA_CACHE_PATH=/tmp/%v.json", testName))

	expectedErrorMessage := "no Host in request URL"

	errMatch, _ := regexp.MatchString(expectedErrorMessage, stderr)
	assert.Error(t, err, "Expected error")
	assert.Truef(t, errMatch, "Expected error message: '%s', got: '%s'", expectedErrorMessage, stderr)

	assert.NotNil(t, stdout, "unexpected stdout")
}

func TestNGINXIntegrationNotValidConfigPath_ExistingDirectory(t *testing.T) {
	testName := helpers.GetTestName(t)

	stdout, stderr, err := runIntegration(t, "CONFIG_PATH=/etc/nginx/", fmt.Sprintf("NRIA_CACHE_PATH=/tmp/%v.json", testName))

	expectedErrorMessage := ": is a directory"

	errMatch, _ := regexp.MatchString(expectedErrorMessage, stderr)
	assert.Error(t, err, "Expected error")
	assert.Truef(t, errMatch, "Expected error message: '%s', got: '%s'", expectedErrorMessage, stderr)

	assert.NotNil(t, stdout, "unexpected stdout")
}

func TestNGINXIntegrationNotValidConfigPath_NonExistingFile(t *testing.T) {
	testName := helpers.GetTestName(t)

	stdout, stderr, err := runIntegration(t, "CONFIG_PATH=/nonExisting/nginx.conf", fmt.Sprintf("NRIA_CACHE_PATH=/tmp/%v.json", testName))

	expectedErrorMessage := "no such file or directory"

	errMatch, _ := regexp.MatchString(expectedErrorMessage, stderr)
	assert.Error(t, err, "Expected error")
	assert.Truef(t, errMatch, "Expected error message: '%s', got: '%s'", expectedErrorMessage, stderr)

	assert.NotNil(t, stdout, "unexpected stdout")
}
