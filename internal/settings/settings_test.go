package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettings(t *testing.T) {
	type testCase struct {
		name string
		run  func(t *testing.T)
	}

	testCases := []testCase{
		{
			name: "SetCurrentProfile with existing profile",
			run: func(t *testing.T) {
				s := &Settings{
					Profiles: map[string]Profile{
						"default": {Endpoint: "https://example.com"},
					},
				}

				require.NoError(t, s.SetCurrentProfile("default"))
				assert.Equal(t, "https://example.com", s.CurrentProfile.Endpoint)
			},
		},
		{
			name: "SetCurrentProfile with non-existent profile",
			run: func(t *testing.T) {
				s := &Settings{
					Profiles: map[string]Profile{
						"default": {Endpoint: "https://example.com"},
					},
				}

				assert.EqualError(t, s.SetCurrentProfile("missing"), "no profile named missing")
			},
		},
		{
			name: "FindProfileByEndpoint found",
			run: func(t *testing.T) {
				s := &Settings{
					Profiles: map[string]Profile{
						"default": {Name: "default", Endpoint: "https://example.com"},
					},
				}

				profile, err := s.FindProfileByEndpoint("https://example.com")
				require.NoError(t, err)
				assert.Equal(t, "default", profile.Name)
			},
		},
		{
			name: "FindProfileByEndpoint not found",
			run: func(t *testing.T) {
				s := &Settings{
					Profiles: map[string]Profile{
						"default": {Endpoint: "https://example.com"},
					},
				}

				_, err := s.FindProfileByEndpoint("https://other.com")
				assert.EqualError(t, err, `no profile found for endpoint "https://other.com"`)
			},
		},
		{
			name: "SetToken",
			run: func(t *testing.T) {
				p := Profile{}
				assert.Nil(t, p.token)

				p.SetToken("my-token")
				require.NotNil(t, p.token)
				assert.Equal(t, "my-token", *p.token)
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, test.run)
	}
}
