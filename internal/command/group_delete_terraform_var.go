package command

import (
	"flag"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type groupDeleteTerraformVarCommand struct {
	*BaseCommand

	key     string
	version *int64
}

// NewGroupDeleteTerraformVarCommandFactory returns a groupDeleteTerraformVarCommand struct.
func NewGroupDeleteTerraformVarCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupDeleteTerraformVarCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupDeleteTerraformVarCommand) validate() error {
	const message = "group-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.key, validation.Required),
	)
}

func (c *groupDeleteTerraformVarCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group delete-terraform-var"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	// Get group to retrieve full path
	group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: toTRN(trn.ResourceTypeGroup, c.arguments[0])})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get group")
		return 1
	}

	// Build TRN: trn:variable:namespace-path/terraform/key
	variableTRN := trn.NewResourceTRN(trn.ResourceTypeVariable, group.FullPath, "terraform", c.key)

	deleteInput := &pb.DeleteNamespaceVariableRequest{
		Id:      variableTRN,
		Version: c.version,
	}

	c.Logger.Debug("group delete-terraform-var input", "input", deleteInput)

	if _, err = c.grpcClient.NamespaceVariablesClient.DeleteNamespaceVariable(c.Context, deleteInput); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete terraform variable")
		return 1
	}

	c.UI.Successf("Terraform variable deleted successfully!")
	return 0
}

func (*groupDeleteTerraformVarCommand) Synopsis() string {
	return "Delete a terraform variable from a group."
}

func (*groupDeleteTerraformVarCommand) Description() string {
	return `
   The group delete-terraform-var command deletes a terraform variable from a group.
`
}

func (*groupDeleteTerraformVarCommand) Usage() string {
	return "tharsis [global options] group delete-terraform-var [options] <group-id>"
}

func (*groupDeleteTerraformVarCommand) Example() string {
	return `
tharsis group delete-terraform-var \
  --key region \
  trn:group:<group_path>
`
}

func (c *groupDeleteTerraformVarCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.key,
		"key",
		"",
		"Variable key.",
	)
	f.Func(
		"version",
		"Metadata version of the resource to be deleted. In most cases, this is not required.",
		func(s string) error {
			v, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return err
			}
			c.version = &v
			return nil
		},
	)

	return f
}
