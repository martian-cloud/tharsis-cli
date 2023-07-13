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

type terraformProviderMirrorListPlatformsCommand struct {
	meta *Metadata
}

// NewTerraformProviderMirrorListPlatformsCommandFactory returns a terraformProviderMirrorListPlatformsCommand struct.
func NewTerraformProviderMirrorListPlatformsCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return terraformProviderMirrorListPlatformsCommand{
			meta: meta,
		}, nil
	}
}

func (c terraformProviderMirrorListPlatformsCommand) Run(args []string) int {
	c.meta.Logger.Debugf("Starting the 'terraform-provider-mirror list-platforms' command with %d arguments:", len(args))
	for ix, arg := range args {
		c.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	settings, err := c.meta.ReadSettings()
	if err != nil {
		c.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		c.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return c.doTerraformProviderMirrorListPlatforms(ctx, client, args)
}

func (c terraformProviderMirrorListPlatformsCommand) doTerraformProviderMirrorListPlatforms(ctx context.Context, client *tharsis.Client, opts []string) int {
	c.meta.Logger.Debugf("will do terraform-provider-mirror list-platforms, %d opts: %#v", len(opts), opts)

	defs := buildJSONOptionDefs(optparser.OptionDefinitions{})
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(c.meta.BinaryName+" terraform-provider-mirror list-platforms", defs, opts)
	if err != nil {
		c.meta.Logger.Error(output.FormatError("failed to parse terraform-provider-mirror list-platforms options", err))
		return cli.RunResultHelp
	}
	if len(cmdArgs) < 1 {
		c.meta.Logger.Error(output.FormatError("missing terraform-provider-mirror list-platforms version id", nil))
		return cli.RunResultHelp
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive terraform-provider-mirror list-platforms arguments: %s", cmdArgs)
		c.meta.Logger.Error(output.FormatError(msg, nil))
		return cli.RunResultHelp
	}

	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		c.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	input := &sdktypes.GetTerraformProviderPlatformMirrorsByVersionInput{
		VersionMirrorID: cmdArgs[0],
	}

	c.meta.Logger.Debugf("terraform-provider-mirror list-platforms input: %#v", input)

	platformMirrors, err := client.TerraformProviderPlatformMirror.GetProviderPlatformMirrorsByVersion(ctx, input)
	if err != nil {
		c.meta.Logger.Error(output.FormatError("failed to get a list of platform mirrors", err))
		return 1
	}

	if toJSON {
		buf, err := objectToJSON(platformMirrors)
		if err != nil {
			c.meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		c.meta.UI.Output(string(buf))
	} else {
		// Format the output.
		tableInput := make([][]string, len(platformMirrors)+1)
		tableInput[0] = []string{"id", "operating system", "architecture", "version mirror id"}
		for ix, pm := range platformMirrors {
			tableInput[ix+1] = []string{pm.Metadata.ID, pm.OS, pm.Arch, pm.VersionMirror.Metadata.ID}
		}
		c.meta.UI.Output(tableformatter.FormatTable(tableInput))
	}

	return 0
}

func (terraformProviderMirrorListPlatformsCommand) Synopsis() string {
	return "List all platforms associated with a mirrored provider version."
}

func (c terraformProviderMirrorListPlatformsCommand) Help() string {
	return fmt.Sprintf(`
Usage: %s [global options] terraform-provider-mirror list-platforms [options] <version_id>

   The terraform-provider-mirror list-platforms command lists
   all platforms that are mirrored for a Terraform provider
   version.

%s

`, c.meta.BinaryName, buildHelpText(buildJSONOptionDefs(optparser.OptionDefinitions{})))
}
