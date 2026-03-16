package command

import (
	"flag"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type terraformProviderMirrorListVersionsCommand struct {
	*BaseCommand

	limit  int
	cursor *string
	sortBy *pb.TerraformProviderVersionMirrorSortableField
	toJSON bool
}

// NewTerraformProviderMirrorListVersionsCommandFactory returns a terraformProviderMirrorListVersionsCommand struct.
func NewTerraformProviderMirrorListVersionsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderMirrorListVersionsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *terraformProviderMirrorListVersionsCommand) validate() error {
	const message = "namespace-path is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.limit, validation.Min(0), validation.Max(maxPaginationLimit)),
	)
}

func (c *terraformProviderMirrorListVersionsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider-mirror list-versions"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetTerraformProviderVersionMirrorsRequest{
		NamespacePath: c.arguments[0],
		Sort:          c.sortBy,
		PaginationOptions: &pb.PaginationOptions{
			First: ptr.Int32(int32(c.limit)),
			After: c.cursor,
		},
	}

	result, err := c.grpcClient.TerraformProviderMirrorsClient.GetTerraformProviderVersionMirrors(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of provider version mirrors")
		return 1
	}

	if c.toJSON {
		if err := c.UI.JSON(result); err != nil {
			c.UI.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		t := terminal.NewTable("id", "type", "registry_hostname", "registry_namespace", "semantic_version")

		for _, mirror := range result.VersionMirrors {
			t.Rich([]string{
				mirror.Metadata.Id,
				mirror.Type,
				mirror.RegistryHostname,
				mirror.RegistryNamespace,
				mirror.SemanticVersion,
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

func (*terraformProviderMirrorListVersionsCommand) Synopsis() string {
	return "Retrieve a paginated list of provider version mirrors."
}

func (*terraformProviderMirrorListVersionsCommand) Description() string {
	return `
   The terraform-provider-mirror list-versions command prints information
   about provider version mirrors in a namespace. Supports pagination and sorting.
`
}

func (*terraformProviderMirrorListVersionsCommand) Usage() string {
	return "tharsis [global options] terraform-provider-mirror list-versions [options] <namespace-path>"
}

func (*terraformProviderMirrorListVersionsCommand) Example() string {
	return `
tharsis terraform-provider-mirror list-versions \
  --sort-by CREATED_AT_DESC \
  --limit 10 \
  <namespace_path>
`
}

func (c *terraformProviderMirrorListVersionsCommand) Flags() *flag.FlagSet {
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
			value, ok := pb.TerraformProviderVersionMirrorSortableField_value[strings.ToUpper(s)]
			if !ok {
				return fmt.Errorf("invalid sort-by value: %s (valid values: %v)", s, slices.Collect(maps.Keys(pb.TerraformProviderVersionMirrorSortableField_value)))
			}
			c.sortBy = pb.TerraformProviderVersionMirrorSortableField(value).Enum()
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
