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

// managedIdentityAccessRuleCreateCommand is the top-level structure for the managed-identity-access-rule create command.
type managedIdentityAccessRuleCreateCommand struct {
	meta *Metadata
}

// NewManagedIdentityAccessRuleCreateCommandFactory returns a managedIdentityAccessRuleCreateCommand struct.
func NewManagedIdentityAccessRuleCreateCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return managedIdentityAccessRuleCreateCommand{
			meta: meta,
		}, nil
	}
}

func (m managedIdentityAccessRuleCreateCommand) Run(args []string) int {
	m.meta.Logger.Debugf("Starting the 'managed-identity-access-rule create' command with %d arguments:", len(args))
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

	return m.doManagedIdentityAccessRuleCreate(ctx, client, args)
}

func (m managedIdentityAccessRuleCreateCommand) doManagedIdentityAccessRuleCreate(ctx context.Context,
	client *tharsis.Client, opts []string,
) int {
	m.meta.Logger.Debugf("will do managed-identity-access-rule create, %d opts", len(opts))

	defs := buildManagedIdentityAccessRuleCreateDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(m.meta.BinaryName+" managed-identity-access-rule create", defs, opts)
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to parse managed-identity-access-rule create options", err))
		return 1
	}
	if len(cmdArgs) > 0 {
		msg := fmt.Sprintf("excessive managed-identity-access-rule create arguments: %s", cmdArgs)
		m.meta.Logger.Error(output.FormatError(msg, nil), m.HelpManagedIdentityAccessRuleCreate())
		return 1
	}

	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	managedIdentityPath := getOption("managed-identity-path", "", cmdOpts)[0]
	ruleType := getOption("rule-type", "", cmdOpts)[0]
	moduleAttestationPolicies := getOptionSlice("module-attestation-policy", cmdOpts)
	runStage := getOption("run-stage", "", cmdOpts)[0]
	allowedUsers := getOptionSlice("allowed-user", cmdOpts)
	allowedServiceAccounts := getOptionSlice("allowed-service-account", cmdOpts)
	allowedTeams := getOptionSlice("allowed-team", cmdOpts)

	// Get the managed identity from its path.
	managedIdentity, err := client.ManagedIdentity.GetManagedIdentity(ctx, &sdktypes.GetManagedIdentityInput{
		Path: &managedIdentityPath,
	})
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to get managed identity", err))
		return 1
	}

	// Build the module attestation policies.
	policies, err := buildModuleAttestationPolicies(moduleAttestationPolicies)
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to parse/build module attestation policies", err))
		return 1
	}

	input := &sdktypes.CreateManagedIdentityAccessRuleInput{
		Type:                      sdktypes.ManagedIdentityAccessRuleType(ruleType),
		ModuleAttestationPolicies: policies,
		ManagedIdentityID:         managedIdentity.Metadata.ID,
		RunStage:                  sdktypes.JobType(runStage),
		AllowedUsers:              allowedUsers,
		AllowedServiceAccounts:    allowedServiceAccounts,
		AllowedTeams:              allowedTeams,
	}
	m.meta.Logger.Debugf("managed-identity-access-rule create input: %#v", input)

	managedIdentityAccessRule, err := client.ManagedIdentity.CreateManagedIdentityAccessRule(ctx, input)
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to create managed identity access rule", err))
		return 1
	}

	return outputManagedIdentityAccessRule(m.meta, toJSON, managedIdentityAccessRule)
}

