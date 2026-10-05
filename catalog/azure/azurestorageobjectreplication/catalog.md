# Azure Storage Object Replication

Deploys an object replication policy between TWO Azure Storage Accounts -- asynchronous, rule-driven copying of block blobs from containers on a source account to containers on a destination account. This is the storage-level answer to cross-region DR (replicate to an account in a paired region), data distribution (fan out to a read-local copy), and tenant offboarding or archival -- without any application-side copy jobs. One policy spans exactly one account pair; Azure materializes it on BOTH accounts (the destination holds the authoritative copy that assigns rule IDs, the source holds the mirror), which the IaC modules handle as one unit -- this kind IS the pair.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Object Replication Policy (both sides)** -- the destination-side policy (authoritative) and the source-side mirror, with your container-to-container rules, backfill choices, and prefix filters

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **Two AzureStorageAccounts**, both prepared for replication: blob versioning AND change feed enabled on the SOURCE (the account spec's `blobProperties`), blob versioning on the DESTINATION -- Azure rejects the policy at apply time otherwise. Versioning rules out hierarchical-namespace (ADLS Gen2) accounts.
- **The containers on both sides** each rule names -- they must live on the policy's respective accounts.

## Deploy

### Console

Open the deployment store, find **Azure Storage Object Replication**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Cross-Region DR Replication** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureStorageObjectReplication
metadata:
  name: invoices-dr
  org: acme-corp
  env: prod
spec:
  sourceStorageAccountId:
    valueFrom:
      kind: AzureStorageAccount
      name: primary-account
      fieldPath: status.outputs.storage_account_id
  destinationStorageAccountId:
    valueFrom:
      kind: AzureStorageAccount
      name: dr-account
      fieldPath: status.outputs.storage_account_id
  rules:
    - sourceContainerName:
        valueFrom:
          kind: AzureStorageContainer
          name: invoices-container
          fieldPath: status.outputs.container_name
      destinationContainerName:
        valueFrom:
          kind: AzureStorageContainer
          name: invoices-replica-container
          fieldPath: status.outputs.container_name
      copyBlobsCreatedAfter: Everything
```

```shell
planton apply -f replication.yaml
```

This bootstraps DR for one container: the whole container backfills once, then new blobs stream asynchronously. An Infra Job tracks the provisioning in real time.

### InfraChart

When both accounts, both containers, and the policy deploy in the same InfraChart, wire every reference with ValueFromRef:

```yaml
spec:
  sourceStorageAccountId:
    valueFrom:
      kind: AzureStorageAccount
      name: primary-account
      fieldPath: status.outputs.storage_account_id
  destinationStorageAccountId:
    valueFrom:
      kind: AzureStorageAccount
      name: dr-account
      fieldPath: status.outputs.storage_account_id
  rules:
    - sourceContainerName:
        valueFrom:
          kind: AzureStorageContainer
          name: invoices-container
          fieldPath: status.outputs.container_name
      destinationContainerName:
        valueFrom:
          kind: AzureStorageContainer
          name: invoices-replica-container
          fieldPath: status.outputs.container_name
```

The InfraPipeline resolves the dependency graph and deploys the accounts, then the containers, then the policy with the resolved values.

## Key Configuration

These are the most important decisions when configuring a replication policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The account pair** -- `sourceStorageAccountId` and `destinationStorageAccountId` are both fixed at creation: one policy spans exactly one pair. A paired-region destination is the DR pattern; same-region is the distribution/archival pattern.

**Rules** -- each rule maps ONE source container to ONE destination container; a policy carries 1-1000 rules and a container pair appears in at most one rule.

**Backfill** -- per-rule `copyBlobsCreatedAfter` decides what existing data joins the copy: `OnlyNewObjects` (the default -- no backfill), `Everything` (the whole container -- takes time proportional to its size), or an RFC 3339 UTC instant (blobs created after that moment).

**Prefix filters** -- per-rule `prefixMatch` narrows replication to blobs whose names start with one of the prefixes (an INCLUDE filter, ARM's own semantics). Empty means every blob.

**Honesty about RPO** -- replication is asynchronous with NO RPO guarantee unless the accounts opt into replication metrics and you monitor them.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureStorageAccount** | `sourceStorageAccountId`, `destinationStorageAccountId` | `status.outputs.storage_account_id` |
| **AzureStorageContainer** | per-rule `sourceContainerName`, `destinationContainerName` | `status.outputs.container_name` |

### What This Kind Provides

This kind has no consumable outputs: nothing downstream composes with a replication policy, so no output has a ValueFromRef consumer. `status.outputs` still carries what operations needs -- `policy_id` (the server-assigned GUID shared by both sides, what `az storage account or-policy show --policy-id` and the monitoring surfaces key on) and the two ARM IDs (`source_object_replication_id`, `destination_object_replication_id`), because one policy materializes as TWO ARM resources, one on each account.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Cross-region DR** -- destination in the paired region, `Everything` backfill on the critical containers; the initial copy runs long, then steady-state replication takes over. Start from the **Cross-Region DR Replication** preset.

**Prefix-scoped distribution** -- fan out only a name subtree (`reports/2026`) to a read-local copy in each serving geography. Start from the **Prefix-Scoped Content Distribution** preset.

## Works With

- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- both ends of the pair; the source needs versioning + change feed, the destination versioning
- [**Azure Storage Container**](/infra-catalog/azure-storage-container) -- the per-rule source and destination containers, referenced by name
