package command

import (
	"flag"
	"fmt"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type groupGetTerraformVarCommand struct {
	*BaseCommand

	key           string
	showSensitive bool
	toJSON        bool
}

// NewGroupGetTerraformVarCommandFactory returns a groupGetTerraformVarCommand struct.
func NewGroupGetTerraformVarCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupGetTerraformVarCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupGetTerraformVarCommand) validate() error {
	const message = "group-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.key, validation.Required),
	)
}

func (c *groupGetTerraformVarCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group get-terraform-var"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	// Get group to retrieve full path
	group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: toTRN(trn.ResourceTypeGroup, c.arguments[0])})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get group")
		return 1
	}

	input := &pb.GetNamespaceVariableByIDRequest{
		Id: trn.NewResourceTRN(trn.ResourceTypeVariable, group.FullPath, pb.VariableCategory_terraform.String(), c.key),
	}

	c.Logger.Debug("group get-terraform-var input", "input", input)

	variable, err := c.grpcClient.NamespaceVariablesClient.GetNamespaceVariableByID(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get terraform variable")
		return 1
	}

	// If showing sensitive value, fetch the variable version
	if c.showSensitive && variable.Sensitive {
		versionInput := &pb.GetNamespaceVariableVersionByIDRequest{
			Id:                    variable.LatestVersionId,
			IncludeSensitiveValue: true,
		}

		version, err := c.grpcClient.NamespaceVariablesClient.GetNamespaceVariableVersionByID(c.Context, versionInput)
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to get variable version")
			return 1
		}

		// Set the value from the version
		variable.Value = version.Value
	}

	return outputNamespaceVariable(c.UI, c.toJSON, c.showSensitive, variable)
}

func (*groupGetTerraformVarCommand) Synopsis() string {
	return "Get a terraform variable for a group."
}

func (*groupGetTerraformVarCommand) Description() string {
	return `
   The group get-terraform-var command retrieves a terraform variable for a group.
`
}

func (*groupGetTerraformVarCommand) Usage() string {
	return "tharsis [global options] group get-terraform-var [options] <group-id>"
}

func (*groupGetTerraformVarCommand) Example() string {
	return `
tharsis group get-terraform-var \
  --key region \
  trn:group:<group_path>
`
}

func (c *groupGetTerraformVarCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.key,
		"key",
		"",
		"Variable key.",
	)
	f.BoolVar(
		&c.showSensitive,
		"show-sensitive",
		false,
		"Show the actual value of sensitive variables (requires appropriate permissions).",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Output in JSON format.",
	)

	return f
}

func outputNamespaceVariable(ui terminal.UI, toJSON bool, showSensitive bool, variable *pb.NamespaceVariable) int {
	if toJSON {
		if err := ui.JSON(variable); err != nil {
			ui.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		displayValue := ptr.ToString(variable.Value)
		if variable.Sensitive && !showSensitive {
			displayValue = "[SENSITIVE]"
		}

		t := terminal.NewTable("key", "value", "namespace_path", "sensitive")
		t.Rich([]string{
			variable.Key,
			displayValue,
			variable.NamespacePath,
			fmt.Sprintf("%t", variable.Sensitive),
		}, nil)

		ui.Table(t)
	}

	return 0
}
