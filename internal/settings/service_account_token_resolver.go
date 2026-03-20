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
	// expirationLeeway causes renewal slightly before the token actually
	// expires, avoiding race conditions where a token expires mid-request.
	expirationLeeway = 30 * time.Second
)

// oidcTokenGetter is a callback for returning the OIDC token that is used to login to the service account.
type oidcTokenGetter func() ([]byte, error)

// tokenInfo holds a service account token and its expiration behind a
// mutex so concurrent callers can safely share a single resolver.
type tokenInfo struct {
	expires *time.Time
	token   string
	mutex   sync.RWMutex
}

// serviceAccountTokenResolver exchanges an OIDC token for a short-lived
// service account token, caching and renewing it transparently.
type serviceAccountTokenResolver struct {
	tokenGetter      oidcTokenGetter
	token            *tokenInfo
	client           *client.Client
	serviceAccountID string
}

func newServiceAccountTokenResolver(
	ctx context.Context,
	serviceAccountID string,
	httpEndpoint string,
	tlsSkipVerify bool,
	tokenGetter oidcTokenGetter,
) (client.TokenGetter, error) {
	// Separate unauthenticated client for token renewal to avoid
	// circular dependency on the token we're trying to obtain.
	c, err := client.New(ctx, &client.Config{
		HTTPEndpoint:  httpEndpoint,
		TLSSkipVerify: tlsSkipVerify,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &serviceAccountTokenResolver{
		tokenGetter:      tokenGetter,
		token:            &tokenInfo{},
		client:           c,
		serviceAccountID: serviceAccountID,
	}, nil
}

func (r *serviceAccountTokenResolver) Token(ctx context.Context) (string, error) {
	if r.isTokenExpired() {
		if err := r.renewToken(ctx); err != nil {
			return "", fmt.Errorf("service account token renewal failed: %w", err)
		}
	}

	r.token.mutex.RLock()
	defer r.token.mutex.RUnlock()
	return r.token.token, nil
}

func (r *serviceAccountTokenResolver) isTokenExpired() bool {
	r.token.mutex.RLock()
	defer r.token.mutex.RUnlock()

	return r.token.expires == nil || !time.Now().Add(expirationLeeway).Before(*r.token.expires)
}

func (r *serviceAccountTokenResolver) renewToken(ctx context.Context) error {
	oidcToken, err := r.tokenGetter()
	if err != nil {
		return fmt.Errorf("failed to get OIDC token: %w", err)
	}

	tokenResp, err := r.client.ServiceAccountsClient.CreateOIDCToken(ctx, &pb.CreateOIDCTokenRequest{
		ServiceAccountId: r.serviceAccountID,
		Token:            string(oidcToken),
	})
	if err != nil {
		return fmt.Errorf("failed to create service account token: %w", err)
	}

	r.token.mutex.Lock()
	r.token.token = string(tokenResp.Token)
	expiresWhen := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	r.token.expires = &expiresWhen
	r.token.mutex.Unlock()

	return nil
}
