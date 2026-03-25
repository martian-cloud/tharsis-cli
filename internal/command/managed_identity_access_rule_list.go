package command

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type managedIdentityAccessRuleListCommand struct {
	*BaseCommand

	managedIdentityID *string
	toJSON            *bool
}

// NewManagedIdentityAccessRuleListCommandFactory returns a managedIdentityAccessRuleListCommand struct.
func NewManagedIdentityAccessRuleListCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &managedIdentityAccessRuleListCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *managedIdentityAccessRuleListCommand) validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments, validation.Empty),
		validation.Field(&c.managedIdentityID, validation.Required, validation.NotNil),
	)
}

func (c *managedIdentityAccessRuleListCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("managed-identity-access-rule list"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetManagedIdentityAccessRulesRequest{
		ManagedIdentityId: *c.managedIdentityID,
	}

	result, err := c.grpcClient.ManagedIdentitiesClient.GetManagedIdentityAccessRules(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of managed identity access rules")
		return 1
	}

	if *c.toJSON {
		if err := c.UI.JSON(result); err != nil {
			c.UI.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		t := terminal.NewTable("id", "type", "run_stage", "verify_state_lineage")

		for _, rule := range result.AccessRules {
			t.Rich([]string{
				rule.Metadata.Id,
				rule.Type,
				rule.RunStage,
				fmt.Sprintf("%t", rule.VerifyStateLineage),
			}, nil)
		}

		c.UI.Table(t)
	}

	return 0
}

func (*managedIdentityAccessRuleListCommand) Synopsis() string {
	return "Retrieve a list of managed identity access rules."
}

func (*managedIdentityAccessRuleListCommand) Description() string {
	return `
   The managed-identity-access-rule list command prints information about
   access rules for a specific managed identity.
`
}

func (*managedIdentityAccessRuleListCommand) Usage() string {
	return "tharsis [global options] managed-identity-access-rule list [options]"
}

func (*managedIdentityAccessRuleListCommand) Example() string {
	return `
tharsis managed-identity-access-rule list \
  -managed-identity-id trn:managed_identity:<group_path>/<managed_identity_name>
`
}

func (c *managedIdentityAccessRuleListCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.managedIdentityID,
		"managed-identity-id",
		"ID of the managed identity to get access rules for.",
	)
	f.StringVar(
		&c.managedIdentityID,
		"managed-identity-path",
		"Resource path of the managed identity to get access rules for.",
		flag.Deprecated("use -managed-identity-id"),
		flag.TransformString(func(s string) string {
			return trn.NewResourceTRN(trn.ResourceTypeManagedIdentity, s)
		}),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
		flag.Default(false),
	)

	return f
}
