package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/acl"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tools/mocks"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type variableMocks struct {
	variables  *mocks.NamespaceVariablesClient
	workspaces *mocks.WorkspacesClient
	groups     *mocks.GroupsClient
	acl        *acl.MockChecker
}

func TestSetVariable(t *testing.T) {
	type testCase struct {
		name        string
		input       *setVariableInput
		mockSetup   func(*variableMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name: "create new terraform variable",
			input: &setVariableInput{
				NamespaceID: "ws1",
				Key:         "region",
				Value:       "us-east-1",
				Category:    "TERRAFORM",
			},
			mockSetup: func(m *variableMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "ws1"}).Return(&pb.Workspace{
					Metadata: &pb.ResourceMetadata{Id: "ws1"},
					FullPath: "group/workspace",
				}, nil)
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(nil)
				m.variables.On("GetNamespaceVariableByID", mock.Anything, &pb.GetNamespaceVariableByIDRequest{
					Id: "trn:variable:group/workspace/terraform/region",
				}).Return(nil, status.Error(codes.NotFound, "not found"))
				m.variables.On("CreateNamespaceVariable", mock.Anything, &pb.CreateNamespaceVariableRequest{
					Key:           "region",
					Value:         "us-east-1",
					Category:      pb.VariableCategory_TERRAFORM,
					NamespacePath: "group/workspace",
				}).Return(&pb.NamespaceVariable{
					Metadata: &pb.ResourceMetadata{Id: "v1"},
					Key:      "region",
					Value:    ptr.String("us-east-1"),
				}, nil)
			},
		},
		{
			name: "acl denial",
			input: &setVariableInput{
				NamespaceID: "ws1",
				Key:         "region",
				Value:       "us-east-1",
				Category:    "TERRAFORM",
			},
			mockSetup: func(m *variableMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "ws1"}).Return(&pb.Workspace{
					Metadata: &pb.ResourceMetadata{Id: "ws1"},
					FullPath: "group/workspace",
				}, nil)
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(status.Error(codes.PermissionDenied, "access denied"))
			},
			expectError: true,
		},
		{
			name: "update existing environment variable",
			input: &setVariableInput{
				NamespaceID: "ws1",
				Key:         "PATH",
				Value:       "/usr/bin",
				Category:    "ENVIRONMENT",
			},
			mockSetup: func(m *variableMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "ws1"}).Return(&pb.Workspace{
					Metadata: &pb.ResourceMetadata{Id: "ws1"},
					FullPath: "group/workspace",
				}, nil)
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(nil)
				m.variables.On("GetNamespaceVariableByID", mock.Anything, &pb.GetNamespaceVariableByIDRequest{
					Id: "trn:variable:group/workspace/environment/PATH",
				}).Return(&pb.NamespaceVariable{
					Metadata: &pb.ResourceMetadata{Id: "v1"},
					Key:      "PATH",
					Value:    ptr.String("/bin"),
				}, nil)
				m.variables.On("UpdateNamespaceVariable", mock.Anything, &pb.UpdateNamespaceVariableRequest{
					Id:    "v1",
					Key:   "PATH",
					Value: "/usr/bin",
				}).Return(&pb.NamespaceVariable{
					Metadata: &pb.ResourceMetadata{Id: "v1"},
					Key:      "PATH",
					Value:    ptr.String("/usr/bin"),
				}, nil)
			},
		},
		{
			name: "invalid category",
			input: &setVariableInput{
				NamespaceID: "ws1",
				Key:         "key",
				Value:       "value",
				Category:    "INVALID",
			},
			expectError: true,
		},
		{
			name: "workspace not found, uses group instead",
			input: &setVariableInput{
				NamespaceID: "group1",
				Key:         "region",
				Value:       "us-west-2",
				Category:    "TERRAFORM",
			},
			mockSetup: func(m *variableMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "group1"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
				m.groups.On("GetGroupByID", mock.Anything, &pb.GetGroupByIDRequest{Id: "group1"}).
					Return(&pb.Group{Metadata: &pb.ResourceMetadata{Id: "group1"}, FullPath: "group1"}, nil)
				m.acl.On("Authorize", mock.Anything, mock.Anything, "group1", trn.ResourceTypeGroup).Return(nil)
				m.variables.On("GetNamespaceVariableByID", mock.Anything, &pb.GetNamespaceVariableByIDRequest{
					Id: "trn:variable:group1/terraform/region",
				}).Return(nil, status.Error(codes.NotFound, "not found"))
				m.variables.On("CreateNamespaceVariable", mock.Anything, &pb.CreateNamespaceVariableRequest{
					Key:           "region",
					Value:         "us-west-2",
					Category:      pb.VariableCategory_TERRAFORM,
					NamespacePath: "group1",
				}).Return(&pb.NamespaceVariable{
					Metadata: &pb.ResourceMetadata{Id: "v2"},
					Key:      "region",
					Value:    ptr.String("us-west-2"),
				}, nil)
			},
		},
		{
			name: "namespace not found",
			input: &setVariableInput{
				NamespaceID: "nonexistent",
				Key:         "key",
				Value:       "value",
				Category:    "TERRAFORM",
			},
			mockSetup: func(m *variableMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "nonexistent"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
				m.groups.On("GetGroupByID", mock.Anything, &pb.GetGroupByIDRequest{Id: "nonexistent"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &variableMocks{
				variables:  mocks.NewNamespaceVariablesClient(t),
				groups:     mocks.NewGroupsClient(t),
				workspaces: mocks.NewWorkspacesClient(t),
				acl:        acl.NewMockChecker(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					NamespaceVariablesClient: testMocks.variables,
					WorkspacesClient:         testMocks.workspaces,
					GroupsClient:             testMocks.groups,
				},
				acl: testMocks.acl,
			}

			_, handler := setVariable(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, output.Success)
		})
	}
}

