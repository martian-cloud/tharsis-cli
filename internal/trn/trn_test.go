package trn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToPath(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		expected   string
	}{
		{
			name:       "trn with path",
			identifier: "trn:workspace:group/my-workspace",
			expected:   "group/my-workspace",
		},
		{
			name:       "trn with nested path",
			identifier: "trn:group:parent/child/grandchild",
			expected:   "parent/child/grandchild",
		},
		{
			name:       "non-trn returns unchanged",
			identifier: "group/my-workspace",
			expected:   "group/my-workspace",
		},
		{
			name:       "empty string",
			identifier: "",
			expected:   "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, ToPath(test.identifier))
		})
	}
}

func TestToPathParts(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		expected   []string
	}{
		{
			name:       "trn with path",
			identifier: "trn:workspace:group/my-workspace",
			expected:   []string{"group", "my-workspace"},
		},
		{
			name:       "non-trn path",
			identifier: "parent/child",
			expected:   []string{"parent", "child"},
		},
		{
			name:       "single segment",
			identifier: "top-level",
			expected:   []string{"top-level"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, ToPathParts(test.identifier))
		})
	}
}

func TestIsTRN(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		expected   bool
	}{
		{
			name:       "valid trn",
			identifier: "trn:workspace:group/ws",
			expected:   true,
		},
		{
			name:       "plain path",
			identifier: "group/ws",
			expected:   false,
		},
		{
			name:       "empty string",
			identifier: "",
			expected:   false,
		},
		{
			name:       "partial prefix",
			identifier: "trn",
			expected:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, IsTRN(test.identifier))
		})
	}
}

func TestToTRN(t *testing.T) {
	tests := []struct {
		name         string
		resourceType ResourceType
		identifier   string
		expected     string
	}{
		{
			name:         "already a trn",
			resourceType: ResourceTypeWorkspace,
			identifier:   "trn:workspace:group/ws",
			expected:     "trn:workspace:group/ws",
		},
		{
			name:         "path converted to trn",
			resourceType: ResourceTypeWorkspace,
			identifier:   "group/my-workspace",
			expected:     "trn:workspace:group/my-workspace",
		},
		{
			name:         "gid returned as-is",
			resourceType: ResourceTypeWorkspace,
			identifier:   "V19mNDdhYzEwYi01OGNjLTQzNzItYTU2Ny0wZTAyYjJjM2Q0Nzk", // base64url of W_f47ac10b-58cc-4372-a567-0e02b2c3d479
			expected:     "V19mNDdhYzEwYi01OGNjLTQzNzItYTU2Ny0wZTAyYjJjM2Q0Nzk",
		},
		{
			name:         "group path",
			resourceType: ResourceTypeGroup,
			identifier:   "parent/child",
			expected:     "trn:group:parent/child",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, ToTRN(test.resourceType, test.identifier))
		})
	}
}

func TestNewResourceTRN(t *testing.T) {
	tests := []struct {
		name         string
		resourceType ResourceType
		args         []string
		expected     string
	}{
		{
			name:         "single path segment",
			resourceType: ResourceTypeGroup,
			args:         []string{"my-group"},
			expected:     "trn:group:my-group",
		},
		{
			name:         "multiple path segments",
			resourceType: ResourceTypeWorkspace,
			args:         []string{"group", "workspace"},
			expected:     "trn:workspace:group/workspace",
		},
		{
			name:         "full path as single arg",
			resourceType: ResourceTypeManagedIdentity,
			args:         []string{"group/my-identity"},
			expected:     "trn:managed_identity:group/my-identity",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, NewResourceTRN(test.resourceType, test.args...))
		})
	}
}
