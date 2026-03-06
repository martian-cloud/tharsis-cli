# Tharsis CLI Commands

## Available Commands

Currently, the following commands are available:

```
apply                                      Apply a single run.
configure                                  Create or update a profile.
destroy                                    Destroy the workspace state.
documentation                              Perform command documentation operations.
group                                      Do operations on groups.
managed-identity                           Do operations on a managed identity.
managed-identity-access-rule               Do operations on a managed identity access rule.
managed-identity-alias                     Do operations on a managed identity alias.
module                                     Do operations on a terraform module.
plan                                       Create a speculative plan
run                                        Do operations on runs.
runner-agent                               Do operations on runner agents.
service-account                            Create an authentication token for a service account.
sso                                        Log in to the OAuth2 provider and return an authentication token.
terraform-provider                         Do operations on a terraform provider.
terraform-provider-mirror                  Mirror Terraform providers from any Terraform registry.
version                                    Get the CLI's version.
workspace                                  Do operations on workspaces.
```

***

### Command: apply

##### Apply a single run.

The apply command applies a run to create, update, or destroy
infrastructure resources. Supports setting run-scoped Terraform
and environment variables, auto-approving changes, using remote
module sources, and specifying Terraform versions.

***

### Command: configure

##### Create or update a profile.

**Subcommands:**

```
delete                                     Remove a profile.
list                                       Show all profiles.
```

```
Usage: tharsis configure [options]
```

   The configure command creates or updates a profile. If no
   options are specified, the command prompts for values.

<details>
<summary>Expand options</summary>

- `--http-endpoint`: The Tharsis HTTP API endpoint (in URL format).

- `--insecure-tls-skip-verify`: Allow TLS but disable verification of the gRPC server's certificate chain and hostname. This should ONLY be true for testing as it could allow the CLI to connect to an impersonated server.

- `--profile`: The name of the profile to set.

</details>

##### Example:

```
tharsis configure \
  --http-endpoint https://api.tharsis.example.com \
  --profile prod-example
```

---

#### configure delete

##### Remove a profile.

```
Usage: tharsis configure delete <name>
```

   The configure delete command removes a profile and its
   credentials with the given name.

##### Example:

```
tharsis configure delete prod-example
```

---

#### configure list

##### Show all profiles.

```
Usage: tharsis configure list
```

   The configure list command prints information about all profiles.

##### Example:

```
tharsis configure list
```

***

### Command: destroy

##### Destroy the workspace state.

The destroy command destroys all infrastructure resources managed
by a workspace. Similar to apply, it supports setting run-scoped
Terraform and environment variables, auto-approving changes, and
using remote module sources.

***

### Command: documentation

##### Perform command documentation operations.

**Subcommands:**

```
generate                                   Generate documentation of commands.
```

The documentation command(s) perform operations on the documentation.

---

#### documentation generate

##### Generate documentation of commands.

```
Usage: tharsis [global options] documentation generate
```

  The documentation generate command generates markdown documentation
  for the entire CLI.

<details>
<summary>Expand options</summary>

- `--output`: The output filename.

</details>

##### Example:

```
tharsis documentation generate
```

***

### Command: group

##### Do operations on groups.

**Subcommands:**

```
create                                     Create a new group.
delete                                     Delete a group.
get                                        Get a single group.
list                                       Retrieve a paginated list of groups.
list-memberships                           Retrieve a list of group memberships.
update                                     Update a group.
```

Groups are containers for organizing workspaces hierarchically.
They can be nested and inherit variables and managed identities
to children. Use group commands to create, update, delete groups,
set Terraform and environment variables, manage memberships, and
migrate groups between parents.

---

#### group create

##### Create a new group.

```
Usage: tharsis [global options] group create [options] <name>
```

   The group create command creates a new group. It allows
   setting a group's description (optional). Shows final
   output as JSON, if specified. Idempotent when used with
   --if-not-exists option.

<details>
<summary>Expand options</summary>

- `--description`: Description for the new group.

- `--if-not-exists`: Create a group if it does not already exist.

- `--json`: Show final output as JSON.

- `--parent-group-id`: Parent group ID.

</details>

##### Example:

```
tharsis group create \
  --parent-group-id trn:group:ops \
  --description "Operations group" \
  my-group
```

---

#### group delete

##### Delete a group.

```
Usage: tharsis [global options] group delete [options] <id>
```

   The group delete command deletes a group by its ID. Includes
   a force flag to delete the group even if resources are
   deployed (dangerous!).

