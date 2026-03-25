package command

import (
	"encoding/base64"
	"fmt"
	"maps"
	"slices"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// managedIdentityCreateCommand is the top-level structure for the managed identity create command.
type managedIdentityCreateCommand struct {
	*BaseCommand

	name                               *string
	groupID                            *string
	identityType                       *string
	description                        *string
	awsFederatedRole                   *string
	azureFederatedClientID             *string
	azureFederatedTenantID             *string
	tharsisFederatedServiceAccountPath *string
	kubernetesFederatedAudience        *string
	toJSON                             *bool
}

var _ Command = (*managedIdentityCreateCommand)(nil)

func (c *managedIdentityCreateCommand) validate() error {
	if c.name != nil && len(c.arguments) > 0 {
		return fmt.Errorf("name must be provided as either an argument or -name flag, not both")
	}

	if c.name == nil && len(c.arguments) == 0 {
		return fmt.Errorf("name is required either as an argument or -name flag")
	}

	return validation.ValidateStruct(c,
		validation.Field(&c.groupID, validation.Required, validation.NotNil),
		validation.Field(&c.identityType, validation.Required, validation.NotNil),
		validation.Field(&c.awsFederatedRole,
			validation.When(c.identityType != nil && *c.identityType == pb.ManagedIdentityType_aws_federated.String(), validation.Required),
		),
		validation.Field(&c.azureFederatedClientID,
			validation.When(c.identityType != nil && *c.identityType == pb.ManagedIdentityType_azure_federated.String(), validation.Required),
		),
		validation.Field(&c.azureFederatedTenantID,
			validation.When(c.identityType != nil && *c.identityType == pb.ManagedIdentityType_azure_federated.String(), validation.Required),
		),
		validation.Field(&c.tharsisFederatedServiceAccountPath,
			validation.When(c.identityType != nil && *c.identityType == pb.ManagedIdentityType_tharsis_federated.String(), validation.Required),
		),
		validation.Field(&c.kubernetesFederatedAudience,
			validation.When(c.identityType != nil && *c.identityType == pb.ManagedIdentityType_kubernetes_federated.String(), validation.Required),
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

	// Deprecated -name flag support.
	if c.name != nil {
		c.arguments = append(c.arguments, *c.name)
	}

	encodedData, err := c.encodeIdentityData()
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to encode managed identity data")
		return 1
	}

	input := &pb.CreateManagedIdentityRequest{
		Type:        pb.ManagedIdentityType(pb.ManagedIdentityType_value[*c.identityType]),
		Name:        c.arguments[0],
		Description: ptr.ToString(c.description),
		GroupId:     *c.groupID,
		Data:        encodedData,
	}

	createdIdentity, err := c.grpcClient.ManagedIdentitiesClient.CreateManagedIdentity(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create a managed identity")
		return 1
	}

	return outputManagedIdentity(c.UI, *c.toJSON, createdIdentity)
}

func (c *managedIdentityCreateCommand) encodeIdentityData() (string, error) {
	dataMap := map[string]string{
		pb.ManagedIdentityType_aws_federated.String():        fmt.Sprintf(`{"role":"%s"}`, ptr.ToString(c.awsFederatedRole)),
		pb.ManagedIdentityType_azure_federated.String():      fmt.Sprintf(`{"clientId":"%s","tenantId":"%s"}`, ptr.ToString(c.azureFederatedClientID), ptr.ToString(c.azureFederatedTenantID)),
		pb.ManagedIdentityType_tharsis_federated.String():    fmt.Sprintf(`{"serviceAccountPath":"%s"}`, ptr.ToString(c.tharsisFederatedServiceAccountPath)),
		pb.ManagedIdentityType_kubernetes_federated.String(): fmt.Sprintf(`{"audience":"%s"}`, ptr.ToString(c.kubernetesFederatedAudience)),
	}

	dataToEncode, ok := dataMap[*c.identityType]
	if !ok {
		return "", fmt.Errorf("unknown managed identity type %s", *c.identityType)
	}

	return base64.StdEncoding.EncodeToString([]byte(dataToEncode)), nil
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
  -group-id trn:group:<group_path> \
  -type aws_federated \
  -aws-federated-role arn:aws:iam::123456789012:role/MyRole \
  -description "AWS production role" \
  aws-prod
`
}

func (c *managedIdentityCreateCommand) Flags() *flag.Set {
	identityTypes := slices.Collect(maps.Keys(pb.ManagedIdentityType_value))

	f := flag.NewSet("Command options")
	f.StringVar(
		&c.name,
		"name",
		"The name of the managed identity.",
		flag.Deprecated("pass name as an argument"),
	)
	f.StringVar(
		&c.groupID,
		"group-id",
		"Group ID or TRN where the managed identity will be created.",
	)
	f.StringVar(
		&c.groupID,
		"group-path",
		"The group path where the managed identity will be created.",
		flag.Deprecated("use -group-id"),
		flag.TransformString(func(s string) string {
			return trn.NewResourceTRN(trn.ResourceTypeGroup, s)
		}),
	)
	f.StringVar(
		&c.identityType,
		"type",
		"The type of managed identity.",
		flag.ValidValues(identityTypes...),
		flag.PredictValues(identityTypes...),
	)
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
