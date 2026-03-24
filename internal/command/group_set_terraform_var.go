package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type groupSetTerraformVarCommand struct {
	*BaseCommand

	key       string
	value     string
	sensitive bool
}

// NewGroupSetTerraformVarCommandFactory returns a groupSetTerraformVarCommand struct.
func NewGroupSetTerraformVarCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupSetTerraformVarCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupSetTerraformVarCommand) validate() error {
	const message = "group-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.key, validation.Required),
		validation.Field(&c.value, validation.Required),
	)
}

func (c *groupSetTerraformVarCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group set-terraform-var"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	// Get group to retrieve full path
	group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: trn.ToTRN(trn.ResourceTypeGroup, c.arguments[0])})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get group")
		return 1
	}

	// Build TRN and check if variable exists
	variableTRN := trn.NewResourceTRN(trn.ResourceTypeVariable, group.FullPath, pb.VariableCategory_terraform.String(), c.key)
	existingVar, err := c.grpcClient.NamespaceVariablesClient.GetNamespaceVariableByID(c.Context, &pb.GetNamespaceVariableByIDRequest{
		Id: variableTRN,
	})
	if err != nil && status.Code(err) != codes.NotFound {
		c.UI.ErrorWithSummary(err, "failed to check existing variable")
		return 1
	}

	if existingVar != nil {
		// Variable exists - check if sensitivity matches
		if existingVar.Sensitive != c.sensitive {
			c.UI.Errorf("cannot change sensitive flag - delete and recreate the variable instead")
			return 1
		}

		// Update existing variable
		updateInput := &pb.UpdateNamespaceVariableRequest{
			Id:    existingVar.Metadata.Id,
			Key:   c.key,
			Value: c.value,
		}

		if _, err = c.grpcClient.NamespaceVariablesClient.UpdateNamespaceVariable(c.Context, updateInput); err != nil {
			c.UI.ErrorWithSummary(err, "failed to update terraform variable")
			return 1
		}
	} else {

		// Create new variable
		createInput := &pb.CreateNamespaceVariableRequest{
			NamespacePath: group.FullPath,
			Category:      pb.VariableCategory_terraform,
			Key:           c.key,
			Value:         c.value,
			Sensitive:     c.sensitive,
		}

		if _, err = c.grpcClient.NamespaceVariablesClient.CreateNamespaceVariable(c.Context, createInput); err != nil {
			c.UI.ErrorWithSummary(err, "failed to set terraform variable")
			return 1
		}
	}

	c.UI.Successf("Terraform variable set successfully in group!")
	return 0
}

func (*groupSetTerraformVarCommand) Synopsis() string {
	return "Set a terraform variable for a group."
}

func (*groupSetTerraformVarCommand) Description() string {
	return `
   The group set-terraform-var command creates or updates a terraform variable for a group.
`
}

func (*groupSetTerraformVarCommand) Usage() string {
	return "tharsis [global options] group set-terraform-var [options] <group-id>"
}

func (*groupSetTerraformVarCommand) Example() string {
	return `
tharsis group set-terraform-var \
  --key region \
  --value us-east-1 \
  trn:group:<group_path>
`
}

func (c *groupSetTerraformVarCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.key,
		"key",
		"",
		"Variable key.",
	)
	f.StringVar(
		&c.value,
		"value",
		"",
		"Variable value.",
	)
	f.BoolVar(
		&c.sensitive,
		"sensitive",
		false,
		"Mark variable as sensitive.",
	)

	return f
}
