# Tharsis CLI Commands

## Available Commands

Currently, the following commands are available:

```bash
apply                                       Apply a Terraform run
configure                                   Create or update a profile.
destroy                                     Destroy workspace resources
documentation                               Perform command documentation operations.
group                                       Do operations on groups.
managed-identity                            Do operations on a managed identity.
managed-identity-access-rule                Do operations on a managed identity access rule.
managed-identity-alias                      Do operations on a managed identity alias.
module                                      Do operations on a terraform module.
plan                                        Create a speculative plan
run                                         Do operations on runs.
runner-agent                                Do operations on runner agents.
service-account                             Create an authentication token for a service account.
sso                                         Log in to the OAuth2 provider and return an authentication token.
terraform-provider                          Do operations on a terraform provider.
terraform-provider-mirror                   Mirror Terraform providers from any Terraform registry.
version                                     Get the CLI's version.
workspace                                   Do operations on workspaces.
```

---

## apply

Apply a Terraform run

```bash
tharsis [global options] apply [options] <workspace-id>
```

The apply command creates and applies a Terraform run.
It first creates a plan, then applies it after approval.
Supports setting run-scoped Terraform / environment variables.

Terraform variables may be passed in via supported
options or from the environment with a 'TF*VAR*' prefix.

<details>
<summary>Options</summary>

- `--auto-approve` - Skip interactive approval of the plan.

- `--comment` - Comment for the apply.

- `--directory-path` - The path of the root module's directory.

- `--env-var` - An environment variable as a key=value pair.

- `--env-var-file` - The path to an environment variables file.

- `--input` - Ask for input for variables if not directly set.

- `--module-source` - Remote module source specification.

- `--module-version` - Remote module version number--defaults to latest.

- `--refresh` - Whether to do the usual refresh step.

- `--refresh-only` - Whether to do ONLY a refresh operation.

- `--target` - The Terraform address of the resources to be acted upon.

- `--terraform-version` - The Terraform CLI version to use for the run.

- `--tf-var` - A terraform variable as a key=value pair.

- `--tf-var-file` - The path to a .tfvars variables file.

</details>

:::note Example

```bash
tharsis apply --directory-path ./terraform trn:workspace:ops/my-workspace
```

:::

---

## configure

Create or update a profile.

:::info Subcommands

- `delete                                   ` - Remove a profile.
- `list                                     ` - Show all profiles.
  :::

```bash
tharsis configure [options]
```

The configure command creates or updates a profile. If no
options are specified, the command prompts for values.

<details>
<summary>Options</summary>

- `--http-endpoint` - The Tharsis HTTP API endpoint (in URL format).

- `--insecure-tls-skip-verify` - Allow TLS but disable verification of the gRPC server's certificate chain and hostname. This should ONLY be true for testing as it could allow the CLI to connect to an impersonated server.

- `--profile` - The name of the profile to set.

</details>

:::note Example

```bash
tharsis configure \
  --http-endpoint https://api.tharsis.example.com \
  --profile prod-example
```

:::

---

#### configure delete

Remove a profile.

```bash
tharsis configure delete <name>
```

The configure delete command removes a profile and its
credentials with the given name.

:::note Example

```bash
tharsis configure delete prod-example
```

:::

---

#### configure list

Show all profiles.

```bash
tharsis configure list
```

The configure list command prints information about all profiles.

:::note Example

```bash
tharsis configure list
```

:::

---

## destroy

Destroy workspace resources

```bash
tharsis [global options] destroy [options] <workspace-id>
```

The destroy command destroys resources in a workspace.
It creates a destroy plan, then applies it after approval.
Supports setting run-scoped Terraform / environment variables.

Terraform variables may be passed in via supported
options or from the environment with a 'TF*VAR*' prefix.

<details>
<summary>Options</summary>

- `--auto-approve` - Skip interactive approval of the plan.

- `--comment` - Comment for the destroy.

- `--directory-path` - The path of the root module's directory.

- `--env-var` - An environment variable as a key=value pair.

- `--env-var-file` - The path to an environment variables file.

- `--input` - Ask for input for variables if not directly set.

- `--module-source` - Remote module source specification.

- `--module-version` - Remote module version number--defaults to latest.

- `--refresh` - Whether to do the usual refresh step.

- `--target` - The Terraform address of the resources to be acted upon.

- `--terraform-version` - The Terraform CLI version to use for the run.

- `--tf-var` - A terraform variable as a key=value pair.

- `--tf-var-file` - The path to a .tfvars variables file.

</details>

:::note Example

```bash
tharsis destroy --directory-path ./terraform trn:workspace:ops/my-workspace
```

:::

---

## documentation

Perform command documentation operations.

:::info Subcommands

- `generate                                 ` - Generate documentation of commands.
  :::

The documentation command(s) perform operations on the documentation.

---

#### documentation generate

Generate documentation of commands.

```bash
tharsis [global options] documentation generate
```

The documentation generate command generates markdown documentation
for the entire CLI.

<details>
<summary>Options</summary>

- `--output` - The output filename.

</details>

:::note Example

```bash
tharsis documentation generate
```

:::

---

## group

Do operations on groups.

:::info Subcommands

- `create                                   ` - Create a new group.
- `delete                                   ` - Delete a group.
- `delete-terraform-var                     ` - Delete a terraform variable from a group.
- `get                                      ` - Get a single group.
- `get-terraform-var                        ` - Get a terraform variable for a group.
- `list                                     ` - Retrieve a paginated list of groups.
- `list-environment-vars                    ` - List all environment variables in a group.
- `list-memberships                         ` - Retrieve a list of group memberships.
- `list-terraform-vars                      ` - List all terraform variables in a group.
- `set-environment-vars                     ` - Set environment variables for a group.
- `set-terraform-var                        ` - Set a terraform variable for a group.
- `set-terraform-vars                       ` - Set terraform variables for a group.
- `update                                   ` - Update a group.
  :::

Groups are containers for organizing workspaces hierarchically.
They can be nested and inherit variables and managed identities
to children. Use group commands to create, update, delete groups,
set Terraform and environment variables, manage memberships, and
migrate groups between parents.

---

#### group create

Create a new group.

```bash
tharsis [global options] group create [options] <name>
```

The group create command creates a new group. It allows
setting a group's description (optional). Shows final
output as JSON, if specified. Idempotent when used with
--if-not-exists option.

<details>
<summary>Options</summary>

- `--description` - Description for the new group.

- `--if-not-exists` - Create a group if it does not already exist.

- `--json` - Show final output as JSON.

- `--parent-group-id` - Parent group ID.

</details>

:::note Example

```bash
tharsis group create \
  --parent-group-id trn:group:ops \
  --description "Operations group" \
  my-group
```

:::

---

#### group delete

Delete a group.

```bash
tharsis [global options] group delete [options] <id>
```

The group delete command deletes a group by its ID. Includes
a force flag to delete the group even if resources are
deployed (dangerous!).

<details>
<summary>Options</summary>

- `--force` - Force delete the group.

