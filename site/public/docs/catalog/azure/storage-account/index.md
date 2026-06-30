---
title: "Storage Account"
description: "Storage Account deployment documentation"
icon: "package"
order: 100
componentName: "azurestorageaccount"
---

# Azure Storage Account

Deploys an Azure Storage Account with configurable account kind, performance tier, replication strategy, access tier, network access controls, blob service properties, and optional blob containers. The component enforces HTTPS-only traffic and TLS 1.2 by default, applies default-deny network rules, and enables soft delete retention for both blobs and containers.

## What Gets Created

When you deploy an AzureStorageAccount resource, Planton provisions:

- **Storage Account** — an `storage.Account` resource in the specified region and resource group, configured with the chosen account kind, performance tier, replication type, access tier, HTTPS enforcement, minimum TLS version, and public nested-item access disabled
- **Network Rules** — default-deny network ACLs with optional IP allowlists, VNet subnet rules, and Azure trusted services bypass
- **Blob Properties** — blob service configuration including versioning, blob soft delete retention, and container soft delete retention
- **Blob Containers** — a `storage.Container` resource for each entry in the `containers` list, with configurable public access levels, parented to the storage account
- **Azure Tags** — resource metadata tags applied to the storage account for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or Planton provider config
- **An Azure Resource Group** where the storage account will be created (can reference an AzureResourceGroup resource)
- **A globally unique name** — Azure Storage Account names must be 3-24 lowercase alphanumeric characters and must be globally unique; the component derives the name from `metadata.name` by stripping dots, underscores, and hyphens, lowercasing, and truncating to 24 characters
- **Network planning** — know which IP ranges and/or VNet subnets need storage access if restricting with network ACLs

## Quick Start

Create a file `storage-account.yaml`:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureStorageAccount
metadata:
  name: my-storage
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AzureStorageAccount.my-storage
spec:
  region: eastus
  resourceGroup: my-rg
```

Deploy:

```shell
planton apply -f storage-account.yaml
```

This creates a StorageV2 storage account with Standard tier, LRS replication, Hot access tier, HTTPS-only traffic, TLS 1.2 minimum, default-deny network ACLs that bypass Azure trusted services, and 7-day soft delete retention for blobs and containers.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the Storage Account (e.g., `eastus`, `westeurope`). | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `accountKind` | `enum` | `STORAGE_V2` | Kind of storage account. Values: `STORAGE_V2` (general-purpose v2, recommended), `BLOB_STORAGE` (specialized blob with access tiers), `BLOCK_BLOB_STORAGE` (premium block blobs, SSD-backed), `FILE_STORAGE` (premium file shares, SSD-backed), `STORAGE` (legacy general-purpose v1). |
| `accountTier` | `enum` | `STANDARD` | Performance tier. Values: `STANDARD` (HDD-backed), `PREMIUM` (SSD-backed, only for specific account kinds). |
| `replicationType` | `enum` | `LRS` | Replication strategy. Values: `LRS` (3 copies in one datacenter), `ZRS` (3 copies across availability zones), `GRS` (6 copies: 3 local + 3 in paired region), `GZRS` (ZRS + geo-replication), `RA_GRS` (read-access geo-redundant), `RA_GZRS` (read-access geo-zone-redundant). |
| `accessTier` | `enum` | `HOT` | Default blob access tier. Only applicable for BlobStorage and StorageV2 account kinds. Values: `HOT` (frequently accessed data), `COOL` (infrequently accessed, 30-day minimum retention). |
| `enableHttpsTrafficOnly` | `bool` | `true` | When true, all requests must use HTTPS. Strongly recommended for security. |
| `minTlsVersion` | `enum` | `TLS1_2` | Minimum TLS version for incoming requests. Values: `TLS1_0`, `TLS1_1`, `TLS1_2` (recommended). |
| `networkRules.defaultAction` | `enum` | `DENY` | Default action when no explicit rule matches. Values: `DENY` (recommended), `ALLOW`. |
| `networkRules.bypassAzureServices` | `bool` | `true` | Allows traffic from trusted Azure services (Backup, Monitor, Event Grid, etc.) even when default action is `DENY`. |
| `networkRules.ipRules` | `string[]` | `[]` | IP addresses or CIDR ranges allowed to access the storage. Maximum 200 entries. |
| `networkRules.virtualNetworkSubnetIds` | `string[]` | `[]` | Azure VNet subnet resource IDs allowed to access the storage. Maximum 100 entries. |
| `blobProperties.enableVersioning` | `bool` | `false` | Enables blob versioning. When enabled, Azure maintains previous versions of blobs for data protection and recovery. |
| `blobProperties.softDeleteRetentionDays` | `int` | `7` | Retention period in days for deleted blobs. Range: 0-365. Set to 0 to disable blob soft delete. |
| `blobProperties.containerSoftDeleteRetentionDays` | `int` | `7` | Retention period in days for deleted containers. Range: 0-365. Set to 0 to disable container soft delete. |
| `containers` | `object[]` | `[]` | List of blob containers to create. Maximum 100 entries. Each entry requires `name` (3-63 lowercase characters) and optionally `accessType`. |
| `containers[].name` | `string` | — | Container name. Must be lowercase, 3-63 characters. | 
| `containers[].accessType` | `enum` | `PRIVATE` | Public access level for the container. Values: `PRIVATE` (no public read access, recommended), `BLOB` (public read for blobs only), `CONTAINER` (public read for container and blobs). |

## Examples

### Development Storage with Open Network Access

A storage account for development with network ACLs set to allow all traffic and reduced soft delete retention:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureStorageAccount
metadata:
  name: dev-storage
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AzureStorageAccount.dev-storage
spec:
  region: eastus
  resourceGroup: dev-rg
  accountKind: STORAGE_V2
  accountTier: STANDARD
  replicationType: LRS
  blobProperties:
    softDeleteRetentionDays: 1
    containerSoftDeleteRetentionDays: 1
  networkRules:
    defaultAction: ALLOW
```

