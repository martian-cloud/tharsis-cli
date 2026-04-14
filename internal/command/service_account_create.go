package command

import (
	"errors"
	"fmt"

	"github.com/aws/smithy-go/ptr"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"google.golang.org/protobuf/encoding/protojson"
)

type serviceAccountCreateCommand struct {
	*BaseCommand

	groupID                 *string
	description             *string
	enableClientCredentials *bool
	oidcTrustPolicies       []string
	toJSON                  *bool
}

var _ Command = (*serviceAccountCreateCommand)(nil)

func (c *serviceAccountCreateCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: name")
	}

	return nil
}

// NewServiceAccountCreateCommandFactory returns a serviceAccountCreateCommand struct.
func NewServiceAccountCreateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &serviceAccountCreateCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *serviceAccountCreateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("service-account create"),
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

	result, err := c.grpcClient.ServiceAccountsClient.CreateServiceAccount(c.Context, &pb.CreateServiceAccountRequest{
		Name:                    c.arguments[0],
		Description:             ptr.ToString(c.description),
		GroupId:                 *c.groupID,
		EnableClientCredentials: *c.enableClientCredentials,
		OidcTrustPolicies:       policies,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create service account")
		return 1
	}

	return c.Output(result, c.toJSON)
}

func (*serviceAccountCreateCommand) Synopsis() string {
	return "Create a new service account."
}

func (*serviceAccountCreateCommand) Usage() string {
	return "tharsis [global options] service-account create [options] <name>"
}

func (*serviceAccountCreateCommand) Description() string {
	return `
   Creates a service account for machine-to-
   machine auth using OIDC trust policies.
   Created within a group for CI/CD pipelines
   and automation workflows.
`
}

func (*serviceAccountCreateCommand) Example() string {
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
tharsis service-account create \
  -group-id "trn:group:<group_path>" \
  -description "<description>" \
  -oidc-trust-policy '{"issuer":"https://gitlab.com","bound_claims_type":"STRING","bound_claims":{"namespace_path":"<namespace_path>"}}' \
  <name>
%[1]s
`, "```")
}

func (c *serviceAccountCreateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.groupID,
		"group-id",
		"Group ID or TRN where the service account will be created.",
		flag.Required(),
	)
	f.StringVar(
		&c.description,
		"description",
		"Description for the service account.",
	)
	f.BoolVar(
		&c.enableClientCredentials,
		"enable-client-credentials",
		"Enable client credentials authentication.",
		flag.Default(false),
	)
	f.StringSliceVar(
		&c.oidcTrustPolicies,
		"oidc-trust-policy",
		"OIDC trust policy as JSON.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
