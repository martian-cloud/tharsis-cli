package command

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/apparentlymart/go-versions/versions"
	tfaddr "github.com/hashicorp/terraform-registry-address"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/provider"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tfe"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type registryConnection struct {
	provider          *provider.Provider
	client            provider.RegistryProtocol
	requestOpts       []provider.RequestOption
	token             *string
	availableVersions []provider.VersionInfo
}

type terraformProviderMirrorSyncCommand struct {
	*BaseCommand

	sg        terminal.StepGroup
	groupID   *string
	version   *string
	platforms []string
}

var _ Command = (*terraformProviderMirrorSyncCommand)(nil)

func (c *terraformProviderMirrorSyncCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: provider fqn")
	}

	if c.groupID == nil {
		return errors.New("group id is required")
	}

	return nil
}

// NewTerraformProviderMirrorSyncCommandFactory returns a new terraformProviderMirrorSyncCommand
func NewTerraformProviderMirrorSyncCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderMirrorSyncCommand{
			BaseCommand: baseCommand,
			sg:          baseCommand.UI.StepGroup(),
		}, nil
	}
}

func (c *terraformProviderMirrorSyncCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider-mirror sync"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	registry, err := c.connectToRegistry()
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to connect to registry")
		return 1
	}

	if err := c.resolveVersion(registry); err != nil {
		c.UI.ErrorWithSummary(err, "failed to resolve version")
		return 1
	}

	versionMirror, err := c.getOrCreateVersionMirror(registry)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get version mirror")
		return 1
	}

	if err := c.uploadMissingPlatforms(registry, versionMirror); err != nil {
		c.UI.ErrorWithSummary(err, "failed to upload platforms")
		return 1
	}

	c.sg.Wait()
	c.UI.Successf("\nProvider platform packages uploaded to mirror successfully!")
	return 0
}

func (c *terraformProviderMirrorSyncCommand) connectToRegistry() (conn *registryConnection, err error) {
	parsedProvider, err := tfaddr.ParseProviderSource(c.arguments[0])
	if err != nil {
		return nil, err
	}

	step := c.sg.Add("Connect to provider registry")
	defer func() { c.finalizeStep(step, err) }()

	registryClient := provider.NewRegistryClient(c.HTTPClient)

	token, err := c.resolveRegistryToken(parsedProvider.Hostname.String())
	if err != nil {
		return nil, err
	}

	prov := &provider.Provider{
		Hostname:  parsedProvider.Hostname.String(),
		Namespace: parsedProvider.Namespace,
		Type:      parsedProvider.Type,
	}

	var registryOpts []provider.RequestOption
	if token != nil {
		registryOpts = append(registryOpts, provider.WithToken(*token))
	}

	availableVersions, err := registryClient.ListVersions(c.Context, prov, registryOpts...)
	if err != nil {
		return nil, err
	}

	step.Update("Connect to provider registry (%d versions available)", len(availableVersions))

	return &registryConnection{
		provider:          prov,
		client:            registryClient,
		requestOpts:       registryOpts,
		token:             token,
		availableVersions: availableVersions,
	}, nil
}

func (c *terraformProviderMirrorSyncCommand) resolveVersion(registry *registryConnection) (err error) {
	if c.version != nil {
		// Normalize partial versions (e.g., 6.31 -> 6.31.0).
		parsed, err := versions.ParseVersion(*c.version)
		if err != nil {
			return err
		}
		*c.version = parsed.String()

		return nil
	}

	step := c.sg.Add("Find latest version")
	defer func() { c.finalizeStep(step, err) }()

	version, err := provider.FindLatestVersion(registry.availableVersions)
	if err != nil {
		return err
	}

	c.version = &version
	step.Update("Find latest version (%s)", *c.version)

	return nil
}

