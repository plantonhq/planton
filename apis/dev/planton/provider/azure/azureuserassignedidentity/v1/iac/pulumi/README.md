# AzureUserAssignedIdentity Pulumi Module

## Overview

This Pulumi module provisions an Azure User-Assigned Managed Identity with optional
RBAC role assignments using the Azure Classic provider (`pulumi-azure`). It creates
a `authorization.UserAssignedIdentity` resource and zero or more
`authorization.Assignment` resources based on the spec's `role_assignments` list.

## Resources Created

- `authorization.UserAssignedIdentity` -- the managed identity
- `authorization.Assignment` (N) -- one per role assignment in the spec

## Inputs

The module receives an `AzureUserAssignedIdentityStackInput` containing:

- `target.spec.region` -- Azure region
- `target.spec.resource_group` -- resource group name (resolved from StringValueOrRef)
- `target.spec.name` -- identity name
- `target.spec.role_assignments` -- list of scope + role_definition_name pairs
- `target.metadata` -- Planton metadata for tagging
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `identity_id` | Azure Resource Manager ID of the identity |
| `principal_id` | Service Principal Object ID |
| `client_id` | Application/Client ID |
| `tenant_id` | Azure AD Tenant ID |

## Key Implementation Details

### Role Assignment Naming

Role assignments are named `{identity-name}-ra-{index}` to ensure uniqueness within
the Pulumi stack. The index corresponds to the position in the `role_assignments` list.

### AAD Replication Handling

`SkipServicePrincipalAadCheck` is set to `true` on all role assignments. Azure AD
replication is eventually consistent -- after creating a managed identity, there can
be a delay before the principal is visible for role assignment. Setting this flag
avoids 403 errors during creation.

### DependsOn

All role assignments explicitly depend on the identity resource to ensure the
principal ID is available before Azure attempts to create the role binding.

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
