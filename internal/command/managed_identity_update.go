package command

import (
	"flag"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// managedIdentityUpdateCommand is the top-level structure for the managed identity update command.
type managedIdentityUpdateCommand struct {
	*BaseCommand

	description                      *string
	awsFederatedRole                 *string
	azureFederatedClientID           *string
	azureFederatedTenantID           *string
	tharsisFederatedServiceAccountID *string
	updateIdentityData               bool
	toJSON                           bool
}

var _ Command = (*managedIdentityUpdateCommand)(nil)

func (c *managedIdentityUpdateCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewManagedIdentityUpdateCommandFactory returns a managedIdentityUpdateCommand struct.
func NewManagedIdentityUpdateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &managedIdentityUpdateCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *managedIdentityUpdateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("managed-identity update"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	managedIdentityID := c.arguments[0]

	var data *string
	if c.updateIdentityData {
		identity, err := c.grpcClient.ManagedIdentitiesClient.GetManagedIdentityByID(c.Context, &pb.GetManagedIdentityByIDRequest{Id: managedIdentityID})
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to get managed identity")
			return 1
		}

		identityType := pb.ManagedIdentityType(pb.ManagedIdentityType_value[identity.Type])

		// Build data based on identity type and provided fields
		switch identityType {
		case pb.ManagedIdentityType_AWS_FEDERATED:
			if c.awsFederatedRole != nil {
				dataStr := fmt.Sprintf(`{"role":"%s"}`, *c.awsFederatedRole)
				data = &dataStr
			}
		case pb.ManagedIdentityType_AZURE_FEDERATED:
			if c.azureFederatedClientID != nil && c.azureFederatedTenantID != nil {
				dataStr := fmt.Sprintf(`{"clientId":"%s","tenantId":"%s"}`, *c.azureFederatedClientID, *c.azureFederatedTenantID)
				data = &dataStr
			}
		case pb.ManagedIdentityType_THARSIS_FEDERATED:
			if c.tharsisFederatedServiceAccountID != nil {
				dataStr := fmt.Sprintf(`{"serviceAccountId":"%s"}`, *c.tharsisFederatedServiceAccountID)
				data = &dataStr
			}
		default:
			c.UI.Errorf("unsupported identity type: %s", identity.Type)
			return 1
		}
	}

	input := &pb.UpdateManagedIdentityRequest{
		Id:          managedIdentityID,
		Description: c.description,
		Data:        data,
	}

	c.Logger.Debug("managed identity update input", "input", input)

	updatedIdentity, err := c.grpcClient.ManagedIdentitiesClient.UpdateManagedIdentity(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update a managed identity")
		return 1
	}

	return outputManagedIdentity(c.UI, c.toJSON, updatedIdentity)
}

func (*managedIdentityUpdateCommand) Synopsis() string {
	return "Update a managed identity."
}

func (*managedIdentityUpdateCommand) Usage() string {
	return "tharsis [global options] managed-identity update [options] <id>"
}

func (*managedIdentityUpdateCommand) Description() string {
	return `
   The managed-identity update command updates a managed identity.
   Currently, it supports updating the description and data.
   Shows final output as JSON, if specified.
`
}

func (*managedIdentityUpdateCommand) Example() string {
	return `
tharsis managed-identity update \
  --description "Updated AWS production role" \
  --aws-federated-role arn:aws:iam::123456789012:role/UpdatedRole \
  trn:managed_identity:ops/my-group/aws-prod
`
}

func (c *managedIdentityUpdateCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"description",
		"Description for the managed identity.",
		func(s string) error {
			c.description = &s
			return nil
		},
	)
	f.Func(
		"aws-federated-role",
		"AWS IAM role. (Only if type is aws_federated)",
		func(s string) error {
			c.awsFederatedRole = &s
			c.updateIdentityData = true
			return nil
		},
	)
	f.Func(
		"azure-federated-client-id",
		"Azure client ID. (Only if type is azure_federated)",
		func(s string) error {
			c.azureFederatedClientID = &s
			c.updateIdentityData = true
			return nil
		},
	)
	f.Func(
		"azure-federated-tenant-id",
		"Azure tenant ID. (Only if type is azure_federated)",
		func(s string) error {
			c.azureFederatedTenantID = &s
			c.updateIdentityData = true
			return nil
		},
	)
	f.Func(
		"tharsis-federated-service-account-id",
		"Tharsis service account ID or TRN. (Only if type is tharsis_federated)",
		func(s string) error {
			c.tharsisFederatedServiceAccountID = &s
			c.updateIdentityData = true
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
