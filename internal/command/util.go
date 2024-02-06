package command

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/varparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

const (
	// Useful for parsing paths, and lists. Used by several modules.
	sep = "/"

	// For clearly marking required option in help text.
	red   = "\033[31m"
	reset = "\033[0m"
)

// Utility module for the command package.  These functions are used by
// multiple modules.

// getOption is called by other modules in this package
func getOption(optName string, defaultValue string, options map[string][]string) []string {
	gotVal, ok := options[optName]
	if ok {
		return gotVal
	}

	return []string{defaultValue}
}

// getOptionSlice returns the slice of values for an option that was passed
// in multiple times to a subcommand. This correctly handles the behavior
// where an empty slice is expected when option wasn't provided.
func getOptionSlice(optName string, options map[string][]string) []string {
	gotVal, ok := options[optName]
	if ok {
		return gotVal
	}

	return []string{}
}

// getBoolOptionValue returns a boolean based on a string option argument.
// If the argument is not reconcilable with true/false, it returns an error.
func getBoolOptionValue(optName string, defaultValue string, options map[string][]string) (bool, error) {
	s := getOption(optName, defaultValue, options)
	switch strings.ToLower(s[0]) {
	case "true", "1":
		return true, nil
	case "false", "0":
		return false, nil
	//
	// If other values should be allowed, they can be put here.
	//
	default:
		return false, fmt.Errorf("invalid argument for --%s option: %s", optName, s)
	}
}

// isNamespacePathValid determines if a path is invalid (starts or ends with a '/').
// Logs an error using the Metadata passed in.
func isNamespacePathValid(meta *Metadata, namespacePath string) bool {
	if namespacePath == "" {
		meta.Logger.Error(output.FormatError("namespace path cannot be empty", nil))
		return false
	}
	if strings.HasPrefix(namespacePath, sep) || strings.HasSuffix(namespacePath, sep) {
		meta.Logger.Error(output.FormatError("namespace path must not begin or end with a forward slash", nil))
		return false
	}

	return true
}

// arePathsValid is a helper function to validate resource paths
// that include a parent like managed identity path.
func isResourcePathValid(meta *Metadata, resourcePath string) bool {
	if strings.LastIndex(resourcePath, sep) == -1 {
		meta.Logger.Error(output.FormatError("resource path is not valid", nil))
		return false
	}
	if strings.HasPrefix(resourcePath, sep) || strings.HasSuffix(resourcePath, sep) {
		meta.Logger.Error(output.FormatError("resource path must not begin or end with a forward slash", nil))
		return false
	}

	return true
}

// buildHelpText build the option part of help text for any given Tharsis subcommand.
func buildHelpText(defs optparser.OptionDefinitions) string {
	var buf []string
	buf = append(buf, "\nOptions:")

	sortedNames, longest := sortOptions(defs)

	for _, option := range sortedNames {
		var required string

		// If an option is required display 'Required.'
		if defs[option].Required {
			required = red + "Required" + reset + "."
		}
		pad := strings.Repeat(" ", longest-len(option))
		buf = append(buf, fmt.Sprintf("\n   --%s%s  %s %s", option, pad, defs[option].Synopsis, required))
	}

	return strings.Join(buf, "\n")
}

func sortOptions(defs optparser.OptionDefinitions) ([]string, int) {
	longestName := 0
	sortedNames := []string{}

	// Find the longest option name.
	for optName := range defs {
		if len(optName) > longestName {
			longestName = len(optName)
		}
		sortedNames = append(sortedNames, optName)
	}

	// Sort the option names.
	sort.Strings(sortedNames)

	return sortedNames, longestName
}

// buildJSONOptionDefs returns jsonDefs OptionDefinition appended to provided defs.
func buildJSONOptionDefs(defs optparser.OptionDefinitions) optparser.OptionDefinitions {
	jsonDefs := optparser.OptionDefinition{
		Arguments: []string{},
		Synopsis:  "Show final output as JSON.",
	}
	defs["json"] = &jsonDefs

	return defs
}

// buildPaginationOptionDefs returns pagination defs shared between several commands.
func buildPaginationOptionDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"cursor": {
			Arguments: []string{"Cursor_String"},
			Synopsis:  "The cursor string for manual pagination.",
		},
		"limit": {
			Arguments: []string{"count"},
			Synopsis:  "Maximum number of result elements to return.",
		},
		"sort-order": {
			Arguments: []string{"Sort_Order"},
			Synopsis:  "Sort in this direction, ASC or DESC.",
		},
	}
}

// objectToJSON marshals object and returns the result as a string.
func objectToJSON(object interface{}) (string, error) {
	buf, err := json.MarshalIndent(object, "", "    ")
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func convertToSetNamespaceVariablesInput(vars []varparser.Variable) []types.SetNamespaceVariablesVariable {
	response := []types.SetNamespaceVariablesVariable{}
	for _, v := range vars {
		response = append(response, types.SetNamespaceVariablesVariable{
			Key:   v.Key,
			Value: v.Value,
			HCL:   v.HCL,
		})
	}
	return response
}
