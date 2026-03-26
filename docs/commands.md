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

Show this usage message.

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
  
The apply command creates and applies a Terraform run.
It first creates a plan, then applies it after approval.
Supports setting run-scoped Terraform / environment variables.

Terraform variables may be passed in via supported
options or from the environment with a 'TF_VAR_' prefix.
  
```shell
tharsis apply -directory-path ./terraform trn:workspace:<workspace_path>
```
  
#### Options
  
#### auto-approve

Skip interactive approval of the plan.\
**Default:** `false`

#### directory-path

The path of the root module's directory.

#### env-var <span style={{color:'green'}}>...</span>

An environment variable as a key=value pair.

#### env-var-file <span style={{color:'green'}}>...</span>

The path to an environment variables file.

#### input

Ask for input for variables if not directly set.\
**Default:** `true`

#### module-source

Remote module source specification.

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
  
The caller-identity command returns information about the
authenticated caller (User or ServiceAccount).
  
```shell
tharsis caller-identity
```
  
#### Options
  
#### json

Show output as JSON.\
**Default:** `false`


---
## configure command
**Create or update a profile.**
  
**Subcommands:**
  
- [`delete`](#configure-delete-subcommand) - Remove a profile.
- [`list`](#configure-list-subcommand) - Show all profiles.
  
The configure command creates or updates a profile. If no
options are specified, the command prompts for values.
  
```shell
tharsis configure \
  -http-endpoint https://api.tharsis.example.com \
  -profile prod-example
```
  
#### Options
  
#### endpoint-url <span style={{color:'orange'}}>!</span>

The Tharsis HTTP API endpoint (in URL format).\
**Deprecated**: use -http-endpoint instead

#### http-endpoint

The Tharsis HTTP API endpoint (in URL format).

#### insecure-tls-skip-verify

Allow TLS but disable verification of the gRPC server's certificate chain and hostname. This should ONLY be true for testing as it could allow the CLI to connect to an impersonated server.\
**Default:** `false`

#### profile

The name of the profile to set.


---
### configure delete subcommand
**Remove a profile.**
  
The configure delete command removes a profile and its
credentials with the given name.
  
```shell
tharsis configure delete prod-example
```
  
---
### configure list subcommand
**Show all profiles.**
  
The configure list command prints information about all profiles.
  
```shell
tharsis configure list
```
  
---
## destroy command
**Destroy workspace resources.**
  
The destroy command destroys resources in a workspace.
It creates a destroy plan, then applies it after approval.
Supports setting run-scoped Terraform / environment variables.

Terraform variables may be passed in via supported
options or from the environment with a 'TF_VAR_' prefix.
  
```shell
tharsis destroy -directory-path ./terraform trn:workspace:<workspace_path>
```
  
#### Options
  
#### auto-approve

Skip interactive approval of the plan.\
**Default:** `false`

#### directory-path

The path of the root module's directory.

#### env-var <span style={{color:'green'}}>...</span>

An environment variable as a key=value pair.

#### env-var-file <span style={{color:'green'}}>...</span>

The path to an environment variables file.

#### input

Ask for input for variables if not directly set.\
**Default:** `true`

#### module-source

Remote module source specification.

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
  
The group add-membership command adds a membership to a group.
Exactly one of -user-id, -service-account-id, or -team-id must be specified.
  
```shell
tharsis group add-membership \
  -role-id trn:role:<role_name> \
  -user-id trn:user:<username> \
  trn:group:<group_path>
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`

#### role <span style={{color:'orange'}}>!</span>

The role for the membership.\
**Deprecated**: use -role-id

#### role-id

The role ID for the membership.

#### service-account-id

The service account ID for the membership.

#### team-id

The team ID for the membership.

#### team-name <span style={{color:'orange'}}>!</span>

The team name for the membership.\
**Deprecated**: use -team-id

#### user-id

The user ID for the membership.

#### username <span style={{color:'orange'}}>!</span>

The username for the membership.\
**Deprecated**: use -user-id


---
### group create subcommand
**Create a new group.**
  
The group create command creates a new group. It allows
setting a group's description (optional). Shows final
output as JSON, if specified. Idempotent when used with
-if-not-exists option.
  
```shell
tharsis group create \
  -parent-group-id trn:group:<group_path> \
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

Show final output as JSON.\
**Default:** `false`

#### parent-group-id

Parent group ID.


---
### group delete subcommand
**Delete a group.**
  
The group delete command deletes a group by its ID. Includes
a force flag to delete the group even if resources are
deployed (dangerous!).
  
```shell
tharsis group delete \
  -force \
  trn:group:<group_path>
```
  
#### Options
  
#### force

Force delete the group.

#### version

Metadata version of the resource to be deleted. In most cases, this is not required.


---
### group delete-terraform-var subcommand
**Delete a terraform variable from a group.**
  
The group delete-terraform-var command deletes a terraform variable from a group.
  
```shell
tharsis group delete-terraform-var \
  -key region \
  trn:group:<group_path>
```
  
#### Options
  
#### key <span style={{color:'red'}}>*</span>

Variable key.

#### version

Metadata version of the resource to be deleted. In most cases, this is not required.


---
### group get subcommand
**Get a single group.**
  
The group get command retrieves a single group by its ID.
Shows output as JSON, if specified.
  
```shell
tharsis group get \
  -json \
  trn:tharsis:group:<group_path>
```
  
#### Options
  
#### json

Show output as JSON.\
**Default:** `false`


---
### group get-membership subcommand
**Get a group membership.**
  
The group get-membership command retrieves details about a specific group membership.
  
```shell
tharsis group get-membership \
  -user-id trn:user:<username> \
  trn:group:<group_path>
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`

#### service-account-id

Service account ID to find the group membership for.

#### team-id

Team ID to find the group membership for.

#### team-name <span style={{color:'orange'}}>!</span>

Team name to find the group membership for.\
**Deprecated**: use -team-id

#### user-id

User ID to find the group membership for.

#### username <span style={{color:'orange'}}>!</span>

Username to find the group membership for.\
**Deprecated**: use -user-id


---
### group get-terraform-var subcommand
**Get a terraform variable for a group.**
  
The group get-terraform-var command retrieves a terraform variable for a group.
  
```shell
tharsis group get-terraform-var \
  -key region \
  trn:group:<group_path>
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`

#### key <span style={{color:'red'}}>*</span>

Variable key.

#### show-sensitive

Show the actual value of sensitive variables (requires appropriate permissions).\
**Default:** `false`


---
### group list subcommand
**Retrieve a paginated list of groups.**
  
The group list command prints information about (likely
multiple) groups. Supports pagination, filtering and
sorting the output.
  
```shell
tharsis group list \
  -parent-id trn:group:<parent_group_path> \
  -sort-by FULL_PATH_ASC \
  -limit 5 \
  -json
```
  
#### Options
  
#### cursor

The cursor string for manual pagination.

#### json

Show final output as JSON.\
**Default:** `false`

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
**Values:** `FULL_PATH_ASC`, `FULL_PATH_DESC`, `GROUP_LEVEL_ASC`, `GROUP_LEVEL_DESC`, `UPDATED_AT_ASC`, `UPDATED_AT_DESC`

#### sort-order <span style={{color:'orange'}}>!</span>

Sort in this direction.\
**Values:** `ASC`, `DESC`\
**Deprecated**: use -sort-by


---
### group list-environment-vars subcommand
**List all environment variables in a group.**
  
The group list-environment-vars command retrieves all terraform
variables from a group and its parent groups.
  
```shell
tharsis group list-environment-vars -show-sensitive trn:group:<group_path>
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`

#### show-sensitive

Show the actual values of sensitive variables (requires appropriate permissions).\
**Default:** `false`


---
### group list-memberships subcommand
**Retrieve a list of group memberships.**
  
The group list-memberships command prints information about
memberships for a specific group.
  
```shell
tharsis group list-memberships trn:group:<group_path>
```
  
#### Options
  
#### json

Show final output as JSON.\
**Default:** `false`


---
### group list-terraform-vars subcommand
**List all terraform variables in a group.**
  
The group list-terraform-vars command retrieves all terraform
variables from a group and its parent groups.
  
```shell
tharsis group list-terraform-vars -show-sensitive trn:group:<group_path>
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`

#### show-sensitive

Show the actual values of sensitive variables (requires appropriate permissions).\
**Default:** `false`


---
### group migrate subcommand
**Migrate a group to a new parent or to top-level.**
  
The group migrate command migrates a group to another parent group or to top-level.
Omit -new-parent-id to migrate to top-level.
  
```shell
tharsis group migrate \
  -new-parent-id trn:group:<parent_group_path> \
  trn:group:<group_path>
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`

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
  
The group remove-membership command removes a membership from a group.
  
```shell
tharsis group remove-membership <id>
```
  
#### Options
  
#### version

Metadata version of the resource to be deleted. In most cases, this is not required.


---
### group set-environment-vars subcommand
**Set environment variables for a group.**
  
The group set-environment-vars command sets environment variables for a group.
Command will overwrite any existing environment variables in the target group!
Note: This command does not support sensitive variables.
  
```shell
tharsis group set-environment-vars \
  -env-var-file vars.env \
  trn:group:<group_path>
```
  
#### Options
  
#### env-var-file <span style={{color:'red'}}>*</span> <span style={{color:'green'}}>...</span>

Path to an environment variables file.


---
### group set-terraform-var subcommand
**Set a terraform variable for a group.**
  
The group set-terraform-var command creates or updates a terraform variable for a group.
  
```shell
tharsis group set-terraform-var \
  -key region \
  -value us-east-1 \
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
  
The group set-terraform-vars command sets terraform variables for a group.
Command will overwrite any existing Terraform variables in the target group!
Note: This command does not support sensitive variables.
  
```shell
tharsis group set-terraform-vars \
  -tf-var-file terraform.tfvars \
  trn:group:<group_path>
```
  
#### Options
  
#### tf-var-file <span style={{color:'red'}}>*</span> <span style={{color:'green'}}>...</span>

Path to a .tfvars file.


---
### group update subcommand
**Update a group.**
  
The group update command updates a group. Currently, it
supports updating the description. Shows final output
as JSON, if specified.
  
```shell
tharsis group update \
  -description "Updated operations group" \
  trn:group:<group_path>
```
  
#### Options
  
#### description

Description for the group.

#### json

Show final output as JSON.\
**Default:** `false`

#### version

Metadata version of the resource to be updated. In most cases, this is not required.


---
### group update-membership subcommand
**Update a group membership.**
  
The group update-membership command updates a group membership's role.
  
```shell
tharsis group update-membership \
  -role-id trn:role:<role_name> \
  <id>
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`

#### role <span style={{color:'orange'}}>!</span>

New role for the membership.\
**Deprecated**: use -role-id

#### role-id <span style={{color:'red'}}>*</span>

The role ID for the membership.

#### version

Metadata version of the resource to be updated. In most cases, this is not required.


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
  
The managed-identity create command creates a new managed identity.
  
```shell
tharsis managed-identity create \
  -group-id trn:group:<group_path> \
  -type aws_federated \
  -aws-federated-role arn:aws:iam::123456789012:role/MyRole \
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

Show final output as JSON.\
**Default:** `false`

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
  
The managed-identity delete command deletes a managed identity.

Use with caution as deleting a managed identity is irreversible!
  
```shell
tharsis managed-identity delete -force trn:managed_identity:<group_path>/<managed_identity_name>
```
  
#### Options
  
#### force

Force delete the managed identity.


---
### managed-identity get subcommand
**Get a single managed identity.**
  
The managed-identity get command prints information about one
managed identity.
  
```shell
tharsis managed-identity get trn:managed_identity:<group_path>/<managed_identity_name>
```
  
#### Options
  
#### json

Show final output as JSON.\
**Default:** `false`


---
### managed-identity update subcommand
**Update a managed identity.**
  
The managed-identity update command updates a managed identity.
Currently, it supports updating the description and data.
Shows final output as JSON, if specified.
  
```shell
tharsis managed-identity update \
  -description "Updated AWS production role" \
  -aws-federated-role arn:aws:iam::123456789012:role/UpdatedRole \
  trn:managed_identity:<group_path>/<managed_identity_name>
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

#### json

Show final output as JSON.\
**Default:** `false`

#### kubernetes-federated-audience

Kubernetes federated audience. The audience should match the client_id configured in your EKS OIDC identity provider. (Only if type is kubernetes_federated)

#### tharsis-federated-service-account-path

Tharsis service account path this managed identity will assume. (Only if type is tharsis_federated)


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
  
The managed-identity-access-rule create command creates a new managed identity access rule.
  
```shell
tharsis managed-identity-access-rule create \
  -managed-identity-id trn:managed_identity:<group_path>/<managed_identity_name> \
  -rule-type eligible_principals \
  -run-stage plan \
  -allowed-user trn:user:<username> \
  -allowed-team trn:team:<team_name>
```
  
#### Options
  
#### allowed-service-account <span style={{color:'green'}}>...</span>

Allowed service account ID.

#### allowed-team <span style={{color:'green'}}>...</span>

Allowed team ID.

#### allowed-user <span style={{color:'green'}}>...</span>

Allowed user ID.

#### json

Show final output as JSON.\
**Default:** `false`

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
  
The managed-identity-access-rule delete command deletes a managed identity access rule.
  
```shell
tharsis managed-identity-access-rule delete <id>
```
  
---
### managed-identity-access-rule get subcommand
**Get a managed identity access rule.**
  
The managed-identity-access-rule get command gets a managed identity access rule by ID.
  
```shell
tharsis managed-identity-access-rule get <id>
```
  
#### Options
  
#### json

Show final output as JSON.\
**Default:** `false`


---
### managed-identity-access-rule list subcommand
**Retrieve a list of managed identity access rules.**
  
The managed-identity-access-rule list command prints information about
access rules for a specific managed identity.
  
```shell
tharsis managed-identity-access-rule list \
  -managed-identity-id trn:managed_identity:<group_path>/<managed_identity_name>
```
  
#### Options
  
#### json

Show final output as JSON.\
**Default:** `false`

#### managed-identity-id

ID of the managed identity to get access rules for.

#### managed-identity-path <span style={{color:'orange'}}>!</span>

Resource path of the managed identity to get access rules for.\
**Deprecated**: use -managed-identity-id


---
### managed-identity-access-rule update subcommand
**Update a managed identity access rule.**
  
The managed-identity-access-rule update command updates an existing managed identity access rule.
  
```shell
tharsis managed-identity-access-rule update \
  -allowed-user trn:user:<username> \
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

Show final output as JSON.\
**Default:** `false`

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
  
The managed-identity-alias create command creates a new managed identity alias.
  
```shell
tharsis managed-identity-alias create \
  -group-id trn:group:<group_path> \
  -alias-source-id trn:managed_identity:<group_path>/<source_identity_name> \
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

Show final output as JSON.\
**Default:** `false`

#### name <span style={{color:'orange'}}>!</span>

The name of the managed identity alias.\
**Deprecated**: pass name as an argument


---
### managed-identity-alias delete subcommand
**Delete a managed identity alias.**
  
The managed-identity-alias delete command deletes a managed identity alias.
  
```shell
tharsis managed-identity-alias delete trn:managed_identity:<group_path>/<managed_identity_name>
```
  
#### Options
  
#### force

Force delete the managed identity alias.


---
## mcp command
**Start the Tharsis MCP server.**
  
The mcp command starts the Tharsis MCP server, enabling AI assistants
to interact with Tharsis resources through the Model Context Protocol.
By default, all toolsets are enabled in read-only mode for safety.

Available toolsets: auth, run, job, configuration_version, workspace, group, variable, managed_identity, documentation, terraform_module, terraform_module_version, terraform_provider, terraform_provider_platform

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
  
```shell
# Start MCP server with production profile in read-only mode
tharsis -p production mcp

# Start with specific toolsets
tharsis mcp -toolsets auth,run

# Start with namespace ACL restrictions
tharsis mcp -namespace-mutation-acl "dev/*,staging/*"

# MCP Client Configuration (mcp.json):
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
  
The module create command creates a new Terraform module. It
requires a group ID and repository URL. The argument should be
in the format: module-name/system (e.g., vpc/aws). Shows final
output as JSON, if specified. Idempotent when used with
-if-not-exists option.
  
```shell
tharsis module create \
  -group-id trn:group:<group_path> \
  -repository-url https://github.com/example/terraform-aws-vpc \
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

Show final output as JSON.\
**Default:** `false`

#### private

Whether the module is private.\
**Default:** `false`

#### repository-url

The repository URL for the module.


---
### module create-attestation subcommand
**Create a new module attestation.**
  
The module create-attestation command creates a new module attestation.
  
```shell
tharsis module create-attestation \
  -description "Attestation for v1.0.0" \
  -data aGVsbG8sIHdvcmxk \
  trn:terraform_module:<module_path>
```
  
#### Options
  
#### data <span style={{color:'red'}}>*</span>

The attestation data (must be a Base64-encoded string).

#### description

Description for the attestation.

#### json

Show final output as JSON.\
**Default:** `false`


---
### module delete subcommand
**Delete a Terraform module.**
  
The module delete command deletes a Terraform module.
  
```shell
tharsis module delete trn:terraform_module:<group_path>/<module_name>/<system>
```
  
---
### module delete-attestation subcommand
**Delete a module attestation.**
  
The module delete-attestation command deletes a module attestation.
  
```shell
tharsis module delete-attestation trn:terraform_module_attestation:<group_path>/<module_name>/<module_system>/<sha_sum>
```
  
---
### module delete-version subcommand
**Delete a module version.**
  
The module delete-version command deletes a module version.
  
```shell
tharsis module delete-version trn:terraform_module_version:<group_path>/<module_name>/<system>/<semantic_version>
```
  
#### Options
  
#### version

Metadata version of the resource to be deleted. In most cases, this is not required.


---
### module get subcommand
**Get a single Terraform module.**
  
The module get command prints information about one Terraform module.
  
```shell
tharsis module get trn:terraform_module:<group_path>/<module_name>/<system>
```
  
#### Options
  
#### json

Show final output as JSON.\
**Default:** `false`


---
### module get-version subcommand
**Get a module version by ID or TRN.**
  
The module get-version command retrieves details about a specific module version.
  
```shell
tharsis module get-version trn:terraform_module_version:<group_path>/<module_name>/<system>/<version>
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`

#### version <span style={{color:'orange'}}>!</span>

A semver compliant version tag to use as a filter.\
**Deprecated**: pass version TRN as argument


---
### module list subcommand
**Retrieve a paginated list of modules.**
  
The module list command prints information about (likely
multiple) modules. Supports pagination, filtering and
sorting the output.
  
```shell
tharsis module list \
  -group-id trn:group:<group_path> \
  -include-inherited \
  -sort-by UPDATED_AT_DESC \
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

Show final output as JSON.\
**Default:** `false`

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
  
The module list-attestations command prints information about attestations
for a specific module. Supports pagination, filtering and sorting.
  
```shell
tharsis module list-attestations \
  -sort-by CREATED_AT_DESC \
  -limit 10 \
  trn:terraform_module:<group_path>/<module_name>/<system>
```
  
#### Options
  
#### cursor

The cursor string for manual pagination.

#### digest

Filter to attestations with this digest.

#### json

Show final output as JSON.\
**Default:** `false`

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
  
The module list-versions command prints information about versions
of a specific module. Supports pagination, filtering and sorting.
  
```shell
tharsis module list-versions \
  -search 1.0 \
  -sort-by CREATED_AT_DESC \
  -limit 10 \
  trn:terraform_module:<group_path>/<module_name>/<system>
```
  
#### Options
  
#### cursor

The cursor string for manual pagination.

#### json

Show final output as JSON.\
**Default:** `false`

#### latest

Filter to only the latest version.

#### limit

Maximum number of result elements to return.\
**Default:** `100`

#### search

Filter to versions containing this substring.

#### semantic-version

Filter to a specific semantic version.

#### sort-by

Sort by this field.

#### sort-order <span style={{color:'orange'}}>!</span>

Sort in this direction.\
**Values:** `ASC`, `DESC`\
**Deprecated**: use -sort-by


---
### module update subcommand
**Update a Terraform module.**
  
The module update command updates a Terraform module.
Currently, it supports updating the repository URL and
private flag. Shows final output as JSON, if specified.
  
```shell
tharsis module update \
  -repository-url https://github.com/example/terraform-aws-vpc-v2 \
  -private true \
  trn:terraform_module:<group_path>/<module_name>/<system>
```
  
#### Options
  
#### json

Show final output as JSON.\
**Default:** `false`

#### private

Whether the module is private.

#### repository-url

The repository URL for the module.

#### version

Metadata version of the resource to be updated. In most cases, this is not required.


---
### module update-attestation subcommand
**Update a module attestation.**
  
The module update-attestation command updates an existing module attestation.
  
```shell
tharsis module update-attestation \
  -description "Updated description" \
  trn:terraform_module_attestation:<group_path>/<module_name>/<system>/<sha_sum>
```
  
#### Options
  
#### description

Description for the attestation.

#### json

Show final output as JSON.\
**Default:** `false`


---
### module upload-version subcommand
**Upload a new module version to the module registry.**
  
The module upload-version command uploads a new
module version to the module registry.
  
```shell
tharsis module upload-version \
  -version 1.0.0 \
  -directory-path ./my-module \
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
  
The plan command creates a speculative plan. It allows viewing
the changes Terraform will make to your infrastructure
without applying them. Supports setting run-scoped
Terraform / environment variables and planning a
destroy run.

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
  
```shell
tharsis plan -directory-path ./terraform trn:workspace:<workspace_path>
```
  
#### Options
  
#### destroy

Designates this run as a destroy operation.\
**Default:** `false`

#### directory-path

The path of the root module's directory.

#### env-var <span style={{color:'green'}}>...</span>

An environment variable as a key=value pair.

#### env-var-file <span style={{color:'green'}}>...</span>

The path to an environment variables file.

#### module-source

Remote module source specification.

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
  
The run cancel command cancels a run. Supports forced cancellation which is useful when a graceful cancel is not enough.
  
```shell
tharsis run cancel -force <id>
```
  
#### Options
  
#### force

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
  
The runner-agent assign-service-account command assigns a service account to a runner agent.
  
```shell
tharsis runner-agent assign-service-account \
  trn:service_account:<group_path>/<service_account_name> \
  trn:runner:<group_path>/<runner_name>
```
  
---
### runner-agent create subcommand
**Create a new runner agent.**
  
The runner-agent create command creates a new runner agent.
  
```shell
tharsis runner-agent create \
  -group-id trn:group:<group_path> \
  -description "Production runner" \
  -run-untagged-jobs \
  -tag prod \
  -tag us-east-1 \
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

Show final output as JSON.\
**Default:** `false`

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
  
The runner-agent delete command deletes a runner agent.
  
```shell
tharsis runner-agent delete trn:runner:<group_path>/<runner_name>
```
  
#### Options
  
#### version

Metadata version of the resource to be deleted. In most cases, this is not required.


---
### runner-agent get subcommand
**Get a runner agent.**
  
The runner-agent get command gets a runner agent by ID.
  
```shell
tharsis runner-agent get trn:runner:<group_path>/<runner_name>
```
  
#### Options
  
#### json

Show final output as JSON.\
**Default:** `false`


---
### runner-agent unassign-service-account subcommand
**Unassign a service account from a runner agent.**
  
The runner-agent unassign-service-account command removes a service account from a runner agent.
  
```shell
tharsis runner-agent unassign-service-account \
  trn:service_account:<group_path>/<service_account_name> \
  trn:runner:<group_path>/<runner_name>
```
  
---
### runner-agent update subcommand
**Update a runner agent.**
  
The runner-agent update command updates an existing runner agent.
  
```shell
tharsis runner-agent update \
  -description "Updated description" \
  -disabled true \
  -tag prod \
  -tag us-west-2 \
  trn:runner:<group_path>/<runner_name>
```
  
#### Options
  
#### description

Description for the runner agent.

#### disabled

Enable or disable the runner agent.

#### json

Show final output as JSON.\
**Default:** `false`

#### run-untagged-jobs

Allow the runner agent to execute jobs without tags.

#### tag <span style={{color:'green'}}>...</span>

Tag for the runner agent.

#### version

Metadata version of the resource to be updated. In most cases, this is not required.


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
  
The service-account create-token command creates a token for a service account using OIDC authentication.
The input token is issued by an identity provider specified in the service account's trust policy.
The output token can be used to authenticate with the API.
  
```shell
tharsis service-account create-token \
  -token <oidc-token> \
  trn:service_account:<group_path>/<service_account_name>
```
  
#### Options
  
#### json

Output in JSON format.\
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
  
The login command starts an embedded web server and opens
a web browser page or tab pointed at said web server.
That redirects to the OAuth2 provider's login page, where
the user can sign in. If there is an SSO scheme active,
that will sign in the user. The login command captures
the authentication token for use in subsequent commands.
  
```shell
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
  
The terraform-provider create command creates a new terraform provider.
  
```shell
tharsis terraform-provider create \
  -group-id trn:group:<group_path> \
  -repository-url https://github.com/example/terraform-provider-example \
  my-provider
```
  
#### Options
  
#### group-id

The ID of the group to create the provider in.

#### json

Output in JSON format.\
**Default:** `false`

#### private

Set to false to allow all groups to view and use the terraform provider.\
**Default:** `true`

#### repository-url

The repository URL for this terraform provider.


---
### terraform-provider upload-version subcommand
**Upload a new Terraform provider version to the provider registry.**
  
The terraform-provider upload-version command uploads a new
Terraform provider version to the provider registry.
  
```shell
tharsis terraform-provider upload-version \
  -directory ./my-provider \
  trn:terraform_provider:<group_path>/<name>
```
  
#### Options
  
#### directory

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
  
The terraform-provider-mirror delete-platform command deletes a terraform provider
platform from a group's mirror. The package will no longer be available for the
associated provider's version and platform.
  
```shell
tharsis terraform-provider-mirror delete-platform <platform-mirror-id>
```
  
---
### terraform-provider-mirror delete-version subcommand
**Delete a terraform provider version from mirror.**
  
The terraform-provider-mirror delete-version command deletes a terraform provider
version and any associated platform binaries from a group's mirror. The -force
option must be used when deleting a provider version which actively hosts
platform binaries.
  
```shell
tharsis terraform-provider-mirror delete-version -force <version-mirror-id>
```
  
#### Options
  
#### force

Skip confirmation prompt.


---
### terraform-provider-mirror get-version subcommand
**Get a mirrored terraform provider version.**
  
The terraform-provider-mirror get-version command retrieves a terraform provider
version from the provider mirror.
  
```shell
tharsis terraform-provider-mirror get-version <version-mirror-id>
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`


---
### terraform-provider-mirror list-platforms subcommand
**Retrieve a paginated list of provider platform mirrors.**
  
The terraform-provider-mirror list-platforms command prints information
about provider platform mirrors for a version mirror. Supports pagination,
filtering and sorting.
  
```shell
tharsis terraform-provider-mirror list-platforms \
  -os linux \
  -architecture amd64 \
  -sort-by CREATED_AT_DESC \
  trn:terraform_provider_version_mirror:<group_path>/<provider_namespace>/<provider_name>/<semantic_version>
```
  
#### Options
  
#### architecture

Filter to platforms with this architecture.

#### cursor

The cursor string for manual pagination.

#### json

Show final output as JSON.\
**Default:** `false`

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
  
The terraform-provider-mirror list-versions command prints information
about provider version mirrors for a namespace. Supports pagination
and sorting.
  
```shell
tharsis terraform-provider-mirror list-versions \
  -sort-by CREATED_AT_DESC \
  -limit 10 \
  <namespace_path>
```
  
#### Options
  
#### cursor

The cursor string for manual pagination.

#### json

Show final output as JSON.\
**Default:** `false`

#### limit

Maximum number of result elements to return.\
**Default:** `100`

#### sort-by

Sort by this field.\
**Values:** `CREATED_AT_ASC`, `CREATED_AT_DESC`, `TYPE_ASC`, `TYPE_DESC`

#### sort-order <span style={{color:'orange'}}>!</span>

Sort in this direction.\
**Values:** `ASC`, `DESC`\
**Deprecated**: use -sort-by


---
### terraform-provider-mirror sync subcommand
**Sync provider platforms from upstream registry to mirror.**
  
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
1. CLI environment variable TF_TOKEN_\<hostname\>
(e.g., TF_TOKEN_registry_example_com)
2. Federated registry: runs service discovery and uses the
token from a matching CLI profile

Fully Qualified Name (FQN) must be formatted as:

\[registry hostname/\]\{registry namespace\}/\{provider name\}

The hostname can be omitted for providers from the default
public Terraform registry (registry.terraform.io).

Examples: registry.terraform.io/hashicorp/aws, hashicorp/aws
  
```shell
tharsis terraform-provider-mirror sync \
  -group-id my-group \
  -version 1.0.0 \
  -platform linux_amd64 \
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
## version command
**Get the CLI's version.**
  
The tharsis version command returns the CLI's version.
  
```shell
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
  
The workspace add-membership command adds a membership to a workspace.
Exactly one of -user-id, -service-account-id, or -team-id must be specified.
  
```shell
tharsis workspace add-membership \
  -role-id trn:role:owner \
  -user-id trn:user:john.smith \
  trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`

#### role <span style={{color:'orange'}}>!</span>

Role name for new membership.\
**Deprecated**: use -role-id

#### role-id

The role ID for the membership.

#### service-account-id

The service account ID for the membership.

#### team-id

The team ID for the membership.

#### team-name <span style={{color:'orange'}}>!</span>

Team name for the new membership.\
**Deprecated**: use -team-id

#### user-id

The user ID for the membership.

#### username <span style={{color:'orange'}}>!</span>

Username for the new membership.\
**Deprecated**: use -user-id


---
### workspace assign-managed-identity subcommand
**Assign a managed identity to a workspace.**
  
The workspace assign-managed-identity command assigns a managed identity to a workspace.
  
```shell
tharsis workspace assign-managed-identity \
  trn:workspace:<workspace_path> \
  trn:managed_identity:<group_path>/<identity_name>
```
  
---
### workspace create subcommand
**Create a new workspace.**
  
The workspace create command creates a new workspace. It
allows setting a workspace's description (optional),
maximum job duration and managed identities. Shows final
output as JSON, if specified. Idempotent when used with
-if-not-exists option.
  
```shell
tharsis workspace create \
  -parent-group-id trn:group:<group_path> \
  -description "Production workspace" \
  -terraform-version "1.5.0" \
  -max-job-duration 60 \
  -prevent-destroy-plan \
  -managed-identity trn:managed_identity:<group_path>/<identity_name> \
  -label env=prod \
  -label team=platform \
  <name>
```
  
#### Options
  
#### description

Description for the new workspace.

#### if-not-exists

Create a workspace if it does not already exist.\
**Default:** `false`

#### json

Show final output as JSON.\
**Default:** `false`

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
  
The workspace delete command deletes a workspace. Includes
a force flag to delete the workspace even if resources are
deployed (dangerous!).

Use with caution as deleting a workspace is irreversible!
  
```shell
tharsis workspace delete -force trn:workspace:<workspace_path>
```
  
#### Options
  
#### force

Force the workspace to delete even if resources are deployed.

#### version

Metadata version of the resource to be deleted. In most cases, this is not required.


---
### workspace delete-terraform-var subcommand
**Delete a terraform variable from a workspace.**
  
The workspace delete-terraform-var command deletes a terraform variable from a workspace.
  
```shell
tharsis workspace delete-terraform-var \
  -key region \
  trn:workspace:<workspace_path>
```
  
#### Options
  
#### key <span style={{color:'red'}}>*</span>

Variable key.

#### version

Metadata version of the resource to be deleted. In most cases, this is not required.


---
### workspace get subcommand
**Get a single workspace.**
  
The workspace get command prints information about one
workspace.
  
```shell
tharsis workspace get trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Show final output as JSON.\
**Default:** `false`


---
### workspace get-assigned-managed-identities subcommand
**Get assigned managed identities for a workspace.**
  
The workspace get-assigned-managed-identities command lists managed identities assigned to a workspace.
  
```shell
tharsis workspace get-assigned-managed-identities trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`


---
### workspace get-membership subcommand
**Get a workspace membership.**
  
The workspace get-membership command retrieves details about a specific workspace membership.
  
```shell
tharsis workspace get-membership \
  -user-id trn:user:<username> \
  trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`

#### service-account-id

Service account ID to find the workspace membership for.

#### team-id

Team ID to find the workspace membership for.

#### team-name <span style={{color:'orange'}}>!</span>

Team name to find the workspace membership for.\
**Deprecated**: use -team-id

#### user-id

User ID to find the workspace membership for.

#### username <span style={{color:'orange'}}>!</span>

Username to find the workspace membership for.\
**Deprecated**: use -user-id


---
### workspace get-terraform-var subcommand
**Get a terraform variable for a workspace.**
  
The workspace get-terraform-var command retrieves a terraform variable for a workspace.
  
```shell
tharsis workspace get-terraform-var \
  -key region \
  trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`

#### key <span style={{color:'red'}}>*</span>

Variable key.

#### show-sensitive

Show the actual value of sensitive variables (requires appropriate permissions).\
**Default:** `false`


---
### workspace label subcommand
**Manage labels on a workspace.**
  
The workspace label command manages labels on a workspace.
It supports adding, updating, removing, and overwriting labels.

Label operations:
key=value  Add or update a label
key-       Remove a label (not allowed with -overwrite)
  
```shell
tharsis workspace label \
  -overwrite \
  trn:workspace:<workspace_path> \
  env=prod \
  tier=frontend
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`

#### overwrite

Replace all existing labels with the specified labels.\
**Default:** `false`


---
### workspace list subcommand
**Retrieve a paginated list of workspaces.**
  
The workspace list command prints information about (likely
multiple) workspaces. Supports pagination, filtering and
sorting the output.
  
```shell
tharsis workspace list \
  -group-id trn:group:<group_path> \
  -label env=prod \
  -label team=platform \
  -sort-by FULL_PATH_ASC \
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

Show final output as JSON.\
**Default:** `false`

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
  
The workspace list-environment-vars command retrieves all environment
variables from a workspace and its parent groups.
  
```shell
tharsis workspace list-environment-vars -show-sensitive trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`

#### show-sensitive

Show the actual values of sensitive variables (requires appropriate permissions).\
**Default:** `false`


---
### workspace list-memberships subcommand
**Retrieve a list of workspace memberships.**
  
The workspace list-memberships command prints information about
memberships for a specific workspace.
  
```shell
tharsis workspace list-memberships trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Show final output as JSON.\
**Default:** `false`


---
### workspace list-terraform-vars subcommand
**List all terraform variables in a workspace.**
  
The workspace list-terraform-vars command retrieves all terraform
variables from a workspace and its parent groups.
  
```shell
tharsis workspace list-terraform-vars -show-sensitive trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`

#### show-sensitive

Show the actual values of sensitive variables (requires appropriate permissions).\
**Default:** `false`


---
### workspace migrate subcommand
**Migrate a workspace to a new group.**
  
The workspace migrate command migrates a workspace to a different group.
  
```shell
tharsis workspace migrate \
  -new-group-id trn:group:<group_path> \
  trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`

#### new-group-id <span style={{color:'red'}}>*</span>

New parent group ID.


---
### workspace outputs subcommand
**Get the state version outputs for a workspace.**
  
The workspace outputs command retrieves the state version outputs for a workspace.

Supported output types:
- Decorated (shows if map, list, etc. default).
- JSON.
- Raw (just the value. limited).

In addition, it supports filtering the output for each of the supported types above with -output-name option.

Combining -raw and -json is not allowed.
  
```shell
tharsis workspace outputs trn:workspace:<workspace_path>
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`

#### output-name

The name of the output variable to use as a filter. Required for -raw option.

#### raw

For any value that can be converted to a string, output just the raw value.\
**Default:** `false`


---
### workspace remove-membership subcommand
**Remove a workspace membership.**
  
The workspace remove-membership command removes a membership from a workspace.
  
```shell
tharsis workspace remove-membership <id>
```
  
#### Options
  
#### version

Metadata version of the resource to be deleted. In most cases, this is not required.


---
### workspace set-environment-vars subcommand
**Set environment variables for a workspace.**
  
The workspace set-environment-vars command sets environment variables for a workspace.
Command will overwrite any existing environment variables in the target workspace!
Note: This command does not support sensitive variables.
  
```shell
tharsis workspace set-environment-vars \
  -env-var-file vars.env \
  trn:workspace:<workspace_path>
```
  
#### Options
  
#### env-var-file <span style={{color:'red'}}>*</span> <span style={{color:'green'}}>...</span>

Path to an environment variables file.


---
### workspace set-terraform-var subcommand
**Set a terraform variable for a workspace.**
  
The workspace set-terraform-var command creates or updates a terraform variable for a workspace.
  
```shell
tharsis workspace set-terraform-var \
  -key region \
  -value us-east-1 \
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
  
The workspace set-terraform-vars command sets terraform variables for a workspace.
Command will overwrite any existing Terraform variables in the target workspace!
Note: This command does not support sensitive variables.
  
```shell
tharsis workspace set-terraform-vars \
  -tf-var-file terraform.tfvars \
  trn:workspace:<workspace_path>
```
  
#### Options
  
#### tf-var-file <span style={{color:'red'}}>*</span> <span style={{color:'green'}}>...</span>

Path to a .tfvars file.


---
### workspace unassign-managed-identity subcommand
**Unassign a managed identity from a workspace.**
  
The workspace unassign-managed-identity command removes a managed identity from a workspace.
  
```shell
tharsis workspace unassign-managed-identity \
  trn:workspace:<workspace_path> \
  trn:managed_identity:<group_path>/<identity_name>
```
  
---
### workspace update subcommand
**Update a workspace.**
  
The workspace update command updates a workspace.
Currently, it supports updating the description and the
maximum job duration. Shows final output as JSON, if
specified.
  
```shell
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

Show final output as JSON.\
**Default:** `false`

#### label <span style={{color:'green'}}>...</span>

Labels for the workspace (key=value).

#### max-job-duration

The amount of minutes before a job is gracefully canceled.

#### prevent-destroy-plan

Whether a run/plan will be prevented from destroying deployed resources.

#### terraform-version

The default Terraform CLI version for the workspace.

#### version

Metadata version of the resource to be updated. In most cases, this is not required.


---
### workspace update-membership subcommand
**Update a workspace membership.**
  
The workspace update-membership command updates a workspace membership's role.
  
```shell
tharsis workspace update-membership \
  -role-id trn:role:<role_name> \
  <id>
```
  
#### Options
  
#### json

Output in JSON format.\
**Default:** `false`

#### role <span style={{color:'orange'}}>!</span>

Role name for the membership.\
**Deprecated**: use -role-id

#### role-id

The role ID for the membership.

#### version

Metadata version of the resource to be updated. In most cases, this is not required.


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