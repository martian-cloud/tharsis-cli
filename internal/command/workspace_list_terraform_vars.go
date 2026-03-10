package command

import (
	"flag"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type workspaceListTerraformVarsCommand struct {
	*BaseCommand

	showSensitive bool
	toJSON        bool
}

// NewWorkspaceListTerraformVarsCommandFactory returns a workspaceListTerraformVarsCommand struct.
func NewWorkspaceListTerraformVarsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceListTerraformVarsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceListTerraformVarsCommand) validate() error {
	const message = "workspace-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *workspaceListTerraformVarsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace list-terraform-vars"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	workspace, err := c.grpcClient.WorkspacesClient.GetWorkspaceByID(c.Context, &pb.GetWorkspaceByIDRequest{Id: c.arguments[0]})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get workspace")
		return 1
	}

	input := &pb.GetNamespaceVariablesRequest{
		NamespacePath: workspace.FullPath,
	}

	c.Logger.Debug("workspace list-terraform-vars input", "input", input)

	result, err := c.grpcClient.NamespaceVariablesClient.GetNamespaceVariables(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to list terraform variables")
		return 1
	}

	// Filter to only terraform variables
	var terraformVars []*pb.NamespaceVariable
	for _, v := range result.Variables {
		if v.Category == "terraform" {
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

func (*workspaceListTerraformVarsCommand) Synopsis() string {
	return "List all terraform variables in a workspace."
}

func (*workspaceListTerraformVarsCommand) Description() string {
	return `
   The workspace list-terraform-vars command retrieves all terraform
   variables from a workspace and its parent workspaces.
`
}

func (*workspaceListTerraformVarsCommand) Usage() string {
	return "tharsis [global options] workspace list-terraform-vars [options] <workspace-id>"
}

func (*workspaceListTerraformVarsCommand) Example() string {
	return `
tharsis workspace list-terraform-vars --show-sensitive trn:workspace:ops/my-workspace
`
}

func (c *workspaceListTerraformVarsCommand) Flags() *flag.FlagSet {
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
