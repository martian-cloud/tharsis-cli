package tools

import (
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/acl"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tharsis"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestListRuns(t *testing.T) {
	type testCase struct {
		name     string
		input    listRunsInput
		runs     []sdktypes.Run
		pageInfo sdktypes.PageInfo
		validate func(*testing.T, listRunsOutput)
	}

	tests := []testCase{
		{
			name:  "list runs without filter",
			input: listRunsInput{},
			runs: []sdktypes.Run{
				{
					Metadata:         sdktypes.ResourceMetadata{ID: "run-1"},
					Status:           sdktypes.RunApplied,
					WorkspacePath:    "group/workspace",
					CreatedBy:        "user@example.com",
					TerraformVersion: "1.5.0",
				},
				{
					Metadata:         sdktypes.ResourceMetadata{ID: "run-2"},
					Status:           sdktypes.RunPending,
					WorkspacePath:    "group/workspace",
					CreatedBy:        "user@example.com",
					TerraformVersion: "1.5.0",
				},
			},
			pageInfo: sdktypes.PageInfo{
				HasNextPage: false,
				Cursor:      "",
			},
			validate: func(t *testing.T, output listRunsOutput) {
				assert.Len(t, output.Runs, 2)
				assert.Equal(t, "run-1", output.Runs[0].ID)
				assert.Equal(t, sdktypes.RunApplied, output.Runs[0].Status)
				assert.False(t, output.PageInfo.HasNextPage)
			},
		},
		{
			name: "list runs with workspace filter",
			input: listRunsInput{
				WorkspacePath: ptr.String("group/workspace"),
			},
			runs: []sdktypes.Run{
				{
					Metadata:         sdktypes.ResourceMetadata{ID: "run-1"},
					Status:           sdktypes.RunApplied,
					WorkspacePath:    "group/workspace",
					CreatedBy:        "user@example.com",
					TerraformVersion: "1.5.0",
				},
			},
			pageInfo: sdktypes.PageInfo{
				HasNextPage: false,
				Cursor:      "",
			},
			validate: func(t *testing.T, output listRunsOutput) {
				assert.Len(t, output.Runs, 1)
				assert.Equal(t, "group/workspace", output.Runs[0].WorkspacePath)
			},
		},
		{
			name: "list runs with pagination",
			input: listRunsInput{
				Limit:  ptr.Int32(10),
				Cursor: ptr.String("cursor-123"),
			},
			runs: []sdktypes.Run{
				{
					Metadata:         sdktypes.ResourceMetadata{ID: "run-3"},
					Status:           sdktypes.RunPending,
					WorkspacePath:    "group/workspace",
					CreatedBy:        "user@example.com",
					TerraformVersion: "1.5.0",
				},
			},
			pageInfo: sdktypes.PageInfo{
				HasNextPage: true,
				Cursor:      "cursor-456",
			},
			validate: func(t *testing.T, output listRunsOutput) {
				assert.Len(t, output.Runs, 1)
				assert.True(t, output.PageInfo.HasNextPage)
				assert.Equal(t, "cursor-456", output.PageInfo.Cursor)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockRun := tharsis.NewRun(t)
			mockACL := acl.NewMockChecker(t)

			mockClient.On("Runs").Return(mockRun)
			mockRun.On("GetRuns", mock.Anything, mock.Anything).Return(
				&sdktypes.GetRunsOutput{
					Runs:     tt.runs,
					PageInfo: &tt.pageInfo,
				}, nil)

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := listRuns(tc)
			_, output, err := handler(t.Context(), nil, tt.input)

			assert.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

func TestGetRun(t *testing.T) {
	runID := "test-run-id"

	tests := []struct {
		name        string
		run         *sdktypes.Run
		aclError    error
		expectError bool
		validate    func(*testing.T, getRunOutput)
	}{
		{
			name: "successful run retrieval",
			run: &sdktypes.Run{
				Metadata:         sdktypes.ResourceMetadata{ID: runID},
				Status:           sdktypes.RunApplied,
				WorkspacePath:    "group/workspace",
				CreatedBy:        "user@example.com",
				TerraformVersion: "1.5.0",
				IsDestroy:        false,
				Speculative:      false,
			},
			validate: func(t *testing.T, output getRunOutput) {
				assert.Equal(t, runID, output.Run.ID)
				assert.Equal(t, sdktypes.RunApplied, output.Run.Status)
				assert.Equal(t, "group/workspace", output.Run.WorkspacePath)
				assert.Equal(t, "user@example.com", output.Run.CreatedBy)
				assert.Equal(t, "1.5.0", output.Run.TerraformVersion)
				assert.False(t, output.Run.IsDestroy)
				assert.False(t, output.Run.Speculative)
			},
		},
		{
			name: "run with plan and apply errors",
			run: &sdktypes.Run{
				Metadata:         sdktypes.ResourceMetadata{ID: runID},
				Status:           sdktypes.RunErrored,
				WorkspacePath:    "group/workspace",
				CreatedBy:        "user@example.com",
				TerraformVersion: "1.5.0",
				Plan: &sdktypes.Plan{
					Status:       sdktypes.PlanErrored,
					ErrorMessage: ptr.String("plan failed"),
				},
				Apply: &sdktypes.Apply{
					Status:       sdktypes.ApplyErrored,
					ErrorMessage: ptr.String("apply failed"),
				},
			},
			validate: func(t *testing.T, output getRunOutput) {
				assert.Equal(t, sdktypes.RunErrored, output.Run.Status)
				assert.NotNil(t, output.Run.Plan)
				assert.Equal(t, sdktypes.PlanErrored, output.Run.Plan.Status)
				assert.Equal(t, "plan failed", *output.Run.Plan.ErrorMessage)
				assert.NotNil(t, output.Run.Apply)
				assert.Equal(t, sdktypes.ApplyErrored, output.Run.Apply.Status)
				assert.Equal(t, "apply failed", *output.Run.Apply.ErrorMessage)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockRun := tharsis.NewRun(t)
			mockACL := acl.NewMockChecker(t)

			if tt.aclError == nil {
				mockClient.On("Runs").Return(mockRun)
				mockRun.On("GetRun", mock.Anything, &sdktypes.GetRunInput{ID: runID}).Return(tt.run, nil)
			}

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := getRun(tc)
			_, output, err := handler(t.Context(), nil, getRunInput{ID: runID})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, output)
				}
			}
		})
	}
}

func TestCreateRun(t *testing.T) {
	workspacePath := "group/workspace"

	tests := []struct {
		name        string
		input       createRunInput
		run         *sdktypes.Run
		aclError    error
		expectError bool
		validate    func(*testing.T, createRunOutput)
	}{
		{
			name: "successful run creation",
			input: createRunInput{
				WorkspacePath: workspacePath,
			},
			run: &sdktypes.Run{
				Metadata:      sdktypes.ResourceMetadata{ID: "run-id"},
				Status:        sdktypes.RunPending,
				WorkspacePath: workspacePath,
			},
			validate: func(t *testing.T, output createRunOutput) {
				assert.Equal(t, "run-id", output.Run.ID)
				assert.Equal(t, sdktypes.RunPending, output.Run.Status)
				assert.Equal(t, workspacePath, output.Run.WorkspacePath)
			},
		},
		{
			name: "run with module source",
			input: createRunInput{
				WorkspacePath: workspacePath,
				ModuleSource:  ptr.String("registry.terraform.io/namespace/module"),
				ModuleVersion: ptr.String("1.0.0"),
			},
			run: &sdktypes.Run{
				Metadata:      sdktypes.ResourceMetadata{ID: "run-id"},
				Status:        sdktypes.RunPending,
				WorkspacePath: workspacePath,
			},
			validate: func(t *testing.T, output createRunOutput) {
				assert.Equal(t, "run-id", output.Run.ID)
			},
		},
		{
			name: "ACL denial",
			input: createRunInput{
				WorkspacePath: workspacePath,
			},
			aclError:    assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockRun := tharsis.NewRun(t)
			mockACL := acl.NewMockChecker(t)

			if tt.aclError == nil {
				mockClient.On("Runs").Return(mockRun)
				mockRun.On("CreateRun", mock.Anything, &sdktypes.CreateRunInput{
					WorkspacePath:          tt.input.WorkspacePath,
					ConfigurationVersionID: tt.input.ConfigurationVersionID,
					ModuleSource:           tt.input.ModuleSource,
					ModuleVersion:          tt.input.ModuleVersion,
					TerraformVersion:       tt.input.TerraformVersion,
					IsDestroy:              tt.input.IsDestroy,
					Speculative:            tt.input.Speculative,
					TargetAddresses:        tt.input.TargetAddresses,
				}).Return(tt.run, nil)
			}
			mockACL.On("Authorize", mock.Anything, mockClient, "trn:workspace:"+workspacePath, trn.ResourceTypeWorkspace).Return(tt.aclError)

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := createRun(tc)
			_, output, err := handler(t.Context(), nil, tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, output)
				}
			}
		})
	}
}

