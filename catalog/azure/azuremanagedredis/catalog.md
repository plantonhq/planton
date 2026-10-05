# Azure Managed Redis

Deploys an Azure Managed Redis instance -- Azure's current-generation Redis service, built on Redis Enterprise. Azure is retiring classic Azure Cache for Redis; Managed Redis is the target for new Redis deployments and the home of the capabilities the classic service never had: Redis modules (search, JSON, probabilistic filters, time series), active multi-primary geo-replication, customer-managed-key encryption, and a keyless-by-default authentication posture. The kind models the cluster and its default database in one spec: the SKU family/size ladder, high availability, engine behavior, modules, persistence, geo-group membership, managed identity, CMK, and network access.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Managed Redis Cluster** -- a Microsoft.Cache/redisEnterprise instance in the specified region and resource group, sized by the chosen SKU (family + memory in one value), with high availability (a replica and the 99.999% zone-redundant SLA) unless explicitly disabled
- **Default Database** -- the Redis process itself, mapped 1-to-1 with the cluster: authentication posture, clustering policy, eviction policy, client protocol, modules, optional persistence, and optional geo-replication-group membership
- **Customer-Managed-Key Encryption** -- created only when `customerManagedKey` is configured; wraps the data-encryption key with your Key Vault key via a user-assigned identity
- **Managed Identity Attachment** -- created only when `identity` is configured; attaches system-assigned and/or user-assigned identities (required for CMK)
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically, merged with any user tags

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the instance will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **Region availability** -- Managed Redis is newer than classic Redis and not yet in every region; check Azure's product-availability table for the current footprint.
- **For customer-managed keys** (optional): an AzureKeyVaultKey in a purge-protected vault and an AzureUserAssignedIdentity granted wrap/unwrap on the key before deployment.

## Deploy

### Console