<details>
<summary>Expand options</summary>

- `--force`: Force delete the group.

- `--version`: Metadata version of the resource to be deleted. In most cases, this is not required.

</details>

##### Example:

```
tharsis group delete \
  --force \
  trn:group:ops/my-group
```

---

#### group get

##### Get a single group.

```
Usage: tharsis [global options] group get [options] <id>
```

   The group get command retrieves a single group by its ID.
   Shows output as JSON, if specified.

<details>
<summary>Expand options</summary>

- `--json`: Show output as JSON.

</details>

##### Example:

```
tharsis group get \
  --json \
  trn:tharsis:group:ops/my-group
```

---

#### group list

##### Retrieve a paginated list of groups.

```
Usage: tharsis [global options] group list [options]
```

   The group list command prints information about (likely
   multiple) groups. Supports pagination, filtering and
   sorting the output.

<details>
<summary>Expand options</summary>

- `--cursor`: The cursor string for manual pagination.

- `--json`: Show final output as JSON.

- `--limit`: Maximum number of result elements to return. Defaults to 100.

- `--parent-id`: Filter to only direct sub-groups of this parent group.

- `--search`: Filter to only groups containing this substring in their path.

- `--sort-by`: Sort by this field (e.g., UPDATED_AT_ASC, UPDATED_AT_DESC, FULL_PATH_ASC, FULL_PATH_DESC).

</details>

##### Example:

```
tharsis group list \
  --parent-id trn:group:top-level/bottom-level \
  --sort-by FULL_PATH_ASC \
  --limit 5 \
  --json
```

---

#### group list-memberships

##### Retrieve a list of group memberships.

```
Usage: tharsis [global options] group list-memberships [options] <group-path>
```

   The group list-memberships command prints information about
   memberships for a specific group.

<details>
<summary>Expand options</summary>

- `--json`: Show final output as JSON.

</details>

##### Example:

```
tharsis group list-memberships top-level/my-group
```

---

#### group update

##### Update a group.

```
Usage: tharsis [global options] group update [options] <id>
```

   The group update command updates a group. Currently, it
   supports updating the description. Shows final output
   as JSON, if specified.

<details>
<summary>Expand options</summary>

- `--description`: Description for the group.

- `--json`: Show final output as JSON.

- `--version`: Metadata version of the resource to be updated. In most cases, this is not required.

</details>

##### Example:

```
tharsis group update \
  --description "Updated operations group" \
  trn:group:ops/my-group
```

***

### Command: managed-identity

##### Do operations on a managed identity.

**Subcommands:**

```
create                                     Create a new managed identity.
delete                                     Delete a managed identity.
get                                        Get a single managed identity.
update                                     Update a managed identity.
```

Managed identities provide OIDC-federated credentials for cloud
providers (AWS, Azure, Kubernetes) without storing secrets. Use
managed-identity commands to create, update, delete, and get
managed identities.

---

#### managed-identity create

##### Create a new managed identity.

```
Usage: tharsis [global options] managed-identity create [options] <name>
```

   The managed-identity create command creates a new managed identity.

<details>
<summary>Expand options</summary>

- `--aws-federated-role`: AWS IAM role. (Only if type is aws_federated)

- `--azure-federated-client-id`: Azure client ID. (Only if type is azure_federated)

- `--azure-federated-tenant-id`: Azure tenant ID. (Only if type is azure_federated)

- `--description`: Description for the managed identity.

- `--group-id`: Group ID or TRN where the managed identity will be created.

- `--json`: Show final output as JSON.

- `--tharsis-federated-service-account-id`: Tharsis service account ID or TRN. (Only if type is tharsis_federated)

- `--type`: The type of managed identity: aws_federated, azure_federated, tharsis_federated.

</details>

##### Example:

```
tharsis managed-identity create \
  --group-id trn:group:ops/my-group \
  --type aws_federated \
  --aws-federated-role arn:aws:iam::123456789012:role/MyRole \
  --description "AWS production role" \
  aws-prod
```

---

#### managed-identity delete

##### Delete a managed identity.

```
Usage: tharsis [global options] managed-identity delete [options] <id>
```

   The managed-identity delete command deletes a managed identity.

   Use with caution as deleting a managed identity is irreversible!

<details>
<summary>Expand options</summary>

- `--force`: Force delete the managed identity.

</details>

##### Example:

```
tharsis managed-identity delete --force trn:managed_identity:ops/my-group/aws-prod
```

