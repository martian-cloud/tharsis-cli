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

// workspaceListTerraformVarsCommand is the top-level structure for the workspace list-terraform-vars command.
type workspaceListTerraformVarsCommand struct {
	meta *Metadata
}

// NewWorkspaceListTerraformVarsCommandFactory returns a workspaceListTerraformVarsCommand struct.
func NewWorkspaceListTerraformVarsCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceListTerraformVarsCommand{
			meta: meta,
		}, nil
	}
}

func (wltv workspaceListTerraformVarsCommand) Run(args []string) int {
	wltv.meta.Logger.Debugf("Starting the 'workspace list-terraform-vars' command with %d arguments:", len(args))
	for ix, arg := range args {
		wltv.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := wltv.meta.GetSDKClient()
	if err != nil {
		wltv.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return wltv.doWorkspaceListTerraformVars(ctx, client, args)
}

func (wltv workspaceListTerraformVarsCommand) doWorkspaceListTerraformVars(ctx context.Context, client *tharsis.Client, opts []string) int {
	wltv.meta.Logger.Debugf("will do workspace list-terraform-vars, %d opts", len(opts))

	defs := wltv.buildWorkspaceListTerraformVarsDefs()

	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wltv.meta.BinaryName+" workspace list-terraform-vars", defs, opts)
	if err != nil {
		wltv.meta.Logger.Error(output.FormatError("failed to parse workspace list-terraform-vars options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		wltv.meta.Logger.Error(output.FormatError("missing workspace list-terraform-vars workspace path", nil), wltv.HelpWorkspaceListTerraformVars())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive workspace list-terraform-vars arguments: %s", cmdArgs)
		wltv.meta.Logger.Error(output.FormatError(msg, nil), wltv.HelpWorkspaceListTerraformVars())
		return 1
	}

	namespacePath := cmdArgs[0]

	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		wltv.meta.UI.Error(output.FormatError("failed to parse boolean value for --json", err))
		return 1
	}

	showSensitive, err := getBoolOptionValue("show-sensitive", "false", cmdOpts)
	if err != nil {
		wltv.meta.UI.Error(output.FormatError("failed to parse boolean value for --show-sensitive", err))
		return 1
	}

	actualPath := trn.ToPath(namespacePath)
	if !isNamespacePathValid(wltv.meta, actualPath) {
		return 1
	}
	// Prepare the inputs - convert path to TRN and use ID field
	trnID := trn.ToTRN(namespacePath, trn.ResourceTypeWorkspace)
	input := &sdktypes.GetWorkspaceInput{ID: &trnID}

	if _, err = client.Workspaces.GetWorkspace(ctx, input); err != nil {
		wltv.meta.Logger.Error(output.FormatError("failed to get workspace", err))
		return 1
	}

	terraformVars, err := listTerraformVariables(ctx, wltv.meta, client, actualPath, showSensitive)
	if err != nil {
		wltv.meta.Logger.Error(output.FormatError("failed to list variables", err))
		return 1
	}

	return outputNamespaceVariables(wltv.meta, toJSON, terraformVars)
}

func (wltv workspaceListTerraformVarsCommand) buildWorkspaceListTerraformVarsDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"show-sensitive": {
			Arguments: []string{},
			Synopsis:  "Show the actual values of sensitive variables (requires appropriate permissions).",
		},
		"json": {
			Arguments: []string{},
			Synopsis:  "Output in JSON format.",
		},
	}
}

func (wltv workspaceListTerraformVarsCommand) Synopsis() string {
	return "List all terraform variables in a workspace."
}

func (wltv workspaceListTerraformVarsCommand) Help() string {
	return wltv.HelpWorkspaceListTerraformVars()
}

// HelpWorkspaceListTerraformVars produces the help string for the 'workspace list-terraform-vars' command.
func (wltv workspaceListTerraformVarsCommand) HelpWorkspaceListTerraformVars() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace list-terraform-vars [options] <workspace>

   The workspace list-terraform-vars command retrieves all terraform
   variables from a workspace and its parent groups.

%s

`, wltv.meta.BinaryName, buildHelpText(wltv.buildWorkspaceListTerraformVarsDefs()))
}
