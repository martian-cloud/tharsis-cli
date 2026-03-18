package tools

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"testing"

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

type configurationVersionMocks struct {
	configurationVersions *mocks.ConfigurationVersionsClient
	acl                   *acl.MockChecker
	tfe                   *tfe.MockRESTClient
}

func TestGetConfigurationVersion(t *testing.T) {
	type testCase struct {
		name        string
		input       *getConfigurationVersionInput
		mockSetup   func(*configurationVersionMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name:  "get configuration version successfully",
			input: &getConfigurationVersionInput{ID: "cv1"},
			mockSetup: func(m *configurationVersionMocks) {
				m.configurationVersions.On("GetConfigurationVersionByID", mock.Anything, &pb.GetConfigurationVersionByIDRequest{Id: "cv1"}).
					Return(&pb.ConfigurationVersion{
						Metadata:    &pb.ResourceMetadata{Id: "cv1", Trn: "trn:configuration_version:cv1"},
						Status:      "uploaded",
						WorkspaceId: "ws1",
						Speculative: false,
					}, nil)
			},
		},
		{
			name:  "configuration version not found",
			input: &getConfigurationVersionInput{ID: "nonexistent"},
			mockSetup: func(m *configurationVersionMocks) {
				m.configurationVersions.On("GetConfigurationVersionByID", mock.Anything, &pb.GetConfigurationVersionByIDRequest{Id: "nonexistent"}).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &configurationVersionMocks{
				configurationVersions: mocks.NewConfigurationVersionsClient(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					ConfigurationVersionsClient: testMocks.configurationVersions,
				},
			}

			_, handler := getConfigurationVersion(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output.ConfigurationVersion)
		})
	}
}

