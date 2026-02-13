# AzureUserAssignedIdentity Pulumi Module -- Architecture Overview

## Purpose

This module is the Pulumi implementation for the `AzureUserAssignedIdentity` OpenMCF
component. It translates the protobuf-defined spec into Azure infrastructure using the
Pulumi Azure Classic SDK.

## Architecture

```
AzureUserAssignedIdentityStackInput
  ├── target (AzureUserAssignedIdentity)
  │     ├── metadata (name, org, env)
  │     └── spec (region, resource_group, name, role_assignments[])
  └── provider_config (credentials)
         │
         ▼
  ┌──────────────────┐
  │  module/main.go   │  Creates provider + identity + N role assignments
  │  module/locals.go │  Resolves StringValueOrRef fields, builds tags
  │  module/outputs.go│  Defines output constant names
  └──────────────────┘
         │
         ▼
  Stack Outputs: identity_id, principal_id, client_id, tenant_id
```

## Resource Graph

```
  authorization.UserAssignedIdentity
         │
         ├── authorization.Assignment[0]  (scope: Key Vault, role: Secrets User)
         ├── authorization.Assignment[1]  (scope: Storage, role: Blob Contributor)
         └── authorization.Assignment[N]  (scope: ..., role: ...)
```

All role assignments depend on the identity and use its `PrincipalId` output.

## Key Implementation Details

### StringValueOrRef Fields

Three field types use `StringValueOrRef`:
- `resource_group` -- references an AzureResourceGroup output
- `scope` (in each role assignment) -- references any Azure resource's output

The platform middleware resolves all `valueFrom` references before the IaC module runs,
so the module simply calls `.GetValue()` to extract the resolved string.

### Role Assignment Iteration

Role assignments are created in a loop over `spec.RoleAssignments`. Each assignment
gets a unique Pulumi resource name based on the identity name and index.

### AAD Replication

Azure AD replication can cause a race condition where the identity's principal ID
is not yet visible when the first role assignment is created. The module handles this
by setting `SkipServicePrincipalAadCheck: true` and adding an explicit `DependsOn`
on the identity resource.
