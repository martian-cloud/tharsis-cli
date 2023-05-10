// Package varparser contains the logic for parsing
// Terraform and environment variables that Tharsis
// API supports. It supports parsing variables
// passed in via flags and from files.
package varparser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// Constants required for this module.
const (
	tfvarsExtension             = ".tfvars"
	tfvarsJSONExtension         = ".tfvars.json"
	autoTfvarsExtension         = ".auto" + tfvarsExtension
	autoTfvarsJSONExtension     = ".auto" + tfvarsJSONExtension
	terraformTfvarsFilename     = "terraform" + tfvarsExtension
	terraformTfvarsJSONFilename = "terraform" + tfvarsJSONExtension
	exportedTfVarPrefix         = "TF_VAR_"
)

// ParseTerraformVariablesInput defines the input for ProcessTerraformVariables.
type ParseTerraformVariablesInput struct {
	TfVarFilePaths []string
	TfVariables    []string
}

// ParseEnvironmentVariablesInput defines the input for ProcessEnvironmentVariables.
type ParseEnvironmentVariablesInput struct {
	EnvVarFilePaths []string
	EnvVariables    []string
}

// Variable represents a parsed terraform or environment variable.
type Variable struct {
	Value    string
	Key      string
	Category sdktypes.VariableCategory
	HCL      bool
}

// VariableParser implements functionalities needed to parse variables.
type VariableParser struct {
	variableMap               map[string]Variable
	moduleDirectory           *string
	withTfVarsFromEnvironment bool
}

// NewVariableParser returns a new VariableProcessor.
func NewVariableParser(moduleDirectory *string, withTfVarsFromEnvironment bool) *VariableParser {
	return &VariableParser{
		moduleDirectory:           moduleDirectory,
		withTfVarsFromEnvironment: withTfVarsFromEnvironment,
	}
}

// ParseTerraformVariables dispatches the functions to parse Terraform
// variables and returns a unique slice of parsed Variables.
//
// Parsing precedence:
// 1. Terraform variables from the environment.
// 2. terraform.tfvars file, if present.
// 3. terraform.tfvars.json file, if present.
// 4. *.auto.tfvars.* files, if present.
// 5. --tf-var-file option(s).
// 6. --tf-var option(s).
func (v *VariableParser) ParseTerraformVariables(input *ParseTerraformVariablesInput) ([]Variable, error) {
	v.variableMap = map[string]Variable{} // Initialize the variable map.

	// Parse exported terraform variables if required.
	if v.withTfVarsFromEnvironment {
		exportedVariables := []string{}
		for _, e := range os.Environ() {
			if strings.HasPrefix(e, exportedTfVarPrefix) {
				exportedVariables = append(exportedVariables, strings.TrimPrefix(e, exportedTfVarPrefix))
			}
		}
		err := v.processStringVariables(exportedVariables, sdktypes.TerraformVariableCategory)
		if err != nil {
			return nil, err
		}
	}

	if err := v.processTerraformTfvarsFiles(); err != nil {
		return nil, err
	}

	if err := v.processAutoTfvarsFiles(); err != nil {
		return nil, err
	}

	if err := v.processTfVarsFile(input.TfVarFilePaths); err != nil {
		return nil, err
	}

	err := v.processStringVariables(input.TfVariables, sdktypes.TerraformVariableCategory)
	if err != nil {
		return nil, err
	}

	variables := []Variable{}
	for _, v := range v.variableMap {
		variables = append(variables, v)
	}

	return variables, nil
}

// ParseEnvironmentVariables dispatches functions to parse environment
// variables and returns a unique slice of parsed Variables.
//
// Parsing precedence:
// 1. --env-var-file option(s).
// 2. --env-var option(s).
func (v *VariableParser) ParseEnvironmentVariables(input *ParseEnvironmentVariablesInput) ([]Variable, error) {
	v.variableMap = map[string]Variable{} // Initialize the variable map.

	if err := v.processEnvVarsFile(input.EnvVarFilePaths); err != nil {
		return nil, err
	}

	err := v.processStringVariables(input.EnvVariables, sdktypes.EnvironmentVariableCategory)
	if err != nil {
		return nil, err
	}

	variables := []Variable{}
	for _, v := range v.variableMap {
		variables = append(variables, v)
	}

	return variables, nil
}

