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

// groupGetTerraformVarCommand is the top-level structure for the group get-terraform-var command.
type groupGetTerraformVarCommand struct {
	meta *Metadata
}

// NewGroupGetTerraformVarCommandFactory returns a groupGetTerraformVarCommand struct.
func NewGroupGetTerraformVarCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return groupGetTerraformVarCommand{
			meta: meta,
		}, nil
	}
}

func (ggv groupGetTerraformVarCommand) Run(args []string) int {
	ggv.meta.Logger.Debugf("Starting the 'group get-terraform-var' command with %d arguments:", len(args))
	for ix, arg := range args {
		ggv.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := ggv.meta.GetSDKClient()
	if err != nil {
		ggv.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return ggv.doGroupGetTerraformVar(ctx, client, args)
}

func (ggv groupGetTerraformVarCommand) doGroupGetTerraformVar(ctx context.Context, client *tharsis.Client, opts []string) int {
	ggv.meta.Logger.Debugf("will do group get-terraform-var, %d opts", len(opts))

	defs := ggv.buildGroupGetTerraformVarDefs()

	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(ggv.meta.BinaryName+" group get-terraform-var", defs, opts)
	if err != nil {
		ggv.meta.Logger.Error(output.FormatError("failed to parse group get-terraform-var options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		ggv.meta.Logger.Error(output.FormatError("missing group get-terraform-var group path", nil), ggv.HelpGroupGetTerraformVar())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive group get-terraform-var arguments: %s", cmdArgs)
		ggv.meta.Logger.Error(output.FormatError(msg, nil), ggv.HelpGroupGetTerraformVar())
		return 1
	}

	namespacePath := cmdArgs[0]
	key := getOption("key", "", cmdOpts)[0]

	if key == "" {
		ggv.meta.Logger.Error(output.FormatError("missing required --key option", nil), ggv.HelpGroupGetTerraformVar())
		return 1
	}

	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		ggv.meta.UI.Error(output.FormatError("failed to parse boolean value for --json", err))
		return 1
	}

	showSensitive, err := getBoolOptionValue("show-sensitive", "false", cmdOpts)
	if err != nil {
		ggv.meta.UI.Error(output.FormatError("failed to parse boolean value for --show-sensitive", err))
		return 1
	}

	actualPath := trn.ToPath(namespacePath)
	if !isNamespacePathValid(ggv.meta, actualPath) {
		return 1
	}

	if _, err = client.Group.GetGroup(ctx, &sdktypes.GetGroupInput{
		Path: &actualPath,
	}); err != nil {
		ggv.meta.Logger.Error(output.FormatError("failed to get group", err))
		return 1
	}

	variable, err := getTerraformVariable(ctx, ggv.meta, client, namespacePath, key, showSensitive)
	if err != nil {
		ggv.meta.Logger.Error(output.FormatError("failed to get variable", err))
		return 1
	}

	return outputNamespaceVariable(ggv.meta, toJSON, variable)
}

func (ggv groupGetTerraformVarCommand) buildGroupGetTerraformVarDefs() optparser.OptionDefinitions {
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

func (ggv groupGetTerraformVarCommand) Synopsis() string {
	return "Get a single terraform variable from a group."
}

func (ggv groupGetTerraformVarCommand) Help() string {
	return ggv.HelpGroupGetTerraformVar()
}

// HelpGroupGetTerraformVar produces the help string for the 'group get-terraform-var' command.
func (ggv groupGetTerraformVarCommand) HelpGroupGetTerraformVar() string {
	return fmt.Sprintf(`
Usage: %s [global options] group get-terraform-var [options] <group>

   The group get-terraform-var command retrieves a single terraform
   variable from a group by its key.

%s

`, ggv.meta.BinaryName, buildHelpText(ggv.buildGroupGetTerraformVarDefs()))
}
