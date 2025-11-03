package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/mitchellh/cli"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/external"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// stateOutputValue represents a state version output value
// and some of its attributes. Necessary for building a map
// that can be marshalled to JSON and displayed easily.
type stateOutputValue struct {
	Type      json.RawMessage `json:"type"`
	Value     json.RawMessage `json:"value"`
	Sensitive bool            `json:"sensitive"`
}

// workspaceOutputsCommand is the top-level structure for the workspace outputs command.
type workspaceOutputsCommand struct {
	meta *Metadata
}

// NewWorkspaceOutputsCommandFactory returns a workspaceOutputsCommand struct.
func NewWorkspaceOutputsCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceOutputsCommand{
			meta: meta,
		}, nil
	}
}

func (wo workspaceOutputsCommand) Run(args []string) int {
	wo.meta.Logger.Debugf("Starting the 'workspace outputs' command with %d arguments:", len(args))
	for ix, arg := range args {
		wo.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := wo.meta.GetSDKClient()
	if err != nil {
		wo.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return wo.doWorkspaceOutputs(ctx, client, args)
}

func (wo workspaceOutputsCommand) doWorkspaceOutputs(ctx context.Context, client *tharsis.Client, opts []string) int {
	wo.meta.Logger.Debugf("will do workspace outputs, %d opts", len(opts))

	defs := wo.buildWorkspaceOutputsDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wo.meta.BinaryName+" workspace outputs", defs, opts)
	if err != nil {
		wo.meta.Logger.Error(output.FormatError("failed to parse workspace outputs options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		wo.meta.Logger.Error(output.FormatError("missing workspace outputs full path", nil), wo.HelpWorkspaceOutputs())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive workspace outputs arguments: %s", cmdArgs)
		wo.meta.Logger.Error(output.FormatError(msg, nil), wo.HelpWorkspaceOutputs())
		return 1
	}

	workspacePath := cmdArgs[0]
	raw, err := getBoolOptionValue("raw", "false", cmdOpts)
	if err != nil {
		wo.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		wo.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	outputName := getOption("output-name", "", cmdOpts)[0]

	// Validate the workspace path.
	if !isNamespacePathValid(wo.meta, workspacePath) {
		return 1
	}

	// Cannot show both raw and json formats simultaneously.
	if raw && toJSON {
		wo.meta.Logger.Error(output.FormatError("must not supply both -raw and -json", nil))
		return 1
	}

	// Must supply outputName when using raw option. Optional otherwise.
	if raw && outputName == "" {
		wo.meta.Logger.Error(output.FormatError("must specify -output-name if specifying -raw", nil))
		return 1
	}

	input := &sdktypes.GetWorkspaceInput{Path: &workspacePath}
	wo.meta.Logger.Debugf("workspace outputs input: %#v", input)

	workspace, err := client.Workspaces.GetWorkspace(ctx, input)
	if err != nil {
		wo.meta.Logger.Error(output.FormatError("failed to get a workspace", err))
		return 1
	}

	wo.meta.Logger.Debugf("workspace outputs found workspace: %s", workspace.FullPath)

	// Check if the workspace has a current state version.
	if workspace.CurrentStateVersion == nil {
		wo.meta.Logger.Error(output.FormatError("workspace does not have a current state version", nil))
		return 1
	}

	outputs := workspace.CurrentStateVersion.Outputs

	// Check if there are any outputs.
	if len(outputs) == 0 {
		wo.meta.Logger.Error(output.FormatError("workspace does not have any state version outputs", nil))
		return 1
	}

	return wo.displayWorkspaceOutput(raw, toJSON, outputName, outputs)
}

// displayWorkspaceOutput is a helper function to modify
// the final output based on the options.
func (wo workspaceOutputsCommand) displayWorkspaceOutput(raw, toJSON bool, outputName string,
	outputs []sdktypes.StateVersionOutput,
) int {
	outputMap, err := wo.buildStateOutputValueMap(outputs)
	if err != nil {
		wo.meta.Logger.Error(output.FormatError("failed to build state version output value map", err))
		return 1
	}

	// Used for the JSON display.
	var (
		val interface{} = outputMap
		ok  bool
	)

	// Check if the output name exists in the state version output.
	if outputName != "" {
		val, ok = outputMap[outputName]
		if !ok {
			msg := fmt.Sprintf("%s does not exist in state version output. Name is case sensitive.", outputName)
			wo.meta.Logger.Error(output.FormatError(msg, nil))
			return 1
		}
	}

	if toJSON {
		buf, err := objectToJSON(val)
		if err != nil {
			wo.meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}

		// Show the output.
		wo.meta.UI.Output(string(buf))
	} else {
		// Sort the slice to give an alphabetized output.
		sort.SliceStable(outputs, func(i, j int) bool {
			return outputs[i].Name < outputs[j].Name
		})

		// Must use original slice as FormatValue requires cty.Value.
		for _, v := range outputs {
			valueFormatted := external.FormatValue(v.Value, 0)

			// Regular output.
			if outputName == "" {
				if v.Sensitive {
					valueFormatted = "<sensitive>" // Obfuscate output if the value is marked as sensitive.
				}

				wo.meta.UI.Output(fmt.Sprintf("%s = %s", v.Name, valueFormatted))
			} else if raw {
				// Keep going until the right output name is found.
				if v.Name != outputName {
					continue
				}

				// Check if raw is being called on a supported type.
				valueString, err := convert.Convert(v.Value, cty.String)
				if err != nil {
					err := fmt.Errorf("%s is of type '%s'. Use -json flag for more complex types", outputName, v.Type.FriendlyName())
					wo.meta.Logger.Error(output.FormatError("-raw is only supported on string, number and boolean types", err))
					return 1
				}

				if valueString.IsNull() {
					err := fmt.Errorf("value for %s is null", outputName)
					wo.meta.Logger.Error(output.FormatError("Unsupported value type", err))
					return 1
				}

				// This will print without a newline.
				fmt.Fprint(os.Stdout, valueString.AsString())
			} else if v.Name == outputName {
				wo.meta.UI.Output(valueFormatted)
			}

		}
	}

	return 0
}

// buildStateOutputValueMap build a map of the outputs name
// to it's value and its attributes. Used for displaying a
// subset of the returned StateVersionOutput values and for
// marshalling the cty.Values into their appropriate types.
func (wo workspaceOutputsCommand) buildStateOutputValueMap(outputs []sdktypes.StateVersionOutput) (map[string]*stateOutputValue, error) {
	valueMap := make(map[string]*stateOutputValue, len(outputs))

	// Build a map of output name --> stateOutputValue.
	for _, output := range outputs {
		value, err := ctyjson.Marshal(output.Value, output.Type)
		if err != nil {
			return nil, err
		}

		valueType, err := ctyjson.MarshalType(output.Type)
		if err != nil {
			return nil, err
		}

		// Assign to the map.
		valueMap[output.Name] = &stateOutputValue{
			Value:     value,
			Type:      valueType,
			Sensitive: output.Sensitive,
		}
	}

	return valueMap, nil
}

// buildWorkspaceOutputsDefs returns defs used by workspace outputs command.
func (wo workspaceOutputsCommand) buildWorkspaceOutputsDefs() optparser.OptionDefinitions {
	rawDefs := optparser.OptionDefinitions{
		"output-name": {
			Arguments: []string{"Output_Name"},
			Synopsis:  "The name of the output variable to use as a filter. Required for -raw option.",
		},
		"raw": {
			Arguments: []string{},
			Synopsis:  "For any value that can be converted to a string, output just the raw value.",
		},
	}

	return buildJSONOptionDefs(rawDefs)
}

func (wo workspaceOutputsCommand) Synopsis() string {
	return "Get the state version outputs for a workspace."
}

func (wo workspaceOutputsCommand) Help() string {
	return wo.HelpWorkspaceOutputs()
}

// HelpWorkspaceOutputs produces the help string for the 'workspace outputs' command.
func (wo workspaceOutputsCommand) HelpWorkspaceOutputs() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace outputs [options] <full_path>

   The workspace outputs command retrieves the state version
   outputs for a workspace.

   Supported output types:
      - Decorated (shows if map, list, etc. default).
      - JSON.
      - Raw (just the value. limited).

   In addition, it supports filtering the output for each
   of the supported types above with --output-name option.

%s


Combining --raw and --json is not allowed.
`, wo.meta.BinaryName, buildHelpText(wo.buildWorkspaceOutputsDefs()))
}