// processStringVariables iterates through the variables slice and splits variables using "=".
func (v *VariableParser) processStringVariables(variables []string, category sdktypes.VariableCategory) error {
	for i, pair := range variables {
		pair = strings.TrimSpace(pair)

		if pair == "" {
			// Skip empty lines or variables.
			continue
		}

		s := strings.SplitN(pair, "=", 2)

		if len(s) != 2 {
			return fmt.Errorf("%s variable is not a key=value pair at position %d", category, i+1)
		}

		key := strings.TrimSpace(s[0])
		val := strings.TrimSpace(s[1]) // Value must be a pointer.

		// Make sure there is a key and value pair, output a helpful error otherwise.
		// Assumes that a value could be empty.
		if key == "" || strings.Contains(key, " ") {
			return fmt.Errorf("%s variable has an invalid key at position %d", category, i+1)
		}

		v.variableMap[key] = Variable{
			Key:      key,
			Value:    val,
			HCL:      false, // Set HCL to false for variable passed in via an argument.
			Category: category,
		}
	}

	return nil
}

// processTfVarsFile parses a .tfvars file.
func (v *VariableParser) processTfVarsFile(filePaths []string) error {
	for _, path := range filePaths {
		parser := hclparse.NewParser()

		var (
			parsedFile *hcl.File
			diags      hcl.Diagnostics
		)

		// Call the respective parser based on the file extension.
		if strings.HasSuffix(path, tfvarsExtension) {
			parsedFile, diags = parser.ParseHCLFile(path)
			if diags.HasErrors() {
				return fmt.Errorf("%s", diags.Error())
			}
		} else if strings.HasSuffix(path, tfvarsJSONExtension) {
			parsedFile, diags = parser.ParseJSONFile(path)
			if diags.HasErrors() {
				return fmt.Errorf("%s", diags.Error())
			}
		} else {
			return fmt.Errorf("file extension must only be either %s or %s", tfvarsExtension, tfvarsJSONExtension)
		}

		// Get only the attributes.
		attributes, diags := parsedFile.Body.JustAttributes()
		if diags.HasErrors() {
			return fmt.Errorf("%s", diags.Error())
		}

		// Get the values for each attribute and create run variables.
		for key, attr := range attributes {
			value, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				return fmt.Errorf("%s", diags.Error())
			}

			bytes, err := ctyjson.Marshal(value, value.Type())
			if err != nil {
				return fmt.Errorf("%s", diags.Error())
			}

			val := string(json.RawMessage(bytes))

			// Remove quotes around string values.
			if value.Type().Equals(cty.String) {
				val = value.AsString()
			}

			v.variableMap[key] = Variable{
				Key:      key,
				Category: sdktypes.TerraformVariableCategory,
				HCL:      !value.Type().Equals(cty.String), // Set HCL if value is not a string type (complex variable).
				Value:    val,
			}
		}
	}

	return nil
}

// processEnvVarsFile reads a environment variables file
// and calls processVariables to store Variables in result map.
func (v *VariableParser) processEnvVarsFile(filePaths []string) error {
	for _, path := range filePaths {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		// Read file line by line into lines slice.
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err = v.processStringVariables(lines, sdktypes.EnvironmentVariableCategory); err != nil {
			return err
		}
	}

	return nil
}

// processAutoTfvarsFiles lists all the files in the module's directory
// and processes the variables from the *.auto.tfvars or *.auto.tfvars.json
func (v *VariableParser) processAutoTfvarsFiles() error {
	if v.moduleDirectory == nil {
		// Nothing to process.
		return nil
	}

	files, err := os.ReadDir(*v.moduleDirectory)
	if err != nil {
		return err
	}

	// Find all the files with the "*.auto.tfvars.*" extensions.
	allFiles := []string{}
	for _, file := range files {
		if !file.IsDir() {
			if strings.HasSuffix(file.Name(), autoTfvarsExtension) ||
				strings.HasSuffix(file.Name(), autoTfvarsJSONExtension) {
				pathToFile := filepath.Join(*v.moduleDirectory, file.Name())
				allFiles = append(allFiles, pathToFile)
			}
		}
	}

	return v.processTfVarsFile(allFiles)
}

// processTerraformTfvarsFiles strictly looks for "terraform.tfvars" and
// "terraform.tfvars.json" files in the module's directory. Calls
// processTfVarsFile to process the file(s) that exist.
func (v *VariableParser) processTerraformTfvarsFiles() error {
	if v.moduleDirectory == nil {
		// Nothing to process.
		return nil
	}

	terraformTfVarsFilepath := filepath.Join(*v.moduleDirectory, terraformTfvarsFilename)
	terraformTfVarsJSONFilepath := filepath.Join(*v.moduleDirectory, terraformTfvarsJSONFilename)

	// Determine which file exists and therefore we have to process.
	filesToProcess := []string{}
	for _, path := range []string{terraformTfVarsFilepath, terraformTfVarsJSONFilepath} {
		if _, err := os.Stat(path); err == nil {
			filesToProcess = append(filesToProcess, path)
		}
	}

	return v.processTfVarsFile(filesToProcess)
}