func TestDeleteVariable(t *testing.T) {
	type testCase struct {
		name        string
		input       *deleteVariableInput
		mockSetup   func(*variableMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name: "delete variable successfully",
			input: &deleteVariableInput{
				NamespaceID: "ws1",
				Key:         "region",
				Category:    "TERRAFORM",
			},
			mockSetup: func(m *variableMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "ws1"}).Return(&pb.Workspace{
					Metadata: &pb.ResourceMetadata{Id: "ws1"},
					FullPath: "group/workspace",
				}, nil)
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(nil)
				m.variables.On("DeleteNamespaceVariable", mock.Anything, &pb.DeleteNamespaceVariableRequest{
					Id: "trn:variable:group/workspace/terraform/region",
				}).Return(nil, nil)
			},
		},
		{
			name: "acl denial",
			input: &deleteVariableInput{
				NamespaceID: "ws1",
				Key:         "region",
				Category:    "TERRAFORM",
			},
			mockSetup: func(m *variableMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "ws1"}).Return(&pb.Workspace{
					Metadata: &pb.ResourceMetadata{Id: "ws1"},
					FullPath: "group/workspace",
				}, nil)
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(status.Error(codes.PermissionDenied, "access denied"))
			},
			expectError: true,
		},
		{
			name: "invalid category",
			input: &deleteVariableInput{
				NamespaceID: "ws1",
				Key:         "key",
				Category:    "INVALID",
			},
			expectError: true,
		},
		{
			name: "variable not found",
			input: &deleteVariableInput{
				NamespaceID: "ws1",
				Key:         "nonexistent",
				Category:    "TERRAFORM",
			},
			mockSetup: func(m *variableMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "ws1"}).Return(&pb.Workspace{
					Metadata: &pb.ResourceMetadata{Id: "ws1"},
					FullPath: "group/workspace",
				}, nil)
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(nil)
				m.variables.On("DeleteNamespaceVariable", mock.Anything, &pb.DeleteNamespaceVariableRequest{
					Id: "trn:variable:group/workspace/terraform/nonexistent",
				}).Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
		{
			name: "workspace not found, uses group instead",
			input: &deleteVariableInput{
				NamespaceID: "group1",
				Key:         "region",
				Category:    "TERRAFORM",
			},
			mockSetup: func(m *variableMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "group1"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
				m.groups.On("GetGroupByID", mock.Anything, &pb.GetGroupByIDRequest{Id: "group1"}).
					Return(&pb.Group{Metadata: &pb.ResourceMetadata{Id: "group1"}, FullPath: "group1"}, nil)
				m.acl.On("Authorize", mock.Anything, mock.Anything, "group1", trn.ResourceTypeGroup).Return(nil)
				m.variables.On("DeleteNamespaceVariable", mock.Anything, &pb.DeleteNamespaceVariableRequest{
					Id: "trn:variable:group1/terraform/region",
				}).Return(nil, nil)
			},
		},
		{
			name: "namespace not found",
			input: &deleteVariableInput{
				NamespaceID: "nonexistent",
				Key:         "key",
				Category:    "TERRAFORM",
			},
			mockSetup: func(m *variableMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "nonexistent"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
				m.groups.On("GetGroupByID", mock.Anything, &pb.GetGroupByIDRequest{Id: "nonexistent"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &variableMocks{
				variables:  mocks.NewNamespaceVariablesClient(t),
				workspaces: mocks.NewWorkspacesClient(t),
				groups:     mocks.NewGroupsClient(t),
				acl:        acl.NewMockChecker(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					NamespaceVariablesClient: testMocks.variables,
					WorkspacesClient:         testMocks.workspaces,
					GroupsClient:             testMocks.groups,
				},
				acl: testMocks.acl,
			}

			_, handler := deleteVariable(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, output.Success)
		})
	}
}

