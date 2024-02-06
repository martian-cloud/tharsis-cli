package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/varparser"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// workspaceSetEnvironmentVarsCommand is the top-level structure for the workspace set-environment-vars command.
type workspaceSetEnvironmentVarsCommand struct {
	meta *Metadata
}

// NewWorkspaceSetEnvironmentVarsCommandFactory returns a workspaceSetEnvironmentVarsCommand struct.
func NewWorkspaceSetEnvironmentVarsCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceSetEnvironmentVarsCommand{
			meta: meta,
		}, nil
	}
}

func (wsv workspaceSetEnvironmentVarsCommand) Run(args []string) int {
	wsv.meta.Logger.Debugf("Starting the 'workspace set-environment-vars' command with %d arguments:", len(args))
	for ix, arg := range args {
		wsv.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := wsv.meta.ReadSettings()
	if err != nil {
		wsv.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		wsv.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return wsv.doWorkspaceSetEnvironmentVars(ctx, client, args)
}

func (wsv workspaceSetEnvironmentVarsCommand) doWorkspaceSetEnvironmentVars(ctx context.Context, client *tharsis.Client, opts []string) int {
	wsv.meta.Logger.Debugf("will do workspace set-environment-vars, %d opts", len(opts))

	defs := buildEnvironmentDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wsv.meta.BinaryName+" workspace set-environment-vars", defs, opts)
	if err != nil {
		wsv.meta.Logger.Error(output.FormatError("failed to parse workspace set-environment-vars options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		wsv.meta.Logger.Error(output.FormatError("missing workspace set-environment-vars workspace path", nil), wsv.HelpWorkspaceSetEnvironmentVars())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive workspace set-environment-vars arguments: %s", cmdArgs)
		wsv.meta.Logger.Error(output.FormatError(msg, nil), wsv.HelpWorkspaceSetEnvironmentVars())
		return 1
	}

	namespacePath := cmdArgs[0]
	envVarFiles := getOptionSlice("env-var-file", cmdOpts)

	// Error is already logged.
	if !isNamespacePathValid(wsv.meta, namespacePath) {
		return 1
	}

	parser := varparser.NewVariableParser(nil, false)

	variables, err := parser.ParseEnvironmentVariables(&varparser.ParseEnvironmentVariablesInput{EnvVarFilePaths: envVarFiles})
	if err != nil {
		wsv.meta.Logger.Error(output.FormatError("failed to process environment variables", err))
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.SetNamespaceVariablesInput{
		NamespacePath: namespacePath,
		Category:      sdktypes.EnvironmentVariableCategory,
		Variables:     convertToSetNamespaceVariablesInput(variables),
	}

	wsv.meta.Logger.Debugf("workspace set-environment-vars input: %#v", input)

	// Set the workspace variables.
	err = client.Variable.SetVariables(ctx, input)
	if err != nil {
		wsv.meta.Logger.Error(output.FormatError("failed to set workspace variables", err))
		return 1
	}

	// Format the output.
	wsv.meta.UI.Output(fmt.Sprintf("Environment variables created successfully in workspace %s", namespacePath))
	return 0
}

func (wsv workspaceSetEnvironmentVarsCommand) Synopsis() string {
	return "Set environment variables for a workspace."
}

func (wsv workspaceSetEnvironmentVarsCommand) Help() string {
	return wsv.HelpWorkspaceSetEnvironmentVars()
}

// HelpWorkspaceSetEnvironmentVars produces the help string for the 'workspace set-environment-vars' command.
func (wsv workspaceSetEnvironmentVarsCommand) HelpWorkspaceSetEnvironmentVars() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace set-environment-vars [options] <workspace>

   The workspace set-environment-vars command sets environment
   variables for a workspace. Expects an option with the path
   to the variable file.

   Command will overwrite any existing environment variables
   in the target workspace!

%s

`, wsv.meta.BinaryName, buildHelpText(buildEnvironmentDefs()))
}
