# Azure Storage Share

Deploys an Azure Files share inside an Azure Storage Account -- the SMB/NFS file system unit of Azure storage. VMs, AKS pods, and container apps mount shares for shared POSIX-style state (lift-and-shift app data, user profiles, CI caches, shared content), and Azure bills, throttles, tiers, and snapshots at the share level. Shares are many-per-account with independent lifecycles, which is why they are a first-class kind referencing the account rather than a list folded into the account's spec.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Azure Files Share** -- a share on the referenced storage account (by ARM ID -- the control-plane path), with your chosen provisioned quota, protocol, performance tier, stored access policies, and data-plane metadata

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureStorageAccount** the share will live in, referenced through `storageAccountId`. The parent is fixed at creation, and its KIND gates what the share can be:
  - **NFS shares and the PREMIUM tier** need a FileStorage (premium file) account.
  - **Quotas above 5120 GB** on standard accounts need the account's `largeFileShareEnabled` (up to 102400 GB).

## Deploy

### Console

Open the deployment store, find **Azure Storage Share**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Team File Share** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureStorageShare
metadata:
  name: team-files
  org: acme-corp
  env: prod
spec:
  storageAccountId:
    valueFrom:
      kind: AzureStorageAccount
      name: app-files
      fieldPath: status.outputs.storage_account_id
  shareName: team-files
  quotaGb: 500
```

```shell
planton apply -f share.yaml
```

This creates an SMB share -- what Windows mounts natively and Linux mounts via cifs -- with a 500 GB provisioned quota. An Infra Job tracks the provisioning in real time.

### InfraChart

When the account and share deploy in the same InfraChart, wire the account reference with ValueFromRef:

```yaml
spec:
  storageAccountId:
    valueFrom:
      kind: AzureStorageAccount
      name: app-files
      fieldPath: status.outputs.storage_account_id
  shareName: team-files
  quotaGb: 500
```

The InfraPipeline resolves the dependency graph, deploys the storage account first, then provisions the share with the resolved ARM ID.

## Key Configuration

These are the most important decisions when configuring a share. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Provisioned quota** -- `quotaGb` is the drive size SMB clients see and the write ceiling Azure enforces. It grows in place; shrinking below used capacity fails. On premium FileStorage accounts the quota is the billing: provisioned capacity is paid whether used or not (minimum 100 GB).

**Protocol** -- `enabledProtocol` is fixed at creation. Unspecified means SMB, which works on every account kind. NFS v4.1 brings POSIX semantics (hard links, chmod) for Linux workloads -- it requires a premium FileStorage account and is reachable only over private network paths.

**Access tier** -- `accessTier` trades at-rest cost against per-operation cost: TransactionOptimized (the standard-account default), HOT, COOL, or PREMIUM (required -- and the only legal value -- on FileStorage accounts). Edits in place.

**Stored access policies** -- `acls` (max 5, Azure's limit) anchor shared-access-signature tokens: revoking or shortening a policy immediately revokes every SAS issued against it -- the operational reason to prefer policy-anchored SAS over ad-hoc SAS. Permission letters follow Azure's strict order: r, w, d, l.

**Metadata** -- `metadata` stores free-form key/value pairs on the share (visible to anyone who can read share properties -- not for secrets). Keys must be valid C# identifiers; Azure lowercases them on read.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureStorageAccount** | `storageAccountId` | `status.outputs.storage_account_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `share_id` | The management-plane ARM ID | ARM reads, policy targets |
| `rbac_scope_id` | The data-plane RBAC scope (`.../fileServices/default/fileshares/{name}` -- deliberately a DIFFERENT segment than the management ID) | Storage File Data SMB Share Reader/Contributor/Elevated Contributor role assignments |
| `share_name` | The share's name | Mount commands, CSI volume definitions, app settings |
| `storage_account_name` | The parent account's name, parsed from the resolved account ID | The account/share pair without a second reference |

There is deliberately NO URL output: the share's data-plane path is the ACCOUNT's file endpoint plus the share name, and only the account knows its real endpoint (partitioned-DNS accounts use a different hostname than the classic shared DNS). Compose mount paths from AzureStorageAccount's `primary_file_endpoint` output + `share_name`.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Team file share** -- SMB on a standard account with an honest quota; mounted by VMs and AKS pods via the Files CSI driver. Start from the **Team File Share** preset.

**NFS premium share** -- NFS v4.1 on a FileStorage account for Linux workloads that need POSIX fidelity; private network paths only. Start from the **NFS Premium Share** preset.

**Policy-anchored access** -- stored access policies anchoring revocable SAS tokens for partner exchange. Start from the **Policy-Anchored Access Share** preset.

## Works With

- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- the parent account and the source of the file endpoint mount paths compose from
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- share-scoped data-plane grants targeting `rbac_scope_id`
