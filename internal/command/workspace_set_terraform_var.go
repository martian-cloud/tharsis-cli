package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// workspaceSetTerraformVarCommand is the top-level structure for the workspace set-terraform-var command.
type workspaceSetTerraformVarCommand struct {
	meta *Metadata
}

// NewWorkspaceSetTerraformVarCommandFactory returns a workspaceSetTerraformVarCommand struct.
func NewWorkspaceSetTerraformVarCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceSetTerraformVarCommand{
			meta: meta,
		}, nil
	}
}

func (wsv workspaceSetTerraformVarCommand) Run(args []string) int {
	wsv.meta.Logger.Debugf("Starting the 'workspace set-terraform-var' command with %d arguments:", len(args))
	for ix, arg := range args {
		wsv.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := wsv.meta.GetSDKClient()
	if err != nil {
		wsv.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return wsv.doWorkspaceSetTerraformVar(ctx, client, args)
}

func (wsv workspaceSetTerraformVarCommand) doWorkspaceSetTerraformVar(ctx context.Context, client *tharsis.Client, opts []string) int {
	wsv.meta.Logger.Debugf("will do workspace set-terraform-var, %d opts", len(opts))

	defs := wsv.buildWorkspaceSetTerraformVarDefs()

	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wsv.meta.BinaryName+" workspace set-terraform-var", defs, opts)
	if err != nil {
		wsv.meta.Logger.Error(output.FormatError("failed to parse workspace set-terraform-var options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		wsv.meta.Logger.Error(output.FormatError("missing workspace set-terraform-var workspace path", nil), wsv.HelpWorkspaceSetTerraformVar())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive workspace set-terraform-var arguments: %s", cmdArgs)
		wsv.meta.Logger.Error(output.FormatError(msg, nil), wsv.HelpWorkspaceSetTerraformVar())
		return 1
	}

	namespacePath := cmdArgs[0]
	key := getOption("key", "", cmdOpts)[0]
	value := getOption("value", "", cmdOpts)[0]

	if key == "" {
		wsv.meta.Logger.Error(output.FormatError("missing required --key option", nil), wsv.HelpWorkspaceSetTerraformVar())
		return 1
	}
	if value == "" {
		wsv.meta.Logger.Error(output.FormatError("missing required --value option", nil), wsv.HelpWorkspaceSetTerraformVar())
		return 1
	}

	sensitive, err := getBoolOptionValue("sensitive", "false", cmdOpts)
	if err != nil {
		wsv.meta.UI.Error(output.FormatError("failed to parse boolean value for --sensitive", err))
		return 1
	}

	actualPath := trn.ToPath(namespacePath)
	if !isNamespacePathValid(wsv.meta, actualPath) {
		return 1
	}

	if _, err = client.Workspaces.GetWorkspace(ctx, &sdktypes.GetWorkspaceInput{
		Path: &actualPath,  // Use extracted path, not original namespacePath
	}); err != nil {
		wsv.meta.Logger.Error(output.FormatError("failed to get workspace", err))
		return 1
	}

	getInput := &sdktypes.GetNamespaceVariableInput{
		ID: trn.NewResourceTRN(trn.ResourceTypeVariable, namespacePath, string(sdktypes.TerraformVariableCategory), key),
	}

	wsv.meta.Logger.Debugf("workspace set-terraform-var get variable input: %#v", getInput)

	variable, err := client.Variable.GetVariable(ctx, getInput)
	if err != nil && !tharsis.IsNotFoundError(err) {
		wsv.meta.Logger.Error(output.FormatError("failed to get variable", err))
		return 1
	}

	if variable != nil {
		if variable.Sensitive != sensitive {
			wsv.meta.Logger.Error(output.FormatError("cannot change sensitive flag - delete and recreate the variable instead", nil))
			return 1
		}

		updateInput := &sdktypes.UpdateNamespaceVariableInput{
			ID:    variable.Metadata.ID,
			Key:   variable.Key,
			Value: value,
		}

		wsv.meta.Logger.Debugf("workspace set-terraform-var update variable input: %#v", updateInput)

		if _, err = client.Variable.UpdateVariable(ctx, updateInput); err != nil {
			wsv.meta.Logger.Error(output.FormatError("failed to update variable", err))
			return 1
		}
	} else {
		// Extract path from TRN if needed - NamespacePath field expects paths, not TRNs
		actualPath = trn.ToPath(namespacePath)
		
		createInput := &sdktypes.CreateNamespaceVariableInput{
			Key:           key,
			Value:         value,
			Category:      sdktypes.TerraformVariableCategory,
			NamespacePath: actualPath,
			Sensitive:     sensitive,
		}

		wsv.meta.Logger.Debugf("workspace set-terraform-var create variable input: %#v", createInput)

		if _, err = client.Variable.CreateVariable(ctx, createInput); err != nil {
			wsv.meta.Logger.Error(output.FormatError("failed to create variable", err))
			return 1
		}
	}

	wsv.meta.UI.Output(fmt.Sprintf("Terraform variable '%s' set successfully in workspace %s", key, namespacePath))
	return 0
}

func (wsv workspaceSetTerraformVarCommand) buildWorkspaceSetTerraformVarDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"key": {
			Arguments: []string{"Variable_Key"},
			Synopsis:  "The key/name of the terraform variable.",
			Required:  true,
		},
		"value": {
			Arguments: []string{"Variable_Value"},
			Synopsis:  "The value of the terraform variable.",
			Required:  true,
		},
		"sensitive": {
			Arguments: []string{},
			Synopsis:  "Mark the variable as sensitive.",
		},
	}
}

func (wsv workspaceSetTerraformVarCommand) Synopsis() string {
	return "Set a single terraform variable for a workspace."
}

func (wsv workspaceSetTerraformVarCommand) Help() string {
	return wsv.HelpWorkspaceSetTerraformVar()
}

// HelpWorkspaceSetTerraformVar produces the help string for the 'workspace set-terraform-var' command.
func (wsv workspaceSetTerraformVarCommand) HelpWorkspaceSetTerraformVar() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace set-terraform-var [options] <workspace>

   The workspace set-terraform-var command sets a single terraform
   variable for a workspace. If the variable already exists, it will
   be deleted and recreated with the new value. Use the --sensitive
   flag to mark the variable as sensitive.

%s

`, wsv.meta.BinaryName, buildHelpText(wsv.buildWorkspaceSetTerraformVarDefs()))
}
