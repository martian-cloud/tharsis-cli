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

// workspaceRemoveMembershipCommand is the top-level structure for the workspace remove-membership command.
type workspaceRemoveMembershipCommand struct {
	meta *Metadata
}

// NewWorkspaceRemoveMembershipCommandFactory returns a workspaceRemoveMembershipCommand struct.
func NewWorkspaceRemoveMembershipCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceRemoveMembershipCommand{
			meta: meta,
		}, nil
	}
}

func (ggc workspaceRemoveMembershipCommand) Run(args []string) int {
	ggc.meta.Logger.Debugf("Starting the 'workspace remove-membership' command with %d arguments:", len(args))
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

	return ggc.doWorkspaceRemoveMembership(ctx, client, args)
}

func (ggc workspaceRemoveMembershipCommand) doWorkspaceRemoveMembership(ctx context.Context, client *tharsis.Client, opts []string) int {
	ggc.meta.Logger.Debugf("will do workspace remove-membership, %d opts", len(opts))

	defs := ggc.buildWorkspaceRemoveMembershipOptionDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(ggc.meta.BinaryName+" workspace remove-membership", defs, opts)
	if err != nil {
		ggc.meta.Logger.Error(output.FormatError("failed to parse workspace remove-membership argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		ggc.meta.Logger.Error(output.FormatError("missing workspace remove-membership ID", nil), ggc.HelpWorkspaceRemoveMembership())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive workspace remove-membership arguments: %s", cmdArgs)
		ggc.meta.Logger.Error(output.FormatError(msg, nil), ggc.HelpWorkspaceRemoveMembership())
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
	ggc.meta.Logger.Debugf("workspace remove-membership input: %#v", input)

	// Remove the membership.
	removedMembership, err := client.NamespaceMembership.DeleteMembership(ctx, input)
	if err != nil {
		ggc.meta.Logger.Error(output.FormatError("failed to remove membership", err))
		return 1
	}

	return outputNamespaceMembership(ggc.meta, toJSON, removedMembership)
}

// buildWorkspaceRemoveMembershipOptionDefs returns the defs used by
// workspace remove-membership command--of which, there are currently none.
func (ggc workspaceRemoveMembershipCommand) buildWorkspaceRemoveMembershipOptionDefs() optparser.OptionDefinitions {
	return buildJSONOptionDefs(optparser.OptionDefinitions{})
}

func (ggc workspaceRemoveMembershipCommand) Synopsis() string {
	return "Remove a membership from a workspace."
}

func (ggc workspaceRemoveMembershipCommand) Help() string {
	return ggc.HelpWorkspaceRemoveMembership()
}

// HelpWorkspaceRemoveMembership prints the help string for the 'workspace remove-membership' command.
func (ggc workspaceRemoveMembershipCommand) HelpWorkspaceRemoveMembership() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace remove-membership [options] <id>

   The workspace remove-membership command removes a membership from a workspace.

%s

`, ggc.meta.BinaryName, buildHelpText(ggc.buildWorkspaceRemoveMembershipOptionDefs()))
}
