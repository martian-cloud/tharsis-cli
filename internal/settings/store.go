package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

const (
	settingsDirectoryName = ".tharsis"
	credentialsFilename   = "credentials.json"
	settingsFilename      = "settings.json"

	// flagWrite truncates existing files to avoid leaving stale content
	// when the new payload is shorter than the old one.
	flagWrite = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
)

// ErrNoSettings is a special error if settings file does not exist.
var ErrNoSettings = errors.New("please run 'tharsis configure' to create your initial settings file")

// fileStore handles reading and writing settings and credentials files.
type fileStore struct {
	settingsPath    string
	credentialsPath string
}

func newFileStore() (*fileStore, error) {
	settingsPath, err := DefaultSettingsFilepath()
	if err != nil {
		return nil, err
	}

	credentialsPath, err := DefaultCredentialsFilepath()
	if err != nil {
		return nil, err
	}

	return &fileStore{
		settingsPath:    settingsPath,
		credentialsPath: credentialsPath,
	}, nil
}

// readSettings loads profiles from the settings file and merges in tokens
// from the credentials file. A missing credentials file is not an error
// because it won't exist until the first `sso login`.
func (fs *fileStore) readSettings() (*Settings, error) {
	data, err := readFileAt(fs.settingsPath)
	if err != nil {
		return nil, err
	}

	var s Settings
	if err = json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	credData, credErr := readFileAt(fs.credentialsPath)
	if credErr != nil && !errors.Is(credErr, ErrNoSettings) {
		return nil, fmt.Errorf("failed to read credentials file: %w", credErr)
	}

	// Tokens are stored separately from settings so the credentials file
	// can have stricter permissions and be gitignored independently.
	var c struct {
		Tokens map[string]string `json:"stateTokens"`
	}

	if credData != nil {
		if err = json.Unmarshal(credData, &c); err != nil {
			return nil, fmt.Errorf("failed to read credentials file: %w", err)
		}
	}

	// Populate non-persistent fields that are derived at load time
	// rather than stored in the JSON.
	for name, profile := range s.Profiles {
		profile.Name = name
		if token, ok := c.Tokens[name]; ok {
			profile.token = &token
		}

		s.Profiles[name] = profile
	}

	return &s, nil
}

// writeSettings persists profiles to the settings file and extracts tokens
// into the credentials file. The two-file split keeps secrets out of the
// main settings file.
func (fs *fileStore) writeSettings(s *Settings) error {
	if err := writeJSON(s, fs.settingsPath); err != nil {
		return err
	}

	c := struct {
		Tokens map[string]string `json:"stateTokens"`
	}{Tokens: map[string]string{}}
	for k, profile := range s.Profiles {
		if profile.token != nil {
			c.Tokens[k] = *profile.token
		}
	}

	return writeJSON(c, fs.credentialsPath)
}

// DefaultSettingsFilepath returns the default settings file path.
func DefaultSettingsFilepath() (string, error) {
	dirname, err := defaultSettingsDirectory()
	if err != nil {
		return "", err
	}

	return filepath.Join(dirname, settingsFilename), nil
}

// DefaultCredentialsFilepath returns the OS-dependent path/name of the default credentials file.
func DefaultCredentialsFilepath() (string, error) {
	dirname, err := defaultSettingsDirectory()
	if err != nil {
		return "", err
	}

	return filepath.Join(dirname, credentialsFilename), nil
}

func readFileAt(path string) ([]byte, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, ErrNoSettings
	}

	data, err := os.ReadFile(path) // nosemgrep: gosec.G304-1
	if err != nil {
		return nil, err
	}

	return data, nil
}

func writeJSON(v any, path string) error {
	buf, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	dirPath := filepath.Dir(path)
	if _, sErr := os.Stat(dirPath); os.IsNotExist(sErr) {
		if mErr := os.MkdirAll(dirPath, 0o700); mErr != nil {
			return fmt.Errorf("failed to create settings directory: %s", mErr)
		}
	}

	writer, err := os.OpenFile(path, flagWrite, 0o600) // nosemgrep: gosec.G304-1
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = writer.Write(buf)
	return err
}

func defaultSettingsDirectory() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, settingsDirectoryName), nil
	}

	u, err := user.Current()
	if err != nil {
		return "", err
	}

	if u.HomeDir == "" {
		return "", fmt.Errorf("failed to find user's home directory")
	}

	return filepath.Join(u.HomeDir, settingsDirectoryName), nil
}
