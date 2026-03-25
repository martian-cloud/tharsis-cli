package command

import (
	"flag"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
)

// managedIdentityAccessRuleGetCommand is the top-level structure for the managed identity access rule get command.
type managedIdentityAccessRuleGetCommand struct {
	*BaseCommand

	toJSON bool
}

var _ Command = (*managedIdentityAccessRuleGetCommand)(nil)

func (c *managedIdentityAccessRuleGetCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewManagedIdentityAccessRuleGetCommandFactory returns a managedIdentityAccessRuleGetCommand struct.
func NewManagedIdentityAccessRuleGetCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &managedIdentityAccessRuleGetCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *managedIdentityAccessRuleGetCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("managed-identity-access-rule get"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetManagedIdentityAccessRuleByIDRequest{Id: c.arguments[0]}

	rule, err := c.grpcClient.ManagedIdentitiesClient.GetManagedIdentityAccessRuleByID(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get managed identity access rule")
		return 1
	}

	return outputManagedIdentityAccessRule(c.UI, c.toJSON, rule)
}

func (*managedIdentityAccessRuleGetCommand) Synopsis() string {
	return "Get a managed identity access rule."
}

func (*managedIdentityAccessRuleGetCommand) Usage() string {
	return "tharsis [global options] managed-identity-access-rule get [options] <id>"
}

func (*managedIdentityAccessRuleGetCommand) Description() string {
	return `
   The managed-identity-access-rule get command gets a managed identity access rule by ID.
`
}

func (*managedIdentityAccessRuleGetCommand) Example() string {
	return `
tharsis managed-identity-access-rule get <id>
`
}

func (c *managedIdentityAccessRuleGetCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}

func outputManagedIdentityAccessRule(ui terminal.UI, toJSON bool, rule *pb.ManagedIdentityAccessRule) int {
	if toJSON {
		if err := ui.JSON(rule); err != nil {
			ui.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		values := []terminal.NamedValue{
			{Name: "ID", Value: rule.Metadata.Id},
			{Name: "TRN", Value: rule.Metadata.Trn},
			{Name: "Type", Value: rule.Type},
			{Name: "Run Stage", Value: rule.RunStage},
			{Name: "Verify State Lineage", Value: rule.VerifyStateLineage},
		}

		if len(rule.AllowedUsers) > 0 {
			values = append(values, terminal.NamedValue{Name: "Allowed Users", Value: strings.Join(rule.AllowedUsers, ", ")})
		}

		if len(rule.AllowedServiceAccounts) > 0 {
			values = append(values, terminal.NamedValue{Name: "Allowed Service Accounts", Value: strings.Join(rule.AllowedServiceAccounts, ", ")})
		}

		if len(rule.AllowedTeams) > 0 {
			values = append(values, terminal.NamedValue{Name: "Allowed Teams", Value: strings.Join(rule.AllowedTeams, ", ")})
		}

		ui.NamedValues(values)
	}

	return 0
}
