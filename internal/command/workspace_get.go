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

// workspaceGetCommand is the top-level structure for the workspace get command.
type workspaceGetCommand struct {
	meta *Metadata
}

// NewWorkspaceGetCommandFactory returns a workspaceCommandGet struct.
func NewWorkspaceGetCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceGetCommand{
			meta: meta,
		}, nil
	}
}

func (wgc workspaceGetCommand) Run(args []string) int {
	wgc.meta.Logger.Debugf("Starting the 'workspace get' command with %d arguments:", len(args))
	for ix, arg := range args {
		wgc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := wgc.meta.GetSDKClient()
	if err != nil {
		wgc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return wgc.doWorkspaceGet(ctx, client, args)
}

func (wgc workspaceGetCommand) doWorkspaceGet(ctx context.Context, client *tharsis.Client, opts []string) int {
	wgc.meta.Logger.Debugf("will do workspace get, %d opts", len(opts))

	// Get one argument, no options allowed.
	defs := buildJSONOptionDefs(optparser.OptionDefinitions{})
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wgc.meta.BinaryName+" workspace get", defs, opts)
	if err != nil {
		wgc.meta.Logger.Error(output.FormatError("failed to parse workspace get argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		wgc.meta.Logger.Error(output.FormatError("missing workspace get full path", nil), wgc.HelpWorkspaceGet())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive workspace get arguments: %s", cmdArgs)
		wgc.meta.Logger.Error(output.FormatError(msg, nil), wgc.HelpWorkspaceGet())
		return 1
	}

	workspacePath := cmdArgs[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		wgc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Error is already logged.
	if !isNamespacePathValid(wgc.meta, workspacePath) {
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.GetWorkspaceInput{Path: &workspacePath}
	wgc.meta.Logger.Debugf("workspace get input: %#v", input)

	// Get the workspace.
	foundWorkspace, err := client.Workspaces.GetWorkspace(ctx, input)
	if err != nil {
		wgc.meta.Logger.Error(output.FormatError("failed to get a workspace", err))
		return 1
	}

	return outputWorkspace(wgc.meta, toJSON, foundWorkspace)
}

func (wgc workspaceGetCommand) Synopsis() string {
	return "Get a single workspace."
}

func (wgc workspaceGetCommand) Help() string {
	return wgc.HelpWorkspaceGet()
}

// HelpWorkspaceGet prints the help string for the 'workspace get' command.
func (wgc workspaceGetCommand) HelpWorkspaceGet() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace get [options] <full_path>

   The workspace get command prints information about one
   workspace.

%s

`, wgc.meta.BinaryName, buildHelpText(buildJSONOptionDefs(optparser.OptionDefinitions{})))
}
