package command

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/tableformatter"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

// managedIdentityData holds decoded (or not-yet-encoded) type-specific data for a managed identity.
type managedIdentityData struct {
	awsRole                   *string
	azureClientID             *string
	azureTenantID             *string
	tharsisServiceAccountPath *string
}

// managedIdentityCreateCommand is the top-level structure for the managed-identity create command.
type managedIdentityCreateCommand struct {
	meta *Metadata
}

// NewManagedIdentityCreateCommandFactory returns a managedIdentityCreateCommand struct.
func NewManagedIdentityCreateCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return managedIdentityCreateCommand{
			meta: meta,
		}, nil
	}
}

func (m managedIdentityCreateCommand) Run(args []string) int {
	m.meta.Logger.Debugf("Starting the 'managed-identity create' command with %d arguments:", len(args))
	for ix, arg := range args {
		m.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := m.meta.GetSDKClient()
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return m.doManagedIdentityCreate(ctx, client, args)
}

func (m managedIdentityCreateCommand) doManagedIdentityCreate(ctx context.Context,
	client *tharsis.Client, opts []string,
) int {
	m.meta.Logger.Debugf("will do managed-identity create, %d opts", len(opts))

	defs := buildManagedIdentityCreateDefs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(m.meta.BinaryName+" managed-identity create", defs, opts)
	if err != nil {
		m.meta.Logger.Error(output.FormatError("failed to parse managed-identity create options", err))
		return 1
	}
	if len(cmdArgs) > 0 {
		msg := fmt.Sprintf("excessive managed-identity create arguments: %s", cmdArgs)
		m.meta.Logger.Error(output.FormatError(msg, nil), m.HelpManagedIdentityCreate())
		return 1
	}

	toJSON, err := getBoolOptionValue("json", "false", cmdOpts)
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}
	description := getOption("description", "", cmdOpts)[0]
	groupPath := getOption("group-path", "", cmdOpts)[0]
	managedIdentityType := sdktypes.ManagedIdentityType(getOption("type", "", cmdOpts)[0])
	name := getOption("name", "", cmdOpts)[0]

	// Process the type-specific options and encode the data string.
	data, err := encodeManagedIdentityData(managedIdentityType, cmdOpts)
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to encode data", err))
		return 1
	}

	input := &sdktypes.CreateManagedIdentityInput{
		Description: description,
		GroupPath:   groupPath,
		Type:        managedIdentityType,
		Name:        name,
		Data:        data,
	}
	m.meta.Logger.Debugf("managed-identity create input: %#v", input)

	managedIdentity, err := client.ManagedIdentity.CreateManagedIdentity(ctx, input)
	if err != nil {
		m.meta.UI.Error(output.FormatError("failed to create managed identity", err))
		return 1
	}

	if err = outputManagedIdentity(m.meta, toJSON, managedIdentity); err != nil {
		m.meta.UI.Error(err.Error())
		return 1
	}

	return 0
}

// encodeManagedIdentityData encodes the data for a managed identity.
func encodeManagedIdentityData(managedIdentityType sdktypes.ManagedIdentityType,
	options map[string][]string,
) (string, error) {
	var toEncode interface{}

	switch managedIdentityType {
	case sdktypes.ManagedIdentityAWSFederated:
		if _, ok := options["aws-federated-role"]; !ok {
			return "", fmt.Errorf("missing required option: aws-federated-role")
		}
		toEncode = struct {
			Role string `json:"role"`
		}{
			Role: getOption("aws-federated-role", "", options)[0],
		}
	case sdktypes.ManagedIdentityAzureFederated:
		if _, ok := options["azure-federated-client-id"]; !ok {
			return "", fmt.Errorf("missing required option: azure-federated-client-id")
		}
		if _, ok := options["azure-federated-tenant-id"]; !ok {
			return "", fmt.Errorf("missing required option: azure-federated-tenant-id")
		}
		toEncode = struct {
			ClientID string `json:"clientId"`
			TenantID string `json:"tenantId"`
		}{
			ClientID: getOption("azure-federated-client-id", "", options)[0],
			TenantID: getOption("azure-federated-tenant-id", "", options)[0],
		}
	case sdktypes.ManagedIdentityTharsisFederated:
		// The service account does not need to be verified.
		// It is possible to create a managed identity that refers to a non-existent service account.
		if _, ok := options["tharsis-federated-service-account-path"]; !ok {
			return "", fmt.Errorf("missing required option: tharsis-federated-service-account-path")
		}
		toEncode = struct {
			ServiceAccountPath string `json:"serviceAccountPath"`
		}{
			ServiceAccountPath: getOption("tharsis-federated-service-account-path", "", options)[0],
		}
	}
	data, err := json.Marshal(toEncode)
	if err != nil {
		return "", fmt.Errorf("failed to encode data: %s", err)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// decodeManagedIdentityData decodes the data for a managed identity.
func decodeManagedIdentityData(managedIdentityType sdktypes.ManagedIdentityType, data string) (*managedIdentityData, error) {
	result := &managedIdentityData{}

	buffer, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode managed identity data: %s", err)
	}

	switch managedIdentityType {
	case sdktypes.ManagedIdentityAWSFederated:
		var target struct {
			Role string `json:"role"`
		}
		if err := json.Unmarshal(buffer, &target); err != nil {
			return nil, fmt.Errorf("failed to unmarshal AWS federated data: %s", err)
		}
		result.awsRole = &target.Role
	case sdktypes.ManagedIdentityAzureFederated:
		var target struct {
			ClientID string `json:"clientId"`
			TenantID string `json:"tenantId"`
		}
		if err := json.Unmarshal(buffer, &target); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Azure federated data: %s", err)
		}
		result.azureClientID = &target.ClientID
		result.azureTenantID = &target.TenantID
	case sdktypes.ManagedIdentityTharsisFederated:
		var target struct {
			ServiceAccountPath string `json:"serviceAccountPath"`
		}
		if err := json.Unmarshal(buffer, &target); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Tharsis federated data: %s", err)
		}
		result.tharsisServiceAccountPath = &target.ServiceAccountPath
	}

	return result, nil
}

