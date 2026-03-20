package settings

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticTokenResolver(t *testing.T) {
	type testCase struct {
		name           string
		token          string
		expectErrorMsg string
	}

	testCases := []testCase{
		{
			name:  "valid token",
			token: "my-token",
		},
		{
			name:           "empty token",
			token:          "",
			expectErrorMsg: "authentication token is empty: run 'tharsis sso login' or set the THARSIS_STATIC_TOKEN environment variable",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			getter, err := newStaticTokenResolver(func() (string, error) {
				return test.token, nil
			})

			if test.expectErrorMsg != "" {
				assert.EqualError(t, err, test.expectErrorMsg)
				return
			}

			require.NoError(t, err)

			token, err := getter.Token(t.Context())
			require.NoError(t, err)
			assert.Equal(t, test.token, token)
		})
	}
}

func TestStaticTokenResolverCallsFunc(t *testing.T) {
	// Verify the func is called on each Token() invocation,
	// not just at construction time.
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
}

func TestTokenResolverStaticToken(t *testing.T) {
	tr := &tokenResolver{
		StaticToken: "preset-token",
	}

	getter, err := tr.resolve(t.Context(), "https://example.com", false, nil)
	require.NoError(t, err)

	token, err := getter.Token(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "preset-token", token)
}

func TestTokenResolverUsesStaticTokenFunc(t *testing.T) {
	// When the env doesn't override StaticToken, the resolver should
	// use the provided staticTokenFunc for re-reading.
	tr := &tokenResolver{
		StaticToken: "file-token",
	}

	funcCalled := false
	staticFunc := func() (string, error) {
		funcCalled = true
		return "file-token", nil
	}

	getter, err := tr.resolve(t.Context(), "https://example.com", false, staticFunc)
	require.NoError(t, err)

	token, err := getter.Token(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "file-token", token)
	assert.True(t, funcCalled)
}

func TestTokenResolverNoCredentials(t *testing.T) {
	tr := &tokenResolver{}

	_, err := tr.resolve(t.Context(), "https://example.com", false, nil)
	assert.ErrorContains(t, err, "missing authentication credentials")
}

func TestTokenResolverServiceAccountIDAndPathConflict(t *testing.T) {
	tr := &tokenResolver{
		ServiceAccountID:    "sa-id",
		ServiceAccountPath:  "sa-path",
		ServiceAccountToken: "sa-token",
	}

	_, err := tr.resolve(t.Context(), "https://example.com", false, nil)
	assert.ErrorContains(t, err, "THARSIS_SERVICE_ACCOUNT_ID and THARSIS_SERVICE_ACCOUNT_PATH cannot both be set")
}
