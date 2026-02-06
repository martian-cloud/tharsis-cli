package command

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/apparentlymart/go-versions/versions"
	logger "github.com/caarlos0/log"
	"github.com/dustin/go-humanize"
	tfaddr "github.com/hashicorp/terraform-registry-address"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/mitchellh/cli"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/provider"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

const (
	// defaultSyncConcurrency is the default number of concurrent platform uploads.
	defaultSyncConcurrency = 4
	// maxSyncConcurrency is the maximum allowed concurrency.
	maxSyncConcurrency = 10
	// progressBarWidth is the width of the progress bar.
	progressBarWidth = 60
	// progressBarPrefix is the prefix for each progress bar line.
	progressBarPrefix = "   \033[34m•\033[0m "
)

// syncPlatformsInput is the input for syncMissingPlatforms.
type syncPlatformsInput struct {
	client           *tharsis.Client
	registryClient   provider.RegistryProtocol
	versionMirror    *sdktypes.TerraformProviderVersionMirror
	provider         *provider.Provider
	missingPlatforms map[string]struct{}
	registryOpts     []provider.RequestOption
	concurrency      int
	totalBytes       *atomic.Int64
}

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
	concurrencyStr := getOption("concurrency", strconv.Itoa(defaultSyncConcurrency), cmdOpts)[0]
	allPlatforms := len(platforms) == 0 // If no platform specified, default to all supported platforms.
	useLatestVersion := version == ""   // If no version specified, use latest version.

	concurrency, err := strconv.Atoi(concurrencyStr)
	if err != nil {
		c.meta.UI.Error(output.FormatError("invalid concurrency value", err))
		return 1
	}
	if concurrency < 1 || concurrency > maxSyncConcurrency {
		c.meta.UI.Error(output.FormatError(fmt.Sprintf("concurrency must be between 1 and %d", maxSyncConcurrency), nil))
		return 1
	}

	// Extract path from TRN if needed, then validate path (error is already logged by validation function)
	actualPath := trn.ToPath(groupPath)
	if !isNamespacePathValid(c.meta, actualPath) {
		return 1
	}

	for _, p := range platforms {
		if !strings.Contains(p, "_") {
			c.meta.UI.Output(output.FormatError("Invalid platform. Must be formatted as <os_arch>", nil))
			return 1
		}
	}

	// Validate and parse FQN.
	slashCount := strings.Count(fqn, "/")
	if slashCount < 1 || slashCount > 2 {
		c.meta.UI.Error(output.FormatError("invalid FQN. Must be [hostname/]namespace/name", nil))
		return 1
	}

	parsedProvider, err := tfaddr.ParseProviderSource(fqn)
	if err != nil {
		c.meta.UI.Error(output.FormatError("failed to parse provider FQN", err))
		return 1
	}

	logger.Info("starting terraform-provider-mirror sync")

	// Create an instance of the provider registry client.
	registryClient := provider.NewRegistryClient(c.meta.HTTPClient)

	prov := &provider.Provider{
		Hostname:  parsedProvider.Hostname.String(),
		Namespace: parsedProvider.Namespace,
		Type:      parsedProvider.Type,
	}

	// Resolve authentication token for private registries.
	var registryOpts []provider.RequestOption
	token, err := c.resolveRegistryToken(prov.Hostname)
	if err != nil {
		c.meta.UI.Output(output.FormatError("failed to resolve registry token", err))
		return 1
	}

	if token != nil {
		registryOpts = append(registryOpts, provider.WithToken(*token))
	}

	availableVersions, err := registryClient.ListVersions(ctx, prov, registryOpts...)
	if err != nil {
		c.meta.UI.Output(output.FormatError("failed to list available provider versions", err))
		return 1
	}

	logger.Info("retrieved provider's supported platforms from the Terraform Registry API")

	if useLatestVersion {
		version, err = provider.FindLatestVersion(availableVersions)
		if err != nil {
			c.meta.UI.Output(output.FormatError("failed to find latest provider version", err))
			return 1
		}
	} else {
		// Normalize partial versions (e.g., 6.31 -> 6.31.0).
		parsed, err := versions.ParseVersion(version)
		if err != nil {
			c.meta.UI.Output(output.FormatError("invalid version format", err))
			return 1
		}

		version = parsed.String()
	}

	logger.WithField("version", version).Info("using terraform provider version")

	versionQueryInput := &sdktypes.GetTerraformProviderVersionMirrorByAddressInput{
		RegistryHostname:  prov.Hostname,
		RegistryNamespace: prov.Namespace,
		Type:              prov.Type,
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
			RegistryToken:     token,
			GroupPath:         groupPath,
			Type:              prov.Type,
			RegistryNamespace: prov.Namespace,
			RegistryHostname:  prov.Hostname,
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

	// Determine which platforms need to be synced.
	missingPlatforms, err := c.getMissingPlatforms(ctx, client, versionMirror, availableVersions, platforms, allPlatforms)
	if err != nil {
		c.meta.UI.Output(output.FormatError("failed to determine missing platforms", err))
		return 1
	}

	if len(missingPlatforms) == 0 {
		c.meta.UI.Output("\nStopping since all platform packages are already mirrored")
		return 0
	}

	// Sync the missing platforms.
	if err := c.syncMissingPlatforms(ctx, &syncPlatformsInput{
		client:           client,
		registryClient:   registryClient,
		versionMirror:    versionMirror,
		provider:         prov,
		missingPlatforms: missingPlatforms,
		registryOpts:     registryOpts,
		concurrency:      concurrency,
		totalBytes:       &atomic.Int64{},
	}); err != nil {
		c.meta.UI.Output(output.FormatError("failed to sync platform packages", err))
		return 1
	}

	c.meta.UI.Output("\nProvider platform packages uploaded to mirror successfully!")

	return 0
}

