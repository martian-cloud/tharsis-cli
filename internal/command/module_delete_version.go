package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// moduleDeleteVersionCommand is the top-level structure for the module delete-version command.
type moduleDeleteVersionCommand struct {
	meta *Metadata
}

// NewModuleDeleteVersionCommandFactory returns a moduleDeleteVersionCommand struct.
func NewModuleDeleteVersionCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return moduleDeleteVersionCommand{
			meta: meta,
		}, nil
	}
}

func (mdc moduleDeleteVersionCommand) Run(args []string) int {
	mdc.meta.Logger.Debugf("Starting the 'module delete-version' command with %d arguments:", len(args))
	for ix, arg := range args {
		mdc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := mdc.meta.GetSDKClient()
	if err != nil {
		mdc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return mdc.doModuleDeleteVersion(ctx, client, args)
}

func (mdc moduleDeleteVersionCommand) doModuleDeleteVersion(ctx context.Context, client *tharsis.Client, opts []string) int {
	mdc.meta.Logger.Debugf("will do module delete-version, %d opts", len(opts))

	// No options to parse.
	_, cmdArgs, err := optparser.ParseCommandOptions(mdc.meta.BinaryName+" module delete-version", optparser.OptionDefinitions{}, opts)
	if err != nil {
		mdc.meta.Logger.Error(output.FormatError("failed to parse module delete-version options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		mdc.meta.Logger.Error(output.FormatError("missing module delete-version ID", nil), mdc.HelpModuleDeleteVersion())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive module delete-version arguments: %s", cmdArgs)
		mdc.meta.Logger.Error(output.FormatError(msg, nil), mdc.HelpModuleDeleteVersion())
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.DeleteTerraformModuleVersionInput{ID: cmdArgs[0]}
	mdc.meta.Logger.Debugf("module delete input: %#v", input)

	// Delete the module version.
	err = client.TerraformModuleVersion.DeleteModuleVersion(ctx, input)
	if err != nil {
		mdc.meta.Logger.Error(output.FormatError("failed to delete a module version", err))
		return 1
	}

	// Cannot show the deleted module version, but say something.
	mdc.meta.UI.Output("module version delete succeeded.")

	return 0
}

func (mdc moduleDeleteVersionCommand) Synopsis() string {
	return "Delete a module version."
}

func (mdc moduleDeleteVersionCommand) Help() string {
	return mdc.HelpModuleDeleteVersion()
}

// HelpModuleDeleteVersion produces the help string for the 'module delete-version' command.
func (mdc moduleDeleteVersionCommand) HelpModuleDeleteVersion() string {
	return fmt.Sprintf(`
Usage: %s [global options] module delete-version <id>

   The module delete-version command deletes a module version.

   Use with caution as deleting a module version is irreversible!

`, mdc.meta.BinaryName)
}