- `--version` - Metadata version of the resource to be deleted. In most cases, this is not required.

</details>

:::note Example

```bash
tharsis group delete \
  --force \
  trn:group:ops/my-group
```

:::

---

#### group delete-terraform-var

Delete a terraform variable from a group.

```bash
tharsis [global options] group delete-terraform-var [options] <group-id>
```

The group delete-terraform-var command deletes a terraform variable from a group.

<details>
<summary>Options</summary>

- `--key` - Variable key.

- `--version` - Metadata version of the resource to be deleted. In most cases, this is not required.

</details>

:::note Example

```bash
tharsis group delete-terraform-var \
  --key region \
  trn:group:ops/my-group
```

:::

---

#### group get

Get a single group.

```bash
tharsis [global options] group get [options] <id>
```

The group get command retrieves a single group by its ID.
Shows output as JSON, if specified.

<details>
<summary>Options</summary>

- `--json` - Show output as JSON.

</details>

:::note Example

```bash
tharsis group get \
  --json \
  trn:tharsis:group:ops/my-group
```

:::

---

#### group get-terraform-var

Get a terraform variable for a group.

```bash
tharsis [global options] group get-terraform-var [options] <group-id>
```

The group get-terraform-var command retrieves a terraform variable for a group.

<details>
<summary>Options</summary>

- `--json` - Output in JSON format.

- `--key` - Variable key.

- `--show-sensitive` - Show the actual value of sensitive variables (requires appropriate permissions).

</details>

:::note Example

```bash
tharsis group get-terraform-var \
  --key region \
  trn:group:ops/my-group
```

:::

---

#### group list

Retrieve a paginated list of groups.

```bash
tharsis [global options] group list [options]
```

The group list command prints information about (likely
multiple) groups. Supports pagination, filtering and
sorting the output.

<details>
<summary>Options</summary>

- `--cursor` - The cursor string for manual pagination.

- `--json` - Show final output as JSON.

- `--limit` - Maximum number of result elements to return. Defaults to 100.

- `--parent-id` - Filter to only direct sub-groups of this parent group.

- `--search` - Filter to only groups containing this substring in their path.

- `--sort-by` - Sort by this field (e.g., UPDATED_AT_ASC, UPDATED_AT_DESC, FULL_PATH_ASC, FULL_PATH_DESC).

</details>

:::note Example

```bash
tharsis group list \
  --parent-id trn:group:top-level/bottom-level \
  --sort-by FULL_PATH_ASC \
  --limit 5 \
  --json
```

:::

---

#### group list-environment-vars

List all environment variables in a group.

```bash
tharsis [global options] group list-environment-vars [options] <group-id>
```

The group list-environment-vars command retrieves all terraform
variables from a group and its parent groups.

<details>
<summary>Options</summary>

- `--json` - Output in JSON format.

- `--show-sensitive` - Show the actual values of sensitive variables (requires appropriate permissions).

</details>

:::note Example

```bash
tharsis group list-environment-vars --show-sensitive trn:group:ops/my-group
```

:::

---

#### group list-memberships

Retrieve a list of group memberships.

```bash
tharsis [global options] group list-memberships [options] <group-path>
```

The group list-memberships command prints information about
memberships for a specific group.

<details>
<summary>Options</summary>

- `--json` - Show final output as JSON.

</details>

:::note Example

```bash
tharsis group list-memberships top-level/my-group
```

:::

---

#### group list-terraform-vars

List all terraform variables in a group.

```bash
tharsis [global options] group list-terraform-vars [options] <group-id>
```

The group list-terraform-vars command retrieves all terraform
variables from a group and its parent groups.

<details>
<summary>Options</summary>

- `--json` - Output in JSON format.

- `--show-sensitive` - Show the actual values of sensitive variables (requires appropriate permissions).

</details>

:::note Example

```bash
tharsis group list-terraform-vars --show-sensitive trn:group:ops/my-group
```

:::

---

#### group set-environment-vars

Set environment variables for a group.

```bash
tharsis [global options] group set-environment-vars [options] <group-id>
```

The group set-environment-vars command sets environment variables for a group.
Command will overwrite any existing environment variables in the target group!
Note: This command does not support sensitive variables.

<details>
<summary>Options</summary>

- `--env-var-file` - Path to an environment variables file (can be specified multiple times).

</details>

:::note Example

```bash
tharsis group set-environment-vars \
  --env-var-file vars.env \
  trn:group:ops/my-group
```

:::

---

#### group set-terraform-var

Set a terraform variable for a group.

```bash
tharsis [global options] group set-terraform-var [options] <group-id>
```

The group set-terraform-var command creates or updates a terraform variable for a group.

<details>
<summary>Options</summary>

- `--key` - Variable key.

- `--sensitive` - Mark variable as sensitive.

- `--value` - Variable value.

</details>

:::note Example

```bash
tharsis group set-terraform-var \
  --key region \
  --value us-east-1 \
  trn:group:ops/my-group
```

:::

---

#### group set-terraform-vars

Set terraform variables for a group.

```bash
tharsis [global options] group set-terraform-vars [options] <group-id>
```

The group set-terraform-vars command sets terraform variables for a group.
Command will overwrite any existing Terraform variables in the target group!
Note: This command does not support sensitive variables.

<details>
<summary>Options</summary>

- `--tf-var-file` - Path to a .tfvars file (can be specified multiple times).

</details>

:::note Example

```bash
tharsis group set-terraform-vars \
  --tf-var-file terraform.tfvars \
  trn:group:ops/my-group
```

:::

---

#### group update

Update a group.

```bash
tharsis [global options] group update [options] <id>
```

The group update command updates a group. Currently, it
supports updating the description. Shows final output
as JSON, if specified.

<details>
<summary>Options</summary>

- `--description` - Description for the group.

- `--json` - Show final output as JSON.

- `--version` - Metadata version of the resource to be updated. In most cases, this is not required.

</details>

:::note Example

```bash
tharsis group update \
  --description "Updated operations group" \
  trn:group:ops/my-group
```

:::

---

## managed-identity

Do operations on a managed identity.

:::info Subcommands

- `create                                   ` - Create a new managed identity.
- `delete                                   ` - Delete a managed identity.
- `get                                      ` - Get a single managed identity.
- `update                                   ` - Update a managed identity.
  :::

Managed identities provide OIDC-federated credentials for cloud
providers (AWS, Azure, Kubernetes) without storing secrets. Use
managed-identity commands to create, update, delete, and get
managed identities.

---

#### managed-identity create

Create a new managed identity.

```bash
tharsis [global options] managed-identity create [options] <name>
```

The managed-identity create command creates a new managed identity.

<details>
<summary>Options</summary>

- `--aws-federated-role` - AWS IAM role. (Only if type is aws_federated)

- `--azure-federated-client-id` - Azure client ID. (Only if type is azure_federated)

- `--azure-federated-tenant-id` - Azure tenant ID. (Only if type is azure_federated)

- `--description` - Description for the managed identity.

