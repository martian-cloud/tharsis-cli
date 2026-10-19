package command

import (
	"errors"
	"time"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/trn"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/run"
)

type destroyCommand struct {
	*BaseCommand

	directoryPath            *string
	moduleSource             *string
	moduleVersion            *string
	terraformVersion         *string
	tfVarFiles               []string
	envVarFiles              []string
	tfVariables              []string
	envVariables             []string
	targetAddresses          []string
	annotations              []string
	autoApprove              *bool
	input                    *bool
	waitTimeout              *time.Duration
	refresh                  *bool
	includeModulePrereleases *bool
}

var _ Command = (*destroyCommand)(nil)

// NewDestroyCommandFactory returns a destroyCommand struct.
func NewDestroyCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &destroyCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *destroyCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: workspace id")
	}

	return nil
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

	tokenResolver, err := curSettings.CurrentProfile.NewTokenResolver(c.Context)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create token getter")
		return 1
	}

	defer tokenResolver.Close()

	// The first Ctrl-C stops following and offers to cancel the run; a second exits immediately.
	ctx, cancelCtx, release := interruptContexts(c.Context)
	defer release()

	runMgr, err := run.NewManager(c.grpcClient, tokenResolver, c.HTTPClient, curSettings.CurrentProfile.Endpoint, c.Logger, c.UI)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create run manager")
		return 1
	}

	// Create non-speculative destroy run
	createdRun, err := runMgr.CreateRun(ctx, &run.CreateRunInput{
		WorkspaceID:              trn.TypeWorkspace.Normalize(c.arguments[0]),
		DirectoryPath:            c.directoryPath,
		ModuleSource:             c.moduleSource,
		ModuleVersion:            c.moduleVersion,
		TerraformVersion:         c.terraformVersion,
		TfVarFiles:               c.tfVarFiles,
		EnvVarFiles:              c.envVarFiles,
		TfVariables:              c.tfVariables,
		EnvVariables:             c.envVariables,
		TargetAddresses:          c.targetAddresses,
		Annotations:              c.annotations,
		IsDestroy:                true,
		IsSpeculative:            false,
		AutoApply:                *c.autoApprove,
		Refresh:                  c.refresh,
		IncludeModulePrereleases: *c.includeModulePrereleases,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to destroy")
		return 1
	}

	return followRunAndReport(ctx, cancelCtx, c.UI, runMgr, createdRun, &run.FollowOptions{
		Input:       *c.input,
		WaitTimeout: *c.waitTimeout,
	})
}

func (*destroyCommand) Synopsis() string {
	return "Destroy workspace resources."
}

func (*destroyCommand) Description() string {
	return `
   Destroys all resources in a workspace. Creates a
   destroy plan, then applies after approval.

` + followHelp("destroy", true) + `
   Terraform variables may be passed in via supported
   options or from the environment with a 'TF_VAR_' prefix.
`
}

func (*destroyCommand) Usage() string {
	return "tharsis [global options] destroy [options] <workspace-id>"
}

func (*destroyCommand) Example() string {
	return `
tharsis destroy -directory-path "./terraform" trn:workspace:<workspace_path>
`
}

func (c *destroyCommand) Flags() *flag.Set {
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
		&c.autoApprove,
		"auto-approve",
		"Skip interactive approval of the plan.",
		flag.Default(false),
	)
	addFollowFlags(f, &c.input, &c.waitTimeout)
	f.BoolVar(
		&c.refresh,
		"refresh",
		"Whether to do the usual refresh step.",
		flag.Default(true),
	)
	f.BoolVar(
		&c.includeModulePrereleases,
		"include-module-prereleases",
		"When module-version is empty or a constraint range, allow prerelease module versions to be selected as latest.",
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
	f.StringSliceVar(
		&c.annotations,
		"annotation",
		`An annotation as a JSON object with a key, value, and optional link, e.g. {"key":"commit","value":"abc123","link":"https://..."}.`,
	)

	f.MutuallyExclusive("directory-path", "module-source")

	return f
}
