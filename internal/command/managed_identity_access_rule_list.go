package command

import (
	"flag"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type managedIdentityAccessRuleListCommand struct {
	*BaseCommand

	toJSON bool
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
	const message = "managed-identity-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
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
		ManagedIdentityId: c.arguments[0],
	}

	c.Logger.Debug("managed-identity-access-rule list input", "input", input)

	result, err := c.client.ManagedIdentitiesClient.GetManagedIdentityAccessRules(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of managed identity access rules")
		return 1
	}

	if c.toJSON {
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
	return "tharsis [global options] managed-identity-access-rule list [options] <managed-identity-id>"
}

func (*managedIdentityAccessRuleListCommand) Example() string {
	return `
tharsis managed-identity-access-rule list \
  trn:managed_identity:ops/my-identity
`
}

func (c *managedIdentityAccessRuleListCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}
