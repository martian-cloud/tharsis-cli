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

// managedIdentityDeleteCommand is the top-level structure for the managed-identity delete command.
type managedIdentityDeleteCommand struct {
	meta *Metadata
}

// NewManagedIdentityDeleteCommandFactory returns a managedIdentityDeleteCommand struct.
func NewManagedIdentityDeleteCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return managedIdentityDeleteCommand{
			meta: meta,
		}, nil
	}
}

func (m managedIdentityDeleteCommand) Run(args []string) int {
	m.meta.Logger.Debugf("Starting the 'managed-identity delete' command with %d arguments:", len(args))
	for ix, arg := range args {
		m.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := m.meta.GetSDKClient()
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return m.doManagedIdentityDelete(ctx, client, args)
}

func (m managedIdentityDeleteCommand) doManagedIdentityDelete(ctx context.Context, client *tharsis.Client, opts []string) int {
	m.meta.Logger.Debugf("will do managed-identity delete, %d opts", len(opts))

	// No options to parse.
	_, cmdArgs, err := optparser.ParseCommandOptions(m.meta.BinaryName+" managed-identity delete", optparser.OptionDefinitions{}, opts)
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to parse managed-identity delete options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		m.meta.Logger.Error(output.FormatError("missing managed-identity delete path", nil), m.HelpManagedIdentityDelete())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive managed-identity delete arguments: %s", cmdArgs)
		m.meta.Logger.Error(output.FormatError(msg, nil), m.HelpManagedIdentityDelete())
		return 1
	}

	managedIdentityPath := cmdArgs[0]
	if !isResourcePathValid(m.meta, managedIdentityPath) {
		return 1
	}

	managedIdentity, err := client.ManagedIdentity.GetManagedIdentity(ctx, &sdktypes.GetManagedIdentityInput{
		Path: &managedIdentityPath,
	})
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to get managed identity", err))
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.DeleteManagedIdentityInput{ID: managedIdentity.Metadata.ID}
	m.meta.Logger.Debugf("managed-identity delete input: %#v", input)

	// Delete the managed identity.
	err = client.ManagedIdentity.DeleteManagedIdentity(ctx, input)
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to delete managed identity", err))
		return 1
	}

	// Cannot show the deleted managed identity, but say something.
	m.meta.UI.Output("managed-identity delete succeeded.")

	return 0
}

func (m managedIdentityDeleteCommand) Synopsis() string {
	return "Delete a managed identity."
}

func (m managedIdentityDeleteCommand) Help() string {
	return m.HelpManagedIdentityDelete()
}

// HelpManagedIdentityDelete produces the help string for the 'managed-identity delete' command.
func (m managedIdentityDeleteCommand) HelpManagedIdentityDelete() string {
	return fmt.Sprintf(`
Usage: %s [global options] managed-identity delete <managed-identity-path>

   The managed-identity delete command deletes a managed identity.

   Use with caution as deleting a managed identity is irreversible!

`, m.meta.BinaryName)
}
