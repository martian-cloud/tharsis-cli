package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// versionCommand returns the remote API backend version this CLI connects to.
type versionCommand struct {
	*BaseCommand

	toJSON bool
}

var _ Command = (*versionCommand)(nil)

func (c *versionCommand) validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments, validation.Empty),
	)
}

// NewVersionCommandFactory returns an instance of versionCommand.
func NewVersionCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &versionCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *versionCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("version"),
		WithInputValidator(c.validate),
	); code != 0 {
		return code
	}

	version := struct {
		CLI string `json:"cli"`
	}{
		CLI: c.Version,
	}

	if c.toJSON {
		if err := c.UI.JSON(version); err != nil {
			c.UI.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		c.UI.Output(version.CLI)
	}

	return 0
}

func (c *versionCommand) Synopsis() string {
	return "Get the CLI's version."
}

func (c *versionCommand) Usage() string {
	return "tharsis [global options] version"
}

func (c *versionCommand) Description() string {
	return `
  The tharsis version command returns the CLI's version.
`
}

func (c *versionCommand) Example() string {
	return `
tharsis version --json
`
}

func (c *versionCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}
