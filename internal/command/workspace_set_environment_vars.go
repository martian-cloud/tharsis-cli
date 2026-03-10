package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/varparser"
)

type workspaceSetEnvironmentVarsCommand struct {
	*BaseCommand

	envVarFiles []string
}

// NewWorkspaceSetEnvironmentVarsCommandFactory returns a workspaceSetEnvironmentVarsCommand struct.
func NewWorkspaceSetEnvironmentVarsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceSetEnvironmentVarsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceSetEnvironmentVarsCommand) validate() error {
	const message = "workspace-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.envVarFiles, validation.Required),
	)
}

func (c *workspaceSetEnvironmentVarsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace set-environment-vars"),
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

	parser := varparser.NewVariableParser(nil, false)

	variables, err := parser.ParseEnvironmentVariables(&varparser.ParseEnvironmentVariablesInput{EnvVarFilePaths: c.envVarFiles})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to process environment variables")
		return 1
	}

	pbVariables := make([]*pb.SetNamespaceVariablesInputVariable, len(variables))
	for i, v := range variables {
		pbVariables[i] = &pb.SetNamespaceVariablesInputVariable{
			Key:   v.Key,
			Value: v.Value,
		}
	}

	input := &pb.SetNamespaceVariablesRequest{
		NamespacePath: workspace.FullPath,
		Category:      pb.VariableCategory_ENVIRONMENT,
		Variables:     pbVariables,
	}

	c.Logger.Debug("workspace set-environment-vars input", "input", input)

	if _, err = c.grpcClient.NamespaceVariablesClient.SetNamespaceVariables(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to set environment variables")
		return 1
	}

	c.UI.Successf("Environment variables set successfully in workspace %s", workspace.FullPath)
	return 0
}

func (*workspaceSetEnvironmentVarsCommand) Synopsis() string {
	return "Set environment variables for a workspace."
}

func (*workspaceSetEnvironmentVarsCommand) Description() string {
	return `
   The workspace set-environment-vars command sets environment variables for a workspace.
   Command will overwrite any existing environment variables in the target workspace!
   Note: This command does not support sensitive variables.
`
}

func (*workspaceSetEnvironmentVarsCommand) Usage() string {
	return "tharsis [global options] workspace set-environment-vars [options] <workspace-id>"
}

func (*workspaceSetEnvironmentVarsCommand) Example() string {
	return `
tharsis workspace set-environment-vars \
  --env-var-file vars.env \
  trn:workspace:ops/my-workspace
`
}

func (c *workspaceSetEnvironmentVarsCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"env-var-file",
		"Path to an environment variables file (can be specified multiple times).",
		func(s string) error {
			c.envVarFiles = append(c.envVarFiles, s)
			return nil
		},
	)

	return f
}