func TestSetTerraformVariablesFromFile(t *testing.T) {
	type testCase struct {
		name        string
		setupFile   func() string
		input       func(string) *setTerraformVariablesFromFileInput
		mockSetup   func(*variableMocks)
		expectError bool
		expectCount int
	}

	testCases := []testCase{
		{
			name: "set variables from file successfully",
			setupFile: func() string {
				file := filepath.Join(t.TempDir(), "terraform.tfvars")
				_ = os.WriteFile(file, []byte(`region = "us-east-1"
instance_type = "t2.micro"`), 0600)
				return file
			},
			input: func(file string) *setTerraformVariablesFromFileInput {
				return &setTerraformVariablesFromFileInput{
					NamespaceID: "ws1",
					FilePath:    file,
				}
			},
			mockSetup: func(m *variableMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "ws1"}).Return(&pb.Workspace{
					Metadata: &pb.ResourceMetadata{Id: "ws1"},
					FullPath: "group/workspace",
				}, nil)
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(nil)
				m.variables.On("SetNamespaceVariables", mock.Anything, mock.MatchedBy(func(req *pb.SetNamespaceVariablesRequest) bool {
					return req.NamespacePath == "group/workspace" && req.Category == pb.VariableCategory_TERRAFORM && len(req.Variables) == 2
				})).Return(nil, nil)
			},
			expectCount: 2,
		},
		{
			name: "acl denial",
			setupFile: func() string {
				file := filepath.Join(t.TempDir(), "terraform.tfvars")
				_ = os.WriteFile(file, []byte(`region = "us-east-1"`), 0600)
				return file
			},
			input: func(file string) *setTerraformVariablesFromFileInput {
				return &setTerraformVariablesFromFileInput{
					NamespaceID: "ws1",
					FilePath:    file,
				}
			},
			mockSetup: func(m *variableMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "ws1"}).Return(&pb.Workspace{
					Metadata: &pb.ResourceMetadata{Id: "ws1"},
					FullPath: "group/workspace",
				}, nil)
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(status.Error(codes.PermissionDenied, "access denied"))
			},
			expectError: true,
		},
		{
			name: "file does not exist",
			setupFile: func() string {
				return "/nonexistent/file.tfvars"
			},
			input: func(file string) *setTerraformVariablesFromFileInput {
				return &setTerraformVariablesFromFileInput{
					NamespaceID: "ws1",
					FilePath:    file,
				}
			},
			mockSetup: func(m *variableMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "ws1"}).Return(&pb.Workspace{
					Metadata: &pb.ResourceMetadata{Id: "ws1"},
					FullPath: "group/workspace",
				}, nil)
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(nil)
			},
			expectError: true,
		},
		{
			name: "workspace not found, uses group instead",
			setupFile: func() string {
				file := filepath.Join(t.TempDir(), "terraform.tfvars")
				_ = os.WriteFile(file, []byte(`region = "us-west-2"`), 0600)
				return file
			},
			input: func(file string) *setTerraformVariablesFromFileInput {
				return &setTerraformVariablesFromFileInput{
					NamespaceID: "group1",
					FilePath:    file,
				}
			},
			mockSetup: func(m *variableMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "group1"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
				m.groups.On("GetGroupByID", mock.Anything, &pb.GetGroupByIDRequest{Id: "group1"}).
					Return(&pb.Group{Metadata: &pb.ResourceMetadata{Id: "group1"}, FullPath: "group1"}, nil)
				m.acl.On("Authorize", mock.Anything, mock.Anything, "group1", trn.ResourceTypeGroup).Return(nil)
				m.variables.On("SetNamespaceVariables", mock.Anything, mock.MatchedBy(func(req *pb.SetNamespaceVariablesRequest) bool {
					return req.NamespacePath == "group1" && req.Category == pb.VariableCategory_TERRAFORM && len(req.Variables) == 1
				})).Return(nil, nil)
			},
			expectCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file := tc.setupFile()
			input := tc.input(file)

			testMocks := &variableMocks{
				variables:  mocks.NewNamespaceVariablesClient(t),
				workspaces: mocks.NewWorkspacesClient(t),
				groups:     mocks.NewGroupsClient(t),
				acl:        acl.NewMockChecker(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					NamespaceVariablesClient: testMocks.variables,
					WorkspacesClient:         testMocks.workspaces,
					GroupsClient:             testMocks.groups,
				},
				acl: testMocks.acl,
			}

			_, handler := setTerraformVariablesFromFile(toolCtx)
			_, output, err := handler(t.Context(), nil, input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, output.Success)
			assert.Equal(t, tc.expectCount, output.Count)
		})
	}
}

