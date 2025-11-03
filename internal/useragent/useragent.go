// Package useragent provides user agent functionality.
package useragent

import (
	"fmt"
	"runtime"
)

// BuildUserAgent creates a user agent string for the Tharsis CLI
func BuildUserAgent(version string) string {
	return fmt.Sprintf("tharsis-cli/%s (%s; %s)", version, runtime.GOOS, runtime.GOARCH)
}
