package settings

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

const (
	expirationLeeway = 30 * time.Second // seconds of leeway before expiration time
)

// oidcTokenGetter is a callback for returning the OIDC token that is used to login to the service account
type oidcTokenGetter func() ([]byte, error)

type tokenInfo struct {
	expires *time.Time
	token   string
	mutex   sync.RWMutex
}

// serviceAccountTokenGetter implements the client.TokenGetter interface.
type serviceAccountTokenGetter struct {
	tokenGetter oidcTokenGetter
	// The temporary/dynamic service account token, with expiration.
	// For thread safety, the token and its expiration are protected by a mutex.
	token            *tokenInfo
	client           *client.Client
	serviceAccountID string
}

// newServiceAccountTokenGetter returns a new instance of this getter.
func newServiceAccountTokenGetter(
	ctx context.Context,
	serviceAccountID string,
	httpEndpoint string,
	tlsSkipVerify bool,
	tokenGetter oidcTokenGetter,
) (client.TokenGetter, error) {
	// Create a client for the renewal of the service account token.
	c, err := client.New(ctx, &client.Config{
		HTTPEndpoint:  httpEndpoint,
		TLSSkipVerify: tlsSkipVerify,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &serviceAccountTokenGetter{
		tokenGetter:      tokenGetter,
		token:            &tokenInfo{},
		client:           c,
		serviceAccountID: serviceAccountID,
	}, nil
}

func (p *serviceAccountTokenGetter) Token(ctx context.Context) (string, error) {
	if p.isTokenExpired() {
		err := p.renewToken(ctx)
		if err != nil {
			return "", fmt.Errorf("service account token renewal failed: %w", err)
		}
	}

	p.token.mutex.RLock()
	defer p.token.mutex.RUnlock()
	return p.token.token, nil
}

// isTokenExpired returns true if a token was set but has expired, true if no token was ever set,
// and false if a token has been set and has not yet expired.
func (p *serviceAccountTokenGetter) isTokenExpired() bool {
	p.token.mutex.RLock()
	defer p.token.mutex.RUnlock()

	return p.token.expires == nil || !time.Now().Add(expirationLeeway).Before(*p.token.expires)
}

func (p *serviceAccountTokenGetter) renewToken(ctx context.Context) error {
	oidcToken, err := p.tokenGetter()
	if err != nil {
		return fmt.Errorf("failed to get OIDC token: %w", err)
	}

	// Get a new service account token.
	tokenResp, err := p.client.ServiceAccountsClient.CreateOIDCToken(ctx, &pb.CreateOIDCTokenRequest{
		ServiceAccountId: p.serviceAccountID,
		Token:            string(oidcToken),
	})
	if err != nil {
		return fmt.Errorf("failed to create service account token: %w", err)
	}

	// Set the token and its expiration.
	p.token.mutex.Lock()
	p.token.token = string(tokenResp.Token)
	expiresWhen := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	p.token.expires = &expiresWhen
	p.token.mutex.Unlock()

	return nil
}
