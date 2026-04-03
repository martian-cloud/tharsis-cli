package command

import (
	"errors"

	"github.com/aws/smithy-go/ptr"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type terraformProviderMirrorDeleteVersionCommand struct {
	*BaseCommand

	force *bool
}

var _ Command = (*terraformProviderMirrorDeleteVersionCommand)(nil)

// NewTerraformProviderMirrorDeleteVersionCommandFactory returns a terraformProviderMirrorDeleteVersionCommand struct.
func NewTerraformProviderMirrorDeleteVersionCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderMirrorDeleteVersionCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *terraformProviderMirrorDeleteVersionCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: version mirror id")
	}

	return nil
}

func (c *terraformProviderMirrorDeleteVersionCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider-mirror delete-version"),
		WithInputValidator(c.validate),
		WithClient(true),
		WithWarningPrompt("This will permanently delete the provider version mirror."),
	); code != 0 {
		return code
	}

	input := &pb.DeleteTerraformProviderVersionMirrorRequest{
		Id:    c.arguments[0],
		Force: ptr.ToBool(c.force),
	}

	if _, err := c.grpcClient.TerraformProviderMirrorsClient.DeleteTerraformProviderVersionMirror(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete terraform provider version mirror")
		return 1
	}

	c.UI.Successf("Terraform provider version mirror deleted successfully!")
	return 0
}

func (*terraformProviderMirrorDeleteVersionCommand) Synopsis() string {
	return "Delete a terraform provider version from mirror."
}

func (*terraformProviderMirrorDeleteVersionCommand) Description() string {
	return `
   Removes a mirrored provider version and its platform
   binaries. Use -force when the version hosts packages.
`
}

func (*terraformProviderMirrorDeleteVersionCommand) Usage() string {
	return "tharsis [global options] terraform-provider-mirror delete-version [options] <version-mirror-id>"
}

func (*terraformProviderMirrorDeleteVersionCommand) Example() string {
	return `
tharsis terraform-provider-mirror delete-version -force <version-mirror-id>
`
}

func (c *terraformProviderMirrorDeleteVersionCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.force,
		"force",
		"Skip confirmation prompt.",
		flag.Aliases("f"),
	)

	return f
}
