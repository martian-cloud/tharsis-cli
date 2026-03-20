package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetCurrentProfile(t *testing.T) {
	type testCase struct {
		name           string
		profileName    string
		expectErrorMsg string
	}

	testCases := []testCase{
		{
			name:        "existing profile",
			profileName: "default",
		},
		{
			name:           "non-existent profile",
			profileName:    "missing",
			expectErrorMsg: "no profile named missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			s := &Settings{
				Profiles: map[string]Profile{
					"default": {Endpoint: "https://example.com"},
				},
			}

			err := s.SetCurrentProfile(test.profileName)

			if test.expectErrorMsg != "" {
				assert.EqualError(t, err, test.expectErrorMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, "https://example.com", s.CurrentProfile.Endpoint)
		})
	}
}

func TestFindProfileByEndpoint(t *testing.T) {
	type testCase struct {
		name           string
		endpoint       string
		expectName     string
		expectErrorMsg string
	}

	testCases := []testCase{
		{
			name:       "found",
			endpoint:   "https://example.com",
			expectName: "default",
		},
		{
			name:           "not found",
			endpoint:       "https://other.com",
			expectErrorMsg: `no profile found for endpoint "https://other.com"`,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			s := &Settings{
				Profiles: map[string]Profile{
					"default": {Name: "default", Endpoint: "https://example.com"},
				},
			}

			profile, err := s.FindProfileByEndpoint(test.endpoint)

			if test.expectErrorMsg != "" {
				assert.EqualError(t, err, test.expectErrorMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expectName, profile.Name)
			assert.Equal(t, test.endpoint, profile.Endpoint)
		})
	}
}

func TestSetToken(t *testing.T) {
	p := Profile{}
	assert.Nil(t, p.token)

	p.SetToken("my-token")
	require.NotNil(t, p.token)
	assert.Equal(t, "my-token", *p.token)
}
