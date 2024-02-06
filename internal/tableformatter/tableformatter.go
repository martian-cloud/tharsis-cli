// Package tableformatter configures the tablewriter
// library to display data in a tabled format used
// by various CLI commands.
package tableformatter

import (
	"strings"

	"github.com/olekukonko/tablewriter"
)

// This module wraps https://github.com/olekukonko/tablewriter to format list
// output as a table, formatting per example 10 at that site--no borders or
// internal separators, all left aligned.  The intent is to imitate the
// output format of the docker CLI.

// FormatTable formats a 2D array of strings and returns a (potentially long)
// string of the output.  Input is in row-major order.  The first row of input
// is taken as the header and is forced to uppercase.  This function does
// _NOT_ sanitize the input.
func FormatTable(input [][]string) string {
	result := &strings.Builder{}
	table := tablewriter.NewWriter(result)

	table.SetHeader(forceUpper(input[0])) // use row 0 as the header
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)
	table.AppendBulk(input[1:]) // add the rest of the data

	table.Render()
	return result.String()
}

// forceUpper returns a slice of strings converted to uppercase.
func forceUpper(input []string) []string {
	result := make([]string, len(input))

	for ix, s := range input {
		result[ix] = strings.ToUpper(s)
	}

	return result
}
