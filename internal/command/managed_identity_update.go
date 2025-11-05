package command

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// managedIdentityUpdateCommand is the top-level structure for the managed-identity update command.
type managedIdentityUpdateCommand struct {
	meta *Metadata
}

// NewManagedIdentityUpdateCommandFactory returns a managedIdentityUpdateCommand struct.
func NewManagedIdentityUpdateCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return managedIdentityUpdateCommand{
			meta: meta,
		}, nil
	}
}

func (m managedIdentityUpdateCommand) Run(args []string) int {
	m.meta.Logger.Debugf("Starting the 'managed-identity update' command with %d arguments:", len(args))
	for ix, arg := range args {
		m.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := m.meta.GetSDKClient()
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return m.doManagedIdentityUpdate(ctx, client, args)
}

func (m managedIdentityUpdateCommand) doManagedIdentityUpdate(ctx context.Context, client *tharsis.Client, opts []string) int {
	m.meta.Logger.Debugf("will do managed-identity update, %d opts", len(opts))

	defs := buildManagedIdentityUpdateDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(m.meta.BinaryName+" managed-identity update", defs, opts)
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to parse managed-identity update options", err))
		return 1
	}
	if len(cmdArgs) < 1 {
		m.meta.Logger.Error(output.FormatError("missing managed-identity update path", nil), m.HelpManagedIdentityUpdate())
		return 1
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive managed-identity update arguments: %s", cmdArgs)
		m.meta.Logger.Error(output.FormatError(msg, nil), m.HelpManagedIdentityUpdate())
		return 1
	}

	managedIdentityPath := cmdArgs[0]
	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	actualPath := trn.ToPath(managedIdentityPath)
	if !isResourcePathValid(m.meta, actualPath) {
		return 1
	}

	// Get the managed identity so, we can find it's ID.
	trnID := trn.ToTRN(managedIdentityPath, trn.ResourceTypeManagedIdentity)
	getManagedIdentityInput := &sdktypes.GetManagedIdentityInput{ID: &trnID}
	managedIdentity, err := client.ManagedIdentity.GetManagedIdentity(ctx, getManagedIdentityInput)
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to get managed identity", err))
		return 1
	}

	// Decode the pre-existing object's data so the fields can be copied in case they are not changing.
	managedIdentityData, err := decodeManagedIdentityData(managedIdentity.Type, managedIdentity.Data)
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to decode managed identity data", err))
		return 1
	}

	// Process the type-specific command options and encode the data--with defaults from the pre-existing object.
	data, err := encodeManagedIdentityDataWithDefaults(managedIdentity.Type, cmdOpts, managedIdentityData)
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to encode data", err))
		return 1
	}

	// Prepare the inputs.
	input := &sdktypes.UpdateManagedIdentityInput{
		ID:          managedIdentity.Metadata.ID,
		Description: getOption("description", managedIdentity.Description, cmdOpts)[0],
		Data:        data,
	}

	m.meta.Logger.Debugf("managed-identity update input: %#v", input)

	// Update the managed identity.
	updatedManagedIdentity, err := client.ManagedIdentity.UpdateManagedIdentity(ctx, input)
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to update managed identity", err))
		return 1
	}

	if err = outputManagedIdentity(m.meta, toJSON, updatedManagedIdentity); err != nil {
		m.meta.UI.Error(err.Error())
		return 1
	}

	return 0
}

// encodeManagedIdentityDataWithDefaults encodes the data for a managed identity.
func encodeManagedIdentityDataWithDefaults(managedIdentityType sdktypes.ManagedIdentityType,
	options map[string][]string, defaults *managedIdentityData,
) (string, error) {
	var toEncode interface{}

	switch managedIdentityType {
	case sdktypes.ManagedIdentityAWSFederated:
		toEncode = struct {
			Role string `json:"role"`
		}{
			Role: getOption("aws-federated-role", *defaults.awsRole, options)[0],
		}
	case sdktypes.ManagedIdentityAzureFederated:
		toEncode = struct {
			ClientID string `json:"clientId"`
			TenantID string `json:"tenantId"`
		}{
			ClientID: getOption("azure-federated-client-id", *defaults.azureClientID, options)[0],
			TenantID: getOption("azure-federated-tenant-id", *defaults.azureTenantID, options)[0],
		}
	case sdktypes.ManagedIdentityTharsisFederated:
		toEncode = struct {
			ServiceAccountPath string `json:"serviceAccountPath"`
		}{
			ServiceAccountPath: getOption("tharsis-federated-service-account-path",
				*defaults.tharsisServiceAccountPath, options)[0],
		}
	}
	data, err := json.Marshal(toEncode)
	if err != nil {
		return "", fmt.Errorf("failed to encode data: %s", err)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// buildManagedIdentityUpdateDefs returns defs used by managed-identity create command.
func buildManagedIdentityUpdateDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"description": {
			Arguments: []string{"Description"},
			Synopsis:  "Description for the managed identity.",
		},
		"group-path": {
			Arguments: []string{"Group_Path"},
			Synopsis:  "Full path of group where the managed identity will be created.",
		},
		"name": {
			Arguments: []string{"Managed_Identity_Name"},
			Synopsis:  "The name of the managed identity.",
		},
		// The following are the options for the various types of managed identities.
		// They are ignored if specified on a type that does not use them.
		// See the Tharsis API's InputData struct in internal/services/managedidentity/awsfederated/delegate.go
		"aws-federated-role": {
			Arguments: []string{"AWS_Federated_Role"},
			Synopsis:  fmt.Sprintf("AWS IAM role.  (Only if type is %s)", sdktypes.ManagedIdentityAWSFederated),
		},
		// See the API's InputData struct in internal/services/managedidentity/azurefederated/delegate.go
		"azure-federated-client-id": {
			Arguments: []string{"Azure_Federated_Client_ID"},
			Synopsis:  fmt.Sprintf("Azure client ID.  (Only if type is %s)", sdktypes.ManagedIdentityAzureFederated),
		},
		"azure-federated-tenant-id": {
			Arguments: []string{"Azure_Federated_Tenant_ID"},
			Synopsis:  fmt.Sprintf("Azure tenant ID.  (Only if type is %s)", sdktypes.ManagedIdentityAzureFederated),
		},
		// See the API's InputData struct in internal/services/managedidentity/tharsisfederated/delegate.go
		"tharsis-federated-service-account-path": {
			Arguments: []string{"Tharsis_Federated_Service_Account_Path"},
			Synopsis:  fmt.Sprintf("Tharsis service account path.  (Only if type is %s)", sdktypes.ManagedIdentityTharsisFederated),
		},
	}

	return buildJSONOptionDefs(defs)
}

func (m managedIdentityUpdateCommand) Synopsis() string {
	return "Update a managed identity."
}

func (m managedIdentityUpdateCommand) Help() string {
	return m.HelpManagedIdentityUpdate()
}

// HelpManagedIdentityUpdate produces the help string for the 'managed-identity update' command.
func (m managedIdentityUpdateCommand) HelpManagedIdentityUpdate() string {
	return fmt.Sprintf(`
Usage: %s [global options] managed-identity update [options] <managed-identity-path>

   The managed-identity update command updates a managed identity. Shows final
   output as JSON, if specified.

%s

`, m.meta.BinaryName, buildHelpText(buildManagedIdentityUpdateDefs()))
}
