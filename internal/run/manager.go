// Package run provides high-level run management operations for creating and applying Terraform runs.
package run

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/hashicorp/go-hclog"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/varparser"
)

// runStatus represents the overall status of a run.
type runStatus string

// runStatus constants.
const (
	runApplied            runStatus = "applied"
	runApplyQueued        runStatus = "apply_queued"
	runApplying           runStatus = "applying"
	runCanceled           runStatus = "canceled"
	runDiscarded          runStatus = "discarded"
	runErrored            runStatus = "errored"
	runPending            runStatus = "pending"
	runPlanQueued         runStatus = "plan_queued"
	runPlanned            runStatus = "planned"
	runPlannedAndFinished runStatus = "planned_and_finished"
	runPlanning           runStatus = "planning"
	runQueuing            runStatus = "queuing"
	runQueuingApply       runStatus = "queuing_apply"
)

// planStatus represents the status of a plan resource.
type planStatus string

// planStatus constants.
const (
	planCreated  planStatus = "created"
	planCanceled planStatus = "canceled"
	planQueued   planStatus = "queued"
	planErrored  planStatus = "errored"
	planFinished planStatus = "finished"
	planPending  planStatus = "pending"
	planRunning  planStatus = "running"
)

// applyStatus represents the status of an apply resource.
type applyStatus string

// applyStatus constants.
const (
	applyCanceled applyStatus = "canceled"
	applyCreated  applyStatus = "created"
	applyErrored  applyStatus = "errored"
	applyFinished applyStatus = "finished"
	applyPending  applyStatus = "pending"
	applyQueued   applyStatus = "queued"
	applyRunning  applyStatus = "running"
)

// Manager provides high-level run management operations
type Manager struct {
	grpcClient *client.GRPCClient
	tfeClient  client.RESTClient
	logger     hclog.Logger
	ui         terminal.UI
}

// CreateRunInput contains all parameters for creating a run
type CreateRunInput struct {
	WorkspaceID      string
	DirectoryPath    *string
	ModuleSource     *string
	ModuleVersion    *string
	TerraformVersion *string
	TfVarFiles       []string
	EnvVarFiles      []string
	TfVariables      []string
	EnvVariables     []string
	TargetAddresses  []string
	IsDestroy        bool
	IsSpeculative    bool
	Refresh          bool
	RefreshOnly      bool
	// IncludeModulePrereleases, when true and ModuleVersion is unset or a constraint
	// range, allows prerelease module versions to be selected as "latest".
	IncludeModulePrereleases bool
}

// NewManager creates a new run manager
func NewManager(
	grpcClient *client.GRPCClient,
	tokenResolver client.TokenResolver,
	httpClient *http.Client,
	endpoint string,
	logger hclog.Logger,
	ui terminal.UI,
) (*Manager, error) {
	tfeClient, err := client.NewRESTClient(&client.RESTClientConfig{
		Endpoint:      endpoint,
		TokenResolver: tokenResolver,
		HTTPClient:    httpClient,
	})
	if err != nil {
		return nil, err
	}

	return &Manager{
		grpcClient: grpcClient,
		tfeClient:  tfeClient,
		logger:     logger,
		ui:         ui,
	}, nil
}

