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

// moduleDeleteAttestationCommand is the top-level structure for the module delete-attestation command.
type moduleDeleteAttestationCommand struct {
	meta *Metadata
}

// NewModuleDeleteAttestationCommandFactory returns a moduleDeleteAttestationCommand struct.
func NewModuleDeleteAttestationCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return moduleDeleteAttestationCommand{
			meta: meta,
		}, nil
	}
}

func (mdc moduleDeleteAttestationCommand) Run(args []string) int {
	mdc.meta.Logger.Debugf("Starting the 'module delete-attestation' command with %d arguments:", len(args))
	for ix, arg := range args {
		mdc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := mdc.meta.GetSDKClient()
	if err != nil {
		mdc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return mdc.doModuleDeleteAttestation(ctx, client, args)
}

func (mdc moduleDeleteAttestationCommand) doModuleDeleteAttestation(ctx context.Context, client *tharsis.Client, opts []string) int {
	mdc.meta.Logger.Debugf("will do module delete-attestation, %d opts", len(opts))

	// No options to parse.
	_, cmdArgs, err := optparser.ParseCommandOptions(mdc.meta.BinaryName+" module delete-attestation", optparser.OptionDefinitions{}, opts)
	if err != nil {
		mdc.meta.Logger.Error(output.FormatError("failed to parse module delete-attestation options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		mdc.meta.Logger.Error(output.FormatError("missing module delete-attestation ID", nil), mdc.HelpModuleDeleteAttestation())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive module delete-attestation arguments: %s", cmdArgs)
		mdc.meta.Logger.Error(output.FormatError(msg, nil), mdc.HelpModuleDeleteAttestation())
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.DeleteTerraformModuleAttestationInput{ID: cmdArgs[0]}
	mdc.meta.Logger.Debugf("module delete-attestation input: %#v", input)

	// Delete the module attestation.
	err = client.TerraformModuleAttestation.DeleteModuleAttestation(ctx, input)
	if err != nil {
		mdc.meta.Logger.Error(output.FormatError("failed to delete module attestation", err))
		return 1
	}

	// Cannot show the deleted module attestation, but say something.
	mdc.meta.UI.Output("module attestation delete succeeded.")

	return 0
}

func (mdc moduleDeleteAttestationCommand) Synopsis() string {
	return "Delete a module attestation."
}

func (mdc moduleDeleteAttestationCommand) Help() string {
	return mdc.HelpModuleDeleteAttestation()
}

// HelpModuleDeleteAttestation produces the help string for the 'module delete-attestation' command.
func (mdc moduleDeleteAttestationCommand) HelpModuleDeleteAttestation() string {
	return fmt.Sprintf(`
Usage: %s [global options] module delete-attestation <id>

   The module delete-attestation command deletes a module attestation.

   Use with caution as deleting a module attestation is irreversible!

`, mdc.meta.BinaryName)
}
