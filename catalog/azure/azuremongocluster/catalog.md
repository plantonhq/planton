# Azure Mongo Cluster (Cosmos DB for MongoDB vCore)

Deploys an Azure Cosmos DB for MongoDB vCore cluster -- a real MongoDB engine on dedicated vCore tiers, wire-protocol compatible with community drivers, with vertical tiers from a free sandbox to M200, optional sharding, zone-redundant high availability, geo replicas, and point-in-time restore. A cluster is created in one of three modes: Default builds fresh from the sizing fields, GeoReplica builds a cross-region read replica, and PointInTimeRestore clones from backup history -- and Azure never returns the mode on reads, so changing it always replaces the cluster. Continuous backup (35 days) is automatic.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Mongo vCore cluster** -- the cluster itself, in the creation mode the spec declares: compute tier and storage per shard, MongoDB version, high availability, authentication methods, identities, and optional customer-managed-key encryption
- **Firewall rules** -- one per named client-IP range in `firewallRules`
- **Azure Tags** -- Planton-derived metadata tags merged with the manifest's `tags` (user values win on key conflicts)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **The administrator password's source** -- a managed secret (`$secret/<slug>`) or an AzureKeyVaultSecret output; a literal is acceptable only in throwaway environments.

### Azure Subscription

- **A resource group** -- reference an AzureResourceGroup Infra Component or pass an existing group's name.
- **A globally unique name** -- the name becomes the cluster's public hostname (`{name}.mongocluster.cosmos.azure.com`); a taken name fails at deploy time. Prefix with your org.
- **For CMK encryption** (optional) -- an AzureKeyVaultKey (referenced by its VERSIONLESS ID) and an AzureUserAssignedIdentity with unwrap/wrap permissions on the vault, granted BEFORE the cluster is created; Azure validates both at deploy time.
- **For a GeoReplica** -- the source cluster must carry the `GeoReplicas` preview feature (a create-time, ForceNew list on the source).

## Deploy

### Console

Open the deployment store, find **Azure Mongo Cluster**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields: the creation mode, sizing, authentication, and network posture. Start from the **Production Cluster** or **Free Sandbox** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMongoCluster
metadata:
  name: orders-db
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: prod-data
      fieldPath: status.outputs.resource_group_name
  name: acme-orders-db
  region: eastus
  administratorUsername: mongoadmin
  administratorPassword:
    value: $secret/mongo-admin-password
  version: "8.0"
  computeTier: M30
  storageSizeInGb: 128
  shardCount: 1
  highAvailabilityMode: ZoneRedundantPreferred
  authenticationMethods:
    - NativeAuth
    - MicrosoftEntraID
  firewallRules:
    - name: office
      startIpAddress: 203.0.113.0
      endIpAddress: 203.0.113.255
```

```shell
planton apply -f mongo-cluster.yaml
```

This creates a Default-mode MongoDB 8.0 cluster on dedicated M30 compute with zone-redundant HA, both authentication worlds enabled, and one office IP range allowed through the firewall. An Infra Job tracks the provisioning in real time.

### InfraChart

When the cluster's dependencies deploy in the same InfraChart, wire them with ValueFromRef:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: prod-data
      fieldPath: status.outputs.resource_group_name
  name: acme-orders-db
  region: eastus
  administratorUsername: mongoadmin
  administratorPassword:
    valueFrom:
      kind: AzureKeyVaultSecret
      name: mongo-admin-password
      fieldPath: status.outputs.secret_id
  version: "8.0"
  computeTier: M30
  storageSizeInGb: 128
  shardCount: 1
  highAvailabilityMode: Disabled
```

The InfraPipeline resolves the dependency graph, deploys the resource group and the vaulted password first, then provisions the cluster -- and Azure Mongo Cluster User grants downstream can reference this cluster's `mongo_cluster_id` in the same chart.

## Key Configuration

These are the most important decisions when configuring a Mongo vCore cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**vCore or RU? Answer it before anything else** -- Azure sells two MongoDB-compatible products. This one (vCore) is a REAL MongoDB engine on dedicated compute: predictable hourly cost, full query surface, community drivers and tools work unchanged. The other (a Cosmos DB account with the Mongo API) is the request-unit model: serverless economics, per-operation pricing, a compatibility layer rather than the engine. Lift-and-shift workloads and anything using change streams, transactions, or aggregation-heavy queries belong HERE; spiky tiny workloads that would idle a vCore belong there.

**The mode you create with is the mode you die with** -- Azure never returns `createMode` on reads, so the provider replaces the cluster on ANY mode change; there is no "promote this replica" or "re-parent this clone" through configuration. Promotion of a GeoReplica is an Azure-side operation (portal/CLI); after promoting, import the promoted cluster rather than editing the replica's manifest into Default mode -- that edit is a replacement that would destroy it.

