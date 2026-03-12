// Package trn provides utilities for working with Tharsis Resource Names (TRNs).
package trn

import (
	"fmt"
	"strings"
)

// ResourceType represents the type of resource for TRN construction.
type ResourceType string

// ResourceType constants.
const (
	ResourceTypeWorkspace                      ResourceType = "workspace"
	ResourceTypeGroup                          ResourceType = "group"
	ResourceTypeManagedIdentity                ResourceType = "managed_identity"
	ResourceTypeVariable                       ResourceType = "variable"
	ResourceTypeRun                            ResourceType = "run"
	ResourceTypeConfigurationVersion           ResourceType = "configuration_version"
	ResourceTypeTerraformModule                ResourceType = "terraform_module"
	ResourceTypeTerraformModuleVersion         ResourceType = "terraform_module_version"
	ResourceTypeTerraformProvider              ResourceType = "terraform_provider"
	ResourceTypeTerraformProviderVersionMirror ResourceType = "terraform_provider_version_mirror"
	ResourceTypeTerraformProviderPlatform      ResourceType = "terraform_provider_platform"
	ResourceTypeFederatedRegistry              ResourceType = "federated_registry"
	ResourceTypeServiceAccount                 ResourceType = "service_account"
	ResourceTypeTeam                           ResourceType = "team"
	ResourceTypeUser                           ResourceType = "user"
	ResourceTypeRole                           ResourceType = "role"
	ResourceTypeRunner                         ResourceType = "runner"
)

const (
	// TRNPrefix is the prefix for a Tharsis resource name.
	TRNPrefix = "trn:"
)

// ToPath extracts the path portion from a TRN.
// If the input is not a TRN, returns it unchanged.
func ToPath(identifier string) string {
	if !IsTRN(identifier) {
		return identifier
	}

	// Extract path from TRN format: trn:type:path
	parts := strings.Split(identifier, ":")
	if len(parts) >= 3 {
		return parts[2]
	}

	return identifier
}

// ToPathParts extracts the path portion from a TRN and splits it into parts.
// If the input is not a TRN, splits it as-is.
func ToPathParts(identifier string) []string {
	path := ToPath(identifier)
	return strings.Split(path, "/")
}

// IsTRN returns true if the identifier is a valid TRN format.
func IsTRN(identifier string) bool {
	return strings.HasPrefix(identifier, TRNPrefix)
}

// ToTRN converts a path or TRN to a TRN format.
// If the input is already a TRN, it returns it unchanged.
// If the input is a path, it converts it to a TRN with the specified resource type.
func ToTRN(resourceType ResourceType, identifier string) string {
	// If it's already a TRN, return as-is
	if IsTRN(identifier) {
		return identifier
	}

	// Convert path to TRN
	return NewResourceTRN(resourceType, identifier)
}

// NewResourceTRN returns a new TRN string for the given resource and
// arguments. This is a helper function for creating TRNs.
func NewResourceTRN(resource ResourceType, a ...string) string {
	return fmt.Sprintf("%s%s:%s", TRNPrefix, resource, strings.Join(a, "/"))
}