// CreateRun creates and executes a run
func (m *Manager) CreateRun(ctx context.Context, input *CreateRunInput) (*pb.Run, error) {
	if input.DirectoryPath != nil && input.ModuleSource != nil {
		return nil, fmt.Errorf("must not supply both directory-path and module-source")
	}

	if input.ModuleSource == nil && input.ModuleVersion != nil {
		return nil, fmt.Errorf("must specify module-source if specifying module-version")
	}

	workspace, err := m.grpcClient.WorkspacesClient.GetWorkspaceByID(ctx, &pb.GetWorkspaceByIDRequest{Id: input.WorkspaceID})
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Set default directory path
	var directoryPath string
	if input.DirectoryPath != nil {
		directoryPath = *input.DirectoryPath
	} else {
		var err error
		directoryPath, err = os.Getwd()
		if err != nil {
			directoryPath = "."
		}
	}

	// Parse variables
	runVariables, err := m.parseVariables(directoryPath, input)
	if err != nil {
		return nil, err
	}

	// Handle configuration version upload if needed
	var configVersionID *string
	if input.ModuleSource == nil {
		if err = processDirectoryPath(directoryPath, input.IsDestroy); err != nil {
			return nil, fmt.Errorf("failed to process directory path: %w", err)
		}

		id, cErr := m.uploadConfigVersion(ctx, workspace.Metadata.Id, directoryPath, input.IsSpeculative)
		if cErr != nil {
			return nil, fmt.Errorf("failed to upload configuration version: %w", cErr)
		}

		configVersionID = &id
	}

	// Create run
	createRunInput := &pb.CreateRunRequest{
		WorkspaceId:              workspace.Metadata.Id,
		ConfigurationVersionId:   configVersionID,
		IsDestroy:                input.IsDestroy,
		ModuleSource:             input.ModuleSource,
		ModuleVersion:            input.ModuleVersion,
		Variables:                runVariables,
		TargetAddresses:          input.TargetAddresses,
		Refresh:                  input.Refresh,
		RefreshOnly:              input.RefreshOnly,
		Speculative:              &input.IsSpeculative,
		TerraformVersion:         input.TerraformVersion,
		IncludeModulePrereleases: &input.IncludeModulePrereleases,
	}

	m.logger.Debug("create run input", "input", createRunInput)

	createdRun, err := m.grpcClient.RunsClient.CreateRun(ctx, createRunInput)
	if err != nil {
		return nil, fmt.Errorf("failed to create run: %w", err)
	}

	m.ui.Output("Waiting on run to start")

	m.logger.Debug("created run", "run_id", createdRun.Metadata.Id)

	// Wait until the plan job has been created before requesting it, to avoid a race
	// where the job does not yet exist. The plan status is the authoritative signal.
	if err = m.waitForRunJob(ctx, createdRun.WorkspaceId, createdRun.Metadata.Id, func(ctx context.Context) (string, error) {
		plan, pErr := m.grpcClient.RunsClient.GetPlanByID(ctx, &pb.GetPlanByIDRequest{Id: createdRun.PlanId})
		if pErr != nil {
			return "", pErr
		}
		return plan.Status, nil
	}, planJobReady); err != nil {
		return nil, fmt.Errorf("failed waiting for plan job: %w", err)
	}

	// Get the job for the plan
	job, err := m.grpcClient.JobsClient.GetLatestJobForPlan(ctx, &pb.GetLatestJobForPlanRequest{
		PlanId: createdRun.PlanId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get job for plan: %w", err)
	}

	// Stream plan logs
	if err = m.streamJobLogs(ctx, job.Metadata.Id); err != nil {
		return nil, fmt.Errorf("failed to stream plan logs: %w", err)
	}

	// Get final run status
	finalRun, err := m.grpcClient.RunsClient.GetRunByID(ctx, &pb.GetRunByIDRequest{
		Id: createdRun.Metadata.Id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get final run: %w", err)
	}

	if finalRun.Status == string(runCanceled) || finalRun.Status == string(runDiscarded) || finalRun.Status == string(runErrored) {
		return nil, fmt.Errorf("run ended with status: %s", finalRun.Status)
	}

	return finalRun, nil
}

// ApplyRun applies a run
func (m *Manager) ApplyRun(ctx context.Context, runID string) (*pb.Run, error) {
	// Apply the run
	appliedRun, err := m.grpcClient.RunsClient.ApplyRun(ctx, &pb.ApplyRunRequest{
		RunId: runID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to apply run: %w", err)
	}

	// Wait until the apply job has been created before requesting it, to avoid a race
	// where the job does not yet exist. The apply status is the authoritative signal.
	if err = m.waitForRunJob(ctx, appliedRun.WorkspaceId, runID, func(ctx context.Context) (string, error) {
		apply, aErr := m.grpcClient.RunsClient.GetApplyByID(ctx, &pb.GetApplyByIDRequest{Id: appliedRun.ApplyId})
		if aErr != nil {
			return "", aErr
		}
		return apply.Status, nil
	}, applyJobReady); err != nil {
		return nil, fmt.Errorf("failed waiting for apply job: %w", err)
	}

	// Get the job for the apply
	job, err := m.grpcClient.JobsClient.GetLatestJobForApply(ctx, &pb.GetLatestJobForApplyRequest{
		ApplyId: appliedRun.ApplyId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get job for apply: %w", err)
	}

	// Stream apply logs
	if err = m.streamJobLogs(ctx, job.Metadata.Id); err != nil {
		return nil, fmt.Errorf("failed to stream apply logs: %w", err)
	}

	// Get final run status
	finalRun, err := m.grpcClient.RunsClient.GetRunByID(ctx, &pb.GetRunByIDRequest{
		Id: runID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get final run: %w", err)
	}

	if finalRun.Status != string(runApplied) {
		return nil, fmt.Errorf("apply ended with status: %s", finalRun.Status)
	}

	return finalRun, nil
}

// waitForRunJob blocks until ready reports the plan/apply job has been created. It
// checks the current status first, then subscribes to run events as a wake-up signal,
// re-checking the authoritative plan/apply status (getStatus) on each event. It returns
// an error if the plan/apply reaches a final state before a job becomes available, or
// if the run event stream cannot be established or closes first.
func (m *Manager) waitForRunJob(
	ctx context.Context,
	workspaceID, runID string,
	getStatus func(context.Context) (string, error),
	ready func(status string) (bool, error),
) error {
	// Check current state first; the subscription does not replay current state, so a
	// transition that already happened could otherwise be missed.
	status, err := getStatus(ctx)
	if err != nil {
		return err
	}
	if done, rErr := ready(status); rErr != nil || done {
		return rErr
	}

	// Subscribe to run events for wake-ups. The server keeps the subscription open
	// until the RPC is canceled, so use a child context that is canceled on return to
	// tear the stream down once the job is available.
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	stream, err := m.grpcClient.RunsClient.SubscribeToRunEvents(subCtx, &pb.SubscribeToRunEventsRequest{
		WorkspaceId: &workspaceID,
		RunId:       &runID,
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to run events: %w", err)
	}

	for {
		// Block until the next run event, then re-check the authoritative status.
		_, recvErr := stream.Recv()

		status, err := getStatus(ctx)
		if err != nil {
			return err
		}
		done, err := ready(status)
		if err != nil {
			return err
		}
		if done {
			return nil
		}

		// The stream closed before the job became available; the status above is the
		// most current we can get, so report the stream failure.
		if recvErr != nil {
			return fmt.Errorf("run event stream closed before a job was available: %w", recvErr)
		}
	}
}

// planJobReady reports whether the plan's job has been created, based on the plan
// status. A job exists once the plan reaches queued and through its terminal states.
// A plan that reaches a final state without a job (canceled) is an error.
func planJobReady(status string) (bool, error) {
	switch planStatus(status) {
	case planQueued, planRunning, planFinished, planErrored:
		// A job exists.
		return true, nil
	case planCanceled:
		return false, fmt.Errorf("plan reached final state before a job was available; status: %s", status)
	default:
		// created, pending: job not created yet.
		return false, nil
	}
}

// applyJobReady reports whether the apply's job has been created, based on the apply
// status. A job exists once the apply reaches queued and through its terminal states.
// An apply that reaches a final state without a job (canceled) is an error.
func applyJobReady(status string) (bool, error) {
	switch applyStatus(status) {
	case applyQueued, applyRunning, applyFinished, applyErrored:
		// A job exists.
		return true, nil
	case applyCanceled:
		return false, fmt.Errorf("apply reached final state before a job was available; status: %s", status)
	default:
		// created, pending: job not created yet.
		return false, nil
	}
}

func (m *Manager) streamJobLogs(ctx context.Context, jobID string) error {
	stream, err := m.grpcClient.JobsClient.SubscribeToJobLogStream(ctx, &pb.SubscribeToJobLogStreamRequest{
		JobId: jobID,
	})
	if err != nil {
		return err
	}

	for {
		event, err := stream.Recv()
		if err != nil {
			return err
		}

		if event.Data != nil {
			logs := strings.TrimSpace(event.Data.Logs)
			if color.NoColor {
				logs = terminal.StripAnsi(logs)
			}
			m.ui.Output(logs)
		}

		if event.Completed {
			break
		}
	}

	return nil
}

func (m *Manager) parseVariables(directoryPath string, input *CreateRunInput) ([]*pb.RunVariableInput, error) {
	// We want terraform variables processed automatically from the environment.
	parser := varparser.NewVariableParser(&directoryPath, true)

	tfVars, err := parser.ParseTerraformVariables(&varparser.ParseTerraformVariablesInput{
		TfVariables:    input.TfVariables,
		TfVarFilePaths: input.TfVarFiles,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse terraform variables: %w", err)
	}

	envVars, err := parser.ParseEnvironmentVariables(&varparser.ParseEnvironmentVariablesInput{
		EnvVariables:    input.EnvVariables,
		EnvVarFilePaths: input.EnvVarFiles,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	// Combine variables
	allVars := append(tfVars, envVars...)
	runVariables := make([]*pb.RunVariableInput, len(allVars))
	for i, v := range allVars {
		runVariables[i] = &pb.RunVariableInput{
			Key:      v.Key,
			Value:    &v.Value,
			Category: string(v.Category),
		}
	}

	return runVariables, nil
}

func (m *Manager) uploadConfigVersion(ctx context.Context, workspaceGID, directoryPath string, isSpeculative bool) (string, error) {
	createdConfigVersion, err := m.grpcClient.ConfigurationVersionsClient.CreateConfigurationVersion(ctx,
		&pb.CreateConfigurationVersionRequest{
			WorkspaceId: workspaceGID,
			Speculative: isSpeculative,
		})
	if err != nil {
		return "", err
	}

	m.logger.Debug("created configuration version", "id", createdConfigVersion.Metadata.Id)

	m.ui.Output("Uploading configuration version")

	// Upload the directory using GID
	if err = m.tfeClient.UploadConfigurationVersion(ctx, &client.UploadConfigurationVersionInput{
		WorkspaceID:     workspaceGID,
		ConfigVersionID: createdConfigVersion.Metadata.Id,
		DirectoryPath:   directoryPath,
	}); err != nil {
		return "", err
	}

	// Wait for upload to complete
	var configVersion *pb.ConfigurationVersion
	for {
		configVersion, err = m.grpcClient.ConfigurationVersionsClient.GetConfigurationVersionByID(ctx, &pb.GetConfigurationVersionByIDRequest{
			Id: createdConfigVersion.Metadata.Id,
		})
		if err != nil {
			return "", err
		}

		if configVersion.Status != "pending" {
			break
		}

		time.Sleep(time.Second)
	}

	if configVersion.Status != "uploaded" {
		return "", fmt.Errorf("upload failed; status is %s", configVersion.Status)
	}

	m.logger.Debug("uploaded configuration version successfully", "status", configVersion.Status)

	return configVersion.Metadata.Id, nil
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

	if !isDestroy {
		hasConfig, err := hasConfigFile(directoryPath)
		if err != nil {
			return err
		}

		if !hasConfig {
			return fmt.Errorf("directory tree has no .tf or .tf.json file and plan is not destroy mode: %s", directoryPath)
		}
	}

	return nil
}

func hasConfigFile(dirPath string) (bool, error) {
	found := false

	// Use filepath.WalkDir to scan the tree.
	err := filepath.WalkDir(dirPath, func(_ string, dirEntry fs.DirEntry, err error) error {
		// Pass through any error generated by WalkDir itself.
		if err != nil {
			return err
		}

		// Skip hidden directories for performance
		if dirEntry.IsDir() && strings.HasPrefix(dirEntry.Name(), ".") {
			return filepath.SkipDir
		}

		// We are interested in regular files and nothing else.
		if !dirEntry.IsDir() {
			name := dirEntry.Name()
			if strings.HasSuffix(name, ".tf") || strings.HasSuffix(name, ".tf.json") {
				found = true
				return filepath.SkipAll // Skip the rest.
			}
		}

		return nil
	})

	return found, err
}
