package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// workspaceDeleteTerraformVarCommand is the top-level structure for the workspace delete-terraform-var command.
type workspaceDeleteTerraformVarCommand struct {
	meta *Metadata
}

// NewWorkspaceDeleteTerraformVarCommandFactory returns a workspaceDeleteTerraformVarCommand struct.
func NewWorkspaceDeleteTerraformVarCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceDeleteTerraformVarCommand{
			meta: meta,
		}, nil
	}
}

func (wdv workspaceDeleteTerraformVarCommand) Run(args []string) int {
	wdv.meta.Logger.Debugf("Starting the 'workspace delete-terraform-var' command with %d arguments:", len(args))
	for ix, arg := range args {
		wdv.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := wdv.meta.GetSDKClient()
	if err != nil {
		wdv.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return wdv.doWorkspaceDeleteTerraformVar(ctx, client, args)
}

func (wdv workspaceDeleteTerraformVarCommand) doWorkspaceDeleteTerraformVar(ctx context.Context, client *tharsis.Client, opts []string) int {
	wdv.meta.Logger.Debugf("will do workspace delete-terraform-var, %d opts", len(opts))

	defs := wdv.buildWorkspaceDeleteTerraformVarDefs()

	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wdv.meta.BinaryName+" workspace delete-terraform-var", defs, opts)
	if err != nil {
		wdv.meta.Logger.Error(output.FormatError("failed to parse workspace delete-terraform-var options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		wdv.meta.Logger.Error(output.FormatError("missing workspace delete-terraform-var workspace path", nil), wdv.HelpWorkspaceDeleteTerraformVar())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive workspace delete-terraform-var arguments: %s", cmdArgs)
		wdv.meta.Logger.Error(output.FormatError(msg, nil), wdv.HelpWorkspaceDeleteTerraformVar())
		return 1
	}

	namespacePath := cmdArgs[0]
	key := getOption("key", "", cmdOpts)[0]

	if key == "" {
		wdv.meta.Logger.Error(output.FormatError("missing required --key option", nil), wdv.HelpWorkspaceDeleteTerraformVar())
		return 1
	}

	actualPath := trn.ToPath(namespacePath)
	if !isNamespacePathValid(wdv.meta, actualPath) {
		return 1
	}

	if _, err = client.Workspaces.GetWorkspace(ctx, &sdktypes.GetWorkspaceInput{
		Path: &actualPath, // Use extracted path, not original namespacePath
	}); err != nil {
		wdv.meta.Logger.Error(output.FormatError("failed to get workspace", err))
		return 1
	}

	input := &sdktypes.DeleteNamespaceVariableInput{
		ID: trn.NewResourceTRN(trn.ResourceTypeVariable, namespacePath, string(sdktypes.TerraformVariableCategory), key),
	}

	wdv.meta.Logger.Debugf("workspace delete-terraform-var input: %#v", input)

	err = client.Variable.DeleteVariable(ctx, input)
	if err != nil {
		wdv.meta.Logger.Error(output.FormatError("failed to delete workspace variable", err))
		return 1
	}

	wdv.meta.UI.Output(fmt.Sprintf("Terraform variable '%s' deleted successfully from workspace %s", key, namespacePath))
	return 0
}

func (wdv workspaceDeleteTerraformVarCommand) buildWorkspaceDeleteTerraformVarDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"key": {
			Arguments: []string{"Variable_Key"},
			Synopsis:  "The key/name of the terraform variable to delete.",
			Required:  true,
		},
	}
}

func (wdv workspaceDeleteTerraformVarCommand) Synopsis() string {
	return "Delete a single terraform variable from a workspace."
}

func (wdv workspaceDeleteTerraformVarCommand) Help() string {
	return wdv.HelpWorkspaceDeleteTerraformVar()
}

// HelpWorkspaceDeleteTerraformVar produces the help string for the 'workspace delete-terraform-var' command.
func (wdv workspaceDeleteTerraformVarCommand) HelpWorkspaceDeleteTerraformVar() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace delete-terraform-var [options] <workspace>

   The workspace delete-terraform-var command deletes a single terraform
   variable from a workspace.

%s

`, wdv.meta.BinaryName, buildHelpText(wdv.buildWorkspaceDeleteTerraformVarDefs()))
}
