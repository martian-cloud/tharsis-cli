package command

import (
	"flag"
	"fmt"
	"maps"
	"strconv"
	"strings"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type moduleListCommand struct {
	*BaseCommand

	limit            *int32
	cursor           *string
	search           *string
	groupID          *string
	includeInherited bool
	sortBy           *pb.TerraformModuleSortableField
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
		validation.Field(&c.limit, validation.Min(0), validation.Max(100), validation.When(c.limit != nil)),
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

	if c.limit == nil {
		c.limit = ptr.Int32(defaultPaginationLimit)
	}

	input := &pb.GetTerraformModulesRequest{
		Sort: c.sortBy,
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		Search:           c.search,
		GroupId:          c.groupID,
		IncludeInherited: c.includeInherited,
	}

	c.Logger.Debug("module list input", "input", input)

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
  --group-id trn:group:top-level \
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
	f.Func(
		"limit",
		"Maximum number of result elements to return. Defaults to 100.",
		func(s string) error {
			i, err := strconv.ParseInt(s, 10, 32)
			if err != nil {
				return err
			}
			c.limit = ptr.Int32(int32(i))
			return nil
		},
	)
	f.Func(
		"sort-by",
		"Sort by this field (e.g., NAME_ASC, NAME_DESC, UPDATED_AT_ASC, UPDATED_AT_DESC).",
		func(s string) error {
			value, ok := pb.TerraformModuleSortableField_value[strings.ToUpper(s)]
			if !ok {
				return fmt.Errorf("invalid sort-by value: %s (valid values: %v)", s, maps.Keys(pb.TerraformModuleSortableField_value))
			}
			c.sortBy = pb.TerraformModuleSortableField(value).Enum()
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