// resolveRegistryToken resolves an authentication token for a provider registry.
// It checks: 1) CLI environment variables (TF_TOKEN_...), 2) federated registries via service discovery.
func (c terraformProviderMirrorSyncCommand) resolveRegistryToken(hostname string) (*string, error) {
	if hostname == provider.TerraformPublicRegistryHost {
		c.meta.Logger.Debugf("skipping token resolution for %s", hostname)
		return nil, nil
	}

	// Build the TF_TOKEN_ environment variable name.
	envVar := "TF_TOKEN_" + strings.ReplaceAll(strings.ReplaceAll(hostname, ".", "_"), "-", "__")

	// 1. Check CLI environment variable.
	if token := os.Getenv(envVar); token != "" {
		c.meta.Logger.Debugf("found token in CLI environment variable %s", envVar)
		return &token, nil
	}

	// 2. Run service discovery and look for a federated registry match.
	c.meta.Logger.Debugf("checking for federated registry match for %s", hostname)
	serviceURL, err := c.discoverServiceURL(hostname)
	if err != nil {
		// Discovery failed, assume public registry.
		c.meta.Logger.Debugf("service discovery failed, assuming public registry: %v", err)
		return nil, nil
	}

	c.meta.Logger.Debugf("looking for federated registry profile matching %s", serviceURL.Host)
	currentSettings, err := c.meta.ReadSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to read settings: %w", err)
	}

	for _, profile := range currentSettings.Profiles {
		profileURL, pErr := url.Parse(profile.TharsisURL)
		if pErr != nil {
			continue
		}

		if profileURL.Host == serviceURL.Host && profile.Token != nil {
			c.meta.Logger.Debugf("found federated registry profile for %s", serviceURL.Host)
			return profile.Token, nil
		}
	}

	// No matching profile, assume public registry.
	c.meta.Logger.Debugf("no federated registry match found, assuming public registry")
	return nil, nil
}

// discoverServiceURL wraps disco.DiscoverServiceURL and silences its debug logging.
func (terraformProviderMirrorSyncCommand) discoverServiceURL(hostname string) (*url.URL, error) {
	origLogOutput := log.Writer()
	log.SetOutput(io.Discard)
	defer log.SetOutput(origLogOutput)

	return disco.New().DiscoverServiceURL(svchost.Hostname(hostname), provider.ProvidersServiceID)
}

// getMissingPlatforms returns platforms that need to be synced.
func (c terraformProviderMirrorSyncCommand) getMissingPlatforms(
	ctx context.Context,
	client *tharsis.Client,
	versionMirror *sdktypes.TerraformProviderVersionMirror,
	availableVersions []provider.VersionInfo,
	platforms []string,
	allPlatforms bool,
) (map[string]struct{}, error) {
	c.meta.Logger.Debugf("getting missing platforms for version mirror %s", versionMirror.Metadata.ID)

	platformsInput := &sdktypes.GetTerraformProviderPlatformMirrorsByVersionInput{
		VersionMirrorID: versionMirror.Metadata.ID,
	}

	platformsList, err := client.TerraformProviderPlatformMirror.GetProviderPlatformMirrorsByVersion(ctx, platformsInput)
	if err != nil {
		return nil, fmt.Errorf("failed to list platforms for provider version: %w", err)
	}

	c.meta.Logger.Debugf("found %d existing platforms", len(platformsList))

	existingPlatforms := map[string]struct{}{}
	for _, p := range platformsList {
		existingPlatforms[fmt.Sprintf("%s_%s", p.OS, p.Arch)] = struct{}{}
	}

	missingPlatforms := map[string]struct{}{}
	if allPlatforms {
		target, err := versions.ParseVersion(versionMirror.SemanticVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to parse target provider version: %w", err)
		}

		for _, vi := range availableVersions {
			v, err := versions.ParseVersion(vi.Version)
			if err != nil {
				return nil, fmt.Errorf("failed to parse provider version: %w", err)
			}

			if v.Same(target) {
				for _, p := range vi.Platforms {
					key := fmt.Sprintf("%s_%s", p.OS, p.Arch)
					if _, ok := existingPlatforms[key]; !ok {
						missingPlatforms[key] = struct{}{}
					}
				}

				break
			}
		}
	} else {
		for _, p := range platforms {
			if _, ok := existingPlatforms[p]; !ok {
				missingPlatforms[p] = struct{}{}
			}
		}
	}

	c.meta.Logger.Debugf("found %d missing platforms to sync", len(missingPlatforms))

	return missingPlatforms, nil
}

