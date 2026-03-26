// Package main contains the necessary functions for
// building the help menu and configuring the CLI
// library with all subcommand routes.
package main

import (
	"context"
	"io"
	"os"
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
)

const (
	// logLevelEnvVar is the environment variable for setting the log level.
	logLevelEnvVar = "THARSIS_CLI_LOG"
	// profileEnvVar is the environment variable for setting the active profile.
	profileEnvVar = "THARSIS_PROFILE"
	// noColorEnvVar is the environment variable for disabling colored output.
	noColorEnvVar = "NO_COLOR"
	// binaryName is the CLI binary name.
	binaryName = "tharsis"
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

func main() {
	os.Exit(realMain())
}

// Facilitate testing the main function by wrapping it.
// Now, a test can call realMain without having the os.Exit call getting in the way.
func realMain() int {
	var (
		// rawArgs are the arguments passed to the binary.
		rawArgs = os.Args[1:]
		// profileName is the name of the profile to use.
		profileName *string
		// noColor if true disables the coloring of terminal output
		noColor *bool
		// logLevel controls the verbosity of log output
		logLevel *string
		// logLevels are the valid log levels from hclog
		logLevels = []string{
			hclog.Off.String(),
			hclog.Trace.String(),
			hclog.Debug.String(),
			hclog.Info.String(),
			hclog.Warn.String(),
			hclog.Error.String(),
		}
		// autocompleteFlagInstall is the flag name to install autocomplete
		autocompleteFlagInstall = "enable-autocomplete"
		// autocompleteFlagUninstall is the flag name to uninstall autocomplete
		autocompleteFlagUninstall = "disable-autocomplete"
	)

	globalFlags := flag.NewSet("Global options")
	globalFlags.SetOutput(io.Discard)

	globalFlags.StringVar(
		&profileName,
		"p",
		"Profile to use from the configuration file.",
		flag.Default(settings.DefaultProfileName),
		flag.EnvVar(profileEnvVar),
		flag.Aliases("profile"),
	)

	globalFlags.StringVar(
		&logLevel,
		"log",
		"Set the verbosity of log output for debugging.",
		flag.Default(hclog.Off.String()),
		flag.EnvVar(logLevelEnvVar),
		flag.ValidValues(logLevels...),
		flag.PredictValues(logLevels...),
	)

	globalFlags.BoolVar(
		&noColor,
		"no-color",
		"Disable colored output.",
		flag.Default(false),
		flag.EnvVar(noColorEnvVar),
	)

	// Registered as informational for help/docs display only.
	globalFlags.Informational("v", "Show the version information.", flag.Aliases("version"))
	globalFlags.Informational("h", "Show this usage message.", flag.Aliases("help"))
	globalFlags.Informational(autocompleteFlagInstall, "Install shell autocompletion.")
	globalFlags.Informational(autocompleteFlagUninstall, "Uninstall shell autocompletion.")

	// Parse global flags. Unknown flags (handled by CLI framework) fall through.
	commandArgs := rawArgs
	if err := globalFlags.Parse(rawArgs); err == nil {
		commandArgs = globalFlags.Args()
	}

	if *noColor {
		color.NoColor = true
	}

	log := hclog.New(&hclog.LoggerOptions{
		Name:              binaryName,
		Level:             hclog.LevelFromString(*logLevel),
		Output:            os.Stderr, // Send logs to stderr
		Color:             hclog.AutoColor,
		DisableTime:       true,
		IndependentLevels: true,
	})

	hclog.SetDefault(log)

	log.Debug("",
		"version", Version,
		"binary_name", binaryName,
		"arguments", rawArgs,
		"profile_name", *profileName,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userAgent := useragent.BuildUserAgent(Version)

	// Create HTTP client with retry logic
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.Logger = log
	retryClient.RetryWaitMin = 10 * time.Second
	retryClient.RetryWaitMax = 60 * time.Second
	httpClient := retryClient.StandardClient()

	// Add User-Agent header to all requests
	httpClient.Transport = &useragent.Transport{
		UserAgent: userAgent,
		Base:      httpClient.Transport,
	}

	baseCommand := &command.BaseCommand{
		Context:              ctx,
		BinaryName:           binaryName,
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

	availableCommands, err := commands(baseCommand, globalFlags)
	if err != nil {
		log.Error(err.Error())
		return 1
	}

	c := cli.CLI{
		Name:        binaryName,
		Version:     Version,
		Args:        commandArgs,
		Commands:    availableCommands,
		HelpFunc:    helpFunc(cli.BasicHelpFunc(binaryName), globalFlags),
		HelpWriter:  os.Stdout,
		ErrorWriter: os.Stderr,

		// Shell autocompletion via posener/complete.
		Autocomplete:          true,
		AutocompleteInstall:   autocompleteFlagInstall,
		AutocompleteUninstall: autocompleteFlagUninstall,

		// nil prevents our global flags (-p, -log, -no-color) from appearing
		// on subcommand completions. They're only relevant before the subcommand.
		AutocompleteGlobalFlags: nil,

		// false lets mitchellh/cli add its own flags (-help, -version, etc.)
		// to the root command's Flags map, keeping them root-only.
		AutocompleteNoDefaultFlags: false,
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Error(err.Error())
		return 1
	}

	return exitStatus
}
