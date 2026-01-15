// Package tools provides MCP tool implementations for Tharsis CLI.
package tools

import (
	"fmt"
	"net/http"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/acl"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tharsis"
	sdk "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
)

// ToolContext holds dependencies for tool execution.
type ToolContext struct {
	tharsisURL   string
	profileName  string
	clientGetter func() (tharsis.Client, error)
	acl          acl.Checker
	httpClient   *http.Client
}

// ToolContextOption configures a ToolContext.
type ToolContextOption func(*ToolContext) error

// WithACLPatterns sets ACL patterns for the tool context.
func WithACLPatterns(patterns string) ToolContextOption {
	return func(tc *ToolContext) error {
		checker, err := acl.NewChecker(patterns)
		if err != nil {
			return fmt.Errorf("failed to initialize ACL checker: %w", err)
		}
		tc.acl = checker
		return nil
	}
}

// NewToolContext creates a new tool context.
func NewToolContext(tharsisURL, profileName string, httpClient *http.Client, clientGetter func() (*sdk.Client, error), opts ...ToolContextOption) (*ToolContext, error) {
	wrappedGetter := func() (tharsis.Client, error) {
		c, err := clientGetter()
		if err != nil {
			return nil, err
		}
		return tharsis.NewClient(c), nil
	}

	tc := &ToolContext{
		tharsisURL:   tharsisURL,
		profileName:  profileName,
		clientGetter: wrappedGetter,
		httpClient:   httpClient,
	}

	for _, opt := range opts {
		if err := opt(tc); err != nil {
			return nil, err
		}
	}

	if tc.acl == nil {
		checker, err := acl.NewChecker("")
		if err != nil {
			return nil, fmt.Errorf("failed to initialize ACL checker: %w", err)
		}
		tc.acl = checker
	}

	return tc, nil
}
