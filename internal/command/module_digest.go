package command

import (
	"encoding/hex"
	"errors"
	"os"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/slug"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
)

type moduleDigestCommand struct {
	*BaseCommand

	directoryPath *string
	toJSON        *bool
}

var _ Command = (*moduleDigestCommand)(nil)

// NewModuleDigestCommandFactory returns a moduleDigestCommand struct.
func NewModuleDigestCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleDigestCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *moduleDigestCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

func (c *moduleDigestCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("module digest"),
		WithInputValidator(c.validate),
	); code != 0 {
		return code
	}

	s, err := slug.NewSlug(*c.directoryPath)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create module package")
		return 1
	}
	defer os.Remove(s.SlugPath)

	digest := hex.EncodeToString(s.SHASum)

	if *c.toJSON {
		if err := c.UI.JSON(struct {
			Digest string `json:"digest"`
		}{Digest: digest}); err != nil {
			c.UI.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		c.UI.NamedValues([]terminal.NamedValue{
			{Name: "Digest", Value: digest},
		})
	}

	return 0
}

func (*moduleDigestCommand) Synopsis() string {
	return "Compute the SHA256 digest for a module version package."
}

func (*moduleDigestCommand) Description() string {
	return `
   Packages the module directory and returns its SHA256
   digest. Useful for verifying deterministic builds or
   pre-computing the digest before uploading.
`
}

func (*moduleDigestCommand) Usage() string {
	return "tharsis [global options] module digest [options]"
}

func (*moduleDigestCommand) Example() string {
	return `
tharsis module digest -directory-path "./my-module"
tharsis module digest -directory-path "./my-module" -json
`
}

func (c *moduleDigestCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.directoryPath,
		"directory-path",
		"The path of the terraform module's directory.",
		flag.Default("."),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
		flag.Default(false),
	)

	return f
}
