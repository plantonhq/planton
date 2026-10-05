# AzureRoleAssignment

## Overview

`AzureRoleAssignment` provisions an Azure RBAC role assignment -- the grant of a
role to a principal at a scope. It is the atomic unit of authorization in Azure:
everything an identity is allowed to do is the sum of the role assignments that
target it.

## Why a First-Class Resource?

Role assignments are real infrastructure with their own lifecycle:

- **Many per principal, many per scope** -- a single identity typically holds
  grants on a registry (pull), a key vault (read secrets), and a storage account
  (write blobs); each grant is an independent record with an independent lifecycle
- **Composition seam** -- a grant connects two other resources (a principal and a
  scope) without owning either; modeling it as its own node keeps identities,
  grants, and protected resources independently deployable and auditable
- **Never mutate what you reference** -- a component that references an identity
  must not modify its permissions; first-class assignments are how grants are
  expressed without hidden side effects

## Key Features

- **Any ARM scope** -- management group, subscription, resource group, or a single
  resource; scope defaults to referencing an `AzureResourceGroup`, and any other
  resource's ID output works via an explicit `valueFrom`
- **Built-in or custom roles** -- reference a role by built-in name ("Reader") or
  by role definition ID (custom roles)
- **Managed-identity-first composition** -- `principal_id` defaults to referencing
  an `AzureUserAssignedIdentity`'s principal ID output; literal object IDs cover
  users, groups, and externally managed service principals
- **ABAC conditions** -- attach an Azure attribute-based access control condition
  to narrow when the role applies
- **Cross-tenant delegation** -- supports Azure Lighthouse scenarios via
  `delegated_managed_identity_resource_id`

## When to Use

- Granting a workload identity access to the Azure resources it operates on
  (AcrPull on a registry, Key Vault Secrets User on a vault)
- Granting CI/CD deploy identities scoped rights on an environment's resource group
- Assigning custom roles at subscription or management-group scope
- Expressing an environment's full authorization graph in an Infra Chart

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `scope` | StringValueOrRef | Yes | ARM scope of the grant (defaults to an AzureResourceGroup reference) |
| `role_definition_name` | string | one-of | Built-in role name, e.g. "Reader" |
| `role_definition_id` | StringValueOrRef | one-of | Fully-scoped role definition ID (defaults to an AzureRoleDefinition reference — the composed custom-role path; literals bind existing definitions) |
| `principal_id` | StringValueOrRef | Yes | Azure AD OBJECT ID of the grantee (defaults to an AzureUserAssignedIdentity reference) |
| `principal_type` | enum | No | SERVICE_PRINCIPAL, USER, or GROUP; inferred by Azure when omitted |
| `description` | string | No | Audit note recorded on the assignment |
| `condition` | string | No | Azure ABAC condition expression |
| `condition_version` | string | No | "1.0" or "2.0" (only with `condition`; Azure defaults to "2.0") |
| `delegated_managed_identity_resource_id` | string | No | Cross-tenant (Lighthouse) delegation only |
| `skip_service_principal_aad_check` | bool | No | Skip the AAD existence check for freshly created service principals |
| `name` | string | No | Pinned GUID for the assignment's ARM resource name |

Exactly one of `role_definition_name` / `role_definition_id` must be set.
Role assignments do not support ARM tags, so the spec carries none.

## Outputs

| Output | Description |
|--------|-------------|
| `role_assignment_id` | Fully-scoped ARM ID of the assignment |
| `name` | The assignment's GUID resource name |
| `scope` | The scope the role was granted at |
| `role_definition_id` | The role definition ID Azure resolved (even when a role name was given) |
| `principal_id` | The principal's Azure AD object ID |
| `principal_type` | The principal type Azure recorded |

## Quick Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureRoleAssignment
metadata:
  name: ci-reader
  org: mycompany
  env: production
spec:
  scope:
    valueFrom:
      name: platform-rg
  roleDefinitionName: Reader
  principalId:
    valueFrom:
      name: ci-deploy-identity
```

The `scope` and `principalId` references resolve through their default kinds
(`AzureResourceGroup` and `AzureUserAssignedIdentity`); an explicit
`kind`/`fieldPath` in `valueFrom` targets any other resource, e.g. a Key Vault:

```yaml
spec:
  scope:
    valueFrom:
      kind: AzureKeyVault
      name: platform-kv
      fieldPath: status.outputs.key_vault_id
  roleDefinitionName: Key Vault Secrets User
  principalId:
    valueFrom:
      name: app-identity
```

## Required Permissions

See [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege action set the deploying principal needs.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