**Free and M25 are sandboxes with walls** -- the burstable tiers reject zone-redundant HA and sharding past one shard, and the Free tier additionally refuses `MicrosoftEntraID` authentication (all enforced at manifest time). One Free cluster per subscription is Azure's own cap, and upgrades away from Free/M25 stage a tier-first update the provider performs itself -- expect two update waves in one apply. Never let a Free-tier proof-of-concept quietly become production: the tier ceiling arrives as throttling, not as an error.

**Password rotation is an update; username is forever** -- `administratorPassword` rotates in place (reference a secret store so rotation is a reference change, not a manifest edit); `administratorUsername` is create-only. The rotation-friendly posture for applications is to not use the administrator at all: grant each app's identity an Azure Mongo Cluster User and keep the admin credential for break-glass. Include `MicrosoftEntraID` in `authenticationMethods` BEFORE creating those grants.

**Storage grows, never shrinks -- and sharding is forever** -- `storageSizeInGb` moves up in place with no path down short of dump-and-restore; size for the working set, not the someday set. `shardCount` is ForceNew (re-sharding is not an in-place operation), so shard only with evidence -- vertical scaling covers most workloads. `storageType` (PremiumSSD vs PremiumSSDv2) is also create-only.

**Identity and CMK transitions are one-way doors** -- adding the FIRST user-assigned identity or removing the LAST one replaces the cluster (Azure rejects the in-place transition; changing the set between non-empty states updates in place). The `customerManagedKey` block cannot be added, removed, or changed after create, and its unwrap identity must be listed in `userAssignedIdentityIds` with vault permissions granted before create. Enabling the Data API after create is staged automatically; turning it back OFF replaces the cluster.

**The connection strings substitute real credentials** -- Azure returns connection strings with a `<user>:<password>` placeholder; the engines substitute the actual administrator credentials into the outputs. Treat `connection_string` and `connection_strings` as secrets end to end and wire consumers by reference so a password rotation propagates.

**Network posture** -- `publicNetworkAccessEnabled` defaults to true; `firewallRules` name the IPv4 ranges allowed through while it is. Rule names are the rules' identity (renaming replaces the rule), and `0.0.0.0`-`255.255.255.255` allows everything including other Azure services -- use it knowingly.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureKeyVaultSecret** (recommended) | `administratorPassword` | `status.outputs.secret_id` |
| **AzureUserAssignedIdentity** (identity/CMK) | `userAssignedIdentityIds`, `customerManagedKey.userAssignedIdentityId` | `status.outputs.identity_id` |
| **AzureKeyVaultKey** (CMK encryption) | `customerManagedKey.keyVaultKeyId` | `status.outputs.versionless_id` |
| **AzureMongoCluster** (GeoReplica mode) | `sourceServerId` | `status.outputs.mongo_cluster_id` |
| **AzureMongoCluster** (PointInTimeRestore mode) | `restore.sourceId` | `status.outputs.mongo_cluster_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `mongo_cluster_id` | The cluster's ARM ID | An Azure Mongo Cluster User's `mongoClusterId`; a replica's `sourceServerId`; a restore's `restore.sourceId` |
| `connection_string` | The primary MongoDB URI with administrator credentials substituted in (sensitive; empty when the cluster has no native administrator) | Application configuration, wired by reference |
| `connection_strings` | Every connection string Azure publishes, keyed by Azure's name (primary plus per-replica and per-mode variants; sensitive) | Read-preference-specific and replica-aware wiring |

`mongo_cluster_name` is also exported but only echoes the manifest's `name` back.

## Common Patterns

**Production cluster** -- Default mode on M30 with zone-redundant HA, MongoDB 8.0, and both authentication worlds enabled: applications onboard through Entra grants while the native administrator stays break-glass. Start from the **Production Cluster** preset.

**Zero-cost sandbox** -- the Free tier with minimum storage, no HA, and a single-address firewall rule for a dev machine; one per subscription, and never quietly promoted to production. Start from the **Free Sandbox** preset.

**Disaster recovery and cloning** -- a GeoReplica in a second region for read scale-out and failover (source needs the `GeoReplicas` preview feature at create), and PointInTimeRestore clones for incident forensics or environment refreshes from the automatic 35-day backup history; both inherit sizing from the source.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- the group the cluster lives in
- [**Azure Mongo Cluster User**](/infra-catalog/azure-mongo-cluster-user) -- Entra-identity grants for passwordless application access
- [**Azure Key Vault Secret**](/infra-catalog/azure-key-vault-secret) -- the vaulted administrator password, referenced not copied
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) -- the customer-managed encryption key (versionless reference)
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- attached identities, including the CMK unwrap identity