- `--group-id` - Group ID or TRN where the managed identity will be created.

- `--json` - Show final output as JSON.

- `--tharsis-federated-service-account-id` - Tharsis service account ID or TRN. (Only if type is tharsis_federated)

- `--type` - The type of managed identity: aws_federated, azure_federated, tharsis_federated.

</details>

:::note Example

```bash
tharsis managed-identity create \
  --group-id trn:group:ops/my-group \
  --type aws_federated \
  --aws-federated-role arn:aws:iam::123456789012:role/MyRole \
  --description "AWS production role" \
  aws-prod
```

:::

---

#### managed-identity delete

Delete a managed identity.

```bash
tharsis [global options] managed-identity delete [options] <id>
```

The managed-identity delete command deletes a managed identity.

Use with caution as deleting a managed identity is irreversible!

<details>
<summary>Options</summary>

- `--force` - Force delete the managed identity.

</details>

:::note Example

```bash
tharsis managed-identity delete --force trn:managed_identity:ops/my-group/aws-prod
```

:::

---

#### managed-identity get

Get a single managed identity.

```bash
tharsis [global options] managed-identity get [options] <id>
```

The managed-identity get command prints information about one
managed identity.

<details>
<summary>Options</summary>

- `--json` - Show final output as JSON.

</details>

:::note Example

```bash
tharsis managed-identity get trn:managed_identity:ops/my-group/aws-prod
```

:::

---

#### managed-identity update

Update a managed identity.

```bash
tharsis [global options] managed-identity update [options] <id>
```

The managed-identity update command updates a managed identity.
Currently, it supports updating the description and data.
Shows final output as JSON, if specified.

<details>
<summary>Options</summary>

- `--aws-federated-role` - AWS IAM role. (Only if type is aws_federated)

- `--azure-federated-client-id` - Azure client ID. (Only if type is azure_federated)

- `--azure-federated-tenant-id` - Azure tenant ID. (Only if type is azure_federated)

- `--description` - Description for the managed identity.

- `--json` - Show final output as JSON.

- `--tharsis-federated-service-account-id` - Tharsis service account ID or TRN. (Only if type is tharsis_federated)

</details>

:::note Example

```bash
tharsis managed-identity update \
  --description "Updated AWS production role" \
  --aws-federated-role arn:aws:iam::123456789012:role/UpdatedRole \
  trn:managed_identity:ops/my-group/aws-prod
```

:::

---

## managed-identity-access-rule

Do operations on a managed identity access rule.

:::info Subcommands

- `create                                   ` - Create a new managed identity access rule.
- `delete                                   ` - Delete a managed identity access rule.
- `get                                      ` - Get a managed identity access rule.
- `list                                     ` - Retrieve a list of managed identity access rules.
- `update                                   ` - Update a managed identity access rule.
  :::

Access rules control which runs can use a managed identity based
on conditions like module source or workspace path. Use these
commands to create, update, delete, list, and get access rules.

---

#### managed-identity-access-rule create

Create a new managed identity access rule.

```bash
tharsis [global options] managed-identity-access-rule create [options]
```

The managed-identity-access-rule create command creates a new managed identity access rule.

<details>
<summary>Options</summary>

- `--allowed-service-account` - Allowed service account ID. (This flag may be repeated)

- `--allowed-team` - Allowed team ID. (This flag may be repeated)

- `--allowed-user` - Allowed user ID. (This flag may be repeated)

- `--json` - Show final output as JSON.

- `--managed-identity-id` - The ID or TRN of the managed identity.

- `--module-attestation-policy` - Module attestation policy in format "[PredicateType=someval,]PublicKeyFile=/path/to/file". (This flag may be repeated)

- `--run-stage` - The run stage: plan or apply.

- `--type` - The type of access rule: eligible_principals or module_attestation.

- `--verify-state-lineage` - Verify state lineage.

</details>

:::note Example

```bash
tharsis managed-identity-access-rule create \
  --managed-identity-id trn:managed_identity:ops/my-identity \
  --type eligible_principals \
  --run-stage plan \
  --allowed-user trn:user:john.smith \
  --allowed-team trn:team:my-team
```

:::

---

#### managed-identity-access-rule delete

Delete a managed identity access rule.

```bash
tharsis [global options] managed-identity-access-rule delete [options] <id>
```

The managed-identity-access-rule delete command deletes a managed identity access rule.

:::note Example

```bash
tharsis managed-identity-access-rule delete TV80ZG...
```

:::

---

#### managed-identity-access-rule get

Get a managed identity access rule.

```bash
tharsis [global options] managed-identity-access-rule get [options] <id>
```

The managed-identity-access-rule get command gets a managed identity access rule by ID.

<details>
<summary>Options</summary>

- `--json` - Show final output as JSON.

</details>

:::note Example

```bash
tharsis managed-identity-access-rule get trn:managed_identity_access_rule:abc123
```

:::

---

#### managed-identity-access-rule list

Retrieve a list of managed identity access rules.

```bash
tharsis [global options] managed-identity-access-rule list [options] <managed-identity-id>
```

The managed-identity-access-rule list command prints information about
access rules for a specific managed identity.

<details>
<summary>Options</summary>

- `--json` - Show final output as JSON.

</details>

:::note Example

```bash
tharsis managed-identity-access-rule list \
  trn:managed_identity:ops/my-identity
```

:::

---

#### managed-identity-access-rule update

Update a managed identity access rule.

```bash
tharsis [global options] managed-identity-access-rule update [options] <id>
```

The managed-identity-access-rule update command updates an existing managed identity access rule.

<details>
<summary>Options</summary>

- `--allowed-service-account` - Allowed service account ID. (This flag may be repeated)

- `--allowed-team` - Allowed team ID. (This flag may be repeated)

- `--allowed-user` - Allowed user ID. (This flag may be repeated)

- `--json` - Show final output as JSON.

- `--module-attestation-policy` - Module attestation policy in format "[PredicateType=someval,]PublicKeyFile=/path/to/file". (This flag may be repeated)

- `--run-stage` - The run stage: plan or apply.

- `--verify-state-lineage` - Verify state lineage (true or false).

</details>

:::note Example

```bash
tharsis managed-identity-access-rule update \
  --run-stage apply \
  --allowed-user trn:user:john.smith \
  TV80ZG...
```

:::

---

## managed-identity-alias

Do operations on a managed identity alias.

:::info Subcommands

- `create                                   ` - Create a new managed identity alias.
- `delete                                   ` - Delete a managed identity alias.
  :::

Aliases allow referencing managed identities from other groups.
Use these commands to create and delete managed identity aliases.

---

#### managed-identity-alias create

Create a new managed identity alias.

```bash
tharsis [global options] managed-identity-alias create [options] <name>
```

The managed-identity-alias create command creates a new managed identity alias.

<details>
<summary>Options</summary>

- `--alias-source-id` - The ID or TRN of the source managed identity.

- `--group-id` - Group ID or TRN where the managed identity alias will be created.

- `--json` - Show final output as JSON.

