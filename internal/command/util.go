package command

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
)

const (
	// Useful for parsing paths, and lists. Used by several modules.
	sep = "/"

	// Add any opt names here when an empty slice needs to be returned.
	emptySlice = "tf-var env-var managed-identity"

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

	// Handle case that would return a slice of length 1 when zero is expected.
	if strings.Contains(emptySlice, optName) {
		return []string{}
	}

	return []string{defaultValue}
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

// objectToJSON marshals object and returns the result as a string.
func objectToJSON(object interface{}) (string, error) {
	buf, err := json.MarshalIndent(object, "", "    ")
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

// The End.
