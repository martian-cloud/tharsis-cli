package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// workspaceDeleteCommand is the top-level structure for the workspace delete command.
type workspaceDeleteCommand struct {
	meta *Metadata
}

// NewWorkspaceDeleteCommandFactory returns a workspaceDeleteCommand struct.
func NewWorkspaceDeleteCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceDeleteCommand{
			meta: meta,
		}, nil
	}
}

func (wdc workspaceDeleteCommand) Run(args []string) int {
	wdc.meta.Logger.Debugf("Starting the 'workspace delete' command with %d arguments:", len(args))
	for ix, arg := range args {
		wdc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := wdc.meta.GetSDKClient()
	if err != nil {
		wdc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return wdc.doWorkspaceDelete(ctx, client, args)
}

func (wdc workspaceDeleteCommand) doWorkspaceDelete(ctx context.Context, client *tharsis.Client, opts []string) int {
	wdc.meta.Logger.Debugf("will do workspace delete, %d opts", len(opts))

	defs := wdc.buildWorkspaceDeleteDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wdc.meta.BinaryName+" workspace delete", defs, opts)
	if err != nil {
		wdc.meta.Logger.Error(output.FormatError("failed to parse workspace delete options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		wdc.meta.Logger.Error(output.FormatError("missing workspace delete full path", nil), wdc.HelpWorkspaceDelete())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive workspace delete arguments: %s", cmdArgs)
		wdc.meta.Logger.Error(output.FormatError(msg, nil), wdc.HelpWorkspaceDelete())
		return 1
	}

	workspacePath := cmdArgs[0]
	force, err := getBoolOptionValue("force", "false", cmdOpts)
	if err != nil {
		wdc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Extract path from TRN if needed, then validate path (error is already logged by validation function)
	actualPath := trn.ToPath(workspacePath)
	if !isNamespacePathValid(wdc.meta, actualPath) {
		return 1
	}

	// Prepare the inputs - convert path to TRN and use ID field
	trnID := trn.ToTRN(workspacePath, trn.ResourceTypeWorkspace)
	input := &sdktypes.DeleteWorkspaceInput{ID: &trnID, Force: &force}
	wdc.meta.Logger.Debugf("workspace delete input: %#v", input)

	// Delete the workspace.
	err = client.Workspaces.DeleteWorkspace(ctx, input)
	if err != nil {
		wdc.meta.Logger.Error(output.FormatError("failed to delete a workspace", err))
		return 1
	}

	// Cannot show the deleted workspace, but say something.
	wdc.meta.UI.Output("workspace delete succeeded.")

	return 0
}

func (wdc workspaceDeleteCommand) buildWorkspaceDeleteDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"force": {
			Arguments: []string{},
			Synopsis:  "Force the workspace to delete even if resources are deployed.",
		},
	}
}

func (wdc workspaceDeleteCommand) Synopsis() string {
	return "Delete a workspace."
}

func (wdc workspaceDeleteCommand) Help() string {
	return wdc.HelpWorkspaceDelete()
}

// HelpWorkspaceDelete produces the help string for the 'workspace delete' command.
func (wdc workspaceDeleteCommand) HelpWorkspaceDelete() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace delete <full_path>

   The workspace delete command deletes a workspace. Includes
   a force flag to delete the workspace even if resources are
   deployed (dangerous!).

   Use with caution as deleting a workspace is irreversible!

%s

`, wdc.meta.BinaryName, buildHelpText(wdc.buildWorkspaceDeleteDefs()))
}
