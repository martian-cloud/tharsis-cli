package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/varparser"
)

type groupSetEnvironmentVarsCommand struct {
	*BaseCommand

	envVarFiles []string
}

var _ Command = (*groupSetEnvironmentVarsCommand)(nil)

// NewGroupSetEnvironmentVarsCommandFactory returns a groupSetEnvironmentVarsCommand struct.
func NewGroupSetEnvironmentVarsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupSetEnvironmentVarsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupSetEnvironmentVarsCommand) validate() error {
	const message = "group-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *groupSetEnvironmentVarsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group set-environment-vars"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: trn.ToTRN(trn.ResourceTypeGroup, c.arguments[0])})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get group")
		return 1
	}

	parser := varparser.NewVariableParser(nil, false)

	variables, err := parser.ParseEnvironmentVariables(&varparser.ParseEnvironmentVariablesInput{EnvVarFilePaths: c.envVarFiles})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to process environment variables")
		return 1
	}

	pbVariables := make([]*pb.SetNamespaceVariablesInputVariable, len(variables))
	for i, v := range variables {
		pbVariables[i] = &pb.SetNamespaceVariablesInputVariable{
			Key:   v.Key,
			Value: v.Value,
		}
	}

	input := &pb.SetNamespaceVariablesRequest{
		NamespacePath: group.FullPath,
		Category:      pb.VariableCategory_environment,
		Variables:     pbVariables,
	}

	if _, err = c.grpcClient.NamespaceVariablesClient.SetNamespaceVariables(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to set environment variables")
		return 1
	}

	c.UI.Successf("Environment variables set successfully in group!")
	return 0
}

func (*groupSetEnvironmentVarsCommand) Synopsis() string {
	return "Set environment variables for a group."
}

func (*groupSetEnvironmentVarsCommand) Description() string {
	return `
   The group set-environment-vars command sets environment variables for a group.
   Command will overwrite any existing environment variables in the target group!
   Note: This command does not support sensitive variables.
`
}

func (*groupSetEnvironmentVarsCommand) Usage() string {
	return "tharsis [global options] group set-environment-vars [options] <group-id>"
}

func (*groupSetEnvironmentVarsCommand) Example() string {
	return `
tharsis group set-environment-vars \
  -env-var-file "vars.env" \
  trn:group:<group_path>
`
}

func (c *groupSetEnvironmentVarsCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringSliceVar(
		&c.envVarFiles,
		"env-var-file",
		"Path to an environment variables file.",
		flag.Required(),
	)

	return f
}