</details>

:::note Example

```bash
tharsis managed-identity-alias create \
  --group-id trn:group:ops/my-group \
  --alias-source-id trn:managed_identity:source-identity \
  prod-identity-alias
```

:::

---

#### managed-identity-alias delete

Delete a managed identity alias.

```bash
tharsis [global options] managed-identity-alias delete [options] <id>
```

The managed-identity-alias delete command deletes a managed identity alias.

<details>
<summary>Options</summary>

- `--force` - Force delete the managed identity alias.

</details>

:::note Example

```bash
tharsis managed-identity-alias delete trn:managed_identity:ops/my-group/prod-identity-alias
```

:::

---

## module

Do operations on a terraform module.

:::info Subcommands

- `create                                   ` - Create a new Terraform module.
- `create-attestation                       ` - Create a new module attestation.
- `delete                                   ` - Delete a Terraform module.
- `delete-attestation                       ` - Delete a module attestation.
- `delete-version                           ` - Delete a module version.
- `get                                      ` - Get a single Terraform module.
- `get-version                              ` - Get a module version by ID or TRN.
- `list                                     ` - Retrieve a paginated list of modules.
- `list-attestations                        ` - Retrieve a paginated list of module attestations.
- `list-versions                            ` - Retrieve a paginated list of module versions.
- `update                                   ` - Update a Terraform module.
- `update-attestation                       ` - Update a module attestation.
  :::

The module registry stores Terraform modules with versioning and
attestation support. Use module commands to create, update, delete
modules, upload versions, manage attestations, and list modules
and versions.

---

#### module create

Create a new Terraform module.

```bash
tharsis [global options] module create [options] <module-name/system>
```

The module create command creates a new Terraform module. It
requires a group ID and repository URL. The argument should be
in the format: module-name/system (e.g., vpc/aws). Shows final
output as JSON, if specified. Idempotent when used with
--if-not-exists option.

<details>
<summary>Options</summary>

- `--group-id` - Parent group ID.

- `--if-not-exists` - Create a module if it does not already exist.

- `--json` - Show final output as JSON.

- `--private` - Whether the module is private.

- `--repository-url` - The repository URL for the module.

</details>

:::note Example

```bash
tharsis module create \
  --group-id trn:group:ops/my-group \
  --repository-url https://github.com/example/terraform-aws-vpc \
  --private \
  vpc/aws
```

:::

---

#### module create-attestation

Create a new module attestation.

```bash
tharsis [global options] module create-attestation [options] <module-id>
```

The module create-attestation command creates a new module attestation.

<details>
<summary>Options</summary>

- `--attestation-data` - The attestation data (must be a Base64-encoded string).

- `--description` - Description for the attestation.

- `--json` - Show final output as JSON.

</details>

:::note Example

```bash
tharsis module create-attestation \
  --description "Attestation for v1.0.0" \
  --attestation-data '{"key":"value"}' \
  trn:terraform_module:ops/installer/aws
```

:::

---

#### module delete

Delete a Terraform module.

```bash
tharsis [global options] module delete [options] <id>
```

The module delete command deletes a Terraform module.

Use with caution as deleting a module is irreversible!

:::note Example

```bash
tharsis module delete trn:terraform_module:ops/my-group/vpc
```

:::

---

#### module delete-attestation

Delete a module attestation.

```bash
tharsis [global options] module delete-attestation [options] <id>
```

The module delete-attestation command deletes a module attestation.

<details>
<summary>Options</summary>

- `--force` - Force delete the module attestation.

</details>

:::note Example

```bash
tharsis module delete-attestation trn:terraform_module_attestation:ops/installer/aws:VE1W
```

:::

---

#### module delete-version

Delete a module version.

```bash
tharsis [global options] module delete-version [options] <version-id>
```

The module delete-version command deletes a module version.

<details>
<summary>Options</summary>

- `--force` - Force deletion without confirmation.

- `--version` - Metadata version of the resource to be deleted. In most cases, this is not required.

</details>

:::note Example

```bash
tharsis module delete-version trn:terraform_module_version:ops/installer/aws/1.0.0
```

:::

---

#### module get

Get a single Terraform module.

```bash
tharsis [global options] module get [options] <id>
```

The module get command prints information about one
Terraform module.

<details>
<summary>Options</summary>

- `--json` - Show final output as JSON.

</details>

:::note Example

```bash
tharsis module get trn:terraform_module:ops/my-group/vpc
```

:::

---

#### module get-version

Get a module version by ID or TRN.

```bash
tharsis [global options] module get-version [options] <version-id>
```

The module get-version command retrieves details about a specific module version.

<details>
<summary>Options</summary>

- `--json` - Output in JSON format.

</details>

:::note Example

```bash
tharsis module get-version trn:terraform_module_version:ops/installer/aws/1.0.0
```

:::

---

#### module list

Retrieve a paginated list of modules.

```bash
tharsis [global options] module list [options]
```

The module list command prints information about (likely
multiple) modules. Supports pagination, filtering and
sorting the output.

<details>
<summary>Options</summary>

- `--cursor` - The cursor string for manual pagination.

- `--group-id` - Filter to only modules in this group.

- `--include-inherited` - Include modules inherited from parent groups.

- `--json` - Show final output as JSON.

- `--limit` - Maximum number of result elements to return. Defaults to 100.

- `--search` - Filter to only modules containing this substring in their path.

- `--sort-by` - Sort by this field (e.g., NAME_ASC, NAME_DESC, UPDATED_AT_ASC, UPDATED_AT_DESC).

</details>

:::note Example

```bash
tharsis module list \
  --group-id trn:group:top-level \
  --include-inherited \
  --sort-by UPDATED_AT_DESC \
  --limit 5 \
  --json
```

:::

---

#### module list-attestations

Retrieve a paginated list of module attestations.

```bash
tharsis [global options] module list-attestations [options] <module-id>
```

The module list-attestations command prints information about attestations
for a specific module. Supports pagination, filtering and sorting.

<details>
<summary>Options</summary>

- `--cursor` - The cursor string for manual pagination.

- `--digest` - Filter to attestations with this digest.

- `--json` - Show final output as JSON.

- `--limit` - Maximum number of result elements to return. Defaults to 100.

- `--sort-by` - Sort by this field (e.g., CREATED_AT_ASC, CREATED_AT_DESC).

</details>

:::note Example

```bash
tharsis module list-attestations \
  --sort-by CREATED_AT_DESC \
  --limit 10 \
  trn:terraform_module:ops/installer/aws
```

:::

---

#### module list-versions

Retrieve a paginated list of module versions.

```bash
tharsis [global options] module list-versions [options] <module-id>
```

The module list-versions command prints information about versions
of a specific module. Supports pagination, filtering and sorting.

<details>
<summary>Options</summary>

- `--cursor` - The cursor string for manual pagination.

- `--json` - Show final output as JSON.

- `--latest` - Filter to only the latest version.

- `--limit` - Maximum number of result elements to return. Defaults to 100.

