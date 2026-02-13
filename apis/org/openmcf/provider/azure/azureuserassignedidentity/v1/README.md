# AzureUserAssignedIdentity

## Overview

`AzureUserAssignedIdentity` provisions an Azure User-Assigned Managed Identity with
RBAC role assignments -- Azure's recommended approach for granting Azure resources
secure, credential-free access to other Azure services.

Unlike system-assigned identities (which are tied to a single resource and deleted
when that resource is deleted), user-assigned identities have an independent lifecycle
and can be shared across multiple Azure resources. This makes them ideal for scenarios
where the same set of permissions is needed by multiple resources, or where the
identity must survive resource recreation.

This component follows the same pattern as `AwsIamRole` and `GcpServiceAccount`,
adapted for Azure's scope-based RBAC model where each role assignment targets a
specific Azure resource.

## Key Features

- **StringValueOrRef resource_group** -- references an `AzureResourceGroup` output,
  enabling proper dependency wiring in infra charts
- **StringValueOrRef scope** in role assignments -- references any Azure resource's
  output for dynamic RBAC scoping in infra charts
- **Bundled role assignments** -- identity + permissions defined together because an
  identity without roles has no practical value (per DD03)
- **skip_service_principal_aad_check** -- automatically set on role assignments to
  avoid race conditions with Azure AD replication

## When to Use

- Before deploying AKS clusters, Function Apps, Web Apps, or Container Apps that
  need to access Azure services (Key Vault, Storage, ACR, etc.)
- As the identity layer in aks-environment (enhanced), function-app-environment,
  and web-app-environment infra charts
- When multiple Azure resources need to share the same set of permissions
- When you need an identity that outlives the resources it is attached to

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `region` | string | Yes | Azure region |
| `resource_group` | StringValueOrRef | Yes | Resource group (literal or valueFrom) |
| `name` | string | Yes | Identity name (3-128 chars) |
| `role_assignments` | repeated RoleAssignment | No | RBAC role bindings |

### RoleAssignment

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `scope` | StringValueOrRef | Yes | Azure resource ID to scope the role to |
| `role_definition_name` | string | Yes | Azure built-in or custom role name |

## Outputs

| Output | Description |
|--------|-------------|
| `identity_id` | Azure Resource Manager ID of the identity |
| `principal_id` | Service Principal Object ID (for RBAC and access policies) |
| `client_id` | Application/Client ID (for SDK configuration) |
| `tenant_id` | Azure AD Tenant ID |

## Quick Example

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureUserAssignedIdentity
metadata:
  name: platform-identity
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: platform-rg
      fieldPath: status.outputs.resource_group_name
  name: prod-platform-identity
  role_assignments:
    - scope:
        valueFrom:
          kind: AzureKeyVault
          name: platform-kv
          fieldPath: status.outputs.key_vault_id
      role_definition_name: Key Vault Secrets User
    - scope:
        valueFrom:
          kind: AzureStorageAccount
          name: platform-storage
          fieldPath: status.outputs.storage_account_id
      role_definition_name: Storage Blob Data Contributor
```

## Downstream Resources

Resources that reference this identity:

- **AzureAksCluster** -- `identity_id` for kubelet or control plane identity
- **AzureFunctionApp** -- `identity_id` for managed identity assignment
- **AzureLinuxWebApp** -- `identity_id` for managed identity assignment
- **AzureContainerApp** -- `identity_id` for managed identity assignment

## References

- [Azure Managed Identities overview](https://learn.microsoft.com/en-us/entra/identity/managed-identities-azure-resources/overview)
- [Azure RBAC built-in roles](https://learn.microsoft.com/en-us/azure/role-based-access-control/built-in-roles)
- Research documentation: [docs/README.md](docs/README.md)
- Examples: [examples.md](examples.md)
