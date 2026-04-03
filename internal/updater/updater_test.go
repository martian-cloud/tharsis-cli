package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetch(t *testing.T) {
	type pkg struct {
		Version string `json:"version"`
	}

	tests := []struct {
		name                  string
		current               string
		packages              []pkg
		serverStatus          int
		expectLatest          string
		expectUpdateAvailable bool
	}{
		{
			name:                  "update available",
			current:               "0.29.0",
			packages:              []pkg{{Version: "v0.30.0"}, {Version: "v0.29.0"}},
			expectLatest:          "v0.30.0",
			expectUpdateAvailable: true,
		},
		{
			name:                  "already on latest",
			current:               "0.30.0",
			packages:              []pkg{{Version: "v0.30.0"}},
			expectLatest:          "v0.30.0",
			expectUpdateAvailable: false,
		},
		{
			name:                  "pre-releases skipped",
			current:               "0.29.0",
			packages:              []pkg{{Version: "v0.31.0-alpha.1"}, {Version: "v0.30.0"}, {Version: "v0.29.0"}},
			expectLatest:          "v0.30.0",
			expectUpdateAvailable: true,
		},
		{
			name:                  "highest semver wins over package order",
			current:               "0.29.0",
			packages:              []pkg{{Version: "v0.30.0"}, {Version: "v0.30.1"}, {Version: "v0.30.2"}},
			expectLatest:          "v0.30.2",
			expectUpdateAvailable: true,
		},
		{
			name:                  "dev build skips check",
			current:               "v0.30.0-26-gf8a2b9d-dirty",
			packages:              []pkg{{Version: "v0.30.0"}},
			expectLatest:          "",
			expectUpdateAvailable: false,
		},
		{
			name:                  "no stable packages",
			current:               "0.29.0",
			packages:              []pkg{{Version: "v0.30.0-alpha.1"}},
			expectLatest:          "",
			expectUpdateAvailable: false,
		},
		{
			name:                  "server error returns empty result",
			current:               "0.29.0",
			serverStatus:          http.StatusInternalServerError,
			expectLatest:          "",
			expectUpdateAvailable: false,
		},
		{
			name:                  "empty package list",
			current:               "0.29.0",
			packages:              []pkg{},
			expectLatest:          "",
			expectUpdateAvailable: false,
		},
		{
			name:                  "current version ahead of latest",
			current:               "0.31.0",
			packages:              []pkg{{Version: "v0.30.0"}},
			expectLatest:          "v0.30.0",
			expectUpdateAvailable: false,
		},
		{
			name:                  "invalid version strings ignored",
			current:               "0.29.0",
			packages:              []pkg{{Version: "not-a-version"}, {Version: "v0.30.0"}},
			expectLatest:          "v0.30.0",
			expectUpdateAvailable: true,
		},
		{
			name:                  "all invalid versions",
			current:               "0.29.0",
			packages:              []pkg{{Version: "abc"}, {Version: "xyz"}},
			expectLatest:          "",
			expectUpdateAvailable: false,
		},
		{
			name:                  "version without v prefix",
			current:               "0.29.0",
			packages:              []pkg{{Version: "0.30.0"}},
			expectLatest:          "v0.30.0",
			expectUpdateAvailable: true,
		},
		{
			name:                  "current equals latest with v prefix mismatch",
			current:               "v0.30.0",
			packages:              []pkg{{Version: "0.30.0"}},
			expectLatest:          "v0.30.0",
			expectUpdateAvailable: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tc.serverStatus != 0 {
					w.WriteHeader(tc.serverStatus)
					return
				}

				if err := json.NewEncoder(w).Encode(tc.packages); err != nil {
					t.Fatal(err)
				}
			}))
			defer srv.Close()

			result := fetch(tc.current, srv.URL, srv.Client())

			assert.Equal(t, tc.expectLatest, result.Latest)
			assert.Equal(t, tc.expectUpdateAvailable, result.Status == StatusUpdateAvailable)

			if tc.expectUpdateAvailable {
				expected := fmt.Sprintf("%s/%s/tharsis_%s_%s_%s",
					downloadBase, tc.expectLatest, tc.expectLatest, runtime.GOOS, runtime.GOARCH)
				assert.Equal(t, expected, result.DownloadURL)
			} else {
				assert.Empty(t, result.DownloadURL)
			}
		})
	}
}