---

#### managed-identity get

##### Get a single managed identity.

```
Usage: tharsis [global options] managed-identity get [options] <id>
```

   The managed-identity get command prints information about one
   managed identity.

<details>
<summary>Expand options</summary>

- `--json`: Show final output as JSON.

</details>

##### Example:

```
tharsis managed-identity get trn:managed_identity:ops/my-group/aws-prod
```

---

#### managed-identity update

##### Update a managed identity.

```
Usage: tharsis [global options] managed-identity update [options] <id>
```

   The managed-identity update command updates a managed identity.
   Currently, it supports updating the description and data.
   Shows final output as JSON, if specified.

<details>
<summary>Expand options</summary>

- `--aws-federated-role`: AWS IAM role. (Only if type is aws_federated)

- `--azure-federated-client-id`: Azure client ID. (Only if type is azure_federated)

- `--azure-federated-tenant-id`: Azure tenant ID. (Only if type is azure_federated)

- `--description`: Description for the managed identity.

- `--json`: Show final output as JSON.

- `--tharsis-federated-service-account-id`: Tharsis service account ID or TRN. (Only if type is tharsis_federated)

</details>

##### Example:

```
tharsis managed-identity update \
  --description "Updated AWS production role" \
  --aws-federated-role arn:aws:iam::123456789012:role/UpdatedRole \
  trn:managed_identity:ops/my-group/aws-prod
```

***

### Command: managed-identity-access-rule

##### Do operations on a managed identity access rule.

**Subcommands:**

```
create                                     Create a new managed identity access rule.
delete                                     Delete a managed identity access rule.
get                                        Get a managed identity access rule.
list                                       Retrieve a list of managed identity access rules.
update                                     Update a managed identity access rule.
```

Access rules control which runs can use a managed identity based
on conditions like module source or workspace path. Use these
commands to create, update, delete, list, and get access rules.

---

#### managed-identity-access-rule create

##### Create a new managed identity access rule.

```
Usage: tharsis [global options] managed-identity-access-rule create [options]
```

   The managed-identity-access-rule create command creates a new managed identity access rule.

<details>
<summary>Expand options</summary>

- `--allowed-service-account-id`: Allowed service account ID. (This flag may be repeated)

- `--allowed-team-id`: Allowed team ID. (This flag may be repeated)

- `--allowed-user-id`: Allowed user ID. (This flag may be repeated)

- `--json`: Show final output as JSON.

- `--managed-identity-id`: The ID or TRN of the managed identity.

- `--module-attestation-policy`: Module attestation policy in format "[PredicateType=someval,]PublicKeyFile=/path/to/file". (This flag may be repeated)

- `--run-stage`: The run stage: plan or apply.

- `--type`: The type of access rule: eligible_principals or module_attestation.

- `--verify-state-lineage`: Verify state lineage.

</details>

##### Example:

```
tharsis managed-identity-access-rule create \
  --managed-identity-id trn:managed_identity:ops/my-identity \
  --type eligible_principals \
  --run-stage plan \
  --allowed-user-id trn:user:john.smith \
  --allowed-team-id trn:team:my-team
```

---

#### managed-identity-access-rule delete

##### Delete a managed identity access rule.

```
Usage: tharsis [global options] managed-identity-access-rule delete [options] <id>
```

   The managed-identity-access-rule delete command deletes a managed identity access rule.

##### Example:

```
tharsis managed-identity-access-rule delete TV80ZG...
```

---

#### managed-identity-access-rule get

##### Get a managed identity access rule.

```
Usage: tharsis [global options] managed-identity-access-rule get [options] <id>
```

   The managed-identity-access-rule get command gets a managed identity access rule by ID.

<details>
<summary>Expand options</summary>

- `--json`: Show final output as JSON.

</details>

##### Example:

```
tharsis managed-identity-access-rule get trn:managed_identity_access_rule:abc123
```

---

#### managed-identity-access-rule list

##### Retrieve a list of managed identity access rules.

```
Usage: tharsis [global options] managed-identity-access-rule list [options] <managed-identity-id>
```

   The managed-identity-access-rule list command prints information about
   access rules for a specific managed identity.

<details>
<summary>Expand options</summary>

- `--json`: Show final output as JSON.

</details>

##### Example:

```
tharsis managed-identity-access-rule list \
  trn:managed_identity:ops/my-identity
```

---

#### managed-identity-access-rule update

##### Update a managed identity access rule.

