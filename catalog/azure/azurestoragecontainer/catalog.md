# Azure Storage Container

Deploys a blob container inside an Azure Storage Account -- the namespace unit of Azure blob storage. Applications organize objects into containers the way filesystems use top-level directories, one per data domain (uploads, logs, backups, artifacts), and Azure scopes anonymous access, data-plane RBAC, encryption scopes, and lifecycle prefixes at the container level. Containers are many-per-account with independent lifecycles, which is why they are a first-class kind referencing the account rather than a list folded into the account's spec.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Blob Container** -- a container on the referenced storage account (by ARM ID -- the control-plane path), with your chosen anonymous-access posture, optional default encryption scope, and data-plane metadata

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureStorageAccount** the container will live in, referenced through `storageAccountId`. The parent is fixed at creation: a container cannot move between accounts.
- **For anonymous access** (BLOB/CONTAINER): the account must allow nested public items -- when the account forbids it, Azure forces the container private regardless.
- **For a default encryption scope**: an AzureStorageEncryptionScope on the SAME account.

## Deploy

### Console

Open the deployment store, find **Azure Storage Container**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Private Container** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureStorageContainer
metadata:
  name: app-uploads
  org: acme-corp
  env: prod
spec:
  storageAccountId:
    valueFrom:
      kind: AzureStorageAccount
      name: app-storage
      fieldPath: status.outputs.storage_account_id
  containerName: uploads
```

```shell
planton apply -f container.yaml
```

This creates a private container named `uploads` on the `app-storage` account -- every read requires authorization, the right posture for everything that is not a public website or CDN origin. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the ValueFromRef above wires the container to its account: the InfraPipeline resolves the dependency graph, deploys the storage account first, then provisions the container with the resolved ARM ID.

## Key Configuration

These are the most important decisions when configuring a container. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Container name** -- `containerName` becomes the URL path segment under the account's blob endpoint. 3-63 lowercase letters, digits, and hyphens; unique within the account. Renaming replaces the container.

**Anonymous access** -- `containerAccessType` decides who can read WITHOUT credentials. Unspecified means PRIVATE (recommended). `BLOB` allows anonymous object reads by direct URL with no listing -- the public-website/CDN-origin pattern. `CONTAINER` additionally allows enumeration and is rarely appropriate. Anonymous access also requires the account's `allowNestedItemsToBePublic`.

**Encryption scope** -- `defaultEncryptionScope` applies sub-account key isolation to blobs that don't name their own scope (e.g. per-tenant keys inside one account); the scope must live on the same account. `encryptionScopeOverrideEnabled: false` makes the scope mandatory for every blob -- a hard isolation boundary. Both fixed at creation.

**Metadata** -- `metadata` stores free-form key/value pairs on the container (visible to anyone who can read container properties -- not for secrets). Keys must be valid C# identifiers; Azure lowercases them on read.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureStorageAccount** | `storageAccountId` | `status.outputs.storage_account_id` |
| **AzureStorageEncryptionScope** | `defaultEncryptionScope` | `status.outputs.encryption_scope_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `container_id` | Azure Resource Manager ID of the container | Container-scoped data-plane role assignments (Storage Blob Data Reader/Contributor) |
| `container_name` | The container's name | SDK clients, function bindings, app settings |
| `storage_account_name` | The parent account's name, parsed from the resolved account ID | The account/container pair without a second reference |

There is deliberately NO URL output: the container's data-plane URL is the ACCOUNT's blob endpoint plus the container name, and only the account knows its real endpoint (partitioned-DNS accounts use a different hostname than the classic shared DNS). Compose URLs from AzureStorageAccount's `primary_blob_endpoint` output + `container_name`.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Private application container** -- the default shape: private access, no encryption scope. Grant applications data access with role assignments scoped to `container_id`. Start from the **Private Container** preset.

**Public CDN origin** -- `containerAccessType: BLOB` on an account that allows nested public items; consumers fetch objects by direct URL, nobody can enumerate. Start from the **Public CDN-Origin Container** preset.

**Tenant-scoped encryption** -- a default encryption scope with the per-blob override blocked; revoking one tenant's scope key renders exactly that tenant's data unreadable. Start from the **Tenant-Scoped Encryption Container** preset.

## Works With

- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- the parent account and the source of the blob endpoint containers compose URLs from
- [**Azure Storage Encryption Scope**](/infra-catalog/azure-storage-encryption-scope) -- the same-account scope a container can adopt as its default encryption boundary
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- container-scoped data-plane grants targeting `container_id`
