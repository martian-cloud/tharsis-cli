package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/run"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type destroyCommand struct {
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
	comment          string
	autoApprove      bool
	input            bool
	refresh          bool
}

// NewDestroyCommandFactory returns a destroyCommand struct.
func NewDestroyCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &destroyCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *destroyCommand) validate() error {
	const message = "workspace-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *destroyCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("destroy"),
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

	// Create non-speculative destroy run
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
		IsDestroy:        true,
		IsSpeculative:    false,
		Refresh:          c.refresh,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create run")
		return 1
	}

	// Check if plan has changes
	if runResult.Status == plannedAndFinished {
		c.UI.Output("Stopping since plan had no changes.")
		return 0
	}

	// Return if input is false and autoApprove is not set
	if !c.input && !c.autoApprove {
		c.UI.Output("Will not apply the plan since -input was false.")
		return 0
	}

	// Handle approval
	if c.autoApprove {
		c.UI.Output("\nAuto-approving.\n")
	} else {
		c.UI.Output("\nDo you approve to destroy the above resources?\n")
		answer, err := c.UI.Input(&terminal.Input{
			Prompt: "  only 'yes' will be accepted: ",
		})
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to ask for approval")
			return 1
		}
		if answer != "yes" {
			c.UI.Output("Approval response was negative. Will NOT destroy resources.")
			return 0
		}
		c.UI.Output("\n\n")
	}

	// Apply the destroy run
	appliedRun, err := runMgr.ApplyRun(c.Context, runResult.Metadata.Id)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to apply run")
		return 1
	}

	c.Logger.Debug("destroy completed", "run_id", appliedRun.Metadata.Id, "status", appliedRun.Status)

	return 0
}

func (*destroyCommand) Synopsis() string {
	return "Destroy workspace resources"
}

func (*destroyCommand) Description() string {
	return `
   The destroy command destroys resources in a workspace.
   It creates a destroy plan, then applies it after approval.
   Supports setting run-scoped Terraform / environment variables.

   Terraform variables may be passed in via supported
   options or from the environment with a 'TF_VAR_' prefix.
`
}

func (*destroyCommand) Usage() string {
	return "tharsis [global options] destroy [options] <workspace-id>"
}

func (*destroyCommand) Example() string {
	return `
tharsis destroy --directory-path ./terraform trn:workspace:<workspace_path>
`
}

func (c *destroyCommand) Flags() *flag.FlagSet {
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
	f.StringVar(
		&c.comment,
		"comment",
		"",
		"Comment for the destroy.",
	)
	f.BoolVar(
		&c.autoApprove,
		"auto-approve",
		false,
		"Skip interactive approval of the plan.",
	)
	f.BoolVar(
		&c.input,
		"input",
		true,
		"Ask for input for variables if not directly set.",
	)
	f.BoolVar(
		&c.refresh,
		"refresh",
		true,
		"Whether to do the usual refresh step.",
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
