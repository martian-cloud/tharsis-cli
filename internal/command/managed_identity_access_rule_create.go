package command

import (
	"flag"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// managedIdentityAccessRuleCreateCommand is the top-level structure for the managed identity access rule create command.
type managedIdentityAccessRuleCreateCommand struct {
	*BaseCommand

	managedIdentityID         string
	ruleType                  *pb.ManagedIdentityAccessRuleType
	runStage                  *pb.JobType
	allowedUsers              []string
	allowedServiceAccounts    []string
	allowedTeams              []string
	verifyStateLineage        bool
	moduleAttestationPolicies []string
	toJSON                    bool
}

var _ Command = (*managedIdentityAccessRuleCreateCommand)(nil)

func (c *managedIdentityAccessRuleCreateCommand) validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.managedIdentityID, validation.Required),
		validation.Field(&c.ruleType, validation.NotNil),
		validation.Field(&c.runStage, validation.NotNil),
	)
}

// NewManagedIdentityAccessRuleCreateCommandFactory returns a managedIdentityAccessRuleCreateCommand struct.
func NewManagedIdentityAccessRuleCreateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &managedIdentityAccessRuleCreateCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *managedIdentityAccessRuleCreateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("managed-identity-access-rule create"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	policies, err := buildModuleAttestationPolicies(c.moduleAttestationPolicies)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to parse module attestation policies")
		return 1
	}

	input := &pb.CreateManagedIdentityAccessRuleRequest{
		Type:                      *c.ruleType,
		RunStage:                  *c.runStage,
		ManagedIdentityId:         c.managedIdentityID,
		AllowedUsers:              c.allowedUsers,
		AllowedServiceAccounts:    c.allowedServiceAccounts,
		AllowedTeams:              c.allowedTeams,
		VerifyStateLineage:        c.verifyStateLineage,
		ModuleAttestationPolicies: policies,
	}

	createdRule, err := c.grpcClient.ManagedIdentitiesClient.CreateManagedIdentityAccessRule(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create a managed identity access rule")
		return 1
	}

	return outputManagedIdentityAccessRule(c.UI, c.toJSON, createdRule)
}

func (*managedIdentityAccessRuleCreateCommand) Synopsis() string {
	return "Create a new managed identity access rule."
}

func (*managedIdentityAccessRuleCreateCommand) Usage() string {
	return "tharsis [global options] managed-identity-access-rule create [options]"
}

func (*managedIdentityAccessRuleCreateCommand) Description() string {
	return `
   The managed-identity-access-rule create command creates a new managed identity access rule.
`
}

func (*managedIdentityAccessRuleCreateCommand) Example() string {
	return `
tharsis managed-identity-access-rule create \
  --managed-identity-id trn:managed_identity:<group_path>/<managed_identity_name> \
  --rule-type eligible_principals \
  --run-stage plan \
  --allowed-user trn:user:<username> \
  --allowed-team trn:team:<team_name>
`
}

func (c *managedIdentityAccessRuleCreateCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.managedIdentityID,
		"managed-identity-id",
		"",
		"The ID or TRN of the managed identity.",
	)
	f.Func(
		"managed-identity-path",
		"Resource path to the managed identity. Deprecated.",
		func(s string) error {
			c.managedIdentityID = trn.NewResourceTRN(trn.ResourceTypeManagedIdentity, s)
			return nil
		},
	)
	f.Func(
		"rule-type",
		"The type of access rule: eligible_principals or module_attestation.",
		func(s string) error {
			val, ok := pb.ManagedIdentityAccessRuleType_value[strings.ToLower(s)]
			if !ok {
				return fmt.Errorf("invalid rule type: %s (valid types: %v)", s, slices.Collect(maps.Keys(pb.ManagedIdentityAccessRuleType_value)))
			}
			c.ruleType = pb.ManagedIdentityAccessRuleType(val).Enum()
			return nil
		},
	)
	f.Func(
		"run-stage",
		"The run stage: plan or apply.",
		func(s string) error {
			val, ok := pb.JobType_value[strings.ToLower(s)]
			if !ok {
				return fmt.Errorf("invalid run stage: %s (valid stages: %v)", s, slices.Collect(maps.Keys(pb.JobType_value)))
			}
			c.runStage = pb.JobType(val).Enum()
			return nil
		},
	)
	f.Func(
		"allowed-user",
		"Allowed user ID. (This flag may be repeated)",
		func(s string) error {
			c.allowedUsers = append(c.allowedUsers, trn.ToTRN(trn.ResourceTypeUser, s))
			return nil
		},
	)
	f.Func(
		"allowed-service-account",
		"Allowed service account ID. (This flag may be repeated)",
		func(s string) error {
			c.allowedServiceAccounts = append(c.allowedServiceAccounts, trn.ToTRN(trn.ResourceTypeServiceAccount, s))
			return nil
		},
	)
	f.Func(
		"allowed-team",
		"Allowed team ID. (This flag may be repeated)",
		func(s string) error {
			c.allowedTeams = append(c.allowedTeams, trn.ToTRN(trn.ResourceTypeTeam, s))
			return nil
		},
	)
	f.BoolVar(
		&c.verifyStateLineage,
		"verify-state-lineage",
		false,
		"Verify state lineage.",
	)
	f.Func(
		"module-attestation-policy",
		"Module attestation policy in format \"[PredicateType=someval,]PublicKeyFile=/path/to/file\". (This flag may be repeated)",
		func(s string) error {
			c.moduleAttestationPolicies = append(c.moduleAttestationPolicies, s)
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

func buildModuleAttestationPolicies(policies []string) ([]*pb.ManagedIdentityAccessRuleModuleAttestationPolicy, error) {
	var result []*pb.ManagedIdentityAccessRuleModuleAttestationPolicy

	for _, policy := range policies {
		var predicateType *string
		var filename string

		for _, kv := range strings.Split(policy, ",") {
			parts := strings.Split(kv, "=")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid module attestation policy format: %s", policy)
			}

			switch parts[0] {
			case "PredicateType":
				predicateType = &parts[1]
			case "PublicKeyFile":
				filename = parts[1]
			default:
				return nil, fmt.Errorf("invalid module attestation policy key: %s", parts[0])
			}
		}

		if filename == "" {
			return nil, fmt.Errorf("missing PublicKeyFile in module attestation policy: %s", policy)
		}

		publicKey, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read public key from file %s: %w", filename, err)
		}

		result = append(result, &pb.ManagedIdentityAccessRuleModuleAttestationPolicy{
			PredicateType: predicateType,
			PublicKey:     string(publicKey),
		})
	}

	return result, nil
}
