// Package trn provides utilities for working with Tharsis Resource Names (TRNs).
package trn

import (
	"fmt"
	"strings"
)

// ResourceType represents the type of resource for TRN construction.
type ResourceType string

const (
	// ResourceTypeWorkspace is the resource type constant for workspace TRNs.
	ResourceTypeWorkspace ResourceType = "workspace"
	// ResourceTypeGroup is the resource type constant for group TRNs.
	ResourceTypeGroup ResourceType = "group"
	// ResourceTypeManagedIdentity is the resource type constant for managed identity TRNs.
	ResourceTypeManagedIdentity ResourceType = "managed_identity"
	// ResourceTypeVariable is the resource type constant for variable TRNs.
	ResourceTypeVariable ResourceType = "variable"
	// ResourceTypeRun is the resource type constant for run TRNs.
	ResourceTypeRun ResourceType = "run"
	// ResourceTypeConfigurationVersion is the resource type constant for configuration version TRNs.
	ResourceTypeConfigurationVersion ResourceType = "configuration_version"
	// ResourceTypeTerraformModule is the resource type constant for Terraform module TRNs.
	ResourceTypeTerraformModule ResourceType = "terraform_module"
	// ResourceTypeTerraformModuleVersion is the resource type constant for Terraform module version TRNs.
	ResourceTypeTerraformModuleVersion ResourceType = "terraform_module_version"
	// ResourceTypeTerraformProvider is the resource type constant for Terraform provider TRNs.
	ResourceTypeTerraformProvider ResourceType = "terraform_provider"
	// ResourceTypeTerraformProviderPlatform is the resource type constant for Terraform provider platform TRNs.
	ResourceTypeTerraformProviderPlatform ResourceType = "terraform_provider_platform"
	// ResourceTypeFederatedRegistry is the resource type constant for federated registry TRNs.
	ResourceTypeFederatedRegistry ResourceType = "federated_registry"
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
	return strings.HasPrefix(identifier, "trn:")
}

// ToTRN converts a path or TRN to a TRN format.
// If the input is already a TRN, it returns it unchanged.
// If the input is a path, it converts it to a TRN with the specified resource type.
func ToTRN(identifier string, resourceType ResourceType) string {
	// If it's already a TRN, return as-is
	if strings.HasPrefix(identifier, "trn:") {
		return identifier
	}

	// Convert path to TRN
	return NewResourceTRN(resourceType, identifier)
}

// NewResourceTRN returns a new TRN string for the given resource and
// arguments. This is a helper function for creating TRNs.
func NewResourceTRN(resource ResourceType, a ...string) string {
	return fmt.Sprintf("trn:%s:%s", resource, strings.Join(a, "/"))
}
