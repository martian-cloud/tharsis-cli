package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/acl"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tharsis"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestListTerraformModuleVersions(t *testing.T) {
	type testCase struct {
		name     string
		input    listTerraformModuleVersionsInput
		versions []sdktypes.TerraformModuleVersion
		pageInfo sdktypes.PageInfo
		validate func(*testing.T, listTerraformModuleVersionsOutput)
	}

	tests := []testCase{
		{
			name: "list versions without filter",
			input: listTerraformModuleVersionsInput{
				ModuleID: "module-id",
			},
			versions: []sdktypes.TerraformModuleVersion{
				{
					Metadata: sdktypes.ResourceMetadata{ID: "version-1", TRN: "trn:terraform_module_version:group/vpc/aws/1.0.0"},
					ModuleID: "module-id",
					Version:  "1.0.0",
					SHASum:   "abc123",
					Status:   "uploaded",
					Latest:   false,
				},
				{
					Metadata: sdktypes.ResourceMetadata{ID: "version-2", TRN: "trn:terraform_module_version:group/vpc/aws/1.1.0"},
					ModuleID: "module-id",
					Version:  "1.1.0",
					SHASum:   "def456",
					Status:   "uploaded",
					Latest:   true,
				},
			},
			pageInfo: sdktypes.PageInfo{
				HasNextPage: false,
			},
			validate: func(t *testing.T, output listTerraformModuleVersionsOutput) {
				assert.Len(t, output.ModuleVersions, 2)
				assert.Equal(t, "version-1", output.ModuleVersions[0].ID)
				assert.Equal(t, "1.0.0", output.ModuleVersions[0].Version)
				assert.Equal(t, "1.1.0", output.ModuleVersions[1].Version)
				assert.True(t, output.ModuleVersions[1].Latest)
				assert.False(t, output.PageInfo.HasNextPage)
			},
		},
		{
			name: "list versions with sort",
			input: func() listTerraformModuleVersionsInput {
				sort := sdktypes.TerraformModuleVersionSortableFieldCreatedAtDesc
				return listTerraformModuleVersionsInput{
					ModuleID: "module-id",
					Sort:     &sort,
				}
			}(),
			versions: []sdktypes.TerraformModuleVersion{
				{
					Metadata: sdktypes.ResourceMetadata{ID: "version-2", TRN: "trn:terraform_module_version:group/vpc/aws/1.1.0"},
					ModuleID: "module-id",
					Version:  "1.1.0",
					Status:   "uploaded",
					Latest:   true,
				},
			},
			pageInfo: sdktypes.PageInfo{},
			validate: func(t *testing.T, output listTerraformModuleVersionsOutput) {
				assert.Len(t, output.ModuleVersions, 1)
				assert.Equal(t, "1.1.0", output.ModuleVersions[0].Version)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockModuleVersion := tharsis.NewTerraformModuleVersion(t)

			mockClient.On("TerraformModuleVersions").Return(mockModuleVersion)
			mockModuleVersion.On("GetModuleVersions", mock.Anything, mock.Anything).
				Return(&sdktypes.GetTerraformModuleVersionsOutput{
					ModuleVersions: tt.versions,
					PageInfo:       &tt.pageInfo,
				}, nil)

			tc := &ToolContext{
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
			}

			_, handler := listTerraformModuleVersions(tc)
			_, output, err := handler(t.Context(), nil, tt.input)

			assert.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

func TestGetTerraformModuleVersion(t *testing.T) {
	versionID := "version-id"

	type testCase struct {
		name     string
		input    getTerraformModuleVersionInput
		version  *sdktypes.TerraformModuleVersion
		validate func(*testing.T, getTerraformModuleVersionOutput)
	}

	tests := []testCase{
		{
			name: "get version by ID",
			input: getTerraformModuleVersionInput{
				ID: versionID,
			},
			version: &sdktypes.TerraformModuleVersion{
				Metadata: sdktypes.ResourceMetadata{ID: versionID, TRN: "trn:terraform_module_version:group/vpc/aws/1.0.0"},
				ModuleID: "module-id",
				Version:  "1.0.0",
				SHASum:   "abc123",
				Status:   "uploaded",
				Latest:   true,
			},
			validate: func(t *testing.T, output getTerraformModuleVersionOutput) {
				assert.Equal(t, versionID, output.ModuleVersion.ID)
				assert.Equal(t, "1.0.0", output.ModuleVersion.Version)
				assert.True(t, output.ModuleVersion.Latest)
			},
		},
		{
			name: "get version by TRN",
			input: getTerraformModuleVersionInput{
				ID: "trn:terraform_module_version:group/vpc/aws/1.0.0",
			},
			version: &sdktypes.TerraformModuleVersion{
				Metadata: sdktypes.ResourceMetadata{ID: versionID, TRN: "trn:terraform_module_version:group/vpc/aws/1.0.0"},
				ModuleID: "module-id",
				Version:  "1.0.0",
				SHASum:   "abc123",
				Status:   "uploaded",
			},
			validate: func(t *testing.T, output getTerraformModuleVersionOutput) {
				assert.Equal(t, "1.0.0", output.ModuleVersion.Version)
				assert.Equal(t, "abc123", output.ModuleVersion.SHASum)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockModuleVersion := tharsis.NewTerraformModuleVersion(t)

			mockClient.On("TerraformModuleVersions").Return(mockModuleVersion)
			mockModuleVersion.On("GetModuleVersion", mock.Anything, mock.MatchedBy(func(input *sdktypes.GetTerraformModuleVersionInput) bool {
				return input.ID != nil && *input.ID == tt.input.ID
			})).Return(tt.version, nil)

			tc := &ToolContext{
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
			}

			_, handler := getTerraformModuleVersion(tc)
			_, output, err := handler(t.Context(), nil, tt.input)

			assert.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

func TestDeleteTerraformModuleVersion(t *testing.T) {
	versionID := "version-id"

	type testCase struct {
		name        string
		aclError    error
		expectError bool
		validate    func(*testing.T, deleteTerraformModuleVersionOutput)
	}

	tests := []testCase{
		{
			name: "successful version deletion",
			validate: func(t *testing.T, output deleteTerraformModuleVersionOutput) {
				assert.True(t, output.Success)
				assert.Contains(t, output.Message, "deleted successfully")
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
			mockModuleVersion := tharsis.NewTerraformModuleVersion(t)
			mockACL := acl.NewMockChecker(t)

			mockACL.On("Authorize", mock.Anything, mockClient, versionID, trn.ResourceTypeTerraformModuleVersion).Return(tt.aclError)

			if tt.aclError == nil {
				mockClient.On("TerraformModuleVersions").Return(mockModuleVersion)
				mockModuleVersion.On("DeleteModuleVersion", mock.Anything, &sdktypes.DeleteTerraformModuleVersionInput{
					ID: versionID,
				}).Return(nil)
			}

			tc := &ToolContext{
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := deleteTerraformModuleVersion(tc)
			_, output, err := handler(t.Context(), nil, deleteTerraformModuleVersionInput{ID: versionID})

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

func TestUploadModuleVersion(t *testing.T) {
	moduleID := "module-id"
	versionID := "version-id"

	type testCase struct {
		name        string
		input       uploadModuleVersionInput
		aclError    error
		expectError bool
		validate    func(*testing.T, uploadModuleVersionOutput)
	}

	tests := []testCase{
		{
			name: "successful upload",
			input: uploadModuleVersionInput{
				ModuleID:      moduleID,
				Version:       "1.0.0",
				DirectoryPath: t.TempDir(),
			},
			validate: func(t *testing.T, output uploadModuleVersionOutput) {
				assert.Contains(t, output.Message, "upload initiated")
				assert.Equal(t, versionID, output.ModuleVersion.ID)
			},
		},
		{
			name: "ACL denial",
			input: uploadModuleVersionInput{
				ModuleID:      moduleID,
				Version:       "1.0.0",
				DirectoryPath: t.TempDir(),
			},
			aclError:    assert.AnError,
			expectError: true,
		},
		{
			name: "directory does not exist",
			input: uploadModuleVersionInput{
				ModuleID:      moduleID,
				Version:       "1.0.0",
				DirectoryPath: "/nonexistent/path",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockModuleVersion := tharsis.NewTerraformModuleVersion(t)
			mockACL := acl.NewMockChecker(t)

			mockACL.On("Authorize", mock.Anything, mockClient, tt.input.ModuleID, trn.ResourceTypeTerraformModule).Return(tt.aclError)

			if tt.aclError == nil && tt.name != "directory does not exist" {
				mockClient.On("TerraformModuleVersions").Return(mockModuleVersion)
				mockModuleVersion.On("CreateModuleVersion", mock.Anything, mock.MatchedBy(func(input *sdktypes.CreateTerraformModuleVersionInput) bool {
					return input.Version == tt.input.Version
				})).Return(&sdktypes.TerraformModuleVersion{
					Metadata: sdktypes.ResourceMetadata{ID: versionID, TRN: "trn:terraform_module_version:group/module/system/1.0.0"},
					Version:  tt.input.Version,
					Status:   "pending",
				}, nil)
				mockModuleVersion.On("UploadModuleVersion", mock.Anything, versionID, mock.Anything).Return(nil)
			}

			tc := &ToolContext{
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := uploadModuleVersion(tc)
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
