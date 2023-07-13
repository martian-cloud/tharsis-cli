package providermirror

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/logger"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

type fakeRestError struct {
	Error string `json:"error"`
}

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(fn),
	}
}

func TestListAvailableProviderVersions(t *testing.T) {
	type testCase struct {
		payloadToReturn any
		expectError     error
		name            string
		statusToReturn  int
	}

	testCases := []testCase{
		{
			name: "successful list provider versions response",
			payloadToReturn: &ListVersionsResponse{
				Versions: []struct {
					Version   string `json:"version"`
					Platforms []struct {
						OS   string `json:"os"`
						Arch string `json:"arch"`
					} `json:"platforms"`
				}{
					{
						Version: "0.1.0",
						Platforms: []struct {
							OS   string `json:"os"`
							Arch string `json:"arch"`
						}{
							{
								OS:   "windows",
								Arch: "amd64",
							},
							{
								OS:   "linux",
								Arch: "arm",
							},
						},
					},
				},
				Warnings: []string(nil),
			},
			statusToReturn: http.StatusOK,
		},
		{
			name: "payload returned warnings",
			payloadToReturn: ListVersionsResponse{
				Warnings: []string{
					"some warning occurred",
				},
			},
			statusToReturn: http.StatusOK,
			expectError:    errors.New("provider versions endpoint returned warnings: some warning occurred"),
		},
		{
			name:            "failed to query for provider versions",
			payloadToReturn: fakeRestError{Error: "version doesn't exist"},
			statusToReturn:  http.StatusNotFound,
			expectError:     errors.New("unexpected status code: 404"),
		},
		{
			name:            "no provider versions found",
			payloadToReturn: &ListVersionsResponse{},
			expectError:     errors.New("no provider versions found"),
			statusToReturn:  http.StatusOK,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			payloadBuf, err := json.Marshal(test.payloadToReturn)
			require.Nil(t, err)

			httpClient := newTestClient(func(req *http.Request) *http.Response {
				// Verify URL is correct.
				assert.Equal(t, req.URL.String(), "http://test/hashicorp/aws/versions")

				return &http.Response{
					StatusCode: test.statusToReturn,
					Body:       io.NopCloser(bytes.NewReader(payloadBuf)),
					Header:     make(http.Header),
				}
			})

			l, _ := logger.NewForTest()

			serviceURL, err := url.Parse("http://test/")
			require.Nil(t, err)

			resolver := NewTerraformProviderPackageResolver(l, serviceURL, httpClient)

			response, err := resolver.ListAvailableProviderVersions(ctx, "hashicorp", "aws")

			if test.expectError != nil {
				assert.Equal(t, test.expectError, err)
				return
			}

			assert.Equal(t, test.payloadToReturn, response)
		})
	}
}

func TestFindProviderPackage(t *testing.T) {
	type testCase struct {
		payloadToReturn any
		expectError     error
		name            string
		statusToReturn  int
	}

	testCases := []testCase{
		{
			name: "successfully locate a provider package",
			payloadToReturn: &PackageQueryResponse{
				DownloadURL: "http://download",
			},
			statusToReturn: http.StatusOK,
		},
		{
			name:            "provider package does not exist",
			payloadToReturn: fakeRestError{Error: "not found"},
			expectError:     errors.New("unexpected status code: 404"),
			statusToReturn:  http.StatusNotFound,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			payloadBuf, err := json.Marshal(test.payloadToReturn)
			require.Nil(t, err)

			httpClient := newTestClient(func(req *http.Request) *http.Response {
				// Verify URL is correct.
				assert.Equal(t, req.URL.String(), "http://test/hashicorp/aws/1.0.0/download/linux/amd64")

				return &http.Response{
					StatusCode: test.statusToReturn,
					Body:       io.NopCloser(bytes.NewReader(payloadBuf)),
					Header:     make(http.Header),
				}
			})

			l, _ := logger.NewForTest()

			serviceURL, err := url.Parse("http://test/")
			require.Nil(t, err)

			resolver := NewTerraformProviderPackageResolver(l, serviceURL, httpClient)

			response, err := resolver.FindProviderPackage(ctx, "linux_amd64", &types.TerraformProviderVersionMirror{
				SemanticVersion:   "1.0.0",
				RegistryNamespace: "hashicorp",
				Type:              "aws",
			})

			if test.expectError != nil {
				assert.Equal(t, test.expectError, err)
				return
			}

			assert.Equal(t, test.payloadToReturn, response)
		})
	}
}

