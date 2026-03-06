// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// This file is from:
// https://github.com/hashicorp/waypoint/blob/ed4b99dbdd9378fdd96e213445aa446a27159816/internal/cli/help.go
// Some modifications have been made to meet Tharsis' use case.

package command

import (
	"flag"
	"strings"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
)

const helpTemplate = `
Usage: {{.Name}} {{.SubcommandName}} SUBCOMMAND

{{indent 2 (trim .Help)}}{{if gt (len .Subcommands) 0}}

Subcommands:
{{- range $value := .Subcommands }}
    {{ $value.NameAligned }}    {{ $value.Synopsis }}{{ end }}

{{- end }}
`

// helpCommand is the structure for the help command.
type helpCommand struct {
	synopsisText string
	helpText     string
}

// NewHelpCommandFactory returns a helpCommand struct.
func NewHelpCommandFactory(synopsisText, helpText string) func() (Command, error) {
	return func() (Command, error) {
		return &helpCommand{
			synopsisText: synopsisText,
			helpText:     helpText,
		}, nil
	}
}

func (c *helpCommand) Run(_ []string) int {
	return cli.RunResultHelp
}

func (c *helpCommand) Synopsis() string {
	return strings.TrimSpace(c.synopsisText)
}

func (c *helpCommand) Usage() string {
	return ""
}

func (c *helpCommand) Description() string {
	if c.helpText == "" {
		return c.synopsisText
	}

	return c.helpText
}

func (c *helpCommand) Example() string {
	return ""
}

func (c *helpCommand) Flags() *flag.FlagSet {
	return nil
}

func (c *helpCommand) HelpTemplate() string {
	return output.CommandHelp(output.CommandHelpInfo{Usage: helpTemplate})
}
