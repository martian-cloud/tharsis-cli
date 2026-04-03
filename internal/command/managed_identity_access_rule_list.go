package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type managedIdentityAccessRuleListCommand struct {
	*BaseCommand

	managedIdentityID *string
	toJSON            *bool
}

var _ Command = (*managedIdentityAccessRuleListCommand)(nil)

// NewManagedIdentityAccessRuleListCommandFactory returns a managedIdentityAccessRuleListCommand struct.
func NewManagedIdentityAccessRuleListCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &managedIdentityAccessRuleListCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *managedIdentityAccessRuleListCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	if c.managedIdentityID == nil {
		return errors.New("managed identity id is required")
	}

	return nil
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

	return c.OutputList(result, c.toJSON, "id", "type", "run_stage", "verify_state_lineage")
}

func (*managedIdentityAccessRuleListCommand) Synopsis() string {
	return "Retrieve a list of managed identity access rules."
}

func (*managedIdentityAccessRuleListCommand) Description() string {
	return `
   Lists all access rules for a managed identity.
`
}

func (*managedIdentityAccessRuleListCommand) Usage() string {
	return "tharsis [global options] managed-identity-access-rule list [options]"
}

func (*managedIdentityAccessRuleListCommand) Example() string {
	return `
tharsis managed-identity-access-rule list \
  -managed-identity-id "trn:managed_identity:<group_path>/<managed_identity_name>"
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
	)

	return f
}
