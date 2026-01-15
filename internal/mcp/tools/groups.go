package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/smithy-go/ptr"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// group represents a Tharsis group in MCP responses.
type group struct {
	ID          string `json:"id" jsonschema:"The unique identifier of the group"`
	Path        string `json:"path" jsonschema:"Full path to the group"`
	Name        string `json:"name" jsonschema:"Group name"`
	Description string `json:"description" jsonschema:"Group description"`
	TRN         string `json:"trn" jsonschema:"Tharsis Resource Name"`
}

// toGroup converts an SDK group to MCP group.
func toGroup(g *sdktypes.Group) group {
	return group{
		ID:          g.Metadata.ID,
		TRN:         g.Metadata.TRN,
		Path:        g.FullPath,
		Name:        g.Name,
		Description: g.Description,
	}
}

// listGroupsInput is the input for listing groups.
type listGroupsInput struct {
	ParentPath *string                      `json:"parent_path,omitempty" jsonschema:"Filter groups to this parent path (e.g. parent-group)"`
	Sort       *sdktypes.GroupSortableField `json:"sort,omitempty" jsonschema:"Sort order: FULL_PATH_ASC or FULL_PATH_DESC"`
	Limit      *int32                       `json:"limit,omitempty" jsonschema:"Maximum number of groups to return (default: 10, max: 50)"`
	Cursor     *string                      `json:"cursor,omitempty" jsonschema:"Pagination cursor from previous response"`
}

// listGroupsOutput is the output for listing groups.
type listGroupsOutput struct {
	Groups   []group  `json:"groups" jsonschema:"List of groups"`
	PageInfo pageInfo `json:"page_info" jsonschema:"Pagination information"`
}

// ListGroups returns an MCP tool for listing groups.
func listGroups(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[listGroupsInput, listGroupsOutput]) {
	tool := mcp.Tool{
		Name:        "list_groups",
		Description: "List Tharsis groups with optional filtering by parent. Supports pagination for large result sets.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Groups",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input listGroupsInput) (*mcp.CallToolResult, listGroupsOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, listGroupsOutput{}, err
		}

		sdkInput := &sdktypes.GetGroupsInput{
			PaginationOptions: buildPaginationOptions(input.Limit, input.Cursor),
			Sort:              input.Sort,
		}

		if input.ParentPath != nil {
			sdkInput.Filter = &sdktypes.GroupFilter{
				ParentPath: input.ParentPath,
			}
		}

		result, err := client.Groups().GetGroups(ctx, sdkInput)
		if err != nil {
			return nil, listGroupsOutput{}, fmt.Errorf("failed to list groups: %w", err)
		}

		groups := make([]group, len(result.Groups))
		for i, g := range result.Groups {
			groups[i] = toGroup(&g)
		}

		return nil, listGroupsOutput{
			Groups:   groups,
			PageInfo: buildPageInfo(result.PageInfo),
		}, nil
	}

	return tool, handler
}

// getGroupInput is the input for getting a group.
type getGroupInput struct {
	ID string `json:"id" jsonschema:"required,Group ID or TRN (e.g. Ul8yZ... or trn:group:parent-group/group-name)"`
}

// getGroupOutput is the output for getting a group.
type getGroupOutput struct {
	Group group `json:"group" jsonschema:"The group details"`
}

// GetGroup returns an MCP tool for getting a group.
func getGroup(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[getGroupInput, getGroupOutput]) {
	tool := mcp.Tool{
		Name:        "get_group",
		Description: "Retrieve details about a Tharsis group.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Group",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input getGroupInput) (*mcp.CallToolResult, getGroupOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, getGroupOutput{}, err
		}

		group, err := client.Groups().GetGroup(ctx, &sdktypes.GetGroupInput{ID: &input.ID})
		if err != nil {
			return nil, getGroupOutput{}, fmt.Errorf("failed to get group: %w", err)
		}

		return nil, getGroupOutput{
			Group: toGroup(group),
		}, nil
	}

	return tool, handler
}

// createGroupInput is the input for creating a group.
type createGroupInput struct {
	Name        string `json:"name" jsonschema:"required,Name of the group"`
	ParentPath  string `json:"parent_path" jsonschema:"required,Path to the parent group"`
	Description string `json:"description,omitempty" jsonschema:"Description of the group"`
}

// createGroupOutput is the output for creating a group.
type createGroupOutput struct {
	Group group `json:"group" jsonschema:"The created group"`
}

