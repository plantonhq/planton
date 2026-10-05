# Azure Storage Account

Deploys an Azure Storage Account -- the multi-service storage primitive that fronts Blob (objects), Files (SMB/NFS shares), Queues, Tables, and Data Lake Storage Gen2 behind one globally-unique DNS name. Kind, performance tier, and replication together pick the SKU; the service-level blocks (blob, file, static website) tune the data services the account exposes. The account is the CONTAINER -- its data-plane children (blob containers, file shares, queues, tables, Data Lake filesystems, local SFTP users, object-replication policies) are their own first-class Infra Components referencing this account's outputs. Blob lifecycle management is the deliberate exception: Azure models it as one per-account policy document, so it is folded in as `lifecycleRules`.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions one Storage Account carrying every configured surface:

- **The account** -- General-Purpose v2 (the default), Premium Block Blobs, Premium File Shares, or a legacy kind; Standard or Premium tier; LRS through RA-GZRS replication; the default blob access tier
- **Data Lake & protocols** -- the hierarchical namespace (real directories, POSIX ACLs, the dfs endpoint), and the SFTP / NFSv3 endpoints built on it
- **Security posture** -- HTTPS enforcement, minimum TLS, shared-key vs Microsoft Entra authorization, the account-wide anonymous-access permission, copy-scope restriction, and the SAS expiration policy
- **Network access** -- the public-endpoint switch, the data-plane firewall (default action plus IP rules, service-endpoint subnet references, trusted-service bypass, and per-resource-instance grants), and the routing preference
- **Managed identity** -- system- and/or user-assigned; the user-assigned identity is how customer-managed-key encryption unwraps its Key Vault key
- **Encryption** -- always on; optionally a customer-managed Key Vault key, infrastructure (double) encryption, and the queue/table account-key scopes
- **Blob service settings** -- versioning, change feed, soft delete for blobs and containers, point-in-time restore, last-access tracking, and CORS
- **File service settings** -- share soft delete, the SMB protocol dials, identity-based authentication (Entra Domain Services / Entra Kerberos / on-premises AD), and CORS
- **Static website hosting and a custom domain** -- the $web container at the account's web endpoint, and a CNAME onto the blob endpoint
- **Account-level immutability (WORM)** -- the default write-once policy every new container inherits
- **Lifecycle management rules** -- tiering (Cool/Cold/Archive) and deletion schedules for blobs, snapshots, and versions
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) merged with your `tags`

Blob containers are NOT created here -- each AzureStorageContainer references `status.outputs.storage_account_id`. The same composition edge serves shares, queues, tables, Data Lake filesystems, encryption scopes, local users, and object replication.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the account will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **The account name** -- 3-24 LOWERCASE letters and digits only (no hyphens), globally unique across all of Azure: it becomes the DNS prefix of every service endpoint ({name}.blob.core.windows.net and friends).
- **For customer-managed keys** -- a Key Vault with purge protection, an AzureKeyVaultKey, and an AzureUserAssignedIdentity holding wrap/unwrap on the vault (the "Key Vault Crypto Service Encryption User" role), all BEFORE the account deploys.
- **For firewall subnet rules** -- each admitted AzureSubnet needs the Microsoft.Storage service endpoint enabled.

## Deploy

### Console

Open the deployment store, find **Azure Storage Account**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the account's surfaces in the order a practitioner decides them -- placement and name, kind/tier/replication, protocols, security posture, network access, identity, encryption, the blob and file services, website, WORM, lifecycle, and tags. Every default dropdown carries an honestly-labeled "Azure default" option, so an account that says nothing behaves exactly like a bare CLI manifest. Start from the **Production Locked-Down Account** preset in the [Presets](#presets) tab for the deny-by-default posture.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureStorageAccount
metadata:
  name: app-storage
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "app-rg"
  # 3-24 lowercase letters/digits, globally unique across Azure.
  accountName: acmeappstorage
  # Zone-redundant: the single-region production recommendation.
  replicationType: ZRS
  # Anonymous container access becomes unrepresentable account-wide.
  allowNestedItemsToBePublic: false
  blobProperties:
    versioningEnabled: true
    deleteRetentionPolicy:
      days: 30
    containerDeleteRetentionPolicy:
      days: 30
