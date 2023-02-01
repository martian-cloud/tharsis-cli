package command

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tableformatter"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// groupListCommand is the top-level structure for the group list command.
type groupListCommand struct {
	meta *Metadata
}

// NewGroupListCommandFactory returns a groupListCommand struct.
func NewGroupListCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return groupListCommand{
			meta: meta,
		}, nil
	}
}

func (glc groupListCommand) Run(args []string) int {
	glc.meta.Logger.Debugf("Starting the 'group list' command with %d arguments:", len(args))
	for ix, arg := range args {
		glc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := glc.meta.ReadSettings()
	if err != nil {
		glc.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		glc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return glc.doGroupList(ctx, client, args)
}

func (glc groupListCommand) doGroupList(ctx context.Context, client *tharsis.Client, opts []string) int {
	glc.meta.Logger.Debugf("will do group list, %d opts: %#v", len(opts), opts)

	defs := glc.buildGroupListDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(glc.meta.BinaryName+" group list", defs, opts)
	if err != nil {
		glc.meta.Logger.Error(output.FormatError("failed to parse group list options", err))
		return 1
	}
	if len(cmdArgs) > 0 {
		msg := fmt.Sprintf("excessive group list arguments: %s", cmdArgs)
		glc.meta.Logger.Error(output.FormatError(msg, nil), glc.HelpGroupList())
		return 1
	}

	// Extract option values.
	toJSON := getOption("json", "", cmdOpts)[0] == "1"
	cursor := getOption("cursor", "", cmdOpts)[0]
	limit, err := strconv.ParseInt(getOption("limit", "100", cmdOpts)[0], 10, 64) // 100 is the maximum allowed by GraphQL
	if err != nil {
		msg := fmt.Sprintf("invalid limit option value: %s", cmdOpts["limit"])
		glc.meta.Logger.Error(output.FormatError(msg, nil))
		return 1
	}
	limit32 := int32(limit)
	filterPath := getOption("parent-path", "", cmdOpts)[0]

	if filterPath != "" && !isNamespacePathValid(glc.meta, filterPath) {
		return 1
	}

	// Leniently default to ascending order unless instructed otherwise.
	sortOrder := sdktypes.GroupSortableFieldFullPathAsc
	if strings.HasSuffix(strings.ToLower(getOption("sort-order", "", cmdOpts)[0]), "desc") {
		sortOrder = sdktypes.GroupSortableFieldFullPathDesc
	}

	// Prepare the inputs.
	input := &sdktypes.GetGroupsInput{
		Sort: &sortOrder,
		PaginationOptions: &sdktypes.PaginationOptions{
			Cursor: &cursor,
			Limit:  &limit32,
		},
		Filter: &sdktypes.GroupFilter{
			ParentPath: &filterPath,
		},
	}
	if cursor == "" {
		input.PaginationOptions.Cursor = nil
	}
	if filterPath == "" {
		// If not filtering, must send nil.
		input.Filter = nil
	}
	glc.meta.Logger.Debugf("group list input: %#v", input)

	// Get the groups.
	groupsOutput, err := client.Group.GetGroups(ctx, input)
	if err != nil {
		glc.meta.Logger.Error(output.FormatError("failed to get a list of groups", err))
		return 1
	}

	if toJSON {
		buf, err := objectToJSON(groupsOutput)
		if err != nil {
			glc.meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		glc.meta.UI.Output(string(buf))
	} else {
		// Format the output.
		tableInput := make([][]string, len(groupsOutput.Groups)+1)
		tableInput[0] = []string{"id", "name", "description", "full path"}
		for ix, group := range groupsOutput.Groups {
			tableInput[ix+1] = []string{group.Metadata.ID, group.Name, group.Description, group.FullPath}
		}
		glc.meta.UI.Output(tableformatter.FormatTable(tableInput))
		//
		// Must return the new cursor at the end of the list of groups.
		glc.meta.UI.Output(fmt.Sprintf("has next page: %v", groupsOutput.PageInfo.HasNextPage))
		if groupsOutput.PageInfo.HasNextPage {
			// Show the next cursor _ONLY_ if there is a next page.
			glc.meta.UI.Output(fmt.Sprintf("next cursor: %s", groupsOutput.PageInfo.Cursor))
		}
	}

	return 0
}

func (glc groupListCommand) buildGroupListDefs() optparser.OptionDefinitions {
	defs := buildPaginationOptionDefs()

	defs["parent-path"] = &optparser.OptionDefinition{
		Arguments: []string{"Parent_Path"},
		Synopsis:  "Filter to only direct sub-groups of this parent group.",
	}

	return buildJSONOptionDefs(defs)
}

func (glc groupListCommand) Synopsis() string {
	return "List groups."
}

func (glc groupListCommand) Help() string {
	return glc.HelpGroupList()
}

// HelpGroupList returns the help string for the 'group list' command.
func (glc groupListCommand) HelpGroupList() string {
	return fmt.Sprintf(`
Usage: %s [global options] group list [options]

   The group list command prints information about (likely
   multiple) groups. Supports pagination, filtering and
   sorting the output.

   Example:

   %s group list \
      --parent-path top-level/bottom-level \
      --limit 5 \
      --json

   Above command will only show five subgroups under
   top-level/bottom-level parent groups in JSON format.

%s

`,
		glc.meta.BinaryName,
		glc.meta.BinaryName,
		buildHelpText(glc.buildGroupListDefs()),
	)
}

// The End.
