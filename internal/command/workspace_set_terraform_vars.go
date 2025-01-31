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

// workspaceSetTerraformVarsCommand is the top-level structure for the workspace set-terraform-vars command.
type workspaceSetTerraformVarsCommand struct {
	meta *Metadata
}

// NewWorkspaceSetTerraformVarsCommandFactory returns a workspaceSetTerraformVarsCommand struct.
func NewWorkspaceSetTerraformVarsCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceSetTerraformVarsCommand{
			meta: meta,
		}, nil
	}
}

func (wsv workspaceSetTerraformVarsCommand) Run(args []string) int {
	wsv.meta.Logger.Debugf("Starting the 'workspace set-terraform-vars' command with %d arguments:", len(args))
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

	return wsv.doWorkspaceSetTerraformVars(ctx, client, args)
}

func (wsv workspaceSetTerraformVarsCommand) doWorkspaceSetTerraformVars(ctx context.Context, client *tharsis.Client, opts []string) int {
	wsv.meta.Logger.Debugf("will do workspace set-terraform-vars, %d opts", len(opts))

	workspaceTerraformDefs := buildTerraformDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wsv.meta.BinaryName+" workspace set-terraform-vars", workspaceTerraformDefs, opts)
	if err != nil {
		wsv.meta.Logger.Error(output.FormatError("failed to parse workspace set-terraform-vars options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		wsv.meta.Logger.Error(output.FormatError("missing workspace set-terraform-vars workspace path", nil), wsv.HelpWorkspaceSetTerraformVars())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive workspace set-terraform-vars arguments: %s", cmdArgs)
		wsv.meta.Logger.Error(output.FormatError(msg, nil), wsv.HelpWorkspaceSetTerraformVars())
		return 1
	}

	namespacePath := cmdArgs[0]
	tfVarFiles := getOptionSlice("tf-var-file", cmdOpts)

	// Error is already logged.
	if !isNamespacePathValid(wsv.meta, namespacePath) {
		return 1
	}

	// Ensure namespace is a workspace.
	if _, err = client.Workspaces.GetWorkspace(ctx, &sdktypes.GetWorkspaceInput{
		Path: &namespacePath,
	}); err != nil {
		wsv.meta.Logger.Error(output.FormatError("failed to get workspace", err))
		return 1
	}

	parser := varparser.NewVariableParser(nil, false)

	variables, err := parser.ParseTerraformVariables(&varparser.ParseTerraformVariablesInput{TfVarFilePaths: tfVarFiles})
	if err != nil {
		wsv.meta.Logger.Error(output.FormatError("failed to process environment variables", err))
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.SetNamespaceVariablesInput{
		NamespacePath: namespacePath,
		Category:      sdktypes.TerraformVariableCategory,
		Variables:     convertToSetNamespaceVariablesInput(variables),
	}

	wsv.meta.Logger.Debugf("workspace set-terraform-vars input: %#v", input)

	// Set the workspace variables.
	err = client.Variable.SetVariables(ctx, input)
	if err != nil {
		wsv.meta.Logger.Error(output.FormatError("failed to set workspace variables", err))
		return 1
	}

	// Format the output.
	wsv.meta.UI.Output(fmt.Sprintf("Terraform variables created successfully in workspace %s", namespacePath))
	return 0
}

func (wsv workspaceSetTerraformVarsCommand) Synopsis() string {
	return "Set terraform variables for a workspace."
}

func (wsv workspaceSetTerraformVarsCommand) Help() string {
	return wsv.HelpWorkspaceSetTerraformVars()
}

// HelpWorkspaceSetTerraformVars produces the help string for the 'workspace set-terraform-vars' command.
func (wsv workspaceSetTerraformVarsCommand) HelpWorkspaceSetTerraformVars() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace set-terraform-vars [options] <workspace>

   The workspace set-terraform-vars command sets Terraform
   variables for a workspace. Expects an option with the
   path to the .tfvars file.

   Command will overwrite any existing Terraform variables
   in the target workspace!

%s

`, wsv.meta.BinaryName, buildHelpText(buildTerraformDefs()))
}
