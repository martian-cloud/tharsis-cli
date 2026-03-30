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

func serve(t *testing.T, packages []struct {
	ID      int    `json:"id"`
	Version string `json:"version"`
}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if err := json.NewEncoder(w).Encode(packages); err != nil {
			t.Fatal(err)
		}
	}))
}

func TestFetch(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		packages []struct {
			ID      int    `json:"id"`
			Version string `json:"version"`
		}
		expectLatest          string
		expectUpdateAvailable bool
	}{
		{
			name:    "update available",
			current: "0.29.0",
			packages: []struct {
				ID      int    `json:"id"`
				Version string `json:"version"`
			}{
				{ID: 2, Version: "v0.30.0"},
				{ID: 1, Version: "v0.29.0"},
			},
			expectLatest:          "v0.30.0",
			expectUpdateAvailable: true,
		},
		{
			name:    "already on latest",
			current: "0.30.0",
			packages: []struct {
				ID      int    `json:"id"`
				Version string `json:"version"`
			}{
				{ID: 1, Version: "v0.30.0"},
			},
			expectLatest:          "v0.30.0",
			expectUpdateAvailable: false,
		},
		{
			name:    "pre-releases skipped",
			current: "0.29.0",
			packages: []struct {
				ID      int    `json:"id"`
				Version string `json:"version"`
			}{
				{ID: 3, Version: "v0.31.0-alpha.1"},
				{ID: 2, Version: "v0.30.0"},
				{ID: 1, Version: "v0.29.0"},
			},
			expectLatest:          "v0.30.0",
			expectUpdateAvailable: true,
		},
		{
			name:    "highest ID wins over version order",
			current: "0.29.0",
			packages: []struct {
				ID      int    `json:"id"`
				Version string `json:"version"`
			}{
				{ID: 3, Version: "v0.30.0"},
				{ID: 5, Version: "v0.30.1"},
				{ID: 4, Version: "v0.30.2"},
			},
			expectLatest:          "v0.30.1",
			expectUpdateAvailable: true,
		},
		{
			name:    "dev build skips check",
			current: "v0.30.0-26-gf8a2b9d-dirty",
			packages: []struct {
				ID      int    `json:"id"`
				Version string `json:"version"`
			}{
				{ID: 1, Version: "v0.30.0"},
			},
			expectLatest:          "",
			expectUpdateAvailable: false,
		},
		{
			name:    "no stable packages",
			current: "0.29.0",
			packages: []struct {
				ID      int    `json:"id"`
				Version string `json:"version"`
			}{
				{ID: 1, Version: "v0.30.0-alpha.1"},
			},
			expectLatest:          "",
			expectUpdateAvailable: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := serve(t, tc.packages)
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

func TestFetchServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	result := fetch("0.29.0", srv.URL, srv.Client())
	assert.Equal(t, Result{}, result)
}
