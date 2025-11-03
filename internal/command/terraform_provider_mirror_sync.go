package command

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	logger "github.com/caarlos0/log" // Allows disabling noise from disco package.
	"github.com/hashicorp/go-cleanhttp"
	tfaddr "github.com/hashicorp/terraform-registry-address"
	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/providermirror"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

type terraformProviderMirrorSyncCommand struct {
	meta *Metadata
}

// NewTerraformProviderMirrorSyncCommandFactory returns a terraformProviderMirrorSyncCommand struct.
func NewTerraformProviderMirrorSyncCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return terraformProviderMirrorSyncCommand{meta: meta}, nil
	}
}

func (c terraformProviderMirrorSyncCommand) Run(args []string) int {
	c.meta.Logger.Debugf("Starting the 'terraform-provider-mirror sync' command with %d arguments:", len(args))
	for ix, arg := range args {
		c.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := c.meta.GetSDKClient()
	if err != nil {
		c.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return c.doTerraformProviderMirrorSync(ctx, client, args)
}

func (c terraformProviderMirrorSyncCommand) doTerraformProviderMirrorSync(ctx context.Context, client *tharsis.Client, opts []string) int {
	c.meta.Logger.Debugf("will do terraform-provider-mirror sync, %d opts", len(opts))

	defs := c.defs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(c.meta.BinaryName+" terraform-provider-mirror sync", defs, opts)
	if err != nil {
		c.meta.Logger.Error(output.FormatError("failed to parse terraform-provider-mirror sync options", err))
		return cli.RunResultHelp
	}
	if len(cmdArgs) < 1 {
		c.meta.Logger.Error(output.FormatError("missing terraform-provider-mirror sync <provider-path>", nil))
		return cli.RunResultHelp
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive terraform-provider-mirror sync arguments: %s", cmdArgs)
		c.meta.Logger.Error(output.FormatError(msg, nil))
		return cli.RunResultHelp
	}

	fqn := cmdArgs[0]
	groupPath := getOption("group-path", "", cmdOpts)[0]
	version := getOption("version", "", cmdOpts)[0]
	platforms := getOptionSlice("platform", cmdOpts)
	allPlatforms := len(platforms) == 0 // If no platform specified, default to all supported platforms.
	useLatestVersion := version == ""   // If no version specified, use latest version.

	// Error is already logged.
	if !isNamespacePathValid(c.meta, groupPath) {
		return 1
	}

	for _, p := range platforms {
		if !strings.Contains(p, "_") {
			c.meta.UI.Output(output.FormatError("Invalid platform. Must be formatted as <os_arch>", nil))
			return 1
		}
	}

	// Validate and parse FQN.
	if strings.Count(fqn, "/") != 2 {
		c.meta.UI.Error(output.FormatError("invalid FQN. Must be hostname/namespace/name", nil))
		return 1
	}

	provider, err := tfaddr.ParseProviderSource(fqn)
	if err != nil {
		c.meta.UI.Error(output.FormatError("failed to parse provider FQN", err))
		return 1
	}

	logger.Info("starting terraform-provider-mirror sync")

	// Discover the providers service URL of the target terraform registry.
	log.Default().SetOutput(io.Discard) // Disable debug messages from disco.
	serviceURL, err := disco.New().DiscoverServiceURL(provider.Hostname, "providers.v1")
	if err != nil {
		c.meta.UI.Output(output.FormatError("failed to discover service URL", err))
		return 1
	}

	// Create an instance of the provider package resolver.
	resolver := providermirror.NewTerraformProviderPackageResolver(c.meta.Logger, serviceURL, cleanhttp.DefaultClient())

	availableVersions, err := resolver.ListAvailableProviderVersions(ctx, provider.Namespace, provider.Type)
	if err != nil {
		c.meta.UI.Output(output.FormatError("failed to list available provider versions", err))
		return 1
	}

	logger.Info("retrieved provider's supported platforms from the Terraform Registry API")

	if useLatestVersion {
		version, err = resolver.FindLatestVersion(availableVersions)
		if err != nil {
			c.meta.UI.Output(output.FormatError("failed to find latest provider version", err))
			return 1
		}
	}

	logger.WithField("version", version).Info("using terraform provider version")

	versionQueryInput := &sdktypes.GetTerraformProviderVersionMirrorByAddressInput{
		RegistryHostname:  provider.Hostname.String(),
		RegistryNamespace: provider.Namespace,
		Type:              provider.Type,
		Version:           version,
		GroupPath:         groupPath,
	}

	c.meta.Logger.Debugf("terraform-provider-mirror sync get version mirror input: %#v", versionQueryInput)

	// Attempt to find the version mirror first incase it already exists.
	versionMirror, err := client.TerraformProviderVersionMirror.GetProviderVersionMirrorByAddress(ctx, versionQueryInput)
	if err != nil && !tharsis.IsNotFoundError(err) {
		c.meta.UI.Error(output.FormatError("failed to get terraform provider version mirror", err))
		return 1
	}

	if versionMirror == nil {
		// Version mirror doesn't exist so, create it.
		versionInput := &sdktypes.CreateTerraformProviderVersionMirrorInput{
			GroupPath:         groupPath,
			Type:              provider.Type,
			RegistryNamespace: provider.Namespace,
			RegistryHostname:  provider.Hostname.String(),
			SemanticVersion:   version,
		}

		c.meta.Logger.Debugf("terraform-provider-mirror sync create version mirror input: %#v", versionInput)

		versionMirror, err = client.TerraformProviderVersionMirror.CreateProviderVersionMirror(ctx, versionInput)
		if err != nil {
			c.meta.UI.Error(output.FormatError("failed to create terraform provider version mirror", err))
			return 1
		}
	}

	logger.WithField("id", versionMirror.Metadata.ID).Info("using terraform provider version mirror id")

	// Get the platforms already available for the version so, we don't reupload ones that already exist.
	platformsInput := &sdktypes.GetTerraformProviderPlatformMirrorsByVersionInput{
		VersionMirrorID: versionMirror.Metadata.ID,
	}

	c.meta.Logger.Debugf("terraform-provider-mirror sync get platforms input: %#v", platformsInput)

	platformsList, err := client.TerraformProviderPlatformMirror.GetProviderPlatformMirrorsByVersion(ctx, platformsInput)
	if err != nil {
		c.meta.UI.Output(output.FormatError("failed to list platforms for provider version", err))
		return 1
	}

	// Map of existing platforms will make it easier to filter.
	existingPlatforms := map[string]struct{}{}
	for _, p := range platformsList {
		existingPlatforms[fmt.Sprintf("%s_%s", p.OS, p.Arch)] = struct{}{}
	}

	// Map allows platforms to be automatically de-duped, incase user accidentally specified one multiple times.
	missingPlatforms := map[string]struct{}{}
	if allPlatforms {
		// Since we're using all platforms, we must filter out ones that are already mirrored.
		missingPlatforms, err = resolver.FilterMissingPlatforms(versionMirror.SemanticVersion, availableVersions, existingPlatforms)
		if err != nil {
			c.meta.UI.Output(output.FormatError("failed to filter missing provider version platforms", err))
			return 1
		}
	} else {
		for _, p := range platforms {
			if _, ok := existingPlatforms[p]; !ok {
				missingPlatforms[p] = struct{}{}
			}
		}
	}

	if len(missingPlatforms) == 0 {
		c.meta.UI.Output("\nStopping since all platform packages are already mirrored")
		// Not an error so, we can gracefully finish.
		return 0
	}

	// Create a directory that'll contain all the packages we download.
	packagesDir, err := os.MkdirTemp("", "terraform-providers-*")
	if err != nil {
		c.meta.UI.Output(output.FormatError("failed to create temporary package directory", err))
		return 1
	}
	defer os.RemoveAll(packagesDir)

	logger.Info("locating and downloading packages...")
	logger.IncreasePadding()

	// To keep track of platform -> package file name we downloaded for uploading.
	platformPackageMap := map[string]string{}

	// Locate the packages and download them.
	for pf := range missingPlatforms {
		logger.Infof("locating and downloading provider package for %s", pf)

		foundResp, err := resolver.FindProviderPackage(ctx, pf, versionMirror)
		if err != nil {
			c.meta.UI.Output(output.FormatError("failed to find provider package at Terraform Registry API", err))
			return 1
		}

		platformPackageMap[pf], err = resolver.DownloadProviderPlatformPackage(ctx, foundResp.DownloadURL, packagesDir)
		if err != nil {
			c.meta.UI.Output(output.FormatError("failed to download provider package", err))
			return 1
		}
	}

	logger.ResetPadding()
	logger.Info("downloaded all provider platform packages")
	logger.Info("starting provider package upload to mirror...")
	logger.IncreasePadding()

	for pf, location := range platformPackageMap {
		logger.Infof("uploading provider package for %s", pf)

		if err := c.uploadProviderPlatformPackage(
			ctx,
			versionMirror.Metadata.ID,
			pf,
			location,
			client,
		); err != nil {
			c.meta.UI.Output(output.FormatError("failed to upload provider package to mirror", err))
			return 1
		}
	}

	logger.ResetPadding()
	c.meta.UI.Output("\nProvider platform packages uploaded to mirror successfully!")

	return 0
}

// uploadProviderPlatformPackage uploads the provider platform package to Tharsis Provider Network mirror.
func (c terraformProviderMirrorSyncCommand) uploadProviderPlatformPackage(
	ctx context.Context,
	versionMirrorID,
	platform,
	packageLocation string,
	client *tharsis.Client,
) error {
	pkgFile, err := os.Open(packageLocation)
	if err != nil {
		return err
	}
	defer pkgFile.Close()

	// Get the OS and architecture separately.
	parts := strings.Split(platform, "_")

	input := &sdktypes.UploadProviderPlatformPackageToMirrorInput{
		VersionMirrorID: versionMirrorID,
		Reader:          pkgFile,
		OS:              parts[0],
		Arch:            parts[1],
	}

	return client.TerraformProviderPlatformMirror.UploadProviderPlatformPackageToMirror(ctx, input)
}

func (terraformProviderMirrorSyncCommand) defs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"platform": {
			Arguments: []string{"Platform"},
			Synopsis:  "Specifies which platform (os_arch) the packages should be uploaded for. Defaults to all supported.",
		},
		"group-path": {
			Arguments: []string{"Group_Path"},
			Synopsis:  "Full path to the root group where this Terraform provider version will be mirrored.",
			Required:  true,
		},
		"version": {
			Arguments: []string{"Version"},
			Synopsis:  "The semantic version of the target Terraform provider. Defaults to latest.",
		},
	}
}

