// Package settings manages CLI profile configuration and authentication.
// Settings and credentials are stored in separate files (~/.tharsis/)
// so that credentials can have stricter permissions and be excluded
// from version control independently.
package settings

import "fmt"

const (
	// DefaultProfileName is the name of the default profile
	DefaultProfileName = "default"
)

// Settings holds the contents of one settings file and
// a pointer to the profile specified by the current command.
type Settings struct {
	// store is retained so WriteSettingsFile can reuse the same
	// paths that were resolved during ReadSettings or NewSettings.
	store          *fileStore         `json:"-"`
	Profiles       map[string]Profile `json:"profiles"`
	CurrentProfile *Profile           `json:"-"`
}

// NewSettings creates a new empty Settings with an initialized store.
func NewSettings() (*Settings, error) {
	store, err := newFileStore()
	if err != nil {
		return nil, err
	}

	return &Settings{store: store}, nil
}

// FindProfileByEndpoint returns the profile matching the given endpoint.
func (s *Settings) FindProfileByEndpoint(endpoint string) (*Profile, error) {
	for _, profile := range s.Profiles {
		if profile.Endpoint == endpoint {
			return &profile, nil
		}
	}

	return nil, fmt.Errorf("no profile found for endpoint %q", endpoint)
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

// WriteSettingsFile writes the settings and credentials to the default location.
func (s *Settings) WriteSettingsFile() error {
	return s.store.writeSettings(s)
}

// ReadSettings reads the settings and credentials from the default location.
func ReadSettings() (*Settings, error) {
	store, err := newFileStore()
	if err != nil {
		return nil, err
	}

	s, err := store.readSettings()
	if err != nil {
		return nil, err
	}

	s.store = store
	return s, nil
}
