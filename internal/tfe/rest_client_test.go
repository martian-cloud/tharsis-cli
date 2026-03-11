package tfe

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTokenGetter struct {
	token string
	err   error
}

func (m *mockTokenGetter) Token(_ context.Context) (string, error) {
	return m.token, m.err
}

func TestNewRESTClient(t *testing.T) {
	client, err := NewRESTClient("https://example.com", &mockTokenGetter{token: "test"}, http.DefaultClient)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestDownloadConfigurationVersion(t *testing.T) {
	type testCase struct {
		name         string
		setupServer  func() *httptest.Server
		expectError  bool
		expectOutput string
	}

	testCases := []testCase{
		{
			name: "successful download",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)
					assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte("test content"))
					require.NoError(t, err)
				}))
			},
			expectOutput: "test content",
		},
		{
			name: "download fails",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := tc.setupServer()
			defer server.Close()

			client, err := NewRESTClient(server.URL, &mockTokenGetter{token: "test-token"}, http.DefaultClient)
			require.NoError(t, err)

			var buf bytes.Buffer
			err = client.DownloadConfigurationVersion(t.Context(), &DownloadConfigurationVersionInput{
				ConfigVersionID: "cv-123",
				Writer:          &buf,
			})

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectOutput, buf.String())
		})
	}
}

func TestUploadModuleVersion(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "module.tar.gz")
	require.NoError(t, os.WriteFile(testFile, []byte("module content"), 0600))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, "module content", string(body))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewRESTClient(server.URL, &mockTokenGetter{token: "test-token"}, http.DefaultClient)
	require.NoError(t, err)

	err = client.UploadModuleVersion(t.Context(), &UploadModuleVersionInput{
		ModuleVersionID: "mv-123",
		PackagePath:     testFile,
	})
	require.NoError(t, err)
}

func TestUploadProviderReadme(t *testing.T) {
	tmpDir := t.TempDir()
	readmeFile := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(readmeFile, []byte("# Provider"), 0600))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, "# Provider", string(body))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewRESTClient(server.URL, &mockTokenGetter{token: "test-token"}, http.DefaultClient)
	require.NoError(t, err)

	err = client.UploadProviderReadme(t.Context(), &UploadProviderReadmeInput{
		ProviderVersionID: "pv-123",
		ReadmePath:        readmeFile,
	})
	require.NoError(t, err)
}

func TestUploadProviderChecksums(t *testing.T) {
	tmpDir := t.TempDir()
	checksumsFile := filepath.Join(tmpDir, "checksums.txt")
	require.NoError(t, os.WriteFile(checksumsFile, []byte("checksums"), 0600))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewRESTClient(server.URL, &mockTokenGetter{token: "test-token"}, http.DefaultClient)
	require.NoError(t, err)

	err = client.UploadProviderChecksums(t.Context(), &UploadProviderChecksumsInput{
		ProviderVersionID: "pv-123",
		ChecksumsPath:     checksumsFile,
	})
	require.NoError(t, err)
}

func TestUploadProviderChecksumSignature(t *testing.T) {
	tmpDir := t.TempDir()
	sigFile := filepath.Join(tmpDir, "checksums.sig")
	require.NoError(t, os.WriteFile(sigFile, []byte("signature"), 0600))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewRESTClient(server.URL, &mockTokenGetter{token: "test-token"}, http.DefaultClient)
	require.NoError(t, err)

	err = client.UploadProviderChecksumSignature(t.Context(), &UploadProviderChecksumSignatureInput{
		ProviderVersionID: "pv-123",
		SignaturePath:     sigFile,
	})
	require.NoError(t, err)
}

func TestUploadProviderPlatformBinary(t *testing.T) {
	tmpDir := t.TempDir()
	binaryFile := filepath.Join(tmpDir, "provider")
	require.NoError(t, os.WriteFile(binaryFile, []byte("binary"), 0600))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewRESTClient(server.URL, &mockTokenGetter{token: "test-token"}, http.DefaultClient)
	require.NoError(t, err)

	err = client.UploadProviderPlatformBinary(t.Context(), &UploadProviderPlatformBinaryInput{
		PlatformID: "plat-123",
		BinaryPath: binaryFile,
	})
	require.NoError(t, err)
}

func TestUploadProviderPlatformPackageToMirror(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, "package data", string(body))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewRESTClient(server.URL, &mockTokenGetter{token: "test-token"}, http.DefaultClient)
	require.NoError(t, err)

	err = client.UploadProviderPlatformPackageToMirror(t.Context(), &UploadProviderPlatformPackageToMirrorInput{
		VersionMirrorID: "vm-123",
		OS:              "linux",
		Arch:            "amd64",
		Reader:          bytes.NewReader([]byte("package data")),
	})
	require.NoError(t, err)
}
