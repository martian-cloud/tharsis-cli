package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tableformatter"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// terraformProviderCreateCommand is the top-level structure for the terraform-provider create command.
type terraformProviderCreateCommand struct {
	meta *Metadata
}

// NewTerraformProviderCreateCommandFactory returns a (Terraform provider Create) Command struct.
func NewTerraformProviderCreateCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return terraformProviderCreateCommand{
			meta: meta,
		}, nil
	}
}

func (tpcc terraformProviderCreateCommand) Run(args []string) int {
	tpcc.meta.Logger.Debugf("Starting the 'terraform-provider create' command with %d arguments:", len(args))
	for ix, arg := range args {
		tpcc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := tpcc.meta.GetSDKClient()
	if err != nil {
		tpcc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return tpcc.doTerraformProviderCreate(ctx, client, args)
}

func (tpcc terraformProviderCreateCommand) doTerraformProviderCreate(ctx context.Context, client *tharsis.Client, opts []string) int {
	tpcc.meta.Logger.Debugf("will do terraform-provider create, %d opts", len(opts))

	defs := tpcc.buildTerraformProviderCreateDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(tpcc.meta.BinaryName+" terraform-provider create", defs, opts)
	if err != nil {
		tpcc.meta.Logger.Error(output.FormatError("failed to parse terraform-provider create options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		tpcc.meta.Logger.Error(output.FormatError("missing terraform-provider create path", nil),
			tpcc.HelpTerraformProviderCreate())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive terraform-provider create arguments: %s", cmdArgs)
		tpcc.meta.Logger.Error(output.FormatError(msg, nil), tpcc.HelpTerraformProviderCreate())
		return 1
	}

	tfProviderPath := cmdArgs[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		tpcc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	repositoryURL := getOption("repository-url", "", cmdOpts)[0]
	private, err := getBoolOptionValue("private", "true", cmdOpts)
	if err != nil {
		tpcc.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	// Error is already logged.
	if !isResourcePathValid(tpcc.meta, tfProviderPath) {
		return 1
	}

	// Create the Terraform provider
	index := strings.LastIndex(tfProviderPath, "/")
	tfProvider, err := client.TerraformProvider.CreateProvider(ctx, &types.CreateTerraformProviderInput{
		Name:          tfProviderPath[index+1:],
		GroupPath:     tfProviderPath[:index],
		Private:       private,
		RepositoryURL: repositoryURL,
	})
	if err != nil {
		tpcc.meta.UI.Error(output.FormatError("failed to create Terraform provider", err))
		return 1
	}

	return tpcc.outputTerraformProvider(toJSON, tfProvider)
}

func (tpcc terraformProviderCreateCommand) outputTerraformProvider(toJSON bool,
	tfProvider *types.TerraformProvider,
) int {
	if toJSON {
		buf, err := objectToJSON(tfProvider)
		if err != nil {
			tpcc.meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		tpcc.meta.UI.Output(string(buf))
	} else {
		tableInput := [][]string{
			{
				"id",
				"name",
				"resource path",
				"registry namespace",
				"private",
				"repository url",
			},
			{
				tfProvider.Metadata.ID,
				tfProvider.Name,
				tfProvider.ResourcePath,
				tfProvider.RegistryNamespace,
				fmt.Sprintf("%t", tfProvider.Private),
				tfProvider.RepositoryURL,
			},
		}
		tpcc.meta.UI.Output(tableformatter.FormatTable(tableInput))
	}

	return 0
}

// buildTerraformProviderCreateDefs returns defs used by terraform-provider create command.
func (tpcc terraformProviderCreateCommand) buildTerraformProviderCreateDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"private": {
			Arguments: []string{"Private"},
			Synopsis:  "Set private to false to allow all groups to view and use the Terraform provider (default=true).",
		},
		"repository-url": {
			Arguments: []string{"Repository_URL"},
			Synopsis:  "The repository URL for this Terraform provider.",
		},
	}

	return buildJSONOptionDefs(defs)
}

func (tpcc terraformProviderCreateCommand) Synopsis() string {
	return "Create a new Terraform provider."
}

func (tpcc terraformProviderCreateCommand) Help() string {
	return tpcc.HelpTerraformProviderCreate()
}

// HelpTerraformProviderCreate produces the help string for the 'terraform-provider create' command.
func (tpcc terraformProviderCreateCommand) HelpTerraformProviderCreate() string {
	return fmt.Sprintf(`
Usage: %s [global options] terraform-provider create [options] <full_path>

   The terraform-provider create command creates a new Terraform provider.

%s

`, tpcc.meta.BinaryName, buildHelpText(tpcc.buildTerraformProviderCreateDefs()))
}