```

```shell
planton apply -f azure-storage-account.yaml
```

This creates a zone-redundant General-Purpose v2 account with versioning and 30-day soft delete; everything unspecified keeps Azure's own default (Standard tier, Hot access tier, TLS 1.2, HTTPS required). An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the account to its resource group -- and wire each data-plane child to the account:

```yaml
# On the AzureStorageAccount:
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: app-rg
      fieldPath: status.outputs.resource_group_name
  networkRules:
    defaultAction: DENY
    bypass:
      - AZURE_SERVICES
    virtualNetworkSubnetIds:
      - valueFrom:
          kind: AzureSubnet
          name: app-subnet
          fieldPath: status.outputs.subnet_id

# On each AzureStorageContainer that lives in the account:
spec:
  storageAccountId:
    valueFrom:
      kind: AzureStorageAccount
      name: app-storage
      fieldPath: status.outputs.storage_account_id
```

The InfraPipeline resolves the dependency graph, deploys the group and subnet first, then the account, then the containers and shares that live in it.

## Key Configuration

These are the most important decisions when configuring a Storage Account. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Kind × tier × replication** -- together they pick the SKU. General-Purpose v2 + Standard serves virtually every workload; Premium pairs with the specialized kinds (Premium Block Blobs for low-latency objects and premium Data Lake, Premium File Shares for enterprise SMB/NFS -- which carries NO blob service). Replication is the durability dial: LRS survives a disk, ZRS a zone, GZRS a region; the RA variants light up read-only secondary endpoints in the paired region. Kind and tier are creation-fixed; replication changes cross the zonal/non-zonal boundary only by replacement.

**Hierarchical namespace** -- Data Lake Storage Gen2: real directories, POSIX ACLs, the dfs endpoint, and the foundation for SFTP and NFSv3. The trade is blob versioning (mutually exclusive) -- and with it point-in-time restore and the WORM policy. Decide per account: analytics accounts take HNS, data-protection accounts take versioning. Fixed at creation.

**Keys vs Entra** -- the account's two access keys are static root credentials (they surface in the outputs, masked). `sharedAccessKeyEnabled: false` is the zero-static-credential posture: every data-plane request authorizes through Microsoft Entra role assignments -- verify every consumer supports Entra auth first. The SAS policy caps user-signed token lifetimes account-wide (LOG to observe, BLOCK to enforce).

**Network access** -- Azure's real default is reachable-from-everywhere; authorization is the only guard. Locking down is two steps: `networkRules.defaultAction: DENY`, then admit exactly the networks applications call from (public IPs, service-endpoint subnets, trusted Microsoft services, specific resource instances). `publicNetworkAccessEnabled: false` removes the public endpoint entirely -- private endpoints only. These rules govern the DATA plane; ARM management operations are never subject to them.

**Encryption ownership** -- always encrypted; `customerManagedKey` brings a Key Vault key you own (rotation, revocation, audit), unwrapped by a user-assigned identity attached via `identity`. Queue and table data join the CMK only when their `*EncryptionKeyType` is ACCOUNT -- a creation-fixed choice. `infrastructureEncryptionEnabled` double-encrypts at rest for regimes that demand it.

**The data-protection ladder** -- versioning keeps overwrites recoverable, soft delete keeps deletions recoverable, the change feed records every change, and point-in-time restore (requiring all three, with a window shorter than the soft-delete window) rolls the whole blob service back to any instant. The account-level `immutabilityPolicy` (WORM) adds regulatory tamper-proofing -- LOCKED is irreversible by design.

**Lifecycle rules** -- tier data down (Cool after 30 days without a read, Archive at 180, delete at 730) and expire old versions so protection does not become a storage bill. Each destination accepts exactly one aging basis; last-access schedules need `blobProperties.lastAccessTimeEnabled`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureUserAssignedIdentity** | `identity.identityIds[]`, `customerManagedKey.userAssignedIdentityId` | `status.outputs.identity_id` |
| **AzureKeyVaultKey** | `customerManagedKey.keyVaultKeyId` | `status.outputs.versionless_id` |
| **AzureSubnet** (per firewall rule) | `networkRules.virtualNetworkSubnetIds[]` | `status.outputs.subnet_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef. The account ARM ID is the composition seam the whole storage family references:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `storage_account_id` | Azure Resource Manager ID of the account | AzureStorageContainer and every other storage satellite; role-assignment scopes; private endpoints |
| `storage_account_name` | The account's name (the DNS prefix) | App-hosting kinds (Function App, Linux Web App) bind it |
| `resource_group_name` | The resource group the account lives in | Co-locating satellites |
| `primary_blob_endpoint` / `primary_blob_host` | The blob endpoint URL and bare hostname | Application config; CDN / Front Door origins; custom-domain CNAMEs |
| `primary_queue_endpoint` / `primary_table_endpoint` / `primary_file_endpoint` / `primary_dfs_endpoint` | Per-service endpoints | Queue/table clients, SMB mounts, analytics engines |
| `primary_web_endpoint` / `primary_web_host` | The static-website endpoint and hostname | The CDN / Front Door origin for static sites |
| `secondary_*_endpoint` (six) | Paired-region read-only endpoints | Read failover -- live only on RA_GRS / RA_GZRS |
| `primary_access_key` / `secondary_access_key` | The account's shared keys (SECRETS -- treat as root passwords) | Only consumers that genuinely require key auth |
| `primary_connection_string` / `secondary_connection_string` / `primary_blob_connection_string` / `secondary_blob_connection_string` | Ready-to-use connection strings (SECRETS) | Function App storage bindings and similar |
| `identity_principal_id` | The system-assigned identity's principal | AzureRoleAssignment grants for the account acting on other resources |
| `blob_service_id` / `file_service_id` / `queue_service_id` / `table_service_id` | Per-service ARM IDs | AzureMonitorDiagnosticSetting targets for DATA-ACCESS logs (the account-level ID exposes account metrics only) |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**General-purpose app storage** -- a General-Purpose v2 account with sensible production defaults; containers, queues, and tables compose as satellites. Start from the **General-Purpose v2 Account** preset.

