package command

import (
	"flag"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// moduleCreateCommand is the top-level structure for the module create command.
type moduleCreateCommand struct {
	*BaseCommand

	groupID       string
	repositoryURL string
	private       bool
	toJSON        bool
	ifNotExists   bool
}

var _ Command = (*moduleCreateCommand)(nil)

func (c *moduleCreateCommand) validate() error {
	const message = "module-name/system is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewModuleCreateCommandFactory returns a moduleCreateCommand struct.
func NewModuleCreateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleCreateCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *moduleCreateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("module create"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	moduleArg := c.arguments[0]

	// Parse module-name/system
	parts := strings.Split(moduleArg, "/")
	var name, system string
	switch len(parts) {
	case 1:
		c.UI.Errorf("argument must be in format: module-name/system or group-path/module-name/system")
		return 1
	case 2:
		if c.groupID == "" {
			c.UI.Errorf("group-id is required when supplying just the module-name/system in the argument")
			return 1
		}
		name = parts[0]
		system = parts[1]
	default:
		if c.groupID != "" {
			c.UI.Errorf("group-id should not be supplied when supplying just the module path in the argument")
			return 1
		}

		// Handle deprecated syntax by extracting name, system, and group path.
		system = parts[len(parts)-1]
		name = parts[len(parts)-2]
		groupPath := strings.Join(parts[:len(parts)-2], "/")
		c.groupID = trn.NewResourceTRN(trn.ResourceTypeGroup, groupPath)
	}

	if c.ifNotExists {
		c.Logger.Debug("getting parent group", "value", c.groupID)

		group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: c.groupID})
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to get group")
			return 1
		}

		checkID := trn.NewResourceTRN(trn.ResourceTypeTerraformModule, group.FullPath, name, system)
		c.Logger.Debug("checking if module exists", "value", checkID)

		module, err := c.grpcClient.TerraformModulesClient.GetTerraformModuleByID(c.Context, &pb.GetTerraformModuleByIDRequest{Id: checkID})
		if err != nil && status.Code(err) != codes.NotFound {
			c.UI.ErrorWithSummary(err, "failed to check module")
			return 1
		}

		if module != nil {
			c.Logger.Debug("module already exists, returning existing module")
			return outputModule(c.UI, c.toJSON, module)
		}
	}

	input := &pb.CreateTerraformModuleRequest{
		Name:          name,
		System:        system,
		GroupId:       c.groupID,
		RepositoryUrl: c.repositoryURL,
		Private:       c.private,
	}

	c.Logger.Debug("module create input", "input", input)

	createdModule, err := c.grpcClient.TerraformModulesClient.CreateTerraformModule(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create a module")
		return 1
	}

	return outputModule(c.UI, c.toJSON, createdModule)
}

func (*moduleCreateCommand) Synopsis() string {
	return "Create a new Terraform module."
}

func (*moduleCreateCommand) Usage() string {
	return "tharsis [global options] module create [options] <module-name/system>"
}

func (*moduleCreateCommand) Description() string {
	return `
   The module create command creates a new Terraform module. It
   requires a group ID and repository URL. The argument should be
   in the format: module-name/system (e.g., vpc/aws). Shows final
   output as JSON, if specified. Idempotent when used with
   --if-not-exists option.
`
}

func (*moduleCreateCommand) Example() string {
	return `
tharsis module create \
  --group-id trn:group:<group_path> \
  --repository-url https://github.com/example/terraform-aws-vpc \
  --private \
  vpc/aws
`
}

func (c *moduleCreateCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.groupID,
		"group-id",
		"",
		"Parent group ID.",
	)
	f.StringVar(
		&c.repositoryURL,
		"repository-url",
		"",
		"The repository URL for the module.",
	)
	f.BoolVar(
		&c.private,
		"private",
		false,
		"Whether the module is private.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)
	f.BoolVar(
		&c.ifNotExists,
		"if-not-exists",
		false,
		"Create a module if it does not already exist.",
	)

	return f
}
