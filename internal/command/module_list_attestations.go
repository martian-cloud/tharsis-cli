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

type moduleListAttestationsCommand struct {
	*BaseCommand

	limit     *int32
	cursor    *string
	digest    *string
	sortBy    *string
	sortOrder *string
	toJSON    *bool
}

var _ Command = (*moduleListAttestationsCommand)(nil)

// NewModuleListAttestationsCommandFactory returns a moduleListAttestationsCommand struct.
func NewModuleListAttestationsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleListAttestationsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *moduleListAttestationsCommand) validate() error {
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

func (c *moduleListAttestationsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("module list-attestations"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	sortByEnum, err := parseSortField[pb.TerraformModuleAttestationSortableField](
		c.sortBy,
		c.sortOrder,
		pb.TerraformModuleAttestationSortableField_value,
	)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to parse sort field")
		return 1
	}

	input := &pb.GetTerraformModuleAttestationsRequest{
		ModuleId: trn.ToTRN(trn.ResourceTypeTerraformModule, c.arguments[0]),
		Sort:     sortByEnum,
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		Digest: c.digest,
	}

	result, err := c.grpcClient.TerraformModulesClient.GetTerraformModuleAttestations(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of module attestations")
		return 1
	}

	return c.OutputList(result, c.toJSON)
}

func (*moduleListAttestationsCommand) Synopsis() string {
	return "Retrieve a paginated list of module attestations."
}

func (*moduleListAttestationsCommand) Description() string {
	return `
   The module list-attestations command prints information about attestations
   for a specific module. Supports pagination, filtering and sorting.
`
}

func (*moduleListAttestationsCommand) Usage() string {
	return "tharsis [global options] module list-attestations [options] <module-id>"
}

func (*moduleListAttestationsCommand) Example() string {
	return `
tharsis module list-attestations \
  -sort-by CREATED_AT_DESC \
  -limit 10 \
  trn:terraform_module:<group_path>/<module_name>/<system>
`
}

func (c *moduleListAttestationsCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.TerraformModuleAttestationSortableField_value))

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
		&c.digest,
		"digest",
		"Filter to attestations with this digest.",
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
