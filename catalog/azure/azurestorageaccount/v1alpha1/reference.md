# AzureStorageAccount

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureStorageAccountSpec** defines the configuration for creating an Azure
Storage Account: the multi-service storage primitive that fronts Blob
(objects), Files (SMB/NFS shares), Queues, Tables, and Data Lake Storage
Gen2 behind one globally-unique DNS name. Kind, performance tier, and
replication together pick the SKU; the service-level blocks (blob, file,
static website) tune the data services the account exposes.

**The account is the container; data-plane children are their own kinds.**
Blob containers are first-class `AzureStorageContainer` resources
referencing this account's `storage_account_id` output -- the account spec
deliberately creates no containers. Blob lifecycle management, by contrast,
is folded in as `lifecycle_rules`: Azure models it as a single per-account
policy document with no independent lifecycle and nothing referencing it.

**Network access** follows Azure's real default: the account is reachable
from all networks until `network_rules` declares otherwise. Locking an
account down is a two-step posture -- `default_action: DENY` plus explicit
IP rules, subnet references, and/or trusted-service bypass. Note that ARM
(control-plane) operations are never subject to these rules; they govern
the data plane only.

**Encryption** is always on. The dials are ownership and depth: bring your
own key with `customer_managed_key` (a Key Vault key unwrapped by a
user-assigned identity), double-encrypt at rest with
`infrastructure_encryption_enabled`, and move queue/table encryption under
the account key scope with the `*_encryption_key_type` fields.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureStorageAccount
metadata:
  name: test-storage-account
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  accountName: plantonhackstorage
  # Exercises the SKU-trio enum mappings (kind/tier/replication) plus the
  # COLD access tier added in v4.
  accountKind: STORAGE_V2
  accountTier: STANDARD
  replicationType: RA_GRS
  accessTier: COOL
  # Security posture: TLS floor, key policy, anonymous-access lockdown,
  # SAS lifetime policy with the BLOCK action (exercises both enum maps).
  minTlsVersion: TLS1_2
  sharedAccessKeyEnabled: true
  defaultToOauthAuthentication: true
  allowNestedItemsToBePublic: false
  allowedCopyScope: AAD
  sasPolicy:
    expirationPeriod: "90.00:00:00"
    expirationAction: BLOCK
  # Exercises the account-scoped queue/table encryption key types.
  queueEncryptionKeyType: ACCOUNT
  tableEncryptionKeyType: ACCOUNT
  infrastructureEncryptionEnabled: true
  # Exercises the identity + CMK pair (the identity must carry the
  # unwrapping user-assigned identity).
  identity:
    type: USER_ASSIGNED
    identityIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/storage-uai
  customerManagedKey:
    keyVaultKeyId:
      value: https://test-vault.vault.azure.net/keys/storage-cmk
    userAssignedIdentityId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/storage-uai
  # Exercises the firewall: DENY default, two bypass classes, IP rules,
  # a subnet reference, and a private-link exception.
  networkRules:
    defaultAction: DENY
    bypass:
      - AZURE_SERVICES
      - METRICS
    ipRules:
      - "203.0.113.0/24"
    virtualNetworkSubnetIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/app
    privateLinkAccess:
      - endpointResourceId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Synapse/workspaces/test-ws
  # Exercises the full blob-service surface: versioning + change feed +
  # both soft-delete policies + point-in-time restore + CORS (incl. the
  # PATCH method) + last-access tracking.
  blobProperties:
    versioningEnabled: true
    changeFeedEnabled: true
    changeFeedRetentionInDays: 90
    lastAccessTimeEnabled: true
    deleteRetentionPolicy:
      days: 30
      permanentDeleteEnabled: true
    containerDeleteRetentionPolicy:
      days: 14
    restorePolicy:
      days: 7
    corsRules:
      - allowedOrigins:
          - https://app.example.com
        allowedMethods:
          - GET
          - PUT
          - PATCH
        allowedHeaders:
          - "*"
        exposedHeaders:
          - x-ms-meta-*
        maxAgeInSeconds: 3600
  # Exercises the file-service block on a standard StorageV2 account
  # (share retention + SMB dials; multichannel stays off -- premium-only).
  shareProperties:
    retentionPolicy:
      days: 14
    smb:
      versions:
        - SMB3.1.1
      authenticationTypes:
        - Kerberos
      kerberosTicketEncryptionType:
        - AES-256
      channelEncryptionType:
        - AES-256-GCM
  # Exercises the standalone static-website resource path.
  staticWebsite:
    indexDocument: index.html
    error404Document: "404.html"
  # Exercises the routing enum + endpoint publication flags.
  routing:
    choice: INTERNET_ROUTING
    publishInternetEndpoints: true
  # Exercises the Entra Kerberos files-auth path with a default share
  # permission (the SHARE_PERMISSION_* enum map).
  azureFilesAuthentication:
    directoryType: AADKERB
    defaultShareLevelPermission: SHARE_PERMISSION_READER
  # Exercises the management-policy fold end to end: a tiering rule with
  # prefix filters, an index-tag rule (base-blob only -- ARM forbids
  # snapshot/version actions with tag filters), and snapshot/version
  # trimming.
  lifecycleRules:
    - name: age-out-logs
      filters:
        blobTypes:
          - BLOCK_BLOB
        prefixMatch:
          - logs/
      actions:
        baseBlob:
          tierToCoolAfterDaysSinceModificationGreaterThan: 30
          tierToArchiveAfterDaysSinceModificationGreaterThan: 90
          deleteAfterDaysSinceModificationGreaterThan: 365
        snapshot:
          deleteAfterDaysSinceCreationGreaterThan: 90
        version:
          changeTierToCoolAfterDaysSinceCreation: 30
          deleteAfterDaysSinceCreation: 180
    - name: expire-tagged
      filters:
        blobTypes:
          - BLOCK_BLOB
        matchBlobIndexTags:
          - name: retention
            value: short
      actions:
        baseBlob:
          deleteAfterDaysSinceCreationGreaterThan: 30
  tags:
    team: platform
    cost-center: eng
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.accountName` | `string` | yes |  |  |
| `spec.accountKind` | `enum` |  |  |  |
| `spec.accountTier` | `enum` |  |  |  |
| `spec.replicationType` | `enum` |  |  |  |
| `spec.accessTier` | `enum` |  |  |  |
| `spec.provisionedBillingModelVersion` | `string` |  |  |  |
| `spec.edgeZone` | `string` |  |  |  |
| `spec.httpsTrafficOnlyEnabled` | `bool` |  | `true` |  |
| `spec.minTlsVersion` | `enum` |  |  |  |
| `spec.sharedAccessKeyEnabled` | `bool` |  | `true` |  |
| `spec.defaultToOauthAuthentication` | `bool` |  |  |  |
| `spec.allowNestedItemsToBePublic` | `bool` |  | `true` |  |
| `spec.publicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.allowedCopyScope` | `enum` |  |  |  |
| `spec.sasPolicy` | `AzureStorageAccountSasPolicy` |  |  |  |
| `spec.sasPolicy.expirationPeriod` | `string` | yes |  |  |
| `spec.sasPolicy.expirationAction` | `enum` |  |  |  |
| `spec.localUserEnabled` | `bool` |  | `true` |  |
| `spec.sftpEnabled` | `bool` |  |  |  |
| `spec.crossTenantReplicationEnabled` | `bool` |  |  |  |
| `spec.isHnsEnabled` | `bool` |  |  |  |
| `spec.nfsv3Enabled` | `bool` |  |  |  |
| `spec.largeFileShareEnabled` | `bool` |  |  |  |
| `spec.dnsEndpointType` | `enum` |  |  |  |
| `spec.infrastructureEncryptionEnabled` | `bool` |  |  |  |
| `spec.queueEncryptionKeyType` | `enum` |  |  |  |
| `spec.tableEncryptionKeyType` | `enum` |  |  |  |
| `spec.identity` | `AzureStorageAccountIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.customerManagedKey` | `AzureStorageAccountCustomerManagedKey` |  |  |  |
| `spec.customerManagedKey.keyVaultKeyId` | `string \| valueFrom` | yes |  | AzureKeyVaultKey (`status.outputs.versionless_id`) |
| `spec.customerManagedKey.userAssignedIdentityId` | `string \| valueFrom` | yes |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.networkRules` | `AzureStorageAccountNetworkRules` |  |  |  |
| `spec.networkRules.defaultAction` | `enum` | yes |  |  |
| `spec.networkRules.bypass` | `[]enum` |  |  |  |
| `spec.networkRules.ipRules` | `[]string` |  |  |  |
| `spec.networkRules.virtualNetworkSubnetIds` | `[]string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.networkRules.privateLinkAccess` | `[]AzureStorageAccountPrivateLinkAccess` |  |  |  |
| `spec.networkRules.privateLinkAccess[].endpointResourceId` | `string` | yes |  |  |
| `spec.networkRules.privateLinkAccess[].endpointTenantId` | `string` |  |  |  |
| `spec.blobProperties` | `AzureStorageAccountBlobProperties` |  |  |  |
| `spec.blobProperties.versioningEnabled` | `bool` |  |  |  |
| `spec.blobProperties.changeFeedEnabled` | `bool` |  |  |  |
| `spec.blobProperties.changeFeedRetentionInDays` | `int32` |  |  |  |
| `spec.blobProperties.defaultServiceVersion` | `string` |  |  |  |
| `spec.blobProperties.lastAccessTimeEnabled` | `bool` |  |  |  |
| `spec.blobProperties.deleteRetentionPolicy` | `AzureStorageAccountDeleteRetentionPolicy` |  |  |  |
| `spec.blobProperties.deleteRetentionPolicy.days` | `int32` |  | `7` |  |
| `spec.blobProperties.deleteRetentionPolicy.permanentDeleteEnabled` | `bool` |  |  |  |
| `spec.blobProperties.containerDeleteRetentionPolicy` | `AzureStorageAccountContainerDeleteRetentionPolicy` |  |  |  |
| `spec.blobProperties.containerDeleteRetentionPolicy.days` | `int32` |  | `7` |  |
| `spec.blobProperties.restorePolicy` | `AzureStorageAccountRestorePolicy` |  |  |  |
| `spec.blobProperties.restorePolicy.days` | `int32` |  |  |  |
| `spec.blobProperties.corsRules` | `[]AzureStorageAccountCorsRule` |  |  |  |
| `spec.blobProperties.corsRules[].allowedOrigins` | `[]string` | yes |  |  |
| `spec.blobProperties.corsRules[].allowedMethods` | `[]string` | yes |  |  |
| `spec.blobProperties.corsRules[].allowedHeaders` | `[]string` | yes |  |  |
| `spec.blobProperties.corsRules[].exposedHeaders` | `[]string` | yes |  |  |
| `spec.blobProperties.corsRules[].maxAgeInSeconds` | `int32` |  |  |  |
| `spec.shareProperties` | `AzureStorageAccountShareProperties` |  |  |  |
| `spec.shareProperties.retentionPolicy` | `AzureStorageAccountShareRetentionPolicy` |  |  |  |
| `spec.shareProperties.retentionPolicy.days` | `int32` |  | `7` |  |
| `spec.shareProperties.smb` | `AzureStorageAccountSmbSettings` |  |  |  |
| `spec.shareProperties.smb.versions` | `[]string` |  |  |  |
| `spec.shareProperties.smb.authenticationTypes` | `[]string` |  |  |  |
| `spec.shareProperties.smb.kerberosTicketEncryptionType` | `[]string` |  |  |  |
| `spec.shareProperties.smb.channelEncryptionType` | `[]string` |  |  |  |
| `spec.shareProperties.smb.multichannelEnabled` | `bool` |  |  |  |
| `spec.shareProperties.corsRules` | `[]AzureStorageAccountCorsRule` |  |  |  |
| `spec.shareProperties.corsRules[].allowedOrigins` | `[]string` | yes |  |  |
| `spec.shareProperties.corsRules[].allowedMethods` | `[]string` | yes |  |  |
| `spec.shareProperties.corsRules[].allowedHeaders` | `[]string` | yes |  |  |
| `spec.shareProperties.corsRules[].exposedHeaders` | `[]string` | yes |  |  |
| `spec.shareProperties.corsRules[].maxAgeInSeconds` | `int32` |  |  |  |
| `spec.staticWebsite` | `AzureStorageAccountStaticWebsite` |  |  |  |
| `spec.staticWebsite.indexDocument` | `string` |  |  |  |
| `spec.staticWebsite.error404Document` | `string` |  |  |  |
| `spec.routing` | `AzureStorageAccountRouting` |  |  |  |
| `spec.routing.choice` | `enum` |  |  |  |
| `spec.routing.publishInternetEndpoints` | `bool` |  |  |  |
| `spec.routing.publishMicrosoftEndpoints` | `bool` |  |  |  |
| `spec.customDomain` | `AzureStorageAccountCustomDomain` |  |  |  |
| `spec.customDomain.name` | `string` | yes |  |  |
| `spec.customDomain.useSubdomain` | `bool` |  |  |  |
| `spec.azureFilesAuthentication` | `AzureStorageAccountAzureFilesAuthentication` |  |  |  |
| `spec.azureFilesAuthentication.directoryType` | `enum` | yes |  |  |
| `spec.azureFilesAuthentication.activeDirectory` | `AzureStorageAccountActiveDirectory` |  |  |  |
| `spec.azureFilesAuthentication.activeDirectory.domainName` | `string` | yes |  |  |
| `spec.azureFilesAuthentication.activeDirectory.domainGuid` | `string` | yes |  |  |
| `spec.azureFilesAuthentication.activeDirectory.domainSid` | `string` |  |  |  |
| `spec.azureFilesAuthentication.activeDirectory.storageSid` | `string` |  |  |  |
| `spec.azureFilesAuthentication.activeDirectory.forestName` | `string` |  |  |  |
| `spec.azureFilesAuthentication.activeDirectory.netbiosDomainName` | `string` |  |  |  |
| `spec.azureFilesAuthentication.defaultShareLevelPermission` | `enum` |  |  |  |
| `spec.immutabilityPolicy` | `AzureStorageAccountImmutabilityPolicy` |  |  |  |
| `spec.immutabilityPolicy.state` | `enum` | yes |  |  |
| `spec.immutabilityPolicy.periodSinceCreationInDays` | `int32` |  |  |  |
| `spec.immutabilityPolicy.allowProtectedAppendWrites` | `bool` |  |  |  |
| `spec.lifecycleRules` | `[]AzureStorageAccountLifecycleRule` |  |  |  |
| `spec.lifecycleRules[].name` | `string` | yes |  |  |
| `spec.lifecycleRules[].enabled` | `bool` |  | `true` |  |
| `spec.lifecycleRules[].filters` | `AzureStorageAccountLifecycleFilters` | yes |  |  |
| `spec.lifecycleRules[].filters.blobTypes` | `[]enum` | yes |  |  |
| `spec.lifecycleRules[].filters.prefixMatch` | `[]string` |  |  |  |
| `spec.lifecycleRules[].filters.matchBlobIndexTags` | `[]AzureStorageAccountLifecycleTagFilter` |  |  |  |
| `spec.lifecycleRules[].filters.matchBlobIndexTags[].name` | `string` | yes |  |  |
| `spec.lifecycleRules[].filters.matchBlobIndexTags[].operation` | `string` |  | `==` |  |
| `spec.lifecycleRules[].filters.matchBlobIndexTags[].value` | `string` | yes |  |  |
| `spec.lifecycleRules[].actions` | `AzureStorageAccountLifecycleActions` | yes |  |  |
| `spec.lifecycleRules[].actions.baseBlob` | `AzureStorageAccountLifecycleBaseBlobActions` |  |  |  |
| `spec.lifecycleRules[].actions.baseBlob.tierToCoolAfterDaysSinceModificationGreaterThan` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.baseBlob.tierToCoolAfterDaysSinceLastAccessTimeGreaterThan` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.baseBlob.tierToCoolAfterDaysSinceCreationGreaterThan` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.baseBlob.autoTierToHotFromCoolEnabled` | `bool` |  |  |  |
| `spec.lifecycleRules[].actions.baseBlob.tierToColdAfterDaysSinceModificationGreaterThan` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.baseBlob.tierToColdAfterDaysSinceLastAccessTimeGreaterThan` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.baseBlob.tierToColdAfterDaysSinceCreationGreaterThan` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.baseBlob.tierToArchiveAfterDaysSinceModificationGreaterThan` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.baseBlob.tierToArchiveAfterDaysSinceLastAccessTimeGreaterThan` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.baseBlob.tierToArchiveAfterDaysSinceCreationGreaterThan` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.baseBlob.tierToArchiveAfterDaysSinceLastTierChangeGreaterThan` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.baseBlob.deleteAfterDaysSinceModificationGreaterThan` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.baseBlob.deleteAfterDaysSinceLastAccessTimeGreaterThan` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.baseBlob.deleteAfterDaysSinceCreationGreaterThan` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.snapshot` | `AzureStorageAccountLifecycleSnapshotActions` |  |  |  |
| `spec.lifecycleRules[].actions.snapshot.changeTierToCoolAfterDaysSinceCreation` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.snapshot.tierToColdAfterDaysSinceCreationGreaterThan` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.snapshot.changeTierToArchiveAfterDaysSinceCreation` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.snapshot.tierToArchiveAfterDaysSinceLastTierChangeGreaterThan` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.snapshot.deleteAfterDaysSinceCreationGreaterThan` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.version` | `AzureStorageAccountLifecycleVersionActions` |  |  |  |
| `spec.lifecycleRules[].actions.version.changeTierToCoolAfterDaysSinceCreation` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.version.tierToColdAfterDaysSinceCreationGreaterThan` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.version.changeTierToArchiveAfterDaysSinceCreation` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.version.tierToArchiveAfterDaysSinceLastTierChangeGreaterThan` | `int32` |  |  |  |
| `spec.lifecycleRules[].actions.version.deleteAfterDaysSinceCreation` | `int32` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region where the account will be created (e.g. "eastus",
"westeurope"). Changing the region replaces the account.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the account will be created in. Can be a
literal resource-group name or a reference to an AzureResourceGroup's
name output. Changing it replaces the account.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.accountName

`string` · required

The account's name: 3-24 lowercase letters and digits ONLY (no hyphens
-- stricter than most Azure names) and GLOBALLY unique across all of
Azure, because it becomes the DNS prefix of every service endpoint
({name}.blob.core.windows.net and friends). Changing the name replaces
the account.

- rule: {"required":true,"string":{"pattern":"^[a-z0-9]{3,24}$"}}

### spec.accountKind

`enum`

The kind of account, which decides which data services exist and which
SKUs are legal. Unspecified means STORAGE_V2 -- the general-purpose v2
account that serves virtually every workload. Only the legacy
STORAGE -> STORAGE_V2 upgrade is an in-place change; any other kind
change replaces the account.

Allowed values (use exactly as shown):

- `azure_storage_account_kind_unspecified` -- Not specified: STORAGE_V2 (the general-purpose v2 account).
- `STORAGE_V2` -- General-purpose v2: blobs, files, queues, tables, and ADLS Gen2 -- the right choice for virtually every workload.
- `BLOB_STORAGE` -- Legacy blob-only account with access tiers. Superseded by STORAGE_V2; use only when a legacy pricing artifact demands it.
- `BLOCK_BLOB_STORAGE` -- Premium block blobs (SSD): low-latency object storage and premium ADLS Gen2. Pairs with PREMIUM tier.
- `FILE_STORAGE` -- Premium file shares (SSD): enterprise SMB/NFS file serving. Pairs with PREMIUM tier.
- `STORAGE` -- Legacy general-purpose v1: predates access tiers and modern features; upgradeable in place to STORAGE_V2.

### spec.accountTier

`enum`

The performance tier. Unspecified means STANDARD (HDD-backed, all
redundancy options). PREMIUM (SSD-backed, low single-digit-ms latency)
pairs with the specialized kinds: BLOCK_BLOB_STORAGE for premium
blobs/ADLS, FILE_STORAGE for premium file shares. Changing the tier
replaces the account.

Allowed values (use exactly as shown):

- `azure_storage_account_tier_unspecified` -- Not specified: STANDARD.
- `STANDARD` -- HDD-backed. All redundancy options; the cost-efficient default.
- `PREMIUM` -- SSD-backed, single-digit-ms latency. Pairs with the specialized kinds (BLOCK_BLOB_STORAGE, FILE_STORAGE) or premium page blobs.

### spec.replicationType

`enum`

How the account's data is replicated. Unspecified means LRS (three
copies in one datacenter -- the dev/test floor). Production guidance:
ZRS to survive a zone loss, GZRS to survive a regional loss with zone
resilience at home, RA_* variants to add a read-only secondary
endpoint in the paired region. Switching between the zonal family
(ZRS/GZRS/RA_GZRS) and the non-zonal family (LRS/GRS/RA_GRS) replaces
the account; changes within a family are in-place.

Allowed values (use exactly as shown):

- `azure_storage_account_replication_type_unspecified` -- Not specified: LRS.
- `LRS` -- Locally-redundant: three copies in one datacenter. Cheapest; no zone or region resilience.
- `ZRS` -- Zone-redundant: three copies across availability zones. The single-region production recommendation.
- `GRS` -- Geo-redundant: LRS at home plus three async copies in the paired region.
- `GZRS` -- Geo-zone-redundant: ZRS at home plus geo-replication. The highest durability tier.
- `RA_GRS` -- GRS plus a READ-ONLY secondary endpoint in the paired region (the secondary endpoints in the outputs become live).
- `RA_GZRS` -- GZRS plus a read-only secondary endpoint.

### spec.accessTier

`enum`

The default access tier for blob data: the cost/latency trade-off
applied to blobs that don't set their own tier. Unspecified lets
Azure apply HOT. Only meaningful for STORAGE_V2, BLOB_STORAGE, and
FILE_STORAGE kinds. COOL (30-day minimum retention) and COLD (90-day)
trade storage price for access price; ACCESS_TIER_PREMIUM is the
read-back of premium accounts.

Allowed values (use exactly as shown):

- `azure_storage_account_access_tier_unspecified` -- Not specified: Azure applies HOT on the kinds that support tiers.
- `HOT` -- Frequent access: highest storage price, lowest access price.
- `COOL` -- Infrequent access (30-day minimum retention): cheaper storage, pricier access.
- `COLD` -- Rare access (90-day minimum retention): between COOL and archive.
- `ACCESS_TIER_PREMIUM` -- The tier premium accounts report; storage on SSD media.

### spec.provisionedBillingModelVersion

`string`

The provisioned billing model version. Set "V2" to select the
provisioned-v2 billing model (currently meaningful for FileStorage
accounts -- capacity/IOPS/throughput are provisioned independently);
leave unset for pay-as-you-go / provisioned v1. Fixed at creation.

- rule: provisioned_billing_model_version must be "V2" or left unset (pay-as-you-go / provisioned v1)

### spec.edgeZone

`string`

The Azure Edge Zone where the account should live, for
ultra-low-latency edge scenarios. Leave unset for a regular regional
deployment (virtually all accounts). Fixed at creation.

### spec.httpsTrafficOnlyEnabled

`bool` · optional (explicit presence)

Whether the account rejects plaintext HTTP and requires HTTPS on
every request. Azure's default is true; disable only for the rare
legacy client that cannot speak TLS (NFSv3 mounts are exempt from
this setting by design).

- default: `true`

### spec.minTlsVersion

`enum`

The minimum TLS version the account accepts. Unspecified applies
TLS1_2 (Azure's default since 2024, the compliance floor, and the
only floor still provisionable -- the legacy 1.0/1.1 floors are
retired; see the enum).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_storage_account_min_tls_version_unspecified` -- Not specified: TLS1_2 (Azure's default and the compliance floor).
- `TLS1_2` -- TLS 1.2 -- the floor for new deployments and the only floor Azure still accepts.

### spec.sharedAccessKeyEnabled

`bool` · optional (explicit presence)

Whether the account's shared access keys (and SAS tokens signed by
them) are accepted for authorization. Azure's default is true.
Setting false forces every data-plane request through Microsoft
Entra -- the zero-static-credential posture; be sure every consumer
supports Entra auth before flipping it.

