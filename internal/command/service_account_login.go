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

// serviceAccountLoginCommand is the top-level structure for the service-account login command.
type serviceAccountLoginCommand struct {
	meta *Metadata
}

// NewServiceAccountLoginCommandFactory returns a serviceAccountLoginCommand struct.
func NewServiceAccountLoginCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return serviceAccountLoginCommand{
			meta: meta,
		}, nil
	}
}

func (sal serviceAccountLoginCommand) Run(args []string) int {
	sal.meta.Logger.Debugf("Starting the 'service-account login' command with %d arguments:", len(args))
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

	return sal.doServiceAccountLogin(ctx, client, args)
}

func (sal serviceAccountLoginCommand) doServiceAccountLogin(ctx context.Context,
	client *tharsis.Client, opts []string) int {
	sal.meta.Logger.Debugf("will do service-account login, %d opts", len(opts))

	defs := sal.buildOptions()

	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(sal.meta.BinaryName+" service-account login",
		defs, opts)
	if err != nil {
		sal.meta.Logger.Error(output.FormatError("failed to parse service-account login options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		sal.meta.Logger.Error(output.FormatError("missing service account path", nil),
			sal.HelpServiceAccountLogin())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive service-account login arguments: %s", cmdArgs)
		sal.meta.Logger.Error(output.FormatError(msg, nil), sal.HelpServiceAccountLogin())
		return 1
	}

	serviceAccountPath := cmdArgs[0]
	token := getOption("token", "", cmdOpts)[0] // required; see help message
	asJSON := getOption("json", "", cmdOpts)[0] == "1"

	// Make sure the service account path has a slash and get the parent group path.
	if !isResourcePathValid(sal.meta, serviceAccountPath) {
		return 1
	}

	input := &sdktypes.ServiceAccountLoginInput{
		ServiceAccountPath: serviceAccountPath,
		Token:              token,
	}

	resp, err := client.ServiceAccount.Login(ctx, input)
	if err != nil {
		sal.meta.Logger.Error(output.FormatError("failed to log in to service account", err))
		return 1
	}

	if asJSON {

		// Marshal and indent the JSON output.
		marshalled, err := objectToJSON(resp)
		if err != nil {
			sal.meta.Logger.Error(output.FormatError("failed to marshal login response", err))
			return 1
		}

		sal.meta.UI.Output(marshalled)
	} else {

		// Just print the token (and not the expiration).
		sal.meta.UI.Output(resp.Token)
	}

	return 0
}

func (sal serviceAccountLoginCommand) Synopsis() string {
	return "Log in as a service account."
}

func (sal serviceAccountLoginCommand) Help() string {
	return sal.HelpServiceAccountLogin()
}

// buildOptions builds the option definitions for the service account login command.
func (sal serviceAccountLoginCommand) buildOptions() optparser.OptionDefinitions {
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

// HelpServiceAccountLogin produces the help string for the 'service-account login' command.
func (sal serviceAccountLoginCommand) HelpServiceAccountLogin() string {
	return fmt.Sprintf(`
Usage: %s [global options] service-account login [options] <service_account_path>

   The service-account login command logs in as a service account.
	 It uses the supplied token to authenticate.
	 It prints a post-login token.

	 The input token is used for the OIDC federated login that is issued by an
	 identity provider specified in the service account's trust policy.

	 The returned token is the service account token that can be used to
	 authenticate with the API.

%s

`, sal.meta.BinaryName, buildHelpText(sal.buildOptions()))
}

// The End.