func (c *terraformProviderMirrorSyncCommand) getOrCreateVersionMirror(registry *registryConnection) (versionMirror *pb.TerraformProviderVersionMirror, err error) {
	step := c.sg.Add("Get version mirror")
	defer func() { c.finalizeStep(step, err) }()

	group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{
		Id: *c.groupID,
	})
	if err != nil {
		return nil, err
	}

	versionMirrorTRN := trn.NewResourceTRN(
		trn.ResourceTypeTerraformProviderVersionMirror,
		group.FullPath,
		registry.provider.Hostname,
		registry.provider.Namespace,
		registry.provider.Type,
		*c.version,
	)

	versionMirror, err = c.grpcClient.TerraformProviderMirrorsClient.GetTerraformProviderVersionMirrorByID(c.Context, &pb.GetTerraformProviderVersionMirrorByIDRequest{
		Id: versionMirrorTRN,
	})
	if err != nil && status.Code(err) != codes.NotFound {
		return nil, err
	}

	if versionMirror != nil {
		return versionMirror, nil
	}

	step.Update("Create version mirror")

	versionMirror, err = c.grpcClient.TerraformProviderMirrorsClient.CreateTerraformProviderVersionMirror(c.Context, &pb.CreateTerraformProviderVersionMirrorRequest{
		GroupPath:         group.FullPath,
		Type:              registry.provider.Type,
		RegistryNamespace: registry.provider.Namespace,
		RegistryHostname:  registry.provider.Hostname,
		SemanticVersion:   *c.version,
		RegistryToken:     registry.token,
	})

	return versionMirror, err
}

func (c *terraformProviderMirrorSyncCommand) uploadMissingPlatforms(registry *registryConnection, versionMirror *pb.TerraformProviderVersionMirror) (err error) {
	step := c.sg.Add("Determine missing platforms")
	defer func() { c.finalizeStep(step, err) }()

	missingPlatforms, err := c.getMissingPlatforms(versionMirror, registry.availableVersions)
	if err != nil {
		return err
	}

	step.Update("Determine missing platforms (%d to sync)", len(missingPlatforms))

	if len(missingPlatforms) == 0 {
		return nil
	}

	step.Done()

	curSettings, err := c.getCurrentSettings()
	if err != nil {
		return err
	}

	tokenGetter, err := curSettings.CurrentProfile.NewTokenGetter(c.Context)
	if err != nil {
		return err
	}

	tfeClient, err := tfe.NewRESTClient(curSettings.CurrentProfile.Endpoint, tokenGetter, c.HTTPClient)
	if err != nil {
		return err
	}

	for platform := range missingPlatforms {
		if err := c.uploadPlatform(tfeClient, registry, versionMirror, platform); err != nil {
			return err
		}
	}

	return nil
}

func (c *terraformProviderMirrorSyncCommand) uploadPlatform(tfeClient tfe.RESTClient, registry *registryConnection, versionMirror *pb.TerraformProviderVersionMirror, platform string) (err error) {
	step := c.sg.Add("Upload platform %s", platform)
	start := time.Now()
	defer func() {
		step.Update("Upload platform %s (%s)", platform, time.Since(start).Round(time.Millisecond))
		c.finalizeStep(step, err)
	}()

	parts := strings.Split(platform, "_")
	if len(parts) != 2 {
		return fmt.Errorf("invalid platform format: %s", platform)
	}
	os, arch := parts[0], parts[1]

	packageInfo, err := registry.client.GetPackageInfo(c.Context, registry.provider, *c.version, os, arch, registry.requestOpts...)
	if err != nil {
		return err
	}

	reader, _, err := registry.client.DownloadPackage(c.Context, packageInfo.DownloadURL)
	if err != nil {
		return err
	}
	defer reader.Close()

	return tfeClient.UploadProviderPlatformPackageToMirror(c.Context, &tfe.UploadProviderPlatformPackageToMirrorInput{
		VersionMirrorID: versionMirror.Metadata.Id,
		OS:              os,
		Arch:            arch,
		Reader:          reader,
	})
}

func (c *terraformProviderMirrorSyncCommand) finalizeStep(step terminal.Step, err error) {
	if err != nil {
		step.Abort()
		c.sg.Wait()

		return
	}

	step.Done()
}

