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
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// moduleListAttestationsCommand is the top-level structure for the module list-attestations command.
type moduleListAttestationsCommand struct {
	meta *Metadata
}

// NewModuleListAttestationsCommandFactory returns a moduleListAttestationsCommand struct.
func NewModuleListAttestationsCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return moduleListAttestationsCommand{
			meta: meta,
		}, nil
	}
}

func (mlc moduleListAttestationsCommand) Run(args []string) int {
	mlc.meta.Logger.Debugf("Starting the 'module list-attestations' command with %d arguments:", len(args))
	for ix, arg := range args {
		mlc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := mlc.meta.GetSDKClient()
	if err != nil {
		mlc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return mlc.doModuleListAttestations(ctx, client, args)
}

func (mlc moduleListAttestationsCommand) doModuleListAttestations(ctx context.Context, client *tharsis.Client, opts []string) int {
	mlc.meta.Logger.Debugf("will do module list-attestations, %d opts: %#v", len(opts), opts)

	defs := mlc.buildModuleListAttestationsDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(mlc.meta.BinaryName+" module list-attestations", defs, opts)
	if err != nil {
		mlc.meta.Logger.Error(output.FormatError("failed to parse module list-attestations options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		mlc.meta.Logger.Error(output.FormatError("missing module list-attestations module path", nil), mlc.HelpModuleListAttestations())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive module list-attestations arguments: %s", cmdArgs)
		mlc.meta.Logger.Error(output.FormatError(msg, nil), mlc.HelpModuleListAttestations())
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
	version := getOption("version", "", cmdOpts)[0]
	digest := getOption("digest", "", cmdOpts)[0]
	sortByOption := strings.ToLower(getOption("sort-by", "", cmdOpts)[0])
	sortOrderOption := strings.ToLower(getOption("sort-order", "", cmdOpts)[0])

	actualPath := trn.ToPath(modulePath)
	if !isResourcePathValid(mlc.meta, actualPath) {
		return 1
	}

	// Get the module so, we can find it's ID.
	module, err := client.TerraformModule.GetModule(ctx, &sdktypes.GetTerraformModuleInput{Path: &actualPath}) // Use extracted path
	if err != nil {
		mlc.meta.Logger.Error(output.FormatError("failed to get module", err))
		return 1
	}

	var versionID *string
	if version != "" {
		version, vErr := client.TerraformModuleVersion.GetModuleVersion(ctx, &sdktypes.GetTerraformModuleVersionInput{
			ModulePath: &actualPath, // Use extracted path
			Version:    &version,
		})
		if vErr != nil {
			mlc.meta.Logger.Error(output.FormatError("failed to get module version", vErr))
			return 1
		}

		versionID = &version.Metadata.ID
	}

	// Leniently default to by created unless instructed otherwise.
	sortBy := "created"
	if strings.ToLower(sortByOption) == "predicate" {
		sortBy = sortByOption
	}

	// Leniently default to ascending order unless instructed otherwise.
	sortOrder := "asc"
	if strings.HasSuffix(sortOrderOption, "desc") {
		sortOrder = sortOrderOption
	}

	// Decode from 2x2 to 1 of 4.
	var sortable sdktypes.TerraformModuleAttestationSortableField
	if sortBy == "created" {
		if sortOrder == "asc" {
			sortable = sdktypes.TerraformModuleAttestationSortableFieldCreatedAtAsc
		} else {
			sortable = sdktypes.TerraformModuleAttestationSortableFieldCreatedAtDesc
		}
	} else {
		if sortOrder == "asc" {
			sortable = sdktypes.TerraformModuleAttestationSortableFieldPredicateAsc
		} else {
			sortable = sdktypes.TerraformModuleAttestationSortableFieldPredicateDesc
		}
	}

	// Prepare the inputs.
	input := &sdktypes.GetTerraformModuleAttestationsInput{
		Sort: &sortable,
		PaginationOptions: &sdktypes.PaginationOptions{
			Cursor: &cursor,
			Limit:  &limit32,
		},
	}

	filter := &sdktypes.TerraformModuleAttestationFilter{}
	if versionID != nil {
		filter.TerraformModuleVersionID = versionID
	} else {
		filter.TerraformModuleID = &module.Metadata.ID
	}

	if digest != "" {
		filter.Digest = &digest
	}

	input.Filter = filter

	if cursor == "" {
		input.PaginationOptions.Cursor = nil
	}

	mlc.meta.Logger.Debugf("module list-attestations input: %#v", input)

	// Get the module attestations.
	attestationsOutput, err := client.TerraformModuleAttestation.GetModuleAttestations(ctx, input)
	if err != nil {
		mlc.meta.Logger.Error(output.FormatError("failed to get a list of module attestations", err))
		return 1
	}

	if toJSON {
		buf, err := objectToJSON(attestationsOutput)
		if err != nil {
			mlc.meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		mlc.meta.UI.Output(string(buf))
	} else {
		// Format the output.
		tableInput := make([][]string, len(attestationsOutput.ModuleAttestations)+1)
		tableInput[0] = []string{"id", "module id", "description", "schema type", "predicate type"}
		for ix, attestation := range attestationsOutput.ModuleAttestations {
			tableInput[ix+1] = []string{
				attestation.Metadata.ID, attestation.ModuleID,
				attestation.Description, attestation.SchemaType, attestation.PredicateType,
			}
		}
		mlc.meta.UI.Output(tableformatter.FormatTable(tableInput))
		// Must return the new cursor at the end of the list of module attestations.
		mlc.meta.UI.Output(fmt.Sprintf("has next page: %v", attestationsOutput.PageInfo.HasNextPage))
		if attestationsOutput.PageInfo.HasNextPage {
			// Show the next cursor _ONLY_ if there is a next page.
			mlc.meta.UI.Output(fmt.Sprintf("next cursor: %s", attestationsOutput.PageInfo.Cursor))
		}
	}

	return 0
}

func (mlc moduleListAttestationsCommand) buildModuleListAttestationsDefs() optparser.OptionDefinitions {
	defs := buildPaginationOptionDefs()

	defs["digest"] = &optparser.OptionDefinition{
		Arguments: []string{"Digest"},
		Synopsis:  "Filter attestations by digest (not applicable for --version).",
	}

	defs["sort-by"] = &optparser.OptionDefinition{
		Arguments: []string{"Sort_By"},
		Synopsis:  "Sort by this field: PREDICATE or CREATED.",
	}

	defs["version"] = &optparser.OptionDefinition{
		Arguments: []string{"Version"},
		Synopsis:  "A semver compliant version tag to list attestations for.",
	}

	return buildJSONOptionDefs(defs)
}

func (mlc moduleListAttestationsCommand) Synopsis() string {
	return "List attestations for a module."
}

func (mlc moduleListAttestationsCommand) Help() string {
	return mlc.HelpModuleListAttestations()
}

// HelpModuleListAttestations returns the help string for the 'module list-attestations' command.
func (mlc moduleListAttestationsCommand) HelpModuleListAttestations() string {
	return fmt.Sprintf(`
Usage: %s [global options] module list-attestations [options] <module-path>

   The module list-attestations command prints information
   about (likely multiple) module attestations. By default,
   lists attestations for a module and optionally lists
   attestations for a module version tag specified by
   --version option. Supports pagination, filtering and
   sorting the output.

   Example:

   %s module list-attestations \
      --limit 5 \
      --json \
      some/module/aws

   Above command will only show five module attestations
   in JSON format.

%s

`,
		mlc.meta.BinaryName,
		mlc.meta.BinaryName,
		buildHelpText(mlc.buildModuleListAttestationsDefs()),
	)
}
