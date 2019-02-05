// +build integration

package integration

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/newrelic/nri-nginx/tests/integration/helpers"
	"github.com/newrelic/nri-nginx/tests/integration/jsonschema"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	simplejson "github.com/bitly/go-simplejson"

)

var iName = "nginx"
var binaryPathName = fmt.Sprintf("/var/db/newrelic-infra/newrelic-integrations/bin/nr-%s", iName)
var update = flag.Bool("test.update", false, "update json-schema file")

func setup() error {
	err := helpers.CheckIntegrationIsInstalled(iName)
	if err != nil {
		return fmt.Errorf("NGINX integration isn't properly installed. Err: %s", err)
	}
	return nil
}

func teardown() error {
	return nil
}

func TestMain(m *testing.M) {
	err := setup()
	if err != nil {
		fmt.Println(err)
		tErr := teardown()
		if tErr != nil {
			fmt.Printf("Error during the teardown of the tests: %s\n", tErr)
		}
		os.Exit(1)
	}
	result := m.Run()
	err = teardown()
	if err != nil {
		fmt.Printf("Error during the teardown of the tests: %s\n", err)
	}
	os.Exit(result)
}

func TestNGINXIntegration(t *testing.T) {
	testName := helpers.GetTestName(t)
	cmd := exec.Command(binaryPathName)
	cmd.Env = []string{
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	}
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err != nil {
		t.Fatalf("It isn't possible to execute NGINX integration binary. Err: %s -- %s", err, errbuf.String())
	}

	schemaPath := filepath.Join("json-schema-files", "nginx-schema.json")
	if *update {
		// The tool json-schema-generator doesn't handle properly this case using '--stdin' option.
		// jsonschema.Generate fails with syntax error. As workaround we use json-schema-generator
		// with option '--file', which works properly
		_, err := ioutil.TempFile("", "tmp-output.json")
		if err != nil {
			t.Fatalf("Cannot create a new temporary file, got error: %v", err)
		}
		defer os.Remove("./tmp-output.json") // clean up
		err = ioutil.WriteFile("tmp-output.json", outbuf.Bytes(), 0644)
		if err != nil {
			t.Fatal(err)
		}
		schema, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("json-schema-generator --file ./tmp-output.json | sed -e 1d")).Output()
		if err != nil {
			t.Fatal(err)
		}

		schemaJSON, err := simplejson.NewJson(schema)
		if err != nil {
			t.Fatalf("Cannot unmarshal JSON schema, got error: %v", err)
		}
		err = helpers.ModifyJSONSchemaGlobal(schemaJSON, iName, 1, "1.0.0")
		if err != nil {
			t.Fatal(err)
		}
		err = helpers.ModifyJSONSchemaInventoryPresent(schemaJSON)
		if err != nil {
			t.Fatal(err)
		}
		err = helpers.ModifyJSONSchemaMetricsPresent(schemaJSON, "NginxSample")
		if err != nil {
			t.Fatal(err)
		}
		schema, err = schemaJSON.MarshalJSON()
		if err != nil {
			t.Fatalf("Cannot marshal JSON schema, got error: %v", err)
		}
		err = ioutil.WriteFile(schemaPath, schema, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = jsonschema.Validate(schemaPath, outbuf.String())
	if err != nil {
		t.Fatalf("The output of NGINX integration doesn't have expected format. Err: %s", err)
	}
}

func TestNGINXIntegrationOnlyMetrics(t *testing.T) {
	testName := helpers.GetTestName(t)
	cmd := exec.Command(binaryPathName)
	cmd.Env = []string{
		"METRICS=true",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	}
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err != nil {
		t.Fatalf("It isn't possible to execute NGINX integration binary. Err: %s -- %s", err, errbuf.String())
	}

	schemaPath := filepath.Join("json-schema-files", "nginx-schema-metrics.json")
	if *update {
		schema, err := jsonschema.Generate(outbuf.String())
		if err != nil {
			t.Fatal(err)
		}
		schemaJSON, err := simplejson.NewJson(schema)
		if err != nil {
			t.Fatalf("Cannot unmarshal JSON schema, got error: %v", err)
		}
		err = helpers.ModifyJSONSchemaGlobal(schemaJSON, iName, 1, "1.0.0")
		if err != nil {
			t.Fatal(err)
		}
		err = helpers.ModifyJSONSchemaNoInventory(schemaJSON)
		if err != nil {
			t.Fatal(err)
		}
		err = helpers.ModifyJSONSchemaMetricsPresent(schemaJSON, "NginxSample")
		if err != nil {
			t.Fatal(err)
		}
		schema, err = schemaJSON.MarshalJSON()
		if err != nil {
			t.Fatalf("Cannot marshal JSON schema, got error: %v", err)
		}
		err = ioutil.WriteFile(schemaPath, schema, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = jsonschema.Validate(schemaPath, outbuf.String())
	if err != nil {
		t.Fatalf("The output of NGINX integration doesn't have expected format. Err: %s", err)
	}
}

func TestNGINXIntegrationOnlyInventory(t *testing.T) {
	testName := helpers.GetTestName(t)
	cmd := exec.Command(binaryPathName)
	cmd.Env = []string{
		"INVENTORY=true",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	}
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err != nil {
		t.Fatalf("It isn't possible to execute NGINX integration binary. Err: %s -- %s", err, errbuf.String())
	}

	schemaPath := filepath.Join("json-schema-files", "nginx-schema-inventory.json")
	if *update {
		// The tool json-schema-generator doesn't handle properly this case using '--stdin' option.
		// jsonschema.Generate fails with syntax error. As workaround we use json-schema-generator
		// with option '--file', which works properly
		_, err := ioutil.TempFile("", "tmp-output.json")
		if err != nil {
			t.Fatalf("Cannot create a new temporary file, got error: %v", err)
		}
		defer os.Remove("./tmp-output.json") // clean up
		err = ioutil.WriteFile("tmp-output.json", outbuf.Bytes(), 0644)
		if err != nil {
			t.Fatal(err)
		}
		schema, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("json-schema-generator --file ./tmp-output.json | sed -e 1d")).Output()
		if err != nil {
			t.Fatal(err)
		}

		schemaJSON, err := simplejson.NewJson(schema)
		if err != nil {
			t.Fatalf("Cannot unmarshal JSON schema, got error: %v", err)
		}
		err = helpers.ModifyJSONSchemaGlobal(schemaJSON, iName, 1, "1.0.0")
		if err != nil {
			t.Fatal(err)
		}
		err = helpers.ModifyJSONSchemaInventoryPresent(schemaJSON)
		if err != nil {
			t.Fatal(err)
		}
		err = helpers.ModifyJSONSchemaNoMetrics(schemaJSON)
		if err != nil {
			t.Fatal(err)
		}
		schema, err = schemaJSON.MarshalJSON()
		if err != nil {
			t.Fatalf("Cannot marshal JSON schema, got error: %v", err)
		}
		err = ioutil.WriteFile(schemaPath, schema, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = jsonschema.Validate(schemaPath, outbuf.String())
	if err != nil {
		t.Fatalf("The output of NGINX integration doesn't have expected format. Err: %s", err)
	}
}

func TestNGINXIntegrationInvalidStatusURL(t *testing.T) {
	t.Skip("Skipping test - fix in the NGINX integration required")
	testName := helpers.GetTestName(t)
	cmd := exec.Command(binaryPathName)
	cmd.Env = []string{
		"STATUS_URL=http://localhost/",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
		"VERBOSE=true",
	}

	expectedErrorMessage := "Cannot fetch metrics data"
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()

	errMatch, _ := regexp.MatchString(expectedErrorMessage, errbuf.String())
	if err == nil || !errMatch {
		t.Fatalf("Expected error message: '%s', got: '%s'", expectedErrorMessage, errbuf.String())
	}
	if outbuf.String() != "" {
		t.Fatalf("Unexpected output: %s", outbuf.String())
	}
}

func TestNGINXIntegrationInvalidStatusURL_NoExistingHost(t *testing.T) {
	testName := helpers.GetTestName(t)
	cmd := exec.Command(binaryPathName)
	cmd.Env = []string{
		"STATUS_URL=http://nonExistingHost/status",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	}

	expectedErrorMessage := "no such host"
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()

	errMatch, _ := regexp.MatchString(expectedErrorMessage, errbuf.String())
	if err == nil || !errMatch {
		t.Fatalf("Expected error message: '%s', got: '%s'", expectedErrorMessage, errbuf.String())
	}
	if outbuf.String() != "" {
		t.Fatalf("Unexpected output: %s", outbuf.String())
	}
}

func TestNGINXIntegrationNotValidURL_NoHttp(t *testing.T) {
	testName := helpers.GetTestName(t)
	cmd := exec.Command(binaryPathName)
	cmd.Env = []string{
		"STATUS_URL=localhost/status",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	}

	expectedErrorMessage := "unsupported protocol scheme"
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()

	errMatch, _ := regexp.MatchString(expectedErrorMessage, errbuf.String())
	if err == nil || !errMatch {
		t.Fatalf("Expected error message: '%s', got: '%s'", expectedErrorMessage, errbuf.String())
	}
	if outbuf.String() != "" {
		t.Fatalf("Unexpected output: %s", outbuf.String())
	}
}

func TestNGINXIntegrationNotValidURL_OnlyHttp(t *testing.T) {
	testName := helpers.GetTestName(t)
	cmd := exec.Command(binaryPathName)
	cmd.Env = []string{
		"STATUS_URL=http://",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	}

	expectedErrorMessage := "no Host in request URL"
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()

	errMatch, _ := regexp.MatchString(expectedErrorMessage, errbuf.String())
	if err == nil || !errMatch {
		t.Fatalf("Expected error message: '%s', got: '%s'", expectedErrorMessage, errbuf.String())
	}
	if outbuf.String() != "" {
		t.Fatalf("Unexpected output: %s", outbuf.String())
	}
}

// This test hangs. Also the expectedErrorMessage is not defined
// because currently the error is not returned. More details:
// https://newrelic.atlassian.net/browse/IHOST-83
func TestNGINXIntegrationNotValidConfigPath_ExistingDirectory(t *testing.T) {
	t.Skip("Skipping test becuase it hangs")
	testName := helpers.GetTestName(t)
	cmd := exec.Command(binaryPathName)
	cmd.Env = []string{
		"CONFIG_PATH=/etc/nginx/",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	}

	expectedErrorMessage := "..."
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()

	errMatch, _ := regexp.MatchString(expectedErrorMessage, errbuf.String())
	if err == nil || !errMatch {
		t.Fatalf("Expected error message: '%s', got: '%s'", expectedErrorMessage, errbuf.String())
	}

	if outbuf.String() != "" {
		t.Fatalf("Unexpected output: %s", outbuf.String())
	}
}

func TestNGINXIntegrationNotValidConfigPath_NonExistingFile(t *testing.T) {
	testName := helpers.GetTestName(t)
	cmd := exec.Command(binaryPathName)
	cmd.Env = []string{
		"CONFIG_PATH=/nonExisting/nginx.conf",
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	}

	expectedErrorMessage := "no such file or directory"
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()

	errMatch, _ := regexp.MatchString(expectedErrorMessage, errbuf.String())
	if err == nil || !errMatch {
		t.Fatalf("Expected error message: '%s', got: '%s'", expectedErrorMessage, errbuf.String())
	}

	if outbuf.String() != "" {
		t.Fatalf("Unexpected output: %s", outbuf.String())
	}
}

func TestNGINXIntegrationNotValidConfigPath_ExistingFile(t *testing.T) {
	t.Skip("Skipping test - fix in the NGINX integration required")
	tmpfile, err := ioutil.TempFile("", "empty.conf")
	if err != nil {
		t.Fatalf("Cannot create a new temporary file, got error: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	testName := helpers.GetTestName(t)
	cmd := exec.Command(binaryPathName)
	cmd.Env = []string{
		fmt.Sprintf("CONFIG_PATH=%s", tmpfile.Name()),
		fmt.Sprintf("NRIA_CACHE_PATH=%v", testName),
	}

	expectedErrorMessage := "Config path for NGINX not correctly set, cannot fetch inventory data"
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err = cmd.Run()
	if err != nil {
		t.Fatalf("It isn't possible to execute NGINX integration binary. Err: %s -- %s", err, errbuf.String())
	}

	schemaPath := filepath.Join("json-schema-files", "nginx-schema-metrics.json")
	if *update {
		schema, err := jsonschema.Generate(outbuf.String())
		if err != nil {
			t.Fatal(err)
		}
		schemaJSON, err := simplejson.NewJson(schema)
		if err != nil {
			t.Fatalf("Cannot unmarshal JSON schema, got error: %v", err)
		}
		err = helpers.ModifyJSONSchemaGlobal(schemaJSON, iName, 1, "1.0.0")
		if err != nil {
			t.Fatal(err)
		}
		err = helpers.ModifyJSONSchemaNoInventory(schemaJSON)
		if err != nil {
			t.Fatal(err)
		}
		err = helpers.ModifyJSONSchemaMetricsPresent(schemaJSON, "NginxSample")
		if err != nil {
			t.Fatal(err)
		}
		schema, err = schemaJSON.MarshalJSON()
		if err != nil {
			t.Fatalf("Cannot marshal JSON schema, got error: %v", err)
		}
		err = ioutil.WriteFile(schemaPath, schema, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = jsonschema.Validate(schemaPath, outbuf.String())
	if err != nil {
		t.Fatalf("The output of NGINX integration doesn't have expected format. Err: %s", err)
	}

	errMatch, _ := regexp.MatchString(expectedErrorMessage, errbuf.String())
	if !errMatch {
		t.Fatalf("Expected warning message: '%s', got: '%s'", expectedErrorMessage, errbuf.String())
	}
}