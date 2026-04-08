package terminal

import (
	"github.com/fatih/color"
)

// Color constants for table entries.
const (
	Yellow = "yellow"
	Green  = "green"
	Red    = "red"
)

// colorizeTableEntry returns the entry value with ANSI color applied when color
// output is enabled and the entry has a recognized color name set.
func colorizeTableEntry(ent TableEntry) string {
	if color.NoColor || ent.Color == "" {
		return ent.Value
	}
	switch ent.Color {
	case Green:
		return color.GreenString(ent.Value)
	case Yellow:
		return color.YellowString(ent.Value)
	case Red:
		return color.RedString(ent.Value)
	default:
		return ent.Value
	}
}

// Table is passed to UI.Table to provide a nicely formatted table.
type Table struct {
	Headers []string
	Rows    [][]TableEntry
}

// NewTable creates a new Table structure that can be used with UI.Table.
func NewTable(headers ...string) *Table {
	return &Table{
		Headers: headers,
	}
}

// TableEntry is a single entry for a table.
type TableEntry struct {
	Value string
	Color string
}

// Rich adds a row to the table.
func (t *Table) Rich(cols []string, colors []string) {
	var row []TableEntry

	for i, col := range cols {
		if i < len(colors) {
			row = append(row, TableEntry{Value: col, Color: colors[i]})
		} else {
			row = append(row, TableEntry{Value: col})
		}
	}

	t.Rows = append(t.Rows, row)
}
