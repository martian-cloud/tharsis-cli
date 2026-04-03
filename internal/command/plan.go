package command

import (
	"errors"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/run"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type planCommand struct {
	*BaseCommand

	directoryPath    *string
	moduleSource     *string
	moduleVersion    *string
	terraformVersion *string
	tfVarFiles       []string
	envVarFiles      []string
	tfVariables      []string
	envVariables     []string
	targetAddresses  []string
	destroy          *bool
	refresh          *bool
	refreshOnly      *bool
}

var _ Command = (*planCommand)(nil)

// NewPlanCommandFactory returns a planCommand struct.
func NewPlanCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &planCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *planCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: workspace id")
	}

	return nil
}

func (c *planCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("plan"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	curSettings, err := c.getCurrentSettings()
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get current settings")
		return 1
	}

	tokenGetter, err := curSettings.CurrentProfile.NewTokenGetter(c.Context)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create token getter")
		return 1
	}

	runMgr, err := run.NewManager(c.grpcClient, tokenGetter, c.HTTPClient, curSettings.CurrentProfile.Endpoint, c.Logger, c.UI)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create run manager")
		return 1
	}

	runResult, err := runMgr.CreateRun(c.Context, &run.CreateRunInput{
		WorkspaceID:      trn.ToTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
		DirectoryPath:    c.directoryPath,
		ModuleSource:     c.moduleSource,
		ModuleVersion:    c.moduleVersion,
		TerraformVersion: c.terraformVersion,
		TfVarFiles:       c.tfVarFiles,
		EnvVarFiles:      c.envVarFiles,
		TfVariables:      c.tfVariables,
		EnvVariables:     c.envVariables,
		TargetAddresses:  c.targetAddresses,
		IsDestroy:        *c.destroy,
		IsSpeculative:    true,
		Refresh:          *c.refresh,
		RefreshOnly:      *c.refreshOnly,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to plan")
		return 1
	}

	c.Logger.Debug("plan completed", "run_id", runResult.Metadata.Id, "status", runResult.Status)

	return 0
}

func (*planCommand) Synopsis() string {
	return "Create a speculative plan."
}

func (*planCommand) Description() string {
	return `
   Creates a speculative plan to preview infrastructure
   changes without applying them. Supports run-scoped
   Terraform and environment variables.

   Terraform variables may be passed in via supported
   options or from the environment with a 'TF_VAR_'
   prefix.

   Variable parsing precedence:
     1. Terraform variables from the environment.
     2. terraform.tfvars file from module's directory, if present.
     3. terraform.tfvars.json file from module's directory, if present.
     4. *.auto.tfvars, *.auto.tfvars.json files from the module's directory, if present.
     5. -tf-var-file option(s).
     6. -tf-var option(s).

   NOTE: If the same variable is assigned multiple values, the last value found will be used.
`
}

func (*planCommand) Usage() string {
	return "tharsis [global options] plan [options] <workspace-id>"
}

func (*planCommand) Example() string {
	return `
tharsis plan -directory-path "./terraform" trn:workspace:<workspace_path>
`
}

func (c *planCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.directoryPath,
		"directory-path",
		"The path of the root module's directory.",
	)
	f.StringVar(
		&c.moduleSource,
		"module-source",
		"Remote module source specification.",
	)
	f.StringVar(
		&c.moduleVersion,
		"module-version",
		"Remote module version number. Uses latest if empty.",
	)
	f.StringVar(
		&c.terraformVersion,
		"terraform-version",
		"The Terraform CLI version to use for the run.",
	)
	f.BoolVar(
		&c.destroy,
		"destroy",
		"Designates this run as a destroy operation.",
		flag.Default(false),
	)
	f.BoolVar(
		&c.refresh,
		"refresh",
		"Whether to do the usual refresh step.",
		flag.Default(true),
	)
	f.BoolVar(
		&c.refreshOnly,
		"refresh-only",
		"Whether to do ONLY a refresh operation.",
		flag.Default(false),
	)
	f.StringSliceVar(
		&c.tfVarFiles,
		"tf-var-file",
		"The path to a .tfvars variables file.",
	)
	f.StringSliceVar(
		&c.envVarFiles,
		"env-var-file",
		"The path to an environment variables file.",
	)
	f.StringSliceVar(
		&c.tfVariables,
		"tf-var",
		"A terraform variable as a key=value pair.",
	)
	f.StringSliceVar(
		&c.envVariables,
		"env-var",
		"An environment variable as a key=value pair.",
	)
	f.StringSliceVar(
		&c.targetAddresses,
		"target",
		"The Terraform address of the resources to be acted upon.",
	)

	f.MutuallyExclusive("directory-path", "module-source")

	return f
}
