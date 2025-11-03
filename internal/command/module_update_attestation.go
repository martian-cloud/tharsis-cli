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

// moduleUpdateAttestationCommand is the top-level structure for the module update-attestation command.
type moduleUpdateAttestationCommand struct {
	meta *Metadata
}

// NewModuleUpdateAttestationCommandFactory returns a moduleUpdateAttestationCommand struct.
func NewModuleUpdateAttestationCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return moduleUpdateAttestationCommand{
			meta: meta,
		}, nil
	}
}

func (muc moduleUpdateAttestationCommand) Run(args []string) int {
	muc.meta.Logger.Debugf("Starting the 'module update-attestation' command with %d arguments:", len(args))
	for ix, arg := range args {
		muc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := muc.meta.GetSDKClient()
	if err != nil {
		muc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return muc.doModuleUpdateAttestation(ctx, client, args)
}

func (muc moduleUpdateAttestationCommand) doModuleUpdateAttestation(ctx context.Context, client *tharsis.Client, opts []string) int {
	muc.meta.Logger.Debugf("will do module update-attestation, %d opts", len(opts))

	defs := muc.buildUpdateOptionDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(muc.meta.BinaryName+" module update-attestation", defs, opts)
	if err != nil {
		muc.meta.Logger.Error(output.FormatError("failed to parse module update-attestation options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		muc.meta.Logger.Error(output.FormatError("missing module update-attestation ID", nil), muc.HelpModuleUpdateAttestation())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive module update-attestation arguments: %s", cmdArgs)
		muc.meta.Logger.Error(output.FormatError(msg, nil), muc.HelpModuleUpdateAttestation())
		return 1
	}

	description := getOption("description", "", cmdOpts)[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		muc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.UpdateTerraformModuleAttestationInput{
		ID:          cmdArgs[0],
		Description: description,
	}
	muc.meta.Logger.Debugf("module update-attestation input: %#v", input)

	// Update the module attestation.
	updatedAttestation, err := client.TerraformModuleAttestation.UpdateModuleAttestation(ctx, input)
	if err != nil {
		muc.meta.Logger.Error(output.FormatError("failed to update module attestation", err))
		return 1
	}

	return outputModuleAttestation(muc.meta, toJSON, updatedAttestation)
}

// buildUpdateOptionDefs returns the common defs used by 'module update-attestation' command.
func (muc moduleUpdateAttestationCommand) buildUpdateOptionDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"description": {
			Arguments: []string{"Description"},
			Synopsis:  "New description for the module attestation.",
		},
	}

	return buildJSONOptionDefs(defs)
}

func (muc moduleUpdateAttestationCommand) Synopsis() string {
	return "Update a module attestation."
}

func (muc moduleUpdateAttestationCommand) Help() string {
	return muc.HelpModuleUpdateAttestation()
}

// HelpModuleUpdateAttestation produces the help string for the 'module update-attestation' command.
func (muc moduleUpdateAttestationCommand) HelpModuleUpdateAttestation() string {
	return fmt.Sprintf(`
Usage: %s [global options] module update-attestation [options] <id>

   The module update-attestation command updates a module
   attestation. Shows final output as JSON, if specified.

%s

`, muc.meta.BinaryName, buildHelpText(muc.buildUpdateOptionDefs()))
}
