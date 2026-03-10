package command

import (
	"flag"
	"fmt"
	"maps"
	"strconv"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// managedIdentityAccessRuleUpdateCommand is the top-level structure for the managed identity access rule update command.
type managedIdentityAccessRuleUpdateCommand struct {
	*BaseCommand

	runStage                  pb.JobType
	allowedUsers              []string
	allowedServiceAccounts    []string
	allowedTeams              []string
	verifyStateLineage        *bool
	moduleAttestationPolicies []string
	toJSON                    bool
}

var _ Command = (*managedIdentityAccessRuleUpdateCommand)(nil)

func (c *managedIdentityAccessRuleUpdateCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.runStage, validation.Required),
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
		RunStage:                  c.runStage,
		AllowedUsers:              c.allowedUsers,
		AllowedServiceAccounts:    c.allowedServiceAccounts,
		AllowedTeams:              c.allowedTeams,
		VerifyStateLineage:        c.verifyStateLineage,
		ModuleAttestationPolicies: policies,
	}

	c.Logger.Debug("managed identity access rule update input", "input", input)

	updatedRule, err := c.grpcClient.ManagedIdentitiesClient.UpdateManagedIdentityAccessRule(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update managed identity access rule")
		return 1
	}

	return outputManagedIdentityAccessRule(c.UI, c.toJSON, updatedRule)
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
  --run-stage apply \
  --allowed-user trn:user:john.smith \
  TV80ZG...
`
}

func (c *managedIdentityAccessRuleUpdateCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"run-stage",
		"The run stage: plan or apply.",
		func(s string) error {
			val, ok := pb.JobType_value[strings.ToUpper(s)]
			if !ok {
				return fmt.Errorf("invalid run stage: %s (valid stages: %v)", s, maps.Keys(pb.JobType_value))
			}
			c.runStage = pb.JobType(val)
			return nil
		},
	)
	f.Func(
		"allowed-user",
		"Allowed user ID. (This flag may be repeated)",
		func(s string) error {
			c.allowedUsers = append(c.allowedUsers, s)
			return nil
		},
	)
	f.Func(
		"allowed-service-account",
		"Allowed service account ID. (This flag may be repeated)",
		func(s string) error {
			c.allowedServiceAccounts = append(c.allowedServiceAccounts, s)
			return nil
		},
	)
	f.Func(
		"allowed-team",
		"Allowed team ID. (This flag may be repeated)",
		func(s string) error {
			c.allowedTeams = append(c.allowedTeams, s)
			return nil
		},
	)
	f.Func(
		"verify-state-lineage",
		"Verify state lineage (true or false).",
		func(s string) error {
			val, err := strconv.ParseBool(s)
			if err != nil {
				return err
			}
			c.verifyStateLineage = &val
			return nil
		},
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