func (terraformProviderMirrorSyncCommand) Synopsis() string {
	return "Upload Terraform provider platform packages to the provider mirror."
}

func (c terraformProviderMirrorSyncCommand) Help() string {
	return fmt.Sprintf(`
Usage: %s [global options] terraform-provider-mirror sync [options] <FQN>

   The terraform-provider-mirror sync command finds and uploads
   Terraform provider platform packages to the Tharsis
   Terraform provider mirror. The --platform option can be used
   multiple times to specify more than one platform. By default,
   this command will download all the platforms the latest
   provider version supports.

   Command will only upload missing provider platform packages
   so, if a package ever needs reuploading, the platform mirror
   must be deleted via "tharsis terraform-provider-mirror
   delete-platform" subcommand prior to running this subcommand.

   ---

   Fully Qualified Name (FQN) must be formatted as:

   {registry hostname}/{registry namespace}/{provider name}

   Example: registry.terraform.io/hashicorp/aws

   ---

   Example:

   %s terraform-provider-mirror sync \
      --platform="windows_amd64" \
      --platform="linux_386" \
      --group-path "top-level" \
      registry.terraform.io/hashicorp/aws

   Above command will only find and upload packages for the
   windows_amd64 and linux_386 platforms for the latest
   version of the "aws" provider to "top-level" group's
   provider mirror.

%s

`, c.meta.BinaryName, c.meta.BinaryName, buildHelpText(c.defs()))
}
