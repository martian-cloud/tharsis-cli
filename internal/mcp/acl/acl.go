// Package acl provides access control for MCP tools.
//
// Patterns support simple wildcard matching (case-insensitive):
//   - "prod" - exact match for "prod"
//   - "prod/*" - matches any path starting with "prod/" (all levels)
//   - "prod/**" - same as "prod/*" (matches all levels)
//   - "prod/team-*" - matches paths starting with "prod/team-"
//
// Restrictions:
//   - Wildcard-only patterns ("*") are not allowed
//   - Patterns cannot start with a wildcard ("*/...")
package acl

//go:generate go tool mockery --name Checker --inpackage --case underscore

import (
	"context"
	"fmt"
	"strings"

	"github.com/ryanuber/go-glob"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

const maxPatternLength = 512

// Checker validates access to namespaces.
type Checker interface {
	Authorize(ctx context.Context, client *client.Client, identifier string, resType trn.ResourceType) error
}

type checker struct {
	patterns    []string
	hasPatterns bool
	cache       map[string]error
}

// NewChecker creates an ACL checker from a comma-separated pattern string.
func NewChecker(patternStr string) (Checker, error) {
	patterns, err := parsePatterns(patternStr)
	if err != nil {
		return nil, err
	}

	return &checker{
		patterns:    patterns,
		hasPatterns: len(patterns) > 0,
		cache:       make(map[string]error),
	}, nil
}

func (a *checker) Authorize(ctx context.Context, client *client.Client, identifier string, resType trn.ResourceType) error {
	if !a.hasPatterns {
		return nil
	}

	if err, ok := a.cache[identifier]; ok {
		return err
	}

	path, err := a.resolveNamespacePath(ctx, client, identifier, resType)
	if err != nil {
		return err
	}

	err = a.matchPath(path)
	a.cache[identifier] = err
	return err
}

func (a *checker) resolveNamespacePath(ctx context.Context, client *client.Client, identifier string, resType trn.ResourceType) (string, error) {
	switch resType {
	case trn.ResourceTypeWorkspace:
		resp, err := client.WorkspacesClient.GetWorkspaceByID(ctx, &pb.GetWorkspaceByIDRequest{Id: identifier})
		if err != nil {
			return "", fmt.Errorf("failed to resolve workspace path: %w", err)
		}
		return resp.FullPath, nil

	case trn.ResourceTypeGroup:
		resp, err := client.GroupsClient.GetGroupByID(ctx, &pb.GetGroupByIDRequest{Id: identifier})
		if err != nil {
			return "", fmt.Errorf("failed to resolve group path: %w", err)
		}
		return resp.FullPath, nil

	case trn.ResourceTypeRun:
		resp, err := client.RunsClient.GetRunByID(ctx, &pb.GetRunByIDRequest{Id: identifier})
		if err != nil {
			return "", fmt.Errorf("failed to resolve run: %w", err)
		}
		pathParts := trn.ToPathParts(resp.Metadata.Trn)
		return strings.Join(pathParts[:len(pathParts)-1], "/"), nil

	case trn.ResourceTypeConfigurationVersion:
		resp, err := client.ConfigurationVersionsClient.GetConfigurationVersionByID(ctx, &pb.GetConfigurationVersionByIDRequest{Id: identifier})
		if err != nil {
			return "", fmt.Errorf("failed to resolve configuration version: %w", err)
		}
		pathParts := trn.ToPathParts(resp.Metadata.Trn)
		return strings.Join(pathParts[:len(pathParts)-1], "/"), nil

	case trn.ResourceTypeTerraformModule:
		resp, err := client.TerraformModulesClient.GetTerraformModuleByID(ctx, &pb.GetTerraformModuleByIDRequest{Id: identifier})
		if err != nil {
			return "", fmt.Errorf("failed to resolve terraform module: %w", err)
		}
		// TRN format: trn:terraform_module:group/name/system
		pathParts := trn.ToPathParts(resp.Metadata.Trn)
		return strings.Join(pathParts[:len(pathParts)-2], "/"), nil

	default:
		return "", fmt.Errorf("unsupported resource type for ACL: %s", resType)
	}
}

func (a *checker) matchPath(path string) error {
	path = strings.ToLower(path)

	for _, pattern := range a.patterns {
		if pattern == path || glob.Glob(pattern, path) {
			return nil
		}
	}

	return fmt.Errorf("access denied: path %q does not match allowed patterns", path)
}

func parsePatterns(patternStr string) ([]string, error) {
	if patternStr == "" {
		return nil, nil
	}

	seen := make(map[string]bool)
	var parsed []string

	for p := range strings.SplitSeq(patternStr, ",") {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}

		if len(p) > maxPatternLength {
			return nil, fmt.Errorf("invalid pattern: pattern exceeds maximum length of %d characters", maxPatternLength)
		}

		if p == "*" {
			return nil, fmt.Errorf("invalid pattern %q: wildcard-only patterns are not allowed", p)
		}

		if strings.HasPrefix(p, "*") {
			return nil, fmt.Errorf("invalid pattern %q: patterns cannot start with a wildcard", p)
		}

		if strings.HasPrefix(p, "/") || strings.HasSuffix(p, "/") {
			return nil, fmt.Errorf("invalid pattern %q: patterns should not start or end with '/'", p)
		}

		if strings.Contains(p, "//") {
			return nil, fmt.Errorf("invalid pattern %q: patterns should not contain '//'", p)
		}

		p = strings.ToLower(p)
		seen[p] = true
		parsed = append(parsed, p)
	}

	return parsed, nil
}
