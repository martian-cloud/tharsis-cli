package command

import (
	"errors"
	"maps"
	"slices"
	"strings"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type vcsProviderListCommand struct {
	*BaseCommand

	limit            *int32
	cursor           *string
	search           *string
	namespacePath    *string
	sortBy           *string
	includeInherited *bool
	toJSON           *bool
}

var _ Command = (*vcsProviderListCommand)(nil)

// NewVCSProviderListCommandFactory returns a vcsProviderListCommand struct.
func NewVCSProviderListCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &vcsProviderListCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *vcsProviderListCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

func (c *vcsProviderListCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("vcs-provider list"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetVCSProvidersRequest{
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		Search:           c.search,
		NamespacePath:    *c.namespacePath,
		IncludeInherited: *c.includeInherited,
	}

	if c.sortBy != nil {
		input.Sort = pb.VCSProviderSortableField(pb.VCSProviderSortableField_value[*c.sortBy]).Enum()
	}

	result, err := c.grpcClient.VCSProvidersClient.GetVCSProviders(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of VCS providers")
		return 1
	}

	return c.OutputList(result, c.toJSON, "trn", "name", "type", "url", "description")
}

func (*vcsProviderListCommand) Synopsis() string {
	return "Retrieve a paginated list of VCS providers."
}

func (*vcsProviderListCommand) Usage() string {
	return "tharsis [global options] vcs-provider list [options]"
}

func (*vcsProviderListCommand) Description() string {
	return `
   Lists VCS providers within a namespace. Providers are
   inherited from parent groups and can be filtered with
   -include-inherited. Supports pagination and sorting.
`
}

func (*vcsProviderListCommand) Example() string {
	return `
tharsis vcs-provider list -namespace-path "<group_path>" -include-inherited -json
`
}

func (c *vcsProviderListCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.VCSProviderSortableField_value))

	f := flag.NewSet("Command options")
	f.StringVar(
		&c.namespacePath,
		"namespace-path",
		"Namespace path to list VCS providers for.",
		flag.Required(),
	)
	f.BoolVar(
		&c.includeInherited,
		"include-inherited",
		"Include VCS providers inherited from parent groups.",
		flag.Default(false),
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
		"Filter to VCS providers containing this substring.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
