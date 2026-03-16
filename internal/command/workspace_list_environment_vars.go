package command

import (
	"flag"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type workspaceListEnvironmentVarsCommand struct {
	*BaseCommand

	showSensitive bool
	toJSON        bool
}

// NewWorkspaceListEnvironmentVarsCommandFactory returns a workspaceListEnvironmentVarsCommand struct.
func NewWorkspaceListEnvironmentVarsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceListEnvironmentVarsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceListEnvironmentVarsCommand) validate() error {
	const message = "workspace-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *workspaceListEnvironmentVarsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace list-environment-vars"),
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

	result, err := c.grpcClient.NamespaceVariablesClient.GetNamespaceVariables(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to list environment variables")
		return 1
	}

	// Filter to only environment variables
	var environmentVars []*pb.NamespaceVariable
	for _, v := range result.Variables {
		if v.Category == pb.VariableCategory_environment.String() {
			environmentVars = append(environmentVars, v)
		}
	}

	// Fetch sensitive values if requested
	if c.showSensitive {
		for _, v := range environmentVars {
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

	return outputNamespaceVariables(c.UI, c.toJSON, c.showSensitive, environmentVars)
}

func (*workspaceListEnvironmentVarsCommand) Synopsis() string {
	return "List all environment variables in a workspace."
}

func (*workspaceListEnvironmentVarsCommand) Description() string {
	return `
   The workspace list-environment-vars command retrieves all terraform
   variables from a workspace and its parent workspaces.
`
}

func (*workspaceListEnvironmentVarsCommand) Usage() string {
	return "tharsis [global options] workspace list-environment-vars [options] <workspace-id>"
}

func (*workspaceListEnvironmentVarsCommand) Example() string {
	return `
tharsis workspace list-environment-vars --show-sensitive trn:workspace:<workspace_path>
`
}

func (c *workspaceListEnvironmentVarsCommand) Flags() *flag.FlagSet {
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