Open the deployment store, find **Azure Managed Redis**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Keyless Balanced Cache** preset in the [Presets](#presets) tab to pre-populate the production default posture.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureManagedRedis
metadata:
  name: app-cache
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  clusterName: acme-app-cache
  skuName: BALANCED_B1
  defaultDatabase:
    accessKeysAuthenticationEnabled: false
```

```shell
planton apply -f managed-redis.yaml
```

This creates a 1 GB Balanced instance with high availability (Azure's default), TLS-only access on port 10000, OSS clustering, volatile-lru eviction, and NO access keys -- the keyless posture where clients authenticate with Entra tokens under access-policy-assignment grants. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the instance to a resource group -- and, for CMK, to a Key Vault key and identity -- deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  customerManagedKey:
    keyVaultKeyId:
      valueFrom:
        kind: AzureKeyVaultKey
        name: redis-cmk
        fieldPath: status.outputs.key_id
    userAssignedIdentityId:
      valueFrom:
        kind: AzureUserAssignedIdentity
        name: redis-identity
        fieldPath: status.outputs.identity_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group, key, and identity first, then provisions the instance with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Managed Redis instance. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**SKU family and size** -- `skuName` packs both into one value: BALANCED (B0-B1000) general-purpose, COMPUTE_OPTIMIZED (X3-X700) for high ops/sec on smaller datasets, MEMORY_OPTIMIZED (M10-M2000) for large datasets, FLASH_OPTIMIZED (A250-A4500) for very large datasets tiered onto NVMe flash. The number is the instance's memory in GB (B0 = 0.5 GB). Azure validates size changes against the live instance -- scalable changes apply in place; rejected ones replace the instance. Geo-replication requires BALANCED_B3 or larger.

**Keyless authentication** -- the reverse of classic Redis: `defaultDatabase.accessKeysAuthenticationEnabled` defaults FALSE. Clients authenticate with Microsoft Entra tokens under AzureManagedRedisAccessPolicyAssignment grants -- object ID as the username, a short-lived token as the password, no secret to leak or rotate. Enable keys only for clients that genuinely cannot use Entra tokens; the key outputs populate only while keys are on.

**Clustering policy** -- `defaultDatabase.clusteringPolicy` unset deploys OSS_CLUSTER (best throughput, cluster-aware clients required). ENTERPRISE_CLUSTER proxies all shards behind one endpoint so any client works -- and is required by the RediSearch module. Changing the policy later RECREATES the database (data lost), so decide at creation.

**Modules** -- up to 4 from `RediSearch`, `RedisJSON`, `RedisBloom`, `RedisTimeSeries`, each at most once. RediSearch requires ENTERPRISE_CLUSTER and NO_EVICTION. The module set is fixed for the database's lifetime -- changing it recreates the database.

**Persistence XOR geo-replication** -- AOF (`persistenceAppendOnlyFileBackupFrequency: 1s`) logs every write; RDB (`persistenceRedisDatabaseBackupFrequency: 1h|6h|12h`) snapshots periodically; set at most one. A database joining a geo-replication group (`geoReplicationGroupName`) can use neither -- the group's cross-region replicas are its durability story -- and supports only the RediSearch/RedisJSON modules.

**Network isolation** -- Managed Redis has NO VNet injection and NO IP firewall; Private Link is the only isolation mechanism. Disable `publicNetworkAccessEnabled` and attach an AzurePrivateEndpoint for private-only access.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureKeyVaultKey** (optional, CMK) | `customerManagedKey.keyVaultKeyId` | `status.outputs.key_id` |
| **AzureUserAssignedIdentity** (optional) | `customerManagedKey.userAssignedIdentityId`, `identity.userAssignedIdentityIds` | `status.outputs.identity_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `managed_redis_id` | Azure Resource Manager ID of the cluster | AzureManagedRedisAccessPolicyAssignment, AzureManagedRedisGeoReplication, AzurePrivateEndpoint |
| `managed_redis_name` | The instance's name (the DNS label) | Sibling resources addressing the cluster within its group |
| `region` | The instance's region | Multi-region composition |
| `resource_group_name` | The instance's resource group | Sibling resource placement |
| `hostname` | `{clusterName}.{region}.redis.azure.net` | Keyless (Entra) clients need only this and the port |
| `database_id` | The default database's ARM ID | The scope Entra grants and geo-links operate on |
| `port` | The TCP port (10000 -- never classic Redis's 6379/6380) | Every client's connection config |
| `primary_access_key` | Primary key (SECRET; empty in the keyless default) | Redis client password, only when keys are enabled |
| `secondary_access_key` | Secondary key (SECRET; empty in the keyless default) | Zero-downtime key rotation |
| `identity_principal_id` | The system-assigned identity's principal ID | Role assignments |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Keyless production cache** -- A Balanced instance with high availability and no access keys; every client rides an Entra grant. Start from the **Keyless Balanced Cache** preset.

**Search + JSON document store** -- RediSearch + RedisJSON on a Memory Optimized instance with Enterprise clustering and noeviction -- a queryable document store with Redis latency. Start from the **Search + JSON Document Store** preset.

**Geo-replicated group member** -- One member of an active multi-primary group: BALANCED_B3+, a shared group name, no persistence. Deploy one per region, then link them with AzureManagedRedisGeoReplication. Start from the **Geo-Replicated Group Member** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the instance is created
- [**Azure Managed Redis Access Policy Assignment**](/infra-catalog/azure-managed-redis-access-policy-assignment) -- grants an Entra identity data-plane access (how clients connect in the keyless default)
- [**Azure Managed Redis Geo Replication**](/infra-catalog/azure-managed-redis-geo-replication) -- links instances declaring the same group name into an active multi-primary group
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) -- the customer-managed key that wraps the data-encryption key
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- the wrap/unwrap identity for CMK, and the principal granted data-plane access
- [**Azure Private Endpoint**](/infra-catalog/azure-private-endpoint) -- private connectivity with the public endpoint disabled
