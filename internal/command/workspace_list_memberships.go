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

// workspaceListMembershipsCommand is the top-level structure for the workspace list-memberships command.
type workspaceListMembershipsCommand struct {
	meta *Metadata
}

// NewWorkspaceListMembershipsCommandFactory returns a workspaceListMembershipsCommand struct.
func NewWorkspaceListMembershipsCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceListMembershipsCommand{
			meta: meta,
		}, nil
	}
}

func (wlm workspaceListMembershipsCommand) Run(args []string) int {
	wlm.meta.Logger.Debugf("Starting the 'workspace list-memberships' command with %d arguments:", len(args))
	for ix, arg := range args {
		wlm.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := wlm.meta.GetSDKClient()
	if err != nil {
		wlm.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return wlm.doWorkspaceListMemberships(ctx, client, args)
}

func (wlm workspaceListMembershipsCommand) doWorkspaceListMemberships(ctx context.Context, client *tharsis.Client, opts []string) int {
	wlm.meta.Logger.Debugf("will do workspace list-memberships, %d opts", len(opts))

	defs := wlm.buildWorkspaceListMembershipsOptionDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wlm.meta.BinaryName+" workspace list-memberships", defs, opts)
	if err != nil {
		wlm.meta.Logger.Error(output.FormatError("failed to parse workspace list-memberships argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		wlm.meta.Logger.Error(output.FormatError("missing workspace list-memberships full path", nil), wlm.HelpWorkspaceListMemberships())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive workspace list-memberships arguments: %s", cmdArgs)
		wlm.meta.Logger.Error(output.FormatError(msg, nil), wlm.HelpWorkspaceListMemberships())
		return 1
	}

	path := cmdArgs[0]

	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		wlm.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Extract path from TRN if needed, then validate path (error is already logged by validation function)
	actualPath := trn.ToPath(path)
	if !isNamespacePathValid(wlm.meta, actualPath) {
		return 1
	}

	// Query for the workspace to make sure it exists and is a workspace.
	trnID := trn.ToTRN(path, trn.ResourceTypeWorkspace)
	_, err = client.Workspaces.GetWorkspace(ctx, &sdktypes.GetWorkspaceInput{ID: &trnID})
	if err != nil {
		wlm.meta.UI.Error(output.FormatError("failed to find workspace", err))
		return 1
	}

	// Prepare the inputs.
	// Extract path from TRN if needed - NamespacePath field expects paths, not TRNs
	actualPath = trn.ToPath(path)
	
	input := &sdktypes.GetNamespaceMembershipsInput{
		NamespacePath: actualPath,
	}
	wlm.meta.Logger.Debugf("workspace list-memberships input: %#v", input)

	// Get the workspace's memberships.
	foundMemberships, err := client.NamespaceMembership.GetMemberships(ctx, input)
	if err != nil {
		wlm.meta.Logger.Error(output.FormatError("failed to list a workspace's memberships", err))
		return 1
	}

	return outputNamespaceMemberships(wlm.meta, toJSON, foundMemberships)
}

// buildWorkspaceListMembershipsOptionDefs returns the defs used by
// workspace list-memberships command.
func (wlm workspaceListMembershipsCommand) buildWorkspaceListMembershipsOptionDefs() optparser.OptionDefinitions {
	return buildJSONOptionDefs(optparser.OptionDefinitions{})
}

func (wlm workspaceListMembershipsCommand) Synopsis() string {
	return "List a workspace's memberships."
}

func (wlm workspaceListMembershipsCommand) Help() string {
	return wlm.HelpWorkspaceListMemberships()
}

// HelpWorkspaceListMemberships prints the help string for the 'workspace list-memberships' command.
func (wlm workspaceListMembershipsCommand) HelpWorkspaceListMemberships() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace list-memberships <full_path>

   The workspace list-memberships command lists a workspace's memberships.

%s

`, wlm.meta.BinaryName, buildHelpText(wlm.buildWorkspaceListMembershipsOptionDefs()))
}
