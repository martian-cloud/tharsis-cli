package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// serviceAccountCreateTokenCommand is the top-level structure for the service-account create-token command.
type serviceAccountCreateTokenCommand struct {
	meta *Metadata
}

// NewServiceAccountCreateTokenCommandFactory returns a serviceAccountCreateTokenCommand struct.
func NewServiceAccountCreateTokenCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return serviceAccountCreateTokenCommand{
			meta: meta,
		}, nil
	}
}

func (sal serviceAccountCreateTokenCommand) Run(args []string) int {
	sal.meta.Logger.Debugf("Starting the 'service-account create-token' command with %d arguments:", len(args))
	for ix, arg := range args {
		sal.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := sal.meta.ReadSettings()
	if err != nil {
		sal.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		sal.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return sal.doServiceAccountCreateToken(ctx, client, args)
}

func (sal serviceAccountCreateTokenCommand) doServiceAccountCreateToken(ctx context.Context,
	client *tharsis.Client, opts []string) int {
	sal.meta.Logger.Debugf("will do service-account create-token, %d opts", len(opts))

	defs := sal.buildOptions()

	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(sal.meta.BinaryName+" service-account create-token",
		defs, opts)
	if err != nil {
		sal.meta.Logger.Error(output.FormatError("failed to parse service-account create-token options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		sal.meta.Logger.Error(output.FormatError("missing service account path", nil),
			sal.HelpServiceAccountCreateToken())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive service-account create-token arguments: %s", cmdArgs)
		sal.meta.Logger.Error(output.FormatError(msg, nil), sal.HelpServiceAccountCreateToken())
		return 1
	}

	serviceAccountPath := cmdArgs[0]
	token := getOption("token", "", cmdOpts)[0] // required; see help message
	asJSON := getOption("json", "", cmdOpts)[0] == "1"

	// Make sure the service account path has a slash and get the parent group path.
	if !isResourcePathValid(sal.meta, serviceAccountPath) {
		return 1
	}

	input := &sdktypes.ServiceAccountCreateTokenInput{
		ServiceAccountPath: serviceAccountPath,
		Token:              token,
	}

	resp, err := client.ServiceAccount.CreateToken(ctx, input)
	if err != nil {
		sal.meta.Logger.Error(output.FormatError("failed to create token for service account", err))
		return 1
	}

	if asJSON {

		// Marshal and indent the JSON output.
		marshalled, err := objectToJSON(resp)
		if err != nil {
			sal.meta.Logger.Error(output.FormatError("failed to marshal create-token response", err))
			return 1
		}

		sal.meta.UI.Output(marshalled)
	} else {

		// Just print the token (and not the expiration).
		sal.meta.UI.Output(resp.Token)
	}

	return 0
}

func (sal serviceAccountCreateTokenCommand) Synopsis() string {
	return "Create a token for a service account."
}

func (sal serviceAccountCreateTokenCommand) Help() string {
	return sal.HelpServiceAccountCreateToken()
}

// buildOptions builds the option definitions for the service account create-token command.
func (sal serviceAccountCreateTokenCommand) buildOptions() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"token": {
			Arguments: []string{"Token"},
			Synopsis:  "Initial authentication token.",
			Required:  true,
		},
		"json": {
			Arguments: []string{},
			Synopsis:  "Show output as JSON.",
		},
	}
	return buildJSONOptionDefs(defs)
}

// HelpServiceAccountCreateToken produces the help string for the 'service-account create-token' command.
func (sal serviceAccountCreateTokenCommand) HelpServiceAccountCreateToken() string {
	return fmt.Sprintf(`
Usage: %s [global options] service-account create-token [options] <service_account_path>

   The service-account create-token command creates a token for a service account.
	 It uses the supplied token to authenticate.
	 It prints an output token.

	 The input token is used for the OIDC federated login that is issued by an
	 identity provider specified in the service account's trust policy.

	 The output token is the service account token that can be used to
	 authenticate with the API.

%s

`, sal.meta.BinaryName, buildHelpText(sal.buildOptions()))
}

// The End.