### Production Storage with Blob Containers and Network Restrictions

A production storage account with geo-redundant replication, blob versioning, restricted network access from office IPs, and pre-created containers for application data:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureStorageAccount
metadata:
  name: prod-storage
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureStorageAccount.prod-storage
spec:
  region: eastus
  resourceGroup: prod-rg
  accountKind: STORAGE_V2
  accountTier: STANDARD
  replicationType: GRS
  accessTier: HOT
  enableHttpsTrafficOnly: true
  minTlsVersion: TLS1_2
  blobProperties:
    enableVersioning: true
    softDeleteRetentionDays: 30
    containerSoftDeleteRetentionDays: 30
  networkRules:
    defaultAction: DENY
    bypassAzureServices: true
    ipRules:
      - "203.0.113.0/24"
      - "198.51.100.42"
  containers:
    - name: application-data
    - name: backups
    - name: logs
      accessType: PRIVATE
```

### Premium Block Blob Storage for High-Performance Workloads

A Premium-tier block blob storage account with zone-redundant replication for latency-sensitive workloads:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureStorageAccount
metadata:
  name: perf-blobs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureStorageAccount.perf-blobs
spec:
  region: westeurope
  resourceGroup: perf-rg
  accountKind: BLOCK_BLOB_STORAGE
  accountTier: PREMIUM
  replicationType: ZRS
  blobProperties:
    enableVersioning: true
    softDeleteRetentionDays: 14
    containerSoftDeleteRetentionDays: 14
  networkRules:
    defaultAction: DENY
    bypassAzureServices: true
    virtualNetworkSubnetIds:
      - /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/perf-rg/providers/Microsoft.Network/virtualNetworks/perf-vnet/subnets/app
  containers:
    - name: hot-data
    - name: cache
```

### Cool-Tier Archival Storage

A storage account with Cool access tier for infrequently accessed data and extended soft delete retention:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureStorageAccount
metadata:
  name: archive-storage
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureStorageAccount.archive-storage
spec:
  region: eastus
  resourceGroup: archive-rg
  accountKind: STORAGE_V2
  accountTier: STANDARD
  replicationType: RA_GRS
  accessTier: COOL
  blobProperties:
    enableVersioning: true
    softDeleteRetentionDays: 365
    containerSoftDeleteRetentionDays: 365
  containers:
    - name: audit-logs
    - name: compliance-records
    - name: historical-data
```

### Using Foreign Key References

Reference an Planton-managed resource group instead of hardcoding the name:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureStorageAccount
metadata:
  name: ref-storage
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureStorageAccount.ref-storage
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  containers:
    - name: uploads
    - name: media
      accessType: BLOB
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `storageAccountId` | `string` | Azure Resource Manager ID of the Storage Account |
| `storageAccountName` | `string` | Name of the Storage Account |
| `primaryBlobEndpoint` | `string` | Primary blob endpoint URL (e.g., `https://{name}.blob.core.windows.net/`) |
| `primaryQueueEndpoint` | `string` | Primary queue endpoint URL (e.g., `https://{name}.queue.core.windows.net/`) |
| `primaryTableEndpoint` | `string` | Primary table endpoint URL (e.g., `https://{name}.table.core.windows.net/`) |
| `primaryFileEndpoint` | `string` | Primary file endpoint URL (e.g., `https://{name}.file.core.windows.net/`) |
| `primaryDfsEndpoint` | `string` | Primary DFS (Data Lake Storage Gen2) endpoint URL |
| `primaryWebEndpoint` | `string` | Primary web endpoint URL for static website hosting |
| `containerUrlMap` | `map<string, string>` | Map of container names to their blob URLs (`https://{name}.blob.core.windows.net/{container}`). Only contains containers created by this stack. |
| `region` | `string` | Azure region where the Storage Account was deployed |
| `resourceGroup` | `string` | Resource group name where the Storage Account was created |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/resource-group) — provides the resource group for storage account placement
- [AzureKeyVault](/docs/catalog/azure/key-vault) — store storage account access keys as vault secrets
- [AzureVpc](/docs/catalog/azure/vpc-virtual-network) — provides VNet subnets for network ACL rules
- [AzureAksCluster](/docs/catalog/azure/aks-cluster) — AKS workloads can mount blob containers as persistent volumes or access storage via managed identity
