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

// groupMigrateCommand is the top-level structure for the group migrate command.
type groupMigrateCommand struct {
	meta *Metadata
}

// NewGroupMigrateCommandFactory returns a groupMigrateCommand struct.
func NewGroupMigrateCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return groupMigrateCommand{
			meta: meta,
		}, nil
	}
}

func (gmc groupMigrateCommand) Run(args []string) int {
	gmc.meta.Logger.Debugf("Starting the 'group migrate' command with %d arguments:", len(args))
	for ix, arg := range args {
		gmc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := gmc.meta.ReadSettings()
	if err != nil {
		gmc.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		gmc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return gmc.doGroupMigrate(ctx, client, args)
}

func (gmc groupMigrateCommand) doGroupMigrate(ctx context.Context, client *tharsis.Client, opts []string) int {
	gmc.meta.Logger.Debugf("will do group migrate, %d opts", len(opts))

	defs := gmc.buildGroupMigrateOptionDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(gmc.meta.BinaryName+" group migrate", defs, opts)
	if err != nil {
		gmc.meta.Logger.Error(output.FormatError("failed to parse group migrate options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		gmc.meta.Logger.Error(output.FormatError("missing group migrate full path", nil), gmc.HelpGroupMigrate())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive group migrate arguments: %s", cmdArgs)
		gmc.meta.Logger.Error(output.FormatError(msg, nil), gmc.HelpGroupMigrate())
		return 1
	}

	path := cmdArgs[0]
	newParent := getOption("new-parent-path", "", cmdOpts)[0]
	toTopLevel, err := getBoolOptionValue("to-top-level", "false", cmdOpts)
	if err != nil {
		gmc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		gmc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Error is already logged.
	if !isNamespacePathValid(gmc.meta, path) {
		return 1
	}

	// Check that options are consistent.
	if (newParent == "") && !toTopLevel {
		gmc.meta.Logger.Error(output.FormatError("Must supply either --new-parent-path or --to-top-level", nil))
		return 1
	}
	if (newParent != "") && toTopLevel {
		gmc.meta.Logger.Error(output.FormatError("Must supply only one of --new-parent-path and --to-top-level", nil))
		return 1
	}

	// Prepare the inputs.
	var newParentPath *string
	if newParent != "" {
		newParentPath = &newParent
	}
	input := &sdktypes.MigrateGroupInput{
		GroupPath:     path,
		NewParentPath: newParentPath,
	}
	gmc.meta.Logger.Debugf("group migrate input: %#v", input)

	// Migrate the group.
	migratedGroup, err := client.Group.MigrateGroup(ctx, input)
	if err != nil {
		gmc.meta.Logger.Error(output.FormatError("failed to migrate a group", err))
		return 1
	}

	return outputGroup(gmc.meta, toJSON, migratedGroup)
}

// buildGroupMigrateOptionDefs returns the defs used by
// group migrate command.
func (gmc groupMigrateCommand) buildGroupMigrateOptionDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"new-parent-path": {
			Arguments: []string{"New_Parent"},
			Synopsis:  "New parent path for the group.",
		},
		"to-top-level": {
			Arguments: []string{}, // zero arguments means it's a bool with no argument
			Synopsis:  "Migrate group to top-level.",
		},
	}

	return buildJSONOptionDefs(defs)
}

func (gmc groupMigrateCommand) Synopsis() string {
	return "Migrate a group."
}

func (gmc groupMigrateCommand) Help() string {
	return gmc.HelpGroupMigrate()
}

// HelpGroupMigrate produces the help string for the 'group migrate' command.
func (gmc groupMigrateCommand) HelpGroupMigrate() string {
	return fmt.Sprintf(`
Usage: %s [global options] group migrate [options] <full_path>

   The group migrate command migrates a group to another parent group
	 or to top-level.

%s

`, gmc.meta.BinaryName, buildHelpText(gmc.buildGroupMigrateOptionDefs()))
}

// The End.
