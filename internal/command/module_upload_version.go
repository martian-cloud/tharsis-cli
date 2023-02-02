package command

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/log"
	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/slug"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// moduleUploadVersionCommand is the top-level structure for the module upload-version command.
type moduleUploadVersionCommand struct {
	meta *Metadata
}

// NewModuleUploadVersionCommandFactory returns a moduleUploadVersionCommand struct.
func NewModuleUploadVersionCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return moduleUploadVersionCommand{
			meta: meta,
		}, nil
	}
}

func (muc moduleUploadVersionCommand) Run(args []string) int {
	muc.meta.Logger.Debugf("Starting the 'module upload-version' command with %d arguments:", len(args))
	for ix, arg := range args {
		muc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Cannot delay reading settings past this point.
	settings, err := muc.meta.ReadSettings()
	if err != nil {
		muc.meta.Logger.Error(output.FormatError("failed to read settings file", err))
		return 1
	}

	client, err := settings.CurrentProfile.GetSDKClient()
	if err != nil {
		muc.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return muc.doModuleUploadVersion(ctx, client, args)
}

func (muc moduleUploadVersionCommand) doModuleUploadVersion(ctx context.Context, client *tharsis.Client, opts []string) int {
	muc.meta.Logger.Debugf("will do module upload-version, %d opts", len(opts))

	defs := muc.buildModuleUploadVersionDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(muc.meta.BinaryName+" module upload-version", defs, opts)
	if err != nil {
		muc.meta.Logger.Error(output.FormatError("failed to parse module upload-version options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		muc.meta.Logger.Error(output.FormatError("missing module upload-version module path", nil), muc.HelpModuleUploadVersion())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive module upload-version arguments: %s", cmdArgs)
		muc.meta.Logger.Error(output.FormatError(msg, nil), muc.HelpModuleUploadVersion())
		return 1
	}

	modulePath := cmdArgs[0]
	directoryPath := getOption("directory-path", "", cmdOpts)[0]
	version := getOption("version", "", cmdOpts)[0]

	// Error is already logged.
	if !isResourcePathValid(muc.meta, modulePath) {
		return 1
	}

	// Make sure the directory path exists and is a directory--to give more precise messages.
	if err = muc.checkDirPath(directoryPath); err != nil {
		muc.meta.Logger.Error(output.FormatError("invalid directory path", err))
		return 1
	}

	log.Info("starting module upload-version...")

	log.WithField("directory-path", directoryPath).Info("creating module package...")

	slugFile, err := os.CreateTemp("", "terraform-slug.tgz")
	if err != nil {
		muc.meta.UI.Error(output.FormatError("failed to create module version", err))
		return 1
	}
	defer os.Remove(slugFile.Name())

	slug, err := slug.NewSlug(directoryPath, slugFile.Name())
	if err != nil {
		muc.meta.UI.Error(output.FormatError("failed to create module package", err))
		return 1
	}

	log.Info("module package successfully created")

	reader, err := slug.Open()
	if err != nil {
		muc.meta.UI.Error(output.FormatError("failed to create module version", err))
		return 1
	}
	defer reader.Close()

	log.WithField("module", modulePath).WithField("version", version).Info("creating module version...")

	// Create module version
	moduleVersion, err := client.TerraformModuleVersion.CreateModuleVersion(ctx, &sdktypes.CreateTerraformModuleVersionInput{
		ModulePath: modulePath,
		Version:    version,
		SHASum:     hex.EncodeToString(slug.SHASum),
	})
	if err != nil {
		muc.meta.UI.Error(output.FormatError("failed to create module version", err))
		return 1
	}

	log.Info("module version successfully created")
	log.Info("starting upload...")

	uploadStartTime := time.Now()

	if err = client.TerraformModuleVersion.UploadModuleVersion(ctx, moduleVersion.Metadata.ID, reader); err != nil {
		muc.meta.UI.Error(output.FormatError("failed to upload module version", err))

		// Delete module version
		if err = client.TerraformModuleVersion.DeleteModuleVersion(ctx, &sdktypes.DeleteTerraformModuleVersionInput{
			ID: moduleVersion.Metadata.ID,
		}); err != nil {
			muc.meta.UI.Error(output.FormatError("failed to delete module version", err))
		}
		return 1
	}

	log.IncreasePadding()

	// Wait for module version upload to complete
	// Wait for the upload to complete:
	var updatedModuleVersion *sdktypes.TerraformModuleVersion
	for {

		log.Info(fmt.Sprintf("upload in progress [%s elapsed]", time.Since(uploadStartTime)))

		updatedModuleVersion, err = client.TerraformModuleVersion.GetModuleVersion(ctx, &sdktypes.GetTerraformModuleVersionInput{
			ID: &moduleVersion.Metadata.ID,
		})
		if err != nil {
			muc.meta.UI.Error(output.FormatError("failed to check module version upload status", err))
			return 1
		}
		if updatedModuleVersion.Status == "uploaded" || updatedModuleVersion.Status == "errored" {
			break
		}
		time.Sleep(2 * time.Second)
	}

	log.ResetPadding()

	if updatedModuleVersion.Status == "errored" {
		log.WithField("error", updatedModuleVersion.Error).Error("module version upload failed")
		muc.meta.UI.Output(updatedModuleVersion.Diagnostics)
		// Delete module version
		if err = client.TerraformModuleVersion.DeleteModuleVersion(ctx, &sdktypes.DeleteTerraformModuleVersionInput{
			ID: moduleVersion.Metadata.ID,
		}); err != nil {
			muc.meta.UI.Error(output.FormatError("failed to delete module version", err))
		}
		return 1
	}

	log.Info("module version upload complete")

	return 0
}

func (muc moduleUploadVersionCommand) checkDirPath(directoryPath string) error {
	dirStat, err := os.Stat(directoryPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory path does not exist: %s", directoryPath)
	}
	if err != nil {
		return fmt.Errorf("failed to stat directory path %s: %s", directoryPath, err)
	}
	if !dirStat.IsDir() {
		return fmt.Errorf("path is not a directory: %s", directoryPath)
	}
	return nil
}

// buildModuleUploadVersionDefs returns defs used by module upload-version command.
func (muc moduleUploadVersionCommand) buildModuleUploadVersionDefs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"directory-path": {
			Arguments: []string{"Directory_Path"},
			Synopsis:  "The path of the terraform module's directory.",
		},
		"version": {
			Arguments: []string{"Version"},
			Synopsis:  "The semantic version for the new module version.",
			Required:  true,
		},
	}
}

func (muc moduleUploadVersionCommand) Synopsis() string {
	return "Upload a new module version to the module registry."
}

func (muc moduleUploadVersionCommand) Help() string {
	return muc.HelpModuleUploadVersion()
}

// HelpModuleUploadVersion produces the help string for the 'module upload-version' command.
func (muc moduleUploadVersionCommand) HelpModuleUploadVersion() string {
	return fmt.Sprintf(`
Usage: %s [global options] module upload-version [options] <module-path>

   The module upload-version command uploads a new
   module version to the module registry.

%s

`, muc.meta.BinaryName, buildHelpText(muc.buildModuleUploadVersionDefs()))
}

// The End.
