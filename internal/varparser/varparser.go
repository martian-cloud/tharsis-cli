// Package varparser contains the logic for parsing
// Terraform and environment variables that Tharsis
// API supports. It supports parsing variables
// passed in via flags and from files.
package varparser

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/log"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// ProcessVariablesInput defines the input for ProcessVariables.
type ProcessVariablesInput struct {
	TfVarFilePath  string
	EnvVarFilePath string
	TfVariables    []string
	EnvVariables   []string
}

// Variable represents a parsed terraform or environment variable.
type Variable struct {
	Value    string
	Key      string
	Category sdktypes.VariableCategory
	HCL      bool
}

// ProcessVariables dispatches the functions to process variables files or variable string and returns the result.
func ProcessVariables(input ProcessVariablesInput) ([]Variable, error) {
	var (
		variables []Variable
		err       error
	)

	// Use variable arguments if file path is not specified.
	if input.TfVarFilePath == "" && input.EnvVarFilePath == "" {
		if len(input.TfVariables) > 0 {
			variables, err = processVariables(input.TfVariables, sdktypes.TerraformVariableCategory)
			if err != nil {
				return nil, err
			}
		}
		if len(input.EnvVariables) > 0 {
			// Append to variables slice incase it is not empty.
			result, pErr := processVariables(input.EnvVariables, sdktypes.EnvironmentVariableCategory)
			if pErr != nil {
				return nil, pErr
			}

			variables = append(variables, result...)
		}
	}

	if input.TfVarFilePath != "" {
		variables, err = processTfVarsFile(input.TfVarFilePath)
		if err != nil {
			return nil, err
		}
	}

	if input.EnvVarFilePath != "" {
		vars, err := processEnvVarsFile(input.EnvVarFilePath)
		if err != nil {
			return nil, err
		}
		variables = append(variables, vars...)
	}

	return variables, nil
}

// processVariables iterates through the variables slice and splits variables
// using "=". Populates a slice of RunVariable and returns the result.
func processVariables(variables []string, category sdktypes.VariableCategory) ([]Variable, error) {
	var runVariables []Variable

	seen := map[string]int{} // Keep track of duplicates.
	for i, pair := range variables {
		pair = strings.TrimSpace(pair)

		if pair == "" {
			// Skip empty lines or variables.
			continue
		}

		s := strings.SplitN(pair, "=", 2)

		if len(s) != 2 {
			return nil, fmt.Errorf("%s variable is not a key=value pair at position %d", category, i+1)
		}

		key := strings.TrimSpace(s[0])
		val := strings.TrimSpace(s[1]) // Value must be a pointer.

		// Make sure there is a key and value pair, output a helpful error otherwise.
		// Assumes that a value could be empty.
		if key == "" || strings.Contains(key, " ") {
			return nil, fmt.Errorf("%s variable has an invalid key at position %d", category, i+1)
		}

		// Log a warning if a variable is being mistakenly overwritten.
		if position, ok := seen[key]; ok {
			log.Warnf("%s variable '%s' was previously declared at position %d", category, key, position)
		}

		// Add the variable and its position, so we can track it.
		seen[key] = i + 1

		// Append variable to slice.
		runVariables = append(runVariables, Variable{
			Key:      key,
			Value:    val,
			HCL:      false, // Set HCL to false for variable passed in via an argument.
			Category: category,
		})
	}

	return runVariables, nil
}

// processTfVarsFile parses a .tfvars file and returns a slice of type RunVariable.
func processTfVarsFile(filePath string) ([]Variable, error) {
	if !strings.HasSuffix(filePath, ".tfvars") {
		return nil, errors.New("filename must end in .tfvars")
	}

	parser := hclparse.NewParser()

	// Parse the given file
	file, diags := parser.ParseHCLFile(filePath)
	if diags.HasErrors() {
		return nil, fmt.Errorf("%s", diags.Error())
	}

	// Get only the attributes.
	attributes, diags := file.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, fmt.Errorf("%s", diags.Error())
	}

	// Get the values for each attribute and create run variables.
	var runVariables []Variable
	for key, attr := range attributes {
		value, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return nil, fmt.Errorf("%s", diags.Error())
		}

		bytes, err := ctyjson.Marshal(value, value.Type())
		if err != nil {
			return nil, fmt.Errorf("%s", diags.Error())
		}

		val := string(json.RawMessage(bytes))

		// Remove quotes around string values.
		if value.Type().Equals(cty.String) {
			val = value.AsString()
		}

		// Append variable to slice.
		runVariables = append(runVariables, Variable{
			Key:      key,
			Category: sdktypes.TerraformVariableCategory,
			HCL:      !value.Type().Equals(cty.String), // Set HCL if value is not a string type (complex variable).
			Value:    val,
		})
	}

	return runVariables, nil
}

// processEnvVarsFile reads a environment variables file
// and calls processVariables to return a slice of type RunVariable.
func processEnvVarsFile(filePath string) ([]Variable, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Read file line by line into lines slice.
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return processVariables(lines, sdktypes.EnvironmentVariableCategory)
}
