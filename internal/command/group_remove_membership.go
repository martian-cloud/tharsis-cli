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

// groupRemoveMembershipCommand is the top-level structure for the group remove-membership command.
type groupRemoveMembershipCommand struct {
	meta *Metadata
}

// NewGroupRemoveMembershipCommandFactory returns a groupRemoveMembershipCommand struct.
func NewGroupRemoveMembershipCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return groupRemoveMembershipCommand{
			meta: meta,
		}, nil
	}
}

func (ggc groupRemoveMembershipCommand) Run(args []string) int {
	ggc.meta.Logger.Debugf("Starting the 'group remove-membership' command with %d arguments:", len(args))
	for ix, arg := range args {
		ggc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := ggc.meta.ReadSettings()
	if err != nil {
		ggc.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		ggc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return ggc.doGroupRemoveMembership(ctx, client, args)
}

func (ggc groupRemoveMembershipCommand) doGroupRemoveMembership(ctx context.Context, client *tharsis.Client, opts []string) int {
	ggc.meta.Logger.Debugf("will do group remove-membership, %d opts", len(opts))

	defs := ggc.buildGroupRemoveMembershipOptionDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(ggc.meta.BinaryName+" group remove-membership", defs, opts)
	if err != nil {
		ggc.meta.Logger.Error(output.FormatError("failed to parse group remove-membership argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		ggc.meta.Logger.Error(output.FormatError("missing group remove-membership ID", nil), ggc.HelpGroupRemoveMembership())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive group remove-membership arguments: %s", cmdArgs)
		ggc.meta.Logger.Error(output.FormatError(msg, nil), ggc.HelpGroupRemoveMembership())
		return 1
	}

	membershipID := cmdArgs[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		ggc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.DeleteNamespaceMembershipInput{
		ID: membershipID,
	}
	ggc.meta.Logger.Debugf("group remove-membership input: %#v", input)

	// Remove the membership.
	removedMembership, err := client.NamespaceMembership.DeleteMembership(ctx, input)
	if err != nil {
		ggc.meta.Logger.Error(output.FormatError("failed to remove membership", err))
		return 1
	}

	return outputNamespaceMembership(ggc.meta, toJSON, removedMembership)
}

// buildGroupRemoveMembershipOptionDefs returns the defs used by
// group remove-membership command--of which, there are currently none.
func (ggc groupRemoveMembershipCommand) buildGroupRemoveMembershipOptionDefs() optparser.OptionDefinitions {
	return buildJSONOptionDefs(optparser.OptionDefinitions{})
}

func (ggc groupRemoveMembershipCommand) Synopsis() string {
	return "Remove a membership from a group."
}

func (ggc groupRemoveMembershipCommand) Help() string {
	return ggc.HelpGroupRemoveMembership()
}

// HelpGroupRemoveMembership prints the help string for the 'group remove-membership' command.
func (ggc groupRemoveMembershipCommand) HelpGroupRemoveMembership() string {
	return fmt.Sprintf(`
Usage: %s [global options] group remove-membership [options] <id>

   The group remove-membership command removes a membership from a group.

%s

`, ggc.meta.BinaryName, buildHelpText(ggc.buildGroupRemoveMembershipOptionDefs()))
}
