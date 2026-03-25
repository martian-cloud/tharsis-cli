// Package command contains all the logic for different commands
// within the CLI. It is the main gateway to doing operations
// against the Tharsis API.
package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/settings"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
)

const (
	// maxPaginationLimit is the default (and max) limit for paginated list commands.
	maxPaginationLimit int32 = 100

	// humanTimeFormat is the format used for displaying timestamps in human-readable output.
	humanTimeFormat = "January 2 2006, 3:04:05 PM MST"
)

// baseOptions contains the different ways to configure the behavior of BaseCommand.
type baseOptions struct {
	flags          *flag.Set
	force          *bool
	inputValidator func() error
	commandName    string
	args           []string
	withClient     bool
	withAuth       bool // Not stored in the settings.
	confirmPrompt  string
}

// BaseOptionsFunc is an alias that allows setting baseOptions.
type BaseOptionsFunc func(*baseOptions) error

// WithFlags sets the FlagSet that needs to be parsed. Return
// values are often set as fields on the caller command's struct.
func WithFlags(flags *flag.Set) BaseOptionsFunc {
	return func(o *baseOptions) error {
		o.flags = flags
		return nil
	}
}

// WithCommandName is the name of the command that called
// BaseCommand.initialize(). It should always be set as
// it allows for helpful debugger statements.
func WithCommandName(name string) BaseOptionsFunc {
	return func(o *baseOptions) error {
		o.commandName = name
		return nil
	}
}

// WithArguments sets the raw arguments that are passed into a command.
// It facilitates the parsing of flags and arguments.
func WithArguments(args []string) BaseOptionsFunc {
	return func(o *baseOptions) error {
		o.args = args
		return nil
	}
}

// WithInputValidator allows the calling command to pass in a input validator func,
// which once called, ensures proper data was passed into command. It can be used
// to make sure a flag was specified, or the value is a URL, etc.
func WithInputValidator(inputValidator func() error) BaseOptionsFunc {
	return func(o *baseOptions) error {
		o.inputValidator = inputValidator
		return nil
	}
}

// WithClient indicates that a gRPC client is needed by the command.
// Callers can set the withAuth parameter to indicate if client
// should be initialized with auth. Client should be available on the
// BaseCommand struct after initialize() has been called.
func WithClient(withAuth bool) BaseOptionsFunc {
	return func(o *baseOptions) error {
		o.withClient = true
		o.withAuth = withAuth
		return nil
	}
}

// WithForcePrompt prompts for confirmation in interactive mode when
// --force option is used to prevent accidental deletions for forceful
// actions. The prompt parameter is the confirmation message shown to the user.
func WithForcePrompt(force *bool, prompt string) BaseOptionsFunc {
	return func(o *baseOptions) error {
		o.force = force
		o.confirmPrompt = prompt
		return nil
	}
}

// BaseCommand contains data needed by all the CLI commands.
// It provides access to the UI, logger and other metadata
// information. Private fields are only populated after
// initialize() has been called and are entirely controllable
// by using the baseOptions above.
type BaseCommand struct {
	Context              context.Context
	Logger               hclog.Logger
	UI                   terminal.UI
	HTTPClient           *http.Client
	grpcClient           *client.Client
	Version              string
	DisplayTitle         string
	BinaryName           string
	CurrentProfileName   string
	DefaultHTTPEndpoint  string
	UserAgent            string
	arguments            []string
	DefaultTLSSkipVerify bool
}

// initialize performs some preliminary tasks for each command. It should be
// one of the first functions called by each of the commands to parse flags,
// arguments and initialize the SDK client, if needed. The values for parsed
// flags are generally stored in the caller command's struct, and any
// arguments are available in the 'arguments' field. Use baseOptions above
// to control what happens when. Errors are already logged, so caller can
// simply check for a non-zero status code.
func (c *BaseCommand) initialize(opts ...BaseOptionsFunc) int {
	// Populate baseOptions struct with options.
	o := baseOptions{}
	for _, opt := range opts {
		if err := opt(&o); err != nil {
			c.UI.ErrorWithSummary(err, "failed to load base command options")
			return 1
		}
	}

	c.Logger.Debug("starting command", "name", o.commandName, "argCount", len(o.args))
	for ix, arg := range o.args {
		c.Logger.Debug("argument", "index", ix, "value", arg)
	}

	if o.flags != nil {
		// Discard any output from flags.
		o.flags.SetOutput(io.Discard)

		// Parse flags.
		if err := o.flags.Parse(o.args); err != nil {
			c.UI.ErrorWithSummary(err, "failed to parse %s options", o.commandName)
			return cli.RunResultHelp
		}

		c.arguments = o.flags.Args()
	} else {
		// There are no flags defined for the command, so default to all arguments.
		c.arguments = o.args
	}

	// Call input validator if there is one.
	if o.inputValidator != nil {
		if err := o.inputValidator(); err != nil {
			c.UI.ErrorWithSummary(err, "failed to validate %s input", o.commandName)
			return cli.RunResultHelp
		}
	}

	// Prompt for confirmation if destructive operation in interactive mode.
	// Only prompt when --force is used.
	if o.confirmPrompt != "" && c.UI.Interactive() && o.force != nil && *o.force {
		confirmed, err := c.UI.Confirm(o.confirmPrompt)
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to confirm")
			return 1
		}

		if !confirmed {
			c.UI.Infof("Operation cancelled.")
			return 1
		}
	}

	if o.withClient {
		curSettings, err := c.getCurrentSettings()
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to get current settings")
			return 1
		}

		client, err := curSettings.CurrentProfile.NewClient(c.Context, o.withAuth, c.UserAgent, c.Logger)
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to get client")
			return 1
		}

		c.grpcClient = client
	}

	return 0
}

