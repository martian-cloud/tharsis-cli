package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// groupDeleteCommand is the top-level structure for the group delete command.
type groupDeleteCommand struct {
	meta *Metadata
}

// NewGroupDeleteCommandFactory returns a groupCommandDelete struct.
func NewGroupDeleteCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return groupDeleteCommand{
			meta: meta,
		}, nil
	}
}

func (gdc groupDeleteCommand) Run(args []string) int {
	gdc.meta.Logger.Debugf("Starting the 'group update' command with %d arguments:", len(args))
	for ix, arg := range args {
		gdc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := gdc.meta.GetSDKClient()
	if err != nil {
		gdc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return gdc.doGroupDelete(ctx, client, args)
}

func (gdc groupDeleteCommand) doGroupDelete(ctx context.Context, client *tharsis.Client, opts []string) int {
	gdc.meta.Logger.Debugf("will do group delete, %d opts", len(opts))

	// No options to parse.
	defs := gdc.buildGroupDeleteDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(gdc.meta.BinaryName+" group delete", defs, opts)
	if err != nil {
		gdc.meta.Logger.Error(output.FormatError("failed to parse group delete options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		gdc.meta.Logger.Error(output.FormatError("missing group delete full path", nil), gdc.HelpGroupDelete())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive group delete arguments: %s", cmdArgs)
		gdc.meta.Logger.Error(output.FormatError(msg, nil), gdc.HelpGroupDelete())
		return 1
	}

	groupPath := cmdArgs[0]
	force, err := getBoolOptionValue("force", "false", cmdOpts)
	if err != nil {
		gdc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Extract path from TRN if needed, then validate path (error is already logged by validation function)
	actualPath := trn.ToPath(groupPath)
	if !isNamespacePathValid(gdc.meta, actualPath) {
		return 1
	}

	// Prepare the inputs.
	// Extract path from TRN if needed - GroupPath field expects paths, not TRNs
	actualPath = trn.ToPath(groupPath)

	input := &sdktypes.DeleteGroupInput{
		GroupPath: &actualPath,
		Force:     &force,
	}
	gdc.meta.Logger.Debugf("group delete input: %#v", input)

	// Delete the group.
	err = client.Group.DeleteGroup(ctx, input)
	if err != nil {
		gdc.meta.Logger.Error(output.FormatError("failed to delete a group", err))
		return 1
	}

	// Cannot show the deleted group, but say something.
	gdc.meta.UI.Output("group delete succeeded.")

	return 0
}

func (groupDeleteCommand) buildGroupDeleteDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"force": {
			Arguments: []string{},
			Synopsis:  "Force the deletion of a group.",
		},
	}
}

func (gdc groupDeleteCommand) Synopsis() string {
	return "Delete a group."
}

func (gdc groupDeleteCommand) Help() string {
	return gdc.HelpGroupDelete()
}

// HelpGroupDelete produces the help string for the 'group delete' command.
func (gdc groupDeleteCommand) HelpGroupDelete() string {
	return fmt.Sprintf(`
Usage: %s [global options] group delete [options] <full_path>

   The group delete command deletes a group.

   Use with caution as deleting a group is irreversible!

%s

`, gdc.meta.BinaryName, buildHelpText(gdc.buildGroupDeleteDefs()))
}
