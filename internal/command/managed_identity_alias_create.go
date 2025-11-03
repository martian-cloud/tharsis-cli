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

// managedIdentityAliasCreateCommand is the top-level structure for the managed-identity-alias create command.
type managedIdentityAliasCreateCommand struct {
	meta *Metadata
}

// NewManagedIdentityAliasCreateCommandFactory returns a managedIdentityAliasCreateCommand struct.
func NewManagedIdentityAliasCreateCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return managedIdentityAliasCreateCommand{
			meta: meta,
		}, nil
	}
}

func (m managedIdentityAliasCreateCommand) Run(args []string) int {
	m.meta.Logger.Debugf("Starting the 'managed-identity-alias create' command with %d arguments:", len(args))
	for ix, arg := range args {
		m.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := m.meta.GetSDKClient()
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return m.doManagedIdentityAliasCreate(ctx, client, args)
}

func (m managedIdentityAliasCreateCommand) doManagedIdentityAliasCreate(ctx context.Context,
	client *tharsis.Client, opts []string,
) int {
	m.meta.Logger.Debugf("will do managed-identity-alias create, %d opts", len(opts))

	defs := buildManagedIdentityAliasCreateDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(m.meta.BinaryName+" managed-identity-alias create", defs, opts)
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to parse managed-identity-alias create options", err))
		return 1
	}
	if len(cmdArgs) > 0 {
		msg := fmt.Sprintf("excessive managed-identity-alias create arguments: %s", cmdArgs)
		m.meta.Logger.Error(output.FormatError(msg, nil), m.HelpManagedIdentityAliasCreate())
		return 1
	}

	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	name := getOption("name", "", cmdOpts)[0]
	sourceID := getOption("alias-source-id", "", cmdOpts)[0]
	sourcePath := getOption("alias-source-path", "", cmdOpts)[0]
	groupPath := getOption("group-path", "", cmdOpts)[0]

	input := &sdktypes.CreateManagedIdentityAliasInput{
		GroupPath: groupPath,
		Name:      name,
	}
	if _, ok := cmdOpts["alias-source-id"]; ok {
		input.AliasSourceID = &sourceID
	}
	if _, ok := cmdOpts["alias-source-path"]; ok {
		input.AliasSourcePath = &sourcePath
	}
	m.meta.Logger.Debugf("managed-identity-alias create input: %#v", input)

	managedIdentityAlias, err := client.ManagedIdentity.CreateManagedIdentityAlias(ctx, input)
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to create managed identity alias", err))
		return 1
	}

	if err = outputManagedIdentity(m.meta, toJSON, managedIdentityAlias); err != nil {
		m.meta.UI.Error(err.Error())
		return 1
	}

	return 0
}

// buildManagedIdentityAliasCreateDefs returns defs used by managed-identity-alias create command.
func buildManagedIdentityAliasCreateDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"name": {
			Arguments: []string{"Managed_Identity_Alias_Name"},
			Synopsis:  "The name of the managed identity alias.",
			Required:  true,
		},
		"alias-source-id": {
			Arguments: []string{""},
			Synopsis:  "The alias source ID.",
		},
		"alias-source-path": {
			Arguments: []string{""},
			Synopsis:  "The alias source path.",
		},
		"group-path": {
			Arguments: []string{"Group_Path"},
			Synopsis:  "Full path of group where the managed identity alias will be created.",
			Required:  true,
		},
	}

	return buildJSONOptionDefs(defs)
}

func (m managedIdentityAliasCreateCommand) Synopsis() string {
	return "Create a new managed identity alias."
}

func (m managedIdentityAliasCreateCommand) Help() string {
	return m.HelpManagedIdentityAliasCreate()
}

// HelpManagedIdentityAliasCreate produces the help string for the 'managed-identity-alias create' command.
func (m managedIdentityAliasCreateCommand) HelpManagedIdentityAliasCreate() string {
	return fmt.Sprintf(`
Usage: %s [global options] managed-identity-alias create [options]

   The managed-identity-alias create command creates a new managed identity alias.

%s

`, m.meta.BinaryName, buildHelpText(buildManagedIdentityAliasCreateDefs()))
}
