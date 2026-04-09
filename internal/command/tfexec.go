package command

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/aws/smithy-go/ptr"
	goslug "github.com/hashicorp/go-slug"
	goversion "github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/posener/complete"
	client "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/client"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tfe"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type tfExecCommand struct {
	*BaseCommand

	workspace *string // -workspace flag value
	tfPath    *string // -tf-path flag value
	workDir   *string // -work-dir flag value
}

var _ Command = (*tfExecCommand)(nil)

// NewTfExecCommandFactory returns a factory for the tf-exec passthrough command.
func NewTfExecCommandFactory(baseCommand *BaseCommand) Factory {
	return func() (Command, error) {
		return &tfExecCommand{BaseCommand: baseCommand}, nil
	}
}

// InterceptHelp implements helpInterceptor. When a terraform subcommand is
// present in RawArgs, it forwards help to that terraform subcommand and returns
// the captured output. When only tf-exec flags are present (no terraform
// subcommand), it returns ("", false) so that Wrapper falls back to the
// standard Tharsis CLI help for tf-exec.
func (c *tfExecCommand) InterceptHelp() (string, bool) {
	// RawArgs[0] is the command name ("tf-exec"); subcommand args follow.
	var subArgs []string
	if len(c.RawArgs) > 1 {
		subArgs = c.RawArgs[1:]
	}

	if !isHelpRequest(subArgs) {
		return "", false
	}

	// Parse tf-exec flags; this populates c.tfPath (used by
	// resolveTerraformBinaryForHelp) and leaves terraform args in fs.Args().
	fs := c.Flags()
	fs.SetOutput(io.Discard)
	_ = fs.Parse(subArgs)

	// Nothing after tf-exec flags means help is for tf-exec itself.
	if len(fs.Args()) == 0 {
		return "", false
	}

	binaryPath, err := c.resolveTerraformBinaryForHelp()
	if err != nil {
		return "", false
	}

	if !filepath.IsAbs(binaryPath) {
		return "", false
	}

	cmd := exec.Command(binaryPath, fs.Args()...) // #nosec G204 -- absolute path validated above
	out, _ := cmd.CombinedOutput()                // non-zero exit for --help is expected
	return string(out), true
}

// Run executes the tf-exec passthrough command.
func (c *tfExecCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("tf-exec"),
		WithClient(true),
	); code != 0 {
		return code
	}

	// initialize enforces -workspace as required; c.arguments holds the
	// remaining terraform passthrough args.

	// Get settings for the token getter and endpoint (same pattern as apply.go).
	curSettings, err := c.getCurrentSettings()
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to read current profile for backend configuration")
		return 1
	}
	profile := curSettings.CurrentProfile

	tokenGetter, err := profile.NewTokenGetter(c.Context)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create token getter")
		return 1
	}

	restClient, err := tfe.NewRESTClient(profile.Endpoint, tokenGetter, c.HTTPClient)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create REST client")
		return 1
	}

	workspace, err := c.grpcClient.WorkspacesClient.GetWorkspaceByID(c.Context, &pb.GetWorkspaceByIDRequest{Id: trn.ToTRN(trn.ResourceTypeWorkspace, ptr.ToString(c.workspace))})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get workspace")
		return 1
	}

	lastRun, err := c.findLastAppliedRun(workspace.Metadata.Id)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to find last applied run")
		return 1
	}

	var tfVersionStr string
	if lastRun != nil {
		tfVersionStr = lastRun.TerraformVersion
	}

	binaryPath, err := c.resolveTerraformBinary(tfVersionStr)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to resolve terraform binary")
		return 1
	}

	env, err := c.buildTerraformEnv(workspace.FullPath, profile.Endpoint, tokenGetter)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to build environment")
		return 1
	}

	backendCfg, err := buildTerraformBackendConfig(profile.Endpoint, workspace.FullPath)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to build backend config")
		return 1
	}

	var workDir string
	var persistentDir bool
	var prevCVMarker string
	if ptr.ToString(c.workDir) != "" {
		workDir = ptr.ToString(c.workDir)
		if err := os.MkdirAll(workDir, 0o750); err != nil {
			c.UI.ErrorWithSummary(err, "failed to create work directory")
			return 1
		}
		hasTF, err := hasTerraformFiles(workDir)
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to check for terraform files")
			return 1
		}
		if !hasTF {
			if err := c.downloadLastRunConfig(restClient, lastRun, workDir); err != nil {
				c.UI.ErrorWithSummary(err, "failed to download workspace configuration")
				return 1
			}
		}
	} else {
		persistentDir = true
		workDir, err = persistentWorkdir(profile.Endpoint, workspace.FullPath)
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to determine work directory")
			return 1
		}
		if err := os.MkdirAll(workDir, 0o750); err != nil {
			c.UI.ErrorWithSummary(err, "failed to create work directory")
			return 1
		}
		if data, readErr := os.ReadFile(filepath.Join(workDir, ".tharsis-cv-id")); readErr == nil {
			prevCVMarker = string(data)
		}
		if err := c.downloadLastRunConfig(restClient, lastRun, workDir); err != nil {
			c.UI.ErrorWithSummary(err, "failed to download workspace configuration")
			return 1
		}
	}

	overridePath := filepath.Join(workDir, "override.tf")
	if err := os.WriteFile(overridePath, []byte(backendCfg), 0o600); err != nil {
		c.UI.ErrorWithSummary(err, "failed to write backend config")
		return 1
	}
	if !persistentDir {
		defer os.Remove(overridePath) //nolint:errcheck
	}

	markerPath := filepath.Join(workDir, ".tharsis-cv-id")
	tfDirPath := filepath.Join(workDir, ".terraform")
	if persistentDir {
		currMarker, readErr := os.ReadFile(markerPath)
		_, statErr := os.Stat(tfDirPath)
		versionUnchanged := readErr == nil && len(currMarker) > 0 && string(currMarker) == prevCVMarker
		if versionUnchanged && statErr == nil {
			c.Logger.Debug("skipping terraform init: config version already initialised", "configVersion", string(currMarker), "workDir", workDir)
			return runExec(binaryPath, c.arguments, env, workDir)
		}
	}

	if code := runExec(binaryPath, []string{"init"}, env, workDir); code != 0 {
		return code
	}

	return runExec(binaryPath, c.arguments, env, workDir)
}