```
Usage: tharsis [global options] managed-identity-access-rule update [options] <id>
```

   The managed-identity-access-rule update command updates an existing managed identity access rule.

<details>
<summary>Expand options</summary>

- `--allowed-service-account-id`: Allowed service account ID. (This flag may be repeated)

- `--allowed-team-id`: Allowed team ID. (This flag may be repeated)

- `--allowed-user-id`: Allowed user ID. (This flag may be repeated)

- `--json`: Show final output as JSON.

- `--module-attestation-policy`: Module attestation policy in format "[PredicateType=someval,]PublicKeyFile=/path/to/file". (This flag may be repeated)

- `--run-stage`: The run stage: plan or apply.

- `--verify-state-lineage`: Verify state lineage (true or false).

</details>

##### Example:

```
tharsis managed-identity-access-rule update \
  --run-stage apply \
  --allowed-user-id trn:user:john.smith \
  TV80ZG...
```

***

### Command: managed-identity-alias

##### Do operations on a managed identity alias.

**Subcommands:**

```
create                                     Create a new managed identity alias.
delete                                     Delete a managed identity alias.
```

Aliases allow referencing managed identities from other groups.
Use these commands to create and delete managed identity aliases.

---

#### managed-identity-alias create

##### Create a new managed identity alias.

```
Usage: tharsis [global options] managed-identity-alias create [options] <name>
```

   The managed-identity-alias create command creates a new managed identity alias.

<details>
<summary>Expand options</summary>

- `--alias-source-id`: The ID or TRN of the source managed identity.

- `--group-id`: Group ID or TRN where the managed identity alias will be created.

- `--json`: Show final output as JSON.

</details>

##### Example:

```
tharsis managed-identity-alias create \
  --group-id trn:group:ops/my-group \
  --alias-source-id trn:managed_identity:source-identity \
  prod-identity-alias
```

---

#### managed-identity-alias delete

##### Delete a managed identity alias.

```
Usage: tharsis [global options] managed-identity-alias delete [options] <id>
```

   The managed-identity-alias delete command deletes a managed identity alias.

<details>
<summary>Expand options</summary>

- `--force`: Force delete the managed identity alias.

</details>

##### Example:

```
tharsis managed-identity-alias delete trn:managed_identity:ops/my-group/prod-identity-alias
```

***

### Command: module

##### Do operations on a terraform module.

**Subcommands:**

```
create                                     Create a new Terraform module.
create-attestation                         Create a new module attestation.
delete                                     Delete a Terraform module.
delete-attestation                         Delete a module attestation.
delete-version                             Delete a module version.
get                                        Get a single Terraform module.
get-version                                Get a module version by ID or TRN.
list                                       Retrieve a paginated list of modules.
list-attestations                          Retrieve a paginated list of module attestations.
list-versions                              Retrieve a paginated list of module versions.
update                                     Update a Terraform module.
update-attestation                         Update a module attestation.
```

The module registry stores Terraform modules with versioning and
attestation support. Use module commands to create, update, delete
modules, upload versions, manage attestations, and list modules
and versions.

---

#### module create

##### Create a new Terraform module.

```
Usage: tharsis [global options] module create [options] <module-name/system>
```

   The module create command creates a new Terraform module. It
   requires a group ID and repository URL. The argument should be
   in the format: module-name/system (e.g., vpc/aws). Shows final
   output as JSON, if specified. Idempotent when used with
   --if-not-exists option.

<details>
<summary>Expand options</summary>

- `--group-id`: Parent group ID.

- `--if-not-exists`: Create a module if it does not already exist.

- `--json`: Show final output as JSON.

- `--private`: Whether the module is private.

- `--repository-url`: The repository URL for the module.

</details>

##### Example:

```
tharsis module create \
  --group-id trn:group:ops/my-group \
  --repository-url https://github.com/example/terraform-aws-vpc \
  --private \
  vpc/aws
```

---

#### module create-attestation

##### Create a new module attestation.

```
Usage: tharsis [global options] module create-attestation [options] <module-id>
```

   The module create-attestation command creates a new module attestation.

<details>
<summary>Expand options</summary>

- `--attestation-data`: The attestation data (must be a Base64-encoded string).

- `--description`: Description for the attestation.

- `--json`: Show final output as JSON.

</details>

##### Example:

```
tharsis module create-attestation \
  --description "Attestation for v1.0.0" \
  --attestation-data '{"key":"value"}' \
  trn:terraform_module:ops/installer/aws
```

