package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tableformatter"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// moduleCreateAttestationCommand is the top-level structure for the module create-attestation command.
type moduleCreateAttestationCommand struct {
	meta *Metadata
}

// NewModuleCreateAttestationCommandFactory returns a moduleCreateAttestationCommand struct.
func NewModuleCreateAttestationCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return moduleCreateAttestationCommand{
			meta: meta,
		}, nil
	}
}

func (mcc moduleCreateAttestationCommand) Run(args []string) int {
	mcc.meta.Logger.Debugf("Starting the 'module create-attestation' command with %d arguments:", len(args))
	for ix, arg := range args {
		mcc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := mcc.meta.GetSDKClient()
	if err != nil {
		mcc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return mcc.doModuleCreateAttestation(ctx, client, args)
}

func (mcc moduleCreateAttestationCommand) doModuleCreateAttestation(ctx context.Context, client *tharsis.Client, opts []string) int {
	mcc.meta.Logger.Debugf("will do module create-attestation, %d opts", len(opts))

	defs := mcc.buildDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(mcc.meta.BinaryName+" module create-attestation", defs, opts)
	if err != nil {
		mcc.meta.Logger.Error(output.FormatError("failed to parse module create-attestation options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		mcc.meta.Logger.Error(output.FormatError("missing module create-attestation module path", nil), mcc.HelpModuleCreateAttestation())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive module create-attestation arguments: %s", cmdArgs)
		mcc.meta.Logger.Error(output.FormatError(msg, nil), mcc.HelpModuleCreateAttestation())
		return 1
	}

	modulePath := cmdArgs[0]
	attestationData := getOption("data", "", cmdOpts)[0]
	description := getOption("description", "", cmdOpts)[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		mcc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Error is already logged.
	if !isResourcePathValid(mcc.meta, modulePath) {
		return 1
	}

	input := &sdktypes.CreateTerraformModuleAttestationInput{
		ModulePath:      modulePath,
		Description:     description,
		AttestationData: attestationData,
	}
	mcc.meta.Logger.Debugf("module create input: %#v", input)

	attestation, err := client.TerraformModuleAttestation.CreateModuleAttestation(ctx, input)
	if err != nil {
		mcc.meta.UI.Error(output.FormatError("failed to create module attestation", err))
		return 1
	}

	return outputModuleAttestation(mcc.meta, toJSON, attestation)
}

// outputModuleAttestation is the final output for most module attestation operations.
func outputModuleAttestation(meta *Metadata, toJSON bool, attestation *sdktypes.TerraformModuleAttestation) int {
	if toJSON {
		buf, err := objectToJSON(attestation)
		if err != nil {
			meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		meta.UI.Output(string(buf))
	} else {
		tableInput := [][]string{
			{"id", "module id", "description", "schema type", "predicate type"},
			{
				attestation.Metadata.ID, attestation.ModuleID, attestation.Description,
				attestation.SchemaType, attestation.PredicateType,
			},
		}
		meta.UI.Output(tableformatter.FormatTable(tableInput))
	}

	return 0
}

// buildModuleCreateAttestationDefs returns defs used by module create-attestation command.
func (mcc moduleCreateAttestationCommand) buildDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"data": {
			Arguments: []string{"Data"},
			Synopsis:  "The Base64-encoded attestation data.",
			Required:  true,
		},
		"description": {
			Arguments: []string{"Description"},
			Synopsis:  "Description for the new module attestation.",
		},
	}

	return buildJSONOptionDefs(defs)
}

func (mcc moduleCreateAttestationCommand) Synopsis() string {
	return "Create a new module attestation."
}

func (mcc moduleCreateAttestationCommand) Help() string {
	return mcc.HelpModuleCreateAttestation()
}

// HelpModuleCreateAttestation produces the help string for the 'module create-attestation' command.
func (mcc moduleCreateAttestationCommand) HelpModuleCreateAttestation() string {
	return fmt.Sprintf(`
Usage: %s [global options] module create-attestation [options] <module-path>

   The module create-attestation command creates a
   new module attestation. Attestation data must
   only be a Base64-encoded string.

%s

`, mcc.meta.BinaryName, buildHelpText(mcc.buildDefs()))
}
