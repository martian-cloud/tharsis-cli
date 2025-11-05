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
)

// ToPath extracts the path portion from a TRN.
// If the input is not a TRN, returns it unchanged.
func ToPath(identifier string) string {
	if !strings.HasPrefix(identifier, "trn:") {
		return identifier
	}
	
	// Extract path from TRN format: trn:type:path
	parts := strings.Split(identifier, ":")
	if len(parts) >= 3 {
		return parts[2]
	}
	
	return identifier
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