---

#### module delete

##### Delete a Terraform module.

```
Usage: tharsis [global options] module delete [options] <id>
```

   The module delete command deletes a Terraform module.

   Use with caution as deleting a module is irreversible!

##### Example:

```
tharsis module delete trn:terraform_module:ops/my-group/vpc
```

---

#### module delete-attestation

##### Delete a module attestation.

```
Usage: tharsis [global options] module delete-attestation [options] <id>
```

   The module delete-attestation command deletes a module attestation.

<details>
<summary>Expand options</summary>

- `--force`: Force delete the module attestation.

</details>

##### Example:

```
tharsis module delete-attestation trn:terraform_module_attestation:ops/installer/aws:VE1W
```

---

#### module delete-version

##### Delete a module version.

```
Usage: tharsis [global options] module delete-version [options] <version-id>
```

   The module delete-version command deletes a module version.

<details>
<summary>Expand options</summary>

- `--force`: Force deletion without confirmation.

- `--version`: Metadata version of the resource to be deleted. In most cases, this is not required.

</details>

##### Example:

```
tharsis module delete-version trn:terraform_module_version:ops/installer/aws/1.0.0
```

---

#### module get

##### Get a single Terraform module.

```
Usage: tharsis [global options] module get [options] <id>
```

   The module get command prints information about one
   Terraform module.

<details>
<summary>Expand options</summary>

- `--json`: Show final output as JSON.

</details>

##### Example:

```
tharsis module get trn:terraform_module:ops/my-group/vpc
```

---

#### module get-version

##### Get a module version by ID or TRN.

```
Usage: tharsis [global options] module get-version [options] <version-id>
```

   The module get-version command retrieves details about a specific module version.

<details>
<summary>Expand options</summary>

- `--json`: Output in JSON format.

</details>

##### Example:

```
tharsis module get-version trn:terraform_module_version:ops/installer/aws/1.0.0
```

---

#### module list

##### Retrieve a paginated list of modules.

```
Usage: tharsis [global options] module list [options]
```

   The module list command prints information about (likely
   multiple) modules. Supports pagination, filtering and
   sorting the output.

<details>
<summary>Expand options</summary>

- `--cursor`: The cursor string for manual pagination.

- `--group-id`: Filter to only modules in this group.

- `--include-inherited`: Include modules inherited from parent groups.

- `--json`: Show final output as JSON.

- `--limit`: Maximum number of result elements to return. Defaults to 100.

- `--search`: Filter to only modules containing this substring in their path.

- `--sort-by`: Sort by this field (e.g., NAME_ASC, NAME_DESC, UPDATED_AT_ASC, UPDATED_AT_DESC).

</details>

##### Example:

```
tharsis module list \
  --group-id trn:group:top-level \
  --include-inherited \
  --sort-by UPDATED_AT_DESC \
  --limit 5 \
  --json
```

---

#### module list-attestations

##### Retrieve a paginated list of module attestations.

```
Usage: tharsis [global options] module list-attestations [options] <module-id>
```

   The module list-attestations command prints information about attestations
   for a specific module. Supports pagination, filtering and sorting.

<details>
<summary>Expand options</summary>

- `--cursor`: The cursor string for manual pagination.

- `--digest`: Filter to attestations with this digest.

- `--json`: Show final output as JSON.

- `--limit`: Maximum number of result elements to return. Defaults to 100.

- `--sort-by`: Sort by this field (e.g., CREATED_AT_ASC, CREATED_AT_DESC).

</details>

##### Example:

```
tharsis module list-attestations \
  --sort-by CREATED_AT_DESC \
  --limit 10 \
  trn:terraform_module:ops/installer/aws
```

---

#### module list-versions

##### Retrieve a paginated list of module versions.

```
Usage: tharsis [global options] module list-versions [options] <module-id>
```

   The module list-versions command prints information about versions
   of a specific module. Supports pagination, filtering and sorting.

<details>
<summary>Expand options</summary>

- `--cursor`: The cursor string for manual pagination.

- `--json`: Show final output as JSON.

- `--latest`: Filter to only the latest version.

- `--limit`: Maximum number of result elements to return. Defaults to 100.

- `--search`: Filter to versions containing this substring.

- `--semantic-version`: Filter to a specific semantic version.

- `--sort-by`: Sort by this field (e.g., CREATED_AT_ASC, CREATED_AT_DESC).

</details>

##### Example:

```
tharsis module list-versions \
  --search 1.0 \
  --sort-by CREATED_AT_DESC \
  --limit 10 \
  trn:terraform_module:ops/installer/aws
```

---

#### module update

##### Update a Terraform module.

```
Usage: tharsis [global options] module update [options] <id>
```

   The module update command updates a Terraform module.
   Currently, it supports updating the repository URL and
   private flag. Shows final output as JSON, if specified.

<details>
<summary>Expand options</summary>

- `--json`: Show final output as JSON.

- `--private`: Whether the module is private.

- `--repository-url`: The repository URL for the module.

- `--version`: Metadata version of the resource to be updated. In most cases, this is not required.

</details>

##### Example:

```
tharsis module update \
  --repository-url https://github.com/example/terraform-aws-vpc-v2 \
  --private true \
  trn:terraform_module:ops/my-group/vpc
```

---

#### module update-attestation

##### Update a module attestation.

```
Usage: tharsis [global options] module update-attestation [options] <id>
```

   The module update-attestation command updates an existing module attestation.

<details>
<summary>Expand options</summary>

- `--description`: Description for the attestation.

- `--json`: Show final output as JSON.

</details>

##### Example:

```
tharsis module update-attestation \
  --description "Updated description" \
  trn:terraform_module_attestation:ops/installer/aws:VE1W
```

***

### Command: plan

##### Create a speculative plan

The plan command creates a speculative plan to view the changes
Terraform will make to your infrastructure without applying them.
Supports setting run-scoped Terraform and environment variables,
planning destroy runs, and using remote module sources.

***

### Command: run

##### Do operations on runs.

Runs are units of execution (plan or apply) that create, update,
or destroy infrastructure resources. Use run commands to cancel
runs gracefully or forcefully.

***

### Command: runner-agent

##### Do operations on runner agents.

**Subcommands:**

```
create                                     Create a new runner agent.
delete                                     Delete a runner agent.
get                                        Get a runner agent.
update                                     Update a runner agent.
```

Runner agents are distributed job executors responsible for
launching Terraform jobs that deploy infrastructure to the cloud.
Use runner-agent commands to create, update, delete, get agents,
and assign or unassign service accounts.

---

#### runner-agent create

##### Create a new runner agent.

```
Usage: tharsis [global options] runner-agent create [options] <name>
```

   The runner-agent create command creates a new runner agent.

<details>
<summary>Expand options</summary>

- `--description`: Description for the runner agent.

- `--group-id`: Group ID or TRN where the runner agent will be created.

- `--json`: Show final output as JSON.

- `--run-untagged-jobs`: Allow the runner agent to execute jobs without tags.

- `--tag`: Tag for the runner agent. (This flag may be repeated)

</details>

##### Example:

```
tharsis runner-agent create \
  --group-id trn:group:ops/my-group \
  --description "Production runner" \
  --run-untagged-jobs \
  --tag prod \
  --tag us-east-1 \
  prod-runner
```

---

#### runner-agent delete

##### Delete a runner agent.

```
Usage: tharsis [global options] runner-agent delete [options] <id>
```

   The runner-agent delete command deletes a runner agent.

<details>
<summary>Expand options</summary>

- `--version`: Metadata version of the resource to be deleted. In most cases, this is not required.

</details>

##### Example:

```
tharsis runner-agent delete trn:runner:ops/prod-runner
```

---

#### runner-agent get

##### Get a runner agent.

```
Usage: tharsis [global options] runner-agent get [options] <id>
```

   The runner-agent get command gets a runner agent by ID.

<details>
<summary>Expand options</summary>

- `--json`: Show final output as JSON.

</details>

##### Example:

```
tharsis runner-agent get trn:runner:ops/prod-runner
```

---

#### runner-agent update

##### Update a runner agent.

```
Usage: tharsis [global options] runner-agent update [options] <id>
```

   The runner-agent update command updates an existing runner agent.

<details>
<summary>Expand options</summary>

- `--description`: Description for the runner agent.

- `--disabled`: Enable or disable the runner agent (true or false).

- `--json`: Show final output as JSON.

- `--run-untagged-jobs`: Allow the runner agent to execute jobs without tags (true or false).

- `--tag`: Tag for the runner agent. (This flag may be repeated)

- `--version`: Metadata version of the resource to be updated. In most cases, this is not required.

</details>

##### Example:

```
tharsis runner-agent update \
  --description "Updated description" \
  --disabled true \
  --tag prod \
  --tag us-west-2 \
  trn:runner:abc123
```