// Synopsis returns a short description of the command.
func (*tfExecCommand) Synopsis() string {
	return "Run terraform with Tharsis auth and workspace variables injected."
}

// PredictArgs returns shell-completion candidates for the terraform subcommand
// positional argument.
func (*tfExecCommand) PredictArgs() complete.Predictor {
	return complete.PredictSet(
		"apply", "console", "destroy", "force-unlock", "get", "graph",
		"import", "metadata", "output", "plan", "providers", "refresh",
		"show", "state", "taint", "test", "untaint", "validate",
	)
}

// Usage returns the usage string for the command.
func (*tfExecCommand) Usage() string {
	return "tharsis [global options] tf-exec -workspace <path|trn> [-tf-path <path>] [-work-dir <path>] [terraform-args...]"
}

// Description returns the full description for the command.
func (*tfExecCommand) Description() string {
	return `
  Runs terraform with Tharsis authentication and workspace variables
  automatically injected into the process environment.

  Available Terraform Subcommands:

    apply          Apply the changes required to reach the desired state
    console        Try Terraform expressions at an interactive command prompt
    destroy        Destroy previously-created infrastructure
    force-unlock   Release a stuck lock on the current workspace
    get            Install or upgrade remote Terraform modules
    graph          Generate a Graphviz graph of the steps in an operation
    import         Associate existing infrastructure with a Terraform resource
    metadata       Metadata related commands
    output         Show output values from your root module
    plan           Show changes required by the current configuration
    providers      Show the providers required for this configuration
    refresh        Update the state to match remote systems
    show           Show the current state or a saved plan
    state          Advanced state management
    taint          Mark a resource instance as not fully functional
    test           Execute integration tests for a module
    untaint        Remove the 'tainted' state from a resource instance
    validate       Check whether the configuration is valid

  Terraform Binary Resolution:

    When -tf-path is not provided, tharsis looks for a cached terraform binary
    matching the workspace's configured version in ~/.tharsis/tf-installs/<version>/.
    If not found, it downloads that exact version from releases.hashicorp.com and
    caches it there for future use.

  Authentication:

    The current profile's auth token is injected as TF_TOKEN_<host> where <host>
    is the Tharsis instance hostname with dots replaced by underscores. This
    authenticates terraform against the Tharsis remote backend.

  Variables:

    All variables configured on the workspace and its parent groups are injected:

      - Terraform variables (category: terraform) -> TF_VAR_<key>=<value>
      - Environment variables (category: environment) -> <key>=<value>

    Sensitive variable values are automatically fetched and injected.

  Exit Code:

    The exact exit code returned by terraform is passed through unchanged.
`
}

