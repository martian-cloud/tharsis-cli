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

// groupUpdateCommand is the top-level structure for the group update command.
type groupUpdateCommand struct {
	meta *Metadata
}

// NewGroupUpdateCommandFactory returns a groupUpdateCommand struct.
func NewGroupUpdateCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return groupUpdateCommand{
			meta: meta,
		}, nil
	}
}

func (guc groupUpdateCommand) Run(args []string) int {
	guc.meta.Logger.Debugf("Starting the 'group update' command with %d arguments:", len(args))
	for ix, arg := range args {
		guc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := guc.meta.GetSDKClient()
	if err != nil {
		guc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return guc.doGroupUpdate(ctx, client, args)
}

func (guc groupUpdateCommand) doGroupUpdate(ctx context.Context, client *tharsis.Client, opts []string) int {
	guc.meta.Logger.Debugf("will do group update, %d opts", len(opts))

	defs := buildCommonUpdateOptionDefs("group")
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(guc.meta.BinaryName+" group update", defs, opts)
	if err != nil {
		guc.meta.Logger.Error(output.FormatError("failed to parse group update options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		guc.meta.Logger.Error(output.FormatError("missing group update full path", nil), guc.HelpGroupUpdate())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive group update arguments: %s", cmdArgs)
		guc.meta.Logger.Error(output.FormatError(msg, nil), guc.HelpGroupUpdate())
		return 1
	}

	path := cmdArgs[0]
	description := getOption("description", "", cmdOpts)[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		guc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Error is already logged.
	if !isNamespacePathValid(guc.meta, path) {
		return 1
	}

	input := &sdktypes.UpdateGroupInput{
		GroupPath:   &path,
		Description: description,
	}
	guc.meta.Logger.Debugf("group update input: %#v", input)

	// Update the group.
	updatedGroup, err := client.Group.UpdateGroup(ctx, input)
	if err != nil {
		guc.meta.Logger.Error(output.FormatError("failed to update a group", err))
		return 1
	}

	return outputGroup(guc.meta, toJSON, updatedGroup)
}

func (guc groupUpdateCommand) Synopsis() string {
	return "Update a group."
}

func (guc groupUpdateCommand) Help() string {
	return guc.HelpGroupUpdate()
}

// HelpGroupUpdate produces the help string for the 'group update' command.
func (guc groupUpdateCommand) HelpGroupUpdate() string {
	return fmt.Sprintf(`
Usage: %s [global options] group update [options] <full_path>

   The group update command updates a group. Currently, it
   supports updating the description. Shows final output
   as JSON, if specified.

%s

`, guc.meta.BinaryName, buildHelpText(buildCommonUpdateOptionDefs("group")))
}