// outputManagedIdentity is the final output for most managed-identity operations.
// This output function will be called by other commands, so it must return the raw error.
func outputManagedIdentity(meta *Metadata, toJSON bool, managedIdentity *sdktypes.ManagedIdentity) error {
	if toJSON {
		buf, err := objectToJSON(managedIdentity)
		if err != nil {
			return fmt.Errorf("failed to encode managed identity to JSON: %s", err)
		}
		meta.UI.Output(string(buf))
	} else {
		tableHead := []string{
			"id",
			"name",
			"resource path",
			"type",
			"description",
			"is alias",
		}
		tableBody := []string{
			managedIdentity.Metadata.ID,
			managedIdentity.Name,
			managedIdentity.ResourcePath,
			string(managedIdentity.Type),
			managedIdentity.Description,
			fmt.Sprintf("%t", managedIdentity.IsAlias),
		}

		// If alias, add alias source ID.
		if managedIdentity.IsAlias {
			tableHead = append(tableHead, "alias source")
			tableBody = append(tableBody, *managedIdentity.AliasSourceID)
		}

		// Add the type-specific data.
		decoded, err := decodeManagedIdentityData(managedIdentity.Type, managedIdentity.Data)
		if err != nil {
			return err
		}

		switch managedIdentity.Type {
		case sdktypes.ManagedIdentityAWSFederated:
			tableHead = append(tableHead, "role")
			tableBody = append(tableBody, *decoded.awsRole)
		case sdktypes.ManagedIdentityAzureFederated:
			tableHead = append(tableHead, "client ID", "tenant ID")
			tableBody = append(tableBody, *decoded.azureClientID, *decoded.azureTenantID)
		case sdktypes.ManagedIdentityTharsisFederated:
			tableHead = append(tableHead, "service account path")
			tableBody = append(tableBody, *decoded.tharsisServiceAccountPath)
		}

		// For now, do not display access rules or a list of aliases.

		meta.UI.Output(tableformatter.FormatTable([][]string{tableHead, tableBody}))
	}

	return nil
}

// buildManagedIdentityCreateDefs returns defs used by managed-identity create command.
func buildManagedIdentityCreateDefs() optparser.OptionDefinitions {
	defs := optparser.OptionDefinitions{
		"description": {
			Arguments: []string{"Description"},
			Synopsis:  "Description for the managed identity.",
		},
		"group-path": {
			Arguments: []string{"Group_Path"},
			Synopsis:  "Full path of group where the managed identity will be created.",
			Required:  true,
		},
		"type": {
			Arguments: []string{""},
			Synopsis:  "The type of managed identity: aws_federated, azure_federated, tharsis_federated.",
			Required:  true,
		},
		"name": {
			Arguments: []string{"Managed_Identity_Name"},
			Synopsis:  "The name of the managed identity.",
			Required:  true,
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

func (m managedIdentityCreateCommand) Synopsis() string {
	return "Create a new managed identity."
}

func (m managedIdentityCreateCommand) Help() string {
	return m.HelpManagedIdentityCreate()
}

// HelpManagedIdentityCreate produces the help string for the 'managed-identity create' command.
func (m managedIdentityCreateCommand) HelpManagedIdentityCreate() string {
	return fmt.Sprintf(`
Usage: %s [global options] managed-identity create [options]

   The managed-identity create command creates a new managed identity.

%s

`, m.meta.BinaryName, buildHelpText(buildManagedIdentityCreateDefs()))
}
