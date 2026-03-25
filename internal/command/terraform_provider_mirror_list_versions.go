package command

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
)

type terraformProviderMirrorListVersionsCommand struct {
	*BaseCommand

	limit     *int32
	cursor    *string
	sortBy    *string
	sortOrder *string
	toJSON    *bool
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
	if c.sortBy != nil && c.sortOrder != nil {
		return fmt.Errorf("cannot use both -sort-by and -sort-order")
	}

	const message = "namespace-path is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
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

	var sortEnum *pb.TerraformProviderVersionMirrorSortableField
	if c.sortBy != nil {
		v := pb.TerraformProviderVersionMirrorSortableField(pb.TerraformProviderVersionMirrorSortableField_value[strings.ToUpper(*c.sortBy)])
		sortEnum = &v
	}

	if c.sortOrder != nil {
		var v pb.TerraformProviderVersionMirrorSortableField
		switch strings.ToUpper(*c.sortOrder) {
		case "ASC":
			v = pb.TerraformProviderVersionMirrorSortableField_CREATED_AT_ASC
		case "DESC":
			v = pb.TerraformProviderVersionMirrorSortableField_CREATED_AT_DESC
		}

		sortEnum = &v
	}

	input := &pb.GetTerraformProviderVersionMirrorsRequest{
		NamespacePath: c.arguments[0],
		Sort:          sortEnum,
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
	}

	result, err := c.grpcClient.TerraformProviderMirrorsClient.GetTerraformProviderVersionMirrors(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of provider version mirrors")
		return 1
	}

	if *c.toJSON {
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
   about provider version mirrors for a namespace. Supports pagination
   and sorting.
`
}

func (*terraformProviderMirrorListVersionsCommand) Usage() string {
	return "tharsis [global options] terraform-provider-mirror list-versions [options] <namespace-path>"
}

func (*terraformProviderMirrorListVersionsCommand) Example() string {
	return `
tharsis terraform-provider-mirror list-versions \
  -sort-by CREATED_AT_DESC \
  -limit 10 \
  <namespace_path>
`
}

func (c *terraformProviderMirrorListVersionsCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.TerraformProviderVersionMirrorSortableField_value))

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
		flag.ValidValues(sortValues...),
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
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
		flag.Default(false),
	)

	return f
}
