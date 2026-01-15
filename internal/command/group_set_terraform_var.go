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

// groupSetTerraformVarCommand is the top-level structure for the group set-terraform-var command.
type groupSetTerraformVarCommand struct {
	meta *Metadata
}

// NewGroupSetTerraformVarCommandFactory returns a groupSetTerraformVarCommand struct.
func NewGroupSetTerraformVarCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return groupSetTerraformVarCommand{
			meta: meta,
		}, nil
	}
}

func (gsv groupSetTerraformVarCommand) Run(args []string) int {
	gsv.meta.Logger.Debugf("Starting the 'group set-terraform-var' command with %d arguments:", len(args))
	for ix, arg := range args {
		gsv.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := gsv.meta.GetSDKClient()
	if err != nil {
		gsv.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return gsv.doGroupSetTerraformVar(ctx, client, args)
}

func (gsv groupSetTerraformVarCommand) doGroupSetTerraformVar(ctx context.Context, client *tharsis.Client, opts []string) int {
	gsv.meta.Logger.Debugf("will do group set-terraform-var, %d opts", len(opts))

	defs := gsv.buildGroupSetTerraformVarDefs()

	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(gsv.meta.BinaryName+" group set-terraform-var", defs, opts)
	if err != nil {
		gsv.meta.Logger.Error(output.FormatError("failed to parse group set-terraform-var options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		gsv.meta.Logger.Error(output.FormatError("missing group set-terraform-var group path", nil), gsv.HelpGroupSetTerraformVar())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive group set-terraform-var arguments: %s", cmdArgs)
		gsv.meta.Logger.Error(output.FormatError(msg, nil), gsv.HelpGroupSetTerraformVar())
		return 1
	}

	namespacePath := cmdArgs[0]
	key := getOption("key", "", cmdOpts)[0]
	value := getOption("value", "", cmdOpts)[0]

	if key == "" {
		gsv.meta.Logger.Error(output.FormatError("missing required --key option", nil), gsv.HelpGroupSetTerraformVar())
		return 1
	}
	if value == "" {
		gsv.meta.Logger.Error(output.FormatError("missing required --value option", nil), gsv.HelpGroupSetTerraformVar())
		return 1
	}

	sensitive, err := getBoolOptionValue("sensitive", "false", cmdOpts)
	if err != nil {
		gsv.meta.UI.Error(output.FormatError("failed to parse boolean value for --sensitive", err))
		return 1
	}

	actualPath := trn.ToPath(namespacePath)
	if !isNamespacePathValid(gsv.meta, actualPath) {
		return 1
	}

	if _, err = client.Group.GetGroup(ctx, &sdktypes.GetGroupInput{
		Path: &actualPath, // Use extracted path, not original namespacePath
	}); err != nil {
		gsv.meta.Logger.Error(output.FormatError("failed to get group", err))
		return 1
	}

	getInput := &sdktypes.GetNamespaceVariableInput{
		ID: trn.NewResourceTRN(trn.ResourceTypeVariable, namespacePath, string(sdktypes.TerraformVariableCategory), key),
	}

	gsv.meta.Logger.Debugf("group set-terraform-var get variable input: %#v", getInput)

	variable, err := client.Variable.GetVariable(ctx, getInput)
	if err != nil && !tharsis.IsNotFoundError(err) {
		gsv.meta.Logger.Error(output.FormatError("failed to get variable", err))
		return 1
	}

	if variable != nil {
		if variable.Sensitive != sensitive {
			gsv.meta.Logger.Error(output.FormatError("cannot change sensitive flag - delete and recreate the variable instead", nil))
			return 1
		}

		updateInput := &sdktypes.UpdateNamespaceVariableInput{
			ID:    variable.Metadata.ID,
			Key:   variable.Key,
			Value: value,
		}

		gsv.meta.Logger.Debugf("group set-terraform-var update variable input: %#v", updateInput)

		if _, err = client.Variable.UpdateVariable(ctx, updateInput); err != nil {
			gsv.meta.Logger.Error(output.FormatError("failed to update variable", err))
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

		gsv.meta.Logger.Debugf("group set-terraform-var create variable input: %#v", createInput)

		if _, err = client.Variable.CreateVariable(ctx, createInput); err != nil {
			gsv.meta.Logger.Error(output.FormatError("failed to create variable", err))
			return 1
		}
	}

	gsv.meta.UI.Output(fmt.Sprintf("Terraform variable '%s' set successfully in group %s", key, namespacePath))
	return 0
}

func (gsv groupSetTerraformVarCommand) buildGroupSetTerraformVarDefs() optparser.OptionDefinitions {
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

func (gsv groupSetTerraformVarCommand) Synopsis() string {
	return "Set a single terraform variable for a group."
}

func (gsv groupSetTerraformVarCommand) Help() string {
	return gsv.HelpGroupSetTerraformVar()
}

// HelpGroupSetTerraformVar produces the help string for the 'group set-terraform-var' command.
func (gsv groupSetTerraformVarCommand) HelpGroupSetTerraformVar() string {
	return fmt.Sprintf(`
Usage: %s [global options] group set-terraform-var [options] <group>

   The group set-terraform-var command sets a single terraform
   variable for a group. If the variable already exists, it will
   be deleted and recreated with the new value. Use the --sensitive
   flag to mark the variable as sensitive.

%s

`, gsv.meta.BinaryName, buildHelpText(gsv.buildGroupSetTerraformVarDefs()))
}
