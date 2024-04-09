package command

import (
	"context"
	"fmt"

	"github.com/aws/smithy-go/ptr"
	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tableformatter"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// groupAddMembershipCommand is the top-level structure for the group add-membership command.
type groupAddMembershipCommand struct {
	meta *Metadata
}

// NewGroupAddMembershipCommandFactory returns a groupAddMembershipCommand struct.
func NewGroupAddMembershipCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return groupAddMembershipCommand{
			meta: meta,
		}, nil
	}
}

func (ggc groupAddMembershipCommand) Run(args []string) int {
	ggc.meta.Logger.Debugf("Starting the 'group add-membership' command with %d arguments:", len(args))
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

	return ggc.doGroupAddMembership(ctx, client, args)
}

func (ggc groupAddMembershipCommand) doGroupAddMembership(ctx context.Context, client *tharsis.Client, opts []string) int {
	ggc.meta.Logger.Debugf("will do group add-membership, %d opts", len(opts))

	defs := ggc.buildGroupAddMembershipOptionDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(ggc.meta.BinaryName+" group add-membership", defs, opts)
	if err != nil {
		ggc.meta.Logger.Error(output.FormatError("failed to parse group add-membership argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		ggc.meta.Logger.Error(output.FormatError("missing group add-membership full path", nil), ggc.HelpGroupAddMembership())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive group add-membership arguments: %s", cmdArgs)
		ggc.meta.Logger.Error(output.FormatError(msg, nil), ggc.HelpGroupAddMembership())
		return 1
	}

	path := cmdArgs[0]
	username := getOption("username", "", cmdOpts)[0]
	serviceAccountID := getOption("service-account-id", "", cmdOpts)[0]
	teamName := getOption("team-name", "", cmdOpts)[0]

	role := getOption("role", "", cmdOpts)[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		ggc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Error is already logged.
	if !isNamespacePathValid(ggc.meta, path) {
		return 1
	}

	// Query for the group to make sure it exists and is a group.
	_, err = client.Group.GetGroup(ctx, &sdktypes.GetGroupInput{Path: &path})
	if err != nil {
		ggc.meta.UI.Error(output.FormatError("failed to find group", err))
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.CreateNamespaceMembershipInput{
		NamespacePath: path,
		Role:          role,
	}
	if username != "" {
		input.Username = &username
	}
	if serviceAccountID != "" {
		input.ServiceAccountID = &serviceAccountID
	}
	if teamName != "" {
		input.TeamName = &teamName
	}
	ggc.meta.Logger.Debugf("group add-membership input: %#v", input)

	// Add the membership to the group.
	addedMembership, err := client.NamespaceMembership.AddMembership(ctx, input)
	if err != nil {
		ggc.meta.Logger.Error(output.FormatError("failed to add membership to a group", err))
		return 1
	}

	return outputNamespaceMemberships(ggc.meta, toJSON, []sdktypes.NamespaceMembership{*addedMembership})
}

// buildGroupAddMembershipOptionDefs returns the defs used by
// group add-membership command.
func (ggc groupAddMembershipCommand) buildGroupAddMembershipOptionDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"username": {
			Arguments: []string{"Username"},
			Synopsis:  "Username for new membership for the group.",
		},
		"service-account-id": {
			Arguments: []string{"Service_Account_ID"},
			Synopsis:  "Service account ID for new membership for the group.",
		},
		"team-name": {
			Arguments: []string{"TeamName"},
			Synopsis:  "Team name for new membership for the group.",
		},
		"role": {
			Arguments: []string{"Role"},
			Synopsis:  "Role for new membership.",
			Required:  true,
		},
	}

	return buildJSONOptionDefs(defs)
}

// outputNamespaceMemberships is the final output for most namespace membership operations.
func outputNamespaceMemberships(meta *Metadata, toJSON bool, memberships []sdktypes.NamespaceMembership) int {
	if toJSON {
		buf, err := objectToJSON(memberships)
		if err != nil {
			meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		meta.UI.Output(string(buf))
	} else {
		tableInput := [][]string{
			{"id", "user id", "service account id", "team id", "role"},
		}
		for _, m := range memberships {
			tableInput = append(tableInput,
				[]string{m.Metadata.ID,
					ptr.ToString(m.UserID),
					ptr.ToString(m.ServiceAccountID),
					ptr.ToString(m.TeamID), m.Role,
				},
			)
		}
		meta.UI.Output(tableformatter.FormatTable(tableInput))
	}

	return 0
}

func (ggc groupAddMembershipCommand) Synopsis() string {
	return "Add a membership to a group."
}

func (ggc groupAddMembershipCommand) Help() string {
	return ggc.HelpGroupAddMembership()
}

// HelpGroupAddMembership prints the help string for the 'group add-membership' command.
func (ggc groupAddMembershipCommand) HelpGroupAddMembership() string {
	return fmt.Sprintf(`
Usage: %s [global options] group add-membership [options] <full_path>

   The group add-membership command adds a membership to a group.

   Note: Supply exactly one of --username, --service-account-id, and --team-name.

%s

`, ggc.meta.BinaryName, buildHelpText(ggc.buildGroupAddMembershipOptionDefs()))
}
