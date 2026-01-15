package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/varparser"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// groupSetEnvironmentVarsCommand is the top-level structure for the group set-environment-vars command.
type groupSetEnvironmentVarsCommand struct {
	meta *Metadata
}

// NewGroupSetEnvironmentVarsCommandFactory returns a groupSetEnvironmentVarsCommand struct.
func NewGroupSetEnvironmentVarsCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return groupSetEnvironmentVarsCommand{
			meta: meta,
		}, nil
	}
}

func (gsv groupSetEnvironmentVarsCommand) Run(args []string) int {
	gsv.meta.Logger.Debugf("Starting the 'group set-environment-vars' command with %d arguments:", len(args))
	for ix, arg := range args {
		gsv.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := gsv.meta.GetSDKClient()
	if err != nil {
		gsv.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return gsv.doGroupSetEnvironmentVars(ctx, client, args)
}

func (gsv groupSetEnvironmentVarsCommand) doGroupSetEnvironmentVars(ctx context.Context, client *tharsis.Client, opts []string) int {
	gsv.meta.Logger.Debugf("will do group set-environment-vars, %d opts", len(opts))

	defs := buildEnvironmentDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(gsv.meta.BinaryName+" group set-environment-vars", defs, opts)
	if err != nil {
		gsv.meta.Logger.Error(output.FormatError("failed to parse group set-environment-vars options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		gsv.meta.Logger.Error(output.FormatError("missing group set-environment-vars group path", nil), gsv.HelpGroupSetEnvironmentVars())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive group set-environment-vars arguments: %s", cmdArgs)
		gsv.meta.Logger.Error(output.FormatError(msg, nil), gsv.HelpGroupSetEnvironmentVars())
		return 1
	}

	namespacePath := cmdArgs[0]
	envVarFiles := getOptionSlice("env-var-file", cmdOpts)

	// Extract path from TRN if needed, then validate path (error is already logged by validation function)
	actualPath := trn.ToPath(namespacePath)
	if !isNamespacePathValid(gsv.meta, actualPath) {
		return 1
	}

	// Ensure namespace is a group.
	if _, err = client.Group.GetGroup(ctx, &sdktypes.GetGroupInput{
		Path: &actualPath, // Use extracted path, not original namespacePath
	}); err != nil {
		gsv.meta.Logger.Error(output.FormatError("failed to get group", err))
		return 1
	}

	parser := varparser.NewVariableParser(nil, false)

	variables, err := parser.ParseEnvironmentVariables(&varparser.ParseEnvironmentVariablesInput{EnvVarFilePaths: envVarFiles})
	if err != nil {
		gsv.meta.Logger.Error(output.FormatError("failed to process environment variables", err))
		return 1
	}

	// Prepare the inputs.
	// Extract path from TRN if needed - NamespacePath field expects paths, not TRNs
	actualPath = trn.ToPath(namespacePath)

	input := &sdktypes.SetNamespaceVariablesInput{
		NamespacePath: actualPath,
		Category:      sdktypes.EnvironmentVariableCategory,
		Variables:     convertToSetNamespaceVariablesInput(variables),
	}

	gsv.meta.Logger.Debugf("group set-environment-vars input: %#v", input)

	// Set the group variables.
	err = client.Variable.SetVariables(ctx, input)
	if err != nil {
		gsv.meta.Logger.Error(output.FormatError("failed to set group variables", err))
		return 1
	}

	// Format the output.
	gsv.meta.UI.Output(fmt.Sprintf("Environment variables created successfully in group %s", namespacePath))
	return 0
}

func buildEnvironmentDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"env-var-file": {
			Arguments: []string{"Env_Var_File"},
			Synopsis:  "The path to an environment variables file.",
			Required:  true,
		},
	}
}

func (gsv groupSetEnvironmentVarsCommand) Synopsis() string {
	return "Set environment variables for a group."
}

func (gsv groupSetEnvironmentVarsCommand) Help() string {
	return gsv.HelpGroupSetEnvironmentVars()
}

// HelpGroupSetEnvironmentVars produces the help string for the 'group set-environment-vars' command.
func (gsv groupSetEnvironmentVarsCommand) HelpGroupSetEnvironmentVars() string {
	return fmt.Sprintf(`
Usage: %s [global options] group set-environment-vars [options] <group>

   The group set-environment-vars command sets environment
   variables for a group. Expects an option with the path
   to the variable file.

   Command will overwrite any existing environment variables
   in the target group!

%s

`, gsv.meta.BinaryName, buildHelpText(buildEnvironmentDefs()))
}
