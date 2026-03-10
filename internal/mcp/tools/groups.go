package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/smithy-go/ptr"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// group represents a Tharsis group in MCP responses.
type group struct {
	ID          string `json:"id" jsonschema:"The unique identifier of the group"`
	Path        string `json:"path" jsonschema:"Full path to the group"`
	Name        string `json:"name" jsonschema:"Group name"`
	Description string `json:"description" jsonschema:"Group description"`
	TRN         string `json:"trn" jsonschema:"Tharsis Resource Name"`
}

// toGroup converts a proto group to MCP group.
func toGroup(g *pb.Group) *group {
	return &group{
		ID:          g.Metadata.Id,
		TRN:         g.Metadata.Trn,
		Path:        g.FullPath,
		Name:        g.Name,
		Description: g.Description,
	}
}

// listGroupsInput is the input for listing groups.
type listGroupsInput struct {
	ParentID *string `json:"parent_id,omitempty" jsonschema:"Filter groups to this parent group ID (e.g. Ul8yZ... or trn:group:parent-group)"`
	Search   *string `json:"search,omitempty" jsonschema:"Search term to filter by group path"`
	Sort     *string `json:"sort,omitempty" jsonschema:"Sort order: FULL_PATH_ASC or FULL_PATH_DESC"`
	Limit    *int32  `json:"limit,omitempty" jsonschema:"Maximum number of groups to return (default: 10, max: 50)"`
	Cursor   *string `json:"cursor,omitempty" jsonschema:"Pagination cursor from previous response"`
}

// listGroupsOutput is the output for listing groups.
type listGroupsOutput struct {
	Groups   []*group `json:"groups" jsonschema:"List of groups"`
	PageInfo pageInfo `json:"page_info" jsonschema:"Pagination information"`
}

// ListGroups returns an MCP tool for listing groups.
func listGroups(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*listGroupsInput, *listGroupsOutput]) {
	tool := mcp.Tool{
		Name:        "list_groups",
		Description: "List Tharsis groups with optional filtering by parent. Supports pagination for large result sets.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "List Groups",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *listGroupsInput) (*mcp.CallToolResult, *listGroupsOutput, error) {
		req := &pb.GetGroupsRequest{
			PaginationOptions: buildPaginationOptions(input.Limit, input.Cursor),
			ParentId:          input.ParentID,
			Search:            input.Search,
			Sort:              toSortEnum[pb.GroupSortableField](input.Sort, pb.GroupSortableField_value),
		}

		resp, err := tc.grpcClient.GroupsClient.GetGroups(ctx, req)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list groups: %w", err)
		}

		groups := make([]*group, len(resp.Groups))
		for i, g := range resp.Groups {
			groups[i] = toGroup(g)
		}

		return nil, &listGroupsOutput{
			Groups:   groups,
			PageInfo: buildPageInfo(resp.PageInfo),
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
	Group *group `json:"group,omitempty" jsonschema:"The group details"`
}

// GetGroup returns an MCP tool for getting a group.
func getGroup(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*getGroupInput, *getGroupOutput]) {
	tool := mcp.Tool{
		Name:        "get_group",
		Description: "Retrieve details about a Tharsis group.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Group",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *getGroupInput) (*mcp.CallToolResult, *getGroupOutput, error) {
		resp, err := tc.grpcClient.GroupsClient.GetGroupByID(ctx, &pb.GetGroupByIDRequest{Id: input.ID})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get group: %w", err)
		}

		return nil, &getGroupOutput{
			Group: toGroup(resp),
		}, nil
	}

	return tool, handler
}

// createGroupInput is the input for creating a group.
type createGroupInput struct {
	Name        string   `json:"name" jsonschema:"required,Name of the group"`
	ParentID    string   `json:"parent_id" jsonschema:"required,ID of the parent group (e.g. Ul8yZ... or trn:group:parent-group)"`
	Description string   `json:"description,omitempty" jsonschema:"Description of the group"`
	RunnerTags  []string `json:"runner_tags,omitempty" jsonschema:"Runner tags for the group"`
}

