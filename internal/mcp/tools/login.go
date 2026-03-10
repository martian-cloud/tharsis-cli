package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/auth"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ssoLoginInput is the input for the SSO login tool.
type ssoLoginInput struct{}

// ssoLoginOutput is the output for the SSO login tool.
type ssoLoginOutput struct {
	Message string `json:"message" jsonschema:"SSO Login status message"`
	Success bool   `json:"success" jsonschema:"Whether SSO login was successful"`
}

// LoginWithSSO is a tools that authenticates against a Tharsis instance with browser-based SSO.
func loginWithSSO(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*ssoLoginInput, *ssoLoginOutput]) {
	tool := mcp.Tool{
		Name:        "login",
		Description: "Authenticates with Tharsis using SSO by opening a browser for OAuth flow and storing the authentication token.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, _ *ssoLoginInput) (*mcp.CallToolResult, *ssoLoginOutput, error) {
		ssoClient, err := auth.NewSSOClient(tc.tharsisURL, auth.WithGRPCClient(tc.grpcClient))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create SSO client: %w", err)
		}

		token, err := ssoClient.PerformLogin(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("login failed: %w", err)
		}

		if err := ssoClient.StoreToken(token); err != nil {
			return nil, nil, fmt.Errorf("failed to store token: %w", err)
		}

		return nil, &ssoLoginOutput{Message: "Successfully logged in to Tharsis", Success: true}, nil
	}

	return tool, handler
}

// getConnectionInfoInput is the input for the get_connection_info tool.
type getConnectionInfoInput struct{}

// getConnectionInfoOutput is the output for the get_connection_info tool.
type getConnectionInfoOutput struct {
	TharsisURL    string  `json:"tharsis_url" jsonschema:"The URL of the connected Tharsis instance"`
	ProfileName   string  `json:"profile_name" jsonschema:"The name of the active Tharsis profile"`
	Authenticated bool    `json:"authenticated" jsonschema:"Whether the user is currently authenticated"`
	TRN           *string `json:"trn,omitempty" jsonschema:"The TRN of the authenticated caller"`
}

// getConnectionInfo returns an MCP tool for retrieving connection information.
func getConnectionInfo(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[*getConnectionInfoInput, *getConnectionInfoOutput]) {
	tool := mcp.Tool{
		Name:        "get_connection_info",
		Description: "Get information about the current Tharsis connection including URL, profile, authentication status, and caller details (User or ServiceAccount).",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Connection Info",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, _ *getConnectionInfoInput) (*mcp.CallToolResult, *getConnectionInfoOutput, error) {
		output := &getConnectionInfoOutput{
			TharsisURL:    tc.tharsisURL,
			ProfileName:   tc.profileName,
			Authenticated: false,
		}

		resp, err := tc.grpcClient.CallerClient.GetCaller(ctx, &emptypb.Empty{})
		if err != nil {
			return nil, output, nil
		}

		output.Authenticated = true

		switch caller := resp.Caller.(type) {
		case *pb.GetCallerResponse_User:
			output.TRN = &caller.User.Metadata.Trn
		case *pb.GetCallerResponse_ServiceAccount:
			output.TRN = &caller.ServiceAccount.Metadata.Trn
		}

		return nil, output, nil
	}

	return tool, handler
}
