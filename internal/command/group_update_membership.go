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

// groupUpdateMembershipCommand is the top-level structure for the group update-membership command.
type groupUpdateMembershipCommand struct {
	meta *Metadata
}

// NewGroupUpdateMembershipCommandFactory returns a groupUpdateMembershipCommand struct.
func NewGroupUpdateMembershipCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return groupUpdateMembershipCommand{
			meta: meta,
		}, nil
	}
}

func (ggc groupUpdateMembershipCommand) Run(args []string) int {
	ggc.meta.Logger.Debugf("Starting the 'group update-membership' command with %d arguments:", len(args))
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

	return ggc.doGroupUpdateMembership(ctx, client, args)
}

func (ggc groupUpdateMembershipCommand) doGroupUpdateMembership(ctx context.Context, client *tharsis.Client, opts []string) int {
	ggc.meta.Logger.Debugf("will do group update-membership, %d opts", len(opts))

	defs := ggc.buildGroupUpdateMembershipOptionDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(ggc.meta.BinaryName+" group update-membership", defs, opts)
	if err != nil {
		ggc.meta.Logger.Error(output.FormatError("failed to parse group update-membership argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		ggc.meta.Logger.Error(output.FormatError("missing group update-membership ID", nil), ggc.HelpGroupUpdateMembership())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive group update-membership arguments: %s", cmdArgs)
		ggc.meta.Logger.Error(output.FormatError(msg, nil), ggc.HelpGroupUpdateMembership())
		return 1
	}

	membershipID := cmdArgs[0]
	role := getOption("role", "", cmdOpts)[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		ggc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.UpdateNamespaceMembershipInput{
		ID:   membershipID,
		Role: role,
	}
	ggc.meta.Logger.Debugf("group update-membership input: %#v", input)

	// Update the membership.
	updatedMembership, err := client.NamespaceMembership.UpdateMembership(ctx, input)
	if err != nil {
		ggc.meta.Logger.Error(output.FormatError("failed to update membership", err))
		return 1
	}

	return outputNamespaceMembership(ggc.meta, toJSON, updatedMembership)
}

// buildGroupUpdateMembershipOptionDefs returns the defs used by
// group update-membership command.
func (ggc groupUpdateMembershipCommand) buildGroupUpdateMembershipOptionDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"role": {
			Arguments: []string{"Role"},
			Synopsis:  "New role for the membership.",
			Required:  true,
		},
	}

	return buildJSONOptionDefs(defs)
}

func (ggc groupUpdateMembershipCommand) Synopsis() string {
	return "Update a group membership."
}

func (ggc groupUpdateMembershipCommand) Help() string {
	return ggc.HelpGroupUpdateMembership()
}

// HelpGroupUpdateMembership prints the help string for the 'group update-membership' command.
func (ggc groupUpdateMembershipCommand) HelpGroupUpdateMembership() string {
	return fmt.Sprintf(`
Usage: %s [global options] group update-membership [options] <id>

   The group update-membership command updates a group membership.

%s

`, ggc.meta.BinaryName, buildHelpText(ggc.buildGroupUpdateMembershipOptionDefs()))
}