***

### Command: service-account

##### Create an authentication token for a service account.

Service accounts provide machine-to-machine authentication for
CI/CD pipelines and automation. Use service-account commands to
create authentication tokens.

***

### Command: sso

##### Log in to the OAuth2 provider and return an authentication token.

**Subcommands:**

```
login                                      Log in to the OAuth2 provider and return an authentication token.
```

The sso command authenticates the CLI with the OAuth2 provider,
and allows making authenticated calls to Tharsis backend.

---

#### sso login

##### Log in to the OAuth2 provider and return an authentication token.

```
Usage: tharsis [global options] sso login
```

   The login command starts an embedded web server and opens
   a web browser page or tab pointed at said web server.
   That redirects to the OAuth2 provider's login page, where
   the user can sign in. If there is an SSO scheme active,
   that will sign in the user. The login command captures
   the authentication token for use in subsequent commands.

##### Example:

```
tharsis sso login
```

***

### Command: terraform-provider

##### Do operations on a terraform provider.

The provider registry stores Terraform providers with versioning
support. Use terraform-provider commands to create providers and
upload provider versions to the registry.

***

### Command: terraform-provider-mirror

##### Mirror Terraform providers from any Terraform registry.

**Subcommands:**

```
list-platforms                             Retrieve a paginated list of provider platform mirrors.
list-versions                              Retrieve a paginated list of provider version mirrors.
```

The provider mirror caches Terraform providers from any registry
for use within a group hierarchy. It supports Terraform's Provider
Network Mirror Protocol and gives root group owners control over
which providers, platform packages, and registries are available.
Use these commands to sync providers, list versions and platforms,
get version details, and delete versions or platforms.

---

#### terraform-provider-mirror list-platforms

##### Retrieve a paginated list of provider platform mirrors.

```
Usage: tharsis [global options] terraform-provider-mirror list-platforms [options] <version-mirror-id>
```

   The terraform-provider-mirror list-platforms command prints information
   about provider platform mirrors for a version mirror. Supports pagination,
   filtering and sorting.

<details>
<summary>Expand options</summary>

- `--architecture`: Filter to platforms with this architecture.

- `--cursor`: The cursor string for manual pagination.

- `--json`: Show final output as JSON.

- `--limit`: Maximum number of result elements to return. Defaults to 100.

- `--os`: Filter to platforms with this OS.

- `--sort-by`: Sort by this field (e.g., CREATED_AT_ASC, CREATED_AT_DESC).

</details>

##### Example:

```
tharsis terraform-provider-mirror list-platforms \
  --os linux \
  --architecture amd64 \
  --sort-by CREATED_AT_DESC \
  trn:terraform_provider_version_mirror:ops/registry.terraform.io/hashicorp/time/0.13.1
```

---

#### terraform-provider-mirror list-versions

##### Retrieve a paginated list of provider version mirrors.

```
Usage: tharsis [global options] terraform-provider-mirror list-versions [options] <namespace-path>
```

   The terraform-provider-mirror list-versions command prints information
   about provider version mirrors in a namespace. Supports pagination and sorting.

<details>
<summary>Expand options</summary>

- `--cursor`: The cursor string for manual pagination.

- `--json`: Show final output as JSON.

- `--limit`: Maximum number of result elements to return. Defaults to 100.

- `--sort-by`: Sort by this field (e.g., CREATED_AT_ASC, CREATED_AT_DESC).

</details>

##### Example:

```
tharsis terraform-provider-mirror list-versions \
  --sort-by CREATED_AT_DESC \
  --limit 10 \
  ops
```

***

### Command: version

##### Get the CLI's version.

```
Usage: tharsis [global options] version
```

  The tharsis version command returns the CLI's version.

<details>
<summary>Expand options</summary>

- `--json`: Show final output as JSON.

</details>

##### Example:

```
tharsis version --json
```

***

### Command: workspace

##### Do operations on workspaces.

**Subcommands:**

```
create                                     Create a new workspace.
delete                                     Delete a workspace.
get                                        Get a single workspace.
list                                       Retrieve a paginated list of workspaces.
list-memberships                           Retrieve a list of workspace memberships.
update                                     Update a workspace.
```

Workspaces contain Terraform deployments, state, runs, and variables.
Use workspace commands to create, update, delete workspaces, assign
and unassign managed identities, set Terraform and environment
variables, manage memberships, and view workspace outputs.

