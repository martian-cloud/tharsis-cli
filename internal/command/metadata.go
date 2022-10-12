package command

import (
	"os"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/logger"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/settings"
)

// Metadata communicates information from the 'main' tharsis module to
// the individual commands run via the CLI platform.
//
// In particular, the BinaryName field allows long-form command help to
// accurately show the binary name, even if the binary is copied.
//
// The logger gives the command execution code access to the shared logger.
//
// The name of the current profile (set by the global option) must be stored
// here in order to communicate it from the top-level module down to to the
// individual command modules prior to reading the settings file.
//
type Metadata struct {
	BinaryName         string
	DisplayTitle       string
	Version            string
	Logger             logger.Logger
	UI                 cli.Ui
	CurrentProfileName string
	//
	// In order to have the Makefile set the default endpoint URL through the
	// main package at build time and the configure command use the value,
	// it is necessary to pass the value through this struct.  If the configure
	// command attempts to reference the value in the main package, that causes
	// an import cycle.
	DefaultEndpointURL string
}

// ReadSettings reads the settings file.  If the force argument is false and
// the settings have already been read, this function does nothing.
func (m *Metadata) ReadSettings() (*settings.Settings, error) {

	// Read the current settings.
	currentSettings, err := settings.ReadSettings(nil)
	if err != nil {
		// Build settings manually if settings don't exist and service account is being used.
		if err == settings.ErrNoSettings && m.DefaultEndpointURL != "" && os.Getenv("THARSIS_SERVICE_ACCOUNT_TOKEN") != "" {
			return &settings.Settings{
				CurrentProfile: &settings.Profile{
					TharsisURL: m.DefaultEndpointURL,
				},
			}, nil
		}
		return nil, err
	}

	// Now, we can set the current profile pointer.
	err = currentSettings.SetCurrentProfile(m.CurrentProfileName)
	if err != nil {
		return nil, err
	}

	m.Logger.Debugf("settings: %#v", currentSettings)

	return currentSettings, nil
}

// The End.
