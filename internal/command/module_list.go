package command

import (
	"fmt"
	"maps"
	"slices"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
)

type moduleListCommand struct {
	*BaseCommand

	limit            *int32
	cursor           *string
	search           *string
	groupID          *string
	sortBy           *string
	sortOrder        *string
	includeInherited *bool
	toJSON           *bool
}

// NewModuleListCommandFactory returns a moduleListCommand struct.
func NewModuleListCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleListCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *moduleListCommand) validate() error {
	if c.sortBy != nil && c.sortOrder != nil {
		return fmt.Errorf("cannot use both -sort-by and -sort-order")
	}

	return validation.ValidateStruct(c,
		validation.Field(&c.arguments, validation.Empty),
	)
}

func (c *moduleListCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("module list"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	sortByEnum, err := parseSortField[pb.TerraformModuleSortableField](
		c.sortBy,
		c.sortOrder,
		pb.TerraformModuleSortableField_value,
	)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to parse sort field")
		return 1
	}

	input := &pb.GetTerraformModulesRequest{
		Sort: sortByEnum,
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		Search:           c.search,
		GroupId:          c.groupID,
		IncludeInherited: *c.includeInherited,
	}

	result, err := c.grpcClient.TerraformModulesClient.GetTerraformModules(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of modules")
		return 1
	}

	if *c.toJSON {
		if err := c.UI.JSON(result); err != nil {
			c.UI.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		t := terminal.NewTable("id", "name", "system", "private")

		for _, module := range result.Modules {
			t.Rich([]string{
				module.GetMetadata().Id,
				module.Name,
				module.System,
				fmt.Sprintf("%t", module.Private),
			}, nil)
		}

		c.UI.Table(t)
		namedValues := []terminal.NamedValue{
			{Name: "Total count", Value: result.GetPageInfo().TotalCount},
			{Name: "Has Next Page", Value: result.GetPageInfo().HasNextPage},
		}
		if result.GetPageInfo().EndCursor != nil {
			namedValues = append(namedValues, terminal.NamedValue{
				Name:  "Next cursor",
				Value: result.GetPageInfo().GetEndCursor(),
			})
		}

		c.UI.NamedValues(namedValues)
	}

	return 0
}

func (*moduleListCommand) Synopsis() string {
	return "Retrieve a paginated list of modules."
}

func (*moduleListCommand) Description() string {
	return `
   The module list command prints information about (likely
   multiple) modules. Supports pagination, filtering and
   sorting the output.
`
}

func (*moduleListCommand) Usage() string {
	return "tharsis [global options] module list [options]"
}

func (*moduleListCommand) Example() string {
	return `
tharsis module list \
  -group-id trn:group:<group_path> \
  -include-inherited \
  -sort-by UPDATED_AT_DESC \
  -limit 5 \
  -json
`
}

func (c *moduleListCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.TerraformModuleSortableField_value))

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
		&c.sortOrder,
		"sort-order",
		"Sort in this direction.",
		flag.Deprecated("use -sort-by"),
		flag.ValidValues("ASC", "DESC"),
		flag.PredictValues("ASC", "DESC"),
	)
	f.StringVar(
		&c.search,
		"search",
		"Filter to only modules containing this substring in their path.",
	)
	f.StringVar(
		&c.groupID,
		"group-id",
		"Filter to only modules in this group.",
	)
	f.BoolVar(
		&c.includeInherited,
		"include-inherited",
		"Include modules inherited from parent groups.",
		flag.Default(false),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
		flag.Default(false),
	)

	return f
}