// Example returns example usage for the command.
func (*tfExecCommand) Example() string {
	return `
tharsis tf-exec -workspace my/group/workspace show
tharsis tf-exec -workspace trn:workspace:my/group/workspace plan
tharsis tf-exec -workspace my/group/workspace -work-dir ./infra apply
`
}

// Flags returns the flag set for the command.
func (c *tfExecCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")

	f.StringVar(&c.workspace, "workspace", "The Tharsis workspace path or TRN (e.g. my/group/workspace or trn:workspace:my/group/workspace).", flag.Required())
	f.StringVar(&c.tfPath, "tf-path", "Path to an existing terraform binary. If omitted, the version from the last applied run is downloaded automatically.")
	f.StringVar(&c.workDir, "work-dir", "Working directory for terraform. If omitted, a persistent cache directory keyed by workspace is used.")

	return f
}

// findLastAppliedRun pages through runs for the given workspace ID and returns
// the most recent run with status "applied" that has a configuration version.
// Returns (nil, nil) when no such run exists.
func (c *tfExecCommand) findLastAppliedRun(workspaceID string) (*pb.Run, error) {
	sort := pb.RunSortableField_CREATED_AT_DESC
	pageSize := int32(20)
	var cursor *string

	for {
		resp, err := c.grpcClient.RunsClient.GetRuns(c.Context, &pb.GetRunsRequest{
			Sort: &sort,
			PaginationOptions: &pb.PaginationOptions{
				First: &pageSize,
				After: cursor,
			},
			WorkspaceId: &workspaceID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list runs: %w", err)
		}
		for _, run := range resp.Runs {
			if run.Status == "applied" && run.ConfigurationVersionId != nil {
				return run, nil
			}
		}
		if resp.PageInfo == nil || !resp.PageInfo.HasNextPage {
			break
		}
		cursor = resp.PageInfo.EndCursor
	}
	return nil, nil
}

// resolveTerraformBinaryForHelp returns a terraform binary path suitable for
// help passthrough, without requiring a workspace. Uses -tf-path when provided,
// otherwise falls back to whatever is on PATH.
func (c *tfExecCommand) resolveTerraformBinaryForHelp() (string, error) {
	if p := ptr.ToString(c.tfPath); p != "" {
		if _, err := os.Stat(p); err != nil {
			return "", fmt.Errorf("-tf-path %q: %w", p, err)
		}
		return p, nil
	}
	path, err := exec.LookPath("terraform")
	if err != nil {
		return "", fmt.Errorf("terraform not found on PATH and -tf-path not set")
	}
	return path, nil
}

// resolveTerraformBinary finds an existing or downloads the required terraform binary.
func (c *tfExecCommand) resolveTerraformBinary(tfVersionStr string) (string, error) {
	if p := ptr.ToString(c.tfPath); p != "" {
		if _, err := os.Stat(p); err != nil {
			return "", fmt.Errorf("-tf-path %q: %w", p, err)
		}
		return p, nil
	}

	if tfVersionStr == "" {
		return "", fmt.Errorf("no previous applied run found to determine Terraform version; use -tf-path to specify a binary")
	}

	tfVersion, err := goversion.NewVersion(tfVersionStr)
	if err != nil {
		return "", fmt.Errorf("invalid terraform version %q: %w", tfVersionStr, err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %w", err)
	}
	cacheDir := filepath.Join(homeDir, ".tharsis", "tf-installs", tfVersionStr)

	finder := &fs.ExactVersion{
		Product:    product.Terraform,
		Version:    tfVersion,
		ExtraPaths: []string{cacheDir},
	}
	if path, err := finder.Find(c.Context); err == nil {
		c.Logger.Debug("found cached terraform binary", "path", path)
		return path, nil
	}

	c.Logger.Debug("downloading terraform", "version", tfVersionStr, "cacheDir", cacheDir)
	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		return "", fmt.Errorf("failed to create cache directory %s: %w", cacheDir, err)
	}

	installer := install.NewInstaller()
	path, err := installer.Install(c.Context, []src.Installable{
		&releases.ExactVersion{
			Product:    product.Terraform,
			Version:    tfVersion,
			InstallDir: cacheDir,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to install terraform %s: %w", tfVersionStr, err)
	}
	return path, nil
}

// downloadLastRunConfig fetches the configuration version associated with the
// last applied run and unpacks it into destDir.
// If lastRun is nil (no qualifying run), the function is a no-op.
func (c *tfExecCommand) downloadLastRunConfig(restClient tfe.RESTClient, lastRun *pb.Run, destDir string) error {
	if lastRun == nil {
		c.Logger.Debug("no applied run available, proceeding without config download", "destDir", destDir)
		return nil
	}

	cvID := *lastRun.ConfigurationVersionId

	markerPath := filepath.Join(destDir, ".tharsis-cv-id")
	if existing, err := os.ReadFile(markerPath); err == nil && string(existing) == cvID {
		c.Logger.Debug("configuration version already present, skipping download", "configVersion", cvID, "destDir", destDir)
		return nil
	}

	c.Logger.Debug("downloading configuration version", "configVersion", cvID, "destDir", destDir)

	tmpFile, err := os.CreateTemp("", "config-version-*.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to create temp file for config download: %w", err)
	}
	defer tmpFile.Close()           //nolint:errcheck
	defer os.Remove(tmpFile.Name()) //nolint:errcheck

	if err := restClient.DownloadConfigurationVersion(c.Context, &tfe.DownloadConfigurationVersionInput{
		ConfigVersionID: cvID,
		Writer:          tmpFile,
	}); err != nil {
		return fmt.Errorf("failed to download configuration version: %w", err)
	}

	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek config archive: %w", err)
	}

	if cleanErr := cleanDestDir(destDir); cleanErr != nil {
		return fmt.Errorf("failed to clean destination directory before unpack: %w", cleanErr)
	}

	if err := goslug.Unpack(tmpFile, destDir); err != nil {
		return fmt.Errorf("failed to unpack configuration: %w", err)
	}

	if err := os.WriteFile(markerPath, []byte(cvID), 0o600); err != nil {
		c.Logger.Info("failed to write config version marker", "error", err)
	}

	return nil
}

// buildTerraformEnv constructs the environment for the terraform subprocess,
// injecting the Tharsis auth token and all workspace variables.
func (c *tfExecCommand) buildTerraformEnv(namespacePath, tharsisURL string, tokenGetter client.TokenGetter) ([]string, error) {
	env := os.Environ()

	if _, err := buildTFTokenEnvKey(tharsisURL); err != nil {
		c.Logger.Debug("skipping TF_TOKEN injection: could not derive env key from URL", "url", tharsisURL, "error", err)
	} else {
		tokenStr, err := tokenGetter.Token(c.Context)
		if err != nil {
			c.Logger.Debug("skipping TF_TOKEN injection: could not get token", "error", err)
		} else {
			var tokenPtr *string
			if tokenStr != "" {
				tokenPtr = &tokenStr
			}
			env = appendAuthToken(env, tharsisURL, tokenPtr)
		}
	}

	resp, err := c.grpcClient.NamespaceVariablesClient.GetNamespaceVariables(c.Context,
		&pb.GetNamespaceVariablesRequest{NamespacePath: namespacePath})
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace variables: %w", err)
	}

	for _, variable := range resp.Variables {
		if strings.ContainsAny(variable.Key, "=\x00") {
			c.Logger.Debug("skipping variable with invalid key", "key", variable.Key)
			continue
		}

		value := variable.Value

		if value == nil && variable.Sensitive && variable.LatestVersionId != "" {
			version, vErr := c.grpcClient.NamespaceVariablesClient.GetNamespaceVariableVersionByID(c.Context,
				&pb.GetNamespaceVariableVersionByIDRequest{
					Id:                    variable.LatestVersionId,
					IncludeSensitiveValue: true,
				})
			if vErr != nil {
				c.Logger.Debug("failed to fetch sensitive value", "key", variable.Key, "error", vErr)
			} else if version != nil {
				value = version.Value
			}
		}

		if value == nil {
			c.Logger.Debug("skipping variable: no value available", "key", variable.Key)
			continue
		}

		switch variable.Category {
		case "terraform":
			env = append(env, "TF_VAR_"+variable.Key+"="+*value)
		case "environment":
			env = append(env, variable.Key+"="+*value)
		}
	}

	return env, nil
}

