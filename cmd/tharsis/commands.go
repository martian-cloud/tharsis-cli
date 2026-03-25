package main

import (
	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/command"
)

// commands returns all the available commands.
func commands(baseCommand *command.BaseCommand) (map[string]cli.CommandFactory, error) {
	// The map of all commands except documentation.
	commandMap := map[string]command.Factory{
		"apply":                                     command.NewApplyCommandFactory(baseCommand),
		"caller-identity":                           command.NewCallerIdentityCommandFactory(baseCommand),
		"configure":                                 command.NewConfigureCommandFactory(baseCommand),
		"configure delete":                          command.NewConfigureDeleteCommandFactory(baseCommand),
		"configure list":                            command.NewConfigureListCommandFactory(baseCommand),
		"destroy":                                   command.NewDestroyCommandFactory(baseCommand),
		"group":                                     command.NewHelpCommandFactory(getHelpText("group")),
		"group get":                                 command.NewGroupGetCommandFactory(baseCommand),
		"group create":                              command.NewGroupCreateCommandFactory(baseCommand),
		"group update":                              command.NewGroupUpdateCommandFactory(baseCommand),
		"group delete":                              command.NewGroupDeleteCommandFactory(baseCommand),
		"group list":                                command.NewGroupListCommandFactory(baseCommand),
		"group migrate":                             command.NewGroupMigrateCommandFactory(baseCommand),
		"group list-memberships":                    command.NewGroupListMembershipsCommandFactory(baseCommand),
		"group get-membership":                      command.NewGroupGetMembershipCommandFactory(baseCommand),
		"group add-membership":                      command.NewGroupAddMembershipCommandFactory(baseCommand),
		"group update-membership":                   command.NewGroupUpdateMembershipCommandFactory(baseCommand),
		"group remove-membership":                   command.NewGroupRemoveMembershipCommandFactory(baseCommand),
		"group get-terraform-var":                   command.NewGroupGetTerraformVarCommandFactory(baseCommand),
		"group set-terraform-var":                   command.NewGroupSetTerraformVarCommandFactory(baseCommand),
		"group delete-terraform-var":                command.NewGroupDeleteTerraformVarCommandFactory(baseCommand),
		"group list-terraform-vars":                 command.NewGroupListTerraformVarsCommandFactory(baseCommand),
		"group set-terraform-vars":                  command.NewGroupSetTerraformVarsCommandFactory(baseCommand),
		"group list-environment-vars":               command.NewGroupListEnvironmentVarsCommandFactory(baseCommand),
		"group set-environment-vars":                command.NewGroupSetEnvironmentVarsCommandFactory(baseCommand),
		"managed-identity":                          command.NewHelpCommandFactory(getHelpText("managed-identity")),
		"managed-identity get":                      command.NewManagedIdentityGetCommandFactory(baseCommand),
		"managed-identity create":                   command.NewManagedIdentityCreateCommandFactory(baseCommand),
		"managed-identity update":                   command.NewManagedIdentityUpdateCommandFactory(baseCommand),
		"managed-identity delete":                   command.NewManagedIdentityDeleteCommandFactory(baseCommand),
		"managed-identity-access-rule":              command.NewHelpCommandFactory(getHelpText("managed-identity-access-rule")),
		"managed-identity-access-rule get":          command.NewManagedIdentityAccessRuleGetCommandFactory(baseCommand),
		"managed-identity-access-rule list":         command.NewManagedIdentityAccessRuleListCommandFactory(baseCommand),
		"managed-identity-access-rule create":       command.NewManagedIdentityAccessRuleCreateCommandFactory(baseCommand),
		"managed-identity-access-rule update":       command.NewManagedIdentityAccessRuleUpdateCommandFactory(baseCommand),
		"managed-identity-access-rule delete":       command.NewManagedIdentityAccessRuleDeleteCommandFactory(baseCommand),
		"managed-identity-alias":                    command.NewHelpCommandFactory(getHelpText("managed-identity-alias")),
		"managed-identity-alias create":             command.NewManagedIdentityAliasCreateCommandFactory(baseCommand),
		"managed-identity-alias delete":             command.NewManagedIdentityAliasDeleteCommandFactory(baseCommand),
		"mcp":                                       command.NewMCPCommandFactory(baseCommand),
		"module":                                    command.NewHelpCommandFactory(getHelpText("module")),
		"module get":                                command.NewModuleGetCommandFactory(baseCommand),
		"module create":                             command.NewModuleCreateCommandFactory(baseCommand),
		"module update":                             command.NewModuleUpdateCommandFactory(baseCommand),
		"module delete":                             command.NewModuleDeleteCommandFactory(baseCommand),
		"module list":                               command.NewModuleListCommandFactory(baseCommand),
		"module list-versions":                      command.NewModuleListVersionsCommandFactory(baseCommand),
		"module list-attestations":                  command.NewModuleListAttestationsCommandFactory(baseCommand),
		"module create-attestation":                 command.NewModuleCreateAttestationCommandFactory(baseCommand),
		"module update-attestation":                 command.NewModuleUpdateAttestationCommandFactory(baseCommand),
		"module delete-attestation":                 command.NewModuleDeleteAttestationCommandFactory(baseCommand),
		"module get-version":                        command.NewModuleGetVersionCommandFactory(baseCommand),
		"module delete-version":                     command.NewModuleDeleteVersionCommandFactory(baseCommand),
		"module upload-version":                     command.NewModuleUploadVersionCommandFactory(baseCommand),
		"plan":                                      command.NewPlanCommandFactory(baseCommand),
		"run":                                       command.NewHelpCommandFactory(getHelpText("run")),
		"run cancel":                                command.NewRunCancelCommandFactory(baseCommand),
		"runner-agent":                              command.NewHelpCommandFactory(getHelpText("runner-agent")),
		"runner-agent get":                          command.NewRunnerAgentGetCommandFactory(baseCommand),
		"runner-agent create":                       command.NewRunnerAgentCreateCommandFactory(baseCommand),
		"runner-agent assign-service-account":       command.NewRunnerAgentAssignServiceAccountCommandFactory(baseCommand),
		"runner-agent unassign-service-account":     command.NewRunnerAgentUnassignServiceAccountCommandFactory(baseCommand),
		"runner-agent update":                       command.NewRunnerAgentUpdateCommandFactory(baseCommand),
		"runner-agent delete":                       command.NewRunnerAgentDeleteCommandFactory(baseCommand),
		"service-account":                           command.NewHelpCommandFactory(getHelpText("service-account")),
		"service-account create-token":              command.NewServiceAccountCreateTokenCommandFactory(baseCommand),
		"sso":                                       command.NewHelpCommandFactory(getHelpText("sso")),
		"sso login":                                 command.NewLoginCommandFactory(baseCommand),
		"terraform-provider":                        command.NewHelpCommandFactory(getHelpText("terraform-provider")),
		"terraform-provider create":                 command.NewTerraformProviderCreateCommandFactory(baseCommand),
		"terraform-provider upload-version":         command.NewTerraformProviderUploadVersionCommandFactory(baseCommand),
		"terraform-provider-mirror":                 command.NewHelpCommandFactory(getHelpText("terraform-provider-mirror")),
		"terraform-provider-mirror get-version":     command.NewTerraformProviderMirrorGetVersionCommandFactory(baseCommand),
		"terraform-provider-mirror list-versions":   command.NewTerraformProviderMirrorListVersionsCommandFactory(baseCommand),
		"terraform-provider-mirror list-platforms":  command.NewTerraformProviderMirrorListPlatformsCommandFactory(baseCommand),
		"terraform-provider-mirror sync":            command.NewTerraformProviderMirrorSyncCommandFactory(baseCommand),
		"terraform-provider-mirror delete-version":  command.NewTerraformProviderMirrorDeleteVersionCommandFactory(baseCommand),
		"terraform-provider-mirror delete-platform": command.NewTerraformProviderMirrorDeletePlatformCommandFactory(baseCommand),
		"version":                                   command.NewVersionCommandFactory(baseCommand),
		"workspace":                                 command.NewHelpCommandFactory(getHelpText("workspace")),
		"workspace get":                             command.NewWorkspaceGetCommandFactory(baseCommand),
		"workspace create":                          command.NewWorkspaceCreateCommandFactory(baseCommand),
		"workspace update":                          command.NewWorkspaceUpdateCommandFactory(baseCommand),
		"workspace delete":                          command.NewWorkspaceDeleteCommandFactory(baseCommand),
		"workspace list":                            command.NewWorkspaceListCommandFactory(baseCommand),
		"workspace migrate":                         command.NewWorkspaceMigrateCommandFactory(baseCommand),
		"workspace list-memberships":                command.NewWorkspaceListMembershipsCommandFactory(baseCommand),
		"workspace get-membership":                  command.NewWorkspaceGetMembershipCommandFactory(baseCommand),
		"workspace add-membership":                  command.NewWorkspaceAddMembershipCommandFactory(baseCommand),
		"workspace update-membership":               command.NewWorkspaceUpdateMembershipCommandFactory(baseCommand),
		"workspace remove-membership":               command.NewWorkspaceRemoveMembershipCommandFactory(baseCommand),
		"workspace assign-managed-identity":         command.NewWorkspaceAssignManagedIdentityCommandFactory(baseCommand),
		"workspace unassign-managed-identity":       command.NewWorkspaceUnassignManagedIdentityCommandFactory(baseCommand),
		"workspace get-assigned-managed-identities": command.NewWorkspaceGetAssignedManagedIdentitiesCommandFactory(baseCommand),
		"workspace outputs":                         command.NewWorkspaceOutputsCommandFactory(baseCommand),
		"workspace label":                           command.NewWorkspaceLabelCommandFactory(baseCommand),
		"workspace get-terraform-var":               command.NewWorkspaceGetTerraformVarCommandFactory(baseCommand),
		"workspace set-terraform-var":               command.NewWorkspaceSetTerraformVarCommandFactory(baseCommand),
		"workspace delete-terraform-var":            command.NewWorkspaceDeleteTerraformVarCommandFactory(baseCommand),
		"workspace list-terraform-vars":             command.NewWorkspaceListTerraformVarsCommandFactory(baseCommand),
		"workspace set-terraform-vars":              command.NewWorkspaceSetTerraformVarsCommandFactory(baseCommand),
		"workspace list-environment-vars":           command.NewWorkspaceListEnvironmentVarsCommandFactory(baseCommand),
		"workspace set-environment-vars":            command.NewWorkspaceSetEnvironmentVarsCommandFactory(baseCommand),
	}

	// 	// Add the documentation commands.
	commandMap["documentation"] = command.NewHelpCommandFactory(getHelpText("documentation"))
	commandMap["documentation generate"] = command.NewDocumentationGenerateCommandFactory(baseCommand, commandMap)

	// Convert CommandFactory to cli.CommandFactory.
	returnMap := map[string]cli.CommandFactory{}
	for name, helpCommandFactory := range commandMap {
		helpCommand, err := helpCommandFactory()
		if err != nil {
			return nil, err
		}

		returnMap[name] = func() (cli.Command, error) {
			return command.NewWrapper(helpCommand), nil
		}
	}

	return returnMap, nil
}
