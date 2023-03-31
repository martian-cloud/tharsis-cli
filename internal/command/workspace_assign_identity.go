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

// workspaceAssignManagedIdentityCommand is the top-level structure for the workspace assign-managed-identity command.
type workspaceAssignManagedIdentityCommand struct {
	meta *Metadata
}

// NewWorkspaceAssignManagedIdentityCommandFactory returns a workspaceAssignManagedIdentityCommand struct.
func NewWorkspaceAssignManagedIdentityCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceAssignManagedIdentityCommand{
			meta: meta,
		}, nil
	}
}

func (wam workspaceAssignManagedIdentityCommand) Run(args []string) int {
	wam.meta.Logger.Debugf("Starting the 'workspace assign-managed-identity' command with %d arguments:", len(args))
	for ix, arg := range args {
		wam.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := wam.meta.ReadSettings()
	if err != nil {
		wam.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		wam.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return wam.doWorkspaceAssignManagedIdentity(ctx, client, args)
}

func (wam workspaceAssignManagedIdentityCommand) doWorkspaceAssignManagedIdentity(ctx context.Context, client *tharsis.Client, opts []string) int {
	wam.meta.Logger.Debugf("will do workspace assign-managed-identity, %d opts", len(opts))

	defs := buildJSONOptionDefs(optparser.OptionDefinitions{})
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wam.meta.BinaryName+" workspace assign-managed-identity", defs, opts)
	if err != nil {
		wam.meta.Logger.Error(output.FormatError("failed to parse workspace assign-managed-identity options", err))
		return 1
	}
	if len(cmdArgs) < 2 {
		wam.meta.Logger.Error(output.FormatError("missing workspace assign-managed-identity workspace or full path", nil), wam.HelpWorkspaceAssignManagedIdentity())
		return 1
	}
	if len(cmdArgs) > 2 {
		msg := fmt.Sprintf("excessive workspace assign-managed-identity arguments: %s", cmdArgs)
		wam.meta.Logger.Error(output.FormatError(msg, nil), wam.HelpWorkspaceAssignManagedIdentity())
		return 1
	}

	workspacePath := cmdArgs[0]
	identityPath := cmdArgs[1]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		wam.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Do some basic validation on workspace path.
	if !isNamespacePathValid(wam.meta, workspacePath) {
		return 1
	}

	// Validate managed identity path.
	if !isResourcePathValid(wam.meta, identityPath) {
		return 1
	}

	wam.meta.Logger.Debugf("workspace assign-managed-identity: workspace: %s, managed identity: %s", workspacePath, identityPath)

	workspace, err := assignManagedIdentities(ctx, workspacePath, []string{identityPath}, client)
	if err != nil {
		wam.meta.Logger.Error(output.FormatError("failed to assign managed identity to workspace", err))
		return 1
	}

	return outputWorkspace(wam.meta, toJSON, workspace)
}

// assignManagedIdentities assigns the managed identities and returns the updated workspace.
func assignManagedIdentities(ctx context.Context, workspacePath string, identityPaths []string, client *tharsis.Client) (*sdktypes.Workspace, error) {
	var (
		createdWorkspace *sdktypes.Workspace
		err              error
	)

	// Assign all managed identities to workspace.
	for _, path := range identityPaths {
		pathCopy := path
		createdWorkspace, err = client.ManagedIdentity.AssignManagedIdentityToWorkspace(ctx,
			&sdktypes.AssignManagedIdentityInput{
				ManagedIdentityPath: &pathCopy,
				WorkspacePath:       workspacePath,
			})
		if err != nil {
			return nil, err
		}
	}

	return createdWorkspace, nil
}

func (wam workspaceAssignManagedIdentityCommand) Synopsis() string {
	return "Assign a managed identity to a workspace."
}

func (wam workspaceAssignManagedIdentityCommand) Help() string {
	return wam.HelpWorkspaceAssignManagedIdentity()
}

// HelpWorkspaceAssignManagedIdentity produces the help string for the 'workspace assign-managed-identity' command.
func (wam workspaceAssignManagedIdentityCommand) HelpWorkspaceAssignManagedIdentity() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace assign-managed-identity [options] <workspace> <full_path>

   The workspace assign-managed-identity command designates
   a managed identity to a workspace. Expects two
   arguments: the first being the full path to the target
   workspace and second being the full path to the managed
   identity.

%s

`, wam.meta.BinaryName, buildHelpText(buildJSONOptionDefs(optparser.OptionDefinitions{})))
}
