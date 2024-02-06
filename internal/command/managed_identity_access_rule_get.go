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

// managedIdentityAccessRuleGetCommand is the top-level structure for the managed-identity-access-rule get command.
type managedIdentityAccessRuleGetCommand struct {
	meta *Metadata
}

// NewManagedIdentityAccessRuleGetCommandFactory returns a managedIdentityAccessRuleGetCommand struct.
func NewManagedIdentityAccessRuleGetCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return managedIdentityAccessRuleGetCommand{
			meta: meta,
		}, nil
	}
}

func (m managedIdentityAccessRuleGetCommand) Run(args []string) int {
	m.meta.Logger.Debugf("Starting the 'managed-identity-access-rule get' command with %d arguments:", len(args))
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

	return m.doManagedIdentityAccessRuleGet(ctx, client, args)
}

func (m managedIdentityAccessRuleGetCommand) doManagedIdentityAccessRuleGet(ctx context.Context,
	client *tharsis.Client, opts []string,
) int {
	m.meta.Logger.Debugf("will do managed-identity-access-rule get, %d opts", len(opts))

	defs := buildManagedIdentityAccessRuleGetDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(m.meta.BinaryName+" managed-identity-access-rule get", defs, opts)
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to parse managed-identity-access-rule get options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		m.meta.Logger.Error(output.FormatError("missing managed identity access rule ID", nil))
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive managed-identity-access-rule get arguments: %s", cmdArgs)
		m.meta.Logger.Error(output.FormatError(msg, nil), m.HelpManagedIdentityAccessRuleGet())
		return 1
	}

	managedIdentityAccessRuleID := cmdArgs[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Get the managed identity access rule from its path.
	managedIdentityAccessRule, err := client.ManagedIdentity.GetManagedIdentityAccessRule(ctx, &sdktypes.GetManagedIdentityAccessRuleInput{
		ID: managedIdentityAccessRuleID,
	})
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to get managed identity access rule", err))
		return 1
	}

	return outputManagedIdentityAccessRule(m.meta, toJSON, managedIdentityAccessRule)
}

// buildManagedIdentityAccessRuleGetDefs returns defs used by managed-identity-access-rule get command.
func buildManagedIdentityAccessRuleGetDefs() optparser.OptionDefinitions {
	return buildJSONOptionDefs(optparser.OptionDefinitions{})
}

func (m managedIdentityAccessRuleGetCommand) Synopsis() string {
	return "Get a managed identity access rule."
}

func (m managedIdentityAccessRuleGetCommand) Help() string {
	return m.HelpManagedIdentityAccessRuleGet()
}

// HelpManagedIdentityAccessRuleGet produces the help string for the 'managed-identity-access-rule get' command.
func (m managedIdentityAccessRuleGetCommand) HelpManagedIdentityAccessRuleGet() string {
	return fmt.Sprintf(`
Usage: %s [global options] managed-identity-access-rule get [options] <managed-identity-access-rule-ID>

%s

`, m.meta.BinaryName, buildHelpText(buildManagedIdentityAccessRuleGetDefs()))
}
