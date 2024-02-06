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

// managedIdentityAccessRuleDeleteCommand is the top-level structure for the managed-identity-access-rule delete command.
type managedIdentityAccessRuleDeleteCommand struct {
	meta *Metadata
}

// NewManagedIdentityAccessRuleDeleteCommandFactory returns a managedIdentityAccessRuleDeleteCommand struct.
func NewManagedIdentityAccessRuleDeleteCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return managedIdentityAccessRuleDeleteCommand{
			meta: meta,
		}, nil
	}
}

func (m managedIdentityAccessRuleDeleteCommand) Run(args []string) int {
	m.meta.Logger.Debugf("Starting the 'managed-identity-access-rule delete' command with %d arguments:", len(args))
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

	return m.doManagedIdentityAccessRuleDelete(ctx, client, args)
}

func (m managedIdentityAccessRuleDeleteCommand) doManagedIdentityAccessRuleDelete(ctx context.Context,
	client *tharsis.Client, opts []string,
) int {
	m.meta.Logger.Debugf("will do managed-identity-access-rule delete, %d opts", len(opts))

	defs := buildManagedIdentityAccessRuleDeleteDefs()
	_, cmdArgs, err := optparser.ParseCommandOptions(m.meta.BinaryName+" managed-identity-access-rule delete", defs, opts)
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to parse managed-identity-access-rule delete options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		m.meta.Logger.Error(output.FormatError("missing managed identity access rule ID", nil))
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive managed-identity-access-rule delete arguments: %s", cmdArgs)
		m.meta.Logger.Error(output.FormatError(msg, nil), m.HelpManagedIdentityAccessRuleDelete())
		return 1
	}

	input := &sdktypes.DeleteManagedIdentityAccessRuleInput{
		ID: cmdArgs[0],
	}
	m.meta.Logger.Debugf("managed-identity-access-rule delete input: %#v", input)

	err = client.ManagedIdentity.DeleteManagedIdentityAccessRule(ctx, input)
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to delete managed identity access rule", err))
		return 1
	}

	// Cannot show the deleted group, but say something.
	m.meta.UI.Output("managed-identity-access-rule delete succeeded.")

	return 0
}

// buildManagedIdentityAccessRuleDeleteDefs returns defs used by managed-identity-access-rule delete command.
func buildManagedIdentityAccessRuleDeleteDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{}
}

func (m managedIdentityAccessRuleDeleteCommand) Synopsis() string {
	return "Delete a managed identity access rule."
}

func (m managedIdentityAccessRuleDeleteCommand) Help() string {
	return m.HelpManagedIdentityAccessRuleDelete()
}

// HelpManagedIdentityAccessRuleDelete produces the help string for the 'managed-identity-access-rule delete' command.
func (m managedIdentityAccessRuleDeleteCommand) HelpManagedIdentityAccessRuleDelete() string {
	return fmt.Sprintf(`
Usage: %s [global options] managed-identity-access-rule delete [options] <managed-identity-access-rule-ID>

   The managed-identity-access-rule delete command deletes a managed identity access rule.

   Use with caution as deleting a managed identity access rule is irreversible!

`, m.meta.BinaryName)
}
