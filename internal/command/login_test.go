package command

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/settings"
)

func TestFindDuplicateEndpointProfiles(t *testing.T) {
	tests := []struct {
		name               string
		profiles           map[string]settings.Profile
		currentProfileName string
		expectedDuplicates []string
	}{
		{
			name: "no duplicates",
			profiles: map[string]settings.Profile{
				"default": {Endpoint: "https://tharsis.example.com"},
				"staging": {Endpoint: "https://tharsis-staging.example.com"},
			},
			currentProfileName: "default",
			expectedDuplicates: nil,
		},
		{
			name: "one duplicate",
			profiles: map[string]settings.Profile{
				"default": {Endpoint: "https://tharsis.example.com"},
				"old":     {Endpoint: "https://tharsis.example.com"},
			},
			currentProfileName: "default",
			expectedDuplicates: []string{"old"},
		},
		{
			name: "multiple duplicates",
			profiles: map[string]settings.Profile{
				"default": {Endpoint: "https://tharsis.example.com"},
				"old1":    {Endpoint: "https://tharsis.example.com"},
				"old2":    {Endpoint: "https://tharsis.example.com"},
			},
			currentProfileName: "default",
			expectedDuplicates: []string{"old1", "old2"},
		},
		{
			name: "different paths are not duplicates",
			profiles: map[string]settings.Profile{
				"default": {Endpoint: "https://tharsis.example.com"},
				"other":   {Endpoint: "https://tharsis.example.com/v2"},
			},
			currentProfileName: "default",
			expectedDuplicates: nil,
		},
		{
			name: "different ports are not duplicates",
			profiles: map[string]settings.Profile{
				"default": {Endpoint: "https://tharsis.example.com:8443"},
				"other":   {Endpoint: "https://tharsis.example.com:9443"},
			},
			currentProfileName: "default",
			expectedDuplicates: nil,
		},
		{
			name: "trailing slash is a different string",
			profiles: map[string]settings.Profile{
				"default": {Endpoint: "https://tharsis.example.com"},
				"other":   {Endpoint: "https://tharsis.example.com/"},
			},
			currentProfileName: "default",
			expectedDuplicates: nil,
		},
		{
			name: "case difference is a different string",
			profiles: map[string]settings.Profile{
				"default": {Endpoint: "https://tharsis.example.com"},
				"other":   {Endpoint: "https://Tharsis.Example.COM"},
			},
			currentProfileName: "default",
			expectedDuplicates: nil,
		},
		{
			name:               "empty profiles map",
			profiles:           map[string]settings.Profile{},
			currentProfileName: "default",
			expectedDuplicates: nil,
		},
		{
			name: "single profile no duplicate",
			profiles: map[string]settings.Profile{
				"default": {Endpoint: "https://tharsis.example.com"},
			},
			currentProfileName: "default",
			expectedDuplicates: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findDuplicateEndpointProfiles(tt.profiles, tt.currentProfileName)
			// Sort for deterministic comparison since map iteration is unordered.
			sort.Strings(result)
			sort.Strings(tt.expectedDuplicates)
			assert.Equal(t, tt.expectedDuplicates, result)
		})
	}
}
