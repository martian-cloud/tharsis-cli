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

// workspaceAddMembershipCommand is the top-level structure for the workspace add-membership command.
type workspaceAddMembershipCommand struct {
	meta *Metadata
}

// NewWorkspaceAddMembershipCommandFactory returns a workspaceAddMembershipCommand struct.
func NewWorkspaceAddMembershipCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceAddMembershipCommand{
			meta: meta,
		}, nil
	}
}

func (ggc workspaceAddMembershipCommand) Run(args []string) int {
	ggc.meta.Logger.Debugf("Starting the 'workspace add-membership' command with %d arguments:", len(args))
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

	return ggc.doWorkspaceAddMembership(ctx, client, args)
}

func (ggc workspaceAddMembershipCommand) doWorkspaceAddMembership(ctx context.Context, client *tharsis.Client, opts []string) int {
	ggc.meta.Logger.Debugf("will do workspace add-membership, %d opts", len(opts))

	defs := ggc.buildWorkspaceAddMembershipOptionDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(ggc.meta.BinaryName+" workspace add-membership", defs, opts)
	if err != nil {
		ggc.meta.Logger.Error(output.FormatError("failed to parse workspace add-membership argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		ggc.meta.Logger.Error(output.FormatError("missing workspace add-membership full path", nil), ggc.HelpWorkspaceAddMembership())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive workspace add-membership arguments: %s", cmdArgs)
		ggc.meta.Logger.Error(output.FormatError(msg, nil), ggc.HelpWorkspaceAddMembership())
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

	// Query for the workspace to make sure it exists and is a workspace.
	_, err = client.Workspaces.GetWorkspace(ctx, &sdktypes.GetWorkspaceInput{Path: &path})
	if err != nil {
		ggc.meta.UI.Error(output.FormatError("failed to find workspace", err))
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
	ggc.meta.Logger.Debugf("workspace add-membership input: %#v", input)

	// Add the membership to the workspace.
	addedMembership, err := client.NamespaceMembership.AddMembership(ctx, input)
	if err != nil {
		ggc.meta.Logger.Error(output.FormatError("failed to add membership to a workspace", err))
		return 1
	}

	return outputNamespaceMembership(ggc.meta, toJSON, addedMembership)
}

// buildWorkspaceAddMembershipOptionDefs returns the defs used by
// workspace add-membership command.
func (ggc workspaceAddMembershipCommand) buildWorkspaceAddMembershipOptionDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"username": {
			Arguments: []string{"Username"},
			Synopsis:  "Username for new membership for the workspace.",
		},
		"service-account-id": {
			Arguments: []string{"Service_Account_ID"},
			Synopsis:  "Service account ID for new membership for the workspace.",
		},
		"team-name": {
			Arguments: []string{"TeamName"},
			Synopsis:  "Team name for new membership for the workspace.",
		},
		"role": {
			Arguments: []string{"Role"},
			Synopsis:  "Role for new membership.",
			Required:  true,
		},
	}

	return buildJSONOptionDefs(defs)
}

func (ggc workspaceAddMembershipCommand) Synopsis() string {
	return "Add a membership to a workspace."
}

func (ggc workspaceAddMembershipCommand) Help() string {
	return ggc.HelpWorkspaceAddMembership()
}

// HelpWorkspaceAddMembership prints the help string for the 'workspace add-membership' command.
func (ggc workspaceAddMembershipCommand) HelpWorkspaceAddMembership() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace add-membership [options] <full_path>

   The workspace add-membership command adds a membership to a workspace.

   Note: Supply exactly one of --username, --service-account-id, and --team-name.

%s

`, ggc.meta.BinaryName, buildHelpText(ggc.buildWorkspaceAddMembershipOptionDefs()))
}
