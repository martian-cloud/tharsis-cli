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
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tfe"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type terraformModuleVersionMocks struct {
	terraformModules *mocks.TerraformModulesClient
	acl              *acl.MockChecker
	tfe              *tfe.MockRESTClient
}

func TestListTerraformModuleVersions(t *testing.T) {
	type testCase struct {
		name          string
		input         *listTerraformModuleVersionsInput
		mockSetup     func(*terraformModuleVersionMocks)
		expectError   bool
		expectResults int
	}

	testCases := []testCase{
		{
			name:  "list module versions successfully",
			input: &listTerraformModuleVersionsInput{ModuleID: "m1", Limit: ptr.Int32(10)},
			mockSetup: func(m *terraformModuleVersionMocks) {
				m.terraformModules.On("GetTerraformModuleVersions", mock.Anything, &pb.GetTerraformModuleVersionsRequest{
					ModuleId:          "m1",
					PaginationOptions: &pb.PaginationOptions{First: ptr.Int32(10)},
				}).Return(&pb.GetTerraformModuleVersionsResponse{
					Versions: []*pb.TerraformModuleVersion{
						{Metadata: &pb.ResourceMetadata{
							Id:  "mv1",
							Trn: "trn:terraform_module_version:group/module/aws/1.0.0",
						},
							ModuleId:        "m1",
							SemanticVersion: "1.0.0",
							Status:          "uploaded",
						},
					},
					PageInfo: &pb.PageInfo{HasNextPage: false},
				}, nil)
			},
			expectResults: 1,
		},
		{
			name:  "list call failed",
			input: &listTerraformModuleVersionsInput{ModuleID: "m1", Limit: ptr.Int32(10)},
			mockSetup: func(m *terraformModuleVersionMocks) {
				m.terraformModules.On("GetTerraformModuleVersions", mock.Anything, &pb.GetTerraformModuleVersionsRequest{
					ModuleId:          "m1",
					PaginationOptions: &pb.PaginationOptions{First: ptr.Int32(10)},
				}).Return(nil, status.Error(codes.Internal, "internal error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &terraformModuleVersionMocks{
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

			_, handler := listTerraformModuleVersions(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, output.ModuleVersions, tc.expectResults)
		})
	}
}

func TestGetTerraformModuleVersion(t *testing.T) {
	type testCase struct {
		name        string
		input       *getTerraformModuleVersionInput
		mockSetup   func(*terraformModuleVersionMocks)
		expectError bool
		expectID    string
	}

	testCases := []testCase{
		{
			name:  "get module version successfully",
			input: &getTerraformModuleVersionInput{ID: "mv1"},
			mockSetup: func(m *terraformModuleVersionMocks) {
				m.terraformModules.On("GetTerraformModuleVersionByID", mock.Anything, &pb.GetTerraformModuleVersionByIDRequest{Id: "mv1"}).Return(&pb.TerraformModuleVersion{
					Metadata:        &pb.ResourceMetadata{Id: "mv1", Trn: "trn:terraform_module_version:group/module/aws/1.0.0"},
					ModuleId:        "m1",
					SemanticVersion: "1.0.0",
					Status:          "uploaded",
				}, nil)
			},
			expectID: "mv1",
		},
		{
			name:  "module version not found",
			input: &getTerraformModuleVersionInput{ID: "nonexistent"},
			mockSetup: func(m *terraformModuleVersionMocks) {
				m.terraformModules.On("GetTerraformModuleVersionByID", mock.Anything, &pb.GetTerraformModuleVersionByIDRequest{Id: "nonexistent"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &terraformModuleVersionMocks{
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

			_, handler := getTerraformModuleVersion(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectID, output.ModuleVersion.ID)
		})
	}
}

func TestDeleteTerraformModuleVersion(t *testing.T) {
	type testCase struct {
		name        string
		input       *deleteTerraformModuleVersionInput
		mockSetup   func(*terraformModuleVersionMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name:  "delete module version successfully",
			input: &deleteTerraformModuleVersionInput{ID: "mv1"},
			mockSetup: func(m *terraformModuleVersionMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "mv1", trn.ResourceTypeTerraformModuleVersion).Return(nil)
				m.terraformModules.On("DeleteTerraformModuleVersion", mock.Anything, &pb.DeleteTerraformModuleVersionRequest{Id: "mv1"}).Return(nil, nil)
			},
		},
		{
			name:  "module version not found",
			input: &deleteTerraformModuleVersionInput{ID: "nonexistent"},
			mockSetup: func(m *terraformModuleVersionMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "nonexistent", trn.ResourceTypeTerraformModuleVersion).Return(nil)
				m.terraformModules.On("DeleteTerraformModuleVersion", mock.Anything, &pb.DeleteTerraformModuleVersionRequest{Id: "nonexistent"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
		{
			name:  "acl denial",
			input: &deleteTerraformModuleVersionInput{ID: "mv1"},
			mockSetup: func(m *terraformModuleVersionMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "mv1", trn.ResourceTypeTerraformModuleVersion).
					Return(status.Error(codes.PermissionDenied, "access denied"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &terraformModuleVersionMocks{
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

			_, handler := deleteTerraformModuleVersion(toolCtx)
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

func TestUploadModuleVersion(t *testing.T) {
	type testCase struct {
		name        string
		setupDir    func() string
		input       func(string) *uploadModuleVersionInput
		mockSetup   func(*terraformModuleVersionMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name: "upload module version successfully",
			setupDir: func() string {
				dir := t.TempDir()
				// Create a minimal terraform file
				_ = os.WriteFile(filepath.Join(dir, "main.tf"), []byte("# test"), 0600)
				return dir
			},
			input: func(dir string) *uploadModuleVersionInput {
				return &uploadModuleVersionInput{
					ModuleID:      "m1",
					Version:       "1.0.0",
					DirectoryPath: dir,
				}
			},
			mockSetup: func(m *terraformModuleVersionMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "m1", trn.ResourceTypeTerraformModule).Return(nil)
				m.terraformModules.On("CreateTerraformModuleVersion", mock.Anything, mock.MatchedBy(func(input *pb.CreateTerraformModuleVersionRequest) bool {
					return input.ModuleId == "m1" && input.Version == "1.0.0" && input.ShaSum != ""
				})).Return(&pb.TerraformModuleVersion{
					Metadata:        &pb.ResourceMetadata{Id: "mv1", Trn: "trn:terraform_module_version:group/module/aws/1.0.0"},
					ModuleId:        "m1",
					SemanticVersion: "1.0.0",
					Status:          "pending",
				}, nil)
				m.tfe.On("UploadModuleVersion", mock.Anything, mock.MatchedBy(func(input *tfe.UploadModuleVersionInput) bool {
					return input.ModuleVersionID == "mv1"
				})).Return(nil)
			},
		},
		{
			name: "acl denial",
			setupDir: func() string {
				return t.TempDir()
			},
			input: func(dir string) *uploadModuleVersionInput {
				return &uploadModuleVersionInput{
					ModuleID:      "m1",
					Version:       "1.0.0",
					DirectoryPath: dir,
				}
			},
			mockSetup: func(m *terraformModuleVersionMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "m1", trn.ResourceTypeTerraformModule).Return(status.Error(codes.PermissionDenied, "access denied"))
			},
			expectError: true,
		},
		{
			name: "directory does not exist",
			setupDir: func() string {
				return "/nonexistent/path"
			},
			input: func(dir string) *uploadModuleVersionInput {
				return &uploadModuleVersionInput{
					ModuleID:      "m1",
					Version:       "1.0.0",
					DirectoryPath: dir,
				}
			},
			mockSetup: func(m *terraformModuleVersionMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "m1", trn.ResourceTypeTerraformModule).Return(nil)
			},
			expectError: true,
		},
		{
			name: "create version fails",
			setupDir: func() string {
				dir := t.TempDir()
				_ = os.WriteFile(filepath.Join(dir, "main.tf"), []byte("# test"), 0600)
				return dir
			},
			input: func(dir string) *uploadModuleVersionInput {
				return &uploadModuleVersionInput{
					ModuleID:      "m1",
					Version:       "1.0.0",
					DirectoryPath: dir,
				}
			},
			mockSetup: func(m *terraformModuleVersionMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "m1", trn.ResourceTypeTerraformModule).Return(nil)
				m.terraformModules.On("CreateTerraformModuleVersion", mock.Anything, mock.MatchedBy(func(input *pb.CreateTerraformModuleVersionRequest) bool {
					return input.ModuleId == "m1" && input.Version == "1.0.0" && input.ShaSum != ""
				})).Return(nil, status.Error(codes.Internal, "internal error"))
			},
			expectError: true,
		},
		{
			name: "upload fails",
			setupDir: func() string {
				dir := t.TempDir()
				_ = os.WriteFile(filepath.Join(dir, "main.tf"), []byte("# test"), 0600)
				return dir
			},
			input: func(dir string) *uploadModuleVersionInput {
				return &uploadModuleVersionInput{
					ModuleID:      "m1",
					Version:       "1.0.0",
					DirectoryPath: dir,
				}
			},
			mockSetup: func(m *terraformModuleVersionMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "m1", trn.ResourceTypeTerraformModule).Return(nil)
				m.terraformModules.On("CreateTerraformModuleVersion", mock.Anything, mock.MatchedBy(func(input *pb.CreateTerraformModuleVersionRequest) bool {
					return input.ModuleId == "m1" && input.Version == "1.0.0" && input.ShaSum != ""
				})).Return(&pb.TerraformModuleVersion{
					Metadata:        &pb.ResourceMetadata{Id: "mv1", Trn: "trn:terraform_module_version:group/module/aws/1.0.0"},
					ModuleId:        "m1",
					SemanticVersion: "1.0.0",
					Status:          "pending",
				}, nil)
				m.tfe.On("UploadModuleVersion", mock.Anything, mock.MatchedBy(func(input *tfe.UploadModuleVersionInput) bool {
					return input.ModuleVersionID == "mv1"
				})).Return(status.Error(codes.Internal, "upload failed"))
				m.terraformModules.On("DeleteTerraformModuleVersion", mock.Anything, &pb.DeleteTerraformModuleVersionRequest{Id: "mv1"}).Return(nil, nil)
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := tc.setupDir()
			input := tc.input(dir)

			testMocks := &terraformModuleVersionMocks{
				terraformModules: mocks.NewTerraformModulesClient(t),
				acl:              acl.NewMockChecker(t),
				tfe:              tfe.NewMockRESTClient(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					TerraformModulesClient: testMocks.terraformModules,
				},
				acl:       testMocks.acl,
				tfeClient: testMocks.tfe,
			}

			_, handler := uploadModuleVersion(toolCtx)
			_, output, err := handler(t.Context(), nil, input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, output.ModuleVersion)
		})
	}
}
