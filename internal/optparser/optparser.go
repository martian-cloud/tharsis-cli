// Package optparser provides the logic for
// parsing flags for any given command. It
// also verifies any required option(s).
package optparser

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

// The optparser module parses the rest of a command line after the CLI
// platform has done what it does.  An option starts with one or two hyphens:
//
//   -option=optarg
//
//   -option optarg
//
// An argument is anything that does not start with a hyphen.  All options
// must appear before any arguments.
//
// Arguments are kept in the order they were received.  Options are put into a
// map, so their order is not preserved.
//
// The exception to some of the above is that global options must precede the
// first non-option command or argument.

// OptionDefinition holds the definition of one option.
//
// The length of the Arguments slice tells how many arguments the option takes.
// The names are used when printing options in the usage message.
//
// The stringVal and boolVal fields are for internal use only.
//
// When the need arises for options of numeric or other non-string types, a field for
// the type of argument value should be added.
type OptionDefinition struct {
	Synopsis  string
	Arguments []string
	stringVal arrayVal
	Required  bool
	boolVal   bool
}

// arrayVal contains values from duplicates flags in one slice.
// Allows passing in multiple variables.
type arrayVal []string

func (a *arrayVal) String() string {
	return ""
}

func (a *arrayVal) Set(value string) error {
	*a = append(*a, strings.TrimSpace(value))
	return nil
}

// OptionDefinitions holds the definitions of multiple options.
type OptionDefinitions map[string]*OptionDefinition // must be pointer

// ParseCommandOptions behaves either as ReadGlobalOptions or as ParseCommandOptions.
// At present, there is no difference in behavior between the two.
func ParseCommandOptions(headline string, definitions OptionDefinitions,
	words []string) (map[string][]string, []string, error) {

	// Set up for flag parsing based on the definitions.
	flags := flag.NewFlagSet(headline, flag.ContinueOnError)
	for optName, optInfo := range definitions {
		if len(optInfo.Arguments) > 0 {
			flags.Var(&optInfo.stringVal, optName, optInfo.Synopsis)
		} else {
			flags.BoolVar(&optInfo.boolVal, optName, false, optInfo.Synopsis)
		}
	}

	// Do the parsing.
	flags.SetOutput(io.Discard) // Don't need flag library to output messages when we have our own.
	err := flags.Parse(words)
	if err != nil {
		return nil, nil, err
	}
	arguments := flags.Args()

	// Extract the values.
	options := map[string][]string{}
	for optName, optInfo := range definitions {
		if len(optInfo.Arguments) > 0 {
			val := optInfo.stringVal
			if len(val) > 0 {
				options[optName] = val
			}
		} else if optInfo.boolVal {
			options[optName] = []string{"1"}
		}
	}

	// Return an error if any required option was missing.
	if err := verifyRequiredOptions(definitions, options); err != nil {
		return nil, nil, err
	}

	return options, arguments, nil
}

// verifyRequiredOptions returns an error if any required option is missing.
func verifyRequiredOptions(definitions OptionDefinitions, options map[string][]string) error {
	missing := []string{}
	for optName, optInfo := range definitions {
		if optInfo.Required {
			// Must check this one.
			if _, ok := options[optName]; !ok {
				missing = append(missing, "--"+optName)
			}
		}
	}

	if len(missing) == 1 {
		return fmt.Errorf("required option was not supplied: %s", missing[0])
	} else if len(missing) > 0 {
		return fmt.Errorf("required options were not supplied: %s", strings.Join(missing, ", "))
	}

	return nil
}
