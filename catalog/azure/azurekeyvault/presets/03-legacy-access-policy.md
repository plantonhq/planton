# Legacy Access-Policy Vault

This preset runs the vault in the legacy access-policy authorization mode:
`rbacAuthorizationEnabled: false` with explicit per-principal permission
lists carried on the vault itself. ARM stores but IGNORES access policies
on an RBAC-mode vault, so this preset exists precisely for the orgs and
workloads that have not yet moved to RBAC.

The example grants one workload identity the consumer surface: read
secrets, read certificates, and wrap/unwrap with keys (the
customer-managed-key operations).

## When to Use

- Existing estates standardized on access policies that cannot switch to
  RBAC yet (the two modes cannot mix -- the vault runs one or the other)
- Third-party or legacy software whose documentation assumes access
  policies

Prefer the RBAC presets for anything new: role assignments participate in
PIM and access reviews, support fine-grained scopes, and compose as
first-class `AzureRoleAssignment` resources.

## Key Configuration Choices

- **`objectId` is the identity's PRINCIPAL id** (the directory object),
  not its client id -- replace the placeholder with the literal object id,
  or with a `valueFrom` reference to an `AzureUserAssignedIdentity`'s
  `principal_id` output in composed environments
- **Permission lists are the capability boundary** -- grant only what the
  principal actually performs; `KEY_WRAP_KEY`/`KEY_UNWRAP_KEY` are the CMK
  consumer pair, management permissions (CREATE/DELETE/PURGE) belong to
  operators
- **Tenant defaults to the vault's own** -- access policies cannot span
  tenants in practice, so the preset leaves `tenantId` unset

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<resource-group-name>` | The resource group to create the vault in | The resource group's `status.outputs.resource_group_name` |
| `myorg-legacy-vault` | 3-24 chars, globally unique | Your naming convention |
| `<principal-object-id>` | Object (principal) ID of the identity being granted access, e.g. an AzureUserAssignedIdentity | The identity resource's `status.outputs.principal_id` |
| `<cost-center>` | Your org's cost-attribution tag value | Your tagging convention |

## Migration Path

Moving to RBAC later is an in-place update: set
`rbacAuthorizationEnabled: true`, drop `accessPolicies`, and create
equivalent `AzureRoleAssignment` grants (the switch requires
Microsoft.Authorization write permission on the vault).
