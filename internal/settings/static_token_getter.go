package settings

import (
	"context"
	"fmt"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
)

// Provides a static token, where the user supplies the static token at initialization time.

// staticTokenGetter implements TokenGetter.
type staticTokenGetter struct {
	token string
}

// newStaticTokenGetter returns a new instance of this getter.
func newStaticTokenGetter(token string) (client.TokenGetter, error) {
	if token == "" {
		return nil, fmt.Errorf("static token was empty")
	}

	staticGetter := staticTokenGetter{
		token: token,
	}
	var getter client.TokenGetter = &staticGetter
	return getter, nil
}

func (p *staticTokenGetter) Token(_ context.Context) (string, error) {
	return p.token, nil
}
