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

// moduleDeleteCommand is the top-level structure for the module delete command.
type moduleDeleteCommand struct {
	meta *Metadata
}

// NewModuleDeleteCommandFactory returns a moduleDeleteCommand struct.
func NewModuleDeleteCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return moduleDeleteCommand{
			meta: meta,
		}, nil
	}
}

func (mdc moduleDeleteCommand) Run(args []string) int {
	mdc.meta.Logger.Debugf("Starting the 'module delete' command with %d arguments:", len(args))
	for ix, arg := range args {
		mdc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := mdc.meta.ReadSettings()
	if err != nil {
		mdc.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		mdc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return mdc.doModuleDelete(ctx, client, args)
}

func (mdc moduleDeleteCommand) doModuleDelete(ctx context.Context, client *tharsis.Client, opts []string) int {
	mdc.meta.Logger.Debugf("will do module delete, %d opts", len(opts))

	// No options to parse.
	_, cmdArgs, err := optparser.ParseCommandOptions(mdc.meta.BinaryName+" module delete", optparser.OptionDefinitions{}, opts)
	if err != nil {
		mdc.meta.Logger.Error(output.FormatError("failed to parse module delete options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		mdc.meta.Logger.Error(output.FormatError("missing module delete path", nil), mdc.HelpModuleDelete())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive module delete arguments: %s", cmdArgs)
		mdc.meta.Logger.Error(output.FormatError(msg, nil), mdc.HelpModuleDelete())
		return 1
	}

	modulePath := cmdArgs[0]

	if !isResourcePathValid(mdc.meta, modulePath) {
		return 1
	}

	module, err := client.TerraformModule.GetModule(ctx, &sdktypes.GetTerraformModuleInput{Path: &modulePath})
	if err != nil {
		mdc.meta.Logger.Error(output.FormatError("failed to get module", err))
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.DeleteTerraformModuleInput{ID: module.Metadata.ID}
	mdc.meta.Logger.Debugf("module delete input: %#v", input)

	// Delete the module.
	err = client.TerraformModule.DeleteModule(ctx, input)
	if err != nil {
		mdc.meta.Logger.Error(output.FormatError("failed to delete module", err))
		return 1
	}

	// Cannot show the deleted module, but say something.
	mdc.meta.UI.Output("module delete succeeded.")

	return 0
}

func (mdc moduleDeleteCommand) Synopsis() string {
	return "Delete a module."
}

func (mdc moduleDeleteCommand) Help() string {
	return mdc.HelpModuleDelete()
}

// HelpModuleDelete produces the help string for the 'module delete' command.
func (mdc moduleDeleteCommand) HelpModuleDelete() string {
	return fmt.Sprintf(`
Usage: %s [global options] module delete <module-path>

   The module delete command deletes a module.

   Use with caution as deleting a module is irreversible!

`, mdc.meta.BinaryName)
}
