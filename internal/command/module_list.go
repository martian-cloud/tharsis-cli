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

// moduleListCommand is the top-level structure for the module list command.
type moduleListCommand struct {
	meta *Metadata
}

// NewModuleListCommandFactory returns a moduleListCommand struct.
func NewModuleListCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return moduleListCommand{
			meta: meta,
		}, nil
	}
}

func (mlc moduleListCommand) Run(args []string) int {
	mlc.meta.Logger.Debugf("Starting the 'module list' command with %d arguments:", len(args))
	for ix, arg := range args {
		mlc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := mlc.meta.ReadSettings()
	if err != nil {
		mlc.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		mlc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return mlc.doModuleList(ctx, client, args)
}

func (mlc moduleListCommand) doModuleList(ctx context.Context, client *tharsis.Client, opts []string) int {
	mlc.meta.Logger.Debugf("will do module list, %d opts: %#v", len(opts), opts)

	defs := mlc.buildModuleListDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(mlc.meta.BinaryName+" module list", defs, opts)
	if err != nil {
		mlc.meta.Logger.Error(output.FormatError("failed to parse module list options", err))
		return 1
	}
	if len(cmdArgs) > 0 {
		msg := fmt.Sprintf("excessive module list arguments: %s", cmdArgs)
		mlc.meta.Logger.Error(output.FormatError(msg, nil), mlc.HelpModuleList())
		return 1
	}

	// Extract option values.
	toJSON := getOption("json", "", cmdOpts)[0] == "1"
	cursor := getOption("cursor", "", cmdOpts)[0]
	limit, err := strconv.ParseInt(getOption("limit", "100", cmdOpts)[0], 10, 64) // 100 is the maximum allowed by GraphQL
	if err != nil {
		msg := fmt.Sprintf("invalid limit option value: %s", cmdOpts["limit"])
		mlc.meta.Logger.Error(output.FormatError(msg, nil))
		return 1
	}
	limit32 := int32(limit)
	search := getOption("search", "", cmdOpts)[0]
	sortByOption := strings.ToLower(getOption("sort-by", "", cmdOpts)[0])
	sortOrderOption := strings.ToLower(getOption("sort-order", "", cmdOpts)[0])

	// Leniently default to by name unless instructed otherwise.
	sortBy := "name"
	if strings.ToLower(sortByOption) == "updated" {
		sortBy = sortByOption
	}

	// Leniently default to ascending order unless instructed otherwise.
	sortOrder := "asc"
	if strings.HasSuffix(sortOrderOption, "desc") {
		sortOrder = sortOrderOption
	}

	// Decode from 2x2 to 1 of 4.
	var sortable sdktypes.TerraformModuleSortableField
	if sortBy == "name" {
		if sortOrder == "asc" {
			sortable = sdktypes.TerraformModuleSortableFieldNameAsc
		} else {
			sortable = sdktypes.TerraformModuleSortableFieldNameDesc
		}
	} else {
		if sortOrder == "asc" {
			sortable = sdktypes.TerraformModuleSortableFieldUpdatedAtAsc
		} else {
			sortable = sdktypes.TerraformModuleSortableFieldUpdatedAtDesc
		}
	}

	// Prepare the inputs.
	input := &sdktypes.GetTerraformModulesInput{
		Sort: &sortable,
		PaginationOptions: &sdktypes.PaginationOptions{
			Cursor: &cursor,
			Limit:  &limit32,
		},
		Filter: &sdktypes.TerraformModuleFilter{
			Search: &search,
		},
	}

	if cursor == "" {
		input.PaginationOptions.Cursor = nil
	}

	if search == "" {
		// If not filtering, must send nil.
		input.Filter = nil
	}

	mlc.meta.Logger.Debugf("module list input: %#v", input)

	// Get the modules.
	modulesOutput, err := client.TerraformModule.GetModules(ctx, input)
	if err != nil {
		mlc.meta.Logger.Error(output.FormatError("failed to get a list of modules", err))
		return 1
	}

	if toJSON {
		buf, err := objectToJSON(modulesOutput)
		if err != nil {
			mlc.meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		mlc.meta.UI.Output(string(buf))
	} else {
		// Format the output.
		tableInput := make([][]string, len(modulesOutput.TerraformModules)+1)
		tableInput[0] = []string{"id", "name", "resource path", "private", "repository url"}
		for ix, module := range modulesOutput.TerraformModules {
			tableInput[ix+1] = []string{module.Metadata.ID, module.Name, module.ResourcePath,
				fmt.Sprintf("%t", module.Private), module.ResourcePath}
		}
		mlc.meta.UI.Output(tableformatter.FormatTable(tableInput))
		//
		// Must return the new cursor at the end of the list of modules.
		mlc.meta.UI.Output(fmt.Sprintf("has next page: %v", modulesOutput.PageInfo.HasNextPage))
		if modulesOutput.PageInfo.HasNextPage {
			// Show the next cursor _ONLY_ if there is a next page.
			mlc.meta.UI.Output(fmt.Sprintf("next cursor: %s", modulesOutput.PageInfo.Cursor))
		}
	}

	return 0
}

// buildModuleListDefs returns the defs used by the 'module list' command.
func (mlc moduleListCommand) buildModuleListDefs() optparser.OptionDefinitions {
	defs := buildPaginationOptionDefs()

	defs["search"] = &optparser.OptionDefinition{
		Arguments: []string{"Search"},
		Synopsis:  "Filter results to this term.",
	}

	defs["sort-by"] = &optparser.OptionDefinition{
		Arguments: []string{"Sort_By"},
		Synopsis:  "Sort by this field: NAME or UPDATED.",
	}

	return buildJSONOptionDefs(defs)
}

func (mlc moduleListCommand) Synopsis() string {
	return "List modules."
}

func (mlc moduleListCommand) Help() string {
	return mlc.HelpModuleList()
}

// HelpModuleList returns the help string for the 'module list' command.
func (mlc moduleListCommand) HelpModuleList() string {
	return fmt.Sprintf(`
Usage: %s [global options] module list [options]

   The module list command prints information about (likely
   multiple) modules. Supports pagination, filtering and
   sorting the output.

   Example:

   %s module list \
      --search aws \
      --limit 5 \
      --json

   Above command will only show five modules matching the
   search term 'aws' in JSON format.

%s

`,
		mlc.meta.BinaryName,
		mlc.meta.BinaryName,
		buildHelpText(mlc.buildModuleListDefs()),
	)
}

// The End.
