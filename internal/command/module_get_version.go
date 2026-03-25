package command

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type moduleGetVersionCommand struct {
	*BaseCommand

	version *string
	toJSON  *bool
}

// NewModuleGetVersionCommandFactory returns a moduleGetVersionCommand struct.
func NewModuleGetVersionCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleGetVersionCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *moduleGetVersionCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *moduleGetVersionCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("module get-version"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	moduleVersionID := c.arguments[0]

	// Handle deprecated -version flag and module path arg.
	if c.version != nil {
		moduleVersionID = trn.NewResourceTRN(trn.ResourceTypeTerraformModuleVersion, moduleVersionID, *c.version)
	}

	input := &pb.GetTerraformModuleVersionByIDRequest{
		Id: moduleVersionID,
	}

	version, err := c.grpcClient.TerraformModulesClient.GetTerraformModuleVersionByID(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get module version")
		return 1
	}

	return outputModuleVersion(c.UI, *c.toJSON, version)
}

func (*moduleGetVersionCommand) Synopsis() string {
	return "Get a module version by ID or TRN."
}

func (*moduleGetVersionCommand) Description() string {
	return `
   The module get-version command retrieves details about a specific module version.
`
}

func (*moduleGetVersionCommand) Usage() string {
	return "tharsis [global options] module get-version [options] <version-id>"
}

func (*moduleGetVersionCommand) Example() string {
	return `
tharsis module get-version trn:terraform_module_version:<group_path>/<module_name>/<system>/<version>
`
}

func (c *moduleGetVersionCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Output in JSON format.",
		flag.Default(false),
	)
	f.StringVar(
		&c.version,
		"version",
		"A semver compliant version tag to use as a filter.",
		flag.Deprecated("pass version TRN as argument"),
	)

	return f
}

func outputModuleVersion(ui terminal.UI, toJSON bool, version *pb.TerraformModuleVersion) int {
	if toJSON {
		if err := ui.JSON(version); err != nil {
			ui.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		values := []terminal.NamedValue{
			{Name: "ID", Value: version.Metadata.Id},
			{Name: "TRN", Value: version.Metadata.Trn},
			{Name: "Version", Value: version.SemanticVersion},
			{Name: "Status", Value: version.Status},
			{Name: "Latest", Value: version.Latest},
			{Name: "SHA Sum", Value: version.ShaSum},
			{Name: "Created By", Value: version.CreatedBy},
			{Name: "Created At", Value: version.Metadata.CreatedAt.AsTime().Local().Format(humanTimeFormat)},
		}

		if version.Error != "" {
			values = append(values, terminal.NamedValue{Name: "Error", Value: version.Error})
		}

		if len(version.Submodules) > 0 {
			values = append(values, terminal.NamedValue{Name: "Submodules", Value: strings.Join(version.Submodules, ", ")})
		}

		ui.NamedValues(values)
	}

	return 0
}
