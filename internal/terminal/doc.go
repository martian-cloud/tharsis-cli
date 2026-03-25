// Package terminal provides terminal UI abstractions for the Tharsis CLI.
//
// Originally licensed under MPL-2.0 from HashiCorp Waypoint Plugin SDK:
// https://github.com/hashicorp/waypoint-plugin-sdk/tree/dcdb2a03f7144a6e9a552351aeadd1791564f70e/terminal
//
// Modifications (not limited to):
//   - Removed glint dependency, using direct stdout writes for Output
//   - Added Successf, Errorf, ErrorWithSummary, Warnf, Infof, JSON, Close, Confirm methods to UI
//   - Removed WAYPOINT_FORCE_EMOJI, using LANG env var for UTF-8 detection
//   - Added RegisterStatus for custom status registration
//   - Added shell command auto-highlighting in examples
//   - Added StripAnsi for removing ANSI escape sequences
package terminal
