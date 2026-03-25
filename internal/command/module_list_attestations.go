package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type moduleListAttestationsCommand struct {
	*BaseCommand

	limit     int
	sortOrder *string
	cursor    *string
	digest    *string
	sortBy    *string
	toJSON    bool
}

// NewModuleListAttestationsCommandFactory returns a moduleListAttestationsCommand struct.
func NewModuleListAttestationsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleListAttestationsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *moduleListAttestationsCommand) validate() error {
	const message = "module-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.limit, validation.Min(0), validation.Max(maxPaginationLimit)),
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
			First: ptr.Int32(int32(c.limit)),
			After: c.cursor,
		},
		Digest: c.digest,
	}

	result, err := c.grpcClient.TerraformModulesClient.GetTerraformModuleAttestations(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of module attestations")
		return 1
	}

	if c.toJSON {
		if err := c.UI.JSON(result); err != nil {
			c.UI.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		t := terminal.NewTable("id", "description", "predicate_type", "schema_type")

		for _, attestation := range result.Attestations {
			t.Rich([]string{
				attestation.Metadata.Id,
				attestation.Description,
				attestation.PredicateType,
				attestation.SchemaType,
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
  --sort-by CREATED_AT_DESC \
  --limit 10 \
  trn:terraform_module:<group_path>/<module_name>/<system>
`
}

func (c *moduleListAttestationsCommand) Flags() *flag.FlagSet {
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
			case "PREDICATE", // Deprecated
				pb.TerraformModuleAttestationSortableField_CREATED_AT_ASC.String(),
				pb.TerraformModuleAttestationSortableField_CREATED_AT_DESC.String(),
				pb.TerraformModuleAttestationSortableField_PREDICATE_ASC.String(),
				pb.TerraformModuleAttestationSortableField_PREDICATE_DESC.String():
				c.sortBy = &v
			case "CREATED": // Deprecated
				c.sortBy = ptr.String("CREATED_AT")
			default:
				return fmt.Errorf("unknown sort by option %s", s)
			}

			return nil
		},
	)
	f.Func(
		"digest",
		"Filter to attestations with this digest.",
		func(s string) error {
			c.digest = &s
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
