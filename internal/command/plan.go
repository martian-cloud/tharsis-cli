package command

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/job"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/varparser"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

const (
	// logLimit sets the limit for log clump size for both plan and apply
	logLimit = 1024 * 1024

	// Run status string prefix for a successful plan.
	// The other good value is "planned_and_finished".
	planSucceededRunPrefix = "planned"

	// Plan status string value for a successful plan.
	planSucceededPlanValue = "finished"
)

type runInput struct {
	workspacePath    string
	directoryPath    string
	tfVarFilePath    string
	envVarFilePath   string
	moduleSource     string
	moduleVersion    string
	terraformVersion string
	tfVariables      []string
	envVariables     []string
	isDestroy        bool
	isSpeculative    bool
}

// planCommand is the top-level structure for the plan command.
type planCommand struct {
	meta *Metadata
}

// NewPlanCommandFactory returns a planCommand struct.
func NewPlanCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return planCommand{
			meta: meta,
		}, nil
	}
}

func (pc planCommand) Run(args []string) int {
	pc.meta.Logger.Debugf("Starting the 'plan' command with %d arguments:", len(args))
	for ix, arg := range args {
		pc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := pc.meta.ReadSettings()
	if err != nil {
		pc.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		pc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return pc.doPlan(ctx, client, args)
}

func (pc planCommand) doPlan(ctx context.Context, client *tharsis.Client, opts []string) int {
	pc.meta.Logger.Debugf("will do plan, %d opts", len(opts))

	// Build option definitions for this command.
	defs := buildPlanDefs()

	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(pc.meta.BinaryName+" plan", defs, opts)
	if err != nil {
		pc.meta.Logger.Error(output.FormatError("failed to parse plan argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		pc.meta.Logger.Error(output.FormatError("missing plan workspace path", nil), pc.HelpPlan())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive plan arguments: %s", cmdArgs)
		pc.meta.Logger.Error(output.FormatError(msg, nil), pc.HelpPlan())
		return 1
	}

	workspacePath := cmdArgs[0]
	directoryPath := getOption("directory-path", "", cmdOpts)[0]
	tfVariables := getOption("tf-var", "", cmdOpts)
	envVariables := getOption("env-var", "", cmdOpts)
	tfVarFile := getOption("tf-var-file", "", cmdOpts)[0]
	envVarFile := getOption("env-var-file", "", cmdOpts)[0]
	terraformVersion := getOption("terraform-version", "", cmdOpts)[0]
	destroy := getOption("destroy", "", cmdOpts)[0] == "1"

	// Error is already logged.
	if !isNamespacePathValid(pc.meta, workspacePath) {
		return 1
	}

	// Do all the inner work of the plan command.  Make it speculative.
	_, exitCode := createRun(ctx, client, pc.meta, &runInput{
		workspacePath:    workspacePath,
		directoryPath:    directoryPath,
		tfVarFilePath:    tfVarFile,
		envVarFilePath:   envVarFile,
		terraformVersion: terraformVersion,
		tfVariables:      tfVariables,
		envVariables:     envVariables,
		isDestroy:        destroy,
		isSpeculative:    true,
	})

	// If there was an error, the error message has already been logged.
	return exitCode
}

// innerPlan does all the inner work of the plan command.
// If there is a problem, log an error and return non-zero.
// The apply command can also use this function.
func createRun(ctx context.Context, client *tharsis.Client, meta *Metadata, input *runInput) (*sdktypes.Run, int) {
	// Must not have both directory path and module source.
	if (input.directoryPath != "") && (input.moduleSource != "") {
		meta.Logger.Error(output.FormatError("must not supply both -directory-path and -module-source", nil))
		return nil, 1
	}

	// If module source is missing, module version is not allowed.
	if (input.moduleSource == "") && (input.moduleVersion != "") {
		meta.Logger.Error(output.FormatError("must specify -module-source if specifying -module-version", nil))
		return nil, 1
	}

	// Must not have both a file path and variable flags.
	if (input.tfVarFilePath != "" || input.envVarFilePath != "") && (len(input.tfVariables) > 0 || len(input.envVariables) > 0) {
		meta.Logger.Error(output.FormatError("either (-tf-var-file / -env-var-file) or (-tf-var / -env-var) may be used", nil))
		return nil, 1
	}

	// Prepare input for variable parser.
	processVariablesInput := varparser.ProcessVariablesInput{
		TfVariables:    input.tfVariables,
		EnvVariables:   input.envVariables,
		TfVarFilePath:  input.tfVarFilePath,
		EnvVarFilePath: input.envVarFilePath,
	}

	// Process variables string or files.
	variables, err := varparser.ProcessVariables(processVariablesInput)
	if err != nil {
		meta.Logger.Error(output.FormatError("failed to process variables", err))
		return nil, 1
	}

	// Verify the workspace path exists.
	foundWorkspace, err := client.Workspaces.GetWorkspace(ctx,
		&sdktypes.GetWorkspaceInput{Path: input.workspacePath})
	if err != nil {
		meta.Logger.Error(output.FormatError("failed to get a workspace", err))
		return nil, 1
	}
	if foundWorkspace == nil {
		meta.Logger.Error(output.FormatError("failed to get a workspace", nil))
		return nil, 1
	}
	meta.Logger.Debugf("plan: found workspace: %#v", foundWorkspace)

	// If module source was not specified, check and maybe default the directory path.
	var createdConfigurationVersionID *string
	if input.moduleSource == "" {
		directoryPath := input.directoryPath

		// If directory path was not specified, default it to cwd or ".".
		if directoryPath == "" {
			var wErr error
			directoryPath, wErr = os.Getwd()
			if wErr != nil {
				directoryPath = "."
			}
		}

		// Check, and process the directory path.
		pErr := processDirectoryPath(directoryPath, input.isDestroy)
		if pErr != nil {
			meta.Logger.Error(output.FormatError(pErr.Error(), nil))
			return nil, 1
		}

		// Create and upload the configuration version.
		if directoryPath != "" {
			id, cErr := createUploadConfigVersion(ctx, client, meta,
				input.workspacePath, directoryPath, input.isSpeculative)
			if cErr != nil {
				meta.Logger.Error(output.FormatError(cErr.Error(), nil))
				return nil, 1
			}
			createdConfigurationVersionID = &id
		}
	}

	// convert string to *string for module source and version
	var moduleSourceP, moduleVersionP *string
	if input.moduleSource != "" {
		moduleSourceP = &input.moduleSource
	}
	if input.moduleVersion != "" {
		moduleVersionP = &input.moduleVersion
	}

	// Inform the user that we're making progress...
	meta.UI.Output("Waiting on run to start")

	createRunInput := &sdktypes.CreateRunInput{
		WorkspacePath:          input.workspacePath,
		ConfigurationVersionID: createdConfigurationVersionID,
		IsDestroy:              input.isDestroy,
		ModuleSource:           moduleSourceP,
		ModuleVersion:          moduleVersionP,
		Variables:              variables,
	}

	if input.terraformVersion != "" {
		createRunInput.TerraformVersion = &input.terraformVersion
	}

	// Call CreateRun
	createdRun, err := client.Run.CreateRun(ctx, createRunInput)
	if err != nil {
		meta.Logger.Error(output.FormatError("failed to create a run", err))
		return nil, 1
	}

	meta.Logger.Debugf("plan: createdRun: %#v", createdRun)

	// Display the logs
	// (We're guaranteed that Plan is not nil.)
	planLogChannel, err := client.Job.GetJobLogs(ctx, &sdktypes.GetJobLogsInput{
		ID:          *createdRun.Plan.CurrentJobID,
		StartOffset: 0,
		Limit:       logLimit,
	})
	if err != nil {
		meta.Logger.Error(output.FormatError("failed to connect to read plan logs", err))
		return nil, 1
	}
	err = job.DisplayLogs(planLogChannel, meta.UI)
	if err != nil {
		meta.Logger.Error(output.FormatError("failed to read plan logs", err))
		return nil, 1
	}

	// Check whether the plan passed (vs. failed).
	plannedRun, err := client.Run.GetRun(ctx, &sdktypes.GetRunInput{ID: createdRun.Metadata.ID})
	if err != nil {
		meta.Logger.Error(output.FormatError("Failed to get post-plan run", err))
		return nil, 1
	}

	// If a plan fails, both plannedRun.Status and plannedRun.Plan.Status are "errored".
	// Details of the failure should be visible above in the logs.
	//
	// If a plan succeeds, plannedRun.Status is either "planned" or "planned_and_finished",
	// while plannedRun.Plan.Status is "finished".
	//
	meta.Logger.Debugf("post-plan run status: %s", plannedRun.Status)
	meta.Logger.Debugf("post-plan run.plan.status: %s", plannedRun.Plan.Status)
	if !strings.HasPrefix(string(plannedRun.Status), planSucceededRunPrefix) {
		// Status is already printed in the jog logs, so no need to log it here.
		return nil, 1
	}
	if plannedRun.Plan.Status != planSucceededPlanValue {
		meta.Logger.Errorf("Plan status: %s", plannedRun.Status)
		return nil, 1
	}

	return plannedRun, 0
}

// processDirectoryPath checks and processes the directory path.
func processDirectoryPath(directoryPath string, isDestroy bool) error {

	// Make sure the directory path exists and is a directory--to give more precise messages.
	// By doing the check here, one check catches both plan and apply commands.
	dirStat, err := os.Stat(directoryPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory path does not exist: %s", directoryPath)
	}
	if err != nil {
		return fmt.Errorf("failed to stat directory path %s: %s", directoryPath, err)
	}
	if !dirStat.IsDir() {
		return fmt.Errorf("path is not a directory: %s", directoryPath)
	}

	// Do pre-plan checks.
	// For future checks, may need to pass in ctx, client, meta, foundWorkspace, and/or isSpeculative
	err = doPrePlanChecks(directoryPath, isDestroy)
	if err != nil {
		return fmt.Errorf("failed pre-plan check: %s", err)
	}

	return nil
}

// createUploadConfigVersion creates and uploads a configuration version,
// then waits for it to finish uploading.  It returns an error code, zero for success.
func createUploadConfigVersion(ctx context.Context, client *tharsis.Client, meta *Metadata,
	workspacePath, directoryPath string, isSpeculative bool) (string, error) {

	// Call CreateConfigurationVersion
	createdConfigurationVersion, err := client.ConfigurationVersion.CreateConfigurationVersion(ctx,
		&sdktypes.CreateConfigurationVersionInput{
			WorkspacePath: workspacePath,
			Speculative:   &isSpeculative,
		})
	if err != nil {
		return "", fmt.Errorf("failed to create a configuration version: %s", err)
	}
	meta.Logger.Debugf("plan: createdConfigurationVersion: %#v", createdConfigurationVersion)

	// Inform the user that we're making progress...
	meta.UI.Output("Uploading configuration version")

	// Call UploadConfigurationVersion
	err = client.ConfigurationVersion.UploadConfigurationVersion(ctx,
		&sdktypes.UploadConfigurationVersionInput{
			WorkspacePath:          workspacePath,
			ConfigurationVersionID: createdConfigurationVersion.Metadata.ID,
			DirectoryPath:          directoryPath,
		})
	if err != nil {
		return "", fmt.Errorf("failed to upload a configuration version: %s", err)
	}
	meta.Logger.Debugf("plan: upload configuration version was successfully launched.")

	// Wait for the upload to complete:
	var updatedConfigurationVersion *sdktypes.ConfigurationVersion
	for {
		updatedConfigurationVersion, err = client.ConfigurationVersion.GetConfigurationVersion(ctx,
			&sdktypes.GetConfigurationVersionInput{ID: createdConfigurationVersion.Metadata.ID})
		if err != nil {
			return "", fmt.Errorf("failed to check for completion of upload: %s", err)
		}
		if updatedConfigurationVersion.Status != "pending" {
			break
		}
	}
	if updatedConfigurationVersion.Status != "uploaded" {
		return "", fmt.Errorf("upload failed; status is %s", updatedConfigurationVersion.Status)
	}
	meta.Logger.Debugf("plan: upload configuration version successfully finished.")

	return createdConfigurationVersion.Metadata.ID, nil
}

// Return a slice of errors--or nil if all checks pass.
func doPrePlanChecks(directoryPath string, isDestroy bool) error {

	errors := []error{}

	// TODO: Add the initial authorization check:
	//	   1. if !tfe.Workspace.Permissions.CanQueueRun:
	//	   insufficient rights to generate a plan
	//	   See https://github.com/hashicorp/terraform/blob/main/internal/backend/remote/backend_plan.go#L27

	// There are also other checks that may be added in the future.

	//	   6. if !backendOperation.HasConfig() && backendOperation.PlanMode != plans.DestroyMode:
	//	   a non-destroy plan requires at least one .tf or .tf.json file in the tree
	if !isDestroy {
		hasConfig, err := hasConfigFile(directoryPath)
		if err != nil {
			errors = append(errors, err)
		}
		if !hasConfig {
			errors = append(errors,
				fmt.Errorf("directory tree has no .tf or .tf.json file and plan is not destroy mode: %s",
					directoryPath))
		}
	}

	// If no errors, return nil.
	// A quick out could be done for exactly one error, but the benefit would be negligible.
	if len(errors) > 0 {
		return combineErrors(errors)
	}

	return nil
}

// Check whether a directory path has at least one file whose name ends in ".tf" or ".tf.json".
// Return true if at least one such file exists.
func hasConfigFile(dirPath string) (bool, error) {
	result := false // no config files found until proven otherwise

	// Use filepath.WalkDir to scan the tree.
	err := filepath.WalkDir(dirPath, func(path string, dirEntry fs.DirEntry, err error) error {

		// Pass through any error generated by WalkDir itself.
		if err != nil {
			return err
		}

		// We are interested in regular files and nothing else.
		if dirEntry.Type() == 0 {
			if strings.HasSuffix(path, ".tf") || strings.HasSuffix(path, ".tf.json") {
				// proven that there is a TF config file
				result = true
				// Unfortunately, there does not appear to be a practical way to tell WalkDir
				// that our work here is done and it can quit walking any more of the tree.
			}
		}

		return nil
	})

	return result, err
}

// Return a single error that contains potentially multiple errors.
func combineErrors(errors []error) error {
	pool := []string{}

	// Convert each error to a string.
	for _, err := range errors {
		pool = append(pool, fmt.Sprintf("%s", err))
	}

	// Now, combine the strings and convert back to error.
	return fmt.Errorf(strings.Join(pool, "; "))
}

// buildCommonRunOptionDefs returns the common option definitions
// used by plan, apply and destroy commands.
func buildCommonRunOptionDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"directory-path": {
			Arguments: []string{"Directory_Path"},
			Synopsis:  "The path of the root module's directory.",
		},
		"tf-var-file": {
			Arguments: []string{"Tf_Var_File"},
			Synopsis:  "The path to a .tfvars variables file.",
		},
		"env-var-file": {
			Arguments: []string{"Env_Var_File"},
			Synopsis:  "The path to an environment variables file.",
		},
		"tf-var": {
			Arguments: []string{"Tf_Var"},
			Synopsis:  "A terraform variable as a key=value pair.",
		},
		"env-var": {
			Arguments: []string{"Env_Var"},
			Synopsis:  "An environment variable as a key=value pair.",
		},
		"terraform-version": {
			Arguments: []string{"Terraform_Version"},
			Synopsis:  "The Terraform CLI version to use for the run.",
		},
	}
}

func buildPlanDefs() optparser.OptionDefinitions {
	defs := buildCommonRunOptionDefs()
	destroyDef := optparser.OptionDefinition{
		Arguments: []string{}, // zero arguments means it's a bool with no argument
		Synopsis:  "Designates this run as a destroy operation.",
	}
	defs["destroy"] = &destroyDef

	return defs
}

func (pc planCommand) Synopsis() string {
	return "Create a speculative plan"
}

func (pc planCommand) Help() string {
	return pc.HelpPlan()
}

// HelpPlan prints the help string for the 'run apply' command.
func (pc planCommand) HelpPlan() string {
	return fmt.Sprintf(`
Usage: %s [global options] plan [options] <workspace>

   The run plan command plans a run. It allows viewing
   the changes Terraform will make to your infrastructure
   without applying them. Supports setting run-scoped
   Terraform / environment variables and planning a
   destroy run.

%s


Combining --tf-var or --env-var and --tf-var-file or --env-var-file is not allowed.

`, pc.meta.BinaryName, buildHelpText(buildPlanDefs()))
}

// The End.
