package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tableformatter"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// groupCreateCommand is the top-level structure for the group create command.
type groupCreateCommand struct {
	meta *Metadata
}

// NewGroupCreateCommandFactory returns a groupCreateCommand struct.
func NewGroupCreateCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return groupCreateCommand{
			meta: meta,
		}, nil
	}
}

func (gcc groupCreateCommand) Run(args []string) int {
	gcc.meta.Logger.Debugf("Starting the 'group create' command with %d arguments:", len(args))
	for ix, arg := range args {
		gcc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := gcc.meta.GetSDKClient()
	if err != nil {
		gcc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return gcc.doGroupCreate(ctx, client, args)
}

func (gcc groupCreateCommand) doGroupCreate(ctx context.Context, client *tharsis.Client, opts []string) int {
	gcc.meta.Logger.Debugf("will do group create, %d opts", len(opts))

	defs := buildCommonCreateOptionDefs("group")
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(gcc.meta.BinaryName+" group create", defs, opts)
	if err != nil {
		gcc.meta.Logger.Error(output.FormatError("failed to parse group create options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		gcc.meta.Logger.Error(output.FormatError("missing group create full path", nil), gcc.HelpGroupCreate())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive group create arguments: %s", cmdArgs)
		gcc.meta.Logger.Error(output.FormatError(msg, nil), gcc.HelpGroupCreate())
		return 1
	}

	groupPath := cmdArgs[0]
	description := getOption("description", "", cmdOpts)[0]
	ifNotExists, err := getBoolOptionValue("if-not-exists", "false", cmdOpts)
	if err != nil {
		gcc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		gcc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	var name, parentPath string

	// Extract path from TRN if needed, then validate path (error is already logged by validation function)
	actualPath := trn.ToPath(groupPath)
	if !isNamespacePathValid(gcc.meta, actualPath) {
		return 1
	}

	// Prepare the inputs.
	index := strings.LastIndex(groupPath, sep)
	if index == -1 {
		// This is a new top-level group
		name = groupPath
	}

	// Check if group already exists.
	if ifNotExists {
		trnID := trn.ToTRN(groupPath, trn.ResourceTypeGroup)
		getGroupInput := &sdktypes.GetGroupInput{ID: &trnID}
		group, gErr := client.Group.GetGroup(ctx, getGroupInput)
		if (gErr != nil) && !tharsis.IsNotFoundError(gErr) {
			gcc.meta.Logger.Error(output.FormatError("failed to check group", gErr))
			return 1
		}

		if group != nil {
			return outputGroup(gcc.meta, toJSON, group)
		}
	}

	// If name is empty, then parent path exists.
	if name == "" {
		name = groupPath[index+1:]
		parentPath = groupPath[:index]
	}

	input := &sdktypes.CreateGroupInput{
		Name:        name,
		ParentPath:  &parentPath,
		Description: description,
	}
	if parentPath == "" {
		// Empty parent ID is turned into nil to signify this is a new top-level group.
		input.ParentPath = nil
	}
	gcc.meta.Logger.Debugf("group create input: %#v", input)

	// Create the group.
	createdGroup, err := client.Group.CreateGroup(ctx, input)
	if err != nil {
		gcc.meta.Logger.Error(output.FormatError("failed to create a group", err))
		return 1
	}

	return outputGroup(gcc.meta, toJSON, createdGroup)
}

// outputGroup is the final output for most group operations.
func outputGroup(meta *Metadata, toJSON bool, group *sdktypes.Group) int {
	if toJSON {
		buf, err := objectToJSON(group)
		if err != nil {
			meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		meta.UI.Output(string(buf))
	} else {
		tableInput := [][]string{
			{"id", "name", "description", "full path"},
			{group.Metadata.ID, group.Name, group.Description, group.FullPath},
		}
		meta.UI.Output(tableformatter.FormatTable(tableInput))
	}

	return 0
}

func (gcc groupCreateCommand) Synopsis() string {
	return "Create a new group."
}

func (gcc groupCreateCommand) Help() string {
	return gcc.HelpGroupCreate()
}

// HelpGroupCreate produces the help string for the 'group create' command.
func (gcc groupCreateCommand) HelpGroupCreate() string {
	return fmt.Sprintf(`
Usage: %s [global options] group create [options] <full_path>

   The group create command creates a new group. It allows
   setting a group's description (optional). Shows final
   output as JSON, if specified. Idempotent when used with
   --if-not-exists option.

%s

`, gcc.meta.BinaryName, buildHelpText(buildCommonCreateOptionDefs("group")))
}
