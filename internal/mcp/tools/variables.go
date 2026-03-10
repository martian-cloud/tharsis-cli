package tools

import (
	"context"
	"fmt"

	"github.com/aws/smithy-go/ptr"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/varparser"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// setVariableInput is the input for setting a variable.
type setVariableInput struct {
	NamespaceID string `json:"namespace_id" jsonschema:"required,Workspace or group ID or TRN (e.g. Ul8yZ... or trn:workspace:group/workspace or trn:group:group/subgroup)"`
	Key         string `json:"key" jsonschema:"required,Variable name/key"`
	Value       string `json:"value" jsonschema:"required,Variable value"`
	Category    string `json:"category" jsonschema:"required,Variable category: TERRAFORM or ENVIRONMENT"`
}

// setVariableOutput is the output for setting a variable.
type setVariableOutput struct {
	Message string `json:"message" jsonschema:"Success message"`
	Success bool   `json:"success" jsonschema:"Whether operation was successful"`
}

// SetVariable returns an MCP tool for setting a namespace variable.
func setVariable(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*setVariableInput, *setVariableOutput]) {
	tool := mcp.Tool{
		Name:        "set_variable",
		Description: "Set a Terraform or environment variable on a workspace or group. Creates if doesn't exist, updates if it does. Note: Sensitive variables are not supported.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Set Variable",
			IdempotentHint:  true,
			DestructiveHint: ptr.Bool(true),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *setVariableInput) (*mcp.CallToolResult, *setVariableOutput, error) {
		categoryValue, ok := pb.VariableCategory_value[input.Category]
		if !ok {
			return nil, nil, fmt.Errorf("invalid category: must be 'TERRAFORM' or 'ENVIRONMENT'")
		}

		namespacePath, resourceType, err := getNamespacePath(ctx, tc.grpcClient, input.NamespaceID)
		if err != nil {
			return nil, nil, err
		}

		if err = tc.acl.Authorize(ctx, tc.grpcClient, input.NamespaceID, resourceType); err != nil {
			return nil, nil, err
		}

		variableID := trn.NewResourceTRN(trn.ResourceTypeVariable, namespacePath, input.Category, input.Key)
		variable, err := tc.grpcClient.NamespaceVariablesClient.GetNamespaceVariableByID(ctx, &pb.GetNamespaceVariableByIDRequest{Id: variableID})
		if err != nil && status.Code(err) != codes.NotFound {
			return nil, nil, fmt.Errorf("failed to get variable: %w", err)
		}

		if variable != nil {
			_, err = tc.grpcClient.NamespaceVariablesClient.UpdateNamespaceVariable(ctx, &pb.UpdateNamespaceVariableRequest{
				Id:    variable.Metadata.Id,
				Key:   variable.Key,
				Value: input.Value,
			})
			if err != nil {
				return nil, nil, fmt.Errorf("failed to update variable: %w", err)
			}
		} else {
			if _, err = tc.grpcClient.NamespaceVariablesClient.CreateNamespaceVariable(ctx, &pb.CreateNamespaceVariableRequest{
				Key:           input.Key,
				Value:         input.Value,
				Category:      pb.VariableCategory(categoryValue),
				NamespacePath: namespacePath,
			}); err != nil {
				return nil, nil, fmt.Errorf("failed to create variable: %w", err)
			}
		}

		return nil, &setVariableOutput{
			Message: fmt.Sprintf("Variable '%s' set successfully in namespace %s", input.Key, namespacePath),
			Success: true,
		}, nil
	}

	return tool, handler
}

// setTerraformVariablesFromFileInput is the input for setting Terraform variables from a file.
type setTerraformVariablesFromFileInput struct {
	NamespaceID string `json:"namespace_id" jsonschema:"required,Workspace or group ID or TRN (e.g. Ul8yZ... or trn:workspace:group/workspace or trn:group:group/subgroup)"`
	FilePath    string `json:"file_path" jsonschema:"required,Path to .tfvars file containing Terraform variable definitions"`
}

// setTerraformVariablesFromFileOutput is the output for setting Terraform variables from a file.
type setTerraformVariablesFromFileOutput struct {
	Message string `json:"message" jsonschema:"Success message"`
	Success bool   `json:"success" jsonschema:"Whether operation was successful"`
	Count   int    `json:"count" jsonschema:"Number of variables set"`
}

