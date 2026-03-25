package tools

import (
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

type terraformModuleMocks struct {
	terraformModules *mocks.TerraformModulesClient
	groups           *mocks.GroupsClient
	acl              *acl.MockChecker
}

func TestListTerraformModules(t *testing.T) {
	type testCase struct {
		name          string
		input         *listTerraformModulesInput
		mockSetup     func(*terraformModuleMocks)
		expectError   bool
		expectResults int
	}

	testCases := []testCase{
		{
			name:  "list modules successfully",
			input: &listTerraformModulesInput{Limit: ptr.Int32(10)},
			mockSetup: func(m *terraformModuleMocks) {
				m.terraformModules.On("GetTerraformModules", mock.Anything, &pb.GetTerraformModulesRequest{
					PaginationOptions: &pb.PaginationOptions{First: ptr.Int32(10)},
				}).Return(&pb.GetTerraformModulesResponse{
					Modules: []*pb.TerraformModule{
						{Metadata: &pb.ResourceMetadata{Id: "m1", Trn: "trn:terraform_module:group/module/aws"}, Name: "module", System: "aws"},
					},
					PageInfo: &pb.PageInfo{HasNextPage: false},
				}, nil)
			},
			expectResults: 1,
		},
		{
			name:  "list call fails",
			input: &listTerraformModulesInput{Limit: ptr.Int32(10)},
			mockSetup: func(m *terraformModuleMocks) {
				m.terraformModules.On("GetTerraformModules", mock.Anything, &pb.GetTerraformModulesRequest{
					PaginationOptions: &pb.PaginationOptions{First: ptr.Int32(10)},
				}).Return(nil, status.Error(codes.Internal, "internal error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &terraformModuleMocks{
				terraformModules: mocks.NewTerraformModulesClient(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					TerraformModulesClient: testMocks.terraformModules,
				},
			}

			_, handler := listTerraformModules(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, output.Modules, tc.expectResults)
		})
	}
}

func TestGetTerraformModule(t *testing.T) {
	type testCase struct {
		name        string
		input       *getTerraformModuleInput
		mockSetup   func(*terraformModuleMocks)
		expectError bool
		expectID    string
	}

	testCases := []testCase{
		{
			name:  "get module successfully",
			input: &getTerraformModuleInput{ID: "m1"},
			mockSetup: func(m *terraformModuleMocks) {
				m.terraformModules.On("GetTerraformModuleByID", mock.Anything, &pb.GetTerraformModuleByIDRequest{Id: "m1"}).Return(&pb.TerraformModule{
					Metadata:      &pb.ResourceMetadata{Id: "m1", Trn: "trn:terraform_module:group/module/aws"},
					Name:          "module",
					System:        "aws",
					RepositoryUrl: "https://github.com/org/repo",
				}, nil)
			},
			expectID: "m1",
		},
		{
			name:  "module not found",
			input: &getTerraformModuleInput{ID: "nonexistent"},
			mockSetup: func(m *terraformModuleMocks) {
				m.terraformModules.On("GetTerraformModuleByID", mock.Anything, &pb.GetTerraformModuleByIDRequest{Id: "nonexistent"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &terraformModuleMocks{
				terraformModules: mocks.NewTerraformModulesClient(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					TerraformModulesClient: testMocks.terraformModules,
				},
			}

			_, handler := getTerraformModule(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectID, output.Module.ID)
		})
	}
}

func TestCreateTerraformModule(t *testing.T) {
	type testCase struct {
		name        string
		input       *createTerraformModuleInput
		mockSetup   func(*terraformModuleMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name: "create module successfully",
			input: &createTerraformModuleInput{
				GroupID:       "g1",
				Name:          "new-module",
				System:        "aws",
				RepositoryURL: "https://github.com/org/repo",
			},
			mockSetup: func(m *terraformModuleMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "g1", trn.ResourceTypeGroup).Return(nil)
				m.terraformModules.On("CreateTerraformModule", mock.Anything, &pb.CreateTerraformModuleRequest{
					GroupId:       "g1",
					Name:          "new-module",
					System:        "aws",
					RepositoryUrl: "https://github.com/org/repo",
				}).Return(&pb.TerraformModule{
					Metadata:      &pb.ResourceMetadata{Id: "m1", Trn: "trn:terraform_module:group1/new-module/aws"},
					Name:          "new-module",
					System:        "aws",
					RepositoryUrl: "https://github.com/org/repo",
				}, nil)
			},
		},
		{
			name: "acl denial",
			input: &createTerraformModuleInput{
				GroupID: "g1",
				Name:    "new-module",
				System:  "aws",
			},
			mockSetup: func(m *terraformModuleMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "g1", trn.ResourceTypeGroup).Return(status.Error(codes.PermissionDenied, "access denied"))
			},
			expectError: true,
		},
		{
			name: "create failed",
			input: &createTerraformModuleInput{
				GroupID:       "g1",
				Name:          "new-module",
				System:        "aws",
				RepositoryURL: "https://github.com/org/repo",
			},
			mockSetup: func(m *terraformModuleMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "g1", trn.ResourceTypeGroup).Return(nil)
				m.terraformModules.On("CreateTerraformModule", mock.Anything, &pb.CreateTerraformModuleRequest{
					GroupId:       "g1",
					Name:          "new-module",
					System:        "aws",
					RepositoryUrl: "https://github.com/org/repo",
				}).Return(nil, status.Error(codes.Aborted, "already exists"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &terraformModuleMocks{
				terraformModules: mocks.NewTerraformModulesClient(t),
				acl:              acl.NewMockChecker(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					TerraformModulesClient: testMocks.terraformModules,
				},
				acl: testMocks.acl,
			}

			_, handler := createTerraformModule(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output)
			require.NotNil(t, output.Module)
		})
	}
}

func TestUpdateTerraformModule(t *testing.T) {
	type testCase struct {
		name        string
		input       *updateTerraformModuleInput
		mockSetup   func(*terraformModuleMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name: "update module successfully",
			input: &updateTerraformModuleInput{
				ID:            "m1",
				RepositoryURL: ptr.String("https://github.com/org/new-repo"),
			},
			mockSetup: func(m *terraformModuleMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "m1", trn.ResourceTypeTerraformModule).Return(nil)
				m.terraformModules.On("UpdateTerraformModule", mock.Anything, &pb.UpdateTerraformModuleRequest{
					Id:            "m1",
					RepositoryUrl: ptr.String("https://github.com/org/new-repo"),
				}).Return(&pb.TerraformModule{
					Metadata:      &pb.ResourceMetadata{Id: "m1", Trn: "trn:terraform_module:group/module/aws"},
					Name:          "module",
					System:        "aws",
					RepositoryUrl: "https://github.com/org/new-repo",
				}, nil)
			},
		},
		{
			name: "acl denial",
			input: &updateTerraformModuleInput{
				ID:            "m1",
				RepositoryURL: ptr.String("https://github.com/org/repo"),
			},
			mockSetup: func(m *terraformModuleMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "m1", trn.ResourceTypeTerraformModule).Return(status.Error(codes.PermissionDenied, "access denied"))
			},
			expectError: true,
		},
		{
			name: "module not found",
			input: &updateTerraformModuleInput{
				ID:            "nonexistent",
				RepositoryURL: ptr.String("https://github.com/org/repo"),
			},
			mockSetup: func(m *terraformModuleMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "nonexistent", trn.ResourceTypeTerraformModule).Return(nil)
				m.terraformModules.On("UpdateTerraformModule", mock.Anything, &pb.UpdateTerraformModuleRequest{
					Id:            "nonexistent",
					RepositoryUrl: ptr.String("https://github.com/org/repo"),
				}).Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &terraformModuleMocks{
				terraformModules: mocks.NewTerraformModulesClient(t),
				acl:              acl.NewMockChecker(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					TerraformModulesClient: testMocks.terraformModules,
				},
				acl: testMocks.acl,
			}

			_, handler := updateTerraformModule(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output)
		})
	}
}

