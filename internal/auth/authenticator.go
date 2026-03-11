package auth

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	"golang.org/x/oauth2"
)

//go:generate go tool mockery --name Authenticator --inpackage --case underscore

// Authenticator handles authentication flow.
type Authenticator interface {
	PerformLogin(ctx context.Context) (*oauth2.Token, error)
	StoreToken(token *oauth2.Token) error
}

// Options contains optional fields for creating an authenticator.
type Options struct {
	logger     hclog.Logger
	ui         terminal.UI
	grpcClient *client.Client
}

// Option is a functional option for authenticator options.
type Option func(*Options)

// WithLogger sets the logger.
func WithLogger(l hclog.Logger) Option {
	return func(c *Options) {
		c.logger = l
	}
}

// WithUI sets the UI for output.
func WithUI(u terminal.UI) Option {
	return func(c *Options) {
		c.ui = u
	}
}

// WithGRPCClient sets the gRPC client.
func WithGRPCClient(grpcClient *client.Client) Option {
	return func(c *Options) {
		c.grpcClient = grpcClient
	}
}
