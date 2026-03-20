package settings

import (
	"context"
	"fmt"
	"log"

	"github.com/qiangxue/go-env"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// tokenResolver picks the right token strategy based on environment variables.
// Priority: service account > static token > error.
// Environment variables always override values set on the struct so CI/CD
// pipelines can inject credentials without modifying the settings file.
type tokenResolver struct {
	ServiceAccountToken string `env:"SERVICE_ACCOUNT_TOKEN,secret"`
	ServiceAccountID    string `env:"SERVICE_ACCOUNT_ID"`
	ServiceAccountPath  string `env:"SERVICE_ACCOUNT_PATH"`
	StaticToken         string `env:"STATIC_TOKEN,secret"`
}

// resolve returns the appropriate client.TokenGetter. staticTokenFunc is
// used when the static token was not overridden by an environment variable,
// allowing the caller to control how the token is fetched (e.g. re-reading
// from the credentials file for long-lived processes).
func (tr *tokenResolver) resolve(
	ctx context.Context,
	httpEndpoint string,
	tlsSkipVerify bool,
	staticTokenFunc func() (string, error),
) (client.TokenGetter, error) {
	// Snapshot before env loading so we can detect if THARSIS_STATIC_TOKEN
	// overrode the default value from the credentials file.
	defaultToken := tr.StaticToken

	if err := tr.loadEnv(); err != nil {
		return nil, err
	}

	if tr.ServiceAccountID != "" && tr.ServiceAccountPath != "" {
		return nil, fmt.Errorf("THARSIS_SERVICE_ACCOUNT_ID and THARSIS_SERVICE_ACCOUNT_PATH cannot both be set")
	}

	// SERVICE_ACCOUNT_PATH is deprecated; convert to TRN for backwards compatibility.
	if tr.ServiceAccountPath != "" {
		tr.ServiceAccountID = trn.NewResourceTRN(trn.ResourceTypeServiceAccount, tr.ServiceAccountPath)
	}

	if tr.ServiceAccountID != "" && tr.ServiceAccountToken != "" {
		return newServiceAccountTokenResolver(
			ctx,
			tr.ServiceAccountID,
			httpEndpoint,
			tlsSkipVerify,
			func() ([]byte, error) {
				return []byte(tr.ServiceAccountToken), nil
			},
		)
	}

	if tr.StaticToken != "" {
		// If the env var didn't override the default, use staticTokenFunc
		// to re-read from the credentials file on each call.
		if staticTokenFunc != nil && tr.StaticToken == defaultToken {
			return newStaticTokenResolver(staticTokenFunc)
		}

		staticToken := tr.StaticToken
		return newStaticTokenResolver(func() (string, error) { return staticToken, nil })
	}

	return nil, fmt.Errorf("missing authentication credentials: either use tharsis sso login to get a token or set the required environment variables: " +
		"THARSIS_STATIC_TOKEN environment variable is used to supply a static token: " +
		"THARSIS_SERVICE_ACCOUNT_ID and THARSIS_SERVICE_ACCOUNT_TOKEN environment variables are required to login using a service account",
	)
}

func (tr *tokenResolver) loadEnv() error {
	if err := env.New("THARSIS_", log.Printf).Load(tr); err != nil {
		return fmt.Errorf("failed to load env variables: %w", err)
	}

	return nil
}
