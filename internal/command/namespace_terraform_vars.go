package command

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tableformatter"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// getTerraformVariable retrieves a single terraform variable from a namespace.
func getTerraformVariable(
	ctx context.Context,
	meta *Metadata,
	client *tharsis.Client,
	namespacePath string,
	key string,
	showSensitive bool,
) (*sdktypes.NamespaceVariable, error) {
	getInput := &sdktypes.GetNamespaceVariableInput{
		ID: trn.NewResourceTRN(trn.ResourceTypeVariable, trn.ToPath(namespacePath), string(sdktypes.TerraformVariableCategory), key),
	}

	meta.Logger.Debugf("get terraform variable input: %#v", getInput)

	variable, err := client.Variable.GetVariable(ctx, getInput)
	if err != nil {
		return nil, err
	}

	// Fetch sensitive value if requested
	if showSensitive && variable.Sensitive {
		sensitiveValue, err := fetchSensitiveValue(ctx, client, variable)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch sensitive value: %w", err)
		}
		variable.Value = sensitiveValue
		// Rate limiting is handled inside fetchSensitiveValue
	}

	return variable, nil
}

// listTerraformVariables retrieves all terraform variables from a namespace.
func listTerraformVariables(
	ctx context.Context,
	_ *Metadata,
	client *tharsis.Client,
	namespacePath string,
	showSensitive bool,
) ([]sdktypes.NamespaceVariable, error) {
	// Get all variables for the namespace
	variables, err := client.Variable.GetVariables(ctx, &sdktypes.GetNamespaceVariablesInput{
		NamespacePath: namespacePath,
	})
	if err != nil {
		return nil, err
	}

	// Filter to only terraform variables
	terraformVars := []sdktypes.NamespaceVariable{}
	for _, v := range variables {
		if v.Category == sdktypes.TerraformVariableCategory {
			terraformVars = append(terraformVars, v)
		}
	}

	// Fetch sensitive values if requested
	if showSensitive {
		for i := range terraformVars {
			if terraformVars[i].Sensitive {
				sensitiveValue, err := fetchSensitiveValue(ctx, client, &terraformVars[i])
				if err != nil {
					return nil, fmt.Errorf("failed to fetch sensitive value: %w", err)
				}
				terraformVars[i].Value = sensitiveValue
				// Rate limiting: sleep for 100ms between sensitive value fetches
				time.Sleep(100 * time.Millisecond)

			}
		}
	}

	// Sort variables by namespace path (descending) then by key (ascending)
	sort.Slice(terraformVars, func(i, j int) bool {
		cmp := strings.Compare(terraformVars[j].NamespacePath, terraformVars[i].NamespacePath)
		if cmp == 0 {
			return strings.Compare(terraformVars[i].Key, terraformVars[j].Key) < 0
		}
		return cmp < 0
	})

	return terraformVars, nil
}

// outputNamespaceVariable is the final output for namespace variable get operations.
func outputNamespaceVariable(meta *Metadata, toJSON bool, variable *sdktypes.NamespaceVariable) int {
	if toJSON {
		buf, err := objectToJSON(variable)
		if err != nil {
			meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		meta.UI.Output(string(buf))
	} else {
		// Display value or mask if sensitive (unless value was explicitly retrieved)
		displayValue := "[SENSITIVE]"
		if variable.Value != nil {
			displayValue = *variable.Value
		}

		tableInput := [][]string{
			{"key", "value", "category", "namespace path", "sensitive"},
			{variable.Key, displayValue, string(variable.Category), variable.NamespacePath, fmt.Sprintf("%t", variable.Sensitive)},
		}
		meta.UI.Output(tableformatter.FormatTable(tableInput))
	}

	return 0
}

// outputNamespaceVariables is the final output for namespace variable list operations.
func outputNamespaceVariables(meta *Metadata, toJSON bool, variables []sdktypes.NamespaceVariable) int {
	if toJSON {
		buf, err := objectToJSON(variables)
		if err != nil {
			meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		meta.UI.Output(string(buf))
	} else {
		if len(variables) == 0 {
			meta.UI.Output("No terraform variables found.")
			return 0
		}

		// Create table with headers
		tableInput := make([][]string, len(variables)+1)
		tableInput[0] = []string{"key", "value", "category", "namespace path", "sensitive"}

		for ix, variable := range variables {
			// Display value or mask if sensitive (unless value was explicitly retrieved)
			displayValue := "[SENSITIVE]"
			if variable.Value != nil {
				displayValue = *variable.Value
			}

			tableInput[ix+1] = []string{
				variable.Key,
				displayValue,
				string(variable.Category),
				variable.NamespacePath,
				fmt.Sprintf("%t", variable.Sensitive),
			}
		}

		meta.UI.Output(tableformatter.FormatTable(tableInput))
	}

	return 0
}

// fetchSensitiveValue retrieves the sensitive value for a variable.
// Returns the sensitive value if available, otherwise returns nil.
func fetchSensitiveValue(ctx context.Context, client *tharsis.Client, variable *types.NamespaceVariable) (*string, error) {
	if !variable.Sensitive || variable.LatestVersionID == "" {
		return nil, nil
	}

	version, err := client.Variable.GetVariableVersion(ctx, &types.GetVariableVersionInput{
		VersionID:             variable.LatestVersionID,
		IncludeSensitiveValue: true,
	})
	if err != nil {
		return nil, err
	}

	if version != nil && version.Value != nil {
		return version.Value, nil
	}

	return nil, nil
}