func TestCreateConfigurationVersion(t *testing.T) {
	type testCase struct {
		name        string
		input       *createConfigurationVersionInput
		mockSetup   func(*configurationVersionMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name: "create and upload configuration version successfully",
			input: &createConfigurationVersionInput{
				WorkspaceID:   "ws1",
				DirectoryPath: "/path/to/config",
				Speculative:   false,
			},
			mockSetup: func(m *configurationVersionMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(nil)
				m.configurationVersions.On("CreateConfigurationVersion", mock.Anything, &pb.CreateConfigurationVersionRequest{
					WorkspaceId: "ws1",
					Speculative: false,
				}).Return(&pb.ConfigurationVersion{
					Metadata:    &pb.ResourceMetadata{Id: "cv1", Trn: "trn:configuration_version:cv1"},
					Status:      "pending",
					WorkspaceId: "ws1",
					Speculative: false,
				}, nil)
				m.tfe.On("UploadConfigurationVersion", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "upload uses resolved workspace ID from response",
			input: &createConfigurationVersionInput{
				WorkspaceID:   "trn:workspace:group/my-workspace",
				DirectoryPath: "/path/to/config",
			},
			mockSetup: func(m *configurationVersionMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "trn:workspace:group/my-workspace", trn.ResourceTypeWorkspace).Return(nil)
				m.configurationVersions.On("CreateConfigurationVersion", mock.Anything, &pb.CreateConfigurationVersionRequest{
					WorkspaceId: "trn:workspace:group/my-workspace",
				}).Return(&pb.ConfigurationVersion{
					Metadata:    &pb.ResourceMetadata{Id: "cv1", Trn: "trn:configuration_version:cv1"},
					Status:      "pending",
					WorkspaceId: "resolved-ws-id",
				}, nil)
				m.tfe.On("UploadConfigurationVersion", mock.Anything, mock.MatchedBy(func(input *tfe.UploadConfigurationVersionInput) bool {
					return input.WorkspaceID == "resolved-ws-id" && input.ConfigVersionID == "cv1"
				})).Return(nil)
			},
		},
		{
			name: "authorization failure",
			input: &createConfigurationVersionInput{
				WorkspaceID:   "ws1",
				DirectoryPath: "/path/to/config",
			},
			mockSetup: func(m *configurationVersionMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(status.Error(codes.PermissionDenied, "access denied"))
			},
			expectError: true,
		},
		{
			name: "create configuration version fails",
			input: &createConfigurationVersionInput{
				WorkspaceID:   "ws1",
				DirectoryPath: "/path/to/config",
			},
			mockSetup: func(m *configurationVersionMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(nil)
				m.configurationVersions.On("CreateConfigurationVersion", mock.Anything, mock.Anything).Return(nil, status.Error(codes.Internal, "internal error"))
			},
			expectError: true,
		},
		{
			name: "upload fails",
			input: &createConfigurationVersionInput{
				WorkspaceID:   "ws1",
				DirectoryPath: "/path/to/config",
			},
			mockSetup: func(m *configurationVersionMocks) {
				m.acl.On("Authorize", mock.Anything, mock.Anything, "ws1", trn.ResourceTypeWorkspace).Return(nil)
				m.configurationVersions.On("CreateConfigurationVersion", mock.Anything, mock.Anything).Return(&pb.ConfigurationVersion{
					Metadata:    &pb.ResourceMetadata{Id: "cv1", Trn: "trn:configuration_version:cv1"},
					Status:      "pending",
					WorkspaceId: "ws1",
				}, nil)
				m.tfe.On("UploadConfigurationVersion", mock.Anything, mock.Anything).Return(status.Error(codes.Internal, "upload failed"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &configurationVersionMocks{
				configurationVersions: mocks.NewConfigurationVersionsClient(t),
				acl:                   acl.NewMockChecker(t),
				tfe:                   tfe.NewMockRESTClient(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					ConfigurationVersionsClient: testMocks.configurationVersions,
				},
				acl:       testMocks.acl,
				tfeClient: testMocks.tfe,
			}

			_, handler := createConfigurationVersion(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, output.ConfigurationVersion)
		})
	}
}

func TestDownloadConfigurationVersion(t *testing.T) {
	type testCase struct {
		name        string
		input       *downloadConfigurationVersionInput
		mockSetup   func(*configurationVersionMocks)
		expectError bool
	}

	testCases := []testCase{
		{
			name:  "download configuration version successfully",
			input: &downloadConfigurationVersionInput{ID: "cv1"},
			mockSetup: func(m *configurationVersionMocks) {
				m.configurationVersions.On("GetConfigurationVersionByID", mock.Anything, &pb.GetConfigurationVersionByIDRequest{Id: "cv1"}).
					Return(&pb.ConfigurationVersion{
						Metadata:    &pb.ResourceMetadata{Id: "cv1"},
						WorkspaceId: "ws1",
					}, nil)
				m.tfe.On("DownloadConfigurationVersion", mock.Anything, mock.MatchedBy(func(input *tfe.DownloadConfigurationVersionInput) bool {
					if input.ConfigVersionID != "cv1" {
						return false
					}
					// Write a valid empty tar.gz to the writer
					gzWriter := gzip.NewWriter(input.Writer)
					tarWriter := tar.NewWriter(gzWriter)
					_ = tarWriter.Close()
					_ = gzWriter.Close()
					return true
				})).Return(nil)
			},
		},
		{
			name:  "download resolves TRN to metadata ID",
			input: &downloadConfigurationVersionInput{ID: "trn:configuration_version:group/ws/cv1"},
			mockSetup: func(m *configurationVersionMocks) {
				m.configurationVersions.On("GetConfigurationVersionByID", mock.Anything, &pb.GetConfigurationVersionByIDRequest{Id: "trn:configuration_version:group/ws/cv1"}).
					Return(&pb.ConfigurationVersion{
						Metadata:    &pb.ResourceMetadata{Id: "resolved-cv-id"},
						WorkspaceId: "ws1",
					}, nil)
				m.tfe.On("DownloadConfigurationVersion", mock.Anything, mock.MatchedBy(func(input *tfe.DownloadConfigurationVersionInput) bool {
					if input.ConfigVersionID != "resolved-cv-id" {
						return false
					}
					gzWriter := gzip.NewWriter(input.Writer)
					tarWriter := tar.NewWriter(gzWriter)
					_ = tarWriter.Close()
					_ = gzWriter.Close()
					return true
				})).Return(nil)
			},
		},
		{
			name:  "get configuration version fails",
			input: &downloadConfigurationVersionInput{ID: "cv1"},
			mockSetup: func(m *configurationVersionMocks) {
				m.configurationVersions.On("GetConfigurationVersionByID", mock.Anything, mock.Anything).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
		{
			name:  "download fails",
			input: &downloadConfigurationVersionInput{ID: "cv1"},
			mockSetup: func(m *configurationVersionMocks) {
				m.configurationVersions.On("GetConfigurationVersionByID", mock.Anything, mock.Anything).
					Return(&pb.ConfigurationVersion{
						Metadata:    &pb.ResourceMetadata{Id: "cv1"},
						WorkspaceId: "ws1",
					}, nil)
				m.tfe.On("DownloadConfigurationVersion", mock.Anything, mock.Anything).
					Return(status.Error(codes.NotFound, "not found"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMocks := &configurationVersionMocks{
				configurationVersions: mocks.NewConfigurationVersionsClient(t),
				tfe:                   tfe.NewMockRESTClient(t),
			}

			if tc.mockSetup != nil {
				tc.mockSetup(testMocks)
			}

			toolCtx := &ToolContext{
				grpcClient: &client.Client{
					ConfigurationVersionsClient: testMocks.configurationVersions,
				},
				tfeClient: testMocks.tfe,
			}

			_, handler := downloadConfigurationVersion(toolCtx)
			_, output, err := handler(t.Context(), nil, tc.input)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, output.OutputPath)
			require.DirExists(t, output.OutputPath)
			t.Cleanup(func() {
				_ = os.RemoveAll(output.OutputPath)
			})
		})
	}
}