// isHelpRequest reports whether any arg in tfArgs is a help flag.
func isHelpRequest(tfArgs []string) bool {
	for _, arg := range tfArgs {
		if arg == "-help" || arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

// runExec runs the binary with args, env, and an optional working directory,
// returning the subprocess exit code.
// binaryPath must be an absolute path; a relative or bare name is rejected to
// prevent PATH-manipulation attacks (CWE-78 / gosec G204).
func runExec(binaryPath string, args, env []string, dir string) int {
	if !filepath.IsAbs(binaryPath) {
		fmt.Fprintf(os.Stderr, "tf-exec: binary path must be absolute, got %q\n", binaryPath)
		return 1
	}
	cmd := exec.Command(binaryPath, args...) // #nosec G204 -- path is caller-validated absolute path from resolveTerraformBinary/resolveTerraformBinaryForHelp
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env
	if dir != "" {
		cmd.Dir = dir
	}
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			}
		}
		return 1
	}
	return 0
}

// appendAuthToken injects the Tharsis auth token into env as TF_TOKEN_<sanitized_host>.
func appendAuthToken(env []string, tharsisURL string, token *string) []string {
	if token == nil || tharsisURL == "" {
		return env
	}
	envKey, err := buildTFTokenEnvKey(tharsisURL)
	if err != nil {
		return env
	}
	return append(env, envKey+"="+*token)
}

