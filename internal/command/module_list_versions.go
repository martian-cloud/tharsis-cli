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

// moduleListVersionsCommand is the top-level structure for the module list-versions command.
type moduleListVersionsCommand struct {
	meta *Metadata
}

// NewModuleListVersionsCommandFactory returns a moduleListVersionsCommand struct.
func NewModuleListVersionsCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return moduleListVersionsCommand{
			meta: meta,
		}, nil
	}
}

func (mlc moduleListVersionsCommand) Run(args []string) int {
	mlc.meta.Logger.Debugf("Starting the 'module list-versions' command with %d arguments:", len(args))
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

	return mlc.doModuleListVersions(ctx, client, args)
}

func (mlc moduleListVersionsCommand) doModuleListVersions(ctx context.Context, client *tharsis.Client, opts []string) int {
	mlc.meta.Logger.Debugf("will do module list-versions, %d opts: %#v", len(opts), opts)

	defs := mlc.buildModuleListVersionsDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(mlc.meta.BinaryName+" module list-versions", defs, opts)
	if err != nil {
		mlc.meta.Logger.Error(output.FormatError("failed to parse module list-versions options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		mlc.meta.Logger.Error(output.FormatError("missing module list-versions module path", nil), mlc.HelpModuleListVersions())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive module list-versions arguments: %s", cmdArgs)
		mlc.meta.Logger.Error(output.FormatError(msg, nil), mlc.HelpModuleListVersions())
		return 1
	}

	// Extract option values.
	modulePath := cmdArgs[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		mlc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	cursor := getOption("cursor", "", cmdOpts)[0]
	limit, err := strconv.ParseInt(getOption("limit", "100", cmdOpts)[0], 10, 64) // 100 is the maximum allowed by GraphQL
	if err != nil {
		msg := fmt.Sprintf("invalid limit option value: %s", cmdOpts["limit"])
		mlc.meta.Logger.Error(output.FormatError(msg, nil))
		return 1
	}
	limit32 := int32(limit)
	sortByOption := strings.ToLower(getOption("sort-by", "", cmdOpts)[0])
	sortOrderOption := strings.ToLower(getOption("sort-order", "", cmdOpts)[0])

	if !isResourcePathValid(mlc.meta, modulePath) {
		return 1
	}

	// Get the module so, we can find it's ID.
	module, err := client.TerraformModule.GetModule(ctx, &sdktypes.GetTerraformModuleInput{Path: &modulePath})
	if err != nil {
		mlc.meta.Logger.Error(output.FormatError("failed to get module", err))
		return 1
	}

	// Leniently default to by created unless instructed otherwise.
	sortBy := "created"
	if strings.ToLower(sortByOption) == "updated" {
		sortBy = sortByOption
	}

	// Leniently default to ascending order unless instructed otherwise.
	sortOrder := "asc"
	if strings.HasSuffix(sortOrderOption, "desc") {
		sortOrder = sortOrderOption
	}

	// Decode from 2x2 to 1 of 4.
	var sortable sdktypes.TerraformModuleVersionSortableField
	if sortBy == "created" {
		if sortOrder == "asc" {
			sortable = sdktypes.TerraformModuleVersionSortableFieldCreatedAtAsc
		} else {
			sortable = sdktypes.TerraformModuleVersionSortableFieldCreatedAtDesc
		}
	} else {
		if sortOrder == "asc" {
			sortable = sdktypes.TerraformModuleVersionSortableFieldUpdatedAtAsc
		} else {
			sortable = sdktypes.TerraformModuleVersionSortableFieldUpdatedAtDesc
		}
	}

	// Prepare the inputs.
	input := &sdktypes.GetTerraformModuleVersionsInput{
		Sort: &sortable,
		PaginationOptions: &sdktypes.PaginationOptions{
			Cursor: &cursor,
			Limit:  &limit32,
		},
		TerraformModuleID: module.Metadata.ID,
	}
	if cursor == "" {
		input.PaginationOptions.Cursor = nil
	}

	mlc.meta.Logger.Debugf("module list-versions input: %#v", input)

	// Get the module versions.
	versionsOutput, err := client.TerraformModuleVersion.GetModuleVersions(ctx, input)
	if err != nil {
		mlc.meta.Logger.Error(output.FormatError("failed to get a list of module versions", err))
		return 1
	}

	if toJSON {
		buf, err := objectToJSON(versionsOutput)
		if err != nil {
			mlc.meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		mlc.meta.UI.Output(string(buf))
	} else {
		// Format the output.
		tableInput := make([][]string, len(versionsOutput.ModuleVersions)+1)
		tableInput[0] = []string{"id", "module id", "version", "shasum", "status", "latest"}
		for ix, moduleVersion := range versionsOutput.ModuleVersions {
			tableInput[ix+1] = []string{
				moduleVersion.Metadata.ID, moduleVersion.ModuleID, moduleVersion.Version,
				moduleVersion.SHASum, moduleVersion.Status, fmt.Sprintf("%t", moduleVersion.Latest),
			}
		}
		mlc.meta.UI.Output(tableformatter.FormatTable(tableInput))
		//
		// Must return the new cursor at the end of the list of module versions.
		mlc.meta.UI.Output(fmt.Sprintf("has next page: %v", versionsOutput.PageInfo.HasNextPage))
		if versionsOutput.PageInfo.HasNextPage {
			// Show the next cursor _ONLY_ if there is a next page.
			mlc.meta.UI.Output(fmt.Sprintf("next cursor: %s", versionsOutput.PageInfo.Cursor))
		}
	}

	return 0
}

func (mlc moduleListVersionsCommand) buildModuleListVersionsDefs() optparser.OptionDefinitions {
	defs := buildPaginationOptionDefs()

	defs["sort-by"] = &optparser.OptionDefinition{
		Arguments: []string{"Sort_By"},
		Synopsis:  "Sort by this field: CREATED or UPDATED.",
	}

	return buildJSONOptionDefs(defs)
}

func (mlc moduleListVersionsCommand) Synopsis() string {
	return "List module versions."
}

func (mlc moduleListVersionsCommand) Help() string {
	return mlc.HelpModuleListVersions()
}

// HelpModuleListVersions returns the help string for the 'module list-versions' command.
func (mlc moduleListVersionsCommand) HelpModuleListVersions() string {
	return fmt.Sprintf(`
Usage: %s [global options] module list-versions [options] <module-path>

   The module list-versions command prints information
   about (likely multiple) module versions. Supports
   pagination, filtering and sorting the output.

   Example:

   %s module list-versions \
      --limit 5 \
      --json \
      some/module/aws

   Above command will only show five module versions
   in JSON format.

%s

`,
		mlc.meta.BinaryName,
		mlc.meta.BinaryName,
		buildHelpText(mlc.buildModuleListVersionsDefs()),
	)
}
