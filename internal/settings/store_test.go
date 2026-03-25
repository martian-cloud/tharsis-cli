package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore(t *testing.T) {
	type testCase struct {
		name string
		run  func(t *testing.T)
	}

	testCases := []testCase{
		{
			name: "read settings with credentials",
			run: func(t *testing.T) {
				fs := setupFileStore(t,
					`{"profiles":{"default":{"endpoint":"https://example.com"}}}`,
					`{"stateTokens":{"default":"my-token"}}`,
				)

				s, err := fs.readSettings()
				require.NoError(t, err)
				assert.Len(t, s.Profiles, 1)
				assert.Equal(t, "default", s.Profiles["default"].Name)
				require.NotNil(t, s.Profiles["default"].token)
				assert.Equal(t, "my-token", *s.Profiles["default"].token)
			},
		},
		{
			name: "read settings without credentials file",
			run: func(t *testing.T) {
				fs := setupFileStore(t,
					`{"profiles":{"default":{"endpoint":"https://example.com"}}}`,
					"",
				)

				s, err := fs.readSettings()
				require.NoError(t, err)
				assert.Len(t, s.Profiles, 1)
				assert.Nil(t, s.Profiles["default"].token)
			},
		},
		{
			name: "read settings with missing settings file",
			run: func(t *testing.T) {
				fs := setupFileStore(t, "", "")

				_, err := fs.readSettings()
				assert.ErrorIs(t, err, ErrNoSettings)
			},
		},
		{
			name: "write settings splits credentials",
			run: func(t *testing.T) {
				dir := t.TempDir()
				fs := &fileStore{
					settingsPath:    filepath.Join(dir, "settings.json"),
					credentialsPath: filepath.Join(dir, "credentials.json"),
				}

				token := "my-token"
				s := &Settings{
					Profiles: map[string]Profile{
						"default": {token: &token, Endpoint: "https://example.com"},
					},
				}

				require.NoError(t, fs.writeSettings(s))

				// Settings file should have the endpoint but not the token.
				data, err := os.ReadFile(fs.settingsPath)
				require.NoError(t, err)

				var written Settings
				require.NoError(t, json.Unmarshal(data, &written))
				assert.Equal(t, "https://example.com", written.Profiles["default"].Endpoint)

				// Credentials file should have the token.
				credData, err := os.ReadFile(fs.credentialsPath)
				require.NoError(t, err)

				var c struct {
					Tokens map[string]string `json:"stateTokens"`
				}
				require.NoError(t, json.Unmarshal(credData, &c))
				assert.Equal(t, "my-token", c.Tokens["default"])
			},
		},
		{
			name: "write settings creates nested directory",
			run: func(t *testing.T) {
				dir := filepath.Join(t.TempDir(), "nested", "dir")
				fs := &fileStore{
					settingsPath:    filepath.Join(dir, "settings.json"),
					credentialsPath: filepath.Join(dir, "credentials.json"),
				}

				s := &Settings{
					Profiles: map[string]Profile{
						"default": {Endpoint: "https://example.com"},
					},
				}

				require.NoError(t, fs.writeSettings(s))
				assert.FileExists(t, fs.settingsPath)
				assert.FileExists(t, fs.credentialsPath)
			},
		},
		{
			name: "round trip preserves all fields",
			run: func(t *testing.T) {
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
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, test.run)
	}
}

// setupFileStore creates a fileStore pointing at temp files, optionally
// writing initial settings and credentials content.
func setupFileStore(t *testing.T, settingsJSON, credentialsJSON string) *fileStore {
	t.Helper()

	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.json")
	credentialsPath := filepath.Join(dir, "credentials.json")

	if settingsJSON != "" {
		require.NoError(t, os.WriteFile(settingsPath, []byte(settingsJSON), 0o600))
	}

	if credentialsJSON != "" {
		require.NoError(t, os.WriteFile(credentialsPath, []byte(credentialsJSON), 0o600))
	}

	return &fileStore{
		settingsPath:    settingsPath,
		credentialsPath: credentialsPath,
	}
}
