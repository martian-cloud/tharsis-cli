package command

import (
	"flag"
	"fmt"
	"maps"
	"strings"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type terraformProviderMirrorListPlatformsCommand struct {
	*BaseCommand

	limit        int
	cursor       *string
	os           *string
	architecture *string
	sortBy       *pb.TerraformProviderPlatformMirrorSortableField
	toJSON       bool
}

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
		validation.Field(&c.limit, validation.Min(0), validation.Max(maxPaginationLimit)),
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

	input := &pb.GetTerraformProviderPlatformMirrorsRequest{
		VersionMirrorId: c.arguments[0],
		Sort:            c.sortBy,
		PaginationOptions: &pb.PaginationOptions{
			First: ptr.Int32(int32(c.limit)),
			After: c.cursor,
		},
		Os:           c.os,
		Architecture: c.architecture,
	}

	c.Logger.Debug("terraform-provider-mirror list-platforms input", "input", input)

	result, err := c.grpcClient.TerraformProviderMirrorsClient.GetTerraformProviderPlatformMirrors(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of provider platform mirrors")
		return 1
	}

	if c.toJSON {
		if err := c.UI.JSON(result); err != nil {
			c.UI.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		t := terminal.NewTable("id", "os", "architecture")

		for _, mirror := range result.PlatformMirrors {
			t.Rich([]string{
				mirror.Metadata.Id,
				mirror.Os,
				mirror.Architecture,
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
  --os linux \
  --architecture amd64 \
  --sort-by CREATED_AT_DESC \
  trn:terraform_provider_version_mirror:<group_path>/<provider_namespace>/<provider_name>/<semantic_version>
`
}

func (c *terraformProviderMirrorListPlatformsCommand) Flags() *flag.FlagSet {
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
			value, ok := pb.TerraformProviderPlatformMirrorSortableField_value[strings.ToUpper(s)]
			if !ok {
				return fmt.Errorf("invalid sort-by value: %s (valid values: %v)", s, maps.Keys(pb.TerraformProviderPlatformMirrorSortableField_value))
			}
			c.sortBy = pb.TerraformProviderPlatformMirrorSortableField(value).Enum()
			return nil
		},
	)
	f.Func(
		"os",
		"Filter to platforms with this OS.",
		func(s string) error {
			c.os = &s
			return nil
		},
	)
	f.Func(
		"architecture",
		"Filter to platforms with this architecture.",
		func(s string) error {
			c.architecture = &s
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
