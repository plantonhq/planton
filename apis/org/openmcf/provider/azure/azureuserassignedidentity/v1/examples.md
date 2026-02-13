# AzureUserAssignedIdentity Examples

## Minimal Configuration

The simplest possible identity -- no role assignments. Useful when roles will be
assigned by other automation or manually.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureUserAssignedIdentity
metadata:
  name: basic-identity
spec:
  region: eastus
  resource_group: my-resource-group
  name: basic-managed-identity
```

## Identity with Key Vault Access

A common pattern: create an identity that can read secrets from Azure Key Vault.
Used by Function Apps, Web Apps, and Container Apps for secure secret access.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureUserAssignedIdentity
metadata:
  name: app-identity
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: prod-identity-rg
  name: prod-app-identity
  role_assignments:
    - scope: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.KeyVault/vaults/prod-secrets-kv
      role_definition_name: Key Vault Secrets User
```

## Identity with Multiple Role Assignments

An identity for an application that needs access to Key Vault (secrets), Storage
(blobs), and Container Registry (pull images).

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureUserAssignedIdentity
metadata:
  name: multi-role-identity
  org: mycompany
  env: production
spec:
  region: westeurope
  resource_group: prod-identity-rg
  name: prod-multi-role-identity
  role_assignments:
    - scope: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.KeyVault/vaults/prod-kv
      role_definition_name: Key Vault Secrets User
    - scope: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Storage/storageAccounts/prodstorage
      role_definition_name: Storage Blob Data Contributor
    - scope: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.ContainerRegistry/registries/prodacr
      role_definition_name: AcrPull
```

## AKS Network Identity

An identity used by AKS for managing network resources. Scoped to the VNet resource
so AKS can create load balancers and manage subnets.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureUserAssignedIdentity
metadata:
  name: aks-network-identity
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: prod-identity-rg
  name: prod-aks-network-identity
  role_assignments:
    - scope: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet
      role_definition_name: Network Contributor
```

## Infra Chart Wiring -- Full Identity Stack

This example demonstrates the primary use case: wiring an identity into an infra chart
where the role assignment scopes reference dynamically created resources.

### Resource Group (Layer 0)

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureResourceGroup
metadata:
  name: platform-rg
  org: mycompany
  env: production
spec:
  name: prod-platform-rg
  region: eastus
```

### Key Vault (Layer 1)

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureKeyVault
metadata:
  name: platform-kv
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: platform-rg
      fieldPath: status.outputs.resource_group_name
  name: prod-platform-kv
```

### User-Assigned Identity (Layer 1)

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
```

### How the Dependency Graph Works

1. OpenMCF creates the Resource Group first (no dependencies)
2. Key Vault and the Identity are both created next (both depend on RG)
3. The role assignment's `scope` resolves to the Key Vault's ID via `valueFrom`
4. The identity is created first, then its role assignments are applied
5. Downstream resources (AKS, Function Apps) can reference `identity_id`

## Subscription-Scoped Reader

An identity with subscription-wide read access. Useful for monitoring and
compliance tools that need to enumerate resources across the subscription.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureUserAssignedIdentity
metadata:
  name: reader-identity
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: shared-identity-rg
  name: subscription-reader-identity
  role_assignments:
    - scope: /subscriptions/00000000-0000-0000-0000-000000000000
      role_definition_name: Reader
```

## Best Practices

1. **Use user-assigned over system-assigned** -- User-assigned identities have
   independent lifecycles and can be shared. System-assigned identities are
   simpler but tied to a single resource.

2. **Least privilege** -- Grant only the specific roles needed. Prefer fine-grained
   roles like "Key Vault Secrets User" over broad roles like "Contributor".

3. **Scope narrowly** -- Assign roles at the most specific scope possible. Prefer
   resource-level scopes over resource-group or subscription scopes.

4. **One identity per concern** -- In enterprise setups, create separate identities
   for different concerns (network, secrets, storage) rather than one identity
   with all permissions.

5. **Use StringValueOrRef in infra charts** -- Always reference resource IDs
   dynamically via `valueFrom` in infra charts instead of hardcoding Azure
   resource IDs. This ensures proper dependency ordering and portability.