// getCurrentSettings returns the current settings in use for the CLI.
func (c *BaseCommand) getCurrentSettings() (*settings.Settings, error) {
	// TODO: Remove migration after a few releases when all users have migrated
	// Migrate old Tharsis settings format to new format
	if err := c.migrateSettings(); err != nil {
		return nil, err
	}

	// Read the current settings.
	currentSettings, err := settings.ReadSettings()
	if err != nil {
		return nil, err
	}

	// Now, we can set the current profile pointer.
	if err := currentSettings.SetCurrentProfile(c.CurrentProfileName); err != nil {
		return nil, err
	}

	c.Logger.Debug("loaded settings", "settings", currentSettings)

	return currentSettings, nil
}

// Close closes any pending resources.
func (c *BaseCommand) Close() error {
	var errs []error

	if closer, ok := c.UI.(io.Closer); ok {
		errs = append(errs, closer.Close())
	}
	if c.grpcClient != nil {
		errs = append(errs, c.grpcClient.Close())
	}

	return errors.Join(errs...)
}

/*
Below methods are to be removed once deprecation is done.
These methods help us preserve the command / option behavior from
the former Graphql-driven CLI / SDK.
*/

// migrateSettings migrates old Tharsis settings format to new format.
// Remove this migration once deprecation is done.
func (c *BaseCommand) migrateSettings() error {
	settingsPath, err := settings.DefaultSettingsFilepath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No settings to migrate
		}
		return err
	}

	var oldFormat struct {
		Profiles map[string]struct {
			TharsisURL string `json:"tharsisURL"`
			StateToken string `json:"stateToken"`
		} `json:"profiles"`
	}

	if err := json.Unmarshal(data, &oldFormat); err != nil {
		return nil // Not old format or corrupted, skip
	}

	// Check if any profile has tharsisURL field (old format)
	needsMigration := false
	for _, profile := range oldFormat.Profiles {
		if profile.TharsisURL != "" {
			needsMigration = true
			break
		}
	}

	if !needsMigration {
		return nil
	}

	c.UI.Output("Migrating settings to new format...")

	// Backup the original settings file before migrating.
	backupPath := settingsPath + ".bak"
	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return fmt.Errorf("failed to create settings backup: %w", err)
	}

	c.UI.Output("  Backup saved to %s", backupPath)

	// Migrate to new format
	newProfiles := make(map[string]settings.Profile)
	tokens := make(map[string]string)

	for name, oldProfile := range oldFormat.Profiles {
		newProfiles[name] = settings.Profile{
			Endpoint: oldProfile.TharsisURL,
		}
		if oldProfile.StateToken != "" {
			tokens[name] = oldProfile.StateToken
		}
	}

	newSettings, err := settings.NewSettings()
	if err != nil {
		return err
	}

	newSettings.Profiles = newProfiles

	// Write migrated settings
	if err := newSettings.WriteSettingsFile(); err != nil {
		return err
	}

	// Write tokens to credentials file if any exist
	if len(tokens) > 0 {
		credsPath, err := settings.DefaultCredentialsFilepath()
		if err != nil {
			return err
		}

		credsData := struct {
			StateTokens map[string]string `json:"stateTokens"`
		}{
			StateTokens: tokens,
		}

		credsJSON, err := json.MarshalIndent(credsData, "", "  ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(credsPath, credsJSON, 0600); err != nil {
			return err
		}
	}

	c.Logger.Info("migrated settings from old Tharsis format to new format")
	c.UI.Output("Settings migration complete.")
	return nil
}

// extractParentPath returns the parent and child from a given path.
// This is deprecated. Remove after users are on the latest CLI.
func extractParentPath(p string) (parent, child string) {
	if index := strings.LastIndex(p, "/"); index != -1 {
		return p[:index], p[index+1:]
	}

	return "", p
}

// parseSortField converts a sort string to an enum value, handling deprecated separate sort-by and sort-order flags.
// Remove after users are on the latest CLI.
func parseSortField[T ~int32](sortBy, sortOrder *string, enumValues map[string]int32) (*T, error) {
	if sortBy == nil {
		return nil, nil
	}

	// Normalize deprecated short names.
	deprecatedAliases := map[string]string{
		"CREATED": "CREATED_AT",
		"UPDATED": "UPDATED_AT",
		"PATH":    "FULL_PATH",
	}

	key := *sortBy
	if alias, ok := deprecatedAliases[key]; ok {
		key = alias
	}

	// Try direct lookup first (new format: FIELD_ORDER).
	sort, ok := enumValues[key]
	if ok {
		enumVal := T(sort)
		return &enumVal, nil
	}

	// Handle deprecated separate sort-by and sort-order flags.
	if sortOrder == nil {
		return nil, fmt.Errorf("sort order must be specified if using deprecated sort-by value %s", *sortBy)
	}

	sortValue := fmt.Sprintf("%s_%s", key, *sortOrder)
	sort, ok = enumValues[sortValue]
	if !ok {
		return nil, fmt.Errorf("unknown sort value %s", sortValue)
	}

	enumVal := T(sort)
	return &enumVal, nil
}
