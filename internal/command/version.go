package command

import (
	"errors"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/updater"
)

// versionCommand returns the remote API backend version this CLI connects to.
type versionCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*versionCommand)(nil)

func (c *versionCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
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

	update := updater.Check(c.Version)

	type versionOutput struct {
		CLI         string `json:"cli"`
		Latest      string `json:"latest,omitempty"`
		DownloadURL string `json:"download_url,omitempty"`
	}

	out := versionOutput{CLI: c.Version}
	if update.Status == updater.StatusUpdateAvailable {
		out.Latest = update.Latest
		out.DownloadURL = update.DownloadURL
	}

	if *c.toJSON {
		if err := c.UI.JSON(out); err != nil {
			c.UI.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		values := []terminal.NamedValue{
			{Name: "Version", Value: out.CLI},
		}

		switch update.Status {
		case updater.StatusUpdateAvailable:
			values = append(values,
				terminal.NamedValue{Name: "Latest", Value: update.Latest},
				terminal.NamedValue{Name: "Download", Value: update.DownloadURL},
			)
		case updater.StatusUpToDate:
			values = append(values, terminal.NamedValue{Name: "Status", Value: "up to date"})
		}

		c.UI.NamedValues(values)
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
  Returns the CLI's version.
`
}

func (c *versionCommand) Example() string {
	return `
tharsis version -json
`
}

func (c *versionCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
		flag.Default(false),
	)

	return f
}
