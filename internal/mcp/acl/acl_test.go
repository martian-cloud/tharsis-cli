package acl

import (
	"strings"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tharsis"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestParsePatterns(t *testing.T) {
	tests := []struct {
		name        string
		patternStr  string
		expected    []string
		expectError bool
	}{
		{
			name:       "valid single pattern",
			patternStr: "prod/*",
			expected:   []string{"prod/*"},
		},
		{
			name:       "valid multiple patterns",
			patternStr: "prod/*,staging",
			expected:   []string{"prod/*", "staging"},
		},
		{
			name:       "empty string",
			patternStr: "",
			expected:   nil,
		},
		{
			name:       "whitespace handling",
			patternStr: " prod/* , staging ",
			expected:   []string{"prod/*", "staging"},
		},
		{
			name:       "duplicate patterns removed",
			patternStr: "prod/*,staging,prod/*,staging",
			expected:   []string{"prod/*", "staging"},
		},
		{
			name:        "wildcard-only pattern rejected",
			patternStr:  "*",
			expectError: true,
		},
		{
			name:        "pattern starting with wildcard rejected",
			patternStr:  "*/workspace",
			expectError: true,
		},
		{
			name:        "leading slash rejected",
			patternStr:  "/prod",
			expectError: true,
		},
		{
			name:        "trailing slash rejected",
			patternStr:  "prod/",
			expectError: true,
		},
		{
			name:        "double slash rejected",
			patternStr:  "prod//workspace",
			expectError: true,
		},
		{
			name:       "double star pattern",
			patternStr: "prod/**",
			expected:   []string{"prod/**"},
		},
		{
			name:       "empty entries ignored",
			patternStr: "prod,,,staging",
			expected:   []string{"prod", "staging"},
		},
		{
			name:        "pattern exceeds max length",
			patternStr:  strings.Repeat("a", maxPatternLength+1),
			expectError: true,
		},
		{
			name:       "pattern at max length allowed",
			patternStr: strings.Repeat("a", maxPatternLength),
			expected:   []string{strings.Repeat("a", maxPatternLength)},
		},
		{
			name:       "uppercase pattern lowercased",
			patternStr: "PROD/*",
			expected:   []string{"prod/*"},
		},
		{
			name:       "mixed case pattern lowercased",
			patternStr: "Prod/Team-*",
			expected:   []string{"prod/team-*"},
		},
		{
			name:       "single segment pattern allowed",
			patternStr: "my-group",
			expected:   []string{"my-group"},
		},
		{
			name:       "workspace pattern with group allowed",
			patternStr: "prod/my-workspace",
			expected:   []string{"prod/my-workspace"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsePatterns(tt.patternStr)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestNewACLChecker(t *testing.T) {
	tests := []struct {
		name        string
		patterns    string
		expectError bool
	}{
		{
			name:     "valid patterns",
			patterns: "prod/*,staging",
		},
		{
			name:     "empty patterns",
			patterns: "",
		},
		{
			name:        "invalid pattern propagates error",
			patterns:    "/prod",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker, err := NewChecker(tt.patterns)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, checker)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, checker)
			}
		})
	}
}