- `--search` - Filter to versions containing this substring.

- `--semantic-version` - Filter to a specific semantic version.

- `--sort-by` - Sort by this field (e.g., CREATED_AT_ASC, CREATED_AT_DESC).

</details>

:::note Example

```bash
tharsis module list-versions \
  --search 1.0 \
  --sort-by CREATED_AT_DESC \
  --limit 10 \
  trn:terraform_module:ops/installer/aws
```

:::

---

#### module update

Update a Terraform module.

```bash
tharsis [global options] module update [options] <id>
```

The module update command updates a Terraform module.
Currently, it supports updating the repository URL and
private flag. Shows final output as JSON, if specified.

<details>
<summary>Options</summary>

- `--json` - Show final output as JSON.

- `--private` - Whether the module is private.

- `--repository-url` - The repository URL for the module.

- `--version` - Metadata version of the resource to be updated. In most cases, this is not required.

</details>

:::note Example

```bash
tharsis module update \
  --repository-url https://github.com/example/terraform-aws-vpc-v2 \
  --private true \
  trn:terraform_module:ops/my-group/vpc
```

:::

---

#### module update-attestation

Update a module attestation.

```bash
tharsis [global options] module update-attestation [options] <id>
```

The module update-attestation command updates an existing module attestation.

<details>
<summary>Options</summary>

- `--description` - Description for the attestation.

- `--json` - Show final output as JSON.

</details>

:::note Example

```bash
tharsis module update-attestation \
  --description "Updated description" \
  trn:terraform_module_attestation:ops/installer/aws:VE1W
```

:::

---

## plan

Create a speculative plan

```bash
tharsis [global options] plan [options] <workspace-id>
```

The plan command creates a speculative plan. It allows viewing
the changes Terraform will make to your infrastructure
without applying them. Supports setting run-scoped
Terraform / environment variables and planning a
destroy run.

Terraform variables may be passed in via supported
options or from the environment with a 'TF*VAR*'
prefix.

Variable parsing precedence: 1. Terraform variables from the environment. 2. terraform.tfvars file from module's directory, if present. 3. terraform.tfvars.json file from module's directory, if present. 4. _.auto.tfvars, _.auto.tfvars.json files from the module's directory, if present. 5. --tf-var-file option(s). 6. --tf-var option(s).

NOTE: If the same variable is assigned multiple values, the last value found will be used.

<details>
<summary>Options</summary>

- `--destroy` - Designates this run as a destroy operation.

- `--directory-path` - The path of the root module's directory.

- `--env-var` - An environment variable as a key=value pair.

- `--env-var-file` - The path to an environment variables file.

- `--module-source` - Remote module source specification.

- `--module-version` - Remote module version number--defaults to latest.

- `--refresh` - Whether to do the usual refresh step.

- `--refresh-only` - Whether to do ONLY a refresh operation.

- `--speculative` - Whether this is a speculative plan.

- `--target` - The Terraform address of the resources to be acted upon.

- `--terraform-version` - The Terraform CLI version to use for the run.

- `--tf-var` - A terraform variable as a key=value pair.

- `--tf-var-file` - The path to a .tfvars variables file.

</details>

:::note Example

```bash
tharsis plan --directory-path ./terraform trn:workspace:ops/my-workspace
```

:::

---

## run

Do operations on runs.

:::info Subcommands

- `cancel                                   ` - Cancel a run.
  :::

Runs are units of execution (plan or apply) that create, update,
or destroy infrastructure resources. Use run commands to cancel
runs gracefully or forcefully.

---

#### run cancel

Cancel a run.

```bash
tharsis [global options] run cancel [options] <run-id>
```

The run cancel command cancels a run. Supports forced cancellation which is useful when a graceful cancel is not enough.

<details>
<summary>Options</summary>

- `--force` - Force the run to cancel.

</details>

:::note Example

```bash
tharsis run cancel --force Ul9kOW
```

:::

---

## runner-agent

Do operations on runner agents.

:::info Subcommands

- `assign-service-account                   ` - Assign a service account to a runner agent.
- `create                                   ` - Create a new runner agent.
- `delete                                   ` - Delete a runner agent.
- `get                                      ` - Get a runner agent.
- `unassign-service-account                 ` - Unassign a service account from a runner agent.
- `update                                   ` - Update a runner agent.
  :::

Runner agents are distributed job executors responsible for
launching Terraform jobs that deploy infrastructure to the cloud.
Use runner-agent commands to create, update, delete, get agents,
and assign or unassign service accounts.

---

#### runner-agent assign-service-account

Assign a service account to a runner agent.

```bash
tharsis [global options] runner-agent assign-service-account [options]
```

The runner-agent assign-service-account command assigns a service account to a runner agent.

<details>
<summary>Options</summary>

- `--runner-id` - The ID of the runner agent.

- `--service-account-id` - The ID of the service account to assign.

</details>

:::note Example

```bash
tharsis runner-agent assign-service-account \
  --runner-id trn:runner:ops/my-runner \
  --service-account-id trn:service_account:ops/my-sa
```

:::

---

#### runner-agent create

Create a new runner agent.

```bash
tharsis [global options] runner-agent create [options] <name>
```

The runner-agent create command creates a new runner agent.

<details>
<summary>Options</summary>

- `--description` - Description for the runner agent.

- `--group-id` - Group ID or TRN where the runner agent will be created.

- `--json` - Show final output as JSON.

- `--run-untagged-jobs` - Allow the runner agent to execute jobs without tags.

- `--tag` - Tag for the runner agent. (This flag may be repeated)

</details>

:::note Example

```bash
tharsis runner-agent create \
  --group-id trn:group:ops/my-group \
  --description "Production runner" \
  --run-untagged-jobs \
  --tag prod \
  --tag us-east-1 \
  prod-runner
```

:::

---

#### runner-agent delete

Delete a runner agent.

```bash
tharsis [global options] runner-agent delete [options] <id>
```

The runner-agent delete command deletes a runner agent.

<details>
<summary>Options</summary>

- `--version` - Metadata version of the resource to be deleted. In most cases, this is not required.

</details>

:::note Example

```bash
tharsis runner-agent delete trn:runner:ops/prod-runner
```

:::

---

#### runner-agent get

Get a runner agent.

```bash
tharsis [global options] runner-agent get [options] <id>
```

The runner-agent get command gets a runner agent by ID.

<details>
<summary>Options</summary>

- `--json` - Show final output as JSON.

</details>

:::note Example

```bash
tharsis runner-agent get trn:runner:ops/prod-runner
```

:::

---

#### runner-agent unassign-service-account

Unassign a service account from a runner agent.

```bash
tharsis [global options] runner-agent unassign-service-account [options]
```

The runner-agent unassign-service-account command removes a service account from a runner agent.

<details>
<summary>Options</summary>

- `--runner-id` - The ID of the runner agent.

- `--service-account-id` - The ID of the service account to unassign.

</details>

:::note Example