// resolveRegistryToken resolves an authentication token for a provider registry.
// It checks: 1) CLI environment variables (TF_TOKEN_...), 2) federated registries via service discovery.
func (c *terraformProviderMirrorSyncCommand) resolveRegistryToken(hostname string) (*string, error) {
	if hostname == provider.TerraformPublicRegistryHost {
		return nil, nil
	}

	// Build the TF_TOKEN_ environment variable name.
	envVar := "TF_TOKEN_" + strings.ReplaceAll(strings.ReplaceAll(hostname, ".", "_"), "-", "__")

	// 1. Check CLI environment variable.
	if token := os.Getenv(envVar); token != "" {
		return &token, nil
	}

	// 2. Run service discovery and look for a federated registry match.
	discovered, err := provider.NewServiceDiscoverer(c.HTTPClient).DiscoverTFEServices(c.Context, hostname)
	if err != nil {
		// Discovery failed, assume public registry.
		return nil, nil
	}

	serviceURL, ok := discovered.Services[provider.ProvidersServiceID]
	if !ok {
		return nil, fmt.Errorf("service url for %q not found", provider.ProvidersServiceID)
	}

	curSettings, err := c.getCurrentSettings()
	if err != nil {
		return nil, err
	}

	for _, profile := range curSettings.Profiles {
		profileURL, pErr := url.Parse(profile.Endpoint)
		if pErr != nil {
			continue
		}

		if profileURL.Host == serviceURL.Host {
			tokenGetter, err := profile.NewTokenGetter(c.Context)
			if err != nil {
				continue
			}

			token, err := tokenGetter.Token(c.Context)
			if err != nil {
				continue
			}

			return &token, nil
		}
	}

	// No matching profile, assume public registry.
	return nil, nil
}

// getMissingPlatforms returns platforms that need to be synced.
func (c *terraformProviderMirrorSyncCommand) getMissingPlatforms(
	versionMirror *pb.TerraformProviderVersionMirror,
	availableVersions []provider.VersionInfo,
) (map[string]struct{}, error) {
	resp, err := c.grpcClient.TerraformProviderMirrorsClient.GetTerraformProviderPlatformMirrors(c.Context, &pb.GetTerraformProviderPlatformMirrorsRequest{
		VersionMirrorId: versionMirror.Metadata.Id,
	})
	if err != nil {
		return nil, err
	}

	existingPlatforms := map[string]struct{}{}
	for _, p := range resp.PlatformMirrors {
		existingPlatforms[fmt.Sprintf("%s_%s", p.Os, p.Architecture)] = struct{}{}
	}

	missingPlatforms := map[string]struct{}{}
	if len(c.platforms) == 0 {
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
		for _, p := range c.platforms {
			if _, ok := existingPlatforms[p]; !ok {
				missingPlatforms[p] = struct{}{}
			}
		}
	}

	return missingPlatforms, nil
}

func (*terraformProviderMirrorSyncCommand) Synopsis() string {
	return "Sync provider platforms from upstream registry to mirror."
}

func (*terraformProviderMirrorSyncCommand) Description() string {
	return `
   Downloads provider platform packages from an upstream
   registry and uploads them to the Tharsis mirror. Use
   -platform multiple times to specify platforms. By default,
   syncs all platforms for the latest version.

   Only missing packages are uploaded. To re-upload, delete
   the platform mirror first via "tharsis
   terraform-provider-mirror delete-platform".

   For private registries, tokens are resolved in order:
   1. TF_TOKEN_<hostname> environment variable
   2. Federated registry service discovery with a
      matching CLI profile

   Fully Qualified Name (FQN) format:

   [registry hostname/]{namespace}/{provider name}

   The hostname can be omitted for providers from the
   default public registry (registry.terraform.io).

   Examples: registry.terraform.io/hashicorp/aws, hashicorp/aws
`
}

func (*terraformProviderMirrorSyncCommand) Usage() string {
	return "tharsis [global options] terraform-provider-mirror sync [options] <provider_fqn>"
}

func (*terraformProviderMirrorSyncCommand) Example() string {
	return `
tharsis terraform-provider-mirror sync \
  -group-id "my-group" \
  -version "1.0.0" \
  -platform "linux_amd64" \
  hashicorp/aws
`
}

func (c *terraformProviderMirrorSyncCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.groupID,
		"group-id",
		"The ID of the root group to create the mirror in.",
	)
	f.StringVar(
		&c.groupID,
		"group-path",
		"Full path to the root group where this Terraform provider version will be mirrored.",
		flag.Deprecated("use -group-id"),
		flag.TransformString(func(s string) string {
			return trn.NewResourceTRN(trn.ResourceTypeGroup, s)
		}),
	)
	f.StringVar(
		&c.version,
		"version",
		"The provider version to sync. If not specified, uses the latest version.",
	)
	f.StringSliceVar(
		&c.platforms,
		"platform",
		"Platform to sync (format: os_arch). If not specified, syncs all platforms.",
	)

	return f
}
