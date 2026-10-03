# AzureRoleDefinition Pulumi Module

## Overview

This Pulumi module provisions a custom Azure RBAC role definition using the
Azure Classic provider (`pulumi-azure`). It creates a single
`authorization.RoleDefinition`: a named, reusable bundle of control-plane and
data-plane permissions that role assignments can then grant to principals.

Name, description, permissions, and assignable scopes update in place; the
creation scope and the pinned GUID are the definition's ARM identity, so
changing either replaces it. Updates and deletes are eventually consistent --
the provider polls until Azure settles, so those operations take a few
minutes. Role definitions carry no ARM tags (the `Microsoft.Authorization`
resource type does not support them).

## Resources Created

- `authorization.RoleDefinition` -- the custom role definition

## Inputs

The module receives an `AzureRoleDefinitionIacInput` containing:

- `target.spec.name` -- the role's tenant-unique display name
- `target.spec.scope` -- the ARM creation scope (references resolved to a literal by the platform)
- `target.spec.permissions` -- permission blocks (actions / not_actions / data_actions / not_data_actions)
- `target.spec.description`, `assignable_scopes`, `role_definition_id` -- optional refinements
- `provider_config` -- Azure credentials (static client secret, keyless web identity, or ambient chain)

## Outputs

| Output | Description |
|--------|-------------|
| `role_definition_id` | Fully-scoped ARM ID of the definition (what a role assignment binds) |
| `role_definition_guid` | The definition's GUID resource name |
| `role_name` | The role's display name as deployed |
| `scope` | The scope the definition was created at |
| `assignable_scopes` | The assignable scopes Azure recorded |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
