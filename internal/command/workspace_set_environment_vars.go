package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/varparser"
)

type workspaceSetEnvironmentVarsCommand struct {
	*BaseCommand

	envVarFiles []string
}

var _ Command = (*workspaceSetEnvironmentVarsCommand)(nil)

// NewWorkspaceSetEnvironmentVarsCommandFactory returns a workspaceSetEnvironmentVarsCommand struct.
func NewWorkspaceSetEnvironmentVarsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceSetEnvironmentVarsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceSetEnvironmentVarsCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: workspace id")
	}

	return nil
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

	workspace, err := c.grpcClient.WorkspacesClient.GetWorkspaceByID(c.Context, &pb.GetWorkspaceByIDRequest{
		Id: trn.ToTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
	})
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
		Category:      pb.VariableCategory_environment,
		Variables:     pbVariables,
	}

	if _, err = c.grpcClient.NamespaceVariablesClient.SetNamespaceVariables(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to set environment variables")
		return 1
	}

	c.UI.Successf("Environment variables set successfully in workspace!")
	return 0
}

func (*workspaceSetEnvironmentVarsCommand) Synopsis() string {
	return "Set environment variables for a workspace."
}

func (*workspaceSetEnvironmentVarsCommand) Description() string {
	return `
   Replaces all environment variables in a workspace from
   a file. Does not support sensitive variables.
`
}

func (*workspaceSetEnvironmentVarsCommand) Usage() string {
	return "tharsis [global options] workspace set-environment-vars [options] <workspace-id>"
}

func (*workspaceSetEnvironmentVarsCommand) Example() string {
	return `
tharsis workspace set-environment-vars \
  -env-var-file "vars.env" \
  trn:workspace:<workspace_path>
`
}

func (c *workspaceSetEnvironmentVarsCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringSliceVar(
		&c.envVarFiles,
		"env-var-file",
		"Path to an environment variables file.",
		flag.Required(),
	)

	return f
}
