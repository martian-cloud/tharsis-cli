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
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/varparser"
	runpkg "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/pkg/run"
	"google.golang.org/protobuf/encoding/protojson"
)

// createRunTimeout bounds the create run request, which isn't canceled by the caller's context once
// it has been sent.
const createRunTimeout = 2 * time.Minute

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
	Annotations      []string
	IsDestroy        bool
	IsSpeculative    bool
	// AutoApply, when true, has the server start the apply phase as soon as the plan (and any
	// post-plan policy stage) clears, instead of waiting at planned for approval.
	AutoApply   bool
	Refresh     *bool
	RefreshOnly bool
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

// parseAnnotations converts the raw --annotation JSON strings into proto run annotations. It uses
// protojson rather than encoding/json so the input is validated against the proto schema: unknown
// fields and wrong-case keys are rejected locally with a clear message instead of being silently
// dropped and reaching the server as empty annotations.
func parseAnnotations(raw []string) ([]*pb.RunAnnotation, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	annotations := make([]*pb.RunAnnotation, len(raw))
	for i, s := range raw {
		if strings.TrimSpace(s) == "null" {
			return nil, fmt.Errorf("failed to parse annotation %q: must be a JSON object with a key and value", s)
		}
		var a pb.RunAnnotation
		if err := protojson.Unmarshal([]byte(s), &a); err != nil {
			return nil, fmt.Errorf("failed to parse annotation %q: %w", s, err)
		}
		// protojson leaves omitted fields as empty strings, so a missing or blank key or value
		// unmarshals cleanly and would only be rejected by the server. Catch it here instead.
		if err := ValidateRunAnnotation(&a); err != nil {
			return nil, fmt.Errorf("failed to parse annotation %q: %w", s, err)
		}
		annotations[i] = &a
	}
	return annotations, nil
}

// ValidateRunAnnotation checks the required fields of a single run annotation. Both the CLI flag
// path (parseAnnotations) and the MCP create_run tool call this so the two entry points reject the
// same malformed annotations (e.g. an empty key or value) before reaching the server.
func ValidateRunAnnotation(a *pb.RunAnnotation) error {
	if a.Key == "" {
		return fmt.Errorf("annotation key cannot be empty")
	}
	if a.Value == "" {
		return fmt.Errorf("annotation value cannot be empty")
	}
	return nil
}

// CreateRun uploads the configuration (when a directory is used) and creates the run. It returns
// the created run without waiting on it; use FollowRun to follow it through its stages.
//
// Canceling ctx stops the upload, but not a create request that has already been sent.
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

	// Parse run annotations
	annotations, err := parseAnnotations(input.Annotations)
	if err != nil {
		return nil, err
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
		Annotations:              annotations,
		Refresh:                  input.Refresh,
		RefreshOnly:              input.RefreshOnly,
		Speculative:              &input.IsSpeculative,
		TerraformVersion:         input.TerraformVersion,
		IncludeModulePrereleases: &input.IncludeModulePrereleases,
		AutoApply:                &input.AutoApply,
	}

	m.logger.Debug("create run input", "input", createRunInput)

	if err = ctx.Err(); err != nil {
		return nil, fmt.Errorf("failed to create run: %w", err)
	}

	// Once the request is sent, see it through even if ctx is canceled (by Ctrl-C) meanwhile: a
	// canceled request may still have created the run, which the caller then couldn't follow or
	// cancel. The caller sees ctx is done when it follows the returned run.
	createCtx, cancelCreate := context.WithTimeout(context.WithoutCancel(ctx), createRunTimeout)
	defer cancelCreate()
	createdRun, err := m.grpcClient.RunsClient.CreateRun(createCtx, createRunInput)
	if err != nil {
		return nil, fmt.Errorf("failed to create run: %w", err)
	}

	m.logger.Debug("created run", "run_id", createdRun.Metadata.Id)

	return createdRun, nil
}

// Overridable for tests; zero uses the pkg/run defaults.
var (
	pollInterval   time.Duration
	readRetryDelay time.Duration
)

// FollowOptions controls how the CLI follows a run.
type FollowOptions struct {
	// Input is false when the user passed -input=false; the CLI then never prompts. A
	// non-interactive terminal is treated the same way for policy gates.
	Input bool
	// Wait is true when the user passed -wait: when a decision can't be asked for, keep following
	// the run while it is made elsewhere (the Tharsis UI). Otherwise following stops and the run is
	// left waiting.
	Wait bool
	// WaitTimeout is how long to wait on a decision, whether asked for at the terminal or (with
	// Wait) being made in the Tharsis UI, before leaving the run waiting and exiting. Zero uses
	// runpkg.DefaultWaitTimeout; there is no way to wait forever.
	WaitTimeout time.Duration
}

// FollowRun follows a created run through each of its stages at the terminal, asking for decisions
// when a stage blocks on a decision. It wires the CLI's hooks (rendering, prompting and interrupt
// handling) into the generic follower in pkg/run.
//
// ctx is canceled by the first interrupt, with a SignalError cause naming the signal, and cleanupCtx
// by the second. At an interactive terminal the first Ctrl-C asks whether to cancel a run that is
// still doing work; the second abandons that question. Any other interrupt leaves the run going.
func (m *Manager) FollowRun(ctx, cleanupCtx context.Context, createdRun *pb.Run, opts *FollowOptions) (*runpkg.Outcome, error) {
	prompter := NewPrompter(m.ui, opts.Input)
	return runpkg.Follow(ctx, m.grpcClient, createdRun, &runpkg.Options{
		Observer:       &renderer{ui: m.ui, runID: createdRun.Metadata.Id},
		Decider:        &cliDecider{ui: m.ui, prompter: prompter, wait: opts.Wait},
		Interrupter:    &cliInterrupter{ui: m.ui, prompter: prompter, followCtx: ctx},
		CleanupContext: cleanupCtx,
		WaitTimeout:    opts.WaitTimeout,
		PollInterval:   pollInterval,
		RetryDelay:     readRetryDelay,
		Logger:         m.logger,
	})
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
