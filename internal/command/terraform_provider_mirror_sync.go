package command

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/apparentlymart/go-versions/versions"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	tfaddr "github.com/hashicorp/terraform-registry-address"
	svchost "github.com/hashicorp/terraform-svchost"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/provider"
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

	sg            terminal.StepGroup
	rootGroupName string
	version       string
	platforms     []string
}

func (c *terraformProviderMirrorSyncCommand) validate() error {
	const message = "provider-fqn is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.rootGroupName, validation.Required),
	)
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

	versionMirror, err := c.getOrCreateVersionMirror(registry)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get version mirror")
		return 1
	}

	if err := c.uploadMissingPlatforms(registry, versionMirror); err != nil {
		c.UI.ErrorWithSummary(err, "failed to upload platforms")
		return 1
	}

	c.UI.Successf("\nProvider platform packages uploaded to mirror successfully!")
	return 0
}

// connectToRegistry establishes connection to the upstream provider registry and resolves authentication.
func (c *terraformProviderMirrorSyncCommand) connectToRegistry() (*registryConnection, error) {
	parsedProvider, err := tfaddr.ParseProviderSource(c.arguments[0])
	if err != nil {
		return nil, err
	}

	step := c.sg.Add("Connect to provider registry")

	registryClient := provider.NewRegistryClient(c.HTTPClient)

	// Resolve authentication token for private registries.
	token, err := c.resolveRegistryToken(parsedProvider.Hostname.String())
	if err != nil {
		step.Abort()
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
		step.Abort()
		return nil, err
	}
	step.Done()

	return &registryConnection{
		provider:          prov,
		client:            registryClient,
		requestOpts:       registryOpts,
		token:             token,
		availableVersions: availableVersions,
	}, nil
}

// getOrCreateVersionMirror resolves the version and gets or creates the version mirror.
func (c *terraformProviderMirrorSyncCommand) getOrCreateVersionMirror(registry *registryConnection) (*pb.TerraformProviderVersionMirror, error) {
	if c.version == "" {
		step := c.sg.Add("Find latest version")
		version, err := provider.FindLatestVersion(registry.availableVersions)
		if err != nil {
			step.Abort()
			return nil, err
		}
		c.version = version
		step.Done()
	} else {
		// Normalize partial versions (e.g., 6.31 -> 6.31.0).
		parsed, err := versions.ParseVersion(c.version)
		if err != nil {
			return nil, err
		}
		c.version = parsed.String()
	}

	step := c.sg.Add("Get version mirror")
	versionMirrorTRN := trn.NewResourceTRN(
		trn.ResourceTypeTerraformProviderVersionMirror,
		c.rootGroupName,
		registry.provider.Hostname,
		registry.provider.Namespace,
		registry.provider.Type,
		c.version,
	)

	// Attempt to find the version mirror first in case it already exists.
	versionMirror, err := c.grpcClient.TerraformProviderMirrorsClient.GetTerraformProviderVersionMirrorByID(c.Context, &pb.GetTerraformProviderVersionMirrorByIDRequest{
		Id: versionMirrorTRN,
	})
	if err != nil && status.Code(err) != codes.NotFound {
		step.Abort()
		return nil, err
	}

	if versionMirror == nil {
		step.Done()
		step = c.sg.Add("Create version mirror")
		// Version mirror doesn't exist, so create it.

		versionMirror, err = c.grpcClient.TerraformProviderMirrorsClient.CreateTerraformProviderVersionMirror(c.Context, &pb.CreateTerraformProviderVersionMirrorRequest{
			GroupPath:         c.rootGroupName,
			Type:              registry.provider.Type,
			RegistryNamespace: registry.provider.Namespace,
			RegistryHostname:  registry.provider.Hostname,
			SemanticVersion:   c.version,
			RegistryToken:     registry.token,
		})
		if err != nil {
			step.Abort()
			return nil, err
		}
	}
	step.Done()

	return versionMirror, nil
}

// uploadMissingPlatforms determines which platforms need to be synced and uploads them to the mirror.
func (c *terraformProviderMirrorSyncCommand) uploadMissingPlatforms(registry *registryConnection, versionMirror *pb.TerraformProviderVersionMirror) error {
	step := c.sg.Add("Determine missing platforms")
	missingPlatforms, err := c.getMissingPlatforms(versionMirror, registry.availableVersions)
	if err != nil {
		step.Abort()
		return err
	}
	step.Done()

	if len(missingPlatforms) == 0 {
		c.UI.Output("\nAll platform packages are already mirrored")
		return nil
	}

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
		parts := strings.Split(platform, "_")
		if len(parts) != 2 {
			return fmt.Errorf("invalid platform format: %s", platform)
		}
		os, arch := parts[0], parts[1]

		step := c.sg.Add("Upload platform %s", platform)

		packageInfo, err := registry.client.GetPackageInfo(c.Context, registry.provider, c.version, os, arch, registry.requestOpts...)
		if err != nil {
			step.Abort()
			return err
		}

		reader, _, err := registry.client.DownloadPackage(c.Context, packageInfo.DownloadURL)
		if err != nil {
			step.Abort()
			return err
		}
		defer reader.Close()

		if err := tfeClient.UploadProviderPlatformPackageToMirror(c.Context, &tfe.UploadProviderPlatformPackageToMirrorInput{
			VersionMirrorID: versionMirror.Metadata.Id,
			OS:              os,
			Arch:            arch,
			Reader:          reader,
		}); err != nil {
			step.Abort()
			return err
		}

		step.Done()
	}

	return nil
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
	serviceURL, err := provider.NewQuietDisco().DiscoverServiceURL(svchost.Hostname(hostname), provider.ProvidersServiceID)
	if err != nil {
		// Discovery failed, assume public registry.
		return nil, nil
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

   Fully Qualified Name (FQN) must be formatted as:

   [registry hostname/]{registry namespace}/{provider name}

   The hostname can be omitted for providers from the default
   public Terraform registry (registry.terraform.io).

   Examples: registry.terraform.io/hashicorp/aws, hashicorp/aws
`
}

func (*terraformProviderMirrorSyncCommand) Usage() string {
	return "tharsis [global options] terraform-provider-mirror sync [options] <provider_fqn>"
}

func (*terraformProviderMirrorSyncCommand) Example() string {
	return `
tharsis terraform-provider-mirror sync \
  --root-group-name my-group \
  --version 1.0.0 \
  --platform linux_amd64 \
  hashicorp/aws
`
}

func (c *terraformProviderMirrorSyncCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.rootGroupName,
		"root-group-name",
		"",
		"The name of the root group to create the mirror in.",
	)
	f.StringVar(
		&c.rootGroupName,
		"group-path",
		"",
		"Full path to the root group where this Terraform provider version will be mirrored. Deprecated.",
	)
	f.StringVar(
		&c.version,
		"version",
		"",
		"The provider version to sync. If not specified, uses the latest version.",
	)
	f.Func(
		"platform",
		"Platform to sync (format: os_arch). Can be specified multiple times. If not specified, syncs all platforms.",
		func(s string) error {
			if len(strings.Split(s, "_")) != 2 {
				return fmt.Errorf("invalid platform format, must be os_arch")
			}
			c.platforms = append(c.platforms, s)
			return nil
		},
	)
	return f
}