func TestApplyRun(t *testing.T) {
	runID := "test-run-id"

	tests := []struct {
		name        string
		run         *sdktypes.Run
		aclError    error
		expectError bool
		validate    func(*testing.T, applyRunOutput)
	}{
		{
			name: "successful run apply",
			run: &sdktypes.Run{
				Metadata: sdktypes.ResourceMetadata{ID: runID},
				Status:   sdktypes.RunApplyQueued,
			},
			validate: func(t *testing.T, output applyRunOutput) {
				assert.Equal(t, runID, output.RunID)
				assert.Equal(t, sdktypes.RunApplyQueued, output.Status)
			},
		},
		{
			name:        "ACL denial",
			aclError:    assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockRun := tharsis.NewRun(t)
			mockACL := acl.NewMockChecker(t)

			if tt.aclError == nil {
				mockClient.On("Runs").Return(mockRun)
				mockRun.On("ApplyRun", mock.Anything, &sdktypes.ApplyRunInput{RunID: runID}).Return(tt.run, nil)
			}
			mockACL.On("Authorize", mock.Anything, mockClient, runID, trn.ResourceTypeRun).Return(tt.aclError)

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := applyRun(tc)
			_, output, err := handler(t.Context(), nil, applyRunInput{ID: runID})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, output)
				}
			}
		})
	}
}

