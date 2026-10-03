# Azure Managed Redis Geo Replication

Links Azure Managed Redis instances into an ACTIVE geo-replication group: every member accepts writes in its own region and Azure merges the datasets with conflict-free semantics -- multi-primary, not the classic primary/warm-standby model. Applications write locally everywhere and read their own region's instance. Group membership is a first-class resource because Azure mutates the replication state of EVERY member when the group changes -- one linking resource manages the whole group, through any one member.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Geo-Replication Links** -- the group-wide linking operation joining the declared members (the managing instance plus 1-4 linked instances, a group of up to 5) into one active replica set. Deleting the resource unlinks all members; each keeps its own copy of the data and becomes independent. Removing a single member's ID force-unlinks just that member -- the designed region-evacuation workflow.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **Two or more Managed Redis instances** already declaring the SAME `geoReplicationGroupName` on their default databases (set at each instance's creation -- joining later recreates the database).
- **Every member meets the group preconditions** Azure enforces at link time: BALANCED_B3 or larger, no AOF/RDB persistence, and only the RediSearch/RedisJSON modules.

## Deploy

### Console

Open the deployment store, find **Azure Managed Redis Geo Replication**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the managing instance, and the linked members. Start from the **Two-Region Active Pair** preset in the [Presets](#presets) tab for the common active-active shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureManagedRedisGeoReplication
metadata:
  name: global-cache-group
  org: acme-corp
  env: prod
spec:
  managedRedisId:
    valueFrom:
      kind: AzureManagedRedis
      name: global-cache-east
      fieldPath: status.outputs.managed_redis_id
  linkedManagedRedisIds:
    - valueFrom:
        kind: AzureManagedRedis
        name: global-cache-west
        fieldPath: status.outputs.managed_redis_id
```

```shell
planton apply -f geo-link.yaml
```

This links the two members into one active group: both accept writes, Azure merges conflict-free, and each application reads and writes its local region. An Infra Job tracks the provisioning in real time.

### InfraChart

The link is the last edge of a multi-region chart -- the pipeline deploys every member first (same group name each), then the one linking resource:

```yaml
spec:
  managedRedisId:
    valueFrom:
      kind: AzureManagedRedis
      name: global-cache-east
      fieldPath: status.outputs.managed_redis_id
  linkedManagedRedisIds:
    - valueFrom:
        kind: AzureManagedRedis
        name: global-cache-west
        fieldPath: status.outputs.managed_redis_id
    - valueFrom:
        kind: AzureManagedRedis
        name: global-cache-europe
        fieldPath: status.outputs.managed_redis_id
```

## Key Configuration

These are the only two decisions -- the group's real configuration lives on the members. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The managing instance** -- `managedRedisId` picks the member the group is administered through. Linking is reciprocal (every member sees the same group state), so create ONE resource per group, never one per member. Fixed at creation. The managing instance is always a member implicitly and must not repeat in the linked list -- Azure rejects self-links at apply time.

**The linked members** -- `linkedManagedRedisIds` carries the OTHER members, 1 to 4 of them. Adding an ID links that instance into the group; removing one force-unlinks it (it keeps its data and becomes independent) -- membership edits are this kind's day-two operation.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureManagedRedis** (managing) | `managedRedisId` | `status.outputs.managed_redis_id` |
| **AzureManagedRedis** (each member) | `linkedManagedRedisIds` | `status.outputs.managed_redis_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `geo_replication_id` | The group's resource ID -- the managing member's ARM ID (the group has no ARM object of its own) | Audit and cross-referencing |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Two-region active pair** -- The common shape: active-active disaster recovery where losing a region loses no writes accepted in the other. Start from the **Two-Region Active Pair** preset.

**Global active mesh** -- Four regions across continents in one write-anywhere group (the maximum is five members). Start from the **Global Active Mesh** preset.

## Works With

- [**Azure Managed Redis**](/infra-catalog/azure-managed-redis) -- the members being linked; each declares the shared group name at creation
- [**Azure Managed Redis Access Policy Assignment**](/infra-catalog/azure-managed-redis-access-policy-assignment) -- grants are per member; geo-replicated applications grant their identity on every member
