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

// managedIdentityAccessRuleUpdateCommand is the top-level structure for the managed-identity-access-rule update command.
type managedIdentityAccessRuleUpdateCommand struct {
	meta *Metadata
}

// NewManagedIdentityAccessRuleUpdateCommandFactory returns a managedIdentityAccessRuleUpdateCommand struct.
func NewManagedIdentityAccessRuleUpdateCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return managedIdentityAccessRuleUpdateCommand{
			meta: meta,
		}, nil
	}
}

func (m managedIdentityAccessRuleUpdateCommand) Run(args []string) int {
	m.meta.Logger.Debugf("Starting the 'managed-identity-access-rule update' command with %d arguments:", len(args))
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

	return m.doManagedIdentityAccessRuleUpdate(ctx, client, args)
}

func (m managedIdentityAccessRuleUpdateCommand) doManagedIdentityAccessRuleUpdate(ctx context.Context,
	client *tharsis.Client, opts []string,
) int {
	m.meta.Logger.Debugf("will do managed-identity-access-rule update, %d opts", len(opts))

	defs := buildManagedIdentityAccessRuleUpdateDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(m.meta.BinaryName+" managed-identity-access-rule update", defs, opts)
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to parse managed-identity-access-rule update options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		m.meta.Logger.Error(output.FormatError("missing managed identity access rule ID", nil))
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive managed-identity-access-rule update arguments: %s", cmdArgs)
		m.meta.Logger.Error(output.FormatError(msg, nil), m.HelpManagedIdentityAccessRuleUpdate())
		return 1
	}

	managedIdentityAccessRuleID := cmdArgs[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	moduleAttestationPolicies := getOptionSlice("module-attestation-policy", cmdOpts)
	allowedUsers := getOptionSlice("allowed-user", cmdOpts)
	allowedServiceAccounts := getOptionSlice("allowed-service-account", cmdOpts)
	allowedTeams := getOptionSlice("allowed-team", cmdOpts)

	var verifyStateLineage *bool
	if _, ok := cmdOpts["verify-state-lineage"]; ok {
		v, vErr := getBoolOptionValue("verify-state-lineage", "false", cmdOpts)
		if vErr != nil {
			m.meta.Logger.Error(output.FormatError("failed to parse -verify-state-lineage option value", vErr))
			return 1
		}
		verifyStateLineage = &v
	}

	// Get the original managed identity access rule from its path.
	managedIdentityAccessRule, err := client.ManagedIdentity.GetManagedIdentityAccessRule(ctx, &sdktypes.GetManagedIdentityAccessRuleInput{
		ID: managedIdentityAccessRuleID,
	})
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to get managed identity access rule", err))
		return 1
	}

	// Build the input based on the original managed identity access rule and the options.
	// If no -allowed-user, -allowed-service-account, or -allowed-team option is specified,
	// empty the respective slice.
	input := &sdktypes.UpdateManagedIdentityAccessRuleInput{
		ID:                     managedIdentityAccessRuleID,
		RunStage:               managedIdentityAccessRule.RunStage,
		AllowedUsers:           allowedUsers,
		AllowedServiceAccounts: allowedServiceAccounts,
		AllowedTeams:           allowedTeams,
		VerifyStateLineage:     verifyStateLineage,
	}

	// If no -module-attestation-policy option is specified, error out if it's a module attestation rule.
	if _, ok := cmdOpts["module-attestation-policy"]; ok {
		input.ModuleAttestationPolicies, err = buildModuleAttestationPolicies(moduleAttestationPolicies)
		if err != nil {
			m.meta.Logger.Error(output.FormatError("failed to parse/build module attestation policies", err))
			return 1
		}
	} else if managedIdentityAccessRule.Type == sdktypes.ManagedIdentityAccessRuleModuleAttestation {
		m.meta.UI.Error(output.FormatError("at least one attestation policy is required", err))
		return 1
	}

	m.meta.Logger.Debugf("managed-identity-access-rule update input: %#v", input)
	managedIdentityAccessRule, err = client.ManagedIdentity.UpdateManagedIdentityAccessRule(ctx, input)
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to update managed identity access rule", err))
		return 1
	}

	return outputManagedIdentityAccessRule(m.meta, toJSON, managedIdentityAccessRule)
}

// buildManagedIdentityAccessRuleUpdateDefs returns defs used by managed-identity-access-rule update command.
func buildManagedIdentityAccessRuleUpdateDefs() optparser.OptionDefinitions {
	return buildManagedIdentityAccessRuleSharedDefs()
}

func (m managedIdentityAccessRuleUpdateCommand) Synopsis() string {
	return "Update a new managed identity access rule."
}

func (m managedIdentityAccessRuleUpdateCommand) Help() string {
	return m.HelpManagedIdentityAccessRuleUpdate()
}

// HelpManagedIdentityAccessRuleUpdate produces the help string for the 'managed-identity-access-rule update' command.
func (m managedIdentityAccessRuleUpdateCommand) HelpManagedIdentityAccessRuleUpdate() string {
	return fmt.Sprintf(`
Usage: %s [global options] managed-identity-access-rule update [options] <managed-identity-access-rule-ID>

%s

`, m.meta.BinaryName, buildHelpText(buildManagedIdentityAccessRuleUpdateDefs()))
}
