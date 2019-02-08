package helpers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bitly/go-simplejson"
	"github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
)

// ValidateJSONSchema validates the input argument against JSON schema. If the
// input is not valid the error is returned. The first argument is the file name
// (without .json extension) of the JSON schema. It is used to build file URI
// required to load the JSON schema. The second argument is the input string that
// is validated.
//
// Deprecated: This function is deprecated. Instead, use Validate function from
// jsonschema package
func ValidateJSONSchema(schemaJsonFileName string, input string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	schemaURI := fmt.Sprintf("file://%s.%s", filepath.Join(pwd, schemaJsonFileName), "json")

	schemaLoader := gojsonschema.NewReferenceLoader(schemaURI)
	documentLoader := gojsonschema.NewStringLoader(input)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("Error loading JSON schema, error: %v", err)
	}

	if result.Valid() {
		return nil
	}
	fmt.Printf("Errors for JSON schema: '%s'\n", schemaURI)
	for _, desc := range result.Errors() {
		fmt.Printf("\t- %s\n", desc)
	}
	fmt.Printf("\n")
	return fmt.Errorf("The output of the integration doesn't have expected JSON format")
}

// GetTestName returns the name of the running test.
func GetTestName(t *testing.T) interface{} {
	v := reflect.ValueOf(*t)
	return v.FieldByName("name")
}

// Deprecated: Instead, use jsonschema.ValidationField
type schemaElement struct {
	schemaField string
	schemaValue interface{}
}

// Deprecated: Instead, use function jsonschema.AddNewElements
func addNewElementsToJSONSchema(location *simplejson.Json, newElements map[string]schemaElement) error {
	var elemLocation *simplejson.Json
	var ok bool
	for key, value := range newElements {
		if elemLocation, ok = location.CheckGet(key); !ok {
			return fmt.Errorf("Cannot update JSON schema with value: %s  for element: %s", value.schemaField, key)
		}
		elemLocation.Set(value.schemaField, value.schemaValue)
	}
	return nil
}

// ModifyJSONSchemaGlobal modifies JSON schema by adding patterns elements for integration name, protocol version and integration version
func ModifyJSONSchemaGlobal(schema *simplejson.Json, integrationName string, protocolVersion int, integrationVersion string) error {
	return addNewElementsToJSONSchema(
		schema.Get("properties"),
		map[string]schemaElement{
			"name":                {"pattern", fmt.Sprintf("^com.newrelic.%s$", integrationName)},
			"protocol_version":    {"pattern", fmt.Sprintf("^%d$", protocolVersion)},
			"integration_version": {"pattern", fmt.Sprintf("^%s$", integrationVersion)},
		})
}

// ModifyJSONSchemaMetricsPresent modifies JSON schema by adding required elements for metrics JSON schema
func ModifyJSONSchemaMetricsPresent(schema *simplejson.Json, eventType string) error {
	mainProperties := schema.Get("properties")

	elementOfMetrics := map[string]schemaElement{
		"metrics": {"minItems", 1},
	}
	err := addNewElementsToJSONSchema(mainProperties, elementOfMetrics)
	if err != nil {
		return err
	}

	metrics, ok := mainProperties.CheckGet("metrics")
	if !ok {
		return fmt.Errorf("Cannot find metrics element")
	}
	metricsItems, ok := metrics.CheckGet("items")
	if !ok {
		return fmt.Errorf("Cannot find metrics items element")
	}
	metricsItemsProperties := metricsItems.Get("properties")
	elementsOfMetricsProperties := map[string]schemaElement{
		"event_type": {"pattern", fmt.Sprintf("^%s$", eventType)},
	}
	err = addNewElementsToJSONSchema(metricsItemsProperties, elementsOfMetricsProperties)
	if err != nil {
		return err
	}

	items, err := metricsItems.Map()
	if err != nil {
		return fmt.Errorf("Not expected metrics structure for 'item' element, got error: %v", err)
	}
	tmp := []interface{}{items}
	metrics.Set("items", tmp)

	// TODO: iterate through list of items and modify each itemSet by adding event_type pattern
	// itemArray, err := metrics.Get("items").Array()
	// for _, itemSet := range itemArray {
	// 	jsonSet := simplejson.Json{itemSet} // fails
	// }

	return nil
}

// ModifyJSONSchemaNoMetrics modifies JSON schema by adding required elements
// assuring that no metrics data exists in the integration output
func ModifyJSONSchemaNoMetrics(schema *simplejson.Json) error {
	return addNewElementsToJSONSchema(
		schema.Get("properties"),
		map[string]schemaElement{
			"metrics": {"maxItems", 0},
		})
}

// ModifyJSONSchemaNoInventory modifies JSON schema by adding required elements
// assuring that no inventory data exists in the integration output
func ModifyJSONSchemaNoInventory(schema *simplejson.Json) error {
	return addNewElementsToJSONSchema(
		schema.Get("properties"),
		map[string]schemaElement{
			"inventory": {"maxProperties", 0},
		})
}

// ModifyJSONSchemaInventoryPresent modifies JSON schema by adding required elements for inventory JSON schema
func ModifyJSONSchemaInventoryPresent(schema *simplejson.Json) error {
	mainProperties := schema.Get("properties")

	elementOfMainInventory := map[string]schemaElement{
		"inventory": {"minProperties", 1},
	}
	err := addNewElementsToJSONSchema(mainProperties, elementOfMainInventory)
	if err != nil {
		return err
	}

	// `required` section for `inventory` should be empty, as none of the inventory data is obligatory
	mainProperties.Get("inventory").Set("required", make([]interface{}, 0))
	return nil
}

// ExecInContainer executes the given command inside the specified container. It returns three values:
// 1st - Standard Output
// 2nd - Standard Error
// 3rd - Runtime error, if any
func ExecInContainer(container string, command []string, envVars ...string) (string, string, error) {
	cmdLine := make([]string, 0, 3+len(command))
	cmdLine = append(cmdLine, "exec", "-i")

	for _, envVar := range envVars {
		cmdLine = append(cmdLine, "-e", envVar)
	}

	cmdLine = append(cmdLine, container)
	cmdLine = append(cmdLine, command...)

	logrus.Debugf("executing: docker %s", strings.Join(cmdLine, " "))

	cmd := exec.Command("docker", cmdLine...)

	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	stdout := outbuf.String()
	stderr := errbuf.String()

	return stdout, stderr, err
}