package command

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/smithy-go/ptr"
	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tableformatter"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

type terraformProviderMirrorListVersionsCommand struct {
	meta *Metadata
}

// NewTerraformProviderMirrorListVersionsCommandFactory returns a terraformProviderMirrorListVersionsCommand struct.
func NewTerraformProviderMirrorListVersionsCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return terraformProviderMirrorListVersionsCommand{
			meta: meta,
		}, nil
	}
}

func (c terraformProviderMirrorListVersionsCommand) Run(args []string) int {
	c.meta.Logger.Debugf("Starting the 'terraform-provider-mirror list-versions' command with %d arguments:", len(args))
	for ix, arg := range args {
		c.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := c.meta.GetSDKClient()
	if err != nil {
		c.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return c.doTerraformProviderMirrorListVersions(ctx, client, args)
}

func (c terraformProviderMirrorListVersionsCommand) doTerraformProviderMirrorListVersions(ctx context.Context, client *tharsis.Client, opts []string) int {
	c.meta.Logger.Debugf("will do terraform-provider-mirror list-versions, %d opts: %#v", len(opts), opts)

	defs := c.defs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(c.meta.BinaryName+" terraform-provider-mirror list-versions", defs, opts)
	if err != nil {
		c.meta.Logger.Error(output.FormatError("failed to parse terraform-provider-mirror list-versions options", err))
		return cli.RunResultHelp
	}
	if len(cmdArgs) < 1 {
		c.meta.Logger.Error(output.FormatError("missing terraform-provider-mirror list-versions group path", nil))
		return cli.RunResultHelp
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive terraform-provider-mirror list-versions arguments: %s", cmdArgs)
		c.meta.Logger.Error(output.FormatError(msg, nil))
		return cli.RunResultHelp
	}

	// Extract option values.
	groupPath := cmdArgs[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		c.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	cursor := getOption("cursor", "", cmdOpts)[0]
	limit, err := strconv.ParseInt(getOption("limit", "100", cmdOpts)[0], 10, 64) // 100 is the maximum allowed by GraphQL
	if err != nil {
		msg := fmt.Sprintf("invalid limit option value: %s", cmdOpts["limit"])
		c.meta.Logger.Error(output.FormatError(msg, nil))
		return 1
	}
	limit32 := int32(limit)
	sortOrderOption := strings.ToLower(getOption("sort-order", "", cmdOpts)[0])

	if !isNamespacePathValid(c.meta, groupPath) {
		return 1
	}

	// Leniently default to ascending order unless instructed otherwise.
	sortOrder := "asc"
	if strings.HasSuffix(sortOrderOption, "desc") {
		sortOrder = sortOrderOption
	}

	var sortable sdktypes.TerraformProviderVersionMirrorSortableField

	if sortOrder == "asc" {
		sortable = sdktypes.TerraformProviderVersionMirrorSortableFieldCreatedAtAsc
	} else {
		sortable = sdktypes.TerraformProviderVersionMirrorSortableFieldCreatedAtDesc
	}

	// Prepare the inputs.
	input := &sdktypes.GetTerraformProviderVersionMirrorsInput{
		Sort: &sortable,
		PaginationOptions: &sdktypes.PaginationOptions{
			Limit: &limit32,
		},
		// We will include all inherited version mirrors, since it's likely
		// the user will be querying from outside the root group but within
		// its hierarchy.
		IncludeInherited: ptr.Bool(true),
		GroupPath:        groupPath,
	}

	if cursor != "" {
		input.PaginationOptions.Cursor = &cursor
	}

	c.meta.Logger.Debugf("terraform-provider-mirror list-versions input: %#v", input)

	versionsOutput, err := client.TerraformProviderVersionMirror.GetProviderVersionMirrors(ctx, input)
	if err != nil {
		c.meta.Logger.Error(output.FormatError("failed to get a list of provider version mirrors", err))
		return 1
	}

	if toJSON {
		buf, err := objectToJSON(versionsOutput)
		if err != nil {
			c.meta.Logger.Error(output.FormatError("failed to get JSON output", err))
			return 1
		}
		c.meta.UI.Output(string(buf))
	} else {
		// Format the output.
		tableInput := make([][]string, len(versionsOutput.VersionMirrors)+1)
		tableInput[0] = []string{"id", "semantic version", "registry hostname", "registry namespace", "type"}
		for ix, vm := range versionsOutput.VersionMirrors {
			tableInput[ix+1] = []string{vm.Metadata.ID, vm.SemanticVersion, vm.RegistryHostname,
				vm.RegistryNamespace, vm.Type}
		}
		c.meta.UI.Output(tableformatter.FormatTable(tableInput))

		c.meta.UI.Output(fmt.Sprintf("has next page: %v", versionsOutput.PageInfo.HasNextPage))
		if versionsOutput.PageInfo.HasNextPage {
			// Show the next cursor _ONLY_ if there is a next page.
			c.meta.UI.Output(fmt.Sprintf("next cursor: %s", versionsOutput.PageInfo.Cursor))
		}
	}

	return 0
}

func (terraformProviderMirrorListVersionsCommand) defs() optparser.OptionDefinitions {
	return buildJSONOptionDefs(buildPaginationOptionDefs())
}

func (terraformProviderMirrorListVersionsCommand) Synopsis() string {
	return "List Terraform Provider versions available via Tharsis provider mirror."
}

func (c terraformProviderMirrorListVersionsCommand) Help() string {
	return fmt.Sprintf(`
Usage: %s [global options] terraform-provider-mirror list-versions [options] <root-group-path>

   The terraform-provider-mirror list-versions command prints
   information about (likely multiple) Terraform provider
   version mirrors. Supports pagination, filtering and
   sorting the output.

   Example:

   %s terraform-provider-mirror list-versions \
      --limit 5 \
      --json \
      top-level

   Above command will only show five version mirrors
   in JSON format from root group top-level.

%s

`,
		c.meta.BinaryName,
		c.meta.BinaryName,
		buildHelpText(c.defs()),
	)
}
