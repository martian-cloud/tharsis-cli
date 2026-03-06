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

type moduleListVersionsCommand struct {
	*BaseCommand

	limit           *int32
	cursor          *string
	search          *string
	latest          *bool
	semanticVersion *string
	sortBy          *pb.TerraformModuleVersionSortableField
	toJSON          bool
}

// NewModuleListVersionsCommandFactory returns a moduleListVersionsCommand struct.
func NewModuleListVersionsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleListVersionsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *moduleListVersionsCommand) validate() error {
	const message = "module-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.limit, validation.Min(0), validation.Max(100), validation.When(c.limit != nil)),
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

	if c.limit == nil {
		c.limit = ptr.Int32(defaultPaginationLimit)
	}

	input := &pb.GetTerraformModuleVersionsRequest{
		ModuleId: c.arguments[0],
		Sort:     c.sortBy,
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		Search:          c.search,
		Latest:          c.latest,
		SemanticVersion: c.semanticVersion,
	}

	c.Logger.Debug("module list-versions input", "input", input)

	result, err := c.client.TerraformModulesClient.GetTerraformModuleVersions(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of module versions")
		return 1
	}

	if c.toJSON {
		if err := c.UI.JSON(result); err != nil {
			c.UI.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		t := terminal.NewTable("id", "version", "status", "latest", "sha_sum")

		for _, version := range result.Versions {
			t.Rich([]string{
				version.Metadata.Id,
				version.SemanticVersion,
				version.Status,
				fmt.Sprintf("%t", version.Latest),
				version.ShaSum,
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
  --search 1.0 \
  --sort-by CREATED_AT_DESC \
  --limit 10 \
  trn:terraform_module:ops/installer/aws
`
}

func (c *moduleListVersionsCommand) Flags() *flag.FlagSet {
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
		"Sort by this field (e.g., CREATED_AT_ASC, CREATED_AT_DESC).",
		func(s string) error {
			value, ok := pb.TerraformModuleVersionSortableField_value[strings.ToUpper(s)]
			if !ok {
				return fmt.Errorf("invalid sort-by value: %s (valid values: %v)", s, maps.Keys(pb.TerraformModuleVersionSortableField_value))
			}
			c.sortBy = pb.TerraformModuleVersionSortableField(value).Enum()
			return nil
		},
	)
	f.Func(
		"search",
		"Filter to versions containing this substring.",
		func(s string) error {
			c.search = &s
			return nil
		},
	)
	f.Func(
		"latest",
		"Filter to only the latest version.",
		func(s string) error {
			b := s == "true"
			c.latest = &b
			return nil
		},
	)
	f.Func(
		"semantic-version",
		"Filter to a specific semantic version.",
		func(s string) error {
			c.semanticVersion = &s
			return nil
		},
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}