// buildTFTokenEnvKey converts a Tharsis URL into the TF_TOKEN_* environment
// variable name that terraform uses for backend authentication.
func buildTFTokenEnvKey(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	host := parsed.Hostname()
	if host == "" {
		return "", fmt.Errorf("no hostname in URL %q", rawURL)
	}
	sanitized := strings.ReplaceAll(host, ".", "_")
	return "TF_TOKEN_" + sanitized, nil
}

// buildTerraformBackendConfig renders the HCL backend configuration block that
// points terraform at the Tharsis remote backend for the given workspace path.
func buildTerraformBackendConfig(tharsisURL, workspacePath string) (string, error) {
	parsed, err := url.Parse(tharsisURL)
	if err != nil {
		return "", fmt.Errorf("invalid TharsisURL %q: %w", tharsisURL, err)
	}
	hostname := parsed.Hostname()
	if hostname == "" {
		return "", fmt.Errorf("no hostname in TharsisURL %q", tharsisURL)
	}

	trimmed := strings.Trim(workspacePath, "/")
	if trimmed == "" {
		return "", fmt.Errorf("workspace path is empty")
	}

	segments := strings.Split(trimmed, "/")
	if len(segments) < 2 {
		return "", fmt.Errorf("workspace path %q must have at least two segments (group/workspace)", workspacePath)
	}

	groupPath := strings.Join(segments[:len(segments)-1], ".")
	workspaceName := segments[len(segments)-1]

	cfg := fmt.Sprintf(`terraform {
  backend "remote" {
    hostname     = %q
    organization = %q
    workspaces {
      name = %q
    }
  }
}
`, hostname, groupPath, workspaceName)

	return cfg, nil
}

// cleanDestDir removes all entries from dir except ".terraform" and ".tharsis-cv-id".
func cleanDestDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if name == ".terraform" || name == ".tharsis-cv-id" {
			continue
		}
		if err := os.RemoveAll(filepath.Join(dir, name)); err != nil {
			return err
		}
	}
	return nil
}

// hasTerraformFiles reports whether dir contains at least one .tf file at the top level.
func hasTerraformFiles(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".tf") {
			return true, nil
		}
	}
	return false, nil
}

// persistentWorkdir returns a stable cache directory path keyed by the SHA-256
// of the Tharsis hostname and workspace path.
func persistentWorkdir(tharsisURL, workspacePath string) (string, error) {
	parsed, err := url.Parse(tharsisURL)
	if err != nil {
		return "", fmt.Errorf("invalid TharsisURL %q: %w", tharsisURL, err)
	}
	hostname := parsed.Hostname()
	if hostname == "" {
		return "", fmt.Errorf("no hostname in TharsisURL %q", tharsisURL)
	}
	sum := sha256.Sum256([]byte(hostname + ":" + workspacePath))
	cacheBase, err := os.UserCacheDir()
	if err != nil {
		cacheBase = os.TempDir()
	}
	return filepath.Join(cacheBase, "tharsis-tf-workdirs", hex.EncodeToString(sum[:])), nil
}