func TestACLChecker_authorize(t *testing.T) {
	tests := []struct {
		name        string
		patterns    string
		identifier  string
		resType     trn.ResourceType
		setupMock   func(*testing.T, *tharsis.MockClient)
		expectError bool
		errorMsg    string
	}{
		{
			name:       "exact match",
			patterns:   "prod",
			identifier: "trn:group:prod",
			resType:    trn.ResourceTypeGroup,
			setupMock: func(t *testing.T, m *tharsis.MockClient) {
				mockGrp := tharsis.NewGroup(t)
				mockGrp.On("GetGroup", t.Context(), &sdktypes.GetGroupInput{ID: ptr.String("trn:group:prod")}).
					Return(&sdktypes.Group{FullPath: "prod"}, nil)
				m.On("Groups").Return(mockGrp)
			},
		},
		{
			name:       "exact match does not match prefix",
			patterns:   "prod",
			identifier: "trn:group:production",
			resType:    trn.ResourceTypeGroup,
			setupMock: func(t *testing.T, m *tharsis.MockClient) {
				mockGrp := tharsis.NewGroup(t)
				mockGrp.On("GetGroup", t.Context(), &sdktypes.GetGroupInput{ID: ptr.String("trn:group:production")}).
					Return(&sdktypes.Group{FullPath: "production"}, nil)
				m.On("Groups").Return(mockGrp)
			},
			expectError: true,
			errorMsg:    "access denied",
		},
		{
			name:       "wildcard match",
			patterns:   "prod/*",
			identifier: "trn:group:prod/app1",
			resType:    trn.ResourceTypeGroup,
			setupMock: func(t *testing.T, m *tharsis.MockClient) {
				mockGrp := tharsis.NewGroup(t)
				mockGrp.On("GetGroup", t.Context(), &sdktypes.GetGroupInput{ID: ptr.String("trn:group:prod/app1")}).
					Return(&sdktypes.Group{FullPath: "prod/app1"}, nil)
				m.On("Groups").Return(mockGrp)
			},
		},
		{
			name:       "suffix wildcard match",
			patterns:   "prod/team-*",
			identifier: "trn:group:prod/team-alpha",
			resType:    trn.ResourceTypeGroup,
			setupMock: func(t *testing.T, m *tharsis.MockClient) {
				mockGrp := tharsis.NewGroup(t)
				mockGrp.On("GetGroup", t.Context(), &sdktypes.GetGroupInput{ID: ptr.String("trn:group:prod/team-alpha")}).
					Return(&sdktypes.Group{FullPath: "prod/team-alpha"}, nil)
				m.On("Groups").Return(mockGrp)
			},
		},
		{
			name:       "suffix wildcard no match",
			patterns:   "prod/team-*",
			identifier: "trn:group:prod/other-alpha",
			resType:    trn.ResourceTypeGroup,
			setupMock: func(t *testing.T, m *tharsis.MockClient) {
				mockGrp := tharsis.NewGroup(t)
				mockGrp.On("GetGroup", t.Context(), &sdktypes.GetGroupInput{ID: ptr.String("trn:group:prod/other-alpha")}).
					Return(&sdktypes.Group{FullPath: "prod/other-alpha"}, nil)
				m.On("Groups").Return(mockGrp)
			},
			expectError: true,
			errorMsg:    "access denied",
		},
		{
			name:       "no match",
			patterns:   "prod/*",
			identifier: "trn:group:staging/app1",
			resType:    trn.ResourceTypeGroup,
			setupMock: func(t *testing.T, m *tharsis.MockClient) {
				mockGrp := tharsis.NewGroup(t)
				mockGrp.On("GetGroup", t.Context(), &sdktypes.GetGroupInput{ID: ptr.String("trn:group:staging/app1")}).
					Return(&sdktypes.Group{FullPath: "staging/app1"}, nil)
				m.On("Groups").Return(mockGrp)
			},
			expectError: true,
			errorMsg:    "access denied",
		},
		{
			name:       "workspace match",
			patterns:   "prod/*",
			identifier: "ws-123",
			resType:    trn.ResourceTypeWorkspace,
			setupMock: func(t *testing.T, m *tharsis.MockClient) {
				mockWS := tharsis.NewWorkspaces(t)
				mockWS.On("GetWorkspace", t.Context(), &sdktypes.GetWorkspaceInput{ID: ptr.String("ws-123")}).
					Return(&sdktypes.Workspace{FullPath: "prod/app1"}, nil)
				m.On("Workspaces").Return(mockWS)
			},
		},
		{
			name:       "empty checker allows all",
			patterns:   "",
			identifier: "trn:group:any/path",
			resType:    trn.ResourceTypeGroup,
		},
		{
			name:       "run resolves to workspace path",
			patterns:   "prod/*",
			identifier: "run-123",
			resType:    trn.ResourceTypeRun,
			setupMock: func(t *testing.T, m *tharsis.MockClient) {
				mockRun := tharsis.NewRun(t)
				mockRun.On("GetRun", t.Context(), &sdktypes.GetRunInput{ID: "run-123"}).
					Return(&sdktypes.Run{
						Metadata: sdktypes.ResourceMetadata{TRN: "trn:run:prod/app1/run-123"},
					}, nil)
				m.On("Runs").Return(mockRun)
			},
		},
		{
			name:       "configuration version resolves to workspace path",
			patterns:   "staging/*",
			identifier: "cv-789",
			resType:    trn.ResourceTypeConfigurationVersion,
			setupMock: func(t *testing.T, m *tharsis.MockClient) {
				mockCV := tharsis.NewConfigurationVersion(t)
				mockCV.On("GetConfigurationVersion", t.Context(), &sdktypes.GetConfigurationVersionInput{ID: "cv-789"}).
					Return(&sdktypes.ConfigurationVersion{
						Metadata: sdktypes.ResourceMetadata{TRN: "trn:configuration_version:staging/app2/cv-789"},
					}, nil)
				m.On("ConfigurationVersions").Return(mockCV)
			},
		},
		{
			name:       "terraform module resolves to group path",
			patterns:   "prod/*",
			identifier: "module-123",
			resType:    trn.ResourceTypeTerraformModule,
			setupMock: func(t *testing.T, m *tharsis.MockClient) {
				mockModule := tharsis.NewTerraformModule(t)
				mockModule.On("GetModule", t.Context(), &sdktypes.GetTerraformModuleInput{ID: ptr.String("module-123")}).
					Return(&sdktypes.TerraformModule{GroupPath: "prod/team"}, nil)
				m.On("TerraformModules").Return(mockModule)
			},
		},
		{
			name:       "terraform module version resolves group path from TRN",
			patterns:   "prod/*",
			identifier: "version-123",
			resType:    trn.ResourceTypeTerraformModuleVersion,
			setupMock: func(t *testing.T, m *tharsis.MockClient) {
				mockModuleVersion := tharsis.NewTerraformModuleVersion(t)
				mockModuleVersion.On("GetModuleVersion", t.Context(), &sdktypes.GetTerraformModuleVersionInput{ID: ptr.String("version-123")}).
					Return(&sdktypes.TerraformModuleVersion{
						Metadata: sdktypes.ResourceMetadata{TRN: "trn:terraform_module_version:prod/team/my-module/aws/1.0.0"},
						ModuleID: "module-456",
					}, nil)
				m.On("TerraformModuleVersions").Return(mockModuleVersion)
			},
		},
		{
			name:       "case insensitive matching",
			patterns:   "PROD/*",
			identifier: "ws-case",
			resType:    trn.ResourceTypeWorkspace,
			setupMock: func(t *testing.T, m *tharsis.MockClient) {
				mockWS := tharsis.NewWorkspaces(t)
				mockWS.On("GetWorkspace", t.Context(), &sdktypes.GetWorkspaceInput{ID: ptr.String("ws-case")}).
					Return(&sdktypes.Workspace{FullPath: "Prod/Workspace"}, nil)
				m.On("Workspaces").Return(mockWS)
			},
		},
		{
			name:       "multiple patterns - second matches",
			patterns:   "prod/*,staging/*",
			identifier: "ws-staging",
			resType:    trn.ResourceTypeWorkspace,
			setupMock: func(t *testing.T, m *tharsis.MockClient) {
				mockWS := tharsis.NewWorkspaces(t)
				mockWS.On("GetWorkspace", t.Context(), &sdktypes.GetWorkspaceInput{ID: ptr.String("ws-staging")}).
					Return(&sdktypes.Workspace{FullPath: "staging/workspace"}, nil)
				m.On("Workspaces").Return(mockWS)
			},
		},
		{
			name:       "wildcard matches deeply nested path",
			patterns:   "prod/*",
			identifier: "ws-deep",
			resType:    trn.ResourceTypeWorkspace,
			setupMock: func(t *testing.T, m *tharsis.MockClient) {
				mockWS := tharsis.NewWorkspaces(t)
				mockWS.On("GetWorkspace", t.Context(), &sdktypes.GetWorkspaceInput{ID: ptr.String("ws-deep")}).
					Return(&sdktypes.Workspace{FullPath: "prod/team/subteam/workspace"}, nil)
				m.On("Workspaces").Return(mockWS)
			},
		},
		{
			name:       "API error propagates",
			patterns:   "prod/*",
			identifier: "ws-error",
			resType:    trn.ResourceTypeWorkspace,
			setupMock: func(t *testing.T, m *tharsis.MockClient) {
				mockWS := tharsis.NewWorkspaces(t)
				mockWS.On("GetWorkspace", t.Context(), &sdktypes.GetWorkspaceInput{ID: ptr.String("ws-error")}).
					Return(nil, assert.AnError)
				m.On("Workspaces").Return(mockWS)
			},
			expectError: true,
			errorMsg:    "failed to resolve workspace path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			if tt.setupMock != nil {
				tt.setupMock(t, mockClient)
			}

			checker, err := NewChecker(tt.patterns)
			require.NoError(t, err)
			require.NotNil(t, checker)

			err = checker.Authorize(t.Context(), mockClient, tt.identifier, tt.resType)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestACLChecker_caching(t *testing.T) {
	mockClient := tharsis.NewMockClient(t)
	mockWS := tharsis.NewWorkspaces(t)

	// Should only be called once due to caching
	mockWS.On("GetWorkspace", t.Context(), &sdktypes.GetWorkspaceInput{ID: ptr.String("ws-123")}).
		Return(&sdktypes.Workspace{FullPath: "prod/app"}, nil).Once()
	mockClient.On("Workspaces").Return(mockWS)

	checker, err := NewChecker("prod/*")
	require.NoError(t, err)

	// Call twice with same identifier
	err = checker.Authorize(t.Context(), mockClient, "ws-123", trn.ResourceTypeWorkspace)
	assert.NoError(t, err)

	err = checker.Authorize(t.Context(), mockClient, "ws-123", trn.ResourceTypeWorkspace)
	assert.NoError(t, err)

	mockWS.AssertNumberOfCalls(t, "GetWorkspace", 1)
}
