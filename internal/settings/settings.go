// Package settings contains the logic for reading,
// writing, associating profiles with a user's settings
// file. It also configures a Tharsis SDK client.
package settings

// Functions to work with the settings file and its siblings.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
)

const (
	// settingsDirectoryName is the name of directory where settings reside.
	settingsDirectoryName = ".tharsis"

	// credentialsFilename is the name of the credentials file.
	credentialsFilename = "credentials.json"

	// settingsFilename is the name of the settings file.
	settingsFilename = "settings.json"

	// DefaultProfileName is the name of the default profile
	DefaultProfileName = "default"

	// Flag for creating or truncating a file and then writing to it.
	flagWrite = os.O_CREATE | os.O_TRUNC | os.O_WRONLY

	// Content of no settings error messages.
	noSettingsMessage = "please run 'tharsis configure' to create your initial settings file"
)

// ErrNoSettings is a special error if settings file does not exist.
var ErrNoSettings = errors.New(noSettingsMessage)

// creds is used to store the credentials to the credentials file.
// Callers should still access the Token field on the Profile.
type creds struct {
	// Map of profile name to its token.
	Tokens map[string]string `json:"stateTokens"`
}

// A settings file can define several profiles, of which exactly one is the "default" profile.

// Settings holds the contents of one settings file and
// a pointer to the profile specified by the current command.
type Settings struct {
	Profiles       map[string]Profile `json:"profiles"`
	CurrentProfile *Profile           `json:"-"` // This field is not persistent, do not write it out.
}

// Profile holds the contents of one profile from a settings file.
type Profile struct {
	token         *string `json:"-"`        // Not persistent, written via creds struct above!
	Endpoint      string  `json:"endpoint"` // HTTP.
	TLSSkipVerify bool    `json:"tlsSkipVerify"`
}

// SetToken sets the token for a profile.
func (p *Profile) SetToken(token string) {
	p.token = &token
}

// ReadSettings reads the settings file.
// If no argument, it reads the default settings file: ~/.tharsis/settings.json
func ReadSettings(name *string) (*Settings, error) {
	data, err := readFile(name, DefaultSettingsFilepath)
	if err != nil {
		return nil, err
	}

	var settings Settings
	if err = json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	if err := readCredentials(&settings); err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	return &settings, nil
}

// readCredentials reads the credentials from the credentials file and
// populates the 'Token' field on the profiles.
func readCredentials(settings *Settings) error {
	data, err := readFile(nil, DefaultCredentialsFilepath)
	if err != nil {
		if errors.Is(err, ErrNoSettings) {
			return nil
		}
		return err
	}

	var c creds
	if err = json.Unmarshal(data, &c); err != nil {
		return err
	}

	// Populate appropriate profile with the token read.
	for name, profile := range settings.Profiles {
		if token, ok := c.Tokens[name]; ok {
			profile.token = &token
			settings.Profiles[name] = profile
		}
	}

	return nil
}

func readFile(name *string, defaultFunc func() (string, error)) ([]byte, error) {
	// Figure out the filename.
	filename, err := resolveFilename(name, defaultFunc)
	if err != nil {
		return nil, err
	}

	// Check whether the file exists, so a more user-friendly error can be returned.
	if _, oErr := os.Stat(filename); os.IsNotExist(oErr) {
		return nil, ErrNoSettings
	}

	// Read the file.
	bytes, err := os.ReadFile(filename) // nosemgrep: gosec.G304-1
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// SetCurrentProfile sets the current profile pointer in a settings object.
func (s *Settings) SetCurrentProfile(profileName string) error {
	foundProfile, ok := s.Profiles[profileName]
	if !ok {
		return fmt.Errorf("no profile named %s", profileName)
	}

	s.CurrentProfile = &foundProfile
	return nil
}

// WriteSettingsFile writes the settings file.
// If no filename argument, it writes the default settings file: ~/.tharsis/settings.json
func (s *Settings) WriteSettingsFile(name *string) error {
	// Figure out the filename.
	filename, err := resolveFilename(name, DefaultSettingsFilepath)
	if err != nil {
		return err
	}

	if err = s.writeFile(s, filename); err != nil {
		return err
	}

	return s.writeCredentialsFile()
}

// writeCredentialsFile writes the credentials file.
// It writes the default credentials file: ~/.tharsis/credentials.json
func (s *Settings) writeCredentialsFile() error {
	filename, err := resolveFilename(nil, DefaultCredentialsFilepath)
	if err != nil {
		return err
	}

	c := &creds{Tokens: map[string]string{}}
	for k, profile := range s.Profiles {
		if profile.token != nil {
			c.Tokens[k] = *profile.token
		}
	}

	return s.writeFile(c, filename)
}

func (s *Settings) writeFile(v any, filename string) error {
	buf, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	// Create the directory if necessary.
	dirPath := filepath.Dir(filename)
	if _, sErr := os.Stat(dirPath); os.IsNotExist(sErr) {
		// Must create the directory.
		mErr := os.MkdirAll(dirPath, 0o700)
		if mErr != nil {
			return fmt.Errorf("failed to create settings directory: %s", mErr)
		}
	}

	writer, err := os.OpenFile(filename, flagWrite, 0o600) // nosemgrep: gosec.G304-1
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = (*writer).Write(buf)
	return err
}

// resolveFilename figures out a filename, including calling a function to find the default name.
func resolveFilename(specified *string, defaultFunc func() (string, error)) (string, error) {
	if specified == nil {
		// Return the default.
		return defaultFunc()
	}

	// Return the specified.
	return *specified, nil
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

// Return the OS-dependent path of the default settings directory.
func defaultSettingsDirectory() (string, error) {
	// If $HOME is set, use it.
	home := os.Getenv("HOME")
	if home != "" {
		return filepath.Join(home, settingsDirectoryName), nil
	}

	// Otherwise, use the Golang calls.
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	homedir := user.HomeDir
	if homedir == "" {
		return "", fmt.Errorf("failed to find user's home directory")
	}
	return filepath.Join(homedir, settingsDirectoryName), nil
}

// NewTokenGetter creates a new token getter for the profile.
func (p *Profile) NewTokenGetter(ctx context.Context) (client.TokenGetter, error) {
	tokenGetter, err := createTokenGetter(ctx, p.token, p.Endpoint, p.TLSSkipVerify)
	if err != nil {
		return nil, err
	}

	return tokenGetter, nil
}

// NewClient returns a Tharsis client based on the specified profile.
func (p *Profile) NewClient(ctx context.Context, withAuth bool, userAgent string, logger hclog.Logger) (*client.Client, error) {
	clientConfig := &client.Config{
		HTTPEndpoint:  p.Endpoint,
		TLSSkipVerify: p.TLSSkipVerify,
		UserAgent:     userAgent,
		Logger:        logger,
	}

	if withAuth {
		tokenGetter, err := p.NewTokenGetter(ctx)
		if err != nil {
			return nil, err
		}

		// Update the client config since values may have been overridden.
		clientConfig.HTTPEndpoint = p.Endpoint
		clientConfig.TLSSkipVerify = p.TLSSkipVerify
		clientConfig.TokenGetter = tokenGetter
	}

	return client.New(ctx, clientConfig)
}
