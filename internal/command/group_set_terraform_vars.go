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

// groupSetTerraformVarsCommand is the top-level structure for the group set-terraform-vars command.
type groupSetTerraformVarsCommand struct {
	meta *Metadata
}

// NewGroupSetTerraformVarsCommandFactory returns a groupSetTerraformVarsCommand struct.
func NewGroupSetTerraformVarsCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return groupSetTerraformVarsCommand{
			meta: meta,
		}, nil
	}
}

func (gsv groupSetTerraformVarsCommand) Run(args []string) int {
	gsv.meta.Logger.Debugf("Starting the 'group set-terraform-vars' command with %d arguments:", len(args))
	for ix, arg := range args {
		gsv.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := gsv.meta.ReadSettings()
	if err != nil {
		gsv.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		gsv.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return gsv.doGroupSetTerraformVars(ctx, client, args)
}

func (gsv groupSetTerraformVarsCommand) doGroupSetTerraformVars(ctx context.Context, client *tharsis.Client, opts []string) int {
	gsv.meta.Logger.Debugf("will do group set-terraform-vars, %d opts", len(opts))

	defs := buildTerraformDefs()

	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(gsv.meta.BinaryName+" group set-terraform-vars", defs, opts)
	if err != nil {
		gsv.meta.Logger.Error(output.FormatError("failed to parse group set-terraform-vars options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		gsv.meta.Logger.Error(output.FormatError("missing group set-terraform-vars group path", nil), gsv.HelpGroupSetTerraformVars())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive group set-terraform-vars arguments: %s", cmdArgs)
		gsv.meta.Logger.Error(output.FormatError(msg, nil), gsv.HelpGroupSetTerraformVars())
		return 1
	}

	namespacePath := cmdArgs[0]
	tfVarFile := getOption("tf-var-file", "", cmdOpts)[0]

	// Error is already logged.
	if !isNamespacePathValid(gsv.meta, namespacePath) {
		return 1
	}

	variables, err := varparser.ProcessVariables(varparser.ProcessVariablesInput{TfVarFilePath: tfVarFile})
	if err != nil {
		gsv.meta.Logger.Error(output.FormatError("failed to process terraform variables", err))
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.SetNamespaceVariablesInput{
		NamespacePath: namespacePath,
		Category:      sdktypes.TerraformVariableCategory,
		Variables:     convertToSetNamespaceVariablesInput(variables),
	}

	gsv.meta.Logger.Debugf("group set-terraform-vars input: %#v", input)

	// Set the group variables.
	err = client.Variable.SetVariables(ctx, input)
	if err != nil {
		gsv.meta.Logger.Error(output.FormatError("failed to set group variables", err))
		return 1
	}

	// Format the output.
	gsv.meta.UI.Output(fmt.Sprintf("Terraform variables created successfully in group %s", namespacePath))
	return 0
}

func buildTerraformDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"tf-var-file": {
			Arguments: []string{"Tf_Var_File"},
			Synopsis:  "The path to a .tfvars variables file.",
			Required:  true,
		},
	}
}

func (gsv groupSetTerraformVarsCommand) Synopsis() string {
	return "Set terraform variables for a group."
}

func (gsv groupSetTerraformVarsCommand) Help() string {
	return gsv.HelpGroupSetTerraformVars()
}

// HelpGroupSetTerraformVars produces the help string for the 'group set-terraform-vars' command.
func (gsv groupSetTerraformVarsCommand) HelpGroupSetTerraformVars() string {
	return fmt.Sprintf(`
Usage: %s [global options] group set-terraform-vars [options] <group>

   The group set-terraform-vars command sets terraform
   variables for a group. Expects an option with the path
   to the .tfvars file.

   Command will overwrite any existing Terraform variables
   in the target group!

%s

`, gsv.meta.BinaryName, buildHelpText(buildTerraformDefs()))
}

// The End.
