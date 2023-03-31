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

// workspaceUnassignManagedIdentityCommand is the top-level structure for the workspace unassign-managed-identity command.
type workspaceUnassignManagedIdentityCommand struct {
	meta *Metadata
}

// NewWorkspaceUnassignManagedIdentityCommandFactory returns a workspaceUnassignManagedIdentityCommand struct.
func NewWorkspaceUnassignManagedIdentityCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceUnassignManagedIdentityCommand{
			meta: meta,
		}, nil
	}
}

func (wam workspaceUnassignManagedIdentityCommand) Run(args []string) int {
	wam.meta.Logger.Debugf("Starting the 'workspace unassign-managed-identity' command with %d arguments:", len(args))
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

	return wam.doWorkspaceUnassignManagedIdentity(ctx, client, args)
}

func (wam workspaceUnassignManagedIdentityCommand) doWorkspaceUnassignManagedIdentity(ctx context.Context, client *tharsis.Client, opts []string) int {
	wam.meta.Logger.Debugf("will do workspace unassign-managed-identity, %d opts", len(opts))

	defs := buildJSONOptionDefs(optparser.OptionDefinitions{})
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wam.meta.BinaryName+" workspace unassign-managed-identity", defs, opts)
	if err != nil {
		wam.meta.Logger.Error(output.FormatError("failed to parse workspace unassign-managed-identity options", err))
		return 1
	}
	if len(cmdArgs) < 2 {
		wam.meta.Logger.Error(output.FormatError("missing workspace unassign-managed-identity workspace or full path", nil), wam.HelpWorkspaceUnassignManagedIdentity())
		return 1
	}
	if len(cmdArgs) > 2 {
		msg := fmt.Sprintf("excessive workspace unassign-managed-identity arguments: %s", cmdArgs)
		wam.meta.Logger.Error(output.FormatError(msg, nil), wam.HelpWorkspaceUnassignManagedIdentity())
		return 1
	}

	workspacePath := cmdArgs[0]
	identityPath := cmdArgs[1]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		wam.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Do some basic validation on paths.
	if !isNamespacePathValid(wam.meta, workspacePath) {
		return 1
	}

	// Validate identity path.
	if !isResourcePathValid(wam.meta, identityPath) {
		return 1
	}

	input := &sdktypes.AssignManagedIdentityInput{
		ManagedIdentityPath: &identityPath,
		WorkspacePath:       workspacePath,
	}

	wam.meta.Logger.Debugf("workspace unassign-managed-identity input: %#v", input)

	workspace, err := client.ManagedIdentity.UnassignManagedIdentityFromWorkspace(ctx, input)
	if err != nil {
		wam.meta.Logger.Error("failed to unassign managed identity from workspace", err)
		return 1
	}

	return outputWorkspace(wam.meta, toJSON, workspace)
}

func (wam workspaceUnassignManagedIdentityCommand) Synopsis() string {
	return "Unassign a managed identity from a workspace."
}

func (wam workspaceUnassignManagedIdentityCommand) Help() string {
	return wam.HelpWorkspaceUnassignManagedIdentity()
}

// HelpWorkspaceUnassignManagedIdentity produces the help string
// for the 'workspace unassign-managed-identity' command.
func (wam workspaceUnassignManagedIdentityCommand) HelpWorkspaceUnassignManagedIdentity() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace unassign-managed-identity [options] <workspace> <full_path>

   The workspace unassign-managed-identity command revokes
   a managed identity from a workspace. Expects two
   arguments: the first being the full path to the target
   workspace and second being the full path to the managed
   identity.

%s

`, wam.meta.BinaryName, buildHelpText(buildJSONOptionDefs(optparser.OptionDefinitions{})))
}