func TestDeleteTerraformModule(t *testing.T) {
	type testCase struct {
		name        string
		input       *deleteTerraformModuleInput
		mockSetup   func(*terraformModuleMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name:  "delete module successfully",
			input: &deleteTerraformModuleInput{ID: "m1"},
			mockSetup: func(m *terraformModuleMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "m1", trn.ResourceTypeTerraformModule).Return(nil)
				m.terraformModules.On("DeleteTerraformModule", mock.Anything, &pb.DeleteTerraformModuleRequest{Id: "m1"}).Return(nil, nil)
			},
		},
		{
			name:  "acl denial",
			input: &deleteTerraformModuleInput{ID: "m1"},
			mockSetup: func(m *terraformModuleMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "m1", trn.ResourceTypeTerraformModule).Return(status.Error(codes.PermissionDenied, "access denied"))
			},
			expectError: true,
		},
		{
			name:  "module not found",
			input: &deleteTerraformModuleInput{ID: "nonexistent"},
			mockSetup: func(m *terraformModuleMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "nonexistent", trn.ResourceTypeTerraformModule).Return(nil)
				m.terraformModules.On("DeleteTerraformModule", mock.Anything, &pb.DeleteTerraformModuleRequest{Id: "nonexistent"}).Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &terraformModuleMocks{
				terraformModules: mocks.NewTerraformModulesClient(t),
				acl:              acl.NewMockChecker(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					TerraformModulesClient: testMocks.terraformModules,
				},
				acl: testMocks.acl,
			}

			_, handler := deleteTerraformModule(toolCtx)
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
