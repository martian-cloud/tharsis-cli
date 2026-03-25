// Package main contains the necessary functions for
// building the help menu and configuring the CLI
// library with all subcommand routes.
package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/command"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/settings"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/useragent"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	// Logger level environment variable
	logLevelEnvVar = "THARSIS_CLI_LOG"
	// Profile environment variable
	profileEnvVar = "THARSIS_PROFILE"
	// User-Agent header name
	userAgentHeader = "User-Agent"
)

var (
	// Version is passed in via ldflags at build time
	Version = "1.0.0"

	// DefaultHTTPEndpoint is passed in via ldflags at build time.
	DefaultHTTPEndpoint string

	// DefaultTLSSkipVerify is passed in via ldflags at build time.
	// Indicates if client should skip verifying the server's
	// certificate chain and domain name.
	DefaultTLSSkipVerify bool
)

// userAgentTransport wraps an http.RoundTripper to add User-Agent header
type userAgentTransport struct {
	userAgent string
	transport http.RoundTripper
}

// RoundTrip implements http.RoundTripper interface by adding User-Agent header to requests
func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set(userAgentHeader, t.userAgent)
	return t.transport.RoundTrip(req)
}

func main() {
	os.Exit(realMain())
}

// Facilitate testing the main function by wrapping it.
// Now, a test can call realMain without having the os.Exit call getting in the way.
func realMain() int {
	var (
		// binaryName is the name of the binary.
		binaryName = filepath.Base(os.Args[0])
		// displayTitle is the name of the binary in title case.
		displayTitle = cases.Title(language.English, cases.Compact).String(binaryName)
		// rawArgs are the arguments passed to the binary.
		rawArgs = os.Args[1:]
		// profileName is the name of the profile to use.
		profileName *string
		// noColor if true disables the coloring of terminal output
		noColor *bool
		// Autocomplete flag names.
		autocompleteFlagInstall   = "enable-autocomplete"
		autocompleteFlagUninstall = "disable-autocomplete"
	)

	// Create a global flagSet.
	globalFlags := flag.NewSet("Global options")
	globalFlags.SetOutput(io.Discard)

	// Default profile from env var, then fall back to default.
	defaultProfile := os.Getenv(profileEnvVar)
	if defaultProfile == "" {
		defaultProfile = settings.DefaultProfileName
	}

	globalFlags.StringVar(
		&profileName,
		"p",
		"Profile name from config file. Overrides THARSIS_PROFILE env var.",
		flag.Default(defaultProfile),
	)

	globalFlags.BoolVar(
		&noColor,
		"no-color",
		"Disable colored output. Also respects NO_COLOR env var.",
		flag.Default(os.Getenv("NO_COLOR") != ""),
	)

	// Set the no color option on the library.
	color.NoColor = *noColor

	// Values are never used since CLI framework can handle them,
	// these are simply meant to facilitate the help output for
	// available global flag.
	var s *string
	var b *bool
	globalFlags.StringVar(&s, "v", "Show the version information.")
	globalFlags.StringVar(&s, "h", "Show this usage message.")
	globalFlags.BoolVar(&b, autocompleteFlagInstall, "Install shell autocompletion.")
	globalFlags.BoolVar(&b, autocompleteFlagUninstall, "Uninstall shell autocompletion.")

	// Check for autocomplete flags - pass directly to CLI without parsing global flags
	isAutocomplete := false
	for _, arg := range rawArgs {
		if arg == "-"+autocompleteFlagInstall || arg == "-"+autocompleteFlagUninstall ||
			arg == "--"+autocompleteFlagInstall || arg == "--"+autocompleteFlagUninstall {
			isAutocomplete = true
			break
		}
	}

	// Only parse global flags if not an autocomplete request
	if !isAutocomplete {
		// Ignore errors since CLI framework will handle them.
		_ = globalFlags.Parse(rawArgs)

		// Apply no-color setting to the color package.
		if *noColor {
			color.NoColor = true
		}
	}

	logLevel := os.Getenv(logLevelEnvVar)
	if logLevel == "" {
		// if log level is not set, keep it off by default.
		logLevel = "off"
	}

	// Log the startup.
	log := hclog.New(&hclog.LoggerOptions{
		Name:              binaryName,
		Level:             hclog.LevelFromString(logLevel),
		Output:            os.Stderr, // Send logs to stderr
		Color:             hclog.AutoColor,
		DisableTime:       true,
		IndependentLevels: true,
	})

	hclog.SetDefault(log)

	log.Debug("",
		"version", Version,
		"binary_name", binaryName,
		"display_title", displayTitle,
		"arguments", rawArgs,
		"profile_name", *profileName,
	)

	// For any variation of "-h" or "-help", simply use "-h".
	// Since help option can be used for any command, we must
	// handle it the same anywhere.
	for ix, arg := range rawArgs {
		if arg == "--h" || arg == "--help" || arg == "-help" {
			rawArgs[ix] = "-h"
		}
	}

	// Only replace "--version" and "--v" at the global level i.e. the first argument.
	// Allows using the same argument in commands and subcommands.
	if len(rawArgs) > 0 && (rawArgs[0] == "--version" || rawArgs[0] == "-version" || rawArgs[0] == "--v") {
		rawArgs[0] = "-v"
	}

	// Get the CLI context.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Build user agent
	userAgent := useragent.BuildUserAgent(Version)

	// Create HTTP client with retry logic
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.Logger = log
	retryClient.RetryWaitMin = 10 * time.Second
	retryClient.RetryWaitMax = 60 * time.Second
	httpClient := retryClient.StandardClient()

	// Add User-Agent header to all requests
	httpClient.Transport = &userAgentTransport{
		userAgent: userAgent,
		transport: httpClient.Transport,
	}

	// Prepare the command metadata struct.
	baseCommand := &command.BaseCommand{
		Context:              ctx,
		BinaryName:           binaryName,
		DisplayTitle:         displayTitle,
		Version:              Version,
		Logger:               log,
		UI:                   terminal.ConsoleUI(ctx),
		CurrentProfileName:   *profileName,
		DefaultHTTPEndpoint:  DefaultHTTPEndpoint,
		DefaultTLSSkipVerify: DefaultTLSSkipVerify,
		UserAgent:            userAgent,
		HTTPClient:           httpClient,
	}

	// Defer closing the base command, so the UI output could finish rendering.
	defer baseCommand.Close()

	commandArgs := rawArgs // No global flag.
	if globalFlags.NFlag() > 0 {
		// A global option was set, so pass remaining arguments to commands,
		// since it will automatically be handled.
		commandArgs = globalFlags.Args()
	}

	availableCommands, err := commands(baseCommand)
	if err != nil {
		log.Error(err.Error())
		return 1
	}

	c := cli.CLI{
		Name:                    binaryName,
		Version:                 Version,
		Args:                    commandArgs,
		Commands:                availableCommands,
		HelpFunc:                helpFunc(cli.BasicHelpFunc(binaryName), globalFlags),
		HelpWriter:              os.Stdout,
		ErrorWriter:             os.Stderr,
		Autocomplete:            true,
		AutocompleteInstall:     autocompleteFlagInstall,
		AutocompleteUninstall:   autocompleteFlagUninstall,
		AutocompleteGlobalFlags: globalAutocompletions(globalFlags),
	}

	// Run the CLI.
	exitStatus, err := c.Run()
	if err != nil {
		log.Error(err.Error())
		return 1
	}

	return exitStatus
}
