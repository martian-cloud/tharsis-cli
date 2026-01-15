package command

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/varparser"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

const (
	// Run status string prefix for a successful plan.
	// The other good value is "planned_and_finished".
	planSucceededRunPrefix = "planned"

	// Plan status string value for a successful plan.
	planSucceededPlanValue = "finished"
)

type runInput struct {
	moduleVersion    string
	directoryPath    string
	terraformVersion string
	workspacePath    string
	moduleSource     string
	envVarFilePath   []string
	tfVarFilePath    []string
	tfVariables      []string
	envVariables     []string
	targetAddresses  []string
	isDestroy        bool
	isSpeculative    bool
	refresh          bool
	refreshOnly      bool
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

	client, err := pc.meta.GetSDKClient()
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
	defs := pc.buildPlanDefs()

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
	moduleSource := getOption("module-source", "", cmdOpts)[0]
	moduleVersion := getOption("module-version", "", cmdOpts)[0]
	tfVariables := getOptionSlice("tf-var", cmdOpts)
	envVariables := getOptionSlice("env-var", cmdOpts)
	tfVarFiles := getOptionSlice("tf-var-file", cmdOpts)
	envVarFiles := getOptionSlice("env-var-file", cmdOpts)
	terraformVersion := getOption("terraform-version", "", cmdOpts)[0]
	destroy, err := getBoolOptionValue("destroy", "false", cmdOpts)
	if err != nil {
		pc.meta.UI.Error(output.FormatError("failed to parse boolean value for -destroy option", err))
		return 1
	}
	targetAddresses := getOptionSlice("target", cmdOpts)
	refresh, err := getBoolOptionValue("refresh", "true", cmdOpts)
	if err != nil {
		pc.meta.UI.Error(output.FormatError("failed to parse boolean value for -refresh option", err))
		return 1
	}
	refreshOnly, err := getBoolOptionValue("refresh-only", "false", cmdOpts)
	if err != nil {
		pc.meta.UI.Error(output.FormatError("failed to parse boolean value for -refresh-only option", err))
		return 1
	}

	// Extract path from TRN if needed, then validate path (error is already logged by validation function)
	actualPath := trn.ToPath(workspacePath)
	if !isNamespacePathValid(pc.meta, actualPath) {
		return 1
	}

	// Do all the inner work of the plan command.  Make it speculative.
	_, exitCode := createRun(ctx, client, pc.meta, &runInput{
		workspacePath:    workspacePath,
		directoryPath:    directoryPath,
		moduleSource:     moduleSource,
		moduleVersion:    moduleVersion,
		tfVarFilePath:    tfVarFiles,
		envVarFilePath:   envVarFiles,
		terraformVersion: terraformVersion,
		tfVariables:      tfVariables,
		envVariables:     envVariables,
		isDestroy:        destroy,
		isSpeculative:    true,
		targetAddresses:  targetAddresses,
		refresh:          refresh,
		refreshOnly:      refreshOnly,
	})

	// If there was an error, the error message has already been logged.
	return exitCode
}

