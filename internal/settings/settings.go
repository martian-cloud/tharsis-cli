// Package settings contains the logic for reading,
// writing, associating profiles with a user's settings
// file. It also configures a Tharsis SDK client.
package settings

// Functions to work with the settings file and its siblings.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdkauth "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/auth"
	sdkconfig "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/config"
)

const (

	// SettingsDirectoryName is the name of directory where settings reside.
	SettingsDirectoryName = ".tharsis"

	// SettingsFilename is the name of the settings file.
	SettingsFilename = "settings.json"

	// DefaultProfileName is the name of the default profile
	DefaultProfileName = "default"

	// Flag for creating or truncating a file and then writing to it.
	flagWrite = os.O_CREATE | os.O_TRUNC | os.O_WRONLY

	// Content of no settings error messages.
	noSettingsMessage = "please run 'tharsis configure' to create your initial settings file"
)

var (

	// ErrNoSettings is a special error if settings file does not exist.
	ErrNoSettings = fmt.Errorf(noSettingsMessage)
)

// A settings file can define several profiles, of which exactly one is the "default" profile.

// Settings holds the contents of one settings file and
// a pointer to the profile specified by the current command.
type Settings struct {
	Profiles       map[string]Profile `json:"profiles"`
	CurrentProfile *Profile           `json:"-"` // This field is not persistent, do not write it out.
}

// Profile holds the contents of one profile from a settings file.
type Profile struct {
	Token      *string `json:"stateToken"`
	TharsisURL string  `json:"tharsisURL"`
}

// ReadSettings reads the settings file.
// If no argument, it reads the default settings file: ~/.tharsis/settings.json
func ReadSettings(name *string) (*Settings, error) {

	// Figure out the filename.
	filename, err := resolveFilename(name, DefaultSettingsFilename)
	if err != nil {
		return nil, err
	}

	// Check whether the file exists, so a more user-friendly error can be returned.
	if _, oErr := os.Stat(filename); os.IsNotExist(oErr) {
		return nil, ErrNoSettings
	}

	// Read the file.
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Decode the file's contents.
	var settings Settings
	err = json.Unmarshal(bytes, &settings)
	if err != nil {
		return nil, err
	}

	return &settings, nil
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

// WriteSettingsFile reads the settings file.
// If no filename argument, it writes the default settings file: ~/.tharsis/settings.json
func (s *Settings) WriteSettingsFile(name *string) error {

	// Figure out the filename.
	filename, err := resolveFilename(name, DefaultSettingsFilename)
	if err != nil {
		return err
	}

	// Encode the file's contents.
	buf, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	// Create the directory if necessary.
	dirPath := filepath.Dir(filename)
	if _, sErr := os.Stat(dirPath); os.IsNotExist(sErr) {
		// Must create the directory.
		mErr := os.MkdirAll(dirPath, 0700)
		if mErr != nil {
			return fmt.Errorf("failed to create settings directory: %s", mErr)
		}
	}

	// Write the file.
	writer, err := os.OpenFile(filename, flagWrite, 0600)
	if err != nil {
		return err
	}
	defer writer.Close()
	_, err = (*writer).Write(buf)
	if err != nil {
		return err
	}

	return nil
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

// DefaultSettingsFilename returns the OS-dependent path/name of the default settings file.
func DefaultSettingsFilename() (string, error) {
	dirname, err := defaultSettingsDirectory()
	if err != nil {
		return "", err
	}

	return filepath.Join(dirname, SettingsFilename), nil
}

// Return the OS-dependent path of the default settings directory.
func defaultSettingsDirectory() (string, error) {

	// If $HOME is set, use it.
	home := os.Getenv("HOME")
	if home != "" {
		return filepath.Join(home, SettingsDirectoryName), nil
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
	return filepath.Join(homedir, SettingsDirectoryName), nil
}


// getAuthProvider attempts to return an SDK authentication provider.
// If not successful but no error happened while reading the file, it returns nil.
func getAuthProvider(profile *Profile) (sdkauth.TokenProvider, error) {

	token := profile.Token
	if token == nil {
		// No token available, so don't return a provider.
		return nil, nil
	}

	// So far, only static tokens are supported by this CLI's settings and so forth.
	provider, err := sdkauth.NewStaticTokenProvider(*token)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

// GetSDKClient returns a Tharsis SDK client with user agent configured
func (p *Profile) GetSDKClient(userAgent string) (*tharsis.Client, error) {
	provider, err := getAuthProvider(p)
	if err != nil {
		return nil, err
	}

	configOptions := []func(*sdkconfig.LoadOptions) error{
		sdkconfig.WithTokenProvider(provider),
		sdkconfig.WithEndpoint(p.TharsisURL),
	}
	
	// Add user agent if provided
	if userAgent != "" {
		configOptions = append(configOptions, sdkconfig.WithUserAgent(userAgent))
	}

	sdkConfig, err := sdkconfig.Load(configOptions...)
	if err != nil {
		return nil, err
	}

	client, err := tharsis.NewClient(sdkConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}
