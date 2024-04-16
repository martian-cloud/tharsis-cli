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

var (
	// limit on number of users to fetch in one page
	usersPerPage = int32(50)
	// estimate that we're likely looking for a more recently updated user than a less recently updated user
	userSortBy = sdktypes.UserSortableFieldUpdatedAtDesc
)

// groupGetMembershipCommand is the top-level structure for the group get-membership command.
type groupGetMembershipCommand struct {
	meta *Metadata
}

// NewGroupGetMembershipCommandFactory returns a groupGetMembershipCommand struct.
func NewGroupGetMembershipCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return groupGetMembershipCommand{
			meta: meta,
		}, nil
	}
}

func (ggm groupGetMembershipCommand) Run(args []string) int {
	ggm.meta.Logger.Debugf("Starting the 'group get-membership' command with %d arguments:", len(args))
	for ix, arg := range args {
		ggm.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := ggm.meta.ReadSettings()
	if err != nil {
		ggm.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		ggm.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return ggm.doGroupGetMembership(ctx, client, args)
}

func (ggm groupGetMembershipCommand) doGroupGetMembership(ctx context.Context, client *tharsis.Client, opts []string) int {
	ggm.meta.Logger.Debugf("will do group get-membership, %d opts", len(opts))

	defs := ggm.buildGroupGetMembershipOptionDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(ggm.meta.BinaryName+" group get-membership", defs, opts)
	if err != nil {
		ggm.meta.Logger.Error(output.FormatError("failed to parse group get-membership argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		ggm.meta.Logger.Error(output.FormatError("missing group get-membership full path", nil), ggm.HelpGroupGetMembership())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive group get-membership arguments: %s", cmdArgs)
		ggm.meta.Logger.Error(output.FormatError(msg, nil), ggm.HelpGroupGetMembership())
		return 1
	}

	path := cmdArgs[0]
	wantUsername := getOption("username", "", cmdOpts)[0]
	wantServiceAccountID := getOption("service-account-id", "", cmdOpts)[0]
	wantTeamName := getOption("team-name", "", cmdOpts)[0]

	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		ggm.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Error is already logged.
	if !isNamespacePathValid(ggm.meta, path) {
		return 1
	}

	// Query for the group to make sure it exists and is a group.
	_, err = client.Group.GetGroup(ctx, &sdktypes.GetGroupInput{Path: &path})
	if err != nil {
		ggm.meta.UI.Error(output.FormatError("failed to find group", err))
		return 1
	}

	return getNamespaceMembership(ctx, ggm.meta, client, toJSON, wantUsername, wantServiceAccountID, wantTeamName, path)
}

// buildGroupGetMembershipOptionDefs returns the defs used by
// group get-membership command.
func (ggm groupGetMembershipCommand) buildGroupGetMembershipOptionDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"username": {
			Arguments: []string{"Username"},
			Synopsis:  "Username to find the group membership.",
		},
		"service-account-id": {
			Arguments: []string{"Service_Account_ID"},
			Synopsis:  "Service account ID to find the group membership.",
		},
		"team-name": {
			Arguments: []string{"TeamName"},
			Synopsis:  "Team name to find the group membership.",
		},
	}

	return buildJSONOptionDefs(defs)
}

func (ggm groupGetMembershipCommand) Synopsis() string {
	return "Get a group membership."
}

func (ggm groupGetMembershipCommand) Help() string {
	return ggm.HelpGroupGetMembership()
}

// HelpGroupGetMembership prints the help string for the 'group get-membership' command.
func (ggm groupGetMembershipCommand) HelpGroupGetMembership() string {
	return fmt.Sprintf(`
Usage: %s [global options] group get-membership <full_path>

   The group get-membership command gets a group's membership.

%s

   Exactly one of --username, --service-account-id, --team-name is required.

`, ggm.meta.BinaryName, buildHelpText(ggm.buildGroupGetMembershipOptionDefs()))
}

////////////////////////////////////////////////////////////////////////////////

// getNamespaceMembership is shared by group get-membership and workspace get-membership.
func getNamespaceMembership(
	ctx context.Context,
	meta *Metadata,
	client *tharsis.Client,
	toJSON bool,
	wantUsername, wantServiceAccountID, wantTeamName, namespacePath string,
) int {

	// Don't allow multiple user, service account, team options.
	countOfWants := 0
	if wantUsername != "" {
		countOfWants++
	}
	if wantServiceAccountID != "" {
		countOfWants++
	}
	if wantTeamName != "" {
		countOfWants++
	}
	if countOfWants < 1 {
		meta.UI.Error(output.FormatError("one of --username, --service-account-id, --team-name required", nil))
		return 1
	}
	if countOfWants > 1 {
		meta.UI.Error(output.FormatError("only one of --username, --service-account-id, --team-name allowed", nil))
		return 1
	}

	// If a username is specified, find the corresponding user ID.
	var wantUser *sdktypes.User
	if wantUsername != "" {
		userPaginator, err := client.User.GetUserPaginator(ctx, &sdktypes.GetUsersInput{
			Sort: &userSortBy,
			PaginationOptions: &sdktypes.PaginationOptions{
				Limit: &usersPerPage,
			},
			Filter: &sdktypes.UserFilter{
				Search: &wantUsername,
			},
		})
		if err != nil {
			meta.UI.Error(output.FormatError("error looking up user paginator", err))
			return 1
		}

		for (wantUser == nil) && userPaginator.HasMore() {

			// Get a page of userPage.
			userPage, nErr := userPaginator.Next(ctx)
			if nErr != nil {
				meta.UI.Error(output.FormatError("error looking up users by name", nErr))
				return 1
			}

			// Find the user if it's on this page.
			for _, u := range userPage.Users {
				copyU := u
				if copyU.Username == wantUsername {
					wantUser = &copyU
					break
				}
			}
		}

		// If not found yet, try another page.
		if wantUser == nil {
			// No such user.
			meta.UI.Error(output.FormatError("user does not exist: "+wantUsername, err))
			return 1
		}
	}

	// If a team name is specified, find the corresponding team ID.
	var wantTeam *sdktypes.Team
	if wantTeamName != "" {
		var err error

		wantTeam, err = client.Team.GetTeam(ctx, &sdktypes.GetTeamInput{Name: &wantTeamName})
		if err != nil {
			meta.UI.Error(output.FormatError("error trying to find team "+wantTeamName, err))
			return 1
		}
		if wantTeam == nil {
			meta.UI.Error(output.FormatError("team does not exist: "+wantTeamName, err))
			return 1
		}
	}

	// Prepare the inputs.
	input := &sdktypes.GetNamespaceMembershipsInput{
		NamespacePath: namespacePath,
	}
	meta.Logger.Debugf("group list-memberships input: %#v", input)

	// Get the group's memberships.
	allMemberships, err := client.NamespaceMembership.GetMemberships(ctx, input)
	if err != nil {
		meta.Logger.Error(output.FormatError("failed to list a group's memberships", err))
		return 1
	}

	// Return only the specified user, service account, or team.
	// Breaks out of outer loop upon finding the first match.
	var result sdktypes.NamespaceMembership // not a pointer in order to avoid pointer aliasing problems
outerLoop:
	for _, m := range allMemberships {
		switch {
		case m.UserID != nil:
			if wantUser != nil {
				if *m.UserID == wantUser.Metadata.ID {
					result = m
					break outerLoop
				}
			}
		case m.ServiceAccountID != nil:
			if *m.ServiceAccountID == wantServiceAccountID {
				result = m
				break outerLoop
			}
		case m.TeamID != nil:
			if wantTeam != nil {
				if *m.TeamID == wantTeam.Metadata.ID {
					result = m
					break outerLoop
				}
			}
		}
	}

	// Handle if the specified membership is not found.
	// Role's zero value is empty string, so use that to detect whether a member was found.
	if result.Role == "" {
		meta.Logger.Error(output.FormatError("did not find the specified membership", err))
		return 1
	}

	return outputNamespaceMembership(meta, toJSON, &result)
}
