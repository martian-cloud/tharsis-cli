package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticTokenResolver(t *testing.T) {
	type testCase struct {
		name string
		run  func(t *testing.T)
	}

	testCases := []testCase{
		{
			name: "valid token",
			run: func(t *testing.T) {
				getter, err := newStaticTokenResolver(func() (string, error) {
					return "my-token", nil
				})
				require.NoError(t, err)

				token, err := getter.Token(t.Context())
				require.NoError(t, err)
				assert.Equal(t, "my-token", token)
			},
		},
		{
			name: "empty token",
			run: func(t *testing.T) {
				_, err := newStaticTokenResolver(func() (string, error) {
					return "", nil
				})
				assert.ErrorContains(t, err, "authentication token is empty")
			},
		},
		{
			name: "func called on each Token invocation",
			run: func(t *testing.T) {
				callCount := 0
				getter, err := newStaticTokenResolver(func() (string, error) {
					callCount++
					return "token", nil
				})
				require.NoError(t, err)

				_, _ = getter.Token(t.Context())
				_, _ = getter.Token(t.Context())

				// 1 call at construction + 2 calls from Token().
				assert.Equal(t, 3, callCount)
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, test.run)
	}
}

func TestTokenResolver(t *testing.T) {
	type testCase struct {
		name string
		run  func(t *testing.T)
	}

	testCases := []testCase{
		{
			name: "static token",
			run: func(t *testing.T) {
				tr := &tokenResolver{StaticToken: "preset-token"}

				getter, err := tr.resolve(t.Context(), "https://example.com", false, nil)
				require.NoError(t, err)

				token, err := getter.Token(t.Context())
				require.NoError(t, err)
				assert.Equal(t, "preset-token", token)
			},
		},
		{
			name: "uses staticTokenFunc when env does not override",
			run: func(t *testing.T) {
				tr := &tokenResolver{StaticToken: "file-token"}

				funcCalled := false
				getter, err := tr.resolve(t.Context(), "https://example.com", false, func() (string, error) {
					funcCalled = true
					return "file-token", nil
				})
				require.NoError(t, err)

				token, err := getter.Token(t.Context())
				require.NoError(t, err)
				assert.Equal(t, "file-token", token)
				assert.True(t, funcCalled)
			},
		},
		{
			name: "no credentials",
			run: func(t *testing.T) {
				tr := &tokenResolver{}

				_, err := tr.resolve(t.Context(), "https://example.com", false, nil)
				assert.ErrorContains(t, err, "missing authentication credentials")
			},
		},
		{
			name: "service account ID and path conflict",
			run: func(t *testing.T) {
				tr := &tokenResolver{
					ServiceAccountID:    "sa-id",
					ServiceAccountPath:  "sa-path",
					ServiceAccountToken: "sa-token",
				}

				_, err := tr.resolve(t.Context(), "https://example.com", false, nil)
				assert.ErrorContains(t, err, "THARSIS_SERVICE_ACCOUNT_ID and THARSIS_SERVICE_ACCOUNT_PATH cannot both be set")
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, test.run)
	}
}
