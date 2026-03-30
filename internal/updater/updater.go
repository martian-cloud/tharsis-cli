// Package updater fetches the latest stable CLI release from GitLab and
// reports whether a newer version is available.
package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	goversion "github.com/hashicorp/go-version"
)

const (
	packagesURL  = "https://gitlab.com/api/v4/projects/infor-cloud%2Fmartian-cloud%2Ftharsis%2Ftharsis-cli/packages?sort=desc&per_page=20"
	downloadBase = "https://gitlab.com/api/v4/projects/infor-cloud%2Fmartian-cloud%2Ftharsis%2Ftharsis-cli/packages/generic/tharsis-cli"
	timeout      = 3 * time.Second
)

// supportedPlatforms is the set of OS/arch combinations published as packages.
var supportedPlatforms = map[string]bool{
	"darwin/amd64":  true,
	"darwin/arm64":  true,
	"freebsd/386":   true,
	"freebsd/amd64": true,
	"freebsd/arm":   true,
	"linux/386":     true,
	"linux/amd64":   true,
	"linux/arm":     true,
	"linux/arm64":   true,
	"openbsd/386":   true,
	"openbsd/amd64": true,
	"solaris/amd64": true,
	"windows/386":   true,
	"windows/amd64": true,
}

// Status represents the outcome of an update check.
type Status int

const (
	// StatusUnknown means the check failed or was skipped (e.g. dev build).
	StatusUnknown Status = iota
	// StatusUpToDate means the current version is the latest stable release.
	StatusUpToDate
	// StatusUpdateAvailable means a newer stable release exists.
	StatusUpdateAvailable
)

// Result holds the outcome of an update check.
type Result struct {
	// Status is the outcome of the check.
	Status Status
	// Latest is the latest stable version. Empty when Status is StatusUnknown.
	Latest string
	// DownloadURL is the direct download link for the current platform.
	// Empty when Status is not StatusUpdateAvailable.
	DownloadURL string
}

// Check fetches the latest stable release and compares it against current.
// Returns an empty Result if the request fails or times out.
func Check(current string) Result {
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: timeout,
			DisableKeepAlives:     true,
		},
	}

	ch := make(chan Result, 1)
	go func() { ch <- fetch(current, packagesURL, client) }()

	select {
	case r := <-ch:
		return r
	case <-time.After(timeout):
		return Result{}
	}
}

func fetch(current, url string, client *http.Client) Result {
	currentVer, err := goversion.NewVersion(current)
	if err != nil || currentVer.Prerelease() != "" {
		// Dev/dirty builds — skip update check.
		return Result{}
	}

	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return Result{}
	}
	defer resp.Body.Close()

	var packages []struct {
		ID      int    `json:"id"`
		Version string `json:"version"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&packages); err != nil {
		return Result{}
	}

	// Find the highest-ID stable (non-pre-release) version.
	var latestVer *goversion.Version
	latestID := -1
	for _, p := range packages {
		v, err := goversion.NewVersion(p.Version)
		if err != nil || v.Prerelease() != "" {
			continue
		}

		if p.ID > latestID {
			latestID = p.ID
			latestVer = v
		}
	}

	if latestVer == nil {
		return Result{}
	}

	latest := latestVer.Original()
	if !strings.HasPrefix(latest, "v") {
		latest = "v" + latest
	}

	platform := runtime.GOOS + "/" + runtime.GOARCH
	if !supportedPlatforms[platform] {
		return Result{}
	}

	downloadURL := fmt.Sprintf("%s/%s/tharsis_%s_%s_%s", downloadBase, latest, latest, runtime.GOOS, runtime.GOARCH)

	if latestVer.GreaterThan(currentVer) {
		return Result{
			Status:      StatusUpdateAvailable,
			Latest:      latest,
			DownloadURL: downloadURL,
		}
	}

	return Result{Status: StatusUpToDate, Latest: latest}
}
