# Azure Storage Data Lake Gen2 Filesystem

Deploys a Data Lake Storage Gen2 filesystem inside an Azure Storage Account -- the namespace unit of an analytics data lake. A filesystem is where hierarchical-namespace (HNS) data lives: Spark, Databricks, Synapse, and the abfss:// driver all address data as `abfss://{filesystem}@{account}.dfs.core.windows.net/path`. Analytics estates conventionally provision one filesystem per data-lake zone (raw, curated, gold) so each zone carries its own access control. Filesystems are many-per-account with independent lifecycles and are the grant boundary for data-plane RBAC and POSIX ACLs, which is why they are a first-class kind rather than a list folded into the account's spec.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Data Lake Gen2 Filesystem** -- a filesystem on the referenced storage account (by ARM ID -- the control-plane path), with optional default encryption scope, root ownership (owner and owning group), the root path's POSIX ACL, and free-form properties

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureStorageAccount** the filesystem will live in, referenced through `storageAccountId`. The parent is fixed at creation. The account should carry `isHnsEnabled: true` -- Azure rejects POSIX owner, group, and ACL settings on flat-namespace accounts at apply time. A filesystem on a flat account is just a blob container wearing a dfs endpoint; create it as an AzureStorageContainer instead.
- **For a default encryption scope**: an AzureStorageEncryptionScope on the SAME account.
- **For identity-owned zones**: an AzureUserAssignedIdentity whose `principal_id` output plays the owner, group, or ACL-principal role.

## Deploy

### Console

Open the deployment store, find **Azure Storage Data Lake Gen2 Filesystem**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Data Lake Zone Filesystem** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureStorageDataLakeGen2Filesystem
metadata:
  name: raw-zone
  org: acme-corp
  env: prod
spec:
  storageAccountId:
    valueFrom:
      kind: AzureStorageAccount
      name: datalake-account
      fieldPath: status.outputs.storage_account_id
  filesystemName: raw
```

```shell
planton apply -f filesystem.yaml
```

This creates the `raw` filesystem on the `datalake-account` account with Azure's default ownership and no ACL -- engines address it as `abfss://raw@{account}.dfs.core.windows.net/`. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the ValueFromRef above wires the filesystem to its account: the InfraPipeline resolves the dependency graph, deploys the HNS account first, then provisions the filesystem with the resolved ARM ID.

## Key Configuration

These are the most important decisions when configuring a filesystem. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Filesystem name** -- `filesystemName` becomes the container segment of every abfss:// and dfs URL. 3-63 lowercase letters, digits, and hyphens, not starting with a hyphen ($root is the one special name). Renaming replaces the filesystem -- and everything stored in it.

**Encryption scope** -- `defaultEncryptionScope` applies sub-account key isolation to just this zone (e.g. a regulated zone under a customer-managed key while the rest of the lake uses platform keys); the scope must live on the same account. Fixed at creation.

**Root ownership** -- `owner` and `group` set the Entra principals that own the root path. The owning user always holds the root's user-class permissions regardless of the ACL. Each takes an object ID (a GUID -- for a managed identity the PRINCIPAL id, not the client id) or the special literal `$superuser`; reference an AzureUserAssignedIdentity's `principal_id` so a workload identity owns its zone.

**The root ACL** -- `aces` is the POSIX access control list on "/" -- the same rwx model as a Linux filesystem, evaluated per request against the caller's Entra identity. ACCESS entries gate the root itself; DEFAULT entries are the template newly created children inherit, which is how a zone's posture propagates to files landing in it. USER/GROUP entries may name a principal via `objectId`; MASK caps every named entry; OTHER covers unmatched callers. Permissions are the strict three-character form (`rwx`, `r-x`, `---`).

**Properties** -- free-form key/value pairs stored on the filesystem. Azure requires the VALUES to be base64-encoded (keys stay plain), e.g. `environment: cHJvZHVjdGlvbg==`. Visible to anyone who can read filesystem properties; not for secrets.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureStorageAccount** | `storageAccountId` | `status.outputs.storage_account_id` |
| **AzureStorageEncryptionScope** | `defaultEncryptionScope` | `status.outputs.encryption_scope_name` |
| **AzureUserAssignedIdentity** | `owner`, `group`, per-ACE `objectId` | `status.outputs.principal_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `filesystem_id` | Azure Resource Manager ID of the filesystem | Zone-scoped data-plane role assignments (Storage Blob Data Reader/Contributor) |
| `filesystem_name` | The filesystem's name | Spark/Databricks/Synapse abfss:// configuration, mount points |
| `storage_account_name` | The parent account's name, parsed from the resolved account ID | The account/filesystem pair without a second reference |

There is deliberately NO abfss:// URL output: the address is the ACCOUNT's dfs endpoint plus the filesystem name, and only the account knows its real endpoint (partitioned-DNS accounts use a different hostname than the classic shared DNS). Compose URLs from the account's dfs endpoint + `filesystem_name`.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Medallion zones** -- one filesystem per zone (raw, curated, gold), each with an owner-manages / group-reads / others-shut-out ACL and DEFAULT twins so new files inherit the posture. Start from the **Data Lake Zone Filesystem** preset.

**Team workspace** -- a named Entra security group owns the workspace root; team membership IS the access list, managed in Entra instead of per-filesystem. Start from the **Team-Scoped Workspace Filesystem** preset.

**Regulated zone under a customer key** -- a default encryption scope backed by Key Vault on just the compliance zone. Start from the **Regulated Zone with Customer-Managed Key** preset.

## Works With

- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- the HNS parent account and the source of the dfs endpoint zones compose URLs from
- [**Azure Storage Encryption Scope**](/infra-catalog/azure-storage-encryption-scope) -- per-zone key isolation through `defaultEncryptionScope`
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- workload identities owning zones and appearing in ACL entries by `principal_id`
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- zone-scoped data-plane grants targeting `filesystem_id`, refined per path by the ACL