// CreateGroup returns an MCP tool for creating a group.
func createGroup(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[createGroupInput, createGroupOutput]) {
	tool := mcp.Tool{
		Name:        "create_group",
		Description: "Create a new Tharsis group under a parent group. Groups organize workspaces and can be nested hierarchically. Note: Cannot create root-level groups.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Create Group",
			DestructiveHint: ptr.Bool(false),
			IdempotentHint:  true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input createGroupInput) (*mcp.CallToolResult, createGroupOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, createGroupOutput{}, err
		}

		if err = tc.acl.Authorize(ctx, client, trn.ToTRN(input.ParentPath, trn.ResourceTypeGroup), trn.ResourceTypeGroup); err != nil {
			return nil, createGroupOutput{}, err
		}

		createInput := &sdktypes.CreateGroupInput{
			Name:        input.Name,
			ParentPath:  &input.ParentPath,
			Description: input.Description,
		}

		group, err := client.Groups().CreateGroup(ctx, createInput)
		if err != nil {
			return nil, createGroupOutput{}, fmt.Errorf("failed to create group: %w", err)
		}

		return nil, createGroupOutput{
			Group: toGroup(group),
		}, nil
	}

	return tool, handler
}

// updateGroupInput is the input for updating a group.
type updateGroupInput struct {
	ID          string `json:"id" jsonschema:"required,Group ID or TRN (e.g. Ul8yZ... or trn:group:parent-group/group-name)"`
	Description string `json:"description,omitempty" jsonschema:"New description for the group"`
}

// updateGroupOutput is the output for updating a group.
type updateGroupOutput struct {
	Group group `json:"group" jsonschema:"The updated group"`
}

// UpdateGroup returns an MCP tool for updating a group.
func updateGroup(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[updateGroupInput, updateGroupOutput]) {
	tool := mcp.Tool{
		Name:        "update_group",
		Description: "Update an existing Tharsis group's description.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Update Group",
			DestructiveHint: ptr.Bool(false),
			IdempotentHint:  true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input updateGroupInput) (*mcp.CallToolResult, updateGroupOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, updateGroupOutput{}, err
		}

		if err = tc.acl.Authorize(ctx, client, input.ID, trn.ResourceTypeGroup); err != nil {
			return nil, updateGroupOutput{}, err
		}

		updateInput := &sdktypes.UpdateGroupInput{
			ID:          &input.ID,
			Description: input.Description,
		}

		updatedGroup, err := client.Groups().UpdateGroup(ctx, updateInput)
		if err != nil {
			return nil, updateGroupOutput{}, fmt.Errorf("failed to update group: %w", err)
		}

		return nil, updateGroupOutput{
			Group: toGroup(updatedGroup),
		}, nil
	}

	return tool, handler
}

// deleteGroupInput is the input for deleting a group.
// Note: Force deletion is intentionally not supported to prevent accidental
// recursive deletion of all child resources. Child resources must be deleted explicitly.
type deleteGroupInput struct {
	ID string `json:"id" jsonschema:"required,Group ID or TRN (e.g. Ul8yZ... or trn:group:parent-group/group-name)"`
}

// deleteGroupOutput is the output for deleting a group.
type deleteGroupOutput struct {
	Message string `json:"message" jsonschema:"Deletion confirmation message"`
	Success bool   `json:"success" jsonschema:"Whether deletion was successful"`
}

// DeleteGroup returns an MCP tool for deleting a group.
func deleteGroup(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[deleteGroupInput, deleteGroupOutput]) {
	tool := mcp.Tool{
		Name:        "delete_group",
		Description: "Delete a Tharsis group. Use with caution as this operation is irreversible.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Group",
			DestructiveHint: ptr.Bool(true),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input deleteGroupInput) (*mcp.CallToolResult, deleteGroupOutput, error) {
		client, err := tc.clientGetter()
		if err != nil {
			return nil, deleteGroupOutput{}, err
		}

		// Get the group to check if it's a top-level group
		group, err := client.Groups().GetGroup(ctx, &sdktypes.GetGroupInput{ID: &input.ID})
		if err != nil {
			return nil, deleteGroupOutput{}, fmt.Errorf("failed to get group: %w", err)
		}

		// Prevent deletion of top-level groups since they're rarely deleted and this protects against accidental LLM actions.
		if !strings.Contains(group.FullPath, "/") {
			return nil, deleteGroupOutput{}, fmt.Errorf("cannot delete top-level group %q: top-level groups cannot be deleted via MCP for safety", group.FullPath)
		}

		if err = tc.acl.Authorize(ctx, client, input.ID, trn.ResourceTypeGroup); err != nil {
			return nil, deleteGroupOutput{}, err
		}

		deleteInput := &sdktypes.DeleteGroupInput{
			ID: &input.ID,
		}

		if err := client.Groups().DeleteGroup(ctx, deleteInput); err != nil {
			return nil, deleteGroupOutput{}, fmt.Errorf("failed to delete group: %w", err)
		}

		return nil, deleteGroupOutput{
			Message: fmt.Sprintf("Group %s deleted successfully", input.ID),
			Success: true,
		}, nil
	}

	return tool, handler
}
