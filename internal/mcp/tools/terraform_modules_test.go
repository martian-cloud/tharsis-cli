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

func TestListTerraformModules(t *testing.T) {
	type testCase struct {
		name     string
		input    listTerraformModulesInput
		modules  []sdktypes.TerraformModule
		pageInfo sdktypes.PageInfo
		validate func(*testing.T, listTerraformModulesOutput)
	}

	tests := []testCase{
		{
			name:  "list modules without filter",
			input: listTerraformModulesInput{},
			modules: []sdktypes.TerraformModule{
				{
					Metadata:  sdktypes.ResourceMetadata{ID: "module-1", TRN: "trn:terraform_module:group/vpc/aws"},
					Name:      "vpc",
					System:    "aws",
					GroupPath: "group",
					Private:   false,
				},
				{
					Metadata:  sdktypes.ResourceMetadata{ID: "module-2", TRN: "trn:terraform_module:group/s3/aws"},
					Name:      "s3",
					System:    "aws",
					GroupPath: "group",
					Private:   true,
				},
			},
			pageInfo: sdktypes.PageInfo{
				HasNextPage: false,
			},
			validate: func(t *testing.T, output listTerraformModulesOutput) {
				assert.Len(t, output.Modules, 2)
				assert.Equal(t, "module-1", output.Modules[0].ID)
				assert.Equal(t, "trn:terraform_module:group/vpc/aws", output.Modules[0].TRN)
				assert.Equal(t, "vpc", output.Modules[0].Name)
				assert.False(t, output.PageInfo.HasNextPage)
			},
		},
		{
			name: "list modules with search",
			input: listTerraformModulesInput{
				Search: ptr.String("vpc"),
			},
			modules: []sdktypes.TerraformModule{
				{
					Metadata:  sdktypes.ResourceMetadata{ID: "module-1", TRN: "trn:terraform_module:group/vpc/aws"},
					Name:      "vpc",
					System:    "aws",
					GroupPath: "group",
				},
			},
			pageInfo: sdktypes.PageInfo{},
			validate: func(t *testing.T, output listTerraformModulesOutput) {
				assert.Len(t, output.Modules, 1)
				assert.Equal(t, "vpc", output.Modules[0].Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockModule := tharsis.NewTerraformModule(t)

			mockClient.On("TerraformModules").Return(mockModule)
			mockModule.On("GetModules", mock.Anything, mock.Anything).
				Return(&sdktypes.GetTerraformModulesOutput{
					TerraformModules: tt.modules,
					PageInfo:         &tt.pageInfo,
				}, nil)

			tc := &ToolContext{
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
			}

			_, handler := listTerraformModules(tc)
			_, output, err := handler(t.Context(), nil, tt.input)

			assert.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

func TestGetTerraformModule(t *testing.T) {
	moduleID := "module-id"

	type testCase struct {
		name     string
		input    getTerraformModuleInput
		module   *sdktypes.TerraformModule
		validate func(*testing.T, getTerraformModuleOutput)
	}

	tests := []testCase{
		{
			name: "get module by ID",
			input: getTerraformModuleInput{
				ID: moduleID,
			},
			module: &sdktypes.TerraformModule{
				Metadata:  sdktypes.ResourceMetadata{ID: moduleID, TRN: "trn:terraform_module:group/vpc/aws"},
				Name:      "vpc",
				System:    "aws",
				GroupPath: "group",
			},
			validate: func(t *testing.T, output getTerraformModuleOutput) {
				assert.Equal(t, moduleID, output.Module.ID)
				assert.Equal(t, "vpc", output.Module.Name)
			},
		},
		{
			name: "get module by TRN",
			input: getTerraformModuleInput{
				ID: "trn:terraform_module:group/vpc/aws",
			},
			module: &sdktypes.TerraformModule{
				Metadata:  sdktypes.ResourceMetadata{ID: moduleID, TRN: "trn:terraform_module:group/vpc/aws"},
				Name:      "vpc",
				System:    "aws",
				GroupPath: "group",
			},
			validate: func(t *testing.T, output getTerraformModuleOutput) {
				assert.Equal(t, "trn:terraform_module:group/vpc/aws", output.Module.TRN)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockModule := tharsis.NewTerraformModule(t)

			mockClient.On("TerraformModules").Return(mockModule)
			mockModule.On("GetModule", mock.Anything, mock.Anything).Return(tt.module, nil)

			tc := &ToolContext{
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
			}

			_, handler := getTerraformModule(tc)
			_, output, err := handler(t.Context(), nil, tt.input)

			assert.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

func TestCreateTerraformModule(t *testing.T) {
	type testCase struct {
		name        string
		input       createTerraformModuleInput
		module      *sdktypes.TerraformModule
		aclError    error
		expectError bool
		validate    func(*testing.T, createTerraformModuleOutput)
	}

	tests := []testCase{
		{
			name: "successful module creation",
			input: createTerraformModuleInput{
				Name:          "vpc",
				System:        "aws",
				GroupPath:     "group",
				RepositoryURL: "https://github.com/org/vpc",
				Private:       false,
			},
			module: &sdktypes.TerraformModule{
				Metadata:      sdktypes.ResourceMetadata{ID: "module-id", TRN: "trn:terraform_module:group/vpc/aws"},
				Name:          "vpc",
				System:        "aws",
				GroupPath:     "group",
				RepositoryURL: "https://github.com/org/vpc",
				Private:       false,
			},
			validate: func(t *testing.T, output createTerraformModuleOutput) {
				assert.Equal(t, "module-id", output.Module.ID)
				assert.Equal(t, "vpc", output.Module.Name)
				assert.Equal(t, "https://github.com/org/vpc", output.Module.RepositoryURL)
			},
		},
		{
			name: "ACL denial",
			input: createTerraformModuleInput{
				Name:          "vpc",
				System:        "aws",
				GroupPath:     "group",
				RepositoryURL: "https://github.com/org/vpc",
			},
			aclError:    assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockModule := tharsis.NewTerraformModule(t)
			mockACL := acl.NewMockChecker(t)

			mockACL.On("Authorize", mock.Anything, mockClient, "trn:group:"+tt.input.GroupPath, trn.ResourceTypeGroup).Return(tt.aclError)

			if tt.aclError == nil {
				mockClient.On("TerraformModules").Return(mockModule)
				mockModule.On("CreateModule", mock.Anything, mock.Anything).Return(tt.module, nil)
			}

			tc := &ToolContext{
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := createTerraformModule(tc)
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

func TestUpdateTerraformModule(t *testing.T) {
	moduleID := "module-id"

	type testCase struct {
		name        string
		input       updateTerraformModuleInput
		module      *sdktypes.TerraformModule
		aclError    error
		expectError bool
		validate    func(*testing.T, updateTerraformModuleOutput)
	}

	tests := []testCase{
		{
			name: "successful module update",
			input: updateTerraformModuleInput{
				ID:            moduleID,
				RepositoryURL: ptr.String("https://github.com/org/new-vpc"),
				Private:       ptr.Bool(true),
			},
			module: &sdktypes.TerraformModule{
				Metadata:      sdktypes.ResourceMetadata{ID: moduleID, TRN: "trn:terraform_module:group/vpc/aws"},
				Name:          "vpc",
				System:        "aws",
				GroupPath:     "group",
				RepositoryURL: "https://github.com/org/new-vpc",
				Private:       true,
			},
			validate: func(t *testing.T, output updateTerraformModuleOutput) {
				assert.Equal(t, moduleID, output.Module.ID)
				assert.Equal(t, "https://github.com/org/new-vpc", output.Module.RepositoryURL)
				assert.True(t, output.Module.Private)
			},
		},
		{
			name: "ACL denial",
			input: updateTerraformModuleInput{
				ID:            moduleID,
				RepositoryURL: ptr.String("https://github.com/org/new-vpc"),
			},
			aclError:    assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := tharsis.NewMockClient(t)
			mockModule := tharsis.NewTerraformModule(t)
			mockACL := acl.NewMockChecker(t)

			mockACL.On("Authorize", mock.Anything, mockClient, moduleID, trn.ResourceTypeTerraformModule).Return(tt.aclError)

			if tt.aclError == nil {
				mockClient.On("TerraformModules").Return(mockModule)
				mockModule.On("UpdateModule", mock.Anything, mock.Anything).Return(tt.module, nil)
			}

			tc := &ToolContext{
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := updateTerraformModule(tc)
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

func TestDeleteTerraformModule(t *testing.T) {
	moduleID := "module-id"

	type testCase struct {
		name        string
		aclError    error
		expectError bool
		validate    func(*testing.T, deleteTerraformModuleOutput)
	}

	tests := []testCase{
		{
			name: "successful module deletion",
			validate: func(t *testing.T, output deleteTerraformModuleOutput) {
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
			mockModule := tharsis.NewTerraformModule(t)
			mockACL := acl.NewMockChecker(t)

			mockACL.On("Authorize", mock.Anything, mockClient, moduleID, trn.ResourceTypeTerraformModule).Return(tt.aclError)

			if tt.aclError == nil {
				mockClient.On("TerraformModules").Return(mockModule)
				mockModule.On("DeleteModule", mock.Anything, &sdktypes.DeleteTerraformModuleInput{
					ID: moduleID,
				}).Return(nil)
			}

			tc := &ToolContext{
				clientGetter: func() (tharsis.Client, error) {
					return mockClient, nil
				},
				acl: mockACL,
			}

			_, handler := deleteTerraformModule(tc)
			_, output, err := handler(t.Context(), nil, deleteTerraformModuleInput{ID: moduleID})

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
