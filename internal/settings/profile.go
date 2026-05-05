package settings

import (
	"context"
	"fmt"

	"github.com/aws/smithy-go/ptr"
	"github.com/hashicorp/go-hclog"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client/token"
)

const (
	// DefaultProfileName is the name of the default profile
	DefaultProfileName = "default"
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

// NewTokenResolver creates a new token resolver for the profile.
// When the token originates from the credentials file, the returned resolver
// re-reads the file on each call so long-lived processes (e.g. MCP server)
// pick up tokens refreshed by a concurrent `sso login`.
func (p *Profile) NewTokenResolver(ctx context.Context) (client.TokenResolver, error) {
	tc := &token.Config{
		StaticToken: ptr.ToString(p.token),
	}

	var staticTokenFunc func() (string, error)
	if p.token != nil {
		endpoint := p.Endpoint
		staticTokenFunc = func() (string, error) {
			s, err := ReadSettings()
			if err != nil {
				return "", fmt.Errorf("failed to re-read settings: %w", err)
			}

			profile, err := s.FindProfileByEndpoint(endpoint)
			if err != nil {
				return "", err
			}

			return ptr.ToString(profile.token), nil
		}
	}

	return tc.Resolve(ctx, p.Endpoint, staticTokenFunc, token.WithTLSSkipVerify(p.TLSSkipVerify))
}

// NewClient returns a Tharsis client based on the specified profile.
func (p *Profile) NewClient(ctx context.Context, withAuth bool, userAgent string, logger hclog.Logger) (*client.GRPCClient, error) {
	clientConfig := &client.GRPCClientConfig{
		HTTPEndpoint:  p.Endpoint,
		TLSSkipVerify: p.TLSSkipVerify,
		UserAgent:     userAgent,
		Logger:        logger,
	}

	if withAuth {
		tokenResolver, err := p.NewTokenResolver(ctx)
		if err != nil {
			return nil, err
		}

		clientConfig.TokenResolver = tokenResolver
	}

	return client.NewGRPCClient(ctx, clientConfig)
}