// createGroupOutput is the output for creating a group.
type createGroupOutput struct {
	Group *group `json:"group,omitempty" jsonschema:"The created group"`
}

// CreateGroup returns an MCP tool for creating a group.
func createGroup(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*createGroupInput, *createGroupOutput]) {
	tool := mcp.Tool{
		Name:        "create_group",
		Description: "Create a new Tharsis group under a parent group. Groups organize workspaces and can be nested hierarchically. Note: Cannot create root-level groups.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Create Group",
			DestructiveHint: ptr.Bool(false),
			IdempotentHint:  true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *createGroupInput) (*mcp.CallToolResult, *createGroupOutput, error) {
		if err := tc.acl.Authorize(ctx, tc.grpcClient, input.ParentID, trn.ResourceTypeGroup); err != nil {
			return nil, nil, err
		}

		resp, err := tc.grpcClient.GroupsClient.CreateGroup(ctx, &pb.CreateGroupRequest{
			Name:        input.Name,
			ParentId:    &input.ParentID,
			Description: input.Description,
			RunnerTags:  input.RunnerTags,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create group: %w", err)
		}

		return nil, &createGroupOutput{
			Group: toGroup(resp),
		}, nil
	}

	return tool, handler
}

// updateGroupInput is the input for updating a group.
type updateGroupInput struct {
	ID          string   `json:"id" jsonschema:"required,Group ID or TRN (e.g. Ul8yZ... or trn:group:parent-group/group-name)"`
	Description *string  `json:"description,omitempty" jsonschema:"New description for the group"`
	RunnerTags  []string `json:"runner_tags,omitempty" jsonschema:"New runner tags for the group"`
}

// updateGroupOutput is the output for updating a group.
type updateGroupOutput struct {
	Group *group `json:"group,omitempty" jsonschema:"The updated group"`
}

// UpdateGroup returns an MCP tool for updating a group.
func updateGroup(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*updateGroupInput, *updateGroupOutput]) {
	tool := mcp.Tool{
		Name:        "update_group",
		Description: "Update an existing Tharsis group's description.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Update Group",
			DestructiveHint: ptr.Bool(false),
			IdempotentHint:  true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *updateGroupInput) (*mcp.CallToolResult, *updateGroupOutput, error) {
		if err := tc.acl.Authorize(ctx, tc.grpcClient, input.ID, trn.ResourceTypeGroup); err != nil {
			return nil, nil, err
		}

		resp, err := tc.grpcClient.GroupsClient.UpdateGroup(ctx, &pb.UpdateGroupRequest{
			Id:          input.ID,
			Description: input.Description,
			RunnerTags:  input.RunnerTags,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to update group: %w", err)
		}

		return nil, &updateGroupOutput{
			Group: toGroup(resp),
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
func deleteGroup(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*deleteGroupInput, *deleteGroupOutput]) {
	tool := mcp.Tool{
		Name:        "delete_group",
		Description: "Delete a Tharsis group. Use with caution as this operation is irreversible.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Group",
			DestructiveHint: ptr.Bool(true),
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *deleteGroupInput) (*mcp.CallToolResult, *deleteGroupOutput, error) {
		grp, err := tc.grpcClient.GroupsClient.GetGroupByID(ctx, &pb.GetGroupByIDRequest{Id: input.ID})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get group: %w", err)
		}

		if !strings.Contains(grp.FullPath, "/") {
			return nil, nil, fmt.Errorf("cannot delete top-level group %q: top-level groups cannot be deleted via MCP for safety", grp.FullPath)
		}

		if err := tc.acl.Authorize(ctx, tc.grpcClient, input.ID, trn.ResourceTypeGroup); err != nil {
			return nil, nil, err
		}

		if _, err := tc.grpcClient.GroupsClient.DeleteGroup(ctx, &pb.DeleteGroupRequest{Id: input.ID}); err != nil {
			return nil, nil, fmt.Errorf("failed to delete group: %w", err)
		}

		return nil, &deleteGroupOutput{
			Message: fmt.Sprintf("Group %s deleted successfully", input.ID),
			Success: true,
		}, nil
	}

	return tool, handler
}
