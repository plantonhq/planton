# AzureUserAssignedIdentity Pulumi Module

## Overview

This Pulumi module provisions a user-assigned managed identity using the
Azure Classic provider (`pulumi-azure`). It creates a single
`authorization.UserAssignedIdentity`: the standalone, credential-free Azure
AD identity workloads authenticate as.

Tags and isolation scope update in place; the name, region, and resource
group are the identity's ARM identity, so changing any of them replaces it --
minting a NEW principal and client ID that invalidates existing grants and
federated trust rules. The identity is deliberately just the identity:
grants live in `AzureRoleAssignment` and keyless trust rules in
`AzureFederatedIdentityCredential`, both referencing this module's outputs.

## Resources Created

- `authorization.UserAssignedIdentity` -- the managed identity

## Inputs

The module receives an `AzureUserAssignedIdentityIacInput` containing:

- `target.spec.region` -- the Azure region (a regional resource)
- `target.spec.resource_group` -- the resource group name (references resolved to a literal by the platform)
- `target.spec.name` -- the identity's name, unique within the resource group
- `target.spec.isolation_scope` -- optional opt-in regional isolation
- `target.spec.tags` -- user tags merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials (static client secret, keyless web identity, or ambient chain)

## Outputs

| Output | Description |
|--------|-------------|
| `identity_id` | ARM ID of the identity (what consumers and federated credentials attach to) |
| `principal_id` | Service principal object ID (what role assignments grant to) |
| `client_id` | Client/application ID (what workloads present to authenticate) |
| `tenant_id` | Azure AD tenant ID |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
