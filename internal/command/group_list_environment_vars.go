package command

import (
	"time"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type groupListEnvironmentVarsCommand struct {
	*BaseCommand

	showSensitive *bool
	toJSON        *bool
}

// NewGroupListEnvironmentVarsCommandFactory returns a groupListEnvironmentVarsCommand struct.
func NewGroupListEnvironmentVarsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupListEnvironmentVarsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupListEnvironmentVarsCommand) validate() error {
	const message = "group-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *groupListEnvironmentVarsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group list-environment-vars"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: c.arguments[0]})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get group")
		return 1
	}

	input := &pb.GetNamespaceVariablesRequest{
		NamespacePath: group.FullPath,
	}

	result, err := c.grpcClient.NamespaceVariablesClient.GetNamespaceVariables(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to list environment variables")
		return 1
	}

	// Filter to only environment variables.
	var environmentVars []*pb.NamespaceVariable
	for _, v := range result.Variables {
		if v.Category == pb.VariableCategory_environment.String() {
			if v.Sensitive && !*c.showSensitive {
				v.Value = ptr.String("[SENSITIVE]")
			}

			environmentVars = append(environmentVars, v)
		}
	}

	// Fetch sensitive values if requested.
	if *c.showSensitive {
		for _, v := range environmentVars {
			if v.Sensitive && v.LatestVersionId != "" {
				versionInput := &pb.GetNamespaceVariableVersionByIDRequest{
					Id:                    v.LatestVersionId,
					IncludeSensitiveValue: true,
				}

				version, err := c.grpcClient.NamespaceVariablesClient.GetNamespaceVariableVersionByID(c.Context, versionInput)
				if err != nil {
					c.UI.ErrorWithSummary(err, "failed to get variable version")
					return 1
				}

				v.Value = version.Value
				// Rate limit to avoid overwhelming the API.
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	return c.OutputProtoList(environmentVars, c.toJSON)
}

func (*groupListEnvironmentVarsCommand) Synopsis() string {
	return "List all environment variables in a group."
}

func (*groupListEnvironmentVarsCommand) Description() string {
	return `
   The group list-environment-vars command retrieves all terraform
   variables from a group and its parent groups.
`
}

func (*groupListEnvironmentVarsCommand) Usage() string {
	return "tharsis [global options] group list-environment-vars [options] <group-id>"
}

func (*groupListEnvironmentVarsCommand) Example() string {
	return `
tharsis group list-environment-vars -show-sensitive trn:group:<group_path>
`
}

func (c *groupListEnvironmentVarsCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.showSensitive,
		"show-sensitive",
		"Show the actual values of sensitive variables (requires appropriate permissions).",
		flag.Default(false),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Output in JSON format.",
		flag.Default(false),
	)

	return f
}
