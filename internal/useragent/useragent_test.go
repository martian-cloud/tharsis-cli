package useragent

import (
	"net/http"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildUserAgent(t *testing.T) {
	ua := BuildUserAgent("1.2.3")

	assert.Contains(t, ua, "tharsis-cli/1.2.3")
	assert.Contains(t, ua, runtime.GOOS)
	assert.Contains(t, ua, runtime.GOARCH)
}

func TestTransport(t *testing.T) {
	t.Run("sets user agent header", func(t *testing.T) {
		var captured *http.Request
		base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
			captured = req
			return &http.Response{StatusCode: 200}, nil
		})

		transport := &Transport{UserAgent: "test-agent/1.0", Base: base}
		req, err := http.NewRequest("GET", "http://example.com", nil)
		require.NoError(t, err)

		_, err = transport.RoundTrip(req)
		require.NoError(t, err)

		assert.Equal(t, "test-agent/1.0", captured.Header.Get("User-Agent"))
	})

	t.Run("overwrites existing user agent", func(t *testing.T) {
		base := roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200}, nil
		})

		transport := &Transport{UserAgent: "new-agent", Base: base}
		req, err := http.NewRequest("GET", "http://example.com", nil)
		require.NoError(t, err)
		req.Header.Set("User-Agent", "old-agent")

		_, err = transport.RoundTrip(req)
		require.NoError(t, err)

		assert.Equal(t, "new-agent", req.Header.Get("User-Agent"))
	})
}

// roundTripFunc adapts a function to http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
