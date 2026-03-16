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

	"github.com/hashicorp/go-hclog"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tfe"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/varparser"
)

const (
	runStatusPlannedPrefix = "planned"
	runStatusApplied       = "applied"
)

// Manager provides high-level run management operations
type Manager struct {
	grpcClient *client.Client
	tfeClient  tfe.RESTClient
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
}

// NewManager creates a new run manager
func NewManager(
	grpcClient *client.Client,
	tokenGetter client.TokenGetter,
	httpClient *http.Client,
	endpoint string,
	logger hclog.Logger,
	ui terminal.UI,
) (*Manager, error) {
	tfeClient, err := tfe.NewRESTClient(endpoint, tokenGetter, httpClient)
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

	m.ui.Output("Waiting on run to start")

	// Create run
	createRunInput := &pb.CreateRunRequest{
		WorkspaceId:            workspace.Metadata.Id,
		ConfigurationVersionId: configVersionID,
		IsDestroy:              input.IsDestroy,
		ModuleSource:           input.ModuleSource,
		ModuleVersion:          input.ModuleVersion,
		Variables:              runVariables,
		TargetAddresses:        input.TargetAddresses,
		Refresh:                input.Refresh,
		RefreshOnly:            input.RefreshOnly,
		Speculative:            &input.IsSpeculative,
		TerraformVersion:       input.TerraformVersion,
	}

	m.logger.Debug("create run input", "input", createRunInput)

	createdRun, err := m.grpcClient.RunsClient.CreateRun(ctx, createRunInput)
	if err != nil {
		return nil, fmt.Errorf("failed to create run: %w", err)
	}

	m.logger.Debug("created run", "run_id", createdRun.Metadata.Id)

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

	if !strings.HasPrefix(finalRun.Status, runStatusPlannedPrefix) {
		return nil, fmt.Errorf("plan ended with status: %s", finalRun.Status)
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

	if finalRun.Status != runStatusApplied {
		return nil, fmt.Errorf("apply ended with status: %s", finalRun.Status)
	}

	return finalRun, nil
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
			m.ui.Output(strings.TrimSpace(event.Data.Logs))
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
	if err = m.tfeClient.UploadConfigurationVersion(ctx, &tfe.UploadConfigurationVersionInput{
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
