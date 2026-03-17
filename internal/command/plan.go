package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
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
	destroy          bool
	refresh          bool
	refreshOnly      bool
}

// NewPlanCommandFactory returns a planCommand struct.
func NewPlanCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &planCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *planCommand) validate() error {
	const message = "workspace-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
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
		WorkspaceID:      toTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
		DirectoryPath:    c.directoryPath,
		ModuleSource:     c.moduleSource,
		ModuleVersion:    c.moduleVersion,
		TerraformVersion: c.terraformVersion,
		TfVarFiles:       c.tfVarFiles,
		EnvVarFiles:      c.envVarFiles,
		TfVariables:      c.tfVariables,
		EnvVariables:     c.envVariables,
		TargetAddresses:  c.targetAddresses,
		IsDestroy:        c.destroy,
		IsSpeculative:    true,
		Refresh:          c.refresh,
		RefreshOnly:      c.refreshOnly,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create run")
		return 1
	}

	c.Logger.Debug("plan completed", "run_id", runResult.Metadata.Id, "status", runResult.Status)

	return 0
}

func (*planCommand) Synopsis() string {
	return "Create a speculative plan"
}

func (*planCommand) Description() string {
	return `
   The plan command creates a speculative plan. It allows viewing
   the changes Terraform will make to your infrastructure
   without applying them. Supports setting run-scoped
   Terraform / environment variables and planning a
   destroy run.

   Terraform variables may be passed in via supported
   options or from the environment with a 'TF_VAR_'
   prefix.

   Variable parsing precedence:
     1. Terraform variables from the environment.
     2. terraform.tfvars file from module's directory, if present.
     3. terraform.tfvars.json file from module's directory, if present.
     4. *.auto.tfvars, *.auto.tfvars.json files from the module's directory, if present.
     5. --tf-var-file option(s).
     6. --tf-var option(s).

   NOTE: If the same variable is assigned multiple values, the last value found will be used.
`
}

func (*planCommand) Usage() string {
	return "tharsis [global options] plan [options] <workspace-id>"
}

func (*planCommand) Example() string {
	return `
tharsis plan --directory-path ./terraform trn:workspace:<workspace_path>
`
}

func (c *planCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)

	f.Func(
		"directory-path",
		"The path of the root module's directory.",
		func(s string) error {
			c.directoryPath = &s
			return nil
		},
	)
	f.Func(
		"module-source",
		"Remote module source specification.",
		func(s string) error {
			c.moduleSource = &s
			return nil
		},
	)
	f.Func(
		"module-version",
		"Remote module version number--defaults to latest.",
		func(s string) error {
			c.moduleVersion = &s
			return nil
		},
	)
	f.Func(
		"terraform-version",
		"The Terraform CLI version to use for the run.",
		func(s string) error {
			c.terraformVersion = &s
			return nil
		},
	)
	f.BoolVar(
		&c.destroy,
		"destroy",
		false,
		"Designates this run as a destroy operation.",
	)
	f.BoolVar(
		&c.refresh,
		"refresh",
		true,
		"Whether to do the usual refresh step.",
	)
	f.BoolVar(
		&c.refreshOnly,
		"refresh-only",
		false,
		"Whether to do ONLY a refresh operation.",
	)
	f.Func(
		"tf-var-file",
		"The path to a .tfvars variables file.",
		func(s string) error {
			c.tfVarFiles = append(c.tfVarFiles, s)
			return nil
		},
	)
	f.Func(
		"env-var-file",
		"The path to an environment variables file.",
		func(s string) error {
			c.envVarFiles = append(c.envVarFiles, s)
			return nil
		},
	)
	f.Func(
		"tf-var",
		"A terraform variable as a key=value pair.",
		func(s string) error {
			c.tfVariables = append(c.tfVariables, s)
			return nil
		},
	)
	f.Func(
		"env-var",
		"An environment variable as a key=value pair.",
		func(s string) error {
			c.envVariables = append(c.envVariables, s)
			return nil
		},
	)
	f.Func(
		"target",
		"The Terraform address of the resources to be acted upon.",
		func(s string) error {
			c.targetAddresses = append(c.targetAddresses, s)
			return nil
		},
	)

	return f
}
