package command

import (
	"fmt"
	"maps"
	"slices"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type moduleListVersionsCommand struct {
	*BaseCommand

	limit           *int32
	cursor          *string
	search          *string
	sortBy          *string
	sortOrder       *string
	latest          *bool
	semanticVersion *string
	toJSON          *bool
}

var _ Command = (*moduleListVersionsCommand)(nil)

// NewModuleListVersionsCommandFactory returns a moduleListVersionsCommand struct.
func NewModuleListVersionsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleListVersionsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *moduleListVersionsCommand) validate() error {
	if c.sortBy != nil && c.sortOrder != nil {
		return fmt.Errorf("cannot use both -sort-by and -sort-order")
	}

	const message = "module-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *moduleListVersionsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("module list-versions"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	sortByEnum, err := parseSortField[pb.TerraformModuleVersionSortableField](
		c.sortBy,
		c.sortOrder,
		pb.TerraformModuleVersionSortableField_value,
	)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to parse sort field")
		return 1
	}

	input := &pb.GetTerraformModuleVersionsRequest{
		ModuleId: trn.ToTRN(trn.ResourceTypeTerraformModule, c.arguments[0]),
		Sort:     sortByEnum,
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		Search:          c.search,
		Latest:          c.latest,
		SemanticVersion: c.semanticVersion,
	}

	result, err := c.grpcClient.TerraformModulesClient.GetTerraformModuleVersions(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of module versions")
		return 1
	}

	return c.OutputList(result, c.toJSON, "trn", "semantic_version", "status", "latest")
}

func (*moduleListVersionsCommand) Synopsis() string {
	return "Retrieve a paginated list of module versions."
}

func (*moduleListVersionsCommand) Description() string {
	return `
   The module list-versions command prints information about versions
   of a specific module. Supports pagination, filtering and sorting.
`
}

func (*moduleListVersionsCommand) Usage() string {
	return "tharsis [global options] module list-versions [options] <module-id>"
}

func (*moduleListVersionsCommand) Example() string {
	return `
tharsis module list-versions \
  -search "1.0" \
  -sort-by "CREATED_AT_DESC" \
  -limit 10 \
  trn:terraform_module:<group_path>/<module_name>/<system>
`
}

func (c *moduleListVersionsCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.TerraformModuleVersionSortableField_value))

	f := flag.NewSet("Command options")
	f.StringVar(
		&c.cursor,
		"cursor",
		"The cursor string for manual pagination.",
	)
	f.Int32Var(
		&c.limit,
		"limit",
		"Maximum number of result elements to return.",
		flag.Default(maxPaginationLimit),
		flag.ValidRange(0, int(maxPaginationLimit)),
	)
	f.StringVar(
		&c.sortBy,
		"sort-by",
		"Sort by this field.",
		flag.PredictValues(sortValues...),
	)
	f.StringVar(
		&c.search,
		"search",
		"Filter to versions containing this substring.",
	)
	f.BoolVar(
		&c.latest,
		"latest",
		"Filter to only the latest version.",
	)
	f.StringVar(
		&c.semanticVersion,
		"semantic-version",
		"Filter to a specific semantic version.",
	)
	f.StringVar(
		&c.sortOrder,
		"sort-order",
		"Sort in this direction.",
		flag.Deprecated("use -sort-by"),
		flag.ValidValues("ASC", "DESC"),
		flag.PredictValues("ASC", "DESC"),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
		flag.Default(false),
	)

	return f
}
