package command

import (
	"errors"
	"maps"
	"slices"
	"strings"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type terraformProviderListCommand struct {
	*BaseCommand

	limit   *int32
	cursor  *string
	search  *string
	groupID *string
	sortBy  *string
	toJSON  *bool
}

var _ Command = (*terraformProviderListCommand)(nil)

// NewTerraformProviderListCommandFactory returns a terraformProviderListCommand struct.
func NewTerraformProviderListCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderListCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *terraformProviderListCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

func (c *terraformProviderListCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider list"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetTerraformProvidersRequest{
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		Search:  c.search,
		GroupId: c.groupID,
	}

	if c.sortBy != nil {
		input.Sort = pb.TerraformProviderSortableField(pb.TerraformProviderSortableField_value[*c.sortBy]).Enum()
	}

	result, err := c.grpcClient.TerraformProvidersClient.GetTerraformProviders(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of terraform providers")
		return 1
	}

	return c.OutputList(result, c.toJSON, "trn", "name", "group_id", "private")
}

func (*terraformProviderListCommand) Synopsis() string {
	return "Retrieve a paginated list of terraform providers."
}

func (*terraformProviderListCommand) Usage() string {
	return "tharsis [global options] terraform-provider list [options]"
}

func (*terraformProviderListCommand) Description() string {
	return `
   Lists Terraform providers within a group with
   pagination and sorting.
`
}

func (*terraformProviderListCommand) Example() string {
	return `
tharsis terraform-provider list -group-id "trn:group:<group_path>" -json
`
}

func (c *terraformProviderListCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.TerraformProviderSortableField_value))

	f := flag.NewSet("Command options")
	f.StringVar(
		&c.groupID,
		"group-id",
		"Group ID to list terraform providers for.",
		flag.Required(),
	)
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
		flag.TransformString(strings.ToUpper),
	)
	f.StringVar(
		&c.search,
		"search",
		"Filter to terraform providers containing this substring.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
