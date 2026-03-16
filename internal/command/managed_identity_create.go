package command

import (
	"encoding/base64"
	"flag"
	"fmt"
	"maps"
	"slices"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// managedIdentityCreateCommand is the top-level structure for the managed identity create command.
type managedIdentityCreateCommand struct {
	*BaseCommand

	name                               *string
	groupID                            string
	identityType                       *pb.ManagedIdentityType
	description                        string
	awsFederatedRole                   string
	azureFederatedClientID             string
	azureFederatedTenantID             string
	tharsisFederatedServiceAccountPath string
	kubernetesFederatedAudience        string
	toJSON                             bool
}

var _ Command = (*managedIdentityCreateCommand)(nil)

func (c *managedIdentityCreateCommand) validate() error {
	const message = "name is required either as an argument or a flag"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.When(c.name == nil).Error(message),
		),
		validation.Field(&c.name, validation.Required.When(len(c.arguments) == 0).Error(message)),
		validation.Field(&c.groupID, validation.Required),
		validation.Field(&c.identityType, validation.NotNil),
		validation.Field(&c.awsFederatedRole,
			validation.When(c.identityType != nil && *c.identityType == pb.ManagedIdentityType_aws_federated, validation.Required),
		),
		validation.Field(&c.azureFederatedClientID,
			validation.When(c.identityType != nil && *c.identityType == pb.ManagedIdentityType_azure_federated, validation.Required),
		),
		validation.Field(&c.azureFederatedTenantID,
			validation.When(c.identityType != nil && *c.identityType == pb.ManagedIdentityType_azure_federated, validation.Required),
		),
		validation.Field(&c.tharsisFederatedServiceAccountPath,
			validation.When(c.identityType != nil && *c.identityType == pb.ManagedIdentityType_tharsis_federated, validation.Required),
		),
		validation.Field(&c.kubernetesFederatedAudience,
			validation.When(c.identityType != nil && *c.identityType == pb.ManagedIdentityType_kubernetes_federated, validation.Required),
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

	if c.name != nil {
		c.arguments = append(c.arguments, *c.name)
	}

	encodedData, err := c.encodeIdentityData()
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to encode managed identity data")
		return 1
	}

	input := &pb.CreateManagedIdentityRequest{
		Type:        *c.identityType,
		Name:        c.arguments[0],
		Description: c.description,
		GroupId:     c.groupID,
		Data:        encodedData,
	}

	createdIdentity, err := c.grpcClient.ManagedIdentitiesClient.CreateManagedIdentity(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create a managed identity")
		return 1
	}

	return outputManagedIdentity(c.UI, c.toJSON, createdIdentity)
}

func (c *managedIdentityCreateCommand) encodeIdentityData() (string, error) {
	dataMap := map[pb.ManagedIdentityType]string{
		pb.ManagedIdentityType_aws_federated:        fmt.Sprintf(`{"role":"%s"}`, c.awsFederatedRole),
		pb.ManagedIdentityType_azure_federated:      fmt.Sprintf(`{"clientId":"%s","tenantId":"%s"}`, c.azureFederatedClientID, c.azureFederatedTenantID),
		pb.ManagedIdentityType_tharsis_federated:    fmt.Sprintf(`{"serviceAccountPath":"%s"}`, c.tharsisFederatedServiceAccountPath),
		pb.ManagedIdentityType_kubernetes_federated: fmt.Sprintf(`{"audience":"%s"}`, c.kubernetesFederatedAudience),
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
  --group-id trn:group:<group_path> \
  --type aws_federated \
  --aws-federated-role arn:aws:iam::123456789012:role/MyRole \
  --description "AWS production role" \
  aws-prod
`
}

func (c *managedIdentityCreateCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"name",
		"The name of the managed identity. Deprecated",
		func(s string) error {
			c.name = &s
			return nil
		},
	)
	f.StringVar(
		&c.groupID,
		"group-id",
		"",
		"Group ID or TRN where the managed identity will be created.",
	)
	f.Func(
		"group-path",
		"The group path where the managed identity will be created. Deprecated.",
		func(s string) error {
			c.groupID = trn.NewResourceTRN(trn.ResourceTypeGroup, s)
			return nil
		},
	)
	f.Func(
		"type",
		"The type of managed identity: aws_federated, azure_federated, tharsis_federated, kubernetes_federated.",
		func(s string) error {
			val, ok := pb.ManagedIdentityType_value[strings.ToLower(s)]
			if !ok {
				return fmt.Errorf("invalid identity type: %s (valid types: %v)", s, slices.Collect(maps.Keys(pb.ManagedIdentityType_value)))
			}
			c.identityType = pb.ManagedIdentityType(val).Enum()
			return nil
		},
	)
	f.StringVar(
		&c.description,
		"description",
		"",
		"Description for the managed identity.",
	)
	f.StringVar(
		&c.awsFederatedRole,
		"aws-federated-role",
		"",
		"AWS IAM role. (Only if type is aws_federated)",
	)
	f.StringVar(
		&c.azureFederatedClientID,
		"azure-federated-client-id",
		"",
		"Azure client ID. (Only if type is azure_federated)",
	)
	f.StringVar(
		&c.azureFederatedTenantID,
		"azure-federated-tenant-id",
		"",
		"Azure tenant ID. (Only if type is azure_federated)",
	)
	f.StringVar(
		&c.tharsisFederatedServiceAccountPath,
		"tharsis-federated-service-account-path",
		"",
		"Tharsis service account path this managed identity will assume. (Only if type is tharsis_federated)",
	)
	f.StringVar(
		&c.kubernetesFederatedAudience,
		"kubernetes-federated-audience",
		"",
		"Kubernetes federated audience. The audience should match the client_id configured in your EKS OIDC identity provider. (Only if type is kubernetes_federated)",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}