// createRun does all the inner work of the plan, apply, and destroy commands.
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

	directoryPath := input.directoryPath

	// If directory path was not specified, default it to cwd or ".".
	if directoryPath == "" {
		var wErr error
		directoryPath, wErr = os.Getwd()
		if wErr != nil {
			directoryPath = "."
		}
	}

	processTFVariablesInput := &varparser.ParseTerraformVariablesInput{
		TfVariables:    input.tfVariables,
		TfVarFilePaths: input.tfVarFilePath,
	}

	meta.Logger.Debugf("plan: ParseTerraformVariables input: %#v", processTFVariablesInput)

	// We want terraform variables processed automatically from the environment.
	parser := varparser.NewVariableParser(&directoryPath, true)

	tfVars, err := parser.ParseTerraformVariables(processTFVariablesInput)
	if err != nil {
		meta.Logger.Error(output.FormatError("failed to parse terraform variables", err))
		return nil, 1
	}

	processEnvVariablesInput := &varparser.ParseEnvironmentVariablesInput{
		EnvVariables:    input.envVariables,
		EnvVarFilePaths: input.envVarFilePath,
	}

	meta.Logger.Debugf("plan: ParseEnvironmentVariables input: %#v", processEnvVariablesInput)

	envVars, err := parser.ParseEnvironmentVariables(processEnvVariablesInput)
	if err != nil {
		meta.Logger.Error(output.FormatError("failed to parse environment variables", err))
		return nil, 1
	}

	// Join both terraform and environment variables into a single slice.
	allVars := []varparser.Variable{}
	allVars = append(allVars, tfVars...)
	allVars = append(allVars, envVars...)

	runVariables := []sdktypes.RunVariable{}
	for _, v := range allVars {
		vCopy := v
		runVariables = append(runVariables, sdktypes.RunVariable{
			Key:      vCopy.Key,
			Value:    &vCopy.Value,
			Category: vCopy.Category,
		})
	}

	// Verify the workspace path exists.
	trnID := trn.ToTRN(input.workspacePath, trn.ResourceTypeWorkspace)
	getWorkspaceInput := &sdktypes.GetWorkspaceInput{ID: &trnID}
	foundWorkspace, err := client.Workspaces.GetWorkspace(ctx, getWorkspaceInput)
	if err != nil {
		meta.Logger.Error(output.FormatError("failed to get a workspace", err))
		return nil, 1
	}

	meta.Logger.Debugf("plan: found workspace: %#v", foundWorkspace)

	// If module source was not specified, check and maybe default the directory path.
	var createdConfigurationVersionID *string
	if input.moduleSource == "" {
		// Check, and process the directory path.
		pErr := processDirectoryPath(directoryPath, input.isDestroy)
		if pErr != nil {
			meta.Logger.Error(output.FormatError("failed to process directory path", pErr))
			return nil, 1
		}

		// Create and upload the configuration version.
		if directoryPath != "" {
			id, cErr := createUploadConfigVersion(ctx, client, meta,
				input.workspacePath, directoryPath, input.isSpeculative)
			if cErr != nil {
				meta.Logger.Error(output.FormatError("failed to upload configuration version", cErr))
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

	// Extract path from TRN if needed - WorkspacePath field expects paths, not TRNs
	actualWorkspacePath := trn.ToPath(input.workspacePath)

	createRunInput := &sdktypes.CreateRunInput{
		WorkspacePath:          actualWorkspacePath,
		ConfigurationVersionID: createdConfigurationVersionID,
		IsDestroy:              input.isDestroy,
		ModuleSource:           moduleSourceP,
		ModuleVersion:          moduleVersionP,
		Variables:              runVariables,
		TargetAddresses:        input.targetAddresses,
		Refresh:                input.refresh,
		RefreshOnly:            input.refreshOnly,
		Speculative:            &input.isSpeculative,
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

	lastSeenLogSize := int32(0)
	logsInput := &sdktypes.JobLogsSubscriptionInput{
		JobID:           *createdRun.Plan.CurrentJobID,
		RunID:           createdRun.Metadata.ID,
		WorkspacePath:   createdRun.WorkspacePath,
		LastSeenLogSize: &lastSeenLogSize,
	}

	meta.Logger.Debugf("plan: job logs input: %#v", logsInput)

	// Subscribe to job log events so we know when to fetch new logs.
	logChannel, err := client.Job.SubscribeToJobLogs(ctx, logsInput)
	if err != nil {
		meta.Logger.Error(output.FormatError("failed to get job logs", err))
		return nil, 1
	}

	for {
		logEvent, ok := <-logChannel
		if !ok {
			// No more logs since channel was closed.
			break
		}

		if logEvent.Error != nil {
			// Catch any incoming errors.
			meta.Logger.Error(output.FormatError("failed to get job logs", logEvent.Error))
			return nil, 1
		}

		meta.UI.Output(strings.TrimSpace(logEvent.Logs))
	}

	plannedRun, err := client.Run.GetRun(ctx, &sdktypes.GetRunInput{ID: createdRun.Metadata.ID})
	if err != nil {
		meta.Logger.Error(output.FormatError("failed to get planned run", err))
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
	workspacePath, directoryPath string, isSpeculative bool,
) (string, error) {
	// Call CreateConfigurationVersion
	// Extract path from TRN if needed - WorkspacePath field expects paths, not TRNs
	actualWorkspacePath := trn.ToPath(workspacePath)

	createdConfigurationVersion, err := client.ConfigurationVersion.CreateConfigurationVersion(ctx,
		&sdktypes.CreateConfigurationVersionInput{
			WorkspacePath: actualWorkspacePath,
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
			WorkspacePath:          actualWorkspacePath,
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
		time.Sleep(time.Second)
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
func combineErrors(errs []error) error {
	pool := []string{}

	// Convert each error to a string.
	for _, err := range errs {
		pool = append(pool, fmt.Sprintf("%s", err))
	}

	// Now, combine the strings and convert back to error.
	return errors.New(strings.Join(pool, "; "))
}

// buildCommonRunOptionDefs returns the common option definitions
// used by plan, apply and destroy commands.
func buildCommonRunOptionDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"directory-path": {
			Arguments: []string{"Directory_Path"},
			Synopsis:  "The path of the root module's directory.",
		},
		"module-source": {
			Arguments: []string{"Module_Source"},
			Synopsis:  "Remote module source specification.",
		},
		"module-version": {
			Arguments: []string{"Module_Version"},
			Synopsis:  "Remote module version number--defaults to latest.",
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
		"target": {
			Arguments: []string{"Resource_Address"},
			Synopsis:  "The Terraform address of the resources to be acted upon. (Use the option multiple times to specify multiple resources.)",
		},
		"refresh": {
			Arguments: []string{"true|false"},
			Synopsis:  "Whether to do the usual refresh step; default is true; use --refresh=false to disable the usual refresh step.",
		},
		"refresh-only": {
			Arguments: []string{"true|false"},
			Synopsis:  "Whether to do ONLY a refresh operation; default is false; use --refresh-only=true to do only a refresh step to update the TF state.",
		},
	}
}

func (pc planCommand) buildPlanDefs() optparser.OptionDefinitions {
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

   Terraform variables may be passed in via supported
   options or from the environment with a 'TF_VAR_'
   prefix.

   Variable parsing precedence:
     1. Terraform variables from the environment.
     2. terraform.tfvars file from module's directory,
        if present.
     3. terraform.tfvars.json file from module's
        directory, if present.
     4. *.auto.tfvars, *.auto.tfvars.json files
        from the module's directory, if present.
     5. --tf-var-file option(s).
     6. --tf-var option(s).

   NOTE: If the same variable is assigned multiple
   values, the last value found will be used. A
   --tf-var option will override the values from a
   *.tfvars file which will override values from
   the environment.

%s

`, pc.meta.BinaryName, buildHelpText(pc.buildPlanDefs()))
}
