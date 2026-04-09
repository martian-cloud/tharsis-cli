---
title: Commands
description: "An introduction to the CLI commands"
---
  
## Available Commands
Currently, the CLI supports the following commands:
  
- [apply](#apply-command) — Apply a Terraform run.
- [caller-identity](#caller-identity-command) — Get the caller's identity.
- [configure](#configure-command) — Create or update a profile.
- [destroy](#destroy-command) — Destroy workspace resources.
- [group](#group-command) — Do operations on groups.
- [managed-identity](#managed-identity-command) — Do operations on a managed identity.
- [managed-identity-access-rule](#managed-identity-access-rule-command) — Do operations on a managed identity access rule.
- [managed-identity-alias](#managed-identity-alias-command) — Do operations on a managed identity alias.
- [mcp](#mcp-command) — Start the Tharsis MCP server.
- [module](#module-command) — Do operations on a terraform module.
- [plan](#plan-command) — Create a speculative plan.
- [run](#run-command) — Do operations on runs.
- [runner-agent](#runner-agent-command) — Do operations on runner agents.
- [service-account](#service-account-command) — Create an authentication token for a service account.
- [sso](#sso-command) — Log in to the OAuth2 provider and return an authentication token.
- [terraform-provider](#terraform-provider-command) — Do operations on a terraform provider.
- [terraform-provider-mirror](#terraform-provider-mirror-command) — Mirror Terraform providers from any Terraform registry.
- [tf-exec](#tf-exec-command) — Run terraform with Tharsis auth and workspace variables injected.
- [version](#version-command) — Get the CLI's version.
- [workspace](#workspace-command) — Do operations on workspaces.
  
:::tip
`tharsis [command]` or `tharsis [command] -h` will output the help menu for that specific command.
:::
:::info
Commands and options may evolve between major versions. Options **must** come before any arguments.
:::
:::tip Have a question?
Check the [FAQ](#frequently-asked-questions-faq) to see if there's already an answer.
:::
:::info Legend
- <span style={{color:'red'}}>\*&nbsp;&nbsp;</span> required
- <span style={{color:'orange'}}>!&nbsp;&nbsp;</span> deprecated
- <span style={{color:'green'}}>...</span> repeatable
:::
  
---
## Global Options
  
#### disable-autocomplete

Uninstall shell autocompletion.

#### enable-autocomplete

Install shell autocompletion.

#### h, help

Show help output.

#### log

Set the verbosity of log output for debugging.\
**Values:** `debug`, `error`, `info`, `off`, `trace`, `warn`\
**Default:** `off`\
**Env:** `THARSIS_CLI_LOG`

#### no-color

Disable colored output.\
**Default:** `false`\
**Env:** `NO_COLOR`

#### p, profile

Profile to use from the configuration file.\
**Default:** `default`\
**Env:** `THARSIS_PROFILE`

#### v, version

Show the version information.


---
## apply command
**Apply a Terraform run.**
  
Creates and applies a Terraform run. First creates a
plan, then applies after approval. Supports run-scoped
Terraform and environment variables.

Terraform variables may be passed in via supported
options or from the environment with a 'TF_VAR_' prefix.
  
```bash
tharsis apply -directory-path "./terraform" trn:workspace:<workspace_path>
```
  
#### Options
  
#### auto-approve

Skip interactive approval of the plan.\
**Default:** `false`

#### directory-path

The path of the root module's directory.\
**Conflicts:** `module-source`

#### env-var <span style={{color:'green'}}>...</span>

An environment variable as a key=value pair.

#### env-var-file <span style={{color:'green'}}>...</span>

The path to an environment variables file.

#### input

Ask for input for variables if not directly set.\
**Default:** `true`

#### module-source

Remote module source specification.\
**Conflicts:** `directory-path`

#### module-version

Remote module version number. Uses latest if empty.

#### refresh

Whether to do the usual refresh step.\
**Default:** `true`

#### refresh-only

Whether to do ONLY a refresh operation.\
**Default:** `false`

#### target <span style={{color:'green'}}>...</span>

The Terraform address of the resources to be acted upon.

#### terraform-version

The Terraform CLI version to use for the run.

#### tf-var <span style={{color:'green'}}>...</span>

A terraform variable as a key=value pair.

#### tf-var-file <span style={{color:'green'}}>...</span>

The path to a .tfvars variables file.


---
## caller-identity command
**Get the caller's identity.**
  
Returns information about the authenticated caller
(User or ServiceAccount).
  
```bash
tharsis caller-identity
```
  
#### Options
  
#### json

Show final output as JSON.


---
## configure command
**Create or update a profile.**
  
**Subcommands:**
  
- [`delete`](#configure-delete-subcommand) - Remove a profile.
- [`list`](#configure-list-subcommand) - Show all profiles.
  
Creates or updates a profile. If no options are
specified, the command prompts for values.
  
```bash
tharsis configure \
  -http-endpoint "https://api.tharsis.example.com" \
  -profile "prod-example"
```
  
#### Options
  
#### endpoint-url <span style={{color:'orange'}}>!</span>

The Tharsis HTTP API endpoint.\
**Deprecated**: use -http-endpoint instead

#### http-endpoint

The Tharsis HTTP API endpoint.

#### insecure-tls-skip-verify

Allow TLS but disable verification of the gRPC server's certificate chain and hostname. Only use for testing as it could allow connecting to an impersonated server.\
**Default:** `false`

#### profile

The name of the profile to set.


---
### configure delete subcommand
**Remove a profile.**
  
Removes a profile and its stored credentials.
  
```bash
tharsis configure delete prod-example
```
  
---
### configure list subcommand
**Show all profiles.**
  
Displays all configured profiles and their endpoints.
  
```bash
tharsis configure list
```
  
---
## destroy command
**Destroy workspace resources.**
  
Destroys all resources in a workspace. Creates a
destroy plan, then applies after approval.

Terraform variables may be passed in via supported
options or from the environment with a 'TF_VAR_' prefix.
  
```bash
tharsis destroy -directory-path "./terraform" trn:workspace:<workspace_path>
```
  
#### Options
  
#### auto-approve

Skip interactive approval of the plan.\
**Default:** `false`

#### directory-path

The path of the root module's directory.\
**Conflicts:** `module-source`

#### env-var <span style={{color:'green'}}>...</span>

An environment variable as a key=value pair.

#### env-var-file <span style={{color:'green'}}>...</span>

The path to an environment variables file.

#### input

Ask for input for variables if not directly set.\
**Default:** `true`

#### module-source

Remote module source specification.\
**Conflicts:** `directory-path`

#### module-version

Remote module version number. Uses latest if empty.

#### refresh

Whether to do the usual refresh step.\
**Default:** `true`

#### target <span style={{color:'green'}}>...</span>

The Terraform address of the resources to be acted upon.

#### terraform-version

The Terraform CLI version to use for the run.

#### tf-var <span style={{color:'green'}}>...</span>

A terraform variable as a key=value pair.

#### tf-var-file <span style={{color:'green'}}>...</span>

The path to a .tfvars variables file.


---
## group command
**Do operations on groups.**
  
**Subcommands:**
  
- [`add-membership`](#group-add-membership-subcommand) - Add a membership to a group.
- [`create`](#group-create-subcommand) - Create a new group.
- [`delete`](#group-delete-subcommand) - Delete a group.
- [`delete-terraform-var`](#group-delete-terraform-var-subcommand) - Delete a terraform variable from a group.
- [`get`](#group-get-subcommand) - Get a single group.
- [`get-membership`](#group-get-membership-subcommand) - Get a group membership.
- [`get-terraform-var`](#group-get-terraform-var-subcommand) - Get a terraform variable for a group.
- [`list`](#group-list-subcommand) - Retrieve a paginated list of groups.
- [`list-environment-vars`](#group-list-environment-vars-subcommand) - List all environment variables in a group.
- [`list-memberships`](#group-list-memberships-subcommand) - Retrieve a list of group memberships.
- [`list-terraform-vars`](#group-list-terraform-vars-subcommand) - List all terraform variables in a group.
- [`migrate`](#group-migrate-subcommand) - Migrate a group to a new parent or to top-level.
- [`remove-membership`](#group-remove-membership-subcommand) - Remove a group membership.
- [`set-environment-vars`](#group-set-environment-vars-subcommand) - Set environment variables for a group.
- [`set-terraform-var`](#group-set-terraform-var-subcommand) - Set a terraform variable for a group.
- [`set-terraform-vars`](#group-set-terraform-vars-subcommand) - Set terraform variables for a group.
- [`update`](#group-update-subcommand) - Update a group.
- [`update-membership`](#group-update-membership-subcommand) - Update a group membership.
  
Groups are containers for organizing workspaces hierarchically.
They can be nested and inherit variables and managed identities
to children. Use group commands to create, update, delete groups,
set Terraform and environment variables, manage memberships, and
migrate groups between parents.
  
---
### group add-membership subcommand
**Add a membership to a group.**
  
Grants a user, service account, or team access to a
group. Exactly one identity flag must be specified.
  
```bash
tharsis group add-membership \
  -role-id "trn:role:<role_name>" \
  -user-id "trn:user:<username>" \
  trn:group:<group_path>
```
  
#### Options
  
#### json

Show final output as JSON.

#### role <span style={{color:'orange'}}>!</span>

The role for the membership.\
**Deprecated**: use -role-id

#### role-id

The role ID for the membership.

#### service-account-id

The service account ID for the membership.\
**Conflicts:** `user-id`, `team-id`, `username`, `team-name`

#### team-id

The team ID for the membership.\
**Conflicts:** `user-id`, `service-account-id`, `username`, `team-name`

#### team-name <span style={{color:'orange'}}>!</span>

The team name for the membership.\
**Deprecated**: use -team-id\
**Conflicts:** `user-id`, `service-account-id`, `team-id`, `username`

#### user-id

The user ID for the membership.\
**Conflicts:** `service-account-id`, `team-id`, `username`, `team-name`

#### username <span style={{color:'orange'}}>!</span>

The username for the membership.\
**Deprecated**: use -user-id\
**Conflicts:** `user-id`, `service-account-id`, `team-id`, `team-name`


---
### group create subcommand
**Create a new group.**
  
Creates a new group under a parent group with an
optional description.
  
```bash
tharsis group create \
  -parent-group-id "trn:group:<group_path>" \
  -description "Operations group" \
  <name>
```
  
#### Options
  
#### description

Description for the new group.

#### if-not-exists

Create a group if it does not already exist.\
**Default:** `false`

#### json

Show final output as JSON.

#### parent-group-id

Parent group ID.


---
### group delete subcommand
**Delete a group.**
  
Permanently removes a group. Use -force to delete
even if resources are deployed.
  
```bash
tharsis group delete \
  -force \
  trn:group:<group_path>
```
  
#### Options
  
#### force, f

Force delete the group.

#### version

Optimistic locking version. Usually not required.


---
### group delete-terraform-var subcommand
**Delete a terraform variable from a group.**
  
Removes a Terraform variable from a group.
  
```bash
tharsis group delete-terraform-var \
  -key "region" \
  trn:group:<group_path>
```
  
#### Options
  
#### key <span style={{color:'red'}}>*</span>

Variable key.

#### version

Optimistic locking version. Usually not required.


---
### group get subcommand
**Get a single group.**
  
Retrieves details about a group by ID or path.
  
```bash
tharsis group get \
  -json \
  trn:tharsis:group:<group_path>
```
  
#### Options
  
#### json

Show final output as JSON.


---
### group get-membership subcommand
**Get a group membership.**
  
Retrieves details about a specific group membership.
  
```bash
tharsis group get-membership \
  -user-id "trn:user:<username>" \
  trn:group:<group_path>
```
  
#### Options
  
#### json

Show final output as JSON.

#### service-account-id

Service account ID to find the group membership for.\
**Conflicts:** `user-id`, `team-id`, `username`, `team-name`

#### team-id

Team ID to find the group membership for.\
**Conflicts:** `user-id`, `service-account-id`, `username`, `team-name`

#### team-name <span style={{color:'orange'}}>!</span>

Team name to find the group membership for.\
**Deprecated**: use -team-id\
**Conflicts:** `user-id`, `service-account-id`, `team-id`, `username`

#### user-id

User ID to find the group membership for.\
**Conflicts:** `service-account-id`, `team-id`, `username`, `team-name`

#### username <span style={{color:'orange'}}>!</span>

Username to find the group membership for.\
**Deprecated**: use -user-id\
**Conflicts:** `user-id`, `service-account-id`, `team-id`, `team-name`


---
### group get-terraform-var subcommand
**Get a terraform variable for a group.**
  
Retrieves a Terraform variable from a group.
  
```bash
tharsis group get-terraform-var \
  -key "region" \
  trn:group:<group_path>
```
  
#### Options
  
#### json

Show final output as JSON.

#### key <span style={{color:'red'}}>*</span>

Variable key.

#### show-sensitive

Show the actual value of sensitive variables (requires appropriate permissions).\
**Default:** `false`


---
### group list subcommand
**Retrieve a paginated list of groups.**
  
Lists groups with pagination, filtering, and sorting.
  
```bash
tharsis group list \
  -parent-id "trn:group:<parent_group_path>" \
  -sort-by "FULL_PATH_ASC" \
  -limit 5 \
  -json
```
  
#### Options
  
#### cursor

The cursor string for manual pagination.

#### json

Show final output as JSON.

#### limit

Maximum number of result elements to return.\
**Default:** `100`

#### parent-id

Filter to only direct sub-groups of this parent group.

#### parent-path <span style={{color:'orange'}}>!</span>

Filter to only direct sub-groups of this parent group.\
**Deprecated**: use -parent-id

#### search

Filter to only groups containing this substring in their path.

#### sort-by

Sort by this field.\
**Values:** `FULL_PATH_ASC`, `FULL_PATH_DESC`, `GROUP_LEVEL_ASC`, `GROUP_LEVEL_DESC`, `UPDATED_AT_ASC`, `UPDATED_AT_DESC`\
**Conflicts:** `sort-order`

#### sort-order <span style={{color:'orange'}}>!</span>

Sort in this direction.\
**Values:** `ASC`, `DESC`\
**Deprecated**: use -sort-by\
**Conflicts:** `sort-by`


---
### group list-environment-vars subcommand
**List all environment variables in a group.**
  
Lists all environment variables from a group and its
parent groups.
  
```bash
tharsis group list-environment-vars -show-sensitive trn:group:<group_path>
```
  
#### Options
  
#### json

Show final output as JSON.

#### show-sensitive

Show the actual values of sensitive variables (requires appropriate permissions).\
**Default:** `false`


---
### group list-memberships subcommand
**Retrieve a list of group memberships.**
  
Lists all memberships for a group.
  
```bash
tharsis group list-memberships trn:group:<group_path>
```
  
#### Options
  
#### json

Show final output as JSON.


---
### group list-terraform-vars subcommand
**List all terraform variables in a group.**
  
Lists all Terraform variables from a group and its
parent groups.
  
```bash
tharsis group list-terraform-vars -show-sensitive trn:group:<group_path>
```
  
#### Options
  
#### json

Show final output as JSON.

#### show-sensitive

Show the actual values of sensitive variables (requires appropriate permissions).\
**Default:** `false`


---
### group migrate subcommand
**Migrate a group to a new parent or to top-level.**
  
Moves a group to a different parent or to top-level.
  
```bash
tharsis group migrate \
  -new-parent-id "trn:group:<parent_group_path>" \
  trn:group:<group_path>
```
  
#### Options
  
#### json

Show final output as JSON.

#### new-parent-id

New parent group ID. Omit to migrate to top-level.

#### new-parent-path <span style={{color:'orange'}}>!</span>

New parent path for the group.\
**Deprecated**: use -new-parent-id

#### to-top-level <span style={{color:'orange'}}>!</span>

Migrate group to top level.\
**Deprecated**: omit -new-parent-id instead


---
### group remove-membership subcommand
**Remove a group membership.**
  
Revokes a membership from a group.
  
```bash
tharsis group remove-membership <id>
```
  
#### Options
  
#### version

Optimistic locking version. Usually not required.


---
### group set-environment-vars subcommand
**Set environment variables for a group.**
  
Replaces all environment variables in a group from a
file. Does not support sensitive variables.
  
```bash
tharsis group set-environment-vars \
  -env-var-file "vars.env" \
  trn:group:<group_path>
```
  
#### Options
  
#### env-var-file <span style={{color:'red'}}>*</span> <span style={{color:'green'}}>...</span>

Path to an environment variables file.


---
### group set-terraform-var subcommand
**Set a terraform variable for a group.**
  
Creates or updates a Terraform variable for a group.
  
```bash
tharsis group set-terraform-var \
  -key "region" \
  -value "us-east-1" \
  trn:group:<group_path>
```
  
#### Options
  
#### key <span style={{color:'red'}}>*</span>

Variable key.

#### sensitive

Mark variable as sensitive.\
**Default:** `false`

#### value <span style={{color:'red'}}>*</span>

Variable value.


---
### group set-terraform-vars subcommand
**Set terraform variables for a group.**
  
Replaces all Terraform variables in a group from a
tfvars file. Does not support sensitive variables.
  
```bash
tharsis group set-terraform-vars \
  -tf-var-file "terraform.tfvars" \
  trn:group:<group_path>
```
  
#### Options
  
#### tf-var-file <span style={{color:'red'}}>*</span> <span style={{color:'green'}}>...</span>

Path to a .tfvars file.


---
### group update subcommand
**Update a group.**
  
Modifies a group's description.
  
```bash
tharsis group update \
  -description "Updated operations group" \
  trn:group:<group_path>
```
  
#### Options
  
#### description

Description for the group.

#### json

Show final output as JSON.

#### version

Optimistic locking version. Usually not required.


---
### group update-membership subcommand
**Update a group membership.**
  
Changes the role of an existing group membership.
  
```bash
tharsis group update-membership \
  -role-id "trn:role:<role_name>" \
  <id>
```
  
#### Options
  
#### json

Show final output as JSON.

#### role <span style={{color:'orange'}}>!</span>

New role for the membership.\
**Deprecated**: use -role-id

#### role-id <span style={{color:'red'}}>*</span>

The role ID for the membership.

#### version

Optimistic locking version. Usually not required.


---
## managed-identity command
**Do operations on a managed identity.**
  
**Subcommands:**
  
- [`create`](#managed-identity-create-subcommand) - Create a new managed identity.
- [`delete`](#managed-identity-delete-subcommand) - Delete a managed identity.
- [`get`](#managed-identity-get-subcommand) - Get a single managed identity.
- [`update`](#managed-identity-update-subcommand) - Update a managed identity.
  
Managed identities provide OIDC-federated credentials for cloud
providers (AWS, Azure, Kubernetes) without storing secrets. Use
managed-identity commands to create, update, delete, and get
managed identities.
  
---
### managed-identity create subcommand
**Create a new managed identity.**
  
Creates a new managed identity for OIDC-federated
cloud provider authentication.
  
```bash
tharsis managed-identity create \
  -group-id "trn:group:<group_path>" \
  -type "aws_federated" \
  -aws-federated-role "arn:aws:iam::123456789012:role/MyRole" \
  -description "AWS production role" \
  aws-prod
```
  
#### Options
  
#### aws-federated-role

AWS IAM role. (Only if type is aws_federated)

#### azure-federated-client-id

Azure client ID. (Only if type is azure_federated)

#### azure-federated-tenant-id

Azure tenant ID. (Only if type is azure_federated)

#### description

Description for the managed identity.

#### group-id

Group ID or TRN where the managed identity will be created.

#### group-path <span style={{color:'orange'}}>!</span>

The group path where the managed identity will be created.\
**Deprecated**: use -group-id

#### json

Show final output as JSON.

#### kubernetes-federated-audience

Kubernetes federated audience. The audience should match the client_id configured in your EKS OIDC identity provider. (Only if type is kubernetes_federated)

#### name <span style={{color:'orange'}}>!</span>

The name of the managed identity.\
**Deprecated**: pass name as an argument

#### tharsis-federated-service-account-path

Tharsis service account path this managed identity will assume. (Only if type is tharsis_federated)

#### type

The type of managed identity.\
**Values:** `aws_federated`, `azure_federated`, `kubernetes_federated`, `tharsis_federated`


---
### managed-identity delete subcommand
**Delete a managed identity.**
  
Permanently removes a managed identity. This action
is irreversible.
  
```bash
tharsis managed-identity delete -force trn:managed_identity:<group_path>/<managed_identity_name>
```
  
#### Options
  
#### force, f

Force delete the managed identity.


---
### managed-identity get subcommand
**Get a single managed identity.**
  
Retrieves details about a managed identity.
  
```bash
tharsis managed-identity get trn:managed_identity:<group_path>/<managed_identity_name>
```
  
#### Options
  
#### json

Show final output as JSON.


---
### managed-identity update subcommand
**Update a managed identity.**
  
Modifies a managed identity's description or data.
  
```bash
tharsis managed-identity update \
  -description "Updated AWS production role" \
  -aws-federated-role "arn:aws:iam::123456789012:role/UpdatedRole" \
  trn:managed_identity:<group_path>/<managed_identity_name>
```
  
#### Options
  
#### aws-federated-role

AWS IAM role. (Only if type is aws_federated)\
**Conflicts:** `azure-federated-client-id`, `tharsis-federated-service-account-path`, `kubernetes-federated-audience`

#### azure-federated-client-id

Azure client ID. (Only if type is azure_federated)\
**Conflicts:** `aws-federated-role`, `tharsis-federated-service-account-path`, `kubernetes-federated-audience`

#### azure-federated-tenant-id

Azure tenant ID. (Only if type is azure_federated)

#### description

Description for the managed identity.

#### json

Show final output as JSON.

#### kubernetes-federated-audience

Kubernetes federated audience. The audience should match the client_id configured in your EKS OIDC identity provider. (Only if type is kubernetes_federated)\
**Conflicts:** `aws-federated-role`, `azure-federated-client-id`, `tharsis-federated-service-account-path`

#### tharsis-federated-service-account-path

Tharsis service account path this managed identity will assume. (Only if type is tharsis_federated)\
**Conflicts:** `aws-federated-role`, `azure-federated-client-id`, `kubernetes-federated-audience`


---
## managed-identity-access-rule command
**Do operations on a managed identity access rule.**
  
**Subcommands:**
  
- [`create`](#managed-identity-access-rule-create-subcommand) - Create a new managed identity access rule.
- [`delete`](#managed-identity-access-rule-delete-subcommand) - Delete a managed identity access rule.
- [`get`](#managed-identity-access-rule-get-subcommand) - Get a managed identity access rule.
- [`list`](#managed-identity-access-rule-list-subcommand) - Retrieve a list of managed identity access rules.
- [`update`](#managed-identity-access-rule-update-subcommand) - Update a managed identity access rule.
  
Access rules control which runs can use a managed identity based
on conditions like module source or workspace path. Use these
commands to create, update, delete, list, and get access rules.
  
---
### managed-identity-access-rule create subcommand
**Create a new managed identity access rule.**
  
Creates an access rule that controls which workspaces
can use a managed identity.
  
```bash
tharsis managed-identity-access-rule create \
  -managed-identity-id "trn:managed_identity:<group_path>/<managed_identity_name>" \
  -rule-type "eligible_principals" \
  -run-stage "plan" \
  -allowed-user "trn:user:<username>" \
  -allowed-team "trn:team:<team_name>"
```
  
#### Options
  
#### allowed-service-account <span style={{color:'green'}}>...</span>

Allowed service account ID.

#### allowed-team <span style={{color:'green'}}>...</span>

Allowed team ID.

#### allowed-user <span style={{color:'green'}}>...</span>

Allowed user ID.

#### json

Show final output as JSON.

#### managed-identity-id

The ID or TRN of the managed identity.

#### managed-identity-path <span style={{color:'orange'}}>!</span>

Resource path to the managed identity.\
**Deprecated**: use -managed-identity-id

#### module-attestation-policy <span style={{color:'green'}}>...</span>

Module attestation policy in format "[PredicateType=someval,]PublicKeyFile=/path/to/file".

#### rule-type <span style={{color:'red'}}>*</span>

The type of access rule.\
**Values:** `eligible_principals`, `module_attestation`

#### run-stage <span style={{color:'red'}}>*</span>

The run stage.\
**Values:** `apply`, `plan`

#### verify-state-lineage

Verify state lineage.\
**Default:** `false`


---
### managed-identity-access-rule delete subcommand
**Delete a managed identity access rule.**
  
Removes an access rule from a managed identity.
  
```bash
tharsis managed-identity-access-rule delete <id>
```
  
---
### managed-identity-access-rule get subcommand
**Get a managed identity access rule.**
  
Retrieves details about a managed identity access rule.
  
```bash
tharsis managed-identity-access-rule get <id>
```
  
#### Options
  
#### json

Show final output as JSON.


---
### managed-identity-access-rule list subcommand
**Retrieve a list of managed identity access rules.**
  
Lists all access rules for a managed identity.
  
```bash
tharsis managed-identity-access-rule list \
  -managed-identity-id "trn:managed_identity:<group_path>/<managed_identity_name>"
```
  
#### Options
  
#### json

Show final output as JSON.

#### managed-identity-id

ID of the managed identity to get access rules for.

#### managed-identity-path <span style={{color:'orange'}}>!</span>

Resource path of the managed identity to get access rules for.\
**Deprecated**: use -managed-identity-id


---
### managed-identity-access-rule update subcommand
**Update a managed identity access rule.**
  
Modifies an existing managed identity access rule.
  
```bash
tharsis managed-identity-access-rule update \
  -allowed-user "trn:user:<username>" \
  <id>
```
  
#### Options
  
#### allowed-service-account <span style={{color:'green'}}>...</span>

Allowed service account ID.

#### allowed-team <span style={{color:'green'}}>...</span>

Allowed team ID.

#### allowed-user <span style={{color:'green'}}>...</span>

Allowed user ID.

#### json

Show final output as JSON.

#### module-attestation-policy <span style={{color:'green'}}>...</span>

Module attestation policy in format "[PredicateType=someval,]PublicKeyFile=/path/to/file".

#### verify-state-lineage

Verify state lineage.


---
## managed-identity-alias command
**Do operations on a managed identity alias.**
  
**Subcommands:**
  
- [`create`](#managed-identity-alias-create-subcommand) - Create a new managed identity alias.
- [`delete`](#managed-identity-alias-delete-subcommand) - Delete a managed identity alias.
  
Aliases allow referencing managed identities from other groups.
Use these commands to create and delete managed identity aliases.
  
---
### managed-identity-alias create subcommand
**Create a new managed identity alias.**
  
Creates an alias that references an existing managed
identity in another group.
  
```bash
tharsis managed-identity-alias create \
  -group-id "trn:group:<group_path>" \
  -alias-source-id "trn:managed_identity:<group_path>/<source_identity_name>" \
  prod-identity-alias
```
  
#### Options
  
#### alias-source-id

The ID or TRN of the source managed identity.

#### alias-source-path <span style={{color:'orange'}}>!</span>

The alias source path.\
**Deprecated**: use -alias-source-id

#### group-id

Group ID or TRN where the managed identity alias will be created.

#### group-path <span style={{color:'orange'}}>!</span>

Full path of the group where the managed identity alias will be created.\
**Deprecated**: use -group-id

#### json

Show final output as JSON.

#### name <span style={{color:'orange'}}>!</span>

The name of the managed identity alias.\
**Deprecated**: pass name as an argument


---
### managed-identity-alias delete subcommand
**Delete a managed identity alias.**
  
Removes a managed identity alias.
  
```bash
tharsis managed-identity-alias delete trn:managed_identity:<group_path>/<managed_identity_name>
```
  
#### Options
  
#### force, f

Force delete the managed identity alias.


---
## mcp command
**Start the Tharsis MCP server.**
  
Starts the Tharsis MCP server, enabling AI assistants to interact
with Tharsis resources through the Model Context Protocol.
By default, all toolsets are enabled in read-only mode for safety.

Available toolsets:
auth, run, job, configuration_version, workspace, group,
variable, managed_identity, documentation, terraform_module,
terraform_module_version, terraform_provider,
terraform_provider_platform

Environment variables (command-line options take precedence):
THARSIS_MCP_TOOLSETS               Comma-separated list of toolsets to enable
THARSIS_MCP_TOOLS                  Comma-separated list of individual tools to enable
THARSIS_MCP_READ_ONLY              Enable read-only mode (true/false)
THARSIS_MCP_NAMESPACE_MUTATION_ACL ACL patterns for namespace mutations

Access Control (ACL) Patterns:

Control which namespaces (groups and workspaces) can be modified using
simple wildcard patterns. ACL patterns apply to write operations (create,
update, delete, apply) to prevent accidental changes to production resources.
Read operations (get, list) are only restricted by user permissions.

Patterns are case-insensitive and support:
- Exact match: "prod" matches only "prod"
- Wildcard: "prod/*" matches any path starting with "prod/" (all levels)
- Prefix/suffix: "prod/team-*" matches "prod/team-alpha", "prod/team-beta"

Tip: Wildcards match across all path levels. To match a specific resource,
use exact paths like "prod/workspace" instead of "prod/*".

Examples:
- "prod" - Allow access to the "prod" group only
- "prod/workspace" - Allow access to specific workspace
- "prod/*" - Allow access to all resources under "prod" at any depth
- "prod/team-*" - Allow access to resources matching "prod/team-*"
- "dev,staging" - Allow access to "dev" and "staging" (comma-separated)

Restrictions:
- Wildcard-only patterns ("*") are not allowed
- Patterns cannot start with a wildcard ("*/workspace")
  
```bash
# Start MCP server with production profile in read-only mode
tharsis -p production mcp

# Start with specific toolsets
tharsis mcp -toolsets auth,run

# Start with namespace ACL restrictions
tharsis mcp -namespace-mutation-acl "dev/*,staging/*"
```

MCP Client Configuration (mcp.json):
```json
{
  "mcpServers": {
    "tharsis-prod": {
      "command": "tharsis",
      "args": ["-p", "production", "mcp"],
      "env": {"THARSIS_MCP_READ_ONLY": "true"},
      "disabled": false,
      "autoApprove": []
    },
    "tharsis-dev": {
      "command": "tharsis",
      "args": ["-p", "development", "mcp"],
      "env": {"THARSIS_MCP_TOOLSETS": "auth,run"},
      "disabled": false,
      "autoApprove": []
    }
  }
}
```
  
#### Options
  
#### namespace-mutation-acl

ACL patterns for namespace mutations (comma-separated).\
**Env:** `THARSIS_MCP_NAMESPACE_MUTATION_ACL`

#### read-only

Enable read-only mode (disables write tools).\
**Env:** `THARSIS_MCP_READ_ONLY`

#### tools

Comma-separated list of individual tools to enable.\
**Env:** `THARSIS_MCP_TOOLS`

#### toolsets

Comma-separated list of toolsets to enable.\
**Env:** `THARSIS_MCP_TOOLSETS`


---
## module command
**Do operations on a terraform module.**
  
**Subcommands:**
  
- [`create`](#module-create-subcommand) - Create a new Terraform module.
- [`create-attestation`](#module-create-attestation-subcommand) - Create a new module attestation.
- [`delete`](#module-delete-subcommand) - Delete a Terraform module.
- [`delete-attestation`](#module-delete-attestation-subcommand) - Delete a module attestation.
- [`delete-version`](#module-delete-version-subcommand) - Delete a module version.
- [`get`](#module-get-subcommand) - Get a single Terraform module.
- [`get-version`](#module-get-version-subcommand) - Get a module version by ID or TRN.
- [`list`](#module-list-subcommand) - Retrieve a paginated list of modules.
- [`list-attestations`](#module-list-attestations-subcommand) - Retrieve a paginated list of module attestations.
- [`list-versions`](#module-list-versions-subcommand) - Retrieve a paginated list of module versions.
- [`update`](#module-update-subcommand) - Update a Terraform module.
- [`update-attestation`](#module-update-attestation-subcommand) - Update a module attestation.
- [`upload-version`](#module-upload-version-subcommand) - Upload a new module version to the module registry.
  
The module registry stores Terraform modules with versioning and
attestation support. Use module commands to create, update, delete
modules, upload versions, manage attestations, and list modules
and versions.
  
---
### module create subcommand
**Create a new Terraform module.**
  
Creates a new Terraform module in the registry.
Argument format: module-name/system (e.g., vpc/aws).
  
```bash
tharsis module create \
  -group-id "trn:group:<group_path>" \
  -repository-url "https://github.com/example/terraform-aws-vpc" \
  -private \
  vpc/aws
```
  
#### Options
  
#### group-id

Parent group ID.

#### if-not-exists

Create a module if it does not already exist.\
**Default:** `false`

#### json

Show final output as JSON.

#### private

Whether the module is private.\
**Default:** `false`

#### repository-url

The repository URL for the module.


---
### module create-attestation subcommand
**Create a new module attestation.**
  
Creates a signed attestation for a module version
to verify its integrity.
  
```bash
tharsis module create-attestation \
  -description "Attestation for v1.0.0" \
  -data "aGVsbG8sIHdvcmxk" \
  trn:terraform_module:<module_path>
```
  
#### Options
  
#### data <span style={{color:'red'}}>*</span>

The attestation data (must be a Base64-encoded string).

#### description

Description for the attestation.

#### json

Show final output as JSON.


---
### module delete subcommand
**Delete a Terraform module.**
  
Permanently removes a module and all its versions
from the registry.
  
```bash
tharsis module delete trn:terraform_module:<group_path>/<module_name>/<system>
```
  
---
### module delete-attestation subcommand
**Delete a module attestation.**
  
Removes an attestation from a module.
  
```bash
tharsis module delete-attestation trn:terraform_module_attestation:<group_path>/<module_name>/<module_system>/<sha_sum>
```
  
---
### module delete-version subcommand
**Delete a module version.**
  
Removes a specific version of a module from the
registry.
  
```bash
tharsis module delete-version trn:terraform_module_version:<group_path>/<module_name>/<system>/<semantic_version>
```
  
#### Options
  
#### version

Optimistic locking version. Usually not required.


---
### module get subcommand
**Get a single Terraform module.**
  
Retrieves details about a Terraform module.
  
```bash
tharsis module get trn:terraform_module:<group_path>/<module_name>/<system>
```
  
#### Options
  
#### json

Show final output as JSON.


---
### module get-version subcommand
**Get a module version by ID or TRN.**
  
Retrieves details about a specific module version.
  
```bash
tharsis module get-version trn:terraform_module_version:<group_path>/<module_name>/<system>/<version>
```
  
#### Options
  
#### json

Show final output as JSON.

#### version <span style={{color:'orange'}}>!</span>

A semver compliant version tag to use as a filter.\
**Deprecated**: pass version TRN as argument


---
### module list subcommand
**Retrieve a paginated list of modules.**
  
Lists modules with pagination, filtering, and sorting.
  
```bash
tharsis module list \
  -group-id "trn:group:<group_path>" \
  -include-inherited \
  -sort-by "UPDATED_AT_DESC" \
  -limit 5 \
  -json
```
  
#### Options
  
#### cursor

The cursor string for manual pagination.

#### group-id

Filter to only modules in this group.

#### include-inherited

Include modules inherited from parent groups.\
**Default:** `false`

#### json

Show final output as JSON.

#### limit

Maximum number of result elements to return.\
**Default:** `100`

#### search

Filter to only modules containing this substring in their path.

#### sort-by

Sort by this field.

#### sort-order <span style={{color:'orange'}}>!</span>

Sort in this direction.\
**Values:** `ASC`, `DESC`\
**Deprecated**: use -sort-by


---
### module list-attestations subcommand
**Retrieve a paginated list of module attestations.**
  
Lists attestations for a module with pagination
and sorting.
  
```bash
tharsis module list-attestations \
  -sort-by "CREATED_AT_DESC" \
  -limit 10 \
  trn:terraform_module:<group_path>/<module_name>/<system>
```
  
#### Options
  
#### cursor

The cursor string for manual pagination.

#### digest

Filter to attestations with this digest.

#### json

Show final output as JSON.

#### limit

Maximum number of result elements to return.\
**Default:** `100`

#### sort-by

Sort by this field.

#### sort-order <span style={{color:'orange'}}>!</span>

Sort in this direction.\
**Values:** `ASC`, `DESC`\
**Deprecated**: use -sort-by


---
### module list-versions subcommand
**Retrieve a paginated list of module versions.**
  
Lists versions of a module with pagination and sorting.
  
```bash
tharsis module list-versions \
  -search "1.0" \
  -sort-by "CREATED_AT_DESC" \
  -limit 10 \
  trn:terraform_module:<group_path>/<module_name>/<system>
```
  
#### Options
  
#### cursor

The cursor string for manual pagination.

#### json

Show final output as JSON.

#### latest

Filter to only the latest version.\
**Conflicts:** `semantic-version`

#### limit

Maximum number of result elements to return.\
**Default:** `100`

#### search

Filter to versions containing this substring.

#### semantic-version

Filter to a specific semantic version.\
**Conflicts:** `latest`

#### sort-by

Sort by this field.

#### sort-order <span style={{color:'orange'}}>!</span>

Sort in this direction.\
**Values:** `ASC`, `DESC`\
**Deprecated**: use -sort-by


---
### module update subcommand
**Update a Terraform module.**
  
Modifies a module's repository URL or visibility.
  
```bash
tharsis module update \
  -repository-url "https://github.com/example/terraform-aws-vpc-v2" \
  -private true \
  trn:terraform_module:<group_path>/<module_name>/<system>
```
  
#### Options
  
#### json

Show final output as JSON.

#### private

Whether the module is private.

#### repository-url

The repository URL for the module.

#### version

Optimistic locking version. Usually not required.


---
### module update-attestation subcommand
**Update a module attestation.**
  
Modifies an existing module attestation.
  
```bash
tharsis module update-attestation \
  -description "Updated description" \
  trn:terraform_module_attestation:<group_path>/<module_name>/<system>/<sha_sum>
```
  
#### Options
  
#### description

Description for the attestation.

#### json

Show final output as JSON.


---
### module upload-version subcommand
**Upload a new module version to the module registry.**
  
Packages and uploads a new module version to the
registry.
  
```bash
tharsis module upload-version \
  -version "1.0.0" \
  -directory-path "./my-module" \
  trn:terraform_module:<group_path>/<module_name>/<system>
```
  
#### Options
  
#### directory-path

The path of the terraform module's directory.\
**Default:** `.`

#### version <span style={{color:'red'}}>*</span>

The semantic version for the new module version.


---
## plan command
**Create a speculative plan.**
  
Creates a speculative plan to preview infrastructure
changes without applying them. Supports run-scoped
Terraform and environment variables.

Terraform variables may be passed in via supported
options or from the environment with a 'TF_VAR_'
prefix.

Variable parsing precedence:
1. Terraform variables from the environment.
2. terraform.tfvars file from module's directory, if present.
3. terraform.tfvars.json file from module's directory, if present.
4. *.auto.tfvars, *.auto.tfvars.json files from the module's directory, if present.
5. -tf-var-file option(s).
6. -tf-var option(s).

NOTE: If the same variable is assigned multiple values, the last value found will be used.
  
```bash
tharsis plan -directory-path "./terraform" trn:workspace:<workspace_path>
```
  
#### Options
  
#### destroy

Designates this run as a destroy operation.\
**Default:** `false`

#### directory-path

The path of the root module's directory.\
**Conflicts:** `module-source`

#### env-var <span style={{color:'green'}}>...</span>

An environment variable as a key=value pair.

#### env-var-file <span style={{color:'green'}}>...</span>

The path to an environment variables file.

#### module-source

Remote module source specification.\
**Conflicts:** `directory-path`

#### module-version

Remote module version number. Uses latest if empty.

#### refresh

Whether to do the usual refresh step.\
**Default:** `true`

#### refresh-only

Whether to do ONLY a refresh operation.\
**Default:** `false`

#### target <span style={{color:'green'}}>...</span>

The Terraform address of the resources to be acted upon.

#### terraform-version

The Terraform CLI version to use for the run.

#### tf-var <span style={{color:'green'}}>...</span>

A terraform variable as a key=value pair.

#### tf-var-file <span style={{color:'green'}}>...</span>

The path to a .tfvars variables file.


---
## run command
**Do operations on runs.**
  
**Subcommands:**
  
- [`cancel`](#run-cancel-subcommand) - Cancel a run.
  
Runs are units of execution (plan or apply) that create, update,
or destroy infrastructure resources. Use run commands to cancel
runs gracefully or forcefully.
  
---
### run cancel subcommand
**Cancel a run.**
  
Stops a running or pending run. Use -force when
graceful cancellation is not sufficient.
  
```bash
tharsis run cancel -force <id>
```
  
#### Options
  
#### force, f

Force the run to cancel.


---
## runner-agent command
**Do operations on runner agents.**
  
**Subcommands:**
  
- [`assign-service-account`](#runner-agent-assign-service-account-subcommand) - Assign a service account to a runner agent.
- [`create`](#runner-agent-create-subcommand) - Create a new runner agent.
- [`delete`](#runner-agent-delete-subcommand) - Delete a runner agent.
- [`get`](#runner-agent-get-subcommand) - Get a runner agent.
- [`unassign-service-account`](#runner-agent-unassign-service-account-subcommand) - Unassign a service account from a runner agent.
- [`update`](#runner-agent-update-subcommand) - Update a runner agent.
  
Runner agents are distributed job executors responsible for
launching Terraform jobs that deploy infrastructure to the cloud.
Use runner-agent commands to create, update, delete, get agents,
and assign or unassign service accounts.
  
---
### runner-agent assign-service-account subcommand
**Assign a service account to a runner agent.**
  
Grants a service account permission to use a runner
agent.
  
```bash
tharsis runner-agent assign-service-account \
  trn:service_account:<group_path>/<service_account_name> \
  trn:runner:<group_path>/<runner_name>
```
  
---
### runner-agent create subcommand
**Create a new runner agent.**
  
Creates a new runner agent for executing Terraform
jobs.
  
```bash
tharsis runner-agent create \
  -group-id "trn:group:<group_path>" \
  -description "Production runner" \
  -run-untagged-jobs \
  -tag "prod" \
  -tag "us-east-1" \
  prod-runner
```
  
#### Options
  
#### description

Description for the runner agent.

#### disabled

Whether the runner is disabled.

#### group-id

Group ID or TRN where the runner agent will be created.

#### group-path <span style={{color:'orange'}}>!</span>

Full path of group where runner will be created.\
**Deprecated**: use -group-id

#### json

Show final output as JSON.

#### run-untagged-jobs

Allow the runner agent to execute jobs without tags.\
**Default:** `false`

#### runner-name <span style={{color:'orange'}}>!</span>

Name of the new runner agent.\
**Deprecated**: pass name as an argument

#### tag <span style={{color:'green'}}>...</span>

Tag for the runner agent.


---
### runner-agent delete subcommand
**Delete a runner agent.**
  
Permanently removes a runner agent.
  
```bash
tharsis runner-agent delete trn:runner:<group_path>/<runner_name>
```
  
#### Options
  
#### version

Optimistic locking version. Usually not required.


---
### runner-agent get subcommand
**Get a runner agent.**
  
Retrieves details about a runner agent.
  
```bash
tharsis runner-agent get trn:runner:<group_path>/<runner_name>
```
  
#### Options
  
#### json

Show final output as JSON.


---
### runner-agent unassign-service-account subcommand
**Unassign a service account from a runner agent.**
  
Revokes a service account's access to a runner agent.
  
```bash
tharsis runner-agent unassign-service-account \
  trn:service_account:<group_path>/<service_account_name> \
  trn:runner:<group_path>/<runner_name>
```
  
---
### runner-agent update subcommand
**Update a runner agent.**
  
Modifies an existing runner agent's configuration.
  
```bash
tharsis runner-agent update \
  -description "Updated description" \
  -disabled true \
  -tag "prod" \
  -tag "us-west-2" \
  trn:runner:<group_path>/<runner_name>
```
  
#### Options
  
#### description

Description for the runner agent.

#### disabled

Enable or disable the runner agent.

#### json

Show final output as JSON.

#### run-untagged-jobs

Allow the runner agent to execute jobs without tags.

#### tag <span style={{color:'green'}}>...</span>

Tag for the runner agent.

#### version

Optimistic locking version. Usually not required.


---
## service-account command
**Create an authentication token for a service account.**
  
**Subcommands:**
  
- [`create-token`](#service-account-create-token-subcommand) - Create a token for a service account.
  
Service accounts provide machine-to-machine authentication for
CI/CD pipelines and automation. Use service-account commands to
create authentication tokens.
  
---
### service-account create-token subcommand
**Create a token for a service account.**
  
Exchanges an identity provider token for a Tharsis
API token using OIDC authentication.
  
```bash
tharsis service-account create-token \
  -token "<oidc-token>" \
  trn:service_account:<group_path>/<service_account_name>
```
  
#### Options
  
#### json

Show final output as JSON.\
**Default:** `false`

#### token <span style={{color:'red'}}>*</span>

Initial authentication token from identity provider.


---
## sso command
**Log in to the OAuth2 provider and return an authentication token.**
  
**Subcommands:**
  
- [`login`](#sso-login-subcommand) - Log in to the OAuth2 provider and return an authentication token.
  
The sso command authenticates the CLI with the OAuth2 provider,
and allows making authenticated calls to Tharsis backend.
  
---
### sso login subcommand
**Log in to the OAuth2 provider and return an authentication token.**
  
Starts an embedded web server and opens a browser to the
OAuth2 provider's login page. If SSO is active, the user
is signed in automatically. The authentication token is
captured and stored for use in subsequent commands.
  
```bash
tharsis sso login
```
  
---
## terraform-provider command
**Do operations on a terraform provider.**
  
**Subcommands:**
  
- [`create`](#terraform-provider-create-subcommand) - Create a new terraform provider.
- [`upload-version`](#terraform-provider-upload-version-subcommand) - Upload a new Terraform provider version to the provider registry.
  
The provider registry stores Terraform providers with versioning
support. Use terraform-provider commands to create providers and
upload provider versions to the registry.
  
---
### terraform-provider create subcommand
**Create a new terraform provider.**
  
Creates a new Terraform provider in the registry.
  
```bash
tharsis terraform-provider create \
  -group-id "trn:group:<group_path>" \
  -repository-url "https://github.com/example/terraform-provider-example" \
  my-provider
```
  
#### Options
  
#### group-id

The ID of the group to create the provider in.

#### json

Show final output as JSON.

#### private

Set to false to allow all groups to view and use the terraform provider.\
**Default:** `true`

#### repository-url

The repository URL for this terraform provider.


---
### terraform-provider upload-version subcommand
**Upload a new Terraform provider version to the provider registry.**
  
Packages and uploads a new provider version to the
registry.
  
```bash
tharsis terraform-provider upload-version \
  -directory-path "./my-provider" \
  trn:terraform_provider:<group_path>/<name>
```
  
#### Options
  
#### directory-path

The path of the terraform provider's directory.\
**Default:** `.`


---
## terraform-provider-mirror command
**Mirror Terraform providers from any Terraform registry.**
  
**Subcommands:**
  
- [`delete-platform`](#terraform-provider-mirror-delete-platform-subcommand) - Delete a terraform provider platform from mirror.
- [`delete-version`](#terraform-provider-mirror-delete-version-subcommand) - Delete a terraform provider version from mirror.
- [`get-version`](#terraform-provider-mirror-get-version-subcommand) - Get a mirrored terraform provider version.
- [`list-platforms`](#terraform-provider-mirror-list-platforms-subcommand) - Retrieve a paginated list of provider platform mirrors.
- [`list-versions`](#terraform-provider-mirror-list-versions-subcommand) - Retrieve a paginated list of provider version mirrors.
- [`sync`](#terraform-provider-mirror-sync-subcommand) - Sync provider platforms from upstream registry to mirror.
  
The provider mirror caches Terraform providers from any registry
for use within a group hierarchy. It supports Terraform's Provider
Network Mirror Protocol and gives root group owners control over
which providers, platform packages, and registries are available.
Use these commands to sync providers, list versions and platforms,
get version details, and delete versions or platforms.
  
---
### terraform-provider-mirror delete-platform subcommand
**Delete a terraform provider platform from mirror.**
  
Removes a platform binary from the provider mirror.
  
```bash
tharsis terraform-provider-mirror delete-platform <platform-mirror-id>
```
  
---
### terraform-provider-mirror delete-version subcommand
**Delete a terraform provider version from mirror.**
  
Removes a mirrored provider version and its platform
binaries. Use -force when the version hosts packages.
  
```bash
tharsis terraform-provider-mirror delete-version -force <version-mirror-id>
```
  
#### Options
  
#### force, f

Skip confirmation prompt.


---
### terraform-provider-mirror get-version subcommand
**Get a mirrored terraform provider version.**
  
Retrieves details about a mirrored provider version.
  
```bash
tharsis terraform-provider-mirror get-version <version-mirror-id>
```
  
#### Options
  
#### json

Show final output as JSON.


---
### terraform-provider-mirror list-platforms subcommand
**Retrieve a paginated list of provider platform mirrors.**
  
Lists mirrored platform binaries for a provider
version with pagination and sorting.
  
```bash
tharsis terraform-provider-mirror list-platforms \
  -os "linux" \
  -architecture "amd64" \
  -sort-by "CREATED_AT_DESC" \
  trn:terraform_provider_version_mirror:<group_path>/<provider_namespace>/<provider_name>/<semantic_version>
```
  
#### Options
  
#### architecture

Filter to platforms with this architecture.

#### cursor

The cursor string for manual pagination.

#### json

Show final output as JSON.

#### limit

Maximum number of result elements to return.\
**Default:** `100`

#### os

Filter to platforms with this OS.

#### sort-by

Sort by this field.\
**Values:** `CREATED_AT_ASC`, `CREATED_AT_DESC`


---
### terraform-provider-mirror list-versions subcommand
**Retrieve a paginated list of provider version mirrors.**
  
Lists mirrored provider versions with pagination
and sorting.
  
```bash
tharsis terraform-provider-mirror list-versions \
  -sort-by "CREATED_AT_DESC" \
  -limit 10 \
  <namespace_path>
```
  
#### Options
  
#### cursor

The cursor string for manual pagination.

#### json

Show final output as JSON.

#### limit

Maximum number of result elements to return.\
**Default:** `100`

#### sort-by

Sort by this field.\
**Values:** `CREATED_AT_ASC`, `CREATED_AT_DESC`, `TYPE_ASC`, `TYPE_DESC`\
**Conflicts:** `sort-order`

#### sort-order <span style={{color:'orange'}}>!</span>

Sort in this direction.\
**Values:** `ASC`, `DESC`\
**Deprecated**: use -sort-by\
**Conflicts:** `sort-by`


---
### terraform-provider-mirror sync subcommand
**Sync provider platforms from upstream registry to mirror.**
  
Downloads provider platform packages from an upstream
registry and uploads them to the Tharsis mirror. Use
-platform multiple times to specify platforms. By default,
syncs all platforms for the latest version.

Only missing packages are uploaded. To re-upload, delete
the platform mirror first via "tharsis
terraform-provider-mirror delete-platform".

For private registries, tokens are resolved in order:
1. TF_TOKEN_\<hostname\> environment variable
2. Federated registry service discovery with a
matching CLI profile

Fully Qualified Name (FQN) format:

\[registry hostname/\]\{namespace\}/\{provider name\}

The hostname can be omitted for providers from the
default public registry (registry.terraform.io).

Examples: registry.terraform.io/hashicorp/aws, hashicorp/aws
  
```bash
tharsis terraform-provider-mirror sync \
  -group-id "my-group" \
  -version "1.0.0" \
  -platform "linux_amd64" \
  hashicorp/aws
```
  
#### Options
  
#### group-id

The ID of the root group to create the mirror in.

#### group-path <span style={{color:'orange'}}>!</span>

Full path to the root group where this Terraform provider version will be mirrored.\
**Deprecated**: use -group-id

#### platform <span style={{color:'green'}}>...</span>

Platform to sync (format: os_arch). If not specified, syncs all platforms.

#### version

The provider version to sync. If not specified, uses the latest version.


---
## tf-exec command
**Run terraform with Tharsis auth and workspace variables injected.**
  
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
matching the workspace's configured version in ~/.tharsis/tf-installs/\<version\>/.
If not found, it downloads that exact version from releases.hashicorp.com and
caches it there for future use.

Authentication:

The current profile's auth token is injected as TF_TOKEN_\<host\> where \<host\>
is the Tharsis instance hostname with dots replaced by underscores. This
authenticates terraform against the Tharsis remote backend.

Variables:

All variables configured on the workspace and its parent groups are injected:

- Terraform variables (category: terraform) -\> TF_VAR_\<key\>=\<value\>
- Environment variables (category: environment) -\> \<key\>=\<value\>

Sensitive variable values are automatically fetched and injected.

Exit Code:

The exact exit code returned by terraform is passed through unchanged.
  
```bash
tharsis tf-exec -workspace my/group/workspace show
tharsis tf-exec -workspace trn:workspace:my/group/workspace plan
tharsis tf-exec -workspace my/group/workspace -work-dir ./infra apply
```
  
#### Options
  
#### tf-path

Path to an existing terraform binary. If omitted, the version from the last applied run is downloaded automatically.

#### work-dir

Working directory for terraform. If omitted, a persistent cache directory keyed by workspace is used.

#### workspace <span style={{color:'red'}}>*</span>

The Tharsis workspace path or TRN (e.g. my/group/workspace or trn:workspace:my/group/workspace).


---
## version command
**Get the CLI's version.**
  
Returns the CLI's version.
  
```bash
tharsis version -json
```
  
#### Options
  
#### json

Show final output as JSON.\
**Default:** `false`


---
## workspace command
**Do operations on workspaces.**
  
**Subcommands:**
  
- [`add-membership`](#workspace-add-membership-subcommand) - Add a membership to a workspace.
- [`assign-managed-identity`](#workspace-assign-managed-identity-subcommand) - Assign a managed identity to a workspace.
- [`create`](#workspace-create-subcommand) - Create a new workspace.
- [`delete`](#workspace-delete-subcommand) - Delete a workspace.
- [`delete-terraform-var`](#workspace-delete-terraform-var-subcommand) - Delete a terraform variable from a workspace.
- [`get`](#workspace-get-subcommand) - Get a single workspace.
- [`get-assigned-managed-identities`](#workspace-get-assigned-managed-identities-subcommand) - Get assigned managed identities for a workspace.
- [`get-membership`](#workspace-get-membership-subcommand) - Get a workspace membership.
- [`get-terraform-var`](#workspace-get-terraform-var-subcommand) - Get a terraform variable for a workspace.
- [`label`](#workspace-label-subcommand) - Manage labels on a workspace.
- [`list`](#workspace-list-subcommand) - Retrieve a paginated list of workspaces.
- [`list-environment-vars`](#workspace-list-environment-vars-subcommand) - List all environment variables in a workspace.
- [`list-memberships`](#workspace-list-memberships-subcommand) - Retrieve a list of workspace memberships.
- [`list-terraform-vars`](#workspace-list-terraform-vars-subcommand) - List all terraform variables in a workspace.
- [`migrate`](#workspace-migrate-subcommand) - Migrate a workspace to a new group.
- [`outputs`](#workspace-outputs-subcommand) - Get the state version outputs for a workspace.
- [`remove-membership`](#workspace-remove-membership-subcommand) - Remove a workspace membership.
- [`set-environment-vars`](#workspace-set-environment-vars-subcommand) - Set environment variables for a workspace.
- [`set-terraform-var`](#workspace-set-terraform-var-subcommand) - Set a terraform variable for a workspace.
- [`set-terraform-vars`](#workspace-set-terraform-vars-subcommand) - Set terraform variables for a workspace.
- [`unassign-managed-identity`](#workspace-unassign-managed-identity-subcommand) - Unassign a managed identity from a workspace.
- [`update`](#workspace-update-subcommand) - Update a workspace.
- [`update-membership`](#workspace-update-membership-subcommand) - Update a workspace membership.
  
Workspaces contain Terraform deployments, state, runs, and variables.
Use workspace commands to create, update, delete workspaces, assign
and unassign managed identities, set Terraform and environment
variables, manage memberships, and view workspace outputs.
  
---
### workspace add-membership subcommand
**Add a membership to a workspace.**
  
Grants a user, service account, or team access to a
workspace. Exactly one identity flag must be specified.
  
```bash
tharsis workspace add-membership \
  -role-id "trn:role:owner" \
  -user-id "trn:user:john.smith" \
  trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Show final output as JSON.

#### role <span style={{color:'orange'}}>!</span>

Role name for new membership.\
**Deprecated**: use -role-id

#### role-id

The role ID for the membership.

#### service-account-id

The service account ID for the membership.\
**Conflicts:** `user-id`, `team-id`, `username`, `team-name`

#### team-id

The team ID for the membership.\
**Conflicts:** `user-id`, `service-account-id`, `username`, `team-name`

#### team-name <span style={{color:'orange'}}>!</span>

Team name for the new membership.\
**Deprecated**: use -team-id\
**Conflicts:** `user-id`, `service-account-id`, `team-id`, `username`

#### user-id

The user ID for the membership.\
**Conflicts:** `service-account-id`, `team-id`, `username`, `team-name`

#### username <span style={{color:'orange'}}>!</span>

Username for the new membership.\
**Deprecated**: use -user-id\
**Conflicts:** `user-id`, `service-account-id`, `team-id`, `team-name`


---
### workspace assign-managed-identity subcommand
**Assign a managed identity to a workspace.**
  
Assigns a managed identity to a workspace for cloud
provider authentication.
  
```bash
tharsis workspace assign-managed-identity \
  trn:workspace:<workspace_path> \
  trn:managed_identity:<group_path>/<identity_name>
```
  
---
### workspace create subcommand
**Create a new workspace.**
  
Creates a new workspace with optional description,
max job duration, and managed identity assignments.
  
```bash
tharsis workspace create \
  -parent-group-id "trn:group:<group_path>" \
  -description "Production workspace" \
  -terraform-version "1.5.0" \
  -max-job-duration 60 \
  -prevent-destroy-plan \
  -managed-identity "trn:managed_identity:<group_path>/<identity_name>" \
  -label "env=prod" \
  -label "team=platform" \
  <name>
```
  
#### Options
  
#### description

Description for the new workspace.

#### if-not-exists

Create a workspace if it does not already exist.\
**Default:** `false`

#### json

Show final output as JSON.

#### label <span style={{color:'green'}}>...</span>

Labels for the new workspace (key=value).

#### managed-identity <span style={{color:'green'}}>...</span>

The ID of a managed identity to assign.

#### max-job-duration

The amount of minutes before a job is gracefully canceled (Default 720).

#### parent-group-id

Parent group ID.

#### prevent-destroy-plan

Whether a run/plan will be prevented from destroying deployed resources.\
**Default:** `false`

#### terraform-version

The default Terraform CLI version for the new workspace.


---
### workspace delete subcommand
**Delete a workspace.**
  
Permanently removes a workspace. Use -force to delete
even if resources are deployed.
  
```bash
tharsis workspace delete -force trn:workspace:<workspace_path>
```
  
#### Options
  
#### force, f

Force the workspace to delete even if resources are deployed.

#### version

Optimistic locking version. Usually not required.


---
### workspace delete-terraform-var subcommand
**Delete a terraform variable from a workspace.**
  
Removes a Terraform variable from a workspace.
  
```bash
tharsis workspace delete-terraform-var \
  -key "region" \
  trn:workspace:<workspace_path>
```
  
#### Options
  
#### key <span style={{color:'red'}}>*</span>

Variable key.

#### version

Optimistic locking version. Usually not required.


---
### workspace get subcommand
**Get a single workspace.**
  
Retrieves details about a workspace by ID or path.
  
```bash
tharsis workspace get trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Show final output as JSON.


---
### workspace get-assigned-managed-identities subcommand
**Get assigned managed identities for a workspace.**
  
Lists all managed identities assigned to a workspace.
  
```bash
tharsis workspace get-assigned-managed-identities trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Show final output as JSON.


---
### workspace get-membership subcommand
**Get a workspace membership.**
  
Retrieves details about a specific workspace membership.
  
```bash
tharsis workspace get-membership \
  -user-id "trn:user:<username>" \
  trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Show final output as JSON.

#### service-account-id

Service account ID to find the workspace membership for.\
**Conflicts:** `user-id`, `team-id`, `username`, `team-name`

#### team-id

Team ID to find the workspace membership for.\
**Conflicts:** `user-id`, `service-account-id`, `username`, `team-name`

#### team-name <span style={{color:'orange'}}>!</span>

Team name to find the workspace membership for.\
**Deprecated**: use -team-id\
**Conflicts:** `user-id`, `service-account-id`, `team-id`, `username`

#### user-id

User ID to find the workspace membership for.\
**Conflicts:** `service-account-id`, `team-id`, `username`, `team-name`

#### username <span style={{color:'orange'}}>!</span>

Username to find the workspace membership for.\
**Deprecated**: use -user-id\
**Conflicts:** `user-id`, `service-account-id`, `team-id`, `team-name`


---
### workspace get-terraform-var subcommand
**Get a terraform variable for a workspace.**
  
Retrieves a Terraform variable from a workspace.
  
```bash
tharsis workspace get-terraform-var \
  -key "region" \
  trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Show final output as JSON.

#### key <span style={{color:'red'}}>*</span>

Variable key.

#### show-sensitive

Show the actual value of sensitive variables (requires appropriate permissions).\
**Default:** `false`


---
### workspace label subcommand
**Manage labels on a workspace.**
  
Adds, updates, removes, or overwrites labels on a
workspace.

Label operations:
key=value  Add or update a label
key-       Remove a label (not allowed with -overwrite)
  
```bash
tharsis workspace label \
  -overwrite \
  trn:workspace:<workspace_path> \
  env=prod \
  tier=frontend
```
  
#### Options
  
#### json

Show final output as JSON.

#### overwrite

Replace all existing labels with the specified labels.\
**Default:** `false`


---
### workspace list subcommand
**Retrieve a paginated list of workspaces.**
  
Lists workspaces with pagination, filtering, and sorting.
  
```bash
tharsis workspace list \
  -group-id "trn:group:<group_path>" \
  -label "env=prod" \
  -label "team=platform" \
  -sort-by "FULL_PATH_ASC" \
  -limit 5 \
  -json
```
  
#### Options
  
#### cursor

The cursor string for manual pagination.

#### group-id

Filter to only workspaces in this group.

#### group-path <span style={{color:'orange'}}>!</span>

Filter to only workspaces in this group path.\
**Deprecated**: use -group-id

#### json

Show final output as JSON.

#### label <span style={{color:'green'}}>...</span>

Filter by label (key=value).

#### limit

Maximum number of result elements to return.\
**Default:** `100`

#### search

Filter to only workspaces containing this substring in their path.

#### sort-by

Sort by this field.

#### sort-order <span style={{color:'orange'}}>!</span>

Sort in this direction.\
**Values:** `ASC`, `DESC`\
**Deprecated**: use -sort-by


---
### workspace list-environment-vars subcommand
**List all environment variables in a workspace.**
  
Lists all environment variables from a workspace and
its parent groups.
  
```bash
tharsis workspace list-environment-vars -show-sensitive trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Show final output as JSON.

#### show-sensitive

Show the actual values of sensitive variables (requires appropriate permissions).\
**Default:** `false`


---
### workspace list-memberships subcommand
**Retrieve a list of workspace memberships.**
  
Lists all memberships for a workspace.
  
```bash
tharsis workspace list-memberships trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Show final output as JSON.


---
### workspace list-terraform-vars subcommand
**List all terraform variables in a workspace.**
  
Lists all Terraform variables from a workspace and its
parent groups.
  
```bash
tharsis workspace list-terraform-vars -show-sensitive trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Show final output as JSON.

#### show-sensitive

Show the actual values of sensitive variables (requires appropriate permissions).\
**Default:** `false`


---
### workspace migrate subcommand
**Migrate a workspace to a new group.**
  
Moves a workspace to a different group.
  
```bash
tharsis workspace migrate \
  -new-group-id "trn:group:<group_path>" \
  trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Show final output as JSON.

#### new-group-id <span style={{color:'red'}}>*</span>

New parent group ID.


---
### workspace outputs subcommand
**Get the state version outputs for a workspace.**
  
Retrieves the Terraform state outputs for a workspace.

Supported output types:
- Decorated (shows if map, list, etc. default).
- JSON.
- Raw (just the value, limited).

Use -output-name to filter to a specific output.
  
```bash
tharsis workspace outputs trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Show final output as JSON.\
**Default:** `false`\
**Conflicts:** `raw`

#### output-name

The name of the output variable to use as a filter. Required for -raw option.

#### raw

For any value that can be converted to a string, output just the raw value.\
**Default:** `false`\
**Conflicts:** `json`


---
### workspace remove-membership subcommand
**Remove a workspace membership.**
  
Revokes a membership from a workspace.
  
```bash
tharsis workspace remove-membership <id>
```
  
#### Options
  
#### version

Optimistic locking version. Usually not required.


---
### workspace set-environment-vars subcommand
**Set environment variables for a workspace.**
  
Replaces all environment variables in a workspace from
a file. Does not support sensitive variables.
  
```bash
tharsis workspace set-environment-vars \
  -env-var-file "vars.env" \
  trn:workspace:<workspace_path>
```
  
#### Options
  
#### env-var-file <span style={{color:'red'}}>*</span> <span style={{color:'green'}}>...</span>

Path to an environment variables file.


---
### workspace set-terraform-var subcommand
**Set a terraform variable for a workspace.**
  
Creates or updates a Terraform variable for a workspace.
  
```bash
tharsis workspace set-terraform-var \
  -key "region" \
  -value "us-east-1" \
  trn:workspace:<workspace_path>
```
  
#### Options
  
#### key <span style={{color:'red'}}>*</span>

Variable key.

#### sensitive

Mark variable as sensitive.\
**Default:** `false`

#### value <span style={{color:'red'}}>*</span>

Variable value.


---
### workspace set-terraform-vars subcommand
**Set terraform variables for a workspace.**
  
Replaces all Terraform variables in a workspace from a
tfvars file. Does not support sensitive variables.
  
```bash
tharsis workspace set-terraform-vars \
  -tf-var-file "terraform.tfvars" \
  trn:workspace:<workspace_path>
```
  
#### Options
  
#### tf-var-file <span style={{color:'red'}}>*</span> <span style={{color:'green'}}>...</span>

Path to a .tfvars file.


---
### workspace unassign-managed-identity subcommand
**Unassign a managed identity from a workspace.**
  
Removes a managed identity assignment from a workspace.
  
```bash
tharsis workspace unassign-managed-identity \
  trn:workspace:<workspace_path> \
  trn:managed_identity:<group_path>/<identity_name>
```
  
---
### workspace update subcommand
**Update a workspace.**
  
Modifies a workspace's description or max job duration.
  
```bash
tharsis workspace update \
  -description "Updated production workspace" \
  -terraform-version "1.6.0" \
  -max-job-duration 120 \
  -prevent-destroy-plan true \
  trn:workspace:<workspace_path>
```
  
#### Options
  
#### description

Description for the workspace.

#### json

Show final output as JSON.

#### label <span style={{color:'green'}}>...</span>

Labels for the workspace (key=value).

#### max-job-duration

The amount of minutes before a job is gracefully canceled.

#### prevent-destroy-plan

Whether a run/plan will be prevented from destroying deployed resources.

#### terraform-version

The default Terraform CLI version for the workspace.

#### version

Optimistic locking version. Usually not required.


---
### workspace update-membership subcommand
**Update a workspace membership.**
  
Changes the role of an existing workspace membership.
  
```bash
tharsis workspace update-membership \
  -role-id "trn:role:<role_name>" \
  <id>
```
  
#### Options
  
#### json

Show final output as JSON.

#### role <span style={{color:'orange'}}>!</span>

Role name for the membership.\
**Deprecated**: use -role-id

#### role-id

The role ID for the membership.

#### version

Optimistic locking version. Usually not required.


---
## Frequently asked questions (FAQ)
### Is configuring a profile necessary?
By default, the CLI will use the default Tharsis endpoint passed in at build-time. Unless a different endpoint is needed, no profile configuration is necessary. Simply run `tharsis sso login` and the `default` profile will be created and stored in the settings file.
### How do I use profiles?
The profile can be specified using the `-p` global flag or the `THARSIS_PROFILE` environment variable. The flag **must** come before a command name. For example, `tharsis -p local group list` will list all the groups using the Tharsis endpoint in the `local` profile. Service accounts can use profiles in the same manner as human users.
### Where are the settings and credentials files located?
The settings file is located at `~/.tharsis/settings.json` and contains profile configuration (endpoints, options). Credentials are stored separately in `~/.tharsis/credentials.json` so they can have stricter permissions.
:::caution
**Never** share the credentials file as it contains sensitive data like the authentication token from SSO!
:::
### How do I disable colored output?
Set the `NO_COLOR` environment variable to any value to disable colored output. For example, `NO_COLOR=1 tharsis group list`.
### Can I use Terraform variables from the CLI's environment inside a run?
Yes, environment variables with the `TF_VAR_` prefix are passed as Terraform variables with the prefix stripped. For example, `TF_VAR_region=us-east-1` sets a Terraform variable named `region` to `us-east-1`.