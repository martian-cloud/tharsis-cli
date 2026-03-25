package settings

import (
	"context"
	"fmt"

	"github.com/aws/smithy-go/ptr"
	"github.com/hashicorp/go-hclog"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
)

// Profile holds the contents of one profile from a settings file.
type Profile struct {
	// token and Name are not serialized to the settings JSON; tokens live
	// in a separate credentials file for security, and Name is derived
	// from the map key during deserialization.
	token         *string `json:"-"`
	Name          string  `json:"-"`
	Endpoint      string  `json:"endpoint"`
	TLSSkipVerify bool    `json:"tlsSkipVerify"`
}

// SetToken sets the token for a profile.
func (p *Profile) SetToken(token string) {
	p.token = &token
}

// NewTokenGetter creates a new token getter for the profile.
// When the token originates from the credentials file, the returned getter
// re-reads the file on each call so long-lived processes (e.g. MCP server)
// pick up tokens refreshed by a concurrent `sso login`.
func (p *Profile) NewTokenGetter(ctx context.Context) (client.TokenGetter, error) {
	resolver := &tokenResolver{
		StaticToken: ptr.ToString(p.token),
	}

	return resolver.resolve(ctx, p.Endpoint, p.TLSSkipVerify, p.tokenFunc())
}

// NewClient returns a Tharsis client based on the specified profile.
func (p *Profile) NewClient(ctx context.Context, withAuth bool, userAgent string, logger hclog.Logger) (*client.Client, error) {
	clientConfig := &client.Config{
		HTTPEndpoint:  p.Endpoint,
		TLSSkipVerify: p.TLSSkipVerify,
		UserAgent:     userAgent,
		Logger:        logger,
	}

	if withAuth {
		tokenGetter, err := p.NewTokenGetter(ctx)
		if err != nil {
			return nil, err
		}

		clientConfig.TokenGetter = tokenGetter
	}

	return client.New(ctx, clientConfig)
}

// tokenFunc returns a function that re-reads the token from the credentials
// file on each call. This ensures long-lived processes see tokens refreshed
// by concurrent `sso login` invocations rather than using a stale value.
func (p *Profile) tokenFunc() func() (string, error) {
	if p.token == nil {
		return nil
	}

	return func() (string, error) {
		s, err := ReadSettings()
		if err != nil {
			return "", fmt.Errorf("failed to re-read settings: %w", err)
		}

		profile, err := s.FindProfileByEndpoint(p.Endpoint)
		if err != nil {
			return "", err
		}

		return ptr.ToString(profile.token), nil
	}
}
