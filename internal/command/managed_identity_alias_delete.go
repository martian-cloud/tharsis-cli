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

// managedIdentityAliasDeleteCommand is the top-level structure for the managed-identity-alias delete command.
type managedIdentityAliasDeleteCommand struct {
	meta *Metadata
}

// NewManagedIdentityAliasDeleteCommandFactory returns a managedIdentityAliasDeleteCommand struct.
func NewManagedIdentityAliasDeleteCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return managedIdentityAliasDeleteCommand{
			meta: meta,
		}, nil
	}
}

func (m managedIdentityAliasDeleteCommand) Run(args []string) int {
	m.meta.Logger.Debugf("Starting the 'managed-identity-alias delete' command with %d arguments:", len(args))
	for ix, arg := range args {
		m.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := m.meta.ReadSettings()
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return m.doManagedIdentityAliasDelete(ctx, client, args)
}

func (m managedIdentityAliasDeleteCommand) doManagedIdentityAliasDelete(ctx context.Context, client *tharsis.Client, opts []string) int {
	m.meta.Logger.Debugf("will do managed-identity-alias delete, %d opts", len(opts))

	// No options to parse.
	_, cmdArgs, err := optparser.ParseCommandOptions(m.meta.BinaryName+" managed-identity-alias delete", optparser.OptionDefinitions{}, opts)
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to parse managed-identity-alias delete options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		m.meta.Logger.Error(output.FormatError("missing managed-identity-alias delete path", nil), m.HelpManagedIdentityAliasDelete())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive managed-identity-alias delete arguments: %s", cmdArgs)
		m.meta.Logger.Error(output.FormatError(msg, nil), m.HelpManagedIdentityAliasDelete())
		return 1
	}

	managedIdentityAliasPath := cmdArgs[0]
	if !isResourcePathValid(m.meta, managedIdentityAliasPath) {
		return 1
	}

	managedIdentityAlias, err := client.ManagedIdentity.GetManagedIdentity(ctx, &sdktypes.GetManagedIdentityInput{
		Path: &managedIdentityAliasPath,
	})
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to get managed identity alias", err))
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.DeleteManagedIdentityAliasInput{ID: managedIdentityAlias.Metadata.ID}
	m.meta.Logger.Debugf("managed-identity-alias delete input: %#v", input)

	// Delete the managed identity alias.
	err = client.ManagedIdentity.DeleteManagedIdentityAlias(ctx, input)
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to delete managed identity alias", err))
		return 1
	}

	// Cannot show the deleted managed identity alias, but say something.
	m.meta.UI.Output("managed-identity-alias delete succeeded.")

	return 0
}

func (m managedIdentityAliasDeleteCommand) Synopsis() string {
	return "Delete a managed identity alias."
}

func (m managedIdentityAliasDeleteCommand) Help() string {
	return m.HelpManagedIdentityAliasDelete()
}

// HelpManagedIdentityAliasDelete produces the help string for the 'managed-identity-alias delete' command.
func (m managedIdentityAliasDeleteCommand) HelpManagedIdentityAliasDelete() string {
	return fmt.Sprintf(`
Usage: %s [global options] managed-identity-alias delete <managed-identity-alias-path>

   The managed-identity-alias delete command deletes a managed identity alias.

   Use with caution as deleting a managed identity alias is irreversible!

`, m.meta.BinaryName)
}
