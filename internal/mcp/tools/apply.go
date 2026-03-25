package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// apply represents a Terraform apply in MCP responses.
type apply struct {
	ID           string  `json:"id" jsonschema:"Unique identifier for this apply"`
	TRN          string  `json:"trn" jsonschema:"Tharsis Resource Name"`
	Status       string  `json:"status" jsonschema:"Current status: created, queued, pending, running, finished, errored, or canceled"`
	TriggeredBy  string  `json:"triggered_by" jsonschema:"Username or service account that triggered this apply"`
	ErrorMessage *string `json:"error_message,omitempty" jsonschema:"Error details if the apply failed"`
}

// toApply converts a proto apply to MCP apply.
func toApply(a *pb.Apply) *apply {
	return &apply{
		ID:           a.Metadata.Id,
		TRN:          a.Metadata.Trn,
		Status:       a.Status,
		TriggeredBy:  a.TriggeredBy,
		ErrorMessage: a.ErrorMessage,
	}
}

// getApplyInput is the input for getting an apply.
type getApplyInput struct {
	ID string `json:"id" jsonschema:"required,Apply ID or TRN (e.g. Ul8yZ... or trn:apply:...)"`
}

// getApplyOutput is the output for getting an apply.
type getApplyOutput struct {
	Apply *apply `json:"apply,omitempty" jsonschema:"The apply details"`
}

// GetApply returns an MCP tool for getting an apply.
func getApply(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*getApplyInput, *getApplyOutput]) {
	tool := mcp.Tool{
		Name:        "get_apply",
		Description: "Get details of a Terraform apply by ID.",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Apply",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, input *getApplyInput) (*mcp.CallToolResult, *getApplyOutput, error) {
		resp, err := tc.grpcClient.RunsClient.GetApplyByID(ctx, &pb.GetApplyByIDRequest{Id: input.ID})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get apply: %w", err)
		}

		return nil, &getApplyOutput{
			Apply: toApply(resp),
		}, nil
	}

	return tool, handler
}
