package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tableformatter"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// moduleGetVersionCommand is the top-level structure for the module get-version command.
type moduleGetVersionCommand struct {
	meta *Metadata
}

// NewModuleGetVersionCommandFactory returns a moduleGetVersionCommand struct.
func NewModuleGetVersionCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return moduleGetVersionCommand{
			meta: meta,
		}, nil
	}
}

func (mgc moduleGetVersionCommand) Run(args []string) int {
	mgc.meta.Logger.Debugf("Starting the 'module get-version' command with %d arguments:", len(args))
	for ix, arg := range args {
		mgc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := mgc.meta.GetSDKClient()
	if err != nil {
		mgc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return mgc.doModuleGetVersion(ctx, client, args)
}

func (mgc moduleGetVersionCommand) doModuleGetVersion(ctx context.Context, client *tharsis.Client, opts []string) int {
	mgc.meta.Logger.Debugf("will do module get-version, %d opts", len(opts))

	defs := mgc.buildModuleGetVersionDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(mgc.meta.BinaryName+" module get-version", defs, opts)
	if err != nil {
		mgc.meta.Logger.Error(output.FormatError("failed to parse module get-version argument", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		mgc.meta.Logger.Error(output.FormatError("missing module get-version path", nil), mgc.HelpModuleGetVersion())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive module get-version arguments: %s", cmdArgs)
		mgc.meta.Logger.Error(output.FormatError(msg, nil), mgc.HelpModuleGetVersion())
		return 1
	}

	modulePath := cmdArgs[0]
	versionTag := getOption("version", "", cmdOpts)[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		mgc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	actualPath := trn.ToPath(modulePath)
	if !isResourcePathValid(mgc.meta, actualPath) {
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.GetTerraformModuleVersionInput{ModulePath: &actualPath}  // Use extracted path

	if versionTag != "" {
		input.Version = &versionTag
	}
	mgc.meta.Logger.Debugf("module get-version input: %#v", input)

	// Get the module version.
	foundModule, err := client.TerraformModuleVersion.GetModuleVersion(ctx, input)
	if err != nil {
		mgc.meta.Logger.Error(output.FormatError("failed to get module version", err))
		return 1
	}

	return mgc.outputModuleVersion(mgc.meta, toJSON, foundModule)
}

// outputModuleVersion is the final output for most module version operations.
func (mgc moduleGetVersionCommand) outputModuleVersion(meta *Metadata, toJSON bool, moduleVersion *sdktypes.TerraformModuleVersion) int {
	if toJSON {
		buf, err := objectToJSON(moduleVersion)
		if err != nil {
			meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		meta.UI.Output(string(buf))
	} else {
		tableInput := [][]string{
			{
				"id",
				"module id",
				"version",
				"shasum",
				"status",
				"latest",
			},
			{
				moduleVersion.Metadata.ID,
				moduleVersion.ModuleID,
				moduleVersion.Version,
				moduleVersion.SHASum,
				moduleVersion.Status,
				fmt.Sprintf("%t", moduleVersion.Latest),
			},
		}
		meta.UI.Output(tableformatter.FormatTable(tableInput))
	}

	return 0
}

// buildModuleGetDefs returns defs used by module get command.
func (mgc moduleGetVersionCommand) buildModuleGetVersionDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"version": {
			Arguments: []string{"Version"},
			Synopsis:  "A semver compliant version tag to use as a filter.",
		},
	}

	return buildJSONOptionDefs(defs)
}

func (mgc moduleGetVersionCommand) Synopsis() string {
	return "Get a single module version."
}

func (mgc moduleGetVersionCommand) Help() string {
	return mgc.HelpModuleGetVersion()
}

// HelpModuleGetVersion prints the help string for the 'module get-version' command.
func (mgc moduleGetVersionCommand) HelpModuleGetVersion() string {
	return fmt.Sprintf(`
Usage: %s [global options] module get-version [options] <module-path>

   The module get-version command prints information
   about a module's version. Returns latest by default
   unless --version option is specified.

%s

`, mgc.meta.BinaryName, buildHelpText(mgc.buildModuleGetVersionDefs()))
}
