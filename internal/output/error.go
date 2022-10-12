// Package output is responsible for formatting
// the final output / errors the user sees in the
// CLI. It automatically parses any GraphQL errors
// from a given error and displays the 'detail'
// field if present.
package output

import (
	"encoding/json"
	"strings"

	"github.com/hasura/go-graphql-client"
)

const (
	bar           = "| "
	reset         = "\033[0m"
	redBar        = "\033[31m" + bar
	newline       = "\n"
	redBarNewline = redBar + newline
	flagProvided  = "flag provided"
)

// FormatError outputs the error with proper formatting for CLI.
func FormatError(summary string, err error) string {

	msg := formatError(summary, "")

	if err != nil {
		// Avoid duplicate err message when flag library already outputs one.
		// flag library will automatically output "flag provided but not defined"
		if strings.Contains(err.Error(), flagProvided) {
			return ""
		}

		errString := parseGraphQLErrors(err)
		return msg + formatError("", errString)
	}

	return msg + newline
}

// ParseGraphQLErrors parses the 'detail' field from a GraphQL Message body.
// If 'detail' is not there returns only the error's 'message' field content.
// If the JSON couldn't be parsed returns original message to caller.
func parseGraphQLErrors(err error) string {
	var m map[string]string

	// Check if error is a graphQL Errors type.
	anErr, ok := err.(graphql.Errors)
	if !ok {
		// Return original error.
		return err.Error()
	}

	// Parse the GraphQL error message.
	message := anErr[0].Message

	// Check if there is embedded JSON within message.
	// Return only message field if no JSON found.
	begin := strings.Index(message, "\"{")
	if begin == -1 {
		return message
	}

	message = strings.ReplaceAll(message[begin+1:len(message)-1], "\\", "")

	// Parse the 'detail' JSON message.
	_ = json.Unmarshal([]byte(message), &m)

	// If no JSON was parsed return original error back to caller.
	if len(m) == 0 {
		return err.Error()
	}

	return m["detail"]
}

func formatError(summary, err string) string {
	if summary != "" {
		return newline +
			redBarNewline +
			redBar +
			"Error: " +
			reset +
			summary +
			newline +
			redBarNewline +
			reset
	}

	return redBar +
		reset +
		err +
		newline +
		redBarNewline +
		reset
}
