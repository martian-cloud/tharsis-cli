// Package main contains the necessary functions for
// building the help menu and configuring the CLI
// library with all subcommand routes.
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
)

func helpFunc(commands map[string]cli.CommandFactory) string {
	longestName := 0
	sortedNames := []string{}

	// Find the longest command name.
	for cmdName := range commands {
		if len(cmdName) > longestName {
			longestName = len(cmdName)
		}
		sortedNames = append(sortedNames, cmdName)
	}

	// Sort the command names.
	sort.Strings(sortedNames)

	return strings.TrimSpace(fmt.Sprintf(`
Usage: %s [global options] <command> [arguments]

Commands:
%s

Global options (if any, must come before the first command):
%s
`, filepath.Base(os.Args[0]),
		showCommands(sortedNames, commands, longestName), showGlobalOptions(globalOptionNames)))
}

func showCommands(sortedNames []string, commands map[string]cli.CommandFactory, longest int) string {
	var buf []string

	for _, commandName := range sortedNames {
		commandFunc := commands[commandName]

		command, err := commandFunc()
		if err != nil {
			log.Printf("Command factory failed during help function: %s", commandName)
			continue
		}

		pad := strings.Repeat(" ", longest-len(commandName))
		buf = append(buf, fmt.Sprintf("%s%s  %s", commandName, pad, command.Synopsis()))
	}

	// No trailing newline.
	return strings.Join(buf, "\n")
}

// showGlobalOptions adds -list and -version to the list
func showGlobalOptions(options optparser.OptionDefinitions) string {
	var buf bytes.Buffer

	sortedNames, longest := sortGlobalOptionNames(options)
	for _, optNamePlus := range sortedNames {
		pad := strings.Repeat(" ", longest-len(optNamePlus))
		optName := strings.Split(optNamePlus, "=")[0] // least costly to just take off the =... again
		optSynopsis := options[optName].Synopsis
		buf.WriteString(fmt.Sprintf("-%s%s  %s\n", optNamePlus, pad, optSynopsis))
	}

	// Has a trailing newline.
	return buf.String()
}

func sortGlobalOptionNames(options optparser.OptionDefinitions) ([]string, int) {
	sorted := []string{}
	longest := 0

	for name := range options {
		var optArgs string
		if len(options[name].Arguments) > 0 {
			optArgs = "=" + strings.Join(options[name].Arguments, ",")
		}
		namePlus := name + optArgs
		if len(namePlus) > longest {
			longest = len(namePlus)
		}
		sorted = append(sorted, namePlus)
	}

	// Sort the command names.
	sort.Strings(sorted)

	return sorted, longest
}