```bash
tharsis runner-agent unassign-service-account \
  --runner-id trn:runner:ops/my-runner \
  --service-account-id trn:service_account:ops/my-sa
```

:::

---

#### runner-agent update

Update a runner agent.

```bash
tharsis [global options] runner-agent update [options] <id>
```

The runner-agent update command updates an existing runner agent.

<details>
<summary>Options</summary>

- `--description` - Description for the runner agent.

- `--disabled` - Enable or disable the runner agent (true or false).

- `--json` - Show final output as JSON.

- `--run-untagged-jobs` - Allow the runner agent to execute jobs without tags (true or false).

- `--tag` - Tag for the runner agent. (This flag may be repeated)

- `--version` - Metadata version of the resource to be updated. In most cases, this is not required.

</details>

:::note Example

```bash
tharsis runner-agent update \
  --description "Updated description" \
  --disabled true \
  --tag prod \
  --tag us-west-2 \
  trn:runner:abc123
```

:::

---

## service-account

Create an authentication token for a service account.

:::info Subcommands

- `create-oidc-token                        ` - Create a token for a service account using OIDC.
  :::

Service accounts provide machine-to-machine authentication for
CI/CD pipelines and automation. Use service-account commands to
create authentication tokens.

---

#### service-account create-oidc-token

Create a token for a service account using OIDC.

```bash
tharsis [global options] service-account create-oidc-token [options] <service-account-id>
```

The service-account create-oidc-token command creates a token for a service account using OIDC authentication.
The input token is issued by an identity provider specified in the service account's trust policy.
The output token can be used to authenticate with the API.

<details>
<summary>Options</summary>

- `--json` - Output in JSON format.

- `--token` - Initial authentication token from identity provider.

</details>

:::note Example

```bash
tharsis service-account create-oidc-token \
  --token <oidc-token> \
  trn:service_account:ops/my-sa
```

:::

---

## sso

Log in to the OAuth2 provider and return an authentication token.

:::info Subcommands

- `login                                    ` - Log in to the OAuth2 provider and return an authentication token.
  :::

The sso command authenticates the CLI with the OAuth2 provider,
and allows making authenticated calls to Tharsis backend.

---

#### sso login

Log in to the OAuth2 provider and return an authentication token.

```bash
tharsis [global options] sso login
```

The login command starts an embedded web server and opens
a web browser page or tab pointed at said web server.
That redirects to the OAuth2 provider's login page, where
the user can sign in. If there is an SSO scheme active,
that will sign in the user. The login command captures
the authentication token for use in subsequent commands.

:::note Example

```bash
tharsis sso login
```

:::

---

## terraform-provider

Do operations on a terraform provider.

:::info Subcommands

- `create                                   ` - Create a new terraform provider.
  :::

The provider registry stores Terraform providers with versioning
support. Use terraform-provider commands to create providers and
upload provider versions to the registry.

---

#### terraform-provider create

Create a new terraform provider.

```bash
tharsis [global options] terraform-provider create [options] <provider-name>
```

The terraform-provider create command creates a new terraform provider.

<details>
<summary>Options</summary>

- `--group-id` - The ID of the group to create the provider in.

- `--json` - Output in JSON format.

- `--private` - Set to false to allow all groups to view and use the terraform provider.

- `--repository-url` - The repository URL for this terraform provider.

</details>

:::note Example

```bash
tharsis terraform-provider create \
  --group-id trn:group:ops/my-group \
  --repository-url https://github.com/example/terraform-provider-example \
  my-provider
```

:::

---

## terraform-provider-mirror

Mirror Terraform providers from any Terraform registry.

:::info Subcommands

- `delete-platform                          ` - Delete a terraform provider platform from mirror.
- `delete-version                           ` - Delete a terraform provider version from mirror.
- `get-version                              ` - Get a mirrored terraform provider version.
- `list-platforms                           ` - Retrieve a paginated list of provider platform mirrors.
- `list-versions                            ` - Retrieve a paginated list of provider version mirrors.
  :::

The provider mirror caches Terraform providers from any registry
for use within a group hierarchy. It supports Terraform's Provider
Network Mirror Protocol and gives root group owners control over
which providers, platform packages, and registries are available.
Use these commands to sync providers, list versions and platforms,
get version details, and delete versions or platforms.

---

#### terraform-provider-mirror delete-platform

Delete a terraform provider platform from mirror.

```bash
tharsis [global options] terraform-provider-mirror delete-platform [options] <platform-mirror-id>
```

The terraform-provider-mirror delete-platform command deletes a terraform provider
platform from a group's mirror. The package will no longer be available for the
associated provider's version and platform.

:::note Example

```bash
tharsis terraform-provider-mirror delete-platform <platform-mirror-id>
```

:::

---

#### terraform-provider-mirror delete-version

Delete a terraform provider version from mirror.

```bash
tharsis [global options] terraform-provider-mirror delete-version [options] <version-mirror-id>
```

The terraform-provider-mirror delete-version command deletes a terraform provider
version and any associated platform binaries from a group's mirror.

<details>
<summary>Options</summary>

- `--force` - Skip confirmation prompt.

</details>

:::note Example

```bash
tharsis terraform-provider-mirror delete-version --force <version-mirror-id>
```

:::

---

#### terraform-provider-mirror get-version

Get a mirrored terraform provider version.

```bash
tharsis [global options] terraform-provider-mirror get-version [options] <version-mirror-id>
```

The terraform-provider-mirror get-version command retrieves a terraform provider
version from the provider mirror.

<details>
<summary>Options</summary>

- `--json` - Output in JSON format.

</details>

:::note Example

```bash
tharsis terraform-provider-mirror get-version <version-mirror-id>
```

:::

---

#### terraform-provider-mirror list-platforms

Retrieve a paginated list of provider platform mirrors.

```bash
tharsis [global options] terraform-provider-mirror list-platforms [options] <version-mirror-id>
```

The terraform-provider-mirror list-platforms command prints information
about provider platform mirrors for a version mirror. Supports pagination,
filtering and sorting.

<details>
<summary>Options</summary>

- `--architecture` - Filter to platforms with this architecture.

- `--cursor` - The cursor string for manual pagination.

- `--json` - Show final output as JSON.

- `--limit` - Maximum number of result elements to return. Defaults to 100.

- `--os` - Filter to platforms with this OS.

- `--sort-by` - Sort by this field (e.g., CREATED_AT_ASC, CREATED_AT_DESC).

</details>

:::note Example

```bash
tharsis terraform-provider-mirror list-platforms \
  --os linux \
  --architecture amd64 \
  --sort-by CREATED_AT_DESC \
  trn:terraform_provider_version_mirror:ops/registry.terraform.io/hashicorp/time/0.13.1
```

:::

---

#### terraform-provider-mirror list-versions

Retrieve a paginated list of provider version mirrors.

```bash
tharsis [global options] terraform-provider-mirror list-versions [options] <namespace-path>
```

The terraform-provider-mirror list-versions command prints information
about provider version mirrors in a namespace. Supports pagination and sorting.

<details>
<summary>Options</summary>

