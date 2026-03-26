package command

import (
	"encoding/base64"
	"fmt"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// managedIdentityUpdateCommand is the top-level structure for the managed identity update command.
type managedIdentityUpdateCommand struct {
	*BaseCommand

	description                        *string
	awsFederatedRole                   *string
	azureFederatedClientID             *string
	azureFederatedTenantID             *string
	tharsisFederatedServiceAccountPath *string
	kubernetesFederatedAudience        *string
	toJSON                             *bool
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

func (c *managedIdentityUpdateCommand) hasDataUpdate() bool {
	return c.awsFederatedRole != nil ||
		c.azureFederatedClientID != nil ||
		c.azureFederatedTenantID != nil ||
		c.tharsisFederatedServiceAccountPath != nil ||
		c.kubernetesFederatedAudience != nil
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

	managedIdentityID := trn.ToTRN(trn.ResourceTypeManagedIdentity, c.arguments[0])

	var encodedData *string
	if c.hasDataUpdate() {
		identity, err := c.grpcClient.ManagedIdentitiesClient.GetManagedIdentityByID(c.Context, &pb.GetManagedIdentityByIDRequest{Id: managedIdentityID})
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to get managed identity")
			return 1
		}

		// Build data based on identity type and provided fields.
		var data string
		switch identity.Type {
		case pb.ManagedIdentityType_aws_federated.String():
			if c.awsFederatedRole != nil {
				data = fmt.Sprintf(`{"role":"%s"}`, *c.awsFederatedRole)
			}
		case pb.ManagedIdentityType_azure_federated.String():
			if c.azureFederatedClientID != nil && c.azureFederatedTenantID != nil {
				data = fmt.Sprintf(`{"clientId":"%s","tenantId":"%s"}`, *c.azureFederatedClientID, *c.azureFederatedTenantID)
			}
		case pb.ManagedIdentityType_tharsis_federated.String():
			if c.tharsisFederatedServiceAccountPath != nil {
				data = fmt.Sprintf(`{"serviceAccountPath":"%s"}`, *c.tharsisFederatedServiceAccountPath)
			}
		case pb.ManagedIdentityType_kubernetes_federated.String():
			if c.kubernetesFederatedAudience != nil {
				data = fmt.Sprintf(`{"audience":"%s"}`, *c.kubernetesFederatedAudience)
			}
		default:
			c.UI.Errorf("unsupported identity type: %s", identity.Type)
			return 1
		}

		if data == "" {
			c.UI.Errorf("no valid data provided for managed identity type %s", identity.Type)
			return 1
		}

		encodedData = ptr.String(base64.StdEncoding.EncodeToString([]byte(data)))
	}

	input := &pb.UpdateManagedIdentityRequest{
		Id:          managedIdentityID,
		Description: c.description,
		Data:        encodedData,
	}

	updatedIdentity, err := c.grpcClient.ManagedIdentitiesClient.UpdateManagedIdentity(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update a managed identity")
		return 1
	}

	return c.OutputProto(updatedIdentity, c.toJSON)
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
  -description "Updated AWS production role" \
  -aws-federated-role arn:aws:iam::123456789012:role/UpdatedRole \
  trn:managed_identity:<group_path>/<managed_identity_name>
`
}

func (c *managedIdentityUpdateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.description,
		"description",
		"Description for the managed identity.",
	)
	f.StringVar(
		&c.awsFederatedRole,
		"aws-federated-role",
		"AWS IAM role. (Only if type is aws_federated)",
	)
	f.StringVar(
		&c.azureFederatedClientID,
		"azure-federated-client-id",
		"Azure client ID. (Only if type is azure_federated)",
	)
	f.StringVar(
		&c.azureFederatedTenantID,
		"azure-federated-tenant-id",
		"Azure tenant ID. (Only if type is azure_federated)",
	)
	f.StringVar(
		&c.tharsisFederatedServiceAccountPath,
		"tharsis-federated-service-account-path",
		"Tharsis service account path this managed identity will assume. (Only if type is tharsis_federated)",
	)
	f.StringVar(
		&c.kubernetesFederatedAudience,
		"kubernetes-federated-audience",
		"Kubernetes federated audience. The audience should match the client_id configured in your EKS OIDC identity provider. (Only if type is kubernetes_federated)",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
		flag.Default(false),
	)

	return f
}
