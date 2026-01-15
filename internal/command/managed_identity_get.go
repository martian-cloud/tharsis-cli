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

// managedIdentityGetCommand is the top-level structure for the managed-identity get command.
type managedIdentityGetCommand struct {
	meta *Metadata
}

// NewManagedIdentityGetCommandFactory returns a managedIdentityGetCommand struct.
func NewManagedIdentityGetCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return managedIdentityGetCommand{
			meta: meta,
		}, nil
	}
}

func (m managedIdentityGetCommand) Run(args []string) int {
	m.meta.Logger.Debugf("Starting the 'managed-identity get' command with %d arguments:", len(args))
	for ix, arg := range args {
		m.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := m.meta.GetSDKClient()
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return m.doManagedIdentityGet(ctx, client, args)
}

func (m managedIdentityGetCommand) doManagedIdentityGet(ctx context.Context, client *tharsis.Client, opts []string) int {
	m.meta.Logger.Debugf("will do managed-identity get, %d opts", len(opts))

	defs := buildJSONOptionDefs(optparser.OptionDefinitions{})
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(m.meta.BinaryName+" managed-identity get", defs, opts)
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to parse managed-identity get argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		m.meta.Logger.Error(output.FormatError("missing managed-identity get path", nil), m.HelpManagedIdentityGet())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive managed-identity get arguments: %s", cmdArgs)
		m.meta.Logger.Error(output.FormatError(msg, nil), m.HelpManagedIdentityGet())
		return 1
	}

	managedIdentityPath := cmdArgs[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	actualPath := trn.ToPath(managedIdentityPath)
	if !isResourcePathValid(m.meta, actualPath) {
		return 1
	}

	// Prepare the inputs - convert path to TRN and use ID field
	trnID := trn.ToTRN(managedIdentityPath, trn.ResourceTypeManagedIdentity)
	input := &sdktypes.GetManagedIdentityInput{ID: &trnID}
	m.meta.Logger.Debugf("managed-identity get input: %#v", input)

	// Get the managed identity.
	foundManagedIdentity, err := client.ManagedIdentity.GetManagedIdentity(ctx, input)
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to get managed identity", err))
		return 1
	}

	if err = outputManagedIdentity(m.meta, toJSON, foundManagedIdentity); err != nil {
		m.meta.UI.Error(err.Error())
		return 1
	}

	return 0
}

func (m managedIdentityGetCommand) Synopsis() string {
	return "Get a single managed identity."
}

func (m managedIdentityGetCommand) Help() string {
	return m.HelpManagedIdentityGet()
}

// HelpManagedIdentityGet prints the help string for the 'managed-identity get' command.
func (m managedIdentityGetCommand) HelpManagedIdentityGet() string {
	return fmt.Sprintf(`
Usage: %s [global options] managed-identity get [options] <managed-identity-path>

   The managed-identity get command prints information about one managed identity.

%s

`, m.meta.BinaryName, buildHelpText(buildJSONOptionDefs(optparser.OptionDefinitions{})))
}