// outputManagedIdentityAccessRule is the final output for most managed-identity-access-rule operations.
// It displays exactly one access rule.
func outputManagedIdentityAccessRule(meta *Metadata, toJSON bool,
	managedIdentityAccessRule *sdktypes.ManagedIdentityAccessRule,
) int {
	if toJSON {
		buf, err := objectToJSON(managedIdentityAccessRule)
		if err != nil {
			meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		meta.UI.Output(string(buf))
		return 0
	}

	return outputManagedIdentityAccessRulesTable(meta, []sdktypes.ManagedIdentityAccessRule{*managedIdentityAccessRule})
}

// outputManagedIdentityAccessRules is the final output for the managed-identity-access-rule list operation.
func outputManagedIdentityAccessRules(meta *Metadata, toJSON bool,
	managedIdentityAccessRules []sdktypes.ManagedIdentityAccessRule,
) int {
	if toJSON {
		buf, err := objectToJSON(managedIdentityAccessRules)
		if err != nil {
			meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		meta.UI.Output(string(buf))
		return 0
	}

	return outputManagedIdentityAccessRulesTable(meta, managedIdentityAccessRules)

}

// outputManagedIdentityAccessRulesTable is the final output for managed-identity-access-rule
// operations when JSON has not been requested.
func outputManagedIdentityAccessRulesTable(meta *Metadata,
	managedIdentityAccessRules []sdktypes.ManagedIdentityAccessRule,
) int {
	tableInput := [][]string{
		{"id", "type", "run-stage"},
	}

	for _, rule := range managedIdentityAccessRules {
		tableInput = append(tableInput, []string{
			rule.Metadata.ID,
			string(rule.Type),
			string(rule.RunStage),
		})
	}

	meta.UI.Output(tableformatter.FormatTable(tableInput))
	// For now, this function does not display the module attestation policies.

	return 0
}

// buildManagedIdentityAccessRuleSharedDefs returns defs used by managed-identity-access-rule create and update commands.
func buildManagedIdentityAccessRuleSharedDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"module-attestation-policy": {
			Arguments: []string{"Module_Attestation_Policy"},
			Synopsis:  "One module attestation policy: --module-attestation-policy=\"[PredicateType=someval,]PublicKeyFile=/path/to/file\".",
		},
		"allowed-user": {
			Arguments: []string{"Allowed_User"},
			Synopsis:  "One allowed username (repeatable).",
		},
		"allowed-service-account": {
			Arguments: []string{"Allowed_Service_Account"},
			Synopsis:  "One allowed service account resource path (repeatable).",
		},
		"allowed-team": {
			Arguments: []string{"Allowed_Team"},
			Synopsis:  "One allowed team name (repeatable).",
		},
	}
	return buildJSONOptionDefs(defs)
}

// buildManagedIdentityAccessRuleCreateDefs returns defs used by managed-identity-access-rule create command.
func buildManagedIdentityAccessRuleCreateDefs() optparser.OptionDefinitions {
	defs := buildManagedIdentityAccessRuleSharedDefs()
	defs["managed-identity-path"] = &optparser.OptionDefinition{
		Arguments: []string{"Managed_Identity_Path"},
		Synopsis:  "Resource path to the managed identity.",
		Required:  true,
	}
	defs["run-stage"] = &optparser.OptionDefinition{
		Arguments: []string{"Run_Stage"},
		Synopsis:  "Which run stage (job type) for this rule: plan or apply.",
		Required:  true,
	}
	defs["rule-type"] = &optparser.OptionDefinition{
		Arguments: []string{"Rule_Type"},
		Synopsis:  "Which type of rule to create: 'eligible_principals' or 'module_attestation'.",
		Required:  true,
	}
	return defs
}

func (m managedIdentityAccessRuleCreateCommand) Synopsis() string {
	return "Create a new managed identity access rule."
}

func (m managedIdentityAccessRuleCreateCommand) Help() string {
	return m.HelpManagedIdentityAccessRuleCreate()
}

// HelpManagedIdentityAccessRuleCreate produces the help string for the 'managed-identity-access-rule create' command.
func (m managedIdentityAccessRuleCreateCommand) HelpManagedIdentityAccessRuleCreate() string {
	return fmt.Sprintf(`
Usage: %s [global options] managed-identity-access-rule create [options]

%s

`, m.meta.BinaryName, buildHelpText(buildManagedIdentityAccessRuleCreateDefs()))
}
