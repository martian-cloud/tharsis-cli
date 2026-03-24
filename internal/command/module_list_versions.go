package command

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type moduleListVersionsCommand struct {
	*BaseCommand

	limit           int
	sortOrder       *string
	cursor          *string
	search          *string
	latest          *bool
	semanticVersion *string
	sortBy          *string
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
		validation.Field(&c.limit, validation.Min(0), validation.Max(maxPaginationLimit)),
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
			First: ptr.Int32(int32(c.limit)),
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
  trn:terraform_module:<group_path>/<module_name>/<system>
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
	f.IntVar(
		&c.limit,
		"limit",
		maxPaginationLimit,
		"Maximum number of result elements to return.",
	)
	f.Func(
		"sort-by",
		"Sort by this field (e.g., CREATED_AT_ASC, CREATED_AT_DESC).",
		func(s string) error {
			// TODO: Update to use PB types and validate with PB map once deprecation is done.
			switch v := strings.ToUpper(s); v {
			case pb.TerraformModuleVersionSortableField_CREATED_AT_ASC.String(),
				pb.TerraformModuleVersionSortableField_CREATED_AT_DESC.String(),
				pb.TerraformModuleVersionSortableField_UPDATED_AT_ASC.String(),
				pb.TerraformModuleVersionSortableField_UPDATED_AT_DESC.String():
				c.sortBy = &v
			case "UPDATED": // Deprecated
				c.sortBy = ptr.String("UPDATED_AT")
			case "CREATED": // Deprecated
				c.sortBy = ptr.String("CREATED_AT")
			default:
				return fmt.Errorf("unknown sort by option %s", s)
			}

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
	f.BoolFunc(
		"latest",
		"Filter to only the latest version.",
		func(s string) error {
			b, err := strconv.ParseBool(s)
			if err != nil {
				return err
			}

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
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}
