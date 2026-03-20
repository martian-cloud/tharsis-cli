package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStoreReadSettings(t *testing.T) {
	type testCase struct {
		name            string
		settingsJSON    string
		credentialsJSON string
		noCreds         bool
		expectError     bool
		expectProfiles  int
		expectTokenSet  bool
	}

	testCases := []testCase{
		{
			name:            "settings with credentials",
			settingsJSON:    `{"profiles":{"default":{"endpoint":"https://example.com"}}}`,
			credentialsJSON: `{"stateTokens":{"default":"my-token"}}`,
			expectProfiles:  1,
			expectTokenSet:  true,
		},
		{
			name:           "settings without credentials file",
			settingsJSON:   `{"profiles":{"default":{"endpoint":"https://example.com"}}}`,
			noCreds:        true,
			expectProfiles: 1,
			expectTokenSet: false,
		},
		{
			name:        "missing settings file",
			expectError: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			settingsPath := filepath.Join(dir, "settings.json")
			credentialsPath := filepath.Join(dir, "credentials.json")

			if test.settingsJSON != "" {
				require.NoError(t, os.WriteFile(settingsPath, []byte(test.settingsJSON), 0o600))
			}

			if test.credentialsJSON != "" {
				require.NoError(t, os.WriteFile(credentialsPath, []byte(test.credentialsJSON), 0o600))
			}

			fs := &fileStore{
				settingsPath:    settingsPath,
				credentialsPath: credentialsPath,
			}

			s, err := fs.readSettings()

			if test.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, s.Profiles, test.expectProfiles)

			if test.expectTokenSet {
				p := s.Profiles["default"]
				require.NotNil(t, p.token)
				assert.Equal(t, "my-token", *p.token)
			}

			if test.expectProfiles > 0 {
				assert.Equal(t, "default", s.Profiles["default"].Name)
			}
		})
	}
}

func TestFileStoreWriteSettings(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.json")
	credentialsPath := filepath.Join(dir, "credentials.json")

	fs := &fileStore{
		settingsPath:    settingsPath,
		credentialsPath: credentialsPath,
	}

	token := "my-token"
	s := &Settings{
		Profiles: map[string]Profile{
			"default": {
				token:    &token,
				Endpoint: "https://example.com",
			},
		},
	}

	require.NoError(t, fs.writeSettings(s))

	// Verify settings file.
	data, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var written Settings
	require.NoError(t, json.Unmarshal(data, &written))
	assert.Equal(t, "https://example.com", written.Profiles["default"].Endpoint)

	// Verify credentials file.
	credData, err := os.ReadFile(credentialsPath)
	require.NoError(t, err)

	var c struct {
		Tokens map[string]string `json:"stateTokens"`
	}
	require.NoError(t, json.Unmarshal(credData, &c))
	assert.Equal(t, "my-token", c.Tokens["default"])
}

func TestFileStoreWriteSettingsCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "dir")
	settingsPath := filepath.Join(dir, "settings.json")
	credentialsPath := filepath.Join(dir, "credentials.json")

	fs := &fileStore{
		settingsPath:    settingsPath,
		credentialsPath: credentialsPath,
	}

	s := &Settings{
		Profiles: map[string]Profile{
			"default": {Endpoint: "https://example.com"},
		},
	}

	require.NoError(t, fs.writeSettings(s))
	assert.FileExists(t, settingsPath)
	assert.FileExists(t, credentialsPath)
}

func TestFileStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	fs := &fileStore{
		settingsPath:    filepath.Join(dir, "settings.json"),
		credentialsPath: filepath.Join(dir, "credentials.json"),
	}

	token := "round-trip-token"
	original := &Settings{
		Profiles: map[string]Profile{
			"prod": {
				token:         &token,
				Endpoint:      "https://prod.example.com",
				TLSSkipVerify: true,
			},
		},
	}

	require.NoError(t, fs.writeSettings(original))

	loaded, err := fs.readSettings()
	require.NoError(t, err)

	p := loaded.Profiles["prod"]
	assert.Equal(t, "prod", p.Name)
	assert.Equal(t, "https://prod.example.com", p.Endpoint)
	assert.True(t, p.TLSSkipVerify)
	require.NotNil(t, p.token)
	assert.Equal(t, "round-trip-token", *p.token)
}
