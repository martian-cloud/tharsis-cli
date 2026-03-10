package command

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/external"
)

// stateOutputValue represents a state version output value
// and some of its attributes. Necessary for building a map
// that can be marshalled to JSON and displayed easily.
type stateOutputValue struct {
	Type      json.RawMessage `json:"type"`
	Value     json.RawMessage `json:"value"`
	Sensitive bool            `json:"sensitive"`
}

type workspaceOutputsCommand struct {
	*BaseCommand

	outputName string
	raw        bool
	toJSON     bool
}

// NewWorkspaceOutputsCommandFactory returns a workspaceOutputsCommand struct.
func NewWorkspaceOutputsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceOutputsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceOutputsCommand) validate() error {
	const message = "workspace-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.raw,
			validation.When(c.toJSON, validation.Empty.Error("must not supply both -raw and -json")),
			validation.When(c.outputName == "", validation.Empty.Error("must specify -output-name if specifying -raw")),
		),
	)
}

func (c *workspaceOutputsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace outputs"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	workspace, err := c.grpcClient.WorkspacesClient.GetWorkspaceByID(c.Context, &pb.GetWorkspaceByIDRequest{
		Id: c.arguments[0],
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get workspace")
		return 1
	}

	if workspace.CurrentStateVersionId == "" {
		c.UI.Output("workspace does not have a current state version")
		return 1
	}

	result, err := c.grpcClient.StateVersionsClient.GetStateVersionOutputs(c.Context, &pb.GetStateVersionOutputsRequest{
		StateVersionId: workspace.CurrentStateVersionId,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get state version outputs")
		return 1
	}

	if len(result.StateVersionOutputs) == 0 {
		c.UI.Output("workspace does not have any state version outputs")
		return 1
	}

	return c.displayWorkspaceOutput(result.StateVersionOutputs)
}

func (c *workspaceOutputsCommand) displayWorkspaceOutput(outputs []*pb.StateVersionOutput) int {
	valueMap := make(map[string]*stateOutputValue, len(outputs))

	for _, output := range outputs {
		valueMap[output.Name] = &stateOutputValue{
			Value:     output.Value,
			Type:      output.Type,
			Sensitive: output.Sensitive,
		}
	}

	var (
		val any = valueMap
		ok  bool
	)

	if c.outputName != "" {
		val, ok = valueMap[c.outputName]
		if !ok {
			c.UI.Errorf("%s does not exist in state version output. Name is case sensitive.", c.outputName)
			return 1
		}
	}

	if c.toJSON {
		if err := c.UI.JSON(val); err != nil {
			c.UI.ErrorWithSummary(err, "failed to output JSON")
			return 1
		}
		return 0
	}

	sort.SliceStable(outputs, func(i, j int) bool {
		return outputs[i].Name < outputs[j].Name
	})

	for _, v := range outputs {
		ctyType, err := ctyjson.UnmarshalType(v.Type)
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to unmarshal type")
			return 1
		}

		ctyValue, err := ctyjson.Unmarshal(v.Value, ctyType)
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to unmarshal value")
			return 1
		}

		valueFormatted := external.FormatValue(ctyValue, 0)

		if c.outputName == "" {
			if v.Sensitive {
				valueFormatted = "<sensitive>"
			}
			c.UI.Output(fmt.Sprintf("%s = %s", v.Name, valueFormatted))
		} else if c.raw {
			if v.Name != c.outputName {
				continue
			}

			valueString, err := convert.Convert(ctyValue, cty.String)
			if err != nil {
				c.UI.Errorf("-raw is only supported on string, number and boolean types: %s is of type '%s'. Use -json flag for more complex types", c.outputName, ctyType.FriendlyName())
				return 1
			}

			if valueString.IsNull() {
				c.UI.Errorf("Unsupported value type: value for %s is null", c.outputName)
				return 1
			}

			fmt.Fprint(os.Stdout, valueString.AsString())
		} else if v.Name == c.outputName {
			c.UI.Output(valueFormatted)
		}
	}

	return 0
}

func (*workspaceOutputsCommand) Synopsis() string {
	return "Get the state version outputs for a workspace."
}

func (*workspaceOutputsCommand) Description() string {
	return `
   The workspace outputs command retrieves the state version outputs for a workspace.

   Supported output types:
      - Decorated (shows if map, list, etc. default).
      - JSON.
      - Raw (just the value. limited).

   In addition, it supports filtering the output for each of the supported types above with --output-name option.

   Combining --raw and --json is not allowed.
`
}

func (*workspaceOutputsCommand) Usage() string {
	return "tharsis [global options] workspace outputs [options] <workspace-id>"
}

func (*workspaceOutputsCommand) Example() string {
	return `
tharsis workspace outputs trn:workspace:ops/my-workspace
`
}

func (c *workspaceOutputsCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.outputName,
		"output-name",
		"",
		"The name of the output variable to use as a filter. Required for -raw option.",
	)
	f.BoolVar(
		&c.raw,
		"raw",
		false,
		"For any value that can be converted to a string, output just the raw value.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Output in JSON format.",
	)

	return f
}
