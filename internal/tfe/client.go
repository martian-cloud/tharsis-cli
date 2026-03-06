// Package tfe provides a REST client for interacting with the upstream Terraform-compatible API.
package tfe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/hashicorp/go-slug"
)

const (
	tarFlagWrite = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	tarMode      = 0600
	tfeV2Service = "tfe.v2"
)

// tokenGetter provides authentication tokens
type tokenGetter interface {
	Token(ctx context.Context) (string, error)
}

// RESTClient handles REST API calls to the upstream Terraform-compatible API
type RESTClient struct {
	endpoint    string
	tokenGetter tokenGetter
	httpClient  *http.Client
}

// NewClient creates a new REST client for interacting with the upstream Terraform REST API
func NewClient(endpoint string, tokenGetter tokenGetter, httpClient *http.Client) *RESTClient {
	return &RESTClient{
		endpoint:    endpoint,
		tokenGetter: tokenGetter,
		httpClient:  httpClient,
	}
}

// UploadConfigurationVersion uploads a directory as a tar.gz file
func (c *RESTClient) UploadConfigurationVersion(ctx context.Context, workspaceID, configVersionID, directoryPath string) error {
	// Discover TFE services
	serviceDiscovery, err := DiscoverTFEServices(ctx, c.httpClient, c.endpoint)
	if err != nil {
		return fmt.Errorf("failed to discover TFE services: %w", err)
	}

	// Get TFE v2 endpoint from service discovery
	tfeEndpoint, err := serviceDiscovery.GetServiceURL(tfeV2Service)
	if err != nil {
		return err
	}

	tarPath, err := c.makeTarFile(directoryPath)
	if err != nil {
		return err
	}
	defer os.Remove(tarPath)

	stat, err := os.Stat(tarPath)
	if err != nil {
		return err
	}

	tarRdr, err := os.Open(tarPath) // nosemgrep: gosec.G304-1
	if err != nil {
		return err
	}
	defer tarRdr.Close()

	uploadURL, err := url.JoinPath(tfeEndpoint, "workspaces", workspaceID, "configuration-versions", configVersionID, "upload")
	if err != nil {
		return fmt.Errorf("failed to construct upload URL: %w", err)
	}

	return c.doPut(ctx, uploadURL, tarRdr, stat.Size())
}

func (c *RESTClient) makeTarFile(dirPath string) (string, error) {
	stat, err := os.Stat(dirPath)
	if err != nil {
		return "", err
	}
	if !stat.IsDir() {
		return "", fmt.Errorf("not a directory: %s", dirPath)
	}

	tarFile, err := os.CreateTemp("", "uploadConfigurationVersion.tgz")
	if err != nil {
		return "", err
	}
	tarPath := tarFile.Name()

	tgzFileWriter, err := os.OpenFile(tarPath, tarFlagWrite, tarMode)
	if err != nil {
		return "", err
	}
	defer tgzFileWriter.Close()

	if _, err = slug.Pack(dirPath, tgzFileWriter, true); err != nil {
		return "", err
	}

	return tarPath, nil
}

func (c *RESTClient) doPut(ctx context.Context, url string, body io.Reader, length int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, body)
	if err != nil {
		return err
	}

	authToken, err := c.tokenGetter.Token(ctx)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = length

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
