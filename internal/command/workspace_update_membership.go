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

// workspaceUpdateMembershipCommand is the top-level structure for the workspace update-membership command.
type workspaceUpdateMembershipCommand struct {
	meta *Metadata
}

// NewWorkspaceUpdateMembershipCommandFactory returns a workspaceUpdateMembershipCommand struct.
func NewWorkspaceUpdateMembershipCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceUpdateMembershipCommand{
			meta: meta,
		}, nil
	}
}

func (ggc workspaceUpdateMembershipCommand) Run(args []string) int {
	ggc.meta.Logger.Debugf("Starting the 'workspace update-membership' command with %d arguments:", len(args))
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

	return ggc.doWorkspaceUpdateMembership(ctx, client, args)
}

func (ggc workspaceUpdateMembershipCommand) doWorkspaceUpdateMembership(ctx context.Context, client *tharsis.Client, opts []string) int {
	ggc.meta.Logger.Debugf("will do workspace update-membership, %d opts", len(opts))

	defs := ggc.buildWorkspaceUpdateMembershipOptionDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(ggc.meta.BinaryName+" workspace update-membership", defs, opts)
	if err != nil {
		ggc.meta.Logger.Error(output.FormatError("failed to parse workspace update-membership argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		ggc.meta.Logger.Error(output.FormatError("missing workspace update-membership ID", nil), ggc.HelpWorkspaceUpdateMembership())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive workspace update-membership arguments: %s", cmdArgs)
		ggc.meta.Logger.Error(output.FormatError(msg, nil), ggc.HelpWorkspaceUpdateMembership())
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
	ggc.meta.Logger.Debugf("workspace update-membership input: %#v", input)

	// Update the membership.
	updatedMembership, err := client.NamespaceMembership.UpdateMembership(ctx, input)
	if err != nil {
		ggc.meta.Logger.Error(output.FormatError("failed to update membership", err))
		return 1
	}

	return outputNamespaceMembership(ggc.meta, toJSON, updatedMembership)
}

// buildWorkspaceUpdateMembershipOptionDefs returns the defs used by
// workspace update-membership command.
func (ggc workspaceUpdateMembershipCommand) buildWorkspaceUpdateMembershipOptionDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"role": {
			Arguments: []string{"Role"},
			Synopsis:  "New role for the membership.",
			Required:  true,
		},
	}

	return buildJSONOptionDefs(defs)
}

func (ggc workspaceUpdateMembershipCommand) Synopsis() string {
	return "Update a workspace membership."
}

func (ggc workspaceUpdateMembershipCommand) Help() string {
	return ggc.HelpWorkspaceUpdateMembership()
}

// HelpWorkspaceUpdateMembership prints the help string for the 'workspace update-membership' command.
func (ggc workspaceUpdateMembershipCommand) HelpWorkspaceUpdateMembership() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace update-membership [options] <id>

   The workspace update-membership command updates a workspace membership.

%s

`, ggc.meta.BinaryName, buildHelpText(ggc.buildWorkspaceUpdateMembershipOptionDefs()))
}
