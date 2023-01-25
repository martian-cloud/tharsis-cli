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

// workspaceListCommand is the top-level structure for the workspace command.
type workspaceListCommand struct {
	meta *Metadata
}

// NewWorkspaceListCommandFactory returns a workspaceCommandList struct.
func NewWorkspaceListCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return workspaceListCommand{
			meta: meta,
		}, nil
	}
}

func (wlc workspaceListCommand) Run(args []string) int {
	wlc.meta.Logger.Debugf("Starting the 'workspace list' command with %d arguments:", len(args))
	for ix, arg := range args {
		wlc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := wlc.meta.ReadSettings()
	if err != nil {
		wlc.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		wlc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return wlc.doWorkspaceList(ctx, client, args)
}

func (wlc workspaceListCommand) doWorkspaceList(ctx context.Context, client *tharsis.Client, opts []string) int {
	wlc.meta.Logger.Debugf("will do workspace list, %d opts: %#v", len(opts), opts)

	defs := wlc.buildWorkspaceListDefs()

	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(wlc.meta.BinaryName+" workspace list", defs, opts)
	if err != nil {
		wlc.meta.Logger.Error(output.FormatError("failed to parse workspace list options", err))
		return 1
	}
	if len(cmdArgs) > 0 {
		msg := fmt.Sprintf("excessive workspace list arguments: %s", cmdArgs)
		wlc.meta.Logger.Error(output.FormatError(msg, nil), wlc.HelpWorkspaceList())
		return 1
	}

	// Extract option values.
	toJSON := getOption("json", "", cmdOpts)[0] == "1"
	cursor := getOption("cursor", "", cmdOpts)[0]
	limit, err := strconv.ParseInt(getOption("limit", "100", cmdOpts)[0], 10, 64) // 100 is the maximum allowed by GraphQL
	if err != nil {
		msg := fmt.Sprintf("invalid limit option value: %s", cmdOpts["limit"])
		wlc.meta.Logger.Error(output.FormatError(msg, nil))
		return 1
	}
	limit32 := int32(limit)
	filterPath := getOption("group-path", "", cmdOpts)[0]

	// Error is already logged.
	if filterPath != "" && !isNamespacePathValid(wlc.meta, filterPath) {
		return 1
	}

	// Leniently default to by full path unless instructed otherwise.
	sortBy := "path"
	if strings.ToLower(getOption("sort-by", "", cmdOpts)[0]) == "updated" {
		sortBy = "updated"
	}
	// Leniently default to ascending order unless instructed otherwise.
	sortOrder := "asc"
	if strings.HasSuffix(strings.ToLower(getOption("sort-order", "", cmdOpts)[0]), "desc") {
		sortOrder = "desc"
	}
	// Decode from 2x2 to 1 of 4.
	var sortable sdktypes.WorkspaceSortableField
	if sortBy == "path" {
		if sortOrder == "asc" {
			sortable = sdktypes.WorkspaceSortableFieldFullPathAsc
		} else {
			sortable = sdktypes.WorkspaceSortableFieldFullPathDesc
		}
	} else {
		if sortOrder == "asc" {
			sortable = sdktypes.WorkspaceSortableFieldUpdatedAtAsc
		} else {
			sortable = sdktypes.WorkspaceSortableFieldUpdatedAtDesc
		}
	}

	// Prepare the inputs.
	input := &sdktypes.GetWorkspacesInput{
		Sort: &sortable,
		PaginationOptions: &sdktypes.PaginationOptions{
			Cursor: &cursor,
			Limit:  &limit32,
		},
		Filter: &sdktypes.WorkspaceFilter{
			GroupPath: &filterPath,
		},
	}
	if cursor == "" {
		input.PaginationOptions.Cursor = nil
	}
	if filterPath == "" {
		// If not filtering, must send nil.
		input.Filter = nil
	}
	wlc.meta.Logger.Debugf("workspace list input: %#v", input)

	// Get the workspaces.
	workspacesOutput, err := client.Workspaces.GetWorkspaces(ctx, input)
	if err != nil {
		wlc.meta.Logger.Error(output.FormatError("failed to get a list of workspaces", err))
		return 1
	}

	if toJSON {
		buf, err := objectToJSON(workspacesOutput)
		if err != nil {
			wlc.meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		wlc.meta.UI.Output(string(buf))
	} else {
		// Format the output.
		tableInput := make([][]string, len(workspacesOutput.Workspaces)+1)
		tableInput[0] = []string{"name", "fullPath", "description", "id"}
		for ix, workspace := range workspacesOutput.Workspaces {
			tableInput[ix+1] = []string{workspace.Name, workspace.FullPath,
				workspace.Description, workspace.Metadata.ID}
		}
		wlc.meta.UI.Output(tableformatter.FormatTable(tableInput))
		//
		// Must return the new cursor at the end of the list of workspaces.
		wlc.meta.UI.Output(fmt.Sprintf("has next page: %v", workspacesOutput.PageInfo.HasNextPage))
		if workspacesOutput.PageInfo.HasNextPage {
			// Show the next cursor _ONLY_ if there is a next page.
			wlc.meta.UI.Output(fmt.Sprintf("next cursor: %s", workspacesOutput.PageInfo.Cursor))
		}
	}

	return 0
}

func (wlc workspaceListCommand) buildWorkspaceListDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"cursor": {
			Arguments: []string{"Cursor_String"},
			Synopsis:  "The cursor string for manual pagination.",
		},
		"limit": {
			Arguments: []string{"count"},
			Synopsis:  "Maximum number of result elements to return.",
		},
		"group-path": {
			Arguments: []string{"Group_Path"},
			Synopsis:  "Filter to only workspaces in this group.",
		},
		"sort-by": {
			Arguments: []string{"Sort_By"},
			Synopsis:  "Sort by this field: PATH or UPDATED.",
		},
		"sort-order": {
			Arguments: []string{"Sort_Order"},
			Synopsis:  "Sort in this direction, ASC or DESC.",
		},
	}

	return buildJSONOptionDefs(defs)
}

func (wlc workspaceListCommand) Synopsis() string {
	return "List workspaces."
}

func (wlc workspaceListCommand) Help() string {
	return wlc.HelpWorkspaceList()
}

// HelpWorkspaceList produces the help string for the 'workspace list' command.
func (wlc workspaceListCommand) HelpWorkspaceList() string {
	return fmt.Sprintf(`
Usage: %s [global options] workspace list [options]

   The workspace list command lists (likely multiple)
   workspaces. Supports pagination, filtering and sorting
   the output.

   Example:

   %s workspace list \
      --group-path top-level/bottom-level \
      --limit 5 \
      --json

   Above command will only show five workspaces under
   top-level/bottom-level parent groups in JSON format.

%s

`,
		wlc.meta.BinaryName,
		wlc.meta.BinaryName,
		buildHelpText(wlc.buildWorkspaceListDefs()),
	)
}

// The End.
