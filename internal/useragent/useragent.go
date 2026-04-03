// Package useragent provides user agent functionality.
package useragent

import (
	"fmt"
	"net/http"
	"runtime"
)

// BuildUserAgent creates a user agent string for the Tharsis CLI.
func BuildUserAgent(version string) string {
	return fmt.Sprintf("tharsis-cli/%s (%s; %s)", version, runtime.GOOS, runtime.GOARCH)
}

// Transport wraps an http.RoundTripper to add the User-Agent header.
type Transport struct {
	UserAgent string
	Base      http.RoundTripper
}

// RoundTrip implements http.RoundTripper.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", t.UserAgent)
	return t.Base.RoundTrip(req)
}
