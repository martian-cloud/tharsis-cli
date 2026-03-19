package settings

import (
	"context"
	"fmt"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
)

// staticTokenGetter adapts a func() (string, error) to the client.TokenGetter interface.
type staticTokenGetter func() (string, error)

func (f staticTokenGetter) Token(_ context.Context) (string, error) {
	return f()
}

// newStaticTokenGetter returns a TokenGetter that calls tokenFunc on each invocation.
// It validates the token is non-empty at construction time.
func newStaticTokenGetter(tokenFunc func() (string, error)) (client.TokenGetter, error) {
	token, err := tokenFunc()
	if err != nil {
		return nil, err
	}

	if token == "" {
		return nil, fmt.Errorf("static token was empty")
	}

	return staticTokenGetter(tokenFunc), nil
}
