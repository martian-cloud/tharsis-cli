package command

import (
	"errors"
	"maps"
	"slices"
	"strings"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type terraformProviderListVersionsCommand struct {
	*BaseCommand

	limit           *int32
	cursor          *string
	sortBy          *string
	providerID      *string
	semanticVersion *string
	latest          *bool
	toJSON          *bool
}

var _ Command = (*terraformProviderListVersionsCommand)(nil)

// NewTerraformProviderListVersionsCommandFactory returns a terraformProviderListVersionsCommand struct.
func NewTerraformProviderListVersionsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderListVersionsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *terraformProviderListVersionsCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

func (c *terraformProviderListVersionsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider list-versions"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetTerraformProviderVersionsRequest{
		ProviderId: *c.providerID,
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		SemanticVersion: c.semanticVersion,
		Latest:          c.latest,
	}

	if c.sortBy != nil {
		input.Sort = pb.TerraformProviderVersionSortableField(pb.TerraformProviderVersionSortableField_value[*c.sortBy]).Enum()
	}

	result, err := c.grpcClient.TerraformProvidersClient.GetTerraformProviderVersions(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to list terraform provider versions")
		return 1
	}

	return c.OutputList(result, c.toJSON, "trn", "version", "provider_id")
}

func (*terraformProviderListVersionsCommand) Synopsis() string {
	return "Retrieve a paginated list of terraform provider versions."
}

func (*terraformProviderListVersionsCommand) Description() string {
	return `
   Lists versions of a Terraform provider with
   pagination and sorting. Filter by semantic version
   or latest only.
`
}

func (*terraformProviderListVersionsCommand) Usage() string {
	return "tharsis [global options] terraform-provider list-versions [options]"
}

func (*terraformProviderListVersionsCommand) Example() string {
	return `
tharsis terraform-provider list-versions \
  -provider-id "<provider_id>" \
  -sort-by "CREATED_AT_DESC" \
  -limit 10
`
}

func (c *terraformProviderListVersionsCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.TerraformProviderVersionSortableField_value))

	f := flag.NewSet("Command options")
	f.StringVar(
		&c.providerID,
		"provider-id",
		"Provider ID or TRN to list versions for.",
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
		&c.semanticVersion,
		"semantic-version",
		"Filter to a specific semantic version.",
	)
	f.BoolVar(
		&c.latest,
		"latest",
		"Filter to only the latest version.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	f.MutuallyExclusive("latest", "semantic-version")

	return f
}
