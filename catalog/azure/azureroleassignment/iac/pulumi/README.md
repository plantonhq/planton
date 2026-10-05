# AzureRoleAssignment Pulumi Module

## Overview

This Pulumi module provisions an Azure RBAC role assignment using the Azure
Classic provider (`pulumi-azure`). It creates a single
`authorization.Assignment` binding a role (by built-in name or definition ID)
to a principal at an ARM scope.

Role assignments are immutable in Azure: any spec change replaces the
assignment (delete + create). They carry no ARM tags -- the
`Microsoft.Authorization` resource type does not support them.

## Resources Created

- `authorization.Assignment` -- the Azure role assignment

## Inputs

The module receives an `AzureRoleAssignmentIacInput` containing:

- `target.spec.scope` -- the ARM scope of the grant (references resolved to a literal by the platform)
- `target.spec.role_definition_name` / `target.spec.role_definition_id` -- exactly one identifies the role
- `target.spec.principal_id` -- the Azure AD object ID receiving the role
- `target.spec.principal_type`, `description`, `condition`, `condition_version`,
  `delegated_managed_identity_resource_id`, `skip_service_principal_aad_check`, `name` -- optional refinements
- `provider_config` -- Azure credentials (static client secret, keyless web identity, or ambient chain)

## Outputs

| Output | Description |
|--------|-------------|
| `role_assignment_id` | Fully-scoped ARM ID of the assignment |
| `name` | The assignment's GUID resource name |
| `scope` | The scope the role was granted at |
| `role_definition_id` | The role definition ID Azure resolved (even when a role name was given) |
| `principal_id` | The principal's Azure AD object ID |
| `principal_type` | The principal type Azure recorded |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
