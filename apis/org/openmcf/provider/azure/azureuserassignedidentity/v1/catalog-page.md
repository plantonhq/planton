# Azure User Assigned Identity

Deploys an Azure User-Assigned Managed Identity with optional RBAC role assignments. The component creates the identity resource and binds it to specified roles at specified scopes, providing a credential-free authentication mechanism that can be shared across multiple Azure resources.

## What Gets Created

When you deploy an AzureUserAssignedIdentity resource, OpenMCF provisions:

- **User-Assigned Managed Identity** — an `authorization.UserAssignedIdentity` resource in the specified region and resource group, tagged with resource metadata for tracking and governance
- **RBAC Role Assignments** — an `authorization.Assignment` for each entry in `roleAssignments`, binding the identity's principal to a specific role at a specific Azure resource scope
- **Azure Tags** — resource metadata tags applied to the identity including resource name, kind, organization, and environment

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the identity will be created (can reference an AzureResourceGroup resource)
- **Azure resource IDs** for any scopes you want to grant role assignments on (Key Vault IDs, Storage Account IDs, subscription IDs, etc.)

## Quick Start

Create a file `identity.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureUserAssignedIdentity
metadata:
  name: my-identity
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureUserAssignedIdentity.my-identity
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-identity
```

Deploy:

```shell
openmcf apply -f identity.yaml
```

This creates a User-Assigned Managed Identity with no role assignments. The identity exists but has no permissions until role assignments are added.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the Managed Identity (e.g., `eastus`, `westeurope`). Must match the region of the resource group. | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `name` | `string` | Name of the User-Assigned Managed Identity. Must be unique within the resource group. | Required, 3-128 characters: alphanumeric, hyphens, underscores |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `roleAssignments` | `RoleAssignment[]` | `[]` | RBAC role assignments granting this identity permissions on Azure resources. Each entry has a `scope` (StringValueOrRef) and a `roleDefinitionName` (string). |
| `roleAssignments[].scope` | `StringValueOrRef` | — | Azure resource ID to scope the role assignment to. Can be a subscription, resource group, or individual resource ID. Supports `valueFrom` to reference another resource's output. |
| `roleAssignments[].roleDefinitionName` | `string` | — | Name of a built-in or custom Azure RBAC role (e.g., `Key Vault Secrets User`, `Storage Blob Data Contributor`, `AcrPull`, `Network Contributor`). |

## Examples

### Identity with No Role Assignments

A bare identity for development or testing. It has no permissions and cannot access any Azure resources until role assignments are added:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureUserAssignedIdentity
metadata:
  name: dev-identity
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureUserAssignedIdentity.dev-identity
spec:
  region: eastus
  resourceGroup: dev-rg
  name: dev-identity
```

### Identity with Key Vault Access

An identity granted read access to Key Vault secrets, suitable for applications that need to retrieve configuration secrets:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureUserAssignedIdentity
metadata:
  name: app-identity
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureUserAssignedIdentity.app-identity
spec:
  region: eastus
  resourceGroup: prod-rg
  name: app-identity
  roleAssignments:
    - scope: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.KeyVault/vaults/prod-vault
      roleDefinitionName: Key Vault Secrets User
```

### Identity with Multiple Role Assignments

An identity for a workload that needs access to several Azure services -- Key Vault for secrets, a Storage Account for blob data, and an ACR for pulling container images:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureUserAssignedIdentity
metadata:
  name: workload-identity
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureUserAssignedIdentity.workload-identity
spec:
  region: westeurope
  resourceGroup: prod-rg
  name: workload-identity
  roleAssignments:
    - scope: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.KeyVault/vaults/prod-vault
      roleDefinitionName: Key Vault Secrets User
    - scope: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Storage/storageAccounts/prodstorage
      roleDefinitionName: Storage Blob Data Contributor
    - scope: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.ContainerRegistry/registries/prodacr
      roleDefinitionName: AcrPull
```

### Using Foreign Key References for Scope

Reference OpenMCF-managed resources instead of hardcoding Azure resource IDs. The `valueFrom` syntax resolves the scope from another resource's stack outputs:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureUserAssignedIdentity
metadata:
  name: ref-identity
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureUserAssignedIdentity.ref-identity
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: platform-rg
      field: status.outputs.resource_group_name
  name: ref-identity
  roleAssignments:
    - scope:
        valueFrom:
          kind: AzureKeyVault
          name: platform-kv
          fieldPath: status.outputs.vault_id
      roleDefinitionName: Key Vault Secrets User
    - scope:
        valueFrom:
          kind: AzureStorageAccount
          name: platform-storage
          fieldPath: status.outputs.storage_account_id
      roleDefinitionName: Storage Blob Data Contributor
```

### Subscription-Level Contributor

An identity with Contributor access at the subscription level, useful for automation or deployment pipelines that manage resources across resource groups:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureUserAssignedIdentity
metadata:
  name: deployer-identity
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureUserAssignedIdentity.deployer-identity
spec:
  region: eastus
  resourceGroup: infra-rg
  name: deployer-identity
  roleAssignments:
    - scope: /subscriptions/00000000-0000-0000-0000-000000000000
      roleDefinitionName: Contributor
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `identity_id` | `string` | Azure Resource Manager ID of the User-Assigned Managed Identity. Format: `/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.ManagedIdentity/userAssignedIdentities/{name}` |
| `principal_id` | `string` | Service Principal Object ID in Azure AD. Used for RBAC role assignments and Key Vault access policies that accept object IDs. |
| `client_id` | `string` | Client ID (Application ID) of the Managed Identity. Applications use this to authenticate as the identity via the Azure SDK (e.g., set as `AZURE_CLIENT_ID`). |
| `tenant_id` | `string` | Azure AD Tenant ID that the Managed Identity belongs to. Used with `client_id` for cross-tenant or multi-tenant scenarios. |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) -- provides the resource group where the identity is created
- [AzureKeyVault](/docs/catalog/azure/azurekeyvault) -- a common scope target for granting secret, key, or certificate access
- [AzureStorageAccount](/docs/catalog/azure/azurestorageaccount) -- a common scope target for granting blob, queue, or table data access
- [AzureAksCluster](/docs/catalog/azure/azureakscluster) -- AKS clusters can use user-assigned identities for kubelet or control plane authentication
