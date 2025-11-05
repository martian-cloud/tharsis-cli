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

// groupDeleteTerraformVarCommand is the top-level structure for the group delete-terraform-var command.
type groupDeleteTerraformVarCommand struct {
	meta *Metadata
}

// NewGroupDeleteTerraformVarCommandFactory returns a groupDeleteTerraformVarCommand struct.
func NewGroupDeleteTerraformVarCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return groupDeleteTerraformVarCommand{
			meta: meta,
		}, nil
	}
}

func (gdv groupDeleteTerraformVarCommand) Run(args []string) int {
	gdv.meta.Logger.Debugf("Starting the 'group delete-terraform-var' command with %d arguments:", len(args))
	for ix, arg := range args {
		gdv.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := gdv.meta.GetSDKClient()
	if err != nil {
		gdv.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return gdv.doGroupDeleteTerraformVar(ctx, client, args)
}

func (gdv groupDeleteTerraformVarCommand) doGroupDeleteTerraformVar(ctx context.Context, client *tharsis.Client, opts []string) int {
	gdv.meta.Logger.Debugf("will do group delete-terraform-var, %d opts", len(opts))

	defs := gdv.buildGroupDeleteTerraformVarDefs()

	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(gdv.meta.BinaryName+" group delete-terraform-var", defs, opts)
	if err != nil {
		gdv.meta.Logger.Error(output.FormatError("failed to parse group delete-terraform-var options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		gdv.meta.Logger.Error(output.FormatError("missing group delete-terraform-var group path", nil), gdv.HelpGroupDeleteTerraformVar())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive group delete-terraform-var arguments: %s", cmdArgs)
		gdv.meta.Logger.Error(output.FormatError(msg, nil), gdv.HelpGroupDeleteTerraformVar())
		return 1
	}

	namespacePath := cmdArgs[0]
	key := getOption("key", "", cmdOpts)[0]

	if key == "" {
		gdv.meta.Logger.Error(output.FormatError("missing required --key option", nil), gdv.HelpGroupDeleteTerraformVar())
		return 1
	}

	actualPath := trn.ToPath(namespacePath)
	if !isNamespacePathValid(gdv.meta, actualPath) {
		return 1
	}

	if _, err = client.Group.GetGroup(ctx, &sdktypes.GetGroupInput{
		Path: &actualPath,  // Use extracted path, not original namespacePath
	}); err != nil {
		gdv.meta.Logger.Error(output.FormatError("failed to get group", err))
		return 1
	}

	input := &sdktypes.DeleteNamespaceVariableInput{
		ID: trn.NewResourceTRN(trn.ResourceTypeVariable, actualPath, string(sdktypes.TerraformVariableCategory), key),  // Use extracted path
	}

	gdv.meta.Logger.Debugf("group delete-terraform-var input: %#v", input)

	err = client.Variable.DeleteVariable(ctx, input)
	if err != nil {
		gdv.meta.Logger.Error(output.FormatError("failed to delete group variable", err))
		return 1
	}

	gdv.meta.UI.Output(fmt.Sprintf("Terraform variable '%s' deleted successfully from group %s", key, namespacePath))
	return 0
}

func (gdv groupDeleteTerraformVarCommand) buildGroupDeleteTerraformVarDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"key": {
			Arguments: []string{"Variable_Key"},
			Synopsis:  "The key/name of the terraform variable to delete.",
			Required:  true,
		},
	}
}

func (gdv groupDeleteTerraformVarCommand) Synopsis() string {
	return "Delete a single terraform variable from a group."
}

func (gdv groupDeleteTerraformVarCommand) Help() string {
	return gdv.HelpGroupDeleteTerraformVar()
}

// HelpGroupDeleteTerraformVar produces the help string for the 'group delete-terraform-var' command.
func (gdv groupDeleteTerraformVarCommand) HelpGroupDeleteTerraformVar() string {
	return fmt.Sprintf(`
Usage: %s [global options] group delete-terraform-var [options] <group>

   The group delete-terraform-var command deletes a single terraform
   variable from a group.

%s

`, gdv.meta.BinaryName, buildHelpText(gdv.buildGroupDeleteTerraformVarDefs()))
}
