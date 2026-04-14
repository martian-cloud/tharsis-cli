package command

import (
	"errors"
	"fmt"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"google.golang.org/protobuf/encoding/protojson"
)

type serviceAccountUpdateCommand struct {
	*BaseCommand

	description             *string
	enableClientCredentials *bool
	version                 *int64
	oidcTrustPolicies       []string
	toJSON                  *bool
}

var _ Command = (*serviceAccountUpdateCommand)(nil)

func (c *serviceAccountUpdateCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewServiceAccountUpdateCommandFactory returns a serviceAccountUpdateCommand struct.
func NewServiceAccountUpdateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &serviceAccountUpdateCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *serviceAccountUpdateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("service-account update"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	var policies []*pb.OIDCTrustPolicy
	for _, raw := range c.oidcTrustPolicies {
		policy := &pb.OIDCTrustPolicy{}
		if err := protojson.Unmarshal([]byte(raw), policy); err != nil {
			c.UI.Errorf(`failed to parse OIDC trust policy JSON, expected format: {"issuer":"...","bound_claims_type":"STRING|GLOB","bound_claims":{"key":"value"}}`)
			return 1
		}

		if policy.Issuer == "" {
			c.UI.Errorf("OIDC trust policy issuer is required")
			return 1
		}

		if len(policy.BoundClaims) == 0 {
			c.UI.Errorf("OIDC trust policy bound_claims is required")
			return 1
		}

		policies = append(policies, policy)
	}

	result, err := c.grpcClient.ServiceAccountsClient.UpdateServiceAccount(c.Context, &pb.UpdateServiceAccountRequest{
		Id:                      c.arguments[0],
		Description:             c.description,
		EnableClientCredentials: c.enableClientCredentials,
		Version:                 c.version,
		OidcTrustPolicies:       policies,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update service account")
		return 1
	}

	return c.Output(result, c.toJSON)
}

func (*serviceAccountUpdateCommand) Synopsis() string {
	return "Update a service account."
}

func (*serviceAccountUpdateCommand) Usage() string {
	return "tharsis [global options] service-account update [options] <id>"
}

func (*serviceAccountUpdateCommand) Description() string {
	return `
   Modifies an existing service account's configuration.
   OIDC trust policies are fully replaced when specified.
`
}

func (*serviceAccountUpdateCommand) Example() string {
	return fmt.Sprintf(`
OIDC trust policy JSON format:

%[1]sjson
{
  "issuer": "https://gitlab.com",
  "bound_claims_type": "STRING",
  "bound_claims": {
    "namespace_path": "<namespace_path>"
  }
}
%[1]s

%[1]sbash
tharsis service-account update \
  -description "<description>" \
  -oidc-trust-policy '{"issuer":"https://gitlab.com","bound_claims_type":"STRING","bound_claims":{"namespace_path":"<namespace_path>"}}' \
  <service_account_id>
%[1]s
`, "```")
}

func (c *serviceAccountUpdateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.description,
		"description",
		"Description for the service account.",
	)
	f.BoolVar(
		&c.enableClientCredentials,
		"enable-client-credentials",
		"Enable client credentials authentication.",
	)
	f.StringSliceVar(
		&c.oidcTrustPolicies,
		"oidc-trust-policy",
		"OIDC trust policy as JSON.",
	)
	f.Int64Var(
		&c.version,
		"version",
		"Optimistic locking version. Usually not required.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
