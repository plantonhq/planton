# Azure Role Assignment

Deploys an Azure RBAC role assignment: the grant of a role to a principal at a scope. A role assignment is the atomic unit of authorization in Azure — everything a user, group, service principal, or managed identity is allowed to do is the sum of the role assignments that target it. Because grants are the most-repeated pattern in any Azure environment, this kind models them as first-class, composable nodes: one assignment per resource, referenceable in InfraCharts, with an independent lifecycle from both the principal and the scope it binds.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Role Assignment** -- one grant record binding a role to a principal at a scope, visible in the portal's IAM blade

The three coordinates of every assignment:

- **scope** -- WHERE the permission applies (management group, subscription, resource group, or an individual resource). Permissions inherit downward.
- **role** -- WHAT is permitted, referenced either by built-in role name (e.g. "Reader") or by role definition ID (custom roles).
- **principalId** -- WHO receives the permission (the Entra object ID of a user, group, service principal, or managed identity).

Azure role assignments are immutable: changing any field replaces the assignment (delete + create). This matches ARM's own model, where an assignment is an atomic grant record rather than a mutable object.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **Authorization-plane rights on the target scope** -- creating role assignments requires `Microsoft.Authorization/roleAssignments/write`, held via Owner, User Access Administrator, or Role Based Access Control Administrator. A deploy failing with "AuthorizationFailed" almost always means the deploying credential has Contributor-level rights but no authorization-plane rights at that scope.

## Deploy

### Console

Open the deployment store, find **Azure Role Assignment**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields -- with a curated built-in role catalog and the object-ID-vs-client-ID trap taught at the moment of choice. Start from the **Managed Identity Grant on a Resource Group** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureRoleAssignment
metadata:
  name: ci-acr-pull
  org: acme-corp
  env: prod
spec:
  scope:
    valueFrom:
      kind: AzureResourceGroup
      name: registry-rg
      fieldPath: status.outputs.resource_group_id
  roleDefinitionName: "AcrPull"
  principalId:
    valueFrom:
      kind: AzureUserAssignedIdentity
      name: ci-deployer
      fieldPath: status.outputs.principal_id
  description: "CI deploy identity pulls from the shared registry"
  skipServicePrincipalAadCheck: true
```

```shell
planton apply -f grant.yaml
```

This grants the built-in `AcrPull` role to the `ci-deployer` managed identity across everything in the `registry-rg` resource group -- the `skipServicePrincipalAadCheck` flag lets the grant deploy in the same pipeline run that creates the identity, since Entra replicates new principals asynchronously and an assignment racing that replication would otherwise fail with "PrincipalNotFound". An Infra Job tracks the provisioning in real time.

### InfraChart

The scope is genuinely polymorphic — any resource's ID output is a valid grant boundary:

```yaml
spec:
  scope:
    valueFrom:
      kind: AzureKeyVault
      name: platform-kv
      fieldPath: status.outputs.key_vault_id
  roleDefinitionName: "Key Vault Crypto Service Encryption User"
  principalId:
    valueFrom:
      kind: AzureDiskEncryptionSet
      name: prod-des
      fieldPath: status.outputs.identity_principal_id
  skipServicePrincipalAadCheck: true
```

This is the grant that completes the disk-encryption chain: the set's system-assigned principal gets exactly the wrap/unwrap role on exactly the vault holding the key.

## Key Configuration

These are the most important decisions when configuring a role assignment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Object ID, not client ID** -- The most common role-assignment mistake: an assignment created with an application (client) ID SUCCEEDS and grants nothing, because no directory object carries that object ID. Referencing an AzureUserAssignedIdentity's `principal_id` output makes the mistake unrepresentable.

**Exactly one way to name the role** -- Built-in roles by name (matched case-insensitively against Azure's catalog); custom roles by fully-scoped definition ID (from an AzureRoleDefinition's `role_definition_id` output). Never both.

**Least-privilege scope** -- Permissions inherit downward, so grant at the narrowest scope that satisfies the use case: the specific vault, not the subscription.

**ABAC conditions** -- An optional condition expression narrows WHEN the role's permissions apply (e.g. only blobs carrying a specific tag). Supported on storage data-plane roles and a growing set of others.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** (default scope) | `scope` | `status.outputs.resource_group_id` |
| **Any Azure resource** (polymorphic scope) | `scope` | the resource's ID output |
| **AzureUserAssignedIdentity** (default principal) | `principalId` | `status.outputs.principal_id` |
| **AzureRoleDefinition** (custom roles) | `roleDefinitionId` | `status.outputs.role_definition_id` (copy the resolved ID) |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `role_assignment_id` | Fully-scoped ARM ID of the assignment | The handle the authorization API uses to fetch or delete the grant -- audit automation keys on it |
| `role_definition_id` | The definition Azure actually bound -- resolved to an ID even when the spec named a built-in role | Knowing exactly which definition matched a case-insensitive role name |
| `principal_type` | The principal type Azure recorded (User, Group, ServicePrincipal) | Confirming what the directory inferred when the spec omitted `principalType` |

The remaining outputs (`name`, `scope`, `principal_id`) echo the grant's coordinates as recorded at deploy time so audit tooling can reason about the assignment without re-reading the spec; they are not typically wired into other Infra Components.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**The identity-grant triad** -- An AzureUserAssignedIdentity plus one assignment per resource it touches, each with `skipServicePrincipalAadCheck: true` so identity and grants deploy in one pipeline run. Start from the **Managed Identity Grant on a Resource Group** preset.

**Custom-role grants** -- An AzureRoleDefinition captures the permission set; assignments bind it by definition ID. Start from the **Custom Role at Subscription Scope** preset.

**Tag-conditioned data access** -- A data-plane role narrowed by an ABAC condition on resource tags. Start from the **ABAC-Conditioned Data Grant** preset.

## Works With

- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- the principal most grants target, via `principal_id`
- [**Azure Role Definition**](/infra-catalog/azure-role-definition) -- custom roles bound by `role_definition_id`
- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- the most common grant boundary
- [**Azure Key Vault**](/infra-catalog/azure-key-vault) -- the classic narrow-scope grant target for CMK and secret consumers