func TestSetEnvironmentVariablesFromFile(t *testing.T) {
	type testCase struct {
		name        string
		setupFile   func() string
		input       func(string) *setEnvironmentVariablesFromFileInput
		mockSetup   func(*variableMocks)
		expectError bool
		expectCount int
	}

	testCases := []testCase{
		{
			name: "set environment variables from file successfully",
			setupFile: func() string {
				file := filepath.Join(t.TempDir(), ".env")
				_ = os.WriteFile(file, []byte(`PATH=/usr/bin
HOME=/home/user`), 0600)
				return file
			},
			input: func(file string) *setEnvironmentVariablesFromFileInput {
				return &setEnvironmentVariablesFromFileInput{
					NamespaceID: "ws1",
					FilePath:    file,
				}
			},
			mockSetup: func(m *variableMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "ws1"}).Return(&pb.Workspace{
					Metadata: &pb.ResourceMetadata{Id: "ws1"},
					FullPath: "group/workspace",
				}, nil)
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(nil)
				m.variables.On("SetNamespaceVariables", mock.Anything, mock.MatchedBy(func(req *pb.SetNamespaceVariablesRequest) bool {
					return req.NamespacePath == "group/workspace" && req.Category == pb.VariableCategory_ENVIRONMENT && len(req.Variables) == 2
				})).Return(nil, nil)
			},
			expectCount: 2,
		},
		{
			name: "acl denial",
			setupFile: func() string {
				file := filepath.Join(t.TempDir(), ".env")
				_ = os.WriteFile(file, []byte(`PATH=/usr/bin`), 0600)
				return file
			},
			input: func(file string) *setEnvironmentVariablesFromFileInput {
				return &setEnvironmentVariablesFromFileInput{
					NamespaceID: "ws1",
					FilePath:    file,
				}
			},
			mockSetup: func(m *variableMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "ws1"}).Return(&pb.Workspace{
					Metadata: &pb.ResourceMetadata{Id: "ws1"},
					FullPath: "group/workspace",
				}, nil)
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(status.Error(codes.PermissionDenied, "access denied"))
			},
			expectError: true,
		},
		{
			name: "workspace not found, uses group instead",
			setupFile: func() string {
				file := filepath.Join(t.TempDir(), ".env")
				_ = os.WriteFile(file, []byte(`PATH=/usr/local/bin`), 0600)
				return file
			},
			input: func(file string) *setEnvironmentVariablesFromFileInput {
				return &setEnvironmentVariablesFromFileInput{
					NamespaceID: "group1",
					FilePath:    file,
				}
			},
			mockSetup: func(m *variableMocks) {
				m.workspaces.On("GetWorkspaceByID", mock.Anything, &pb.GetWorkspaceByIDRequest{Id: "group1"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
				m.groups.On("GetGroupByID", mock.Anything, &pb.GetGroupByIDRequest{Id: "group1"}).
					Return(&pb.Group{Metadata: &pb.ResourceMetadata{Id: "group1"}, FullPath: "group1"}, nil)
				m.acl.On("Authorize", mock.Anything, mock.Anything, "group1", trn.ResourceTypeGroup).Return(nil)
				m.variables.On("SetNamespaceVariables", mock.Anything, mock.MatchedBy(func(req *pb.SetNamespaceVariablesRequest) bool {
					return req.NamespacePath == "group1" && req.Category == pb.VariableCategory_ENVIRONMENT && len(req.Variables) == 1
				})).Return(nil, nil)
			},
			expectCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file := tc.setupFile()
			input := tc.input(file)

			testMocks := &variableMocks{
				variables:  mocks.NewNamespaceVariablesClient(t),
				workspaces: mocks.NewWorkspacesClient(t),
				groups:     mocks.NewGroupsClient(t),
				acl:        acl.NewMockChecker(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					NamespaceVariablesClient: testMocks.variables,
					WorkspacesClient:         testMocks.workspaces,
					GroupsClient:             testMocks.groups,
				},
				acl: testMocks.acl,
			}

			_, handler := setEnvironmentVariablesFromFile(toolCtx)
			_, output, err := handler(t.Context(), nil, input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, output.Success)
			assert.Equal(t, tc.expectCount, output.Count)
		})
	}
}
