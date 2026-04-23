// Package tfe provides a REST client for interacting with the upstream Terraform-compatible API.
package tfe

//go:generate go tool mockery --name RESTClient --inpackage --case underscore

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/provider"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/slug"
)

const (
	contentTypeOctetStream = "application/octet-stream"
)

var _ RESTClient = (*restClient)(nil)

// UploadConfigurationVersionInput is the input for uploading a configuration version
type UploadConfigurationVersionInput struct {
	WorkspaceID     string
	ConfigVersionID string
	DirectoryPath   string
}

// DownloadConfigurationVersionInput is the input for downloading a configuration version
type DownloadConfigurationVersionInput struct {
	ConfigVersionID string
	Writer          io.Writer
}

// UploadModuleVersionInput is the input for uploading a module version
type UploadModuleVersionInput struct {
	ModuleVersionID string
	PackagePath     string
}

// UploadProviderReadmeInput is the input for uploading a provider README
type UploadProviderReadmeInput struct {
	ProviderVersionID string
	ReadmePath        string
}

// UploadProviderChecksumsInput is the input for uploading provider checksums
type UploadProviderChecksumsInput struct {
	ProviderVersionID string
	ChecksumsPath     string
}

// UploadProviderChecksumSignatureInput is the input for uploading provider checksum signature
type UploadProviderChecksumSignatureInput struct {
	ProviderVersionID string
	SignaturePath     string
}

// UploadProviderPlatformBinaryInput is the input for uploading a provider platform binary
type UploadProviderPlatformBinaryInput struct {
	PlatformID string
	BinaryPath string
}

// UploadProviderPlatformPackageToMirrorInput is the input for uploading a provider platform package to mirror
type UploadProviderPlatformPackageToMirrorInput struct {
	VersionMirrorID string
	OS              string
	Arch            string
	Reader          io.Reader
}

// tokenGetter provides authentication tokens
type tokenGetter interface {
	Token(ctx context.Context) (string, error)
}

// RESTClient is the interface for REST client operations
type RESTClient interface {
	UploadConfigurationVersion(ctx context.Context, input *UploadConfigurationVersionInput) error
	DownloadConfigurationVersion(ctx context.Context, input *DownloadConfigurationVersionInput) error
	UploadModuleVersion(ctx context.Context, input *UploadModuleVersionInput) error
	UploadProviderReadme(ctx context.Context, input *UploadProviderReadmeInput) error
	UploadProviderChecksums(ctx context.Context, input *UploadProviderChecksumsInput) error
	UploadProviderChecksumSignature(ctx context.Context, input *UploadProviderChecksumSignatureInput) error
	UploadProviderPlatformBinary(ctx context.Context, input *UploadProviderPlatformBinaryInput) error
	UploadProviderPlatformPackageToMirror(ctx context.Context, input *UploadProviderPlatformPackageToMirrorInput) error
}

// restClient handles REST API calls to the upstream Terraform-compatible API
type restClient struct {
	baseURL           *url.URL
	tokenGetter       tokenGetter
	httpClient        *http.Client
	serviceDiscoverer provider.ServiceDiscoverer
}

// NewRESTClient creates a new REST client for interacting with the upstream Terraform REST API
func NewRESTClient(endpoint string, tokenGetter tokenGetter, httpClient *http.Client) (RESTClient, error) {
	baseURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint: %w", err)
	}

	return &restClient{
		baseURL:           baseURL,
		tokenGetter:       tokenGetter,
		httpClient:        httpClient,
		serviceDiscoverer: provider.NewServiceDiscoverer(httpClient),
	}, nil
}

// UploadConfigurationVersion uploads a directory as a tar.gz file
func (c *restClient) UploadConfigurationVersion(ctx context.Context, input *UploadConfigurationVersionInput) error {
	discovered, err := c.serviceDiscoverer.DiscoverTFEServices(ctx, c.baseURL.String())
	if err != nil {
		return fmt.Errorf("failed to discover tfe v2 service: %w", err)
	}

	serviceURL, ok := discovered.Services[provider.TFEServiceID]
	if !ok {
		return fmt.Errorf("service url for %q not found", provider.TFEServiceID)
	}

	s, err := slug.NewSlug(input.DirectoryPath)
	if err != nil {
		return fmt.Errorf("failed to create slug: %w", err)
	}
	defer os.Remove(s.SlugPath)

	// Get actual file size
	fileInfo, err := os.Stat(s.SlugPath)
	if err != nil {
		return fmt.Errorf("failed to stat slug file: %w", err)
	}

	reader, err := s.Open()
	if err != nil {
		return fmt.Errorf("failed to open slug: %w", err)
	}
	defer reader.Close()

	uploadURL := serviceURL.JoinPath("workspaces", input.WorkspaceID, "configuration-versions", input.ConfigVersionID, "upload").String()

	return c.doPut(ctx, uploadURL, reader, fileInfo.Size())
}

