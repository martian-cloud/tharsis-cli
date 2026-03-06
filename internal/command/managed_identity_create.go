package command

import (
	"flag"
	"fmt"
	"maps"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// managedIdentityCreateCommand is the top-level structure for the managed identity create command.
type managedIdentityCreateCommand struct {
	*BaseCommand

	groupID                          string
	identityType                     pb.ManagedIdentityType
	description                      string
	awsFederatedRole                 *string
	azureFederatedClientID           *string
	azureFederatedTenantID           *string
	tharsisFederatedServiceAccountID *string
	toJSON                           bool
}

var _ Command = (*managedIdentityCreateCommand)(nil)

func (c *managedIdentityCreateCommand) validate() error {
	const message = "name is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.groupID, validation.Required),
		validation.Field(&c.identityType, validation.Required),
		validation.Field(&c.awsFederatedRole,
			validation.When(c.identityType == pb.ManagedIdentityType_AWS_FEDERATED, validation.Required),
		),
		validation.Field(&c.azureFederatedClientID,
			validation.When(c.identityType == pb.ManagedIdentityType_AZURE_FEDERATED, validation.Required),
		),
		validation.Field(&c.azureFederatedTenantID,
			validation.When(c.identityType == pb.ManagedIdentityType_AZURE_FEDERATED, validation.Required),
		),
		validation.Field(&c.tharsisFederatedServiceAccountID,
			validation.When(c.identityType == pb.ManagedIdentityType_THARSIS_FEDERATED, validation.Required),
		),
	)
}

// NewManagedIdentityCreateCommandFactory returns a managedIdentityCreateCommand struct.
func NewManagedIdentityCreateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &managedIdentityCreateCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *managedIdentityCreateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("managed-identity create"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.CreateManagedIdentityRequest{
		Type:        c.identityType,
		Name:        c.arguments[0],
		Description: c.description,
		GroupId:     c.groupID,
		Data:        c.buildIdentityData(),
	}

	c.Logger.Debug("managed identity create input", "input", input)

	createdIdentity, err := c.client.ManagedIdentitiesClient.CreateManagedIdentity(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create a managed identity")
		return 1
	}

	return outputManagedIdentity(c.UI, c.toJSON, createdIdentity)
}

func (c *managedIdentityCreateCommand) buildIdentityData() string {
	dataMap := map[pb.ManagedIdentityType]string{
		pb.ManagedIdentityType_AWS_FEDERATED:     fmt.Sprintf(`{"role":"%s"}`, *c.awsFederatedRole),
		pb.ManagedIdentityType_AZURE_FEDERATED:   fmt.Sprintf(`{"clientId":"%s","tenantId":"%s"}`, *c.azureFederatedClientID, *c.azureFederatedTenantID),
		pb.ManagedIdentityType_THARSIS_FEDERATED: fmt.Sprintf(`{"serviceAccountId":"%s"}`, *c.tharsisFederatedServiceAccountID),
	}
	return dataMap[c.identityType]
}

func (*managedIdentityCreateCommand) Synopsis() string {
	return "Create a new managed identity."
}

func (*managedIdentityCreateCommand) Usage() string {
	return "tharsis [global options] managed-identity create [options] <name>"
}

func (*managedIdentityCreateCommand) Description() string {
	return `
   The managed-identity create command creates a new managed identity.
`
}

func (*managedIdentityCreateCommand) Example() string {
	return `
tharsis managed-identity create \
  --group-id trn:group:ops/my-group \
  --type aws_federated \
  --aws-federated-role arn:aws:iam::123456789012:role/MyRole \
  --description "AWS production role" \
  aws-prod
`
}

func (c *managedIdentityCreateCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.groupID,
		"group-id",
		"",
		"Group ID or TRN where the managed identity will be created.",
	)
	f.Func(
		"type",
		"The type of managed identity: aws_federated, azure_federated, tharsis_federated.",
		func(s string) error {
			val, ok := pb.ManagedIdentityType_value[strings.ToUpper(s)]
			if !ok {
				return fmt.Errorf("invalid identity type: %s (valid types: %v)", s, maps.Keys(pb.ManagedIdentityType_value))
			}
			c.identityType = pb.ManagedIdentityType(val)
			return nil
		},
	)
	f.StringVar(
		&c.description,
		"description",
		"",
		"Description for the managed identity.",
	)
	f.Func(
		"aws-federated-role",
		"AWS IAM role. (Only if type is aws_federated)",
		func(s string) error {
			c.awsFederatedRole = &s
			return nil
		},
	)
	f.Func(
		"azure-federated-client-id",
		"Azure client ID. (Only if type is azure_federated)",
		func(s string) error {
			c.azureFederatedClientID = &s
			return nil
		},
	)
	f.Func(
		"azure-federated-tenant-id",
		"Azure tenant ID. (Only if type is azure_federated)",
		func(s string) error {
			c.azureFederatedTenantID = &s
			return nil
		},
	)
	f.Func(
		"tharsis-federated-service-account-id",
		"Tharsis service account ID or TRN. (Only if type is tharsis_federated)",
		func(s string) error {
			c.tharsisFederatedServiceAccountID = &s
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