- `--cursor` - The cursor string for manual pagination.

- `--json` - Show final output as JSON.

- `--limit` - Maximum number of result elements to return. Defaults to 100.

- `--sort-by` - Sort by this field (e.g., CREATED_AT_ASC, CREATED_AT_DESC).

</details>

:::note Example

```bash
tharsis terraform-provider-mirror list-versions \
  --sort-by CREATED_AT_DESC \
  --limit 10 \
  ops
```

:::

---

## version

Get the CLI's version.

```bash
tharsis [global options] version
```

The tharsis version command returns the CLI's version.

<details>
<summary>Options</summary>

- `--json` - Show final output as JSON.

</details>

:::note Example

```bash
tharsis version --json
```

:::

---

## workspace

Do operations on workspaces.

:::info Subcommands

- `assign-managed-identity                  ` - Assign a managed identity to a workspace.
- `create                                   ` - Create a new workspace.
- `delete                                   ` - Delete a workspace.
- `delete-terraform-var                     ` - Delete a terraform variable from a workspace.
- `get                                      ` - Get a single workspace.
- `get-assigned-managed-identities          ` - Get assigned managed identities for a workspace.
- `get-terraform-var                        ` - Get a terraform variable for a workspace.
- `label                                    ` - Manage labels on a workspace.
- `list                                     ` - Retrieve a paginated list of workspaces.
- `list-environment-vars                    ` - List all environment variables in a workspace.
- `list-memberships                         ` - Retrieve a list of workspace memberships.
- `list-terraform-vars                      ` - List all terraform variables in a workspace.
- `outputs                                  ` - Get the state version outputs for a workspace.
- `set-environment-vars                     ` - Set environment variables for a workspace.
- `set-terraform-var                        ` - Set a terraform variable for a workspace.
- `set-terraform-vars                       ` - Set terraform variables for a workspace.
- `unassign-managed-identity                ` - Unassign a managed identity from a workspace.
- `update                                   ` - Update a workspace.
  :::

Workspaces contain Terraform deployments, state, runs, and variables.
Use workspace commands to create, update, delete workspaces, assign
and unassign managed identities, set Terraform and environment
variables, manage memberships, and view workspace outputs.

---

#### workspace assign-managed-identity

Assign a managed identity to a workspace.

```bash
tharsis [global options] workspace assign-managed-identity [options]
```

The workspace assign-managed-identity command assigns a managed identity to a workspace.

<details>
<summary>Options</summary>

- `--managed-identity-id` - The ID of the managed identity to assign.

- `--workspace-id` - The ID of the workspace.

</details>

:::note Example

```bash
tharsis workspace assign-managed-identity \
  --workspace-id trn:workspace:ops/my-workspace \
  --managed-identity-id trn:managed_identity:ops/my-identity
```

:::

---

#### workspace create

Create a new workspace.

```bash
tharsis [global options] workspace create [options] <name>
```

The workspace create command creates a new workspace. It
allows setting a workspace's description (optional),
maximum job duration and managed identity. Shows final
output as JSON, if specified. Idempotent when used with
--if-not-exists option.

<details>
<summary>Options</summary>

- `--description` - Description for the new workspace.

- `--if-not-exists` - Create a workspace if it does not already exist.

- `--json` - Show final output as JSON.

- `--label` - Labels for the new workspace (key=value). Can be specified multiple times.

- `--managed-identity` - The ID of a managed identity to assign.

- `--max-job-duration` - The amount of minutes before a job is gracefully canceled (Default 720).

- `--parent-group-id` - Parent group ID.

- `--prevent-destroy-plan` - Whether a run/plan will be prevented from destroying deployed resources.

- `--terraform-version` - The default Terraform CLI version for the new workspace.

</details>

:::note Example

```bash
tharsis workspace create \
  --parent-group-id trn:group:ops/my-group \
  --description "Production workspace" \
  --terraform-version "1.5.0" \
  --max-job-duration 60 \
  --prevent-destroy-plan \
  --managed-identity trn:managed_identity:ops/aws-prod \
  --label env=prod \
  --label team=platform \
  my-workspace
```

:::

---

#### workspace delete

Delete a workspace.

```bash
tharsis [global options] workspace delete [options] <id>
```

The workspace delete command deletes a workspace. Includes
a force flag to delete the workspace even if resources are
deployed (dangerous!).

Use with caution as deleting a workspace is irreversible!

<details>
<summary>Options</summary>

- `--force` - Force the workspace to delete even if resources are deployed.

- `--version` - Metadata version of the resource to be deleted. In most cases, this is not required.

</details>

:::note Example

```bash
tharsis workspace delete --force trn:workspace:ops/my-group/my-workspace
```

:::

---

#### workspace delete-terraform-var

Delete a terraform variable from a workspace.

```bash
tharsis [global options] workspace delete-terraform-var [options] <workspace-id>
```

The workspace delete-terraform-var command deletes a terraform variable from a workspace.

<details>
<summary>Options</summary>

- `--key` - Variable key.

- `--version` - Metadata version of the resource to be deleted. In most cases, this is not required.

</details>

:::note Example

```bash
tharsis workspace delete-terraform-var \
  --key region \
  trn:workspace:ops/my-workspace
```

:::

---

#### workspace get

Get a single workspace.

```bash
tharsis [global options] workspace get [options] <id>
```

The workspace get command prints information about one
workspace.

<details>
<summary>Options</summary>

- `--json` - Show final output as JSON.

</details>

:::note Example

```bash
tharsis workspace get trn:workspace:ops/my-group/my-workspace
```

:::

---

#### workspace get-assigned-managed-identities

Get assigned managed identities for a workspace.

```bash
tharsis [global options] workspace get-assigned-managed-identities [options] <workspace-id>
```

The workspace get-assigned-managed-identities command lists managed identities assigned to a workspace.

<details>
<summary>Options</summary>

- `--json` - Output in JSON format.

</details>

:::note Example

```bash
tharsis workspace get-assigned-managed-identities trn:workspace:ops/my-workspace
```

:::

---

#### workspace get-terraform-var

Get a terraform variable for a workspace.

```bash
tharsis [global options] workspace get-terraform-var [options] <workspace-id>
```

The workspace get-terraform-var command retrieves a terraform variable for a workspace.

<details>
<summary>Options</summary>

- `--json` - Output in JSON format.

- `--key` - Variable key.

- `--show-sensitive` - Show the actual value of sensitive variables (requires appropriate permissions).

</details>

:::note Example

```bash
tharsis workspace get-terraform-var \
  --key region \
  trn:workspace:ops/my-workspace
```

:::

---

#### workspace label

Manage labels on a workspace.

```bash
tharsis [global options] workspace label [options] <workspace-id>
```

The workspace label command manages labels on a workspace.
It supports adding, updating, removing, and overwriting labels.

Label operations:
key=value Add or update a label
key- Remove a label (not allowed with --overwrite)

<details>
<summary>Options</summary>

- `--json` - Output in JSON format.