// DownloadConfigurationVersion downloads a configuration version
func (c *restClient) DownloadConfigurationVersion(ctx context.Context, input *DownloadConfigurationVersionInput) error {
	discovered, err := c.serviceDiscoverer.DiscoverTFEServices(ctx, c.baseURL.String())
	if err != nil {
		return fmt.Errorf("failed to discover tfe v2 service: %w", err)
	}

	serviceURL, ok := discovered.Services[provider.TFEServiceID]
	if !ok {
		return fmt.Errorf("service url for %q not found", provider.TFEServiceID)
	}

	downloadURL := serviceURL.JoinPath("configuration-versions", input.ConfigVersionID, "content").String()
	return c.doGet(ctx, downloadURL, input.Writer, contentTypeOctetStream)
}

// UploadModuleVersion uploads a module version package
func (c *restClient) UploadModuleVersion(ctx context.Context, input *UploadModuleVersionInput) error {
	stat, err := os.Stat(input.PackagePath)
	if err != nil {
		return err
	}

	reader, err := os.Open(input.PackagePath) // nosemgrep: gosec.G304-1
	if err != nil {
		return err
	}
	defer reader.Close()

	uploadURL := c.baseURL.JoinPath("v1", "module-registry", "versions", input.ModuleVersionID, "upload").String()

	return c.doPut(ctx, uploadURL, reader, stat.Size())
}

// UploadProviderReadme uploads a provider README file
func (c *restClient) UploadProviderReadme(ctx context.Context, input *UploadProviderReadmeInput) error {
	reader, err := os.Open(input.ReadmePath) // nosemgrep: gosec.G304-1
	if err != nil {
		return err
	}
	defer reader.Close()

	stat, err := reader.Stat()
	if err != nil {
		return err
	}

	uploadURL := c.baseURL.JoinPath("v1", "provider-registry", "versions", input.ProviderVersionID, "readme", "upload").String()

	return c.doPut(ctx, uploadURL, reader, stat.Size())
}

// UploadProviderChecksums uploads provider checksums file
func (c *restClient) UploadProviderChecksums(ctx context.Context, input *UploadProviderChecksumsInput) error {
	reader, err := os.Open(input.ChecksumsPath) // nosemgrep: gosec.G304-1
	if err != nil {
		return err
	}
	defer reader.Close()

	stat, err := reader.Stat()
	if err != nil {
		return err
	}

	uploadURL := c.baseURL.JoinPath("v1", "provider-registry", "versions", input.ProviderVersionID, "checksums", "upload").String()

	return c.doPut(ctx, uploadURL, reader, stat.Size())
}

// UploadProviderChecksumSignature uploads provider checksum signature file
func (c *restClient) UploadProviderChecksumSignature(ctx context.Context, input *UploadProviderChecksumSignatureInput) error {
	reader, err := os.Open(input.SignaturePath) // nosemgrep: gosec.G304-1
	if err != nil {
		return err
	}
	defer reader.Close()

	stat, err := reader.Stat()
	if err != nil {
		return err
	}

	uploadURL := c.baseURL.JoinPath("v1", "provider-registry", "versions", input.ProviderVersionID, "checksums-signature", "upload").String()

	return c.doPut(ctx, uploadURL, reader, stat.Size())
}

// UploadProviderPlatformBinary uploads a provider platform binary
func (c *restClient) UploadProviderPlatformBinary(ctx context.Context, input *UploadProviderPlatformBinaryInput) error {
	reader, err := os.Open(input.BinaryPath) // nosemgrep: gosec.G304-1
	if err != nil {
		return err
	}
	defer reader.Close()

	stat, err := reader.Stat()
	if err != nil {
		return err
	}

	uploadURL := c.baseURL.JoinPath("v1", "provider-registry", "platforms", input.PlatformID, "upload").String()

	return c.doPut(ctx, uploadURL, reader, stat.Size())
}

// UploadProviderPlatformPackageToMirror uploads a provider platform package to mirror
func (c *restClient) UploadProviderPlatformPackageToMirror(ctx context.Context, input *UploadProviderPlatformPackageToMirrorInput) error {
	uploadURL := c.baseURL.JoinPath("v1", "provider-mirror", "providers", input.VersionMirrorID, input.OS, input.Arch, "upload").String()

	return c.doPut(ctx, uploadURL, input.Reader, -1)
}

func (c *restClient) doPut(ctx context.Context, url string, body io.Reader, length int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, body)
	if err != nil {
		return err
	}

	authToken, err := c.tokenGetter.Token(ctx)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", contentTypeOctetStream)

	if length >= 0 {
		req.ContentLength = length
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("upload failed with status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *restClient) doGet(ctx context.Context, url string, writer io.Writer, accept string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	authToken, err := c.tokenGetter.Token(ctx)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)

	if accept != "" {
		req.Header.Set("Accept", accept)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("download failed with status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	_, err = io.Copy(writer, resp.Body)
	return err
}
