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

// workspaceGetTerraformVarCommand is the top-level structure for the workspace get-terraform-var command.
type workspaceGetTerraformVarCommand struct {
	meta *Metadata
}

// NewWorkspaceGetTerraformVarCommandFactory returns a workspaceGetTerraformVarCommand struct.
func NewWorkspaceGetTerraformVarCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceGetTerraformVarCommand{
			meta: meta,
		}, nil
	}
}

func (wgv workspaceGetTerraformVarCommand) Run(args []string) int {
	wgv.meta.Logger.Debugf("Starting the 'workspace get-terraform-var' command with %d arguments:", len(args))
	for ix, arg := range args {
		wgv.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := wgv.meta.GetSDKClient()
	if err != nil {
		wgv.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return wgv.doWorkspaceGetTerraformVar(ctx, client, args)
}

func (wgv workspaceGetTerraformVarCommand) doWorkspaceGetTerraformVar(ctx context.Context, client *tharsis.Client, opts []string) int {
	wgv.meta.Logger.Debugf("will do workspace get-terraform-var, %d opts", len(opts))

	defs := wgv.buildWorkspaceGetTerraformVarDefs()

	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wgv.meta.BinaryName+" workspace get-terraform-var", defs, opts)
	if err != nil {
		wgv.meta.Logger.Error(output.FormatError("failed to parse workspace get-terraform-var options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		wgv.meta.Logger.Error(output.FormatError("missing workspace get-terraform-var workspace path", nil), wgv.HelpWorkspaceGetTerraformVar())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive workspace get-terraform-var arguments: %s", cmdArgs)
		wgv.meta.Logger.Error(output.FormatError(msg, nil), wgv.HelpWorkspaceGetTerraformVar())
		return 1
	}

	namespacePath := cmdArgs[0]
	key := getOption("key", "", cmdOpts)[0]

	if key == "" {
		wgv.meta.Logger.Error(output.FormatError("missing required --key option", nil), wgv.HelpWorkspaceGetTerraformVar())
		return 1
	}

	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		wgv.meta.UI.Error(output.FormatError("failed to parse boolean value for --json", err))
		return 1
	}

	showSensitive, err := getBoolOptionValue("show-sensitive", "false", cmdOpts)
	if err != nil {
		wgv.meta.UI.Error(output.FormatError("failed to parse boolean value for --show-sensitive", err))
		return 1
	}

	actualPath := trn.ToPath(namespacePath)
	if !isNamespacePathValid(wgv.meta, actualPath) {
		return 1
	}

	// Prepare the inputs - convert path to TRN and use ID field
	trnID := trn.ToTRN(namespacePath, trn.ResourceTypeWorkspace)
	input := &sdktypes.GetWorkspaceInput{ID: &trnID}

	if _, err = client.Workspaces.GetWorkspace(ctx, input); err != nil {
		wgv.meta.Logger.Error(output.FormatError("failed to get workspace", err))
		return 1
	}

	variable, err := getTerraformVariable(ctx, wgv.meta, client, namespacePath, key, showSensitive)
	if err != nil {
		wgv.meta.Logger.Error(output.FormatError("failed to get variable", err))
		return 1
	}

	return outputNamespaceVariable(wgv.meta, toJSON, variable)
}

func (wgv workspaceGetTerraformVarCommand) buildWorkspaceGetTerraformVarDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"key": {
			Arguments: []string{"Variable_Key"},
			Synopsis:  "The key/name of the terraform variable.",
			Required:  true,
		},
		"show-sensitive": {
			Arguments: []string{},
			Synopsis:  "Show the actual value of sensitive variables (requires appropriate permissions).",
		},
		"json": {
			Arguments: []string{},
			Synopsis:  "Output in JSON format.",
		},
	}
}

func (wgv workspaceGetTerraformVarCommand) Synopsis() string {
	return "Get a single terraform variable from a workspace."
}

func (wgv workspaceGetTerraformVarCommand) Help() string {
	return wgv.HelpWorkspaceGetTerraformVar()
}

// HelpWorkspaceGetTerraformVar produces the help string for the 'workspace get-terraform-var' command.
func (wgv workspaceGetTerraformVarCommand) HelpWorkspaceGetTerraformVar() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace get-terraform-var [options] <workspace>

   The workspace get-terraform-var command retrieves a single terraform
   variable from a workspace by its key.

%s

`, wgv.meta.BinaryName, buildHelpText(wgv.buildWorkspaceGetTerraformVarDefs()))
}