func TestCancelRun(t *testing.T) {
	runID := "test-run-id"

	tests := []struct {
		name        string
		input       cancelRunInput
		run         *sdktypes.Run
		aclError    error
		expectError bool
		validate    func(*testing.T, cancelRunOutput)
	}{
		{
			name: "successful run cancellation",
			input: cancelRunInput{
				ID: runID,
			},
			run: &sdktypes.Run{
				Metadata: sdktypes.ResourceMetadata{ID: runID},
				Status:   sdktypes.RunCanceled,
			},
			validate: func(t *testing.T, output cancelRunOutput) {
				assert.True(t, output.Success)
				assert.Contains(t, output.Message, runID)
			},
		},
		{
			name: "force cancel run",
			input: cancelRunInput{
				ID:    runID,
				Force: ptr.Bool(true),
			},
			run: &sdktypes.Run{
				Metadata: sdktypes.ResourceMetadata{ID: runID},
				Status:   sdktypes.RunCanceled,
			},
			validate: func(t *testing.T, output cancelRunOutput) {
				assert.True(t, output.Success)
			},
		},
		{
			name: "ACL denial",
			input: cancelRunInput{
				ID: runID,
			},
			aclError:    assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockRun := tharsis.NewRun(t)
			mockACL := acl.NewMockChecker(t)

			if tt.aclError == nil {
				mockClient.On("Runs").Return(mockRun)
				mockRun.On("CancelRun", mock.Anything, &sdktypes.CancelRunInput{
					RunID: tt.input.ID,
					Force: tt.input.Force,
				}).Return(tt.run, nil)
			}
			mockACL.On("Authorize", mock.Anything, mockClient, runID, trn.ResourceTypeRun).Return(tt.aclError)

			tc := &ToolContext{
				tharsisURL:  "https://test.tharsis.io",
				profileName: "test",
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := cancelRun(tc)
			_, output, err := handler(t.Context(), nil, tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, output)
				}
			}
		})
	}
}
