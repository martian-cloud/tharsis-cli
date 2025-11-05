package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// Some variables used from the workspace get-membership command.

// workspaceGetMembershipCommand is the top-level structure for the workspace get-membership command.
type workspaceGetMembershipCommand struct {
	meta *Metadata
}

// NewWorkspaceGetMembershipCommandFactory returns a workspaceGetMembershipCommand struct.
func NewWorkspaceGetMembershipCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceGetMembershipCommand{
			meta: meta,
		}, nil
	}
}

func (wgm workspaceGetMembershipCommand) Run(args []string) int {
	wgm.meta.Logger.Debugf("Starting the 'workspace get-membership' command with %d arguments:", len(args))
	for ix, arg := range args {
		wgm.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := wgm.meta.GetSDKClient()
	if err != nil {
		wgm.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return wgm.doWorkspaceGetMembership(ctx, client, args)
}

func (wgm workspaceGetMembershipCommand) doWorkspaceGetMembership(ctx context.Context, client *tharsis.Client, opts []string) int {
	wgm.meta.Logger.Debugf("will do workspace get-membership, %d opts", len(opts))

	defs := wgm.buildWorkspaceGetMembershipOptionDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wgm.meta.BinaryName+" workspace get-membership", defs, opts)
	if err != nil {
		wgm.meta.Logger.Error(output.FormatError("failed to parse workspace get-membership argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		wgm.meta.Logger.Error(output.FormatError("missing workspace get-membership full path", nil), wgm.HelpWorkspaceGetMembership())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive workspace get-membership arguments: %s", cmdArgs)
		wgm.meta.Logger.Error(output.FormatError(msg, nil), wgm.HelpWorkspaceGetMembership())
		return 1
	}

	path := cmdArgs[0]
	wantUsername := getOption("username", "", cmdOpts)[0]
	wantServiceAccountID := getOption("service-account-id", "", cmdOpts)[0]
	wantTeamName := getOption("team-name", "", cmdOpts)[0]

	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		wgm.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Extract path from TRN if needed, then validate path (error is already logged by validation function)
	actualPath := trn.ToPath(path)
	if !isNamespacePathValid(wgm.meta, actualPath) {
		return 1
	}

	// Query for the workspace to make sure it exists and is a workspace.
	trnID := trn.ToTRN(path, trn.ResourceTypeWorkspace)
	getWorkspaceInput := &sdktypes.GetWorkspaceInput{ID: &trnID}
	_, err = client.Workspaces.GetWorkspace(ctx, getWorkspaceInput)
	if err != nil {
		wgm.meta.UI.Error(output.FormatError("failed to find workspace", err))
		return 1
	}

	return getNamespaceMembership(ctx, wgm.meta, client, toJSON, wantUsername, wantServiceAccountID, wantTeamName, path)
}

// buildWorkspaceGetMembershipOptionDefs returns the defs used by
// workspace get-membership command.
func (wgm workspaceGetMembershipCommand) buildWorkspaceGetMembershipOptionDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"username": {
			Arguments: []string{"Username"},
			Synopsis:  "Username to find the workspace membership.",
		},
		"service-account-id": {
			Arguments: []string{"Service_Account_ID"},
			Synopsis:  "Service account ID to find the workspace membership.",
		},
		"team-name": {
			Arguments: []string{"TeamName"},
			Synopsis:  "Team name to find the workspace membership.",
		},
	}

	return buildJSONOptionDefs(defs)
}

func (wgm workspaceGetMembershipCommand) Synopsis() string {
	return "Get a workspace membership."
}

func (wgm workspaceGetMembershipCommand) Help() string {
	return wgm.HelpWorkspaceGetMembership()
}

// HelpWorkspaceGetMembership prints the help string for the 'workspace get-membership' command.
func (wgm workspaceGetMembershipCommand) HelpWorkspaceGetMembership() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace get-membership <full_path>

   The workspace get-membership command gets a workspace's membership.

%s

   Exactly one of --username, --service-account-id, --team-name is required.

`, wgm.meta.BinaryName, buildHelpText(wgm.buildWorkspaceGetMembershipOptionDefs()))
}
