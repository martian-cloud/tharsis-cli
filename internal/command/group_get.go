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

// groupGetCommand is the top-level structure for the group get command.
type groupGetCommand struct {
	meta *Metadata
}

// NewGroupGetCommandFactory returns a groupGetCommand struct.
func NewGroupGetCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return groupGetCommand{
			meta: meta,
		}, nil
	}
}

func (ggc groupGetCommand) Run(args []string) int {
	ggc.meta.Logger.Debugf("Starting the 'group get' command with %d arguments:", len(args))
	for ix, arg := range args {
		ggc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := ggc.meta.GetSDKClient()
	if err != nil {
		ggc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return ggc.doGroupGet(ctx, client, args)
}

func (ggc groupGetCommand) doGroupGet(ctx context.Context, client *tharsis.Client, opts []string) int {
	ggc.meta.Logger.Debugf("will do group get, %d opts", len(opts))

	defs := buildJSONOptionDefs(optparser.OptionDefinitions{})
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(ggc.meta.BinaryName+" group get", defs, opts)
	if err != nil {
		ggc.meta.Logger.Error(output.FormatError("failed to parse group get argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		ggc.meta.Logger.Error(output.FormatError("missing group get full path", nil), ggc.HelpGroupGet())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive group get arguments: %s", cmdArgs)
		ggc.meta.Logger.Error(output.FormatError(msg, nil), ggc.HelpGroupGet())
		return 1
	}

	path := cmdArgs[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		ggc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Extract path from TRN if needed, then validate path (error is already logged by validation function)
	actualPath := trn.ToPath(path)
	if !isNamespacePathValid(ggc.meta, actualPath) {
		return 1
	}

	// Convert path to TRN and use ID field - API handles TRN detection automatically
	trnID := trn.ToTRN(path, trn.ResourceTypeGroup)
	input := &sdktypes.GetGroupInput{ID: &trnID}
	ggc.meta.Logger.Debugf("group get input: %#v", input)

	// Get the group.
	foundGroup, err := client.Group.GetGroup(ctx, input)
	if err != nil {
		ggc.meta.Logger.Error(output.FormatError("failed to get a group", err))
		return 1
	}

	return outputGroup(ggc.meta, toJSON, foundGroup)
}

func (ggc groupGetCommand) Synopsis() string {
	return "Get a single group."
}

func (ggc groupGetCommand) Help() string {
	return ggc.HelpGroupGet()
}

// HelpGroupGet prints the help string for the 'group get' command.
func (ggc groupGetCommand) HelpGroupGet() string {
	return fmt.Sprintf(`
Usage: %s [global options] group get [options] <full_path>

   The group get command prints information about one group.

%s

`, ggc.meta.BinaryName, buildHelpText(buildJSONOptionDefs(optparser.OptionDefinitions{})))
}