- `--label` - Label operation (key=value to add/update, key- to remove). Can be specified multiple times.

- `--overwrite` - Replace all existing labels with the specified labels.

</details>

:::note Example

```bash
tharsis workspace label \
  --label env=prod \
  --label tier=frontend \
  trn:workspace:ops/my-workspace
```

:::

---

#### workspace list

Retrieve a paginated list of workspaces.

```bash
tharsis [global options] workspace list [options]
```

The workspace list command prints information about (likely
multiple) workspaces. Supports pagination, filtering and
sorting the output.

<details>
<summary>Options</summary>

- `--cursor` - The cursor string for manual pagination.

- `--group-id` - Filter to only workspaces in this group.

- `--json` - Show final output as JSON.

- `--label` - Filter by label (key=value). This flag may be repeated.

- `--limit` - Maximum number of result elements to return. Defaults to 100.

- `--search` - Filter to only workspaces containing this substring in their path.

- `--sort-by` - Sort by this field (e.g., UPDATED_AT_ASC, UPDATED_AT_DESC, FULL_PATH_ASC, FULL_PATH_DESC).

</details>

:::note Example

```bash
tharsis workspace list \
  --group-id trn:group:top-level \
  --label env=prod \
  --label team=platform \
  --sort-by FULL_PATH_ASC \
  --limit 5 \
  --json
```

:::

---

#### workspace list-environment-vars

List all environment variables in a workspace.

```bash
tharsis [global options] workspace list-environment-vars [options] <workspace-id>
```

The workspace list-environment-vars command retrieves all terraform
variables from a workspace and its parent workspaces.

<details>
<summary>Options</summary>

- `--json` - Output in JSON format.

- `--show-sensitive` - Show the actual values of sensitive variables (requires appropriate permissions).

</details>

:::note Example

```bash
tharsis workspace list-environment-vars --show-sensitive trn:workspace:ops/my-workspace
```

:::

---

#### workspace list-memberships

Retrieve a list of workspace memberships.

```bash
tharsis [global options] workspace list-memberships [options] <workspace-path>
```

The workspace list-memberships command prints information about
memberships for a specific workspace.

<details>
<summary>Options</summary>

- `--json` - Show final output as JSON.

</details>

:::note Example

```bash
tharsis workspace list-memberships top-level/my-workspace
```

:::

---

#### workspace list-terraform-vars

List all terraform variables in a workspace.

```bash
tharsis [global options] workspace list-terraform-vars [options] <workspace-id>
```

The workspace list-terraform-vars command retrieves all terraform
variables from a workspace and its parent workspaces.

<details>
<summary>Options</summary>

- `--json` - Output in JSON format.

- `--show-sensitive` - Show the actual values of sensitive variables (requires appropriate permissions).

</details>

:::note Example

```bash
tharsis workspace list-terraform-vars --show-sensitive trn:workspace:ops/my-workspace
```

:::

---

#### workspace outputs

Get the state version outputs for a workspace.

```bash
tharsis [global options] workspace outputs [options] <workspace-id>
```

The workspace outputs command retrieves the state version outputs for a workspace.

Supported output types: - Decorated (shows if map, list, etc. default). - JSON. - Raw (just the value. limited).

In addition, it supports filtering the output for each of the supported types above with --output-name option.

Combining --raw and --json is not allowed.

<details>
<summary>Options</summary>

- `--json` - Output in JSON format.

- `--output-name` - The name of the output variable to use as a filter. Required for -raw option.

- `--raw` - For any value that can be converted to a string, output just the raw value.

</details>

:::note Example

```bash
tharsis workspace outputs trn:workspace:ops/my-workspace
```

:::

---

#### workspace set-environment-vars

Set environment variables for a workspace.

```bash
tharsis [global options] workspace set-environment-vars [options] <workspace-id>
```

The workspace set-environment-vars command sets environment variables for a workspace.
Command will overwrite any existing environment variables in the target workspace!
Note: This command does not support sensitive variables.

<details>
<summary>Options</summary>

- `--env-var-file` - Path to an environment variables file (can be specified multiple times).

</details>

:::note Example

```bash
tharsis workspace set-environment-vars \
  --env-var-file vars.env \
  trn:workspace:ops/my-workspace
```

:::

---

#### workspace set-terraform-var

Set a terraform variable for a workspace.

```bash
tharsis [global options] workspace set-terraform-var [options] <workspace-id>
```

The workspace set-terraform-var command creates or updates a terraform variable for a workspace.

<details>
<summary>Options</summary>

- `--key` - Variable key.

- `--sensitive` - Mark variable as sensitive.

- `--value` - Variable value.

</details>

:::note Example

```bash
tharsis workspace set-terraform-var \
  --key region \
  --value us-east-1 \
  trn:workspace:ops/my-workspace
```

:::

---

#### workspace set-terraform-vars

Set terraform variables for a workspace.

```bash
tharsis [global options] workspace set-terraform-vars [options] <workspace-id>
```

The workspace set-terraform-vars command sets terraform variables for a workspace.
Command will overwrite any existing Terraform variables in the target workspace!
Note: This command does not support sensitive variables.

<details>
<summary>Options</summary>

- `--tf-var-file` - Path to a .tfvars file (can be specified multiple times).

</details>

:::note Example

```bash
tharsis workspace set-terraform-vars \
  --tf-var-file terraform.tfvars \
  trn:workspace:ops/my-workspace
```

:::

---

#### workspace unassign-managed-identity

Unassign a managed identity from a workspace.

```bash
tharsis [global options] workspace unassign-managed-identity [options]
```

The workspace unassign-managed-identity command removes a managed identity from a workspace.

<details>
<summary>Options</summary>

- `--managed-identity-id` - The ID of the managed identity to unassign.

- `--workspace-id` - The ID of the workspace.

</details>

:::note Example

```bash
tharsis workspace unassign-managed-identity \
  --workspace-id trn:workspace:ops/my-workspace \
  --managed-identity-id trn:managed_identity:ops/my-identity
```

:::

---

#### workspace update

Update a workspace.

```bash
tharsis [global options] workspace update [options] <id>
```

The workspace update command updates a workspace.
Currently, it supports updating the description and the
maximum job duration. Shows final output as JSON, if
specified.

<details>
<summary>Options</summary>

- `--description` - Description for the workspace.

- `--json` - Show final output as JSON.

- `--label` - Labels for the workspace (key=value). Can be specified multiple times.

- `--max-job-duration` - The amount of minutes before a job is gracefully canceled.

- `--prevent-destroy-plan` - Whether a run/plan will be prevented from destroying deployed resources.

- `--terraform-version` - The default Terraform CLI version for the workspace.

- `--version` - Metadata version of the resource to be updated. In most cases, this is not required.

</details>

:::note Example

```bash
tharsis workspace update \
  --description "Updated production workspace" \
  --terraform-version "1.6.0" \
  --max-job-duration 120 \
  --prevent-destroy-plan true \
  trn:workspace:ops/my-group/my-workspace
```

:::