**Production locked-down** -- GZRS replication, a DENY firewall admitting declared subnets plus trusted Microsoft services, anonymous access unrepresentable, SAS lifetimes policed, full blob data protection, and a lifecycle schedule that tiers stale data down. Start from the **Production Locked-Down Account** preset.

**Data Lake Gen2** -- the hierarchical namespace for analytics engines, with the versioning trade-off accepted. Start from the **Data Lake Storage Gen2 Account** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the account is created in
- [**Azure Storage Container**](/infra-catalog/azure-storage-container) -- blob containers referencing the account ID (the primary composition edge)
- [**Azure Storage Share**](/infra-catalog/azure-storage-share) -- SMB/NFS file shares on the account's file service
- [**Azure Storage Queue**](/infra-catalog/azure-storage-queue) -- lightweight message queues on the account's queue service
- [**Azure Storage Table**](/infra-catalog/azure-storage-table) -- NoSQL key-attribute tables on the account's table service
- [**Azure Storage Data Lake Gen2 Filesystem**](/infra-catalog/azure-storage-data-lake-gen2-filesystem) -- Data Lake filesystems on a hierarchical-namespace account
- [**Azure Storage Local User**](/infra-catalog/azure-storage-local-user) -- SFTP credentials when the SFTP endpoint is on
- [**Azure Storage Encryption Scope**](/infra-catalog/azure-storage-encryption-scope) -- per-scope encryption keys within the account
- [**Azure Storage Object Replication**](/infra-catalog/azure-storage-object-replication) -- cross-account async blob replication policies
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- unwraps the customer-managed key; composes vault grants before the account exists
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) -- the customer-managed encryption key
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- service-endpoint subnets admitted through the firewall
- [**Azure Private Endpoint**](/infra-catalog/azure-private-endpoint) -- private connectivity when the public endpoint is off
- [**Azure Monitor Diagnostic Setting**](/infra-catalog/azure-monitor-diagnostic-setting) -- data-access telemetry against the per-service IDs
- [**Azure Function App**](/infra-catalog/azure-function-app) and [**Azure Linux Web App**](/infra-catalog/azure-linux-web-app) -- bind the account for their runtime storage
