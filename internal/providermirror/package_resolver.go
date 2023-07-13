// Package providermirror contains functionalities related to resolving
// provider packages for the Tharsis Provider Network Mirror.
package providermirror

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/apparentlymart/go-versions/versions"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/logger"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// zipContentType represents the allowed mime types when downloading a zip archive.
var zipContentType = []string{
	"application/octet-stream", // Some providers have this.
	"application/x-zip-compressed",
	"application/zip",
}

// ListVersionsResponse is the response returned from the Terraform Registry API
// when querying for supported versions for a provider.
// https://developer.hashicorp.com/terraform/internals/provider-registry-protocol#list-available-versions
type ListVersionsResponse struct {
	Versions []struct {
		Version   string `json:"version"`
		Platforms []struct {
			OS   string `json:"os"`
			Arch string `json:"arch"`
		} `json:"platforms"`
	} `json:"versions"`
	Warnings []string `json:"warnings"`
}

// PackageQueryResponse is the response returned when querying for a particular installation package.
// https://developer.hashicorp.com/terraform/internals/provider-registry-protocol#find-a-provider-package
type PackageQueryResponse struct {
	DownloadURL string `json:"download_url"` // Only used to find the provider download URL.
}

// TerraformProviderPackageResolver encapsulates the logic to resolve and download provider
// packages from the Terraform Registry.
type TerraformProviderPackageResolver struct {
	httpClient *http.Client
	logger     logger.Logger
	serviceURL *url.URL
}

// NewTerraformProviderPackageResolver returns an instance of TerraformProviderPackageResolver.
func NewTerraformProviderPackageResolver(
	logger logger.Logger,
	serviceURL *url.URL,
	httpClient *http.Client,
) *TerraformProviderPackageResolver {
	return &TerraformProviderPackageResolver{
		httpClient: httpClient,
		logger:     logger,
		serviceURL: serviceURL,
	}
}

// ListAvailableProviderVersions lists the available provider versions and platforms
// they support by contacting the Terraform Registry API the provider is associated with.
func (m *TerraformProviderPackageResolver) ListAvailableProviderVersions(
	ctx context.Context,
	namespace,
	providerType string,
) (*ListVersionsResponse, error) {
	result, err := url.Parse(path.Join(namespace, providerType, "versions"))
	if err != nil {
		return nil, err
	}

	endpoint := m.serviceURL.ResolveReference(result)

	m.logger.Debugf("Resolved endpoint to list available provider versions: %s", endpoint)

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate http request: %w", err)
	}

	r.Header.Add("Accept", "application/json")

	resp, err := m.httpClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to perform http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the payload to get the available provider versions.
	var decodedBody ListVersionsResponse
	if err = json.NewDecoder(resp.Body).Decode(&decodedBody); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	// Make sure no warnings were returned.
	if len(decodedBody.Warnings) > 0 {
		return nil, fmt.Errorf("provider versions endpoint returned warnings: %s", strings.Join(decodedBody.Warnings, "; "))
	}

	if len(decodedBody.Versions) == 0 {
		return nil, errors.New("no provider versions found")
	}

	return &decodedBody, nil
}

// FindProviderPackage locates the provider package at the Terraform Registry API.
// It is only used to find the package download endpoint since it's generally
// different from the registry API.
func (m *TerraformProviderPackageResolver) FindProviderPackage(
	ctx context.Context,
	platform string,
	versionMirror *sdktypes.TerraformProviderVersionMirror,
) (*PackageQueryResponse, error) {
	// Get the os and arch separately.
	parts := strings.Split(platform, "_")

	// Build the URL to the provider's package URL so, we can find the download URL.
	// Generally, the download endpoint is at a different location than the API.
	p := path.Join(
		versionMirror.RegistryNamespace,
		versionMirror.Type,
		versionMirror.SemanticVersion,
		"download",
		parts[0],
		parts[1],
	)

	result, err := url.Parse(p)
	if err != nil {
		return nil, fmt.Errorf("failed to build package download URL: %w", err)
	}

	endpoint := m.serviceURL.ResolveReference(result)

	m.logger.Debugf("Resolved endpoint for locating provider package: %s", endpoint)

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate http request: %w", err)
	}

	r.Header.Add("Accept", "application/json")

	resp, err := m.httpClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to perform http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var foundResp PackageQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&foundResp); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &foundResp, nil
}

// DownloadProviderPlatformPackage downloads the actual provider package.
func (m *TerraformProviderPackageResolver) DownloadProviderPlatformPackage(
	ctx context.Context,
	endpoint,
	dir string,
) (string, error) {
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to parse download endpoint: %w", err)
	}

	m.logger.Debugf("Resolved endpoint for downloading provider package: %s", endpoint)

	packageFile, err := os.CreateTemp(dir, "terraform-provider-*.zip")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary zip file: %w", err)
	}
	defer packageFile.Close()

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate http request: %w", err)
	}

	resp, err := m.httpClient.Do(r)
	if err != nil {
		return "", fmt.Errorf("failed to perform http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Verify the mime type.
	mimeType := resp.Header.Get("content-type")
	if !isZipMimeType(mimeType) {
		return "", fmt.Errorf("unexpected mime type: expected %v, got %s", zipContentType, mimeType)
	}

	if _, err := io.Copy(packageFile, resp.Body); err != nil {
		return "", fmt.Errorf("failed to copy response to destination file: %w", err)
	}

	return packageFile.Name(), nil
}

// FindLatestVersion finds the latest version a provider supports.
func (m *TerraformProviderPackageResolver) FindLatestVersion(resp *ListVersionsResponse) (string, error) {
	versionsList := make(versions.List, len(resp.Versions))

	for ix, ver := range resp.Versions {
		// Parse the version to make sure it's valid.
		v, err := versions.ParseVersion(ver.Version)
		if err != nil {
			return "", fmt.Errorf("failed to parse provider version: %w", err)
		}

		versionsList[ix] = v
	}

	return versionsList.Newest().String(), nil
}

// FilterMissingPlatforms filters for platforms that are currently not mirrored.
func (m *TerraformProviderPackageResolver) FilterMissingPlatforms(
	targetVersion string,
	resp *ListVersionsResponse,
	existingPlatforms map[string]struct{},
) (map[string]struct{}, error) {
	missingPlatforms := map[string]struct{}{}

	target, err := versions.ParseVersion(targetVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target provider version: %w", err)
	}

	// Find the target version so, we can find missing platforms for it.
	for _, ver := range resp.Versions {
		// Parse the version to make sure it's valid and also for accurate comparison below.
		v, err := versions.ParseVersion(ver.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to parse provider version: %w", err)
		}

		if v.Same(target) {
			// Found the target version, now filter for platforms which are missing from the mirror.
			for _, p := range ver.Platforms {
				key := fmt.Sprintf("%s_%s", p.OS, p.Arch)
				if _, ok := existingPlatforms[key]; !ok {
					missingPlatforms[key] = struct{}{}
				}
			}

			// Don't need to keep looping.
			break
		}
	}

	return missingPlatforms, nil
}

// isZipMimeType verifies the mime type to be a zip equivalent.
func isZipMimeType(contentType string) bool {
	for _, mime := range zipContentType {
		if contentType == mime {
			return true
		}
	}

	return false
}
