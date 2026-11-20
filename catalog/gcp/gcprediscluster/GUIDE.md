# GcpRedisCluster Guide

The judgment this guide protects: a Redis Cluster is paid for by the node
and reached only over Private Service Connect. Size it for the keyspace,
decide how it is reached before it exists, and lock the two security modes
at creation -- they cannot be changed afterwards. Everything else resizes
in place.

## Three Redis-shaped kinds, one decision

The catalog has three managed in-memory stores on Google Cloud.
`GcpRedisInstance` is the legacy single-node Memorystore for Redis,
reached over VPC peering: fine for a small cache, capped at one node.
`GcpMemorystoreInstance` is the new-generation service running the Valkey
engine. This kind is Memorystore for Redis Cluster: the Redis engine,
sharded, replicated, and horizontally scaled. Pick it when a cache has
outgrown one node, when zero-downtime resizing matters, or when clients
speak the OSS Cluster protocol.

## Shards, replicas, nodes, and the bill

Every shard is at least one node; every replica is one more node per
shard. `shardCount x (1 + replicaCount)` nodes bill by the hour at the
`nodeType`'s rate from the moment the cluster exists, serving or idle. The
smallest cluster is one `REDIS_SHARED_CORE_NANO` shard with no replicas.
Replicas are what make a shard survive a node failure -- without one, a
failed shard's slice of the keyspace is unavailable until Google recovers
it. All three levers resize in place; Google rebalances slots across the
new shard count without downtime.

## Connectivity is decided before creation

A cluster is reached only over Private Service Connect. There are two
ways to get endpoints, and the manifest chooses.

With `pscConfigs` naming a consumer network, Google's service connectivity
automation places the endpoints -- a discovery address (and primary and
reader addresses) in that VPC. The automation acts only when a
`GcpServiceConnectionPolicy` with `serviceClass: gcp-memorystore-redis`
already exists on that network in the cluster's region: deploy the policy
first, or creation fails with a connectivity error. Clients bootstrap from
`discovery_endpoint_address`.

With `pscConfigs` empty, the cluster publishes service attachments and
nothing else. Consumers in any VPC or project then reserve an address,
create a regional `GcpGlobalForwardingRule` (empty scheme) targeting each
attachment handle the cluster exposes (`discovery_service_attachment`,
`primary_service_attachment`, `reader_service_attachment`), and register
those rules on the cluster through `GcpRedisClusterEndpointSet`. This is
the multi-VPC pattern and the only way to reach a cluster from a network
the automation cannot own.

## Two modes are frozen at creation

`authorizationMode` and `transitEncryptionMode` are immutable: the
provider replaces the cluster on a change, though Google's API does not
mark them so. Both engines send them explicitly with Google's defaults
(`AUTH_MODE_DISABLED`, `TRANSIT_ENCRYPTION_MODE_DISABLED`) when the
manifest leaves them unset, so the posture never depends on a provider
default -- and so changing your mind later means a new cluster. A blank
is a value: removing either field from an IAM or TLS cluster plans a
replacement with an unauthenticated or plaintext one. Decide at
creation: `AUTH_MODE_IAM_AUTH` puts every command behind an IAM identity;
`TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION` turns on TLS, with
`serverCaMode` choosing which certificate authority clients trust
(Google's per-cluster CA, Google's shared CA, or your own pool in
Certificate Authority Service).

## Durability is layered

Persistence and backups answer different questions. `persistenceConfig`
decides whether a node restart keeps its data: `RDB` snapshots the whole
keyspace on a schedule, `AOF` logs every write with `appendFsync`
choosing how often it reaches disk. `automatedBackupConfig` takes a daily
backup into a managed collection with a retention you choose; a new
cluster can be seeded from one through `managedBackupSource` (or from RDB
files in Cloud Storage through `gcsSource`). Seeding happens once, but the
seed block stays in the manifest for the cluster's life: removing it
replaces the cluster. `crossClusterReplicationConfig` answers the regional
question: a `SECONDARY` in another region, with the primary's shard
count, replicates continuously from a `PRIMARY` and serves reads until it
is promoted.

## Turning something off is a value, not a removal

On a cluster that exists, removing `persistenceConfig`,
`crossClusterReplicationConfig`, `nodeType`, or `serverCaMode` keeps
whatever Google holds: the provider reads each back and sees no change.
Write the off-value instead -- `persistenceConfig.mode: DISABLED`, or
`clusterRole: NONE` with no primary or secondaries to leave a pair --
and write a node type or CA mode explicitly. Switching persistence
modes takes the new mode's block in full, since a field left out keeps
Google's value, and an RDB snapshot anchor, once set, cannot be cleared.
`zoneDistributionConfig` cannot change at all: a change is planned in
place, never sent, and shown again on every plan. `maintenancePolicy`
and `automatedBackupConfig` behave the other way: removing either clears
it.

## What a destroy does

The provider reads `deletionPolicy` before it calls Google.
`PREVENT` fails the destroy right there -- the guard for a cache whose
loss would stampede the store behind it. `ABANDON` walks away from a
running (and billing) cluster without calling Google, so it succeeds
even with deletion protection on. Only `DELETE` reaches Google, where
`deletionProtectionEnabled` (default true) refuses it: turn protection
off and apply that change on its own first, since a change made in the
same destroy run never reaches Google. Both engines send the protection
value explicitly so a manifest that never mentions it behaves the same
everywhere. The maintenance window, when set, starts
on the hour -- Google exposes no finer setting.