// syncMissingPlatforms streams packages from upstream registry to mirror.
func (c terraformProviderMirrorSyncCommand) syncMissingPlatforms(ctx context.Context, input *syncPlatformsInput) error {
	logger.Infof("syncing %d platform(s)...", len(input.missingPlatforms))

	start := time.Now()
	progress := mpb.New(mpb.WithWidth(progressBarWidth))

	var wg sync.WaitGroup
	errCh := make(chan error, len(input.missingPlatforms))
	sem := make(chan struct{}, input.concurrency)

	for pf := range input.missingPlatforms {
		wg.Add(1)
		go func(platform string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			err := c.syncPlatform(ctx, input, platform, progress)
			if err != nil {
				errCh <- err
			}
		}(pf)
	}

	wg.Wait()
	progress.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to sync %d platform(s): %v", len(errs), errs)
	}

	logger.Infof("synced %s in %s", humanize.Bytes(uint64(input.totalBytes.Load())), time.Since(start).Round(time.Millisecond))

	return nil
}

func (c terraformProviderMirrorSyncCommand) syncPlatform(ctx context.Context, input *syncPlatformsInput, platform string, progress *mpb.Progress) error {
	c.meta.Logger.Debugf("syncing platform %s", platform)

	parts := strings.Split(platform, "_")
	packageInfo, err := input.registryClient.GetPackageInfo(ctx, input.provider, input.versionMirror.SemanticVersion, parts[0], parts[1], input.registryOpts...)
	if err != nil {
		return fmt.Errorf("%s: failed to find provider package: %w", platform, err)
	}

	c.meta.Logger.Debugf("downloading from %s", packageInfo.DownloadURL)

	body, contentLength, err := input.registryClient.DownloadPackage(ctx, packageInfo.DownloadURL)
	if err != nil {
		return fmt.Errorf("%s: failed to download provider package: %w", platform, err)
	}
	defer body.Close()

	if contentLength == 0 {
		return fmt.Errorf("%s: provider package has no content", platform)
	}

	c.meta.Logger.Debugf("uploading %s to mirror", humanize.Bytes(uint64(contentLength)))

	bar := progress.AddBar(contentLength,
		mpb.PrependDecorators(decor.Name(progressBarPrefix+platform, decor.WCSyncSpaceR)),
		mpb.AppendDecorators(
			decor.CountersKiloByte("% .2f / % .2f"),
			decor.Percentage(decor.WCSyncSpace),
		),
	)
	reader := bar.ProxyReader(body)
	defer reader.Close()

	err = input.client.TerraformProviderPlatformMirror.UploadProviderPlatformPackageToMirror(ctx, &sdktypes.UploadProviderPlatformPackageToMirrorInput{
		VersionMirrorID: input.versionMirror.Metadata.ID,
		Reader:          reader,
		OS:              parts[0],
		Arch:            parts[1],
	})
	if err != nil {
		bar.Abort(true)
		return fmt.Errorf("%s: failed to upload provider package: %w", platform, err)
	}

	input.totalBytes.Add(contentLength)

	return nil
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
		"concurrency": {
			Arguments: []string{"N"},
			Synopsis:  fmt.Sprintf("Number of concurrent platform uploads. Defaults to %d.", defaultSyncConcurrency),
		},
	}
}

func (terraformProviderMirrorSyncCommand) Synopsis() string {
	return "Sync Terraform provider packages from a registry to the provider mirror."
}

func (c terraformProviderMirrorSyncCommand) Help() string {
	return fmt.Sprintf(`
Usage: %s [global options] terraform-provider-mirror sync [options] <FQN>

   The terraform-provider-mirror sync command downloads Terraform
   provider platform packages from a registry and uploads them to
   the Tharsis provider mirror. The --platform option can be used
   multiple times to specify more than one platform. By default,
   this command will sync all platforms for the latest version.

   Command will only upload missing provider platform packages
   so, if a package ever needs reuploading, the platform mirror
   must be deleted via "tharsis terraform-provider-mirror
   delete-platform" subcommand prior to running this subcommand.

   For private registries, authentication tokens are resolved in
   the following order:
   1. CLI environment variable TF_TOKEN_<hostname>
      (e.g., TF_TOKEN_registry_example_com)
   2. Federated registry: runs service discovery and uses the
      token from a matching CLI profile

   ---

   Fully Qualified Name (FQN) must be formatted as:

   [registry hostname/]{registry namespace}/{provider name}

   The hostname can be omitted for providers from the default
   public Terraform registry (registry.terraform.io).

   Examples: registry.terraform.io/hashicorp/aws, hashicorp/aws

   ---

   Example:

   %s terraform-provider-mirror sync \
      --platform="windows_amd64" \
      --platform="linux_386" \
      --group-path "top-level" \
      hashicorp/aws

   Above command will only find and upload packages for the
   windows_amd64 and linux_386 platforms for the latest
   version of the "aws" provider to "top-level" group's
   provider mirror.

%s

`, c.meta.BinaryName, c.meta.BinaryName, buildHelpText(c.defs()))
}