// SetTerraformVariablesFromFile returns an MCP tool for setting multiple Terraform variables from a file.
func setTerraformVariablesFromFile(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*setTerraformVariablesFromFileInput, *setTerraformVariablesFromFileOutput]) {
	tool := mcp.Tool{
		Name:        "set_terraform_variables_from_file",
		Description: "Set multiple Terraform variables from a .tfvars file on a workspace or group. Overwrites existing Terraform variables.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Set Terraform Variables From File",
			IdempotentHint:  true,
			DestructiveHint: ptr.Bool(true),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *setTerraformVariablesFromFileInput) (*mcp.CallToolResult, *setTerraformVariablesFromFileOutput, error) {
		namespacePath, resourceType, err := getNamespacePath(ctx, tc.grpcClient, input.NamespaceID)
		if err != nil {
			return nil, nil, err
		}

		if err = tc.acl.Authorize(ctx, tc.grpcClient, input.NamespaceID, resourceType); err != nil {
			return nil, nil, err
		}

		parser := varparser.NewVariableParser(nil, false)
		variables, err := parser.ParseTerraformVariables(&varparser.ParseTerraformVariablesInput{
			TfVarFilePaths: []string{input.FilePath},
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse variables file: %w", err)
		}

		setVars := make([]*pb.SetNamespaceVariablesInputVariable, len(variables))
		for i, v := range variables {
			setVars[i] = &pb.SetNamespaceVariablesInputVariable{
				Key:   v.Key,
				Value: v.Value,
			}
		}

		if _, err = tc.grpcClient.NamespaceVariablesClient.SetNamespaceVariables(ctx, &pb.SetNamespaceVariablesRequest{
			NamespacePath: namespacePath,
			Category:      pb.VariableCategory_TERRAFORM,
			Variables:     setVars,
		}); err != nil {
			return nil, nil, fmt.Errorf("failed to set variables: %w", err)
		}

		return nil, &setTerraformVariablesFromFileOutput{
			Message: fmt.Sprintf("Set %d Terraform variables successfully in namespace %s", len(variables), namespacePath),
			Success: true,
			Count:   len(variables),
		}, nil
	}

	return tool, handler
}

// setEnvironmentVariablesFromFileInput is the input for setting environment variables from a file.
type setEnvironmentVariablesFromFileInput struct {
	NamespaceID string `json:"namespace_id" jsonschema:"required,Workspace or group ID or TRN (e.g. Ul8yZ... or trn:workspace:group/workspace or trn:group:group/subgroup)"`
	FilePath    string `json:"file_path" jsonschema:"required,Path to file containing environment variable definitions (KEY=value format)"`
}

// setEnvironmentVariablesFromFileOutput is the output for setting environment variables from a file.
type setEnvironmentVariablesFromFileOutput struct {
	Message string `json:"message" jsonschema:"Success message"`
	Success bool   `json:"success" jsonschema:"Whether operation was successful"`
	Count   int    `json:"count" jsonschema:"Number of variables set"`
}

// SetEnvironmentVariablesFromFile returns an MCP tool for setting multiple environment variables from a file.
func setEnvironmentVariablesFromFile(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*setEnvironmentVariablesFromFileInput, *setEnvironmentVariablesFromFileOutput]) {
	tool := mcp.Tool{
		Name:        "set_environment_variables_from_file",
		Description: "Set multiple environment variables from a file on a workspace or group. Overwrites existing environment variables.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Set Environment Variables From File",
			IdempotentHint:  true,
			DestructiveHint: ptr.Bool(true),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *setEnvironmentVariablesFromFileInput) (*mcp.CallToolResult, *setEnvironmentVariablesFromFileOutput, error) {
		namespacePath, resourceType, err := getNamespacePath(ctx, tc.grpcClient, input.NamespaceID)
		if err != nil {
			return nil, nil, err
		}

		if err = tc.acl.Authorize(ctx, tc.grpcClient, input.NamespaceID, resourceType); err != nil {
			return nil, nil, err
		}

		parser := varparser.NewVariableParser(nil, false)
		variables, err := parser.ParseEnvironmentVariables(&varparser.ParseEnvironmentVariablesInput{
			EnvVarFilePaths: []string{input.FilePath},
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse environment variables file: %w", err)
		}

		setVars := make([]*pb.SetNamespaceVariablesInputVariable, len(variables))
		for i, v := range variables {
			setVars[i] = &pb.SetNamespaceVariablesInputVariable{
				Key:   v.Key,
				Value: v.Value,
			}
		}

		if _, err = tc.grpcClient.NamespaceVariablesClient.SetNamespaceVariables(ctx, &pb.SetNamespaceVariablesRequest{
			NamespacePath: namespacePath,
			Category:      pb.VariableCategory_ENVIRONMENT,
			Variables:     setVars,
		}); err != nil {
			return nil, nil, fmt.Errorf("failed to set environment variables: %w", err)
		}

		return nil, &setEnvironmentVariablesFromFileOutput{
			Message: fmt.Sprintf("Set %d environment variables successfully in namespace %s", len(variables), namespacePath),
			Success: true,
			Count:   len(variables),
		}, nil
	}

	return tool, handler
}

// deleteVariableInput is the input for deleting a variable.
type deleteVariableInput struct {
	NamespaceID string `json:"namespace_id" jsonschema:"required,Workspace or group ID or TRN (e.g. Ul8yZ... or trn:workspace:group/workspace or trn:group:group/subgroup)"`
	Key         string `json:"key" jsonschema:"required,Variable name/key to delete"`
	Category    string `json:"category" jsonschema:"required,Variable category: TERRAFORM or ENVIRONMENT"`
}

// deleteVariableOutput is the output for deleting a variable.
type deleteVariableOutput struct {
	Message string `json:"message" jsonschema:"Success message"`
	Success bool   `json:"success" jsonschema:"Whether operation was successful"`
}

// DeleteVariable returns an MCP tool for deleting a namespace variable.
func deleteVariable(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*deleteVariableInput, *deleteVariableOutput]) {
	tool := mcp.Tool{
		Name:        "delete_variable",
		Description: "Delete a Terraform or environment variable from a workspace or group.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Variable",
			DestructiveHint: ptr.Bool(true),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *deleteVariableInput) (*mcp.CallToolResult, *deleteVariableOutput, error) {
		if _, ok := pb.VariableCategory_value[input.Category]; !ok {
			return nil, nil, fmt.Errorf("invalid category: must be 'TERRAFORM' or 'ENVIRONMENT'")
		}

		namespacePath, resourceType, err := getNamespacePath(ctx, tc.grpcClient, input.NamespaceID)
		if err != nil {
			return nil, nil, err
		}

		if err = tc.acl.Authorize(ctx, tc.grpcClient, input.NamespaceID, resourceType); err != nil {
			return nil, nil, err
		}

		variableID := trn.NewResourceTRN(trn.ResourceTypeVariable, namespacePath, input.Category, input.Key)
		_, err = tc.grpcClient.NamespaceVariablesClient.DeleteNamespaceVariable(ctx, &pb.DeleteNamespaceVariableRequest{Id: variableID})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to delete variable: %w", err)
		}

		return nil, &deleteVariableOutput{
			Message: fmt.Sprintf("Variable '%s' deleted successfully from namespace %s", input.Key, namespacePath),
			Success: true,
		}, nil
	}

	return tool, handler
}

func getNamespacePath(ctx context.Context, c *client.Client, namespaceID string) (string, trn.ResourceType, error) {
	workspace, err := c.WorkspacesClient.GetWorkspaceByID(ctx, &pb.GetWorkspaceByIDRequest{Id: namespaceID})
	if err == nil {
		return workspace.FullPath, trn.ResourceTypeWorkspace, nil
	}

	if status.Code(err) == codes.NotFound {
		group, gErr := c.GroupsClient.GetGroupByID(ctx, &pb.GetGroupByIDRequest{Id: namespaceID})
		if gErr != nil {
			return "", "", fmt.Errorf("failed to get workspace or group: %w", gErr)
		}
		return group.FullPath, trn.ResourceTypeGroup, nil
	}

	return "", "", fmt.Errorf("failed to get namespace: %w", err)
}
