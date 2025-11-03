package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tableformatter"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

type terraformProviderMirrorGetVersionCommand struct {
	meta *Metadata
}

// NewTerraformProviderMirrorGetVersionCommandFactory returns a terraformProviderMirrorGetVersionCommand struct.
func NewTerraformProviderMirrorGetVersionCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return terraformProviderMirrorGetVersionCommand{
			meta: meta,
		}, nil
	}
}

func (c terraformProviderMirrorGetVersionCommand) Run(args []string) int {
	c.meta.Logger.Debugf("Starting the 'terraform-provider-mirror get-version' command with %d arguments:", len(args))
	for ix, arg := range args {
		c.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := c.meta.GetSDKClient()
	if err != nil {
		c.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return c.doTerraformProviderMirrorGetVersion(ctx, client, args)
}

func (c terraformProviderMirrorGetVersionCommand) doTerraformProviderMirrorGetVersion(ctx context.Context, client *tharsis.Client, opts []string) int {
	c.meta.Logger.Debugf("will do terraform-provider get-version, %d opts", len(opts))

	defs := buildJSONOptionDefs(optparser.OptionDefinitions{})
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(c.meta.BinaryName+" terraform-provider-mirror get-version", defs, opts)
	if err != nil {
		c.meta.Logger.Error(output.FormatError("failed to parse terraform-provider-mirror get-version options", err))
		return cli.RunResultHelp
	}
	if len(cmdArgs) < 1 {
		c.meta.Logger.Error(output.FormatError("missing terraform-provider-mirror get-version id", nil))
		return cli.RunResultHelp
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive terraform-provider-mirror get-version arguments: %s", cmdArgs)
		c.meta.Logger.Error(output.FormatError(msg, nil))
		return cli.RunResultHelp
	}

	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		c.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	toGet := &sdktypes.GetTerraformProviderVersionMirrorInput{
		ID: cmdArgs[0],
	}

	c.meta.Logger.Debugf("terraform-provider-mirror get-version input: %#v", toGet)

	versionMirror, err := client.TerraformProviderVersionMirror.GetProviderVersionMirror(ctx, toGet)
	if err != nil {
		c.meta.UI.Error(output.FormatError("failed to get terraform provider version mirror", err))
		return 1
	}

	return outputVersionMirror(c.meta, toJSON, versionMirror)
}

func outputVersionMirror(meta *Metadata, toJSON bool, versionMirror *sdktypes.TerraformProviderVersionMirror) int {
	if toJSON {
		buf, err := objectToJSON(versionMirror)
		if err != nil {
			meta.UI.Output(output.FormatError("failed to get JSON output", err))
			return 1
		}
		meta.UI.Output(string(buf))
	} else {
		tableInput := [][]string{
			{"id", "semantic version", "registry hostname", "registry namespace", "type"},
			{versionMirror.Metadata.ID, versionMirror.SemanticVersion, versionMirror.RegistryHostname,
				versionMirror.RegistryNamespace, versionMirror.Type},
		}

		meta.UI.Output(tableformatter.FormatTable(tableInput))
	}

	return 0
}

func (terraformProviderMirrorGetVersionCommand) Synopsis() string {
	return "Get a mirrored Terraform provider version."
}

func (c terraformProviderMirrorGetVersionCommand) Help() string {
	return fmt.Sprintf(`
Usage: %s [global options] terraform-provider-mirror get-version [options] <id>

   The terraform-provider-mirror get-version command retrieves
   a Terraform Provider version from the provider mirror.

%s

`, c.meta.BinaryName, buildHelpText(buildJSONOptionDefs(optparser.OptionDefinitions{})))
}
