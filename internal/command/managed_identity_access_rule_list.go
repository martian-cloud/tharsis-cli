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

// managedIdentityAccessRuleListCommand is the top-level structure for the managed-identity-access-rule list command.
type managedIdentityAccessRuleListCommand struct {
	meta *Metadata
}

// NewManagedIdentityAccessRuleListCommandFactory returns a managedIdentityAccessRuleListCommand struct.
func NewManagedIdentityAccessRuleListCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return managedIdentityAccessRuleListCommand{
			meta: meta,
		}, nil
	}
}

func (m managedIdentityAccessRuleListCommand) Run(args []string) int {
	m.meta.Logger.Debugf("Starting the 'managed-identity-access-rule list' command with %d arguments:", len(args))
	for ix, arg := range args {
		m.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := m.meta.GetSDKClient()
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return m.doManagedIdentityAccessRuleList(ctx, client, args)
}

func (m managedIdentityAccessRuleListCommand) doManagedIdentityAccessRuleList(ctx context.Context,
	client *tharsis.Client, opts []string,
) int {
	m.meta.Logger.Debugf("will do managed-identity-access-rule list, %d opts", len(opts))

	defs := buildManagedIdentityAccessRuleListDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(m.meta.BinaryName+" managed-identity-access-rule list", defs, opts)
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to parse managed-identity-access-rule list options", err))
		return 1
	}
	if len(cmdArgs) > 0 {
		msg := fmt.Sprintf("excessive managed-identity-access-rule list arguments: %s", cmdArgs)
		m.meta.Logger.Error(output.FormatError(msg, nil), m.HelpManagedIdentityAccessRuleList())
		return 1
	}

	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	managedIdentityPath := getOption("managed-identity-path", "", cmdOpts)[0]

	// Get the managed identity from its path.
	input := &sdktypes.GetManagedIdentityInput{
		Path: &managedIdentityPath,
	}

	m.meta.Logger.Debugf("managed-identity-access-rule list input: %#v", input)

	managedIdentityAccessRules, err := client.ManagedIdentity.GetManagedIdentityAccessRules(ctx, input)
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to list managed identity access rules", err))
		return 1
	}

	return outputManagedIdentityAccessRules(m.meta, toJSON, managedIdentityAccessRules)
}

// buildManagedIdentityAccessRuleListDefs returns defs used by managed-identity-access-rule list command.
func buildManagedIdentityAccessRuleListDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"managed-identity-path": {
			Arguments: []string{"Managed_Identity_Path"},
			Synopsis:  "Resource path to the managed identity.",
			Required:  true,
		},
	}
	return buildJSONOptionDefs(defs)
}

func (m managedIdentityAccessRuleListCommand) Synopsis() string {
	return "List managed identity access rules for a specified managed identity."
}

func (m managedIdentityAccessRuleListCommand) Help() string {
	return m.HelpManagedIdentityAccessRuleList()
}

// HelpManagedIdentityAccessRuleList produces the help string for the 'managed-identity-access-rule list' command.
func (m managedIdentityAccessRuleListCommand) HelpManagedIdentityAccessRuleList() string {
	return fmt.Sprintf(`
Usage: %s [global options] managed-identity-access-rule list [options]

%s

`, m.meta.BinaryName, buildHelpText(buildManagedIdentityAccessRuleListDefs()))
}
