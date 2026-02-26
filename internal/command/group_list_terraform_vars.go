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

// groupListTerraformVarsCommand is the top-level structure for the group list-terraform-vars command.
type groupListTerraformVarsCommand struct {
	meta *Metadata
}

// NewGroupListTerraformVarsCommandFactory returns a groupListTerraformVarsCommand struct.
func NewGroupListTerraformVarsCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return groupListTerraformVarsCommand{
			meta: meta,
		}, nil
	}
}

func (gltv groupListTerraformVarsCommand) Run(args []string) int {
	gltv.meta.Logger.Debugf("Starting the 'group list-terraform-vars' command with %d arguments:", len(args))
	for ix, arg := range args {
		gltv.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := gltv.meta.GetSDKClient()
	if err != nil {
		gltv.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return gltv.doGroupListTerraformVars(ctx, client, args)
}

func (gltv groupListTerraformVarsCommand) doGroupListTerraformVars(ctx context.Context, client *tharsis.Client, opts []string) int {
	gltv.meta.Logger.Debugf("will do group list-terraform-vars, %d opts", len(opts))

	defs := gltv.buildGroupListTerraformVarsDefs()

	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(gltv.meta.BinaryName+" group list-terraform-vars", defs, opts)
	if err != nil {
		gltv.meta.Logger.Error(output.FormatError("failed to parse group list-terraform-vars options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		gltv.meta.Logger.Error(output.FormatError("missing group list-terraform-vars group path", nil), gltv.HelpGroupListTerraformVars())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive group list-terraform-vars arguments: %s", cmdArgs)
		gltv.meta.Logger.Error(output.FormatError(msg, nil), gltv.HelpGroupListTerraformVars())
		return 1
	}

	namespacePath := cmdArgs[0]

	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		gltv.meta.UI.Error(output.FormatError("failed to parse boolean value for --json", err))
		return 1
	}

	showSensitive, err := getBoolOptionValue("show-sensitive", "false", cmdOpts)
	if err != nil {
		gltv.meta.UI.Error(output.FormatError("failed to parse boolean value for --show-sensitive", err))
		return 1
	}

	actualPath := trn.ToPath(namespacePath)
	if !isNamespacePathValid(gltv.meta, actualPath) {
		return 1
	}

	if _, err = client.Group.GetGroup(ctx, &sdktypes.GetGroupInput{
		Path: &actualPath,
	}); err != nil {
		gltv.meta.Logger.Error(output.FormatError("failed to get group", err))
		return 1
	}

	terraformVars, err := listTerraformVariables(ctx, gltv.meta, client, actualPath, showSensitive)
	if err != nil {
		gltv.meta.Logger.Error(output.FormatError("failed to list variables", err))
		return 1
	}

	return outputNamespaceVariables(gltv.meta, toJSON, terraformVars)
}

func (gltv groupListTerraformVarsCommand) buildGroupListTerraformVarsDefs() optparser.OptionDefinitions {
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

func (gltv groupListTerraformVarsCommand) Synopsis() string {
	return "List all terraform variables in a group."
}

func (gltv groupListTerraformVarsCommand) Help() string {
	return gltv.HelpGroupListTerraformVars()
}

// HelpGroupListTerraformVars produces the help string for the 'group list-terraform-vars' command.
func (gltv groupListTerraformVarsCommand) HelpGroupListTerraformVars() string {
	return fmt.Sprintf(`
Usage: %s [global options] group list-terraform-vars [options] <group>

   The group list-terraform-vars command retrieves all terraform
   variables from a group and its parent groups.

%s

`, gltv.meta.BinaryName, buildHelpText(gltv.buildGroupListTerraformVarsDefs()))
}
