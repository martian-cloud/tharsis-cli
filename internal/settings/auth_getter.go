package settings

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/smithy-go/ptr"
	"github.com/qiangxue/go-env"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
)

type tokenGetterOptions struct {
	ServiceAccountToken string `env:"SERVICE_ACCOUNT_TOKEN,secret"`
	ServiceAccountID    string `env:"SERVICE_ACCOUNT_ID"`
	StaticToken         string `env:"STATIC_TOKEN,secret"`
}

func (c *tokenGetterOptions) load() error {
	// Environment variables override load options.
	if err := env.New("THARSIS_", log.Printf).Load(c); err != nil {
		return fmt.Errorf("failed to load env variables: %w", err)
	}

	return nil
}

func createTokenGetter(
	ctx context.Context,
	defaultStaticToken *string,
	httpEndpoint string,
	tlsSkipVerify bool,
) (client.TokenGetter, error) {
	c := &tokenGetterOptions{
		StaticToken: ptr.ToString(defaultStaticToken),
	}

	if err := c.load(); err != nil {
		return nil, err
	}

	if c.ServiceAccountID != "" && c.ServiceAccountToken != "" {
		// A service account token getter from environment variables.
		serviceAccountGetter, err := newServiceAccountTokenGetter(
			ctx,
			c.ServiceAccountID,
			httpEndpoint,
			tlsSkipVerify,
			func() ([]byte, error) {
				return []byte(c.ServiceAccountToken), nil
			},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to obtain a service account token getter: %w", err)
		}

		return serviceAccountGetter, nil
	}

	if c.StaticToken != "" {
		// A static token getter from an environment variable.
		staticGetter, err := newStaticTokenGetter(c.StaticToken)
		if err != nil {
			return nil, fmt.Errorf("failed to obtain a static token getter: %w", err)
		}

		return staticGetter, nil
	}

	return nil, fmt.Errorf("missing authentication credentials: either use tharsis sso login to get a token or set the required environment variables: " +
		"THARSIS_STATIC_TOKEN environment variable is used to supply a static token: " +
		"THARSIS_SERVICE_ACCOUNT_ID and THARSIS_SERVICE_ACCOUNT_TOKEN environment variables are required to login using a service account",
	)
}
