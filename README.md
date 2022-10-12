# Tharsis CLI

Tharsis CLI is a command-line interface to the Tharsis remote Terraform backend. The CLI allows managing groups, workspaces, creating runs from Terraform modules and much more.

- **Groups**: Create, update, delete groups and organize your workspaces in anyway desired.
- **Workspaces**: Create, update, delete workspaces and even assign / unassign managed identities.
- **Variables**: Set Terraform (including complex variables) and environment variables in groups, workspaces, or run-scoped.
- **Authentication**: Authenticate using either SSO or service accounts for M2M.
- **CI/CD**: Integrate easily into a CI/CD pipeline environment and deploy to Tharsis using a [service account](https://gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-docs/-/blob/main/docs/guides/overviews/service_accounts.md).
- **Runs**: Gracefully or forcefully cancel runs.
- **Private Registries**: Deploy modules from private Terraform registries.

## Get started

Instructions on downloading the latest pre-built releases or building a binary from source can be found [here](https://gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-docs/-/blob/main/docs/setup/cli/install.md).

## Documentation

- Tharsis CLI documentation is available at https://gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-docs/-/blob/main/docs/cli/tharsis/intro.md.

## Security

If you've discovered a security vulnerability in Tharsis CLI, please let us know by creating a **confidential** issue in this project.

## Statement of support

Please submit any bugs or feature requests for Tharsis.  Of course, MR's are even better.  :)

## License

Tharsis CLI is distributed under [Mozilla Public License v2.0](https://www.mozilla.org/en-US/MPL/2.0/).
