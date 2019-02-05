package jsonschema

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/xeipuuv/gojsonschema"
)

// Generate generates JSON schema from the input argument using
// json-schema-generator (https://www.npmjs.com/package/json-schema-generator).
func Generate(input string) ([]byte, error) {
	schema, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("echo '%s' | json-schema-generator --stdin | sed -e 1d", input)).Output()
	if err != nil {
		return nil, fmt.Errorf("Error generating JSON schema, error: %v", err)
	}
	return schema, nil
}

// Validate validates the input argument against JSON schema. If the
// input is not valid the error is returned. The first argument is the file name
// of the JSON schema. It is used to build file URI required to load the JSON schema.
// The second argument is the input string that is validated.
func Validate(fileName string, input string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	schemaURI := fmt.Sprintf("file://%s", filepath.Join(pwd, fileName))

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

// ValidationField is a struct used in JSON schema
type ValidationField struct {
	Keyword      string
	KeywordValue interface{}
}

// AddNewElements adds new fields to the JSON schema under specified location
func AddNewElements(location *simplejson.Json, newFields map[string]ValidationField) error {
	var elemLocation *simplejson.Json
	var ok bool
	for name, valueJSON := range newFields {
		if elemLocation, ok = location.CheckGet(name); !ok {
			return fmt.Errorf("Cannot update JSON schema with value: %s  for element: %s", valueJSON.Keyword, name)
		}
		elemLocation.Set(valueJSON.Keyword, valueJSON.KeywordValue)
	}
	return nil
}