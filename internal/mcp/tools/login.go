package tools

import (
	"context"
	"fmt"

	"github.com/aws/smithy-go/ptr"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/auth"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// ssoLoginInput is the input for the SSO login tool.
type ssoLoginInput struct{}

// ssoLoginOutput is the output for the SSO login tool.
type ssoLoginOutput struct {
	Message string `json:"message" jsonschema:"SSO Login status message"`
	Success bool   `json:"success" jsonschema:"Whether SSO login was successful"`
}

// LoginWithSSO is a tools that authenticates against a Tharsis instance with browser-based SSO.
func loginWithSSO(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[ssoLoginInput, ssoLoginOutput]) {
	tool := mcp.Tool{
		Name:        "login",
		Description: "Authenticates with Tharsis using SSO by opening a browser for OAuth flow and storing the authentication token.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, _ ssoLoginInput) (*mcp.CallToolResult, ssoLoginOutput, error) {
		ssoClient, err := auth.NewSSOClient(tc.tharsisURL)
		if err != nil {
			return nil, ssoLoginOutput{}, fmt.Errorf("failed to create SSO client: %w", err)
		}

		token, err := ssoClient.PerformLogin(ctx)
		if err != nil {
			return nil, ssoLoginOutput{}, fmt.Errorf("login failed: %w", err)
		}

		if err := ssoClient.StoreToken(token); err != nil {
			return nil, ssoLoginOutput{}, fmt.Errorf("failed to store token: %w", err)
		}

		return nil, ssoLoginOutput{Message: "Successfully logged in to Tharsis", Success: true}, nil
	}

	return tool, handler
}

// getConnectionInfoInput is the input for the get_connection_info tool.
type getConnectionInfoInput struct{}

// getConnectionInfoOutput is the output for the get_connection_info tool.
type getConnectionInfoOutput struct {
	TharsisURL          string  `json:"tharsis_url" jsonschema:"The URL of the connected Tharsis instance"`
	ProfileName         string  `json:"profile_name" jsonschema:"The name of the active Tharsis profile"`
	Authenticated       bool    `json:"authenticated" jsonschema:"Whether the user is currently authenticated"`
	Username            *string `json:"username,omitempty" jsonschema:"The username of the authenticated user (if User)"`
	Email               *string `json:"email,omitempty" jsonschema:"The email of the authenticated user (if User)"`
	ServiceAccountName  *string `json:"service_account_name,omitempty" jsonschema:"The name of the service account (if ServiceAccount)"`
	ServiceAccountGroup *string `json:"service_account_group,omitempty" jsonschema:"The group path of the service account (if ServiceAccount)"`
	CallerType          *string `json:"caller_type,omitempty" jsonschema:"The type of authenticated caller (User or ServiceAccount)"`
}

// getConnectionInfo returns an MCP tool for retrieving connection information.
func getConnectionInfo(tc *ToolContext) (mcp.Tool, mcp.ToolHandlerFor[getConnectionInfoInput, getConnectionInfoOutput]) {
	tool := mcp.Tool{
		Name:        "get_connection_info",
		Description: "Get information about the current Tharsis connection including URL, profile, authentication status, and caller details (User or ServiceAccount).",
		Annotations: &mcp.ToolAnnotations{
			Title:        "Get Connection Info",
			ReadOnlyHint: true,
		},
	}

	handler := func(ctx context.Context, _ *mcp.CallToolRequest, _ getConnectionInfoInput) (*mcp.CallToolResult, getConnectionInfoOutput, error) {
		output := getConnectionInfoOutput{
			TharsisURL:    tc.tharsisURL,
			ProfileName:   tc.profileName,
			Authenticated: false,
		}

		// Try to get caller info
		client, err := tc.clientGetter()
		if err != nil {
			return nil, output, nil
		}

		caller, err := client.Me().GetCallerInfo(ctx)
		if err != nil {
			return nil, output, nil
		}

		output.Authenticated = true

		switch v := caller.(type) {
		case *types.User:
			output.CallerType = ptr.String("User")
			output.Username = &v.Username
			output.Email = &v.Email
		case *types.ServiceAccount:
			output.CallerType = ptr.String("ServiceAccount")
			output.ServiceAccountName = &v.Name
			output.ServiceAccountGroup = &v.GroupPath
		}

		return nil, output, nil
	}

	return tool, handler
}
