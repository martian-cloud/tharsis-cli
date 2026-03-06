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

type moduleListAttestationsCommand struct {
	*BaseCommand

	limit  *int32
	cursor *string
	digest *string
	sortBy *pb.TerraformModuleAttestationSortableField
	toJSON bool
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
		validation.Field(&c.limit, validation.Min(0), validation.Max(100), validation.When(c.limit != nil)),
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

	if c.limit == nil {
		c.limit = ptr.Int32(defaultPaginationLimit)
	}

	input := &pb.GetTerraformModuleAttestationsRequest{
		ModuleId: c.arguments[0],
		Sort:     c.sortBy,
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		Digest: c.digest,
	}

	c.Logger.Debug("module list-attestations input", "input", input)

	result, err := c.client.TerraformModulesClient.GetTerraformModuleAttestations(c.Context, input)
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
  trn:terraform_module:ops/installer/aws
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
			value, ok := pb.TerraformModuleAttestationSortableField_value[strings.ToUpper(s)]
			if !ok {
				return fmt.Errorf("invalid sort-by value: %s (valid values: %v)", s, maps.Keys(pb.TerraformModuleAttestationSortableField_value))
			}
			c.sortBy = pb.TerraformModuleAttestationSortableField(value).Enum()
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
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}
