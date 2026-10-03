# Azure Event Hub Namespace Customer Managed Key

Configures customer-managed-key (BYOK) encryption on an Event Hubs namespace: event data at rest is encrypted with YOUR Key Vault keys instead of Microsoft-managed keys. Azure models CMK as a configuration applied ONTO an existing namespace -- and this kind mirrors that grain, for a real reason: encrypting with the namespace's own system-assigned identity is only possible as a second step (create the namespace with its identity, grant that identity wrap/unwrap on the vault, then apply this kind). A folded create-time block could never express that sequence. CMK is add-only by Azure's own contract: once enabled it can never be removed short of replacing the namespace itself.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **CMK encryption configuration** -- applied onto the referenced namespace, with your Key Vault keys (1-10) and optionally a second infrastructure-encryption layer

**Add-only lifecycle (Azure's own contract)**: once CMK is enabled it can never be removed -- Azure has no decrypt-back path. Deleting this resource intentionally changes NOTHING on the namespace (the delete is a no-op); returning to Microsoft-managed keys requires replacing the namespace itself. Key ROTATION is routine, though: with versionless key references, vault rotation propagates automatically.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A single-tenant AzureEventHubNamespace** -- CMK requires a namespace on a dedicated cluster (AzureEventHubCluster via `dedicatedClusterId`) or the PREMIUM tier. Multi-tenant BASIC/STANDARD namespaces share hardware and cannot take tenant keys; Azure rejects the encryption patch on them.
- **An AzureKeyVaultKey** in a vault with PURGE PROTECTION enabled -- reference its `versionless_id` output so rotation propagates automatically.
- **The unwrapping identity's vault grant** -- an AzureRoleAssignment of "Key Vault Crypto Service Encryption User" on the vault, targeting either the namespace's system-assigned identity (`identity_principal_id` output) or the user-assigned identity named here, BEFORE this kind applies.

## Deploy

### Console

Open the deployment store, find **Azure Event Hub Namespace Customer Managed Key**, and click **Deploy**. The creation wizard leads with the one-way-door and single-tenant contracts, walks the versionless key list, and teaches both unwrap-identity grant paths. Start from the **BYOK with Rotating Key** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventHubNamespaceCustomerManagedKey
metadata:
  name: telemetry-cmk
  org: acme-corp
  env: prod
spec:
  eventhubNamespaceId:
    valueFrom:
      kind: AzureEventHubNamespace
      name: telemetry-hubs-premium
      fieldPath: status.outputs.namespace_id
  keyVaultKeyIds:
    - valueFrom:
        kind: AzureKeyVaultKey
        name: streaming-key
        fieldPath: status.outputs.versionless_id
  userAssignedIdentityId:
    valueFrom:
      kind: AzureUserAssignedIdentity
      name: cmk-identity
      fieldPath: status.outputs.identity_id
```

```shell
planton apply -f cmk.yaml
```

This applies BYOK encryption to the `telemetry-hubs-premium` namespace with one versionless Key Vault key, unwrapped by the `cmk-identity` user-assigned identity. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the CMK configuration onto a namespace deployed in the same InfraPipeline:

```yaml
spec:
  eventhubNamespaceId:
    valueFrom:
      kind: AzureEventHubNamespace
      name: telemetry-hubs-premium
      fieldPath: status.outputs.namespace_id
```

The InfraPipeline resolves the dependency graph, deploying the namespace (and the identity's vault grant) before this configuration applies.

## Key Configuration

These are the most important decisions when configuring CMK. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The keys** -- `keyVaultKeyIds` (1-10) defaults to referencing AzureKeyVaultKey `versionless_id` outputs, so vault-side rotation propagates automatically with no edit here. Pin versioned IDs only when a compliance regime demands immutable key versions -- and own the manual re-pin at every rotation. The keys' vault must have purge protection enabled.

**The unwrap identity** -- unset uses the namespace's SYSTEM-ASSIGNED identity (grant it the vault access via its `identity_principal_id` output). Set `userAssignedIdentityId` for the fleet pattern: one CMK identity with a standing vault grant, attached to every namespace at creation -- the identity must already ride the namespace's identity block with vault access when this kind applies.

**Infrastructure encryption** -- `infrastructureEncryptionEnabled` adds a second, independent encryption layer beneath your keys. It is fixed the moment CMK is first configured; decide it now.

**What can change later** -- `eventhubNamespaceId` is fixed at creation (the configuration is bound to its namespace for life), while the key list and the unwrap identity edit in place. And remember the add-only lifecycle: deleting this resource changes nothing on the namespace, and there is no path back to Microsoft-managed keys short of replacing the namespace.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureEventHubNamespace** | `eventhubNamespaceId` | `status.outputs.namespace_id` |
| **AzureKeyVaultKey** | `keyVaultKeyIds[]` | `status.outputs.versionless_id` |
| **AzureUserAssignedIdentity** | `userAssignedIdentityId` | `status.outputs.identity_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `customer_managed_key_id` | Azure Resource Manager ID of the CMK configuration | Governance tooling and ARM-level references |

This is deliberately the only output: the kind configures encryption ON the namespace and mints nothing consumable -- applications keep connecting through the namespace and its authorization rules.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**BYOK with rotating key** -- the everyday shape: one versionless key rotated by vault policy, unwrapped by a fleet identity -- the rotation runbook is empty. Start from the **BYOK with Rotating Key** preset.

**Double encryption** -- regulated workloads add the infrastructure-encryption layer beneath the customer keys at first configuration.

## Works With

- [**Azure Event Hub Namespace**](/infra-catalog/azure-event-hub-namespace) -- the single-tenant namespace being encrypted
- [**Azure Event Hub Cluster**](/infra-catalog/azure-event-hub-cluster) -- the dedicated capacity that qualifies a namespace for CMK (or PREMIUM)
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) -- the encryption keys, referenced versionless for automatic rotation
- [**Azure Key Vault**](/infra-catalog/azure-key-vault) -- must have purge protection enabled
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- the optional fleet unwrap identity
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- the "Key Vault Crypto Service Encryption User" grant on the vault
