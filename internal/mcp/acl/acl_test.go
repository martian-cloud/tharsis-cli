package acl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tools/mocks"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		setupMock   func(*client.Client)
		expectError bool
		errorMsg    string
	}{
		{
			name:       "exact match",
			patterns:   "prod",
			identifier: "trn:group:prod",
			resType:    trn.ResourceTypeGroup,
			setupMock: func(c *client.Client) {
				mockGrp := c.GroupsClient.(*mocks.GroupsClient)
				mockGrp.On("GetGroupByID", mock.Anything, &pb.GetGroupByIDRequest{Id: "trn:group:prod"}).
					Return(&pb.Group{FullPath: "prod"}, nil)
			},
		},
		{
			name:       "exact match does not match prefix",
			patterns:   "prod",
			identifier: "trn:group:production",
			resType:    trn.ResourceTypeGroup,
			setupMock: func(c *client.Client) {
				mockGrp := c.GroupsClient.(*mocks.GroupsClient)
				mockGrp.On("GetGroupByID", mock.Anything, &pb.GetGroupByIDRequest{Id: "trn:group:production"}).
					Return(&pb.Group{FullPath: "production"}, nil)
			},
			expectError: true,
			errorMsg:    "access denied",
		},
		{
			name:       "wildcard match",
			patterns:   "prod/*",
			identifier: "trn:group:prod/app1",
			resType:    trn.ResourceTypeGroup,
			setupMock: func(c *client.Client) {
				mockGrp := c.GroupsClient.(*mocks.GroupsClient)
				mockGrp.On("GetGroupByID", mock.Anything, &pb.GetGroupByIDRequest{Id: "trn:group:prod/app1"}).
					Return(&pb.Group{FullPath: "prod/app1"}, nil)
			},
		},
		{
			name:       "suffix wildcard match",
			patterns:   "prod/team-*",
			identifier: "trn:group:prod/team-alpha",
			resType:    trn.ResourceTypeGroup,
			setupMock: func(c *client.Client) {
				mockGrp := c.GroupsClient.(*mocks.GroupsClient)
				mockGrp.On("GetGroupByID", mock.Anything, &pb.GetGroupByIDRequest{Id: "trn:group:prod/team-alpha"}).
					Return(&pb.Group{FullPath: "prod/team-alpha"}, nil)
			},
		},
		{
			name:       "suffix wildcard no match",
			patterns:   "prod/team-*",
			identifier: "trn:group:prod/other-alpha",
			resType:    trn.ResourceTypeGroup,
			setupMock: func(c *client.Client) {
				mockGrp := c.GroupsClient.(*mocks.GroupsClient)
				mockGrp.On("GetGroupByID", mock.Anything, &pb.GetGroupByIDRequest{Id: "trn:group:prod/other-alpha"}).
					Return(&pb.Group{FullPath: "prod/other-alpha"}, nil)
			},
			expectError: true,
			errorMsg:    "access denied",
		},
		{
			name:       "no match",
			patterns:   "prod/*",
			identifier: "trn:group:staging/app1",
			resType:    trn.ResourceTypeGroup,
			setupMock: func(c *client.Client) {
				mockGrp := c.GroupsClient.(*mocks.GroupsClient)
				mockGrp.On("GetGroupByID", mock.Anything, &pb.GetGroupByIDRequest{Id: "trn:group:staging/app1"}).
					Return(&pb.Group{FullPath: "staging/app1"}, nil)
			},
			expectError: true,
			errorMsg:    "access denied",
		},
		{
			name:       "workspace match",
			patterns:   "prod/*",
			identifier: "ws-123",
			resType:    trn.ResourceTypeWorkspace,
			setupMock: func(c *client.Client) {
				mockWS := c.WorkspacesClient.(*mocks.WorkspacesClient)
				mockWS.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "ws-123"}).
					Return(&pb.Workspace{FullPath: "prod/app1"}, nil)
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
			setupMock: func(c *client.Client) {
				mockRun := c.RunsClient.(*mocks.RunsClient)
				mockRun.On("GetRunByID", mock.Anything, &pb.GetRunByIDRequest{Id: "run-123"}).
					Return(&pb.Run{
						Metadata: &pb.ResourceMetadata{Trn: "trn:run:prod/app1/run-123"},
					}, nil)
			},
		},
		{
			name:       "configuration version resolves to workspace path",
			patterns:   "staging/*",
			identifier: "cv-789",
			resType:    trn.ResourceTypeConfigurationVersion,
			setupMock: func(c *client.Client) {
				mockCV := c.ConfigurationVersionsClient.(*mocks.ConfigurationVersionsClient)
				mockCV.On("GetConfigurationVersionByID", mock.Anything, &pb.GetConfigurationVersionByIDRequest{Id: "cv-789"}).
					Return(&pb.ConfigurationVersion{
						Metadata: &pb.ResourceMetadata{Trn: "trn:configuration_version:staging/app2/cv-789"},
					}, nil)
			},
		},
		{
			name:       "terraform module resolves to group path",
			patterns:   "prod/*",
			identifier: "module-123",
			resType:    trn.ResourceTypeTerraformModule,
			setupMock: func(c *client.Client) {
				mockModule := c.TerraformModulesClient.(*mocks.TerraformModulesClient)
				mockModule.On("GetTerraformModuleByID", mock.Anything, &pb.GetTerraformModuleByIDRequest{Id: "module-123"}).
					Return(&pb.TerraformModule{
						Metadata: &pb.ResourceMetadata{Trn: "trn:terraform_module:prod/team/my-module/aws"},
					}, nil)
			},
		},
		{
			name:       "terraform module version resolves group path from TRN",
			patterns:   "prod/*",
			identifier: "version-123",
			resType:    trn.ResourceTypeTerraformModuleVersion,
			setupMock: func(c *client.Client) {
				mockModule := c.TerraformModulesClient.(*mocks.TerraformModulesClient)
				mockModule.On("GetTerraformModuleVersionByID", mock.Anything, &pb.GetTerraformModuleVersionByIDRequest{Id: "version-123"}).
					Return(&pb.TerraformModuleVersion{
						Metadata: &pb.ResourceMetadata{Trn: "trn:terraform_module_version:prod/team/my-module/aws/1.0.0"},
						ModuleId: "module-456",
					}, nil)
			},
		},
		{
			name:       "case insensitive matching",
			patterns:   "PROD/*",
			identifier: "ws-case",
			resType:    trn.ResourceTypeWorkspace,
			setupMock: func(c *client.Client) {
				mockWS := c.WorkspacesClient.(*mocks.WorkspacesClient)
				mockWS.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "ws-case"}).
					Return(&pb.Workspace{FullPath: "Prod/Workspace"}, nil)
			},
		},
		{
			name:       "multiple patterns - second matches",
			patterns:   "prod/*,staging/*",
			identifier: "ws-staging",
			resType:    trn.ResourceTypeWorkspace,
			setupMock: func(c *client.Client) {
				mockWS := c.WorkspacesClient.(*mocks.WorkspacesClient)
				mockWS.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "ws-staging"}).
					Return(&pb.Workspace{FullPath: "staging/workspace"}, nil)
			},
		},
		{
			name:       "wildcard matches deeply nested path",
			patterns:   "prod/*",
			identifier: "ws-deep",
			resType:    trn.ResourceTypeWorkspace,
			setupMock: func(c *client.Client) {
				mockWS := c.WorkspacesClient.(*mocks.WorkspacesClient)
				mockWS.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "ws-deep"}).
					Return(&pb.Workspace{FullPath: "prod/team/subteam/workspace"}, nil)
			},
		},
		{
			name:       "API error propagates",
			patterns:   "prod/*",
			identifier: "ws-error",
			resType:    trn.ResourceTypeWorkspace,
			setupMock: func(c *client.Client) {
				mockWS := c.WorkspacesClient.(*mocks.WorkspacesClient)
				mockWS.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "ws-error"}).
					Return(nil, status.Error(codes.Internal, "internal error"))
			},
			expectError: true,
			errorMsg:    "failed to resolve workspace path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &client.Client{
				GroupsClient:                mocks.NewGroupsClient(t),
				WorkspacesClient:            mocks.NewWorkspacesClient(t),
				RunsClient:                  mocks.NewRunsClient(t),
				ConfigurationVersionsClient: mocks.NewConfigurationVersionsClient(t),
				TerraformModulesClient:      mocks.NewTerraformModulesClient(t),
			}
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
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
	mockClient := &client.Client{
		WorkspacesClient: mocks.NewWorkspacesClient(t),
	}
	mockWS := mockClient.WorkspacesClient.(*mocks.WorkspacesClient)

	// Should only be called once due to caching
	mockWS.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "ws-123"}).
		Return(&pb.Workspace{FullPath: "prod/app"}, nil).Once()

	checker, err := NewChecker("prod/*")
	require.NoError(t, err)

	// Call twice with same identifier
	err = checker.Authorize(t.Context(), mockClient, "ws-123", trn.ResourceTypeWorkspace)
	assert.NoError(t, err)

	err = checker.Authorize(t.Context(), mockClient, "ws-123", trn.ResourceTypeWorkspace)
	assert.NoError(t, err)

	mockWS.AssertNumberOfCalls(t, "GetWorkspaceByID", 1)
}
