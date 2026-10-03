# Azure Role Definition

Deploys a custom Azure RBAC role: a named, reusable bundle of permissions that principals can then be granted through role assignments. Azure ships hundreds of built-in roles, but real organizations routinely need permission sets the built-ins don't express — "Contributor, except role assignments and policy writes", "can restart VMs but not create or delete them", "read-only plus blob data access". A custom role captures such a set once, with a meaningful name, and every grant of it stays consistent as the definition evolves: updating the definition's permissions updates what every existing assignment of it allows.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Custom Role Definition** -- one tenant-visible RBAC role at the chosen scope (management group, subscription, or resource group), with its permission blocks and assignable scopes

A definition grants nothing by itself -- permissions only take effect when an **AzureRoleAssignment** binds the role to a principal at a scope. The definition's `role_definition_id` output is exactly what an assignment's `roleDefinitionId` field consumes.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **Authorization-plane rights on the target scope** -- creating role definitions requires `Microsoft.Authorization/roleDefinitions/write`, held via Owner, User Access Administrator, or Role Based Access Control Administrator. A Contributor-level deploying credential fails with "AuthorizationFailed".

## Deploy

### Console

Open the deployment store, find **Azure Role Definition**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields -- with quick-add templates for the classic permission-block shapes. Start from the **Explicit-Actions Operator Role** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureRoleDefinition
metadata:
  name: acme-vm-operator
  org: acme-corp
  env: prod
spec:
  name: acme-vm-operator
  scope:
    value: "/subscriptions/00000000-0000-0000-0000-000000000000"
  description: "Operate existing VMs: start/stop/restart, no create or delete"
  permissions:
    - actions:
        - Microsoft.Compute/virtualMachines/read
        - Microsoft.Compute/virtualMachines/start/action
        - Microsoft.Compute/virtualMachines/restart/action
        - Microsoft.Compute/virtualMachines/powerOff/action
        - Microsoft.Compute/virtualMachines/deallocate/action
```

```shell
planton apply -f role.yaml
```

This creates a subscription-scoped custom role that can start, stop, restart, and deallocate existing VMs but never create or delete them; assignable scopes are omitted, so Azure defaults them to the definition's own scope. An Infra Job tracks the provisioning in real time.

### InfraChart

Compose the definition with the assignments that grant it -- the assignment consumes the definition's output:

```yaml
# The custom role
spec:
  name: acme-blob-reader
  scope:
    valueFrom:
      kind: AzureResourceGroup
      name: data-rg
      fieldPath: status.outputs.resource_group_id
  permissions:
    - dataActions:
        - Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read
---
# The grant (separate AzureRoleAssignment resource)
spec:
  scope:
    valueFrom:
      kind: AzureResourceGroup
      name: data-rg
      fieldPath: status.outputs.resource_group_id
  roleDefinitionId:
    valueFrom:
      kind: AzureRoleDefinition
      name: acme-blob-reader
      fieldPath: status.outputs.role_definition_id
  principalId:
    valueFrom:
      kind: AzureUserAssignedIdentity
      name: analytics-reader
      fieldPath: status.outputs.principal_id
```

The InfraPipeline resolves the dependency graph: the custom role exists before the grant that binds it.

## Key Configuration

These are the most important decisions when configuring a custom role. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Control plane vs data plane** -- THE distinction to get right. `actions` govern managing resources through ARM ("restart a VM"); `dataActions` govern the data inside them ("read this blob", "read a secret's value in RBAC mode"). A role with `Microsoft.Storage/storageAccounts/*` actions can manage the account yet cannot read one byte of blob data.

**Carve-outs are trims, not denies** -- `notActions`/`notDataActions` subtract from THIS role's own grant; they cannot take away permissions another assignment gives the same principal (that is Azure deny assignments' job). The classic shape: actions `["*"]`, notActions `["Microsoft.Authorization/*/write"]` -- everything except changing RBAC.

**Scope and assignability** -- The definition's scope anchors its ARM ID (ForceNew) and defaults where it can be assigned. Choose the highest scope the role will ever be assigned at; narrow `assignableScopes` below it to pre-authorize where the role may ever be granted. At most one management group may appear in the list.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `scope`, `assignableScopes` | `status.outputs.resource_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `role_definition_id` | Fully-scoped ARM ID of the definition | AzureRoleAssignment's `roleDefinitionId` field -- the reference that binds this custom role to a principal |
| `role_definition_guid` | The definition's GUID resource name | Identifying the definition in authorization-API automation, since assignments track roles by GUID |
| `assignable_scopes` | The scopes Azure recorded -- carries the provider-defaulted own scope when the spec omitted the field | Knowing where grants of this role may exist without re-reading the spec |

The `role_name` and `scope` outputs echo the definition's coordinates as deployed for portal cross-reference and audit tooling; they are not typically wired into other Infra Components.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Operational roles** -- Verb-scoped control-plane roles ("operate, don't create"): the VM operator pattern. Start from the **Explicit-Actions Operator Role** preset.

**Data-access roles** -- Pure data-plane grants with zero management surface: the blob-data-reader pattern. Start from the **Data-Plane Role (Blob Auditor)** preset.

**Admin-minus-carve-out** -- Broad grants with the RBAC-write carve-out, keeping permission changes in the hands of the few. Start from the **Broad Grant with Carve-Outs and Assignable-Scope Control** preset.

## Works With

- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- binds this role to a principal at a scope, consuming `role_definition_id`
- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- a common definition scope and assignable-scope boundary
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- the principal most grants of custom roles target