---

#### workspace create

##### Create a new workspace.

```
Usage: tharsis [global options] workspace create [options] <name>
```

   The workspace create command creates a new workspace. It
   allows setting a workspace's description (optional),
   maximum job duration and managed identity. Shows final
   output as JSON, if specified. Idempotent when used with
   --if-not-exists option.

<details>
<summary>Expand options</summary>

- `--description`: Description for the new workspace.

- `--if-not-exists`: Create a workspace if it does not already exist.

- `--json`: Show final output as JSON.

- `--label`: Labels for the new workspace (key=value). Can be specified multiple times.

- `--managed-identity`: The ID of a managed identity to assign.

- `--max-job-duration`: The amount of minutes before a job is gracefully canceled (Default 720).

- `--parent-group-id`: Parent group ID.

- `--prevent-destroy-plan`: Whether a run/plan will be prevented from destroying deployed resources.

- `--terraform-version`: The default Terraform CLI version for the new workspace.

</details>

##### Example:

```
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

---

#### workspace delete

##### Delete a workspace.

```
Usage: tharsis [global options] workspace delete [options] <id>
```

   The workspace delete command deletes a workspace. Includes
   a force flag to delete the workspace even if resources are
   deployed (dangerous!).

   Use with caution as deleting a workspace is irreversible!

<details>
<summary>Expand options</summary>

- `--force`: Force the workspace to delete even if resources are deployed.

- `--version`: Metadata version of the resource to be deleted. In most cases, this is not required.

</details>

##### Example:

```
tharsis workspace delete --force trn:workspace:ops/my-group/my-workspace
```

---

#### workspace get

##### Get a single workspace.

```
Usage: tharsis [global options] workspace get [options] <id>
```

   The workspace get command prints information about one
   workspace.

<details>
<summary>Expand options</summary>

- `--json`: Show final output as JSON.

</details>

##### Example:

```
tharsis workspace get trn:workspace:ops/my-group/my-workspace
```

---

#### workspace list

##### Retrieve a paginated list of workspaces.

```
Usage: tharsis [global options] workspace list [options]
```

   The workspace list command prints information about (likely
   multiple) workspaces. Supports pagination, filtering and
   sorting the output.

<details>
<summary>Expand options</summary>

- `--cursor`: The cursor string for manual pagination.

- `--group-id`: Filter to only workspaces in this group.

- `--json`: Show final output as JSON.

- `--label`: Filter by label (key=value). This flag may be repeated.

- `--limit`: Maximum number of result elements to return. Defaults to 100.

- `--search`: Filter to only workspaces containing this substring in their path.

- `--sort-by`: Sort by this field (e.g., UPDATED_AT_ASC, UPDATED_AT_DESC, FULL_PATH_ASC, FULL_PATH_DESC).

</details>

##### Example:

```
tharsis workspace list \
  --group-id trn:group:top-level \
  --label env=prod \
  --label team=platform \
  --sort-by FULL_PATH_ASC \
  --limit 5 \
  --json
```

---

#### workspace list-memberships

##### Retrieve a list of workspace memberships.

```
Usage: tharsis [global options] workspace list-memberships [options] <workspace-path>
```

   The workspace list-memberships command prints information about
   memberships for a specific workspace.

<details>
<summary>Expand options</summary>

- `--json`: Show final output as JSON.

</details>

##### Example:

```
tharsis workspace list-memberships top-level/my-workspace
```

---

#### workspace update

##### Update a workspace.

```
Usage: tharsis [global options] workspace update [options] <id>
```

   The workspace update command updates a workspace.
   Currently, it supports updating the description and the
   maximum job duration. Shows final output as JSON, if
   specified.

<details>
<summary>Expand options</summary>

- `--description`: Description for the workspace.

- `--json`: Show final output as JSON.

- `--label`: Labels for the workspace (key=value). Can be specified multiple times.

- `--max-job-duration`: The amount of minutes before a job is gracefully canceled.

- `--prevent-destroy-plan`: Whether a run/plan will be prevented from destroying deployed resources.

- `--terraform-version`: The default Terraform CLI version for the workspace.

- `--version`: Metadata version of the resource to be updated. In most cases, this is not required.

</details>

##### Example:

```
tharsis workspace update \
  --description "Updated production workspace" \
  --terraform-version "1.6.0" \
  --max-job-duration 120 \
  --prevent-destroy-plan true \
  trn:workspace:ops/my-group/my-workspace
```