func TestDownloadProviderPlatformPackage(t *testing.T) {
	type testCase struct {
		expectError    error
		name           string
		contentType    string
		statusToReturn int
	}

	testCases := []testCase{
		{
			name:           "successfully download package",
			contentType:    "application/zip",
			statusToReturn: http.StatusOK,
		},
		{
			name:           "provider package file not found",
			statusToReturn: http.StatusNotFound,
			contentType:    "application/zip",
			expectError:    errors.New("unexpected status code: 404"),
		},
		{
			name:           "content type is not zip",
			contentType:    "application/json",
			expectError:    fmt.Errorf("unexpected mime type: expected %v, got application/json", zipContentType),
			statusToReturn: http.StatusOK,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			httpClient := newTestClient(func(_ *http.Request) *http.Response {
				// Add content type.
				header := make(http.Header)
				header.Add("content-type", test.contentType)

				return &http.Response{
					StatusCode: test.statusToReturn,
					Body:       io.NopCloser(strings.NewReader("package data")),
					Header:     header,
				}
			})

			l, _ := logger.NewForTest()

			serviceURL, err := url.Parse("http://test/")
			require.Nil(t, err)

			resolver := NewTerraformProviderPackageResolver(l, serviceURL, httpClient)

			// Create a temp dir to download the file into.
			tempDir, err := os.MkdirTemp("", "download-provider-test-*.zip")
			require.Nil(t, err)
			defer os.RemoveAll(tempDir)

			response, err := resolver.DownloadProviderPlatformPackage(ctx, "http://download", tempDir)

			if test.expectError != nil {
				assert.Equal(t, test.expectError, err)
				return
			}

			assert.NotEmpty(t, response)
		})
	}
}

func TestFindLatestVersion(t *testing.T) {
	type testCase struct {
		expectErrorMsg       string
		listVersionsResponse *ListVersionsResponse
		name                 string
		expectVersion        string
	}

	testCases := []testCase{
		{
			name: "found latest version 1.5.0",
			listVersionsResponse: &ListVersionsResponse{
				Versions: []struct {
					Version   string `json:"version"`
					Platforms []struct {
						OS   string `json:"os"`
						Arch string `json:"arch"`
					} `json:"platforms"`
				}{
					{
						Version: "0.2.0",
					},
					{
						Version: "1.4.0",
					},
					{
						Version: "1.5.0-pre",
					},
					{
						Version: "1.5.0",
					},
					{
						Version: "1.5.0-rc",
					},
				},
			},
			expectVersion: "1.5.0",
		},
		{
			name: "cannot parse version",
			listVersionsResponse: &ListVersionsResponse{
				Versions: []struct {
					Version   string `json:"version"`
					Platforms []struct {
						OS   string `json:"os"`
						Arch string `json:"arch"`
					} `json:"platforms"`
				}{
					{
						Version: "0.1.0",
					},
					{
						Version: "not parsable",
					},
				},
			},
			expectErrorMsg: "failed to parse provider version: invalid specification; required format is three positive integers separated by periods",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			resolver := NewTerraformProviderPackageResolver(nil, nil, nil)

			response, err := resolver.FindLatestVersion(test.listVersionsResponse)

			if test.expectErrorMsg != "" {
				assert.Equal(t, test.expectErrorMsg, err.Error())
				return
			}

			assert.Equal(t, test.expectVersion, response)
		})
	}
}

func TestFilterMissingPlatforms(t *testing.T) {
	type testCase struct {
		expectErrorMsg       string
		listVersionsResponse *ListVersionsResponse
		existingPlatforms    map[string]struct{}
		expectMissing        map[string]struct{}
		targetVersion        string
		name                 string
	}

	testCases := []testCase{
		{
			name:          "successfully return a map of missing platforms",
			targetVersion: "2.0.0",
			listVersionsResponse: &ListVersionsResponse{
				Versions: []struct {
					Version   string `json:"version"`
					Platforms []struct {
						OS   string `json:"os"`
						Arch string `json:"arch"`
					} `json:"platforms"`
				}{
					{
						Version: "0.1.0",
					},
					{
						Version: "2.0.0",
						Platforms: []struct {
							OS   string `json:"os"`
							Arch string `json:"arch"`
						}{
							{
								OS:   "windows",
								Arch: "amd64",
							},
							{
								OS:   "linux",
								Arch: "arm",
							},
						},
					},
				},
			},
			existingPlatforms: map[string]struct{}{
				"windows_amd64": {},
			},
			expectMissing: map[string]struct{}{
				"linux_arm": {},
			},
		},
		{
			name:          "version not found",
			targetVersion: "1.0.0",
			listVersionsResponse: &ListVersionsResponse{
				Versions: []struct {
					Version   string `json:"version"`
					Platforms []struct {
						OS   string `json:"os"`
						Arch string `json:"arch"`
					} `json:"platforms"`
				}{
					{
						Version: "0.1.0",
					},
				},
			},
			expectMissing: map[string]struct{}{},
		},
		{
			name:          "version is not parsable",
			targetVersion: "2.0.0",
			listVersionsResponse: &ListVersionsResponse{
				Versions: []struct {
					Version   string `json:"version"`
					Platforms []struct {
						OS   string `json:"os"`
						Arch string `json:"arch"`
					} `json:"platforms"`
				}{
					{
						Version: "not parsable",
					},
				},
			},
			expectErrorMsg: "failed to parse provider version: invalid specification; required format is three positive integers separated by periods",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			resolver := NewTerraformProviderPackageResolver(nil, nil, nil)

			response, err := resolver.FilterMissingPlatforms(test.targetVersion, test.listVersionsResponse, test.existingPlatforms)

			if test.expectErrorMsg != "" {
				assert.Equal(t, test.expectErrorMsg, err.Error())
				return
			}

			assert.Equal(t, test.expectMissing, response)
		})
	}
}
