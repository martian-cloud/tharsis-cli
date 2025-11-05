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

// moduleGetCommand is the top-level structure for the module get command.
type moduleGetCommand struct {
	meta *Metadata
}

// NewModuleGetCommandFactory returns a moduleGetCommand struct.
func NewModuleGetCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return moduleGetCommand{
			meta: meta,
		}, nil
	}
}

func (mgc moduleGetCommand) Run(args []string) int {
	mgc.meta.Logger.Debugf("Starting the 'module get' command with %d arguments:", len(args))
	for ix, arg := range args {
		mgc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := mgc.meta.GetSDKClient()
	if err != nil {
		mgc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return mgc.doModuleGet(ctx, client, args)
}

func (mgc moduleGetCommand) doModuleGet(ctx context.Context, client *tharsis.Client, opts []string) int {
	mgc.meta.Logger.Debugf("will do module get, %d opts", len(opts))

	defs := buildJSONOptionDefs(optparser.OptionDefinitions{})
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(mgc.meta.BinaryName+" module get", defs, opts)
	if err != nil {
		mgc.meta.Logger.Error(output.FormatError("failed to parse module get argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		mgc.meta.Logger.Error(output.FormatError("missing module get path", nil), mgc.HelpModuleGet())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive module get arguments: %s", cmdArgs)
		mgc.meta.Logger.Error(output.FormatError(msg, nil), mgc.HelpModuleGet())
		return 1
	}

	modulePath := cmdArgs[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		mgc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	actualPath := trn.ToPath(modulePath)
	if !isResourcePathValid(mgc.meta, actualPath) {
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.GetTerraformModuleInput{Path: &actualPath}  // Use extracted path
	mgc.meta.Logger.Debugf("module get input: %#v", input)

	// Get the Terraform module.
	foundModule, err := client.TerraformModule.GetModule(ctx, input)
	if err != nil {
		mgc.meta.Logger.Error(output.FormatError("failed to get module", err))
		return 1
	}

	return outputModule(mgc.meta, toJSON, foundModule)
}

func (mgc moduleGetCommand) Synopsis() string {
	return "Get a single module."
}

func (mgc moduleGetCommand) Help() string {
	return mgc.HelpModuleGet()
}

// HelpModuleGet prints the help string for the 'module get' command.
func (mgc moduleGetCommand) HelpModuleGet() string {
	return fmt.Sprintf(`
Usage: %s [global options] module get [options] <module-path>

   The module get command prints information about one module.

%s

`, mgc.meta.BinaryName, buildHelpText(buildJSONOptionDefs(optparser.OptionDefinitions{})))
}