- default: `true`

### spec.defaultToOauthAuthentication

`bool`

Whether the Azure portal and tools default to Microsoft Entra
authorization (instead of key-based) when a user browses the
account's data. Azure's default is false. A UX nudge toward Entra --
it does not disable keys (that is shared_access_key_enabled).

### spec.allowNestedItemsToBePublic

`bool` · optional (explicit presence)

Whether individual containers may opt into public (anonymous) read
access. This only PERMITS per-container public access; each
container is still private unless its own access type says
otherwise. Azure and the provider default to false; this spec
deliberately defaults to true so a container-level access type
works without a second account-level switch. Set false to make
anonymous access unrepresentable account-wide, the recommended
posture for anything that isn't a public website/CDN origin.

- default: `true`

### spec.publicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the account's public endpoints accept traffic at all.
Azure's default is true. Setting false removes the public endpoint
entirely -- reachable only via private endpoints -- and makes
network_rules moot.

- default: `true`

### spec.allowedCopyScope

`enum`

Restricts where data can be COPIED to from this account (the
exfiltration guard on the copy APIs). Unspecified leaves copy
unrestricted (Azure's default). AAD limits copy destinations to
accounts in the same tenant; PRIVATE_LINK to accounts wired by
private link.

Allowed values (use exactly as shown):

- `azure_storage_account_allowed_copy_scope_unspecified` -- Not specified: copy destinations are unrestricted (Azure's default).
- `AAD` -- Copy only to accounts in the same Microsoft Entra tenant.
- `PRIVATE_LINK` -- Copy only to accounts connected via private link.

### spec.sasPolicy

`AzureStorageAccountSasPolicy`

The account-wide SAS expiration policy: caps how long user-signed
SAS tokens may live, and whether a violation is logged or blocked.
Omit to leave SAS lifetimes unpoliced (Azure's default).

### spec.sasPolicy.expirationPeriod

`string` · required

The longest lifetime a user-signed SAS token may have, in
"D.HH:MM:SS" form (e.g. "90.00:00:00" for 90 days).

- rule: {"required":true,"string":{"pattern":"^\\d+\\.\\d{2}:\\d{2}:\\d{2}$"}}

### spec.sasPolicy.expirationAction

`enum`

What happens when a SAS token exceeds the period: LOG (record the
violation -- Azure's default) or BLOCK (reject the token).

Allowed values (use exactly as shown):

- `azure_storage_account_sas_expiration_action_unspecified` -- Not specified: LOG (Azure's default).
- `LOG` -- Record a policy violation but accept the token.
- `BLOCK` -- Reject tokens that exceed the expiration period.

### spec.localUserEnabled

`bool` · optional (explicit presence)

Whether local (SFTP) user identities may be created on the account.
Azure's default is true, but local users only matter when
sftp_enabled is true -- they are the credential model SFTP uses.

- default: `true`

### spec.sftpEnabled

`bool`

Whether the account exposes an SFTP endpoint onto its blob storage.
Azure's default is false. Requires is_hns_enabled (SFTP is a Data
Lake Gen2 feature) and bills per enabled hour.

### spec.crossTenantReplicationEnabled

`bool`

Whether object replication may cross Microsoft Entra tenants.
Azure's (and the provider's) default is false; enable only when
an object-replication policy genuinely spans tenants.

### spec.isHnsEnabled

`bool`

Whether the account has a hierarchical namespace (Data Lake Storage
Gen2): real directories, POSIX ACLs, and the dfs endpoint --
required for analytics engines (Spark/Databricks/Synapse), SFTP,
and NFSv3. Mutually exclusive with blob versioning. Fixed at
creation.

### spec.nfsv3Enabled

`bool`

Whether the blob service accepts NFSv3 mounts. Requires
is_hns_enabled, a supported tier/kind pairing (STANDARD + STORAGE_V2
or PREMIUM + BLOCK_BLOB_STORAGE), and LRS or RA_GRS replication.
Fixed at creation.

### spec.largeFileShareEnabled

`bool`

Whether file shares may grow beyond 5 TiB (to 100 TiB). Only
meaningful for STORAGE_V2 and FILE_STORAGE kinds; premium
FileStorage accounts have it on inherently. One-way: Azure cannot
disable it once enabled -- false here means "leave it to Azure",
never "disable" (both engines send the flag only when true).

### spec.dnsEndpointType

`enum`

The DNS endpoint architecture. Unspecified means the classic shared
DNS ({name}.blob.core.windows.net). AZURE_DNS_ZONE gives the account
partitioned DNS for very-high-scale subscriptions (thousands of
accounts); it cannot be combined with blob_properties.restore_policy.
Fixed at creation.

Allowed values (use exactly as shown):

- `azure_storage_account_dns_endpoint_type_unspecified` -- Not specified: the classic shared DNS endpoints.
- `DNS_ENDPOINT_STANDARD` -- Classic shared DNS: {name}.blob.core.windows.net.
- `AZURE_DNS_ZONE` -- Partitioned DNS for very-high-scale subscriptions; incompatible with blob point-in-time restore.

### spec.infrastructureEncryptionEnabled

`bool`

Whether data is DOUBLE-encrypted at rest (two independent encryption
layers with separate keys/algorithms) for regimes that require it.
Azure's default is false. Only STORAGE_V2 accounts, or PREMIUM
BLOCK_BLOB_STORAGE / FILE_STORAGE accounts, support it. Fixed at
creation.

### spec.queueEncryptionKeyType

`enum`

Which key scope encrypts the QUEUE service. Unspecified means
SERVICE (a Microsoft-managed per-service key). ACCOUNT moves queues
under the account's encryption key -- required for queue data to be
covered by customer_managed_key. Not supported on the legacy STORAGE
kind. Fixed at creation.

Allowed values (use exactly as shown):

- `azure_storage_account_encryption_key_type_unspecified` -- Not specified: SERVICE (Azure's default).
- `SERVICE` -- A Microsoft-managed per-service key -- customer-managed keys do NOT cover this service's data.
- `ACCOUNT` -- The account's encryption key scope -- required for the service's data to be covered by customer_managed_key.

### spec.tableEncryptionKeyType

`enum`

Which key scope encrypts the TABLE service. Same semantics and
constraints as queue_encryption_key_type, for tables. Fixed at
creation.

Allowed values (use exactly as shown):

- `azure_storage_account_encryption_key_type_unspecified` -- Not specified: SERVICE (Azure's default).
- `SERVICE` -- A Microsoft-managed per-service key -- customer-managed keys do NOT cover this service's data.
- `ACCOUNT` -- The account's encryption key scope -- required for the service's data to be covered by customer_managed_key.

### spec.identity

`AzureStorageAccountIdentity`

The account's managed identity. Required (with a user-assigned
entry) for customer_managed_key; a system-assigned identity's
principal surfaces in the outputs for AzureRoleAssignment grants.

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the account (its principal surfaces in the outputs for
AzureRoleAssignment grants); USER_ASSIGNED brings identities you
manage and share across resources (required for customer-managed-key
encryption); SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_storage_account_identity_type_unspecified` -- Not specified -- invalid; choose an explicit flavor when the identity block is present.
- `SYSTEM_ASSIGNED` -- Azure creates and rotates an identity bound to the account's lifecycle.
- `USER_ASSIGNED` -- Bring your own AzureUserAssignedIdentity resources (required for customer-managed-key encryption).
- `SYSTEM_AND_USER_ASSIGNED` -- Both flavors together.

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED / SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the account, by ARM ID. Reference
AzureUserAssignedIdentity resources so Key Vault grants can be
composed before the account exists.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.customerManagedKey

`AzureStorageAccountCustomerManagedKey`

Customer-managed-key encryption: the account's data is encrypted
with a Key Vault key you own instead of a Microsoft-managed key.
Requires a user-assigned identity (attached via identity) that has
wrap/unwrap access on the key's vault, and the vault must have
purge protection enabled.

### spec.customerManagedKey.keyVaultKeyId

`string | valueFrom` · required

The Key Vault key that encrypts the account's data, by data-plane
key ID. Defaults to referencing an AzureKeyVaultKey's versionless_id
output so key rotations propagate automatically; pin a versioned ID
only when a compliance regime demands an immutable key version. The
key's vault must have purge protection enabled.

- references: AzureKeyVaultKey (`status.outputs.versionless_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVaultKey, name: <that resource's name>, fieldPath: status.outputs.versionless_id}} -- a bare string does not parse

### spec.customerManagedKey.userAssignedIdentityId

`string | valueFrom` · required

The user-assigned identity Azure uses to unwrap the key, by ARM ID.
Must be one of the identities attached via the account's identity
block, with wrap/unwrap access on the key's vault (a "Key Vault
Crypto Service Encryption User" role assignment, or the equivalent
access policy).

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.networkRules

`AzureStorageAccountNetworkRules`

Data-plane network access control: the default action plus IP,
virtual-network, trusted-service, and private-link exceptions. Omit
to leave the account reachable from all networks (Azure's default).

### spec.networkRules.defaultAction

`enum` · required

What happens to traffic no explicit rule admits. Azure's own
account default is Allow-from-everywhere; declaring this block
requires choosing explicitly. DENY plus the exception lists below
is the production posture.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_storage_account_network_default_action_unspecified` -- Not specified -- invalid; the network_rules block requires an explicit choice.
- `ALLOW` -- Admit traffic no rule matches (the firewall is advisory).
- `DENY` -- Reject traffic no rule matches (the production posture).

### spec.networkRules.bypass

`[]enum`

Traffic classes exempt from the rules. AZURE_SERVICES admits the
trusted Microsoft services (Backup, Monitor, Event Grid, ...);
LOGGING and METRICS admit the classic analytics readers; NONE (the
sole entry) exempts nothing. Unset lets Azure default to
AZURE_SERVICES.

- rule: {"repeated":{"items":{"enum":{"definedOnly":true}}}}

Allowed values (use exactly as shown):

- `azure_storage_account_network_bypass_unspecified` -- Not specified -- not a valid list entry.
- `AZURE_SERVICES` -- Trusted Microsoft services (Backup, Monitor, Event Grid, ...).
- `LOGGING` -- Classic storage-analytics log readers.
- `METRICS` -- Classic storage-analytics metric readers.
- `NONE` -- Exempt nothing (use as the sole entry).

### spec.networkRules.ipRules

`[]string`

Public IPv4 addresses or CIDR ranges admitted to the data plane,
e.g. "203.0.113.0/24". Private-range (RFC 1918) addresses are not
accepted -- VNet traffic is admitted via
virtual_network_subnet_ids instead.

- rule: {"repeated":{"maxItems":"400"}}

### spec.networkRules.virtualNetworkSubnetIds

`[]string | valueFrom`

Subnets admitted to the data plane, by ARM ID. Each subnet must
have the Microsoft.Storage service endpoint enabled. References
AzureSubnet outputs so the network graph composes in one manifest
set.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.networkRules.privateLinkAccess

`[]AzureStorageAccountPrivateLinkAccess`

Resource instances granted private-link-style access through the
firewall (e.g. a Synapse workspace or Azure Backup vault reaching a
locked-down account without a full private endpoint).

### spec.networkRules.privateLinkAccess[].endpointResourceId

`string` · required

The ARM ID of the resource being admitted (e.g. a Synapse
workspace, Azure Backup recovery vault, or Machine Learning
workspace).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.networkRules.privateLinkAccess[].endpointTenantId

`string` · optional (explicit presence)

The tenant of the admitted resource. Leave unset for the
deploying credential's tenant -- the correct value for virtually
every deployment.

- rule: {"string":{"uuid":true}}

### spec.blobProperties

`AzureStorageAccountBlobProperties`

Blob service settings: versioning, soft delete for blobs and
containers, point-in-time restore, change feed, CORS, and last-access
tracking. Not supported on FILE_STORAGE accounts (they have no blob
service).

- rule: restore_policy requires versioning_enabled = true
- rule: restore_policy requires change_feed_enabled = true
- rule: restore_policy requires delete_retention_policy (and its window must exceed the restore window)

### spec.blobProperties.versioningEnabled

`bool`

Whether every blob write keeps the previous version recoverable.
Azure's default is false. The foundation for restore_policy and
immutability_policy; incompatible with hierarchical namespace.

### spec.blobProperties.changeFeedEnabled

`bool`

Whether the change feed (an ordered, replayable log of every blob
change) is recorded. Azure's default is false. Required for
restore_policy.

### spec.blobProperties.changeFeedRetentionInDays

`int32` · optional (explicit presence)

How many days change-feed records are retained, 1-146000. Unset
means infinite retention. Only meaningful with change_feed_enabled.

- rule: {"int32":{"lte":146000,"gte":1}}

### spec.blobProperties.defaultServiceVersion

`string`

The default REST API version the blob service answers with for
requests that don't pin one, e.g. "2020-06-12". Unset lets Azure
choose.

### spec.blobProperties.lastAccessTimeEnabled

`bool`

Whether each blob's last-access time is tracked (a prerequisite for
the lifecycle rules' days-since-last-access conditions). Azure's
default is false.

### spec.blobProperties.deleteRetentionPolicy

`AzureStorageAccountDeleteRetentionPolicy`

Soft delete for BLOBS: deleted blobs are retained and recoverable
for the configured window. Omit to leave blob soft delete off.

### spec.blobProperties.deleteRetentionPolicy.days

`int32` · optional (explicit presence)

How many days deleted blobs remain recoverable, 1-365. Unspecified
applies Azure's default of 7.

- default: `7`
- rule: {"int32":{"lte":365,"gte":1}}

### spec.blobProperties.deleteRetentionPolicy.permanentDeleteEnabled

`bool`

Whether soft-deleted blobs can be PERMANENTLY deleted before the
window ends (an explicit erasure API for right-to-be-forgotten
regimes). Azure's default is false.

### spec.blobProperties.containerDeleteRetentionPolicy

`AzureStorageAccountContainerDeleteRetentionPolicy`

Soft delete for CONTAINERS: deleted containers are retained and
recoverable for the configured window. Omit to leave container
soft delete off.

### spec.blobProperties.containerDeleteRetentionPolicy.days

`int32` · optional (explicit presence)

How many days deleted containers remain recoverable, 1-365.
Unspecified applies Azure's default of 7.

- default: `7`
- rule: {"int32":{"lte":365,"gte":1}}

### spec.blobProperties.restorePolicy

`AzureStorageAccountRestorePolicy`

Point-in-time restore: the whole blob service can be rolled back to
any instant inside the window. Requires versioning, change feed,
and blob soft delete -- and the restore window must be shorter than
the soft-delete window.

### spec.blobProperties.restorePolicy.days

`int32`

How many days back the blob service can be restored to, 1-365.
Must be less than delete_retention_policy.days.

- rule: {"int32":{"lte":365,"gte":1}}

### spec.blobProperties.corsRules

`[]AzureStorageAccountCorsRule`

CORS rules for browser-based access to the blob service, evaluated
in order (max 5).

- rule: {"repeated":{"maxItems":"5"}}

### spec.blobProperties.corsRules[].allowedOrigins

`[]string` · required

The origins allowed to make cross-origin requests, e.g.
"https://app.example.com", or "*" for any origin.

- rule: {"repeated":{"minItems":"1","maxItems":"64"}}

### spec.blobProperties.corsRules[].allowedMethods

`[]string` · required

The HTTP methods the rule admits.

- rule: {"repeated":{"minItems":"1","items":{"cel":[{"id":"storage_account_cors_method_valid","message":"allowed_methods entries must be one of: DELETE, GET, HEAD, MERGE, POST, OPTIONS, PUT, PATCH","expression":"this in ['DELETE', 'GET', 'HEAD', 'MERGE', 'POST', 'OPTIONS', 'PUT', 'PATCH']"}]}}}

### spec.blobProperties.corsRules[].allowedHeaders

`[]string` · required

The request headers the browser may send, e.g. "x-ms-meta-*", or
"*" for all.

- rule: {"repeated":{"minItems":"1","maxItems":"64"}}

### spec.blobProperties.corsRules[].exposedHeaders

`[]string` · required

The response headers exposed to the browser, e.g. "x-ms-meta-*",
or "*" for all.

- rule: {"repeated":{"minItems":"1","maxItems":"64"}}

### spec.blobProperties.corsRules[].maxAgeInSeconds

`int32`

How long (seconds) the browser may cache the preflight response.

- rule: {"int32":{"lte":2000000000,"gte":0}}

### spec.shareProperties

`AzureStorageAccountShareProperties`

File service settings: share soft-delete retention, SMB protocol
dials, and CORS. Supported on FILE_STORAGE accounts and standard
STORAGE_V2 / legacy STORAGE accounts (the kinds that have a file
service).

### spec.shareProperties.retentionPolicy

`AzureStorageAccountShareRetentionPolicy`

Soft delete for file shares: deleted shares are retained and
recoverable for the configured window. Omit to accept Azure's
service-side default (7 days).

### spec.shareProperties.retentionPolicy.days

`int32` · optional (explicit presence)

How many days deleted shares remain recoverable, 1-365.
Unspecified applies Azure's default of 7.

- default: `7`
- rule: {"int32":{"lte":365,"gte":1}}

### spec.shareProperties.smb

`AzureStorageAccountSmbSettings`

SMB protocol dials for file-share mounts: allowed protocol
versions, authentication and encryption suites, and multichannel.

### spec.shareProperties.smb.versions

`[]string`

The SMB protocol versions the service accepts. Unset admits all.

- rule: {"repeated":{"items":{"cel":[{"id":"storage_account_smb_version_valid","message":"versions entries must be one of: SMB2.1, SMB3.0, SMB3.1.1","expression":"this in ['SMB2.1', 'SMB3.0', 'SMB3.1.1']"}]}}}

### spec.shareProperties.smb.authenticationTypes

`[]string`

The authentication methods the service accepts. Unset admits all.

- rule: {"repeated":{"items":{"cel":[{"id":"storage_account_smb_auth_valid","message":"authentication_types entries must be one of: Kerberos, NTLMv2","expression":"this in ['Kerberos', 'NTLMv2']"}]}}}

### spec.shareProperties.smb.kerberosTicketEncryptionType

`[]string`

The Kerberos ticket encryption suites the service accepts. Unset
admits all.

- rule: {"repeated":{"items":{"cel":[{"id":"storage_account_smb_kerb_valid","message":"kerberos_ticket_encryption_type entries must be one of: AES-256, RC4-HMAC","expression":"this in ['AES-256', 'RC4-HMAC']"}]}}}

### spec.shareProperties.smb.channelEncryptionType

`[]string`

The SMB channel encryption suites the service accepts. Unset
admits all.

- rule: {"repeated":{"items":{"cel":[{"id":"storage_account_smb_channel_valid","message":"channel_encryption_type entries must be one of: AES-128-CCM, AES-128-GCM, AES-256-GCM","expression":"this in ['AES-128-CCM', 'AES-128-GCM', 'AES-256-GCM']"}]}}}

### spec.shareProperties.smb.multichannelEnabled

`bool`

Whether SMB Multichannel (multiple parallel network connections
per session) is enabled -- a premium-tier FileStorage feature.
Azure's default is false.

### spec.shareProperties.corsRules

`[]AzureStorageAccountCorsRule`

CORS rules for browser-based access to the file service, evaluated
in order (max 5).

- rule: {"repeated":{"maxItems":"5"}}

### spec.shareProperties.corsRules[].allowedOrigins

`[]string` · required

The origins allowed to make cross-origin requests, e.g.
"https://app.example.com", or "*" for any origin.

- rule: {"repeated":{"minItems":"1","maxItems":"64"}}

### spec.shareProperties.corsRules[].allowedMethods

`[]string` · required

The HTTP methods the rule admits.

- rule: {"repeated":{"minItems":"1","items":{"cel":[{"id":"storage_account_cors_method_valid","message":"allowed_methods entries must be one of: DELETE, GET, HEAD, MERGE, POST, OPTIONS, PUT, PATCH","expression":"this in ['DELETE', 'GET', 'HEAD', 'MERGE', 'POST', 'OPTIONS', 'PUT', 'PATCH']"}]}}}

### spec.shareProperties.corsRules[].allowedHeaders

`[]string` · required

The request headers the browser may send, e.g. "x-ms-meta-*", or
"*" for all.

- rule: {"repeated":{"minItems":"1","maxItems":"64"}}

### spec.shareProperties.corsRules[].exposedHeaders

`[]string` · required

The response headers exposed to the browser, e.g. "x-ms-meta-*",
or "*" for all.

- rule: {"repeated":{"minItems":"1","maxItems":"64"}}

### spec.shareProperties.corsRules[].maxAgeInSeconds

`int32`

How long (seconds) the browser may cache the preflight response.

- rule: {"int32":{"lte":2000000000,"gte":0}}

### spec.staticWebsite

`AzureStorageAccountStaticWebsite`

Static website hosting on the blob service: serves the $web
container at the account's web endpoint. Only STORAGE_V2 and
BLOCK_BLOB_STORAGE accounts support it. The web endpoint surfaces in
the outputs for CDN/Front Door origins.

- rule: static_website requires index_document and/or error_404_document

### spec.staticWebsite.indexDocument

`string`

The document served at the site root and directory paths, e.g.
"index.html".

### spec.staticWebsite.error404Document

`string`

The document served on 404s, e.g. "404.html".

### spec.routing

`AzureStorageAccountRouting`

Network routing preference: whether traffic enters Microsoft's
backbone close to the client (MICROSOFT_ROUTING, the default) or
close to the account (INTERNET_ROUTING), and whether the
routing-specific endpoint sets are published.

### spec.routing.choice

`enum`

Where client traffic enters Microsoft's network. Unspecified means
MICROSOFT_ROUTING (enter near the client -- lowest latency,
Azure's default). INTERNET_ROUTING enters near the account (lower
cost, higher latency).

Allowed values (use exactly as shown):

- `azure_storage_account_routing_choice_unspecified` -- Not specified: MICROSOFT_ROUTING (Azure's default).
- `MICROSOFT_ROUTING` -- Traffic enters Microsoft's backbone close to the client: lowest latency.
- `INTERNET_ROUTING` -- Traffic rides the public internet and enters close to the account: lower cost.

### spec.routing.publishInternetEndpoints

`bool`

Whether the internet-routing endpoint set (the "-internetrouting"
hostnames) is published. Azure's default is false.

### spec.routing.publishMicrosoftEndpoints

`bool`

Whether the Microsoft-routing endpoint set is published. Azure's
default is false.

### spec.customDomain

`AzureStorageAccountCustomDomain`

A custom domain (CNAME) for the blob endpoint, e.g.
"assets.example.com". Azure validates ownership via the CNAME (or
the asverify subdomain when use_subdomain is set).

### spec.customDomain.name

`string` · required

The custom domain name, e.g. "assets.example.com". The domain's
CNAME must point at the account's blob host.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.customDomain.useSubdomain

`bool`

Whether ownership is validated indirectly via the
"asverify.{domain}" CNAME instead of the domain itself -- avoids
downtime when migrating a domain already serving traffic.

### spec.azureFilesAuthentication

`AzureStorageAccountAzureFilesAuthentication`

Identity-based authentication for Azure Files (SMB): Entra Domain
Services, Entra Kerberos, or on-premises Active Directory. Governs
how file-share mounts authenticate users.

- rule: directory_type AD requires the active_directory block (domain coordinates)

### spec.azureFilesAuthentication.directoryType

`enum` · required

The directory service that authenticates SMB mounts: AADDS (Entra
Domain Services), AADKERB (Entra Kerberos -- for hybrid identities
without domain-controller line-of-sight), or AD (on-premises
Active Directory, requires active_directory).

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_storage_account_directory_service_type_unspecified` -- Not specified -- invalid; the block requires an explicit service.
- `AADDS` -- Microsoft Entra Domain Services.
- `AADKERB` -- Microsoft Entra Kerberos (hybrid identities, no domain-controller line of sight needed).
- `AD` -- On-premises Active Directory Domain Services (requires active_directory coordinates).

### spec.azureFilesAuthentication.activeDirectory

`AzureStorageAccountActiveDirectory`

For AD (and optionally AADKERB): the on-premises domain's
coordinates.

### spec.azureFilesAuthentication.activeDirectory.domainName

`string` · required

The AD DS domain's DNS name, e.g. "corp.example.com".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.azureFilesAuthentication.activeDirectory.domainGuid

`string` · required

The AD DS domain's GUID.

- rule: {"required":true,"string":{"uuid":true}}

### spec.azureFilesAuthentication.activeDirectory.domainSid

`string`

The security identifier (SID) of the AD DS domain. Required for
directory_type AD; not used for AADKERB.

### spec.azureFilesAuthentication.activeDirectory.storageSid

`string`

The SID of the computer account created for the storage account in
AD DS. Required for directory_type AD; not used for AADKERB.

### spec.azureFilesAuthentication.activeDirectory.forestName

`string`

The AD DS forest the domain belongs to, e.g. "corp.example.com".
Required for directory_type AD; not used for AADKERB.

### spec.azureFilesAuthentication.activeDirectory.netbiosDomainName

`string`

The NetBIOS name of the AD DS domain, e.g. "CORP". Required for
directory_type AD; not used for AADKERB.

### spec.azureFilesAuthentication.defaultShareLevelPermission

`enum`

The default share-level permission applied to authenticated users
that carry no explicit share-level role assignment. Unspecified
means SHARE_PERMISSION_NONE (explicit assignments only).

Allowed values (use exactly as shown):

- `azure_storage_account_default_share_permission_unspecified` -- Not specified: SHARE_PERMISSION_NONE.
- `SHARE_PERMISSION_NONE` -- No default access -- every user needs an explicit share-level role assignment.
- `SHARE_PERMISSION_READER` -- Read access to shares by default.
- `SHARE_PERMISSION_CONTRIBUTOR` -- Read/write/delete access to shares by default.
- `SHARE_PERMISSION_ELEVATED_CONTRIBUTOR` -- Contributor plus NTFS-permission modification.

### spec.immutabilityPolicy

`AzureStorageAccountImmutabilityPolicy`

Account-level immutability (WORM) policy applied as the default to
every new container. Requires blob versioning. The state machine is
one-way: once LOCKED, the policy (and the account's data) cannot be
un-locked -- start with UNLOCKED and lock only after validating the
configuration. Fixed at creation (the block's presence).

### spec.immutabilityPolicy.state

`enum` · required

The policy's state. Start with UNLOCKED (retention adjustable,
policy deletable) or DISABLED; move to LOCKED only after validating
-- LOCKED is irreversible and makes the retention window a
compliance guarantee Azure itself cannot override.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_storage_account_immutability_state_unspecified` -- Not specified -- invalid; the policy block requires an explicit state.
- `DISABLED` -- The policy exists but is not enforced.
- `UNLOCKED` -- Enforced but adjustable/deletable -- the safe starting state.
- `LOCKED` -- Enforced and IRREVERSIBLE: retention can never be shortened and the policy can never be removed. Enter only from UNLOCKED, after validation.

### spec.immutabilityPolicy.periodSinceCreationInDays

`int32`

How many days each object is immutable from its creation, 1-146000.

- rule: {"int32":{"lte":146000,"gte":1}}

### spec.immutabilityPolicy.allowProtectedAppendWrites

`bool`

Whether new blocks may still be APPENDED to append blobs under the
policy (audit-log pattern: append-only, never modify or delete).

### spec.lifecycleRules

`[]AzureStorageAccountLifecycleRule`

Blob lifecycle management rules: tier blobs down (cool/cold/archive)
and delete them on age/access schedules, filtered by container
prefix, blob type, and index tags. Azure models this as one
per-account policy document -- it lives and dies with the account,
so it is folded here rather than modeled as a standalone kind.

- rule: a rule filtering by blob index tags cannot carry snapshot or version actions (an ARM restriction) -- only base_blob actions

### spec.lifecycleRules[].name

`string` · required

The rule's name, unique within the account's policy.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.lifecycleRules[].enabled

`bool` · optional (explicit presence)

Whether the rule is evaluated. Unspecified applies true --
declaring a rule normally means wanting it active.

- default: `true`

### spec.lifecycleRules[].filters

`AzureStorageAccountLifecycleFilters` · required

What the rule applies to: blob types, name prefixes, and/or index
tags.

- rule: {"required":true}

### spec.lifecycleRules[].filters.blobTypes

`[]enum` · required

The blob types the rule covers. Tiering actions only apply to
BLOCK_BLOB; APPEND_BLOB supports deletion only.

- rule: {"repeated":{"minItems":"1","items":{"enum":{"definedOnly":true}}}}

Allowed values (use exactly as shown):

- `azure_storage_account_lifecycle_blob_type_unspecified` -- Not specified -- not a valid list entry.
- `BLOCK_BLOB` -- Standard block blobs (objects). Supports tiering and deletion.
- `APPEND_BLOB` -- Append blobs (logs). Supports deletion only.

### spec.lifecycleRules[].filters.prefixMatch

`[]string`

Name prefixes the rule covers (e.g. "logs/", "container1/raw/").
Unset covers the whole account.

### spec.lifecycleRules[].filters.matchBlobIndexTags

`[]AzureStorageAccountLifecycleTagFilter`

Blob index tags the rule matches. A rule carrying tag filters
cannot have snapshot or version actions (an ARM restriction).

### spec.lifecycleRules[].filters.matchBlobIndexTags[].name

`string` · required

The tag name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.lifecycleRules[].filters.matchBlobIndexTags[].operation

`string` · optional (explicit presence)

The comparison operator. ARM supports equality only; unspecified
applies "==".

- default: `==`
- rule: operation must be "==" (the only comparison ARM supports)

### spec.lifecycleRules[].filters.matchBlobIndexTags[].value

`string` · required

The tag value to match.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.lifecycleRules[].actions

`AzureStorageAccountLifecycleActions` · required

What the rule does: tiering and deletion schedules for base blobs,
snapshots, and versions.

- rule: {"required":true}
- rule: a lifecycle rule needs at least one of base_blob, snapshot, or version actions

### spec.lifecycleRules[].actions.baseBlob

`AzureStorageAccountLifecycleBaseBlobActions`

Schedules for current (base) blobs.

- rule: tier-to-cool accepts exactly one aging basis: modification, last access, or creation
- rule: tier-to-cold accepts exactly one aging basis: modification, last access, or creation
- rule: tier-to-archive accepts exactly one aging basis: modification, last access, or creation
- rule: delete accepts exactly one aging basis: modification, last access, or creation
- rule: auto_tier_to_hot_from_cool_enabled requires tier_to_cool_after_days_since_last_access_time_greater_than

### spec.lifecycleRules[].actions.baseBlob.tierToCoolAfterDaysSinceModificationGreaterThan

`int32` · optional (explicit presence)

Move to COOL after this many days without modification.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.baseBlob.tierToCoolAfterDaysSinceLastAccessTimeGreaterThan

`int32` · optional (explicit presence)

Move to COOL after this many days without being read.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.baseBlob.tierToCoolAfterDaysSinceCreationGreaterThan

`int32` · optional (explicit presence)

Move to COOL after this many days since creation.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.baseBlob.autoTierToHotFromCoolEnabled

`bool`

Whether a blob tiered to COOL on the last-access schedule moves
back to HOT automatically when it is read again. Requires
tier_to_cool_after_days_since_last_access_time_greater_than.

### spec.lifecycleRules[].actions.baseBlob.tierToColdAfterDaysSinceModificationGreaterThan

`int32` · optional (explicit presence)

Move to COLD after this many days without modification.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.baseBlob.tierToColdAfterDaysSinceLastAccessTimeGreaterThan

`int32` · optional (explicit presence)

Move to COLD after this many days without being read.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.baseBlob.tierToColdAfterDaysSinceCreationGreaterThan

`int32` · optional (explicit presence)

Move to COLD after this many days since creation.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.baseBlob.tierToArchiveAfterDaysSinceModificationGreaterThan

`int32` · optional (explicit presence)

Move to ARCHIVE after this many days without modification.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.baseBlob.tierToArchiveAfterDaysSinceLastAccessTimeGreaterThan

`int32` · optional (explicit presence)

Move to ARCHIVE after this many days without being read.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.baseBlob.tierToArchiveAfterDaysSinceCreationGreaterThan

`int32` · optional (explicit presence)

Move to ARCHIVE after this many days since creation.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.baseBlob.tierToArchiveAfterDaysSinceLastTierChangeGreaterThan

`int32` · optional (explicit presence)

Guard against archive ping-pong: only re-archive a blob this many
days after its last tier change.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.baseBlob.deleteAfterDaysSinceModificationGreaterThan

`int32` · optional (explicit presence)

Delete after this many days without modification.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.baseBlob.deleteAfterDaysSinceLastAccessTimeGreaterThan

`int32` · optional (explicit presence)

Delete after this many days without being read.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.baseBlob.deleteAfterDaysSinceCreationGreaterThan

`int32` · optional (explicit presence)

Delete after this many days since creation.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.snapshot

`AzureStorageAccountLifecycleSnapshotActions`

Schedules for blob snapshots.

### spec.lifecycleRules[].actions.snapshot.changeTierToCoolAfterDaysSinceCreation

`int32` · optional (explicit presence)

Move snapshots to COOL after this many days.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.snapshot.tierToColdAfterDaysSinceCreationGreaterThan

`int32` · optional (explicit presence)

Move snapshots to COLD after this many days.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.snapshot.changeTierToArchiveAfterDaysSinceCreation

`int32` · optional (explicit presence)

Move snapshots to ARCHIVE after this many days.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.snapshot.tierToArchiveAfterDaysSinceLastTierChangeGreaterThan

`int32` · optional (explicit presence)

Guard against archive ping-pong: only re-archive a snapshot this
many days after its last tier change.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.snapshot.deleteAfterDaysSinceCreationGreaterThan

`int32` · optional (explicit presence)

Delete snapshots after this many days.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.version

`AzureStorageAccountLifecycleVersionActions`

Schedules for previous blob versions.

### spec.lifecycleRules[].actions.version.changeTierToCoolAfterDaysSinceCreation

`int32` · optional (explicit presence)

Move versions to COOL after this many days.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.version.tierToColdAfterDaysSinceCreationGreaterThan

`int32` · optional (explicit presence)

Move versions to COLD after this many days.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.version.changeTierToArchiveAfterDaysSinceCreation

`int32` · optional (explicit presence)

Move versions to ARCHIVE after this many days.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.version.tierToArchiveAfterDaysSinceLastTierChangeGreaterThan

`int32` · optional (explicit presence)

Guard against archive ping-pong: only re-archive a version this
many days after its last tier change.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.lifecycleRules[].actions.version.deleteAfterDaysSinceCreation

`int32` · optional (explicit presence)

Delete versions after this many days.

- rule: {"int32":{"lte":99999,"gte":0}}

### spec.tags

`map<string, string>`

Free-form tags applied to the account, merged over the
Planton-derived resource tags (organization, environment, resource
id); a user tag with the same key wins. Tags are Azure's governance
surface -- Azure Policy enforces them and Microsoft Cost Management
groups by them.

## Validation Rules

- `storage_account_access_tier_kind_gate`: access_tier is only supported on STORAGE_V2, BLOB_STORAGE, and FILE_STORAGE accounts
- `storage_account_hns_kind_gate`: is_hns_enabled (Data Lake Gen2) is only supported on STORAGE_V2, BLOB_STORAGE, and BLOCK_BLOB_STORAGE accounts
- `storage_account_sftp_requires_hns`: sftp_enabled requires is_hns_enabled (SFTP is a Data Lake Gen2 feature)
- `storage_account_nfsv3_requires_hns`: nfsv3_enabled requires is_hns_enabled
- `storage_account_nfsv3_tier_kind_gate`: nfsv3_enabled requires STANDARD tier with STORAGE_V2 kind, or PREMIUM tier with BLOCK_BLOB_STORAGE kind
- `storage_account_nfsv3_replication_gate`: nfsv3_enabled requires LRS or RA_GRS replication
- `storage_account_blob_storage_zrs_unsupported`: ZRS replication is not supported for BLOB_STORAGE accounts
- `storage_account_versioning_conflicts_hns`: blob versioning cannot be enabled on a hierarchical-namespace (is_hns_enabled) account
- `storage_account_immutability_requires_versioning`: immutability_policy requires blob_properties.versioning_enabled = true
- `storage_account_infrastructure_encryption_gate`: infrastructure_encryption_enabled requires a STORAGE_V2 account, or a PREMIUM-tier BLOCK_BLOB_STORAGE or FILE_STORAGE account
- `storage_account_dns_zone_restore_conflict`: blob_properties.restore_policy cannot be combined with dns_endpoint_type AZURE_DNS_ZONE
- `storage_account_smb_multichannel_premium_only`: share_properties.smb.multichannel_enabled is only supported on PREMIUM-tier accounts
- `storage_account_blob_properties_kind_gate`: blob_properties is not supported on FILE_STORAGE accounts (they have no blob service)
- `storage_account_share_properties_gate`: share_properties requires a FILE_STORAGE account, or a STANDARD-tier STORAGE_V2 / legacy STORAGE account
- `storage_account_static_website_kind_gate`: static_website is only supported on STORAGE_V2 and BLOCK_BLOB_STORAGE accounts
- `storage_account_large_file_share_kind_gate`: large_file_share_enabled is only supported on STORAGE_V2 and FILE_STORAGE accounts
- `storage_account_queue_key_type_kind_gate`: queue_encryption_key_type ACCOUNT is not supported on the legacy STORAGE kind
- `storage_account_table_key_type_kind_gate`: table_encryption_key_type ACCOUNT is not supported on the legacy STORAGE kind
- `storage_account_cmk_requires_user_assigned_identity`: customer_managed_key requires identity with type USER_ASSIGNED or SYSTEM_AND_USER_ASSIGNED (the unwrapping identity must be attached to the account)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureStorageAccount, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.storage_account_id` | `string` | The Azure Resource Manager ID of the account. This is the primary output: AzureStorageContainer's storage_account_id references it, and role assignments (Storage Blob Data Reader/Contributor) scope to it. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Storage/storageAccounts/{name} |
| `status.outputs.storage_account_name` | `string` | The account's name -- the DNS prefix of every service endpoint, and what app-hosting kinds (Function App, Linux Web App) bind to. |
| `status.outputs.resource_group_name` | `string` | The resource group the account lives in. |
| `status.outputs.primary_blob_endpoint` | `string` | The primary blob endpoint, e.g. "https://{name}.blob.core.windows.net/". Object (blob) reads and writes go here. |
| `status.outputs.primary_blob_host` | `string` | The primary blob HOSTNAME (no scheme/slash) -- the exact value a CDN or Front Door origin, or a custom-domain CNAME, points at. |
| `status.outputs.primary_queue_endpoint` | `string` | The primary queue endpoint, e.g. "https://{name}.queue.core.windows.net/". |
| `status.outputs.primary_table_endpoint` | `string` | The primary table endpoint, e.g. "https://{name}.table.core.windows.net/". |
| `status.outputs.primary_file_endpoint` | `string` | The primary file endpoint, e.g. "https://{name}.file.core.windows.net/". SMB/NFS share mounts and the Files REST API go here. |
| `status.outputs.primary_dfs_endpoint` | `string` | The primary Data Lake Storage Gen2 (dfs) endpoint, e.g. "https://{name}.dfs.core.windows.net/". Analytics engines (Spark/Databricks/Synapse) address hierarchical-namespace data here. |
| `status.outputs.primary_web_endpoint` | `string` | The primary static-website endpoint, e.g. "https://{name}.z13.web.core.windows.net/". Live only when static_website is configured. |
| `status.outputs.primary_web_host` | `string` | The static-website HOSTNAME (no scheme/slash) -- the exact value a CDN or Front Door origin points at for static-site delivery. |
| `status.outputs.secondary_blob_endpoint` | `string` | The secondary (paired-region, read-only) blob endpoint. Populated only for RA_GRS / RA_GZRS replication. |
| `status.outputs.secondary_queue_endpoint` | `string` | The secondary queue endpoint (RA_GRS / RA_GZRS only). |
| `status.outputs.secondary_table_endpoint` | `string` | The secondary table endpoint (RA_GRS / RA_GZRS only). |
| `status.outputs.secondary_file_endpoint` | `string` | The secondary file endpoint (RA_GRS / RA_GZRS only). |
| `status.outputs.secondary_dfs_endpoint` | `string` | The secondary dfs endpoint (RA_GRS / RA_GZRS only). |
| `status.outputs.secondary_web_endpoint` | `string` | The secondary static-website endpoint (RA_GRS / RA_GZRS only). |
| `status.outputs.primary_access_key` | `string` (sensitive) | The account's FIRST shared access key. Static credential material that authorizes EVERY data-plane operation on the account -- treat it like a root password: prefer Entra-based authorization (role assignments against storage_account_id) and reference this key only where a consumer genuinely requires key auth (e.g. a Function App's storage binding). |
| `status.outputs.secondary_access_key` | `string` (sensitive) | The account's SECOND shared access key -- exists so consumers can roll from one key to the other without downtime. The same handling guidance as primary_access_key applies. |
| `status.outputs.primary_connection_string` | `string` (sensitive) | A ready-to-use connection string carrying the account name and the PRIMARY access key. Same secret-handling guidance as the keys. |
| `status.outputs.secondary_connection_string` | `string` (sensitive) | The connection string carrying the SECONDARY access key. |
| `status.outputs.primary_blob_connection_string` | `string` (sensitive) | A blob-service-only connection string (primary key + blob endpoint) -- what some SDKs and app settings expect. |
| `status.outputs.secondary_blob_connection_string` | `string` (sensitive) | The blob-service-only connection string on the secondary key. |
| `status.outputs.identity_principal_id` | `string` | The principal (object) ID of the account's system-assigned identity, populated only when the identity type includes SYSTEM_ASSIGNED. Grant this principal roles to let the account act on other resources (e.g. reading a Key Vault key). |
| `status.outputs.blob_service_id` | `string` | The ARM ID of the account's BLOB service ({storage_account_id}/blobServices/default) -- the diagnostic- settings scope for blob DATA-ACCESS telemetry (StorageRead / StorageWrite / StorageDelete logs). The account-level ID only exposes account metrics; per-operation audit logs live on the service sub-resources, so an AzureMonitorDiagnosticSetting's target_resource_id references this output. Constructed identically on both engines (ARM materializes the service implicitly -- there is nothing to read back). |
| `status.outputs.file_service_id` | `string` | The ARM ID of the account's FILE service ({storage_account_id}/fileServices/default) -- the diagnostic- settings scope for Azure Files data-access telemetry. |
| `status.outputs.queue_service_id` | `string` | The ARM ID of the account's QUEUE service ({storage_account_id}/queueServices/default) -- the diagnostic- settings scope for queue data-access telemetry. |
| `status.outputs.table_service_id` | `string` | The ARM ID of the account's TABLE service ({storage_account_id}/tableServices/default) -- the diagnostic- settings scope for table data-access telemetry. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.customerManagedKey.keyVaultKeyId` | AzureKeyVaultKey | `status.outputs.versionless_id` |
| `spec.customerManagedKey.userAssignedIdentityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.networkRules.virtualNetworkSubnetIds` | AzureSubnet | `status.outputs.subnet_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureAiFoundry | `spec.storageAccountId` | `status.outputs.storage_account_id` |
| AzureBackupContainerStorageAccount | `spec.storageAccountId` | `status.outputs.storage_account_id` |
| AzureBackupProtectedFileShare | `spec.sourceStorageAccountId` | `status.outputs.storage_account_id` |
| AzureCognitiveAccount | `spec.storage[].storageAccountId` | `status.outputs.storage_account_id` |
| AzureComputeGalleryImage | `spec.versions[].storageAccountId` | `status.outputs.storage_account_id` |
| AzureContainerAppEnvironmentStorage | `spec.accountName` | `status.outputs.storage_account_name` |
| AzureContainerAppEnvironmentStorage | `spec.accessKey` | `status.outputs.primary_access_key` |
| AzureContainerInstance | `spec.containers[].volumes[].azureFile.storageAccountName` | `status.outputs.storage_account_name` |
| AzureContainerInstance | `spec.containers[].volumes[].azureFile.storageAccountKey` | `status.outputs.primary_access_key` |
| AzureContainerInstance | `spec.initContainers[].volumes[].azureFile.storageAccountName` | `status.outputs.storage_account_name` |
| AzureContainerInstance | `spec.initContainers[].volumes[].azureFile.storageAccountKey` | `status.outputs.primary_access_key` |
| AzureDataFactoryLinkedService | `spec.azureBlobStorage.serviceEndpoint` | `status.outputs.primary_blob_endpoint` |
| AzureDataFactoryLinkedService | `spec.dataLakeStorageGen2.url` | `status.outputs.primary_dfs_endpoint` |
| AzureDataFactoryTrigger | `spec.blobEvent.storageAccountId` | `status.outputs.storage_account_id` |
| AzureDataProtectionBackupInstance | `spec.blobStorage.storageAccountId` | `status.outputs.storage_account_id` |
| AzureDataProtectionBackupInstance | `spec.dataLakeStorage.storageAccountId` | `status.outputs.storage_account_id` |
| AzureDiskSnapshot | `spec.storageAccountId` | `status.outputs.storage_account_id` |
| AzureEventHub | `spec.captureDescription.destination.storageAccountId` | `status.outputs.storage_account_id` |
| AzureEventgridEventSubscription | `spec.destination.storageQueue.storageAccountId` | `status.outputs.storage_account_id` |
| AzureEventgridEventSubscription | `spec.deadLetter.storageAccountId` | `status.outputs.storage_account_id` |
| AzureFrontDoorOrigin | `spec.privateLink.privateLinkTargetId` | `status.outputs.storage_account_id` |
| AzureFunctionApp | `spec.storageAccountName` | `status.outputs.storage_account_name` |
| AzureFunctionApp | `spec.storageAccountAccessKey` | `status.outputs.primary_access_key` |
| AzureFunctionApp | `spec.storageMounts[].accessKey` | `status.outputs.primary_access_key` |
| AzureFunctionAppFlexConsumption | `spec.storageAccessKey` | `status.outputs.primary_access_key` |
| AzureLinuxWebApp | `spec.storageMounts[].accessKey` | `status.outputs.primary_access_key` |
| AzureMachineLearningWorkspace | `spec.storageAccountId` | `status.outputs.storage_account_id` |
| AzureMonitorDataCollectionRule | `spec.destinations.storageBlobs[].storageAccountId` | `status.outputs.storage_account_id` |
| AzureMonitorDataCollectionRule | `spec.destinations.storageBlobDirects[].storageAccountId` | `status.outputs.storage_account_id` |
| AzureMonitorDataCollectionRule | `spec.destinations.storageTableDirects[].storageAccountId` | `status.outputs.storage_account_id` |
| AzureMonitorDiagnosticSetting | `spec.storageAccountId` | `status.outputs.storage_account_id` |
| AzureNetworkWatcherFlowLog | `spec.storageAccountId` | `status.outputs.storage_account_id` |
| AzureStorageContainer | `spec.storageAccountId` | `status.outputs.storage_account_id` |
| AzureStorageDataLakeGen2Filesystem | `spec.storageAccountId` | `status.outputs.storage_account_id` |
| AzureStorageEncryptionScope | `spec.storageAccountId` | `status.outputs.storage_account_id` |
| AzureStorageLocalUser | `spec.storageAccountId` | `status.outputs.storage_account_id` |
| AzureStorageObjectReplication | `spec.sourceStorageAccountId` | `status.outputs.storage_account_id` |
| AzureStorageObjectReplication | `spec.destinationStorageAccountId` | `status.outputs.storage_account_id` |
| AzureStorageQueue | `spec.storageAccountId` | `status.outputs.storage_account_id` |
| AzureStorageShare | `spec.storageAccountId` | `status.outputs.storage_account_id` |
| AzureStorageTable | `spec.storageAccountId` | `status.outputs.storage_account_id` |

## See Also

- [Overview](../README.md)
