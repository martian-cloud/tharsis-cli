package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
)

type moduleListCommand struct {
	*BaseCommand

	limit            int
	cursor           *string
	search           *string
	groupID          *string
	includeInherited bool
	sortOrder        *string
	sortBy           *string
	toJSON           bool
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
	return validation.ValidateStruct(c,
		validation.Field(&c.limit, validation.Min(0), validation.Max(maxPaginationLimit)),
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
			First: ptr.Int32(int32(c.limit)),
			After: c.cursor,
		},
		Search:           c.search,
		GroupId:          c.groupID,
		IncludeInherited: c.includeInherited,
	}

	result, err := c.grpcClient.TerraformModulesClient.GetTerraformModules(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of modules")
		return 1
	}

	if c.toJSON {
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
  --group-id trn:group:<group_path> \
  --include-inherited \
  --sort-by UPDATED_AT_DESC \
  --limit 5 \
  --json
`
}

func (c *moduleListCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"cursor",
		"The cursor string for manual pagination.",
		func(s string) error {
			c.cursor = &s
			return nil
		},
	)
	f.IntVar(
		&c.limit,
		"limit",
		maxPaginationLimit,
		"Maximum number of result elements to return.",
	)
	f.Func(
		"sort-by",
		"Sort by this field (e.g., NAME_ASC, NAME_DESC, GROUP_LEVEL_ASC, GROUP_LEVEL_DESC, UPDATED_AT_ASC, UPDATED_AT_DESC).",
		func(s string) error {
			// TODO: Update to use PB types and validate with PB map once deprecation is done.
			switch v := strings.ToUpper(s); v {
			case "NAME", // DEPRECATED
				pb.TerraformModuleSortableField_GROUP_LEVEL_ASC.String(),
				pb.TerraformModuleSortableField_GROUP_LEVEL_DESC.String(),
				pb.TerraformModuleSortableField_NAME_ASC.String(),
				pb.TerraformModuleSortableField_NAME_DESC.String(),
				pb.TerraformModuleSortableField_UPDATED_AT_ASC.String(),
				pb.TerraformModuleSortableField_UPDATED_AT_DESC.String():
				c.sortBy = &v
			case "UPDATED": // Deprecated
				c.sortBy = ptr.String("UPDATED_AT")
			default:
				return fmt.Errorf("unknown sort by option %s", s)
			}

			return nil
		},
	)
	f.Func(
		"sort-order",
		"Sort in this direction, ASC or DESC. Deprecated",
		func(s string) error {
			switch v := strings.ToUpper(s); v {
			case "ASC", "DESC":
				c.sortOrder = &v
			default:
				return fmt.Errorf("invalid sort-order value: %s", s)
			}

			return nil
		},
	)
	f.Func(
		"search",
		"Filter to only modules containing this substring in their path.",
		func(s string) error {
			c.search = &s
			return nil
		},
	)
	f.Func(
		"group-id",
		"Filter to only modules in this group.",
		func(s string) error {
			c.groupID = &s
			return nil
		},
	)
	f.BoolVar(
		&c.includeInherited,
		"include-inherited",
		false,
		"Include modules inherited from parent groups.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}
