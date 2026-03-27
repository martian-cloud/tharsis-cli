package command

import (
	"maps"
	"slices"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type terraformProviderMirrorListPlatformsCommand struct {
	*BaseCommand

	limit        *int32
	cursor       *string
	os           *string
	architecture *string
	sortBy       *string
	toJSON       *bool
}

var _ Command = (*terraformProviderMirrorListPlatformsCommand)(nil)

// NewTerraformProviderMirrorListPlatformsCommandFactory returns a terraformProviderMirrorListPlatformsCommand struct.
func NewTerraformProviderMirrorListPlatformsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderMirrorListPlatformsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *terraformProviderMirrorListPlatformsCommand) validate() error {
	const message = "version-mirror-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *terraformProviderMirrorListPlatformsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider-mirror list-platforms"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	var sortEnum *pb.TerraformProviderPlatformMirrorSortableField
	if c.sortBy != nil {
		v := pb.TerraformProviderPlatformMirrorSortableField(pb.TerraformProviderPlatformMirrorSortableField_value[strings.ToUpper(*c.sortBy)])
		sortEnum = &v
	}

	input := &pb.GetTerraformProviderPlatformMirrorsRequest{
		VersionMirrorId: c.arguments[0],
		Sort:            sortEnum,
		PaginationOptions: &pb.PaginationOptions{
			First: c.limit,
			After: c.cursor,
		},
		Os:           c.os,
		Architecture: c.architecture,
	}

	result, err := c.grpcClient.TerraformProviderMirrorsClient.GetTerraformProviderPlatformMirrors(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of provider platform mirrors")
		return 1
	}

	return c.OutputList(result, c.toJSON)
}

func (*terraformProviderMirrorListPlatformsCommand) Synopsis() string {
	return "Retrieve a paginated list of provider platform mirrors."
}

func (*terraformProviderMirrorListPlatformsCommand) Description() string {
	return `
   The terraform-provider-mirror list-platforms command prints information
   about provider platform mirrors for a version mirror. Supports pagination,
   filtering and sorting.
`
}

func (*terraformProviderMirrorListPlatformsCommand) Usage() string {
	return "tharsis [global options] terraform-provider-mirror list-platforms [options] <version-mirror-id>"
}

func (*terraformProviderMirrorListPlatformsCommand) Example() string {
	return `
tharsis terraform-provider-mirror list-platforms \
  -os linux \
  -architecture amd64 \
  -sort-by CREATED_AT_DESC \
  trn:terraform_provider_version_mirror:<group_path>/<provider_namespace>/<provider_name>/<semantic_version>
`
}

func (c *terraformProviderMirrorListPlatformsCommand) Flags() *flag.Set {
	sortValues := slices.Collect(maps.Keys(pb.TerraformProviderPlatformMirrorSortableField_value))

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
		&c.os,
		"os",
		"Filter to platforms with this OS.",
	)
	f.StringVar(
		&c.architecture,
		"architecture",
		"Filter to platforms with this architecture.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
		flag.Default(false),
	)

	return f
}
