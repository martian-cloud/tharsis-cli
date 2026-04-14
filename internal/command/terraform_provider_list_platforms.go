package command

import (
	"errors"
	"maps"
	"slices"
	"strings"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type terraformProviderListPlatformsCommand struct {
	*BaseCommand

	limit             *int32
	cursor            *string
	providerVersionID *string
	providerID        *string
	operatingSystem   *string
	architecture      *string
	sortBy            *string
	toJSON            *bool
}

var _ Command = (*terraformProviderListPlatformsCommand)(nil)

// NewTerraformProviderListPlatformsCommandFactory returns a terraformProviderListPlatformsCommand struct.
func NewTerraformProviderListPlatformsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderListPlatformsCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *terraformProviderListPlatformsCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

func (c *terraformProviderListPlatformsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider list-platforms"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetTerraformProviderPlatformsRequest{
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		ProviderVersionId: c.providerVersionID,
		ProviderId:        c.providerID,
		OperatingSystem:   c.operatingSystem,
		Architecture:      c.architecture,
	}

	if c.sortBy != nil {
		input.Sort = pb.TerraformProviderPlatformSortableField(pb.TerraformProviderPlatformSortableField_value[*c.sortBy]).Enum()
	}

	result, err := c.grpcClient.TerraformProvidersClient.GetTerraformProviderPlatforms(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of terraform provider platforms")
		return 1
	}

	return c.OutputList(result, c.toJSON, "trn", "operating_system", "architecture", "binary_uploaded", "provider_version_id")
}

func (*terraformProviderListPlatformsCommand) Synopsis() string {
	return "Retrieve a paginated list of Terraform provider platforms."
}

func (*terraformProviderListPlatformsCommand) Usage() string {
	return "tharsis [global options] terraform-provider list-platforms [options]"
}

func (*terraformProviderListPlatformsCommand) Description() string {
	return `
   Lists platforms for a Terraform provider. Filter
   by provider version, OS, or architecture.
`
}

func (*terraformProviderListPlatformsCommand) Example() string {
	return `
tharsis terraform-provider list-platforms \
  -provider-version-id "<version_id>" \
  -operating-system "linux" \
  -architecture "amd64" \
  -json
`
}

func (c *terraformProviderListPlatformsCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.TerraformProviderPlatformSortableField_value))

	f := flag.NewSet("Command options")
	f.StringVar(
		&c.providerVersionID,
		"provider-version-id",
		"Filter to platforms for this provider version.",
	)
	f.StringVar(
		&c.providerID,
		"provider-id",
		"Filter to platforms for this provider.",
	)
	f.StringVar(
		&c.operatingSystem,
		"operating-system",
		"Filter to platforms with this operating system.",
	)
	f.StringVar(
		&c.architecture,
		"architecture",
		"Filter to platforms with this architecture.",
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
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
