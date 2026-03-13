package command

import (
	"flag"
	"fmt"
	"time"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type groupListTerraformVarsCommand struct {
	*BaseCommand

	showSensitive bool
	toJSON        bool
}

// NewGroupListTerraformVarsCommandFactory returns a groupListTerraformVarsCommand struct.
func NewGroupListTerraformVarsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupListTerraformVarsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupListTerraformVarsCommand) validate() error {
	const message = "group-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *groupListTerraformVarsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group list-terraform-vars"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: toTRN(trn.ResourceTypeGroup, c.arguments[0])})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get group")
		return 1
	}

	input := &pb.GetNamespaceVariablesRequest{
		NamespacePath: group.FullPath,
	}

	c.Logger.Debug("group list-terraform-vars input", "input", input)

	result, err := c.grpcClient.NamespaceVariablesClient.GetNamespaceVariables(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to list terraform variables")
		return 1
	}

	// Filter to only terraform variables
	var terraformVars []*pb.NamespaceVariable
	for _, v := range result.Variables {
		if v.Category == pb.VariableCategory_terraform.String() {
			terraformVars = append(terraformVars, v)
		}
	}

	// Fetch sensitive values if requested
	if c.showSensitive {
		for _, v := range terraformVars {
			if v.Sensitive && v.LatestVersionId != "" {
				versionInput := &pb.GetNamespaceVariableVersionByIDRequest{
					Id:                    v.LatestVersionId,
					IncludeSensitiveValue: true,
				}

				version, err := c.grpcClient.NamespaceVariablesClient.GetNamespaceVariableVersionByID(c.Context, versionInput)
				if err != nil {
					c.UI.ErrorWithSummary(err, "failed to get variable version")
					return 1
				}

				v.Value = version.Value
				// Rate limit to avoid overwhelming the API
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	return outputNamespaceVariables(c.UI, c.toJSON, c.showSensitive, terraformVars)
}

func (*groupListTerraformVarsCommand) Synopsis() string {
	return "List all terraform variables in a group."
}

func (*groupListTerraformVarsCommand) Description() string {
	return `
   The group list-terraform-vars command retrieves all terraform
   variables from a group and its parent groups.
`
}

func (*groupListTerraformVarsCommand) Usage() string {
	return "tharsis [global options] group list-terraform-vars [options] <group-id>"
}

func (*groupListTerraformVarsCommand) Example() string {
	return `
tharsis group list-terraform-vars --show-sensitive trn:group:<group_path>
`
}

func (c *groupListTerraformVarsCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.showSensitive,
		"show-sensitive",
		false,
		"Show the actual values of sensitive variables (requires appropriate permissions).",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Output in JSON format.",
	)

	return f
}

func outputNamespaceVariables(ui terminal.UI, toJSON bool, showSensitive bool, variables []*pb.NamespaceVariable) int {
	if toJSON {
		if err := ui.JSON(variables); err != nil {
			ui.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
		return 0
	}

	if len(variables) == 0 {
		ui.Output("No variables found")
		return 0
	}

	t := terminal.NewTable("key", "value", "namespace_path", "sensitive")
	for _, v := range variables {
		displayValue := ptr.ToString(v.Value)
		if v.Sensitive && !showSensitive {
			displayValue = "[SENSITIVE]"
		}

		t.Rich([]string{
			v.Key,
			displayValue,
			v.NamespacePath,
			fmt.Sprintf("%t", v.Sensitive),
		}, nil)
	}

	ui.Table(t)
	return 0
}
