package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// managedIdentityAccessRuleUpdateCommand is the top-level structure for the managed identity access rule update command.
type managedIdentityAccessRuleUpdateCommand struct {
	*BaseCommand

	allowedUsers              []string
	allowedServiceAccounts    []string
	allowedTeams              []string
	moduleAttestationPolicies []string
	verifyStateLineage        *bool
	toJSON                    *bool
}

var _ Command = (*managedIdentityAccessRuleUpdateCommand)(nil)

func (c *managedIdentityAccessRuleUpdateCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewManagedIdentityAccessRuleUpdateCommandFactory returns a managedIdentityAccessRuleUpdateCommand struct.
func NewManagedIdentityAccessRuleUpdateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &managedIdentityAccessRuleUpdateCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *managedIdentityAccessRuleUpdateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("managed-identity-access-rule update"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	var policies []*pb.ManagedIdentityAccessRuleModuleAttestationPolicy
	if len(c.moduleAttestationPolicies) > 0 {
		var err error
		policies, err = buildModuleAttestationPolicies(c.moduleAttestationPolicies)
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to parse module attestation policies")
			return 1
		}
	}

	input := &pb.UpdateManagedIdentityAccessRuleRequest{
		Id:                        c.arguments[0],
		AllowedUsers:              c.allowedUsers,
		AllowedServiceAccounts:    c.allowedServiceAccounts,
		AllowedTeams:              c.allowedTeams,
		VerifyStateLineage:        c.verifyStateLineage,
		ModuleAttestationPolicies: policies,
	}

	updatedRule, err := c.grpcClient.ManagedIdentitiesClient.UpdateManagedIdentityAccessRule(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update managed identity access rule")
		return 1
	}

	return c.Output(updatedRule, c.toJSON)
}

func (*managedIdentityAccessRuleUpdateCommand) Synopsis() string {
	return "Update a managed identity access rule."
}

func (*managedIdentityAccessRuleUpdateCommand) Usage() string {
	return "tharsis [global options] managed-identity-access-rule update [options] <id>"
}

func (*managedIdentityAccessRuleUpdateCommand) Description() string {
	return `
   The managed-identity-access-rule update command updates an existing managed identity access rule.
`
}

func (*managedIdentityAccessRuleUpdateCommand) Example() string {
	return `
tharsis managed-identity-access-rule update \
  -allowed-user "trn:user:<username>" \
  <id>
`
}

func (c *managedIdentityAccessRuleUpdateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringSliceVar(
		&c.allowedUsers,
		"allowed-user",
		"Allowed user ID.",
		flag.TransformString(func(s string) string {
			return trn.ToTRN(trn.ResourceTypeUser, s)
		}),
	)
	f.StringSliceVar(
		&c.allowedServiceAccounts,
		"allowed-service-account",
		"Allowed service account ID.",
		flag.TransformString(func(s string) string {
			return trn.ToTRN(trn.ResourceTypeServiceAccount, s)
		}),
	)
	f.StringSliceVar(
		&c.allowedTeams,
		"allowed-team",
		"Allowed team ID.",
		flag.TransformString(func(s string) string {
			return trn.ToTRN(trn.ResourceTypeTeam, s)
		}),
	)
	f.BoolVar(
		&c.verifyStateLineage,
		"verify-state-lineage",
		"Verify state lineage.",
	)
	f.StringSliceVar(
		&c.moduleAttestationPolicies,
		"module-attestation-policy",
		"Module attestation policy in format \"[PredicateType=someval,]PublicKeyFile=/path/to/file\".",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
