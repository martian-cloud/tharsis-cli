package settings

import (
	"context"
	"fmt"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
)

// staticTokenResolver wraps a token-fetching function as a client.TokenGetter.
// It's a func type rather than a struct so that the caller controls whether
// the token is a fixed value or re-read from disk on each call.
type staticTokenResolver func() (string, error)

func (f staticTokenResolver) Token(_ context.Context) (string, error) {
	return f()
}

// newStaticTokenResolver validates the token is non-empty at construction
// time to fail fast rather than on the first API call.
func newStaticTokenResolver(tokenFunc func() (string, error)) (client.TokenGetter, error) {
	token, err := tokenFunc()
	if err != nil {
		return nil, err
	}

	if token == "" {
		return nil, fmt.Errorf("authentication token is empty: run 'tharsis sso login' or set the THARSIS_STATIC_TOKEN environment variable")
	}

	return staticTokenResolver(tokenFunc), nil
}
