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

`authorizationMode` and `transitEncryptionMode` are immutable. Both
engines send them explicitly with Google's defaults
(`AUTH_MODE_DISABLED`, `TRANSIT_ENCRYPTION_MODE_DISABLED`) when the
manifest leaves them unset, so the posture never depends on a provider
default -- and so changing your mind later means a new cluster. Decide at
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
files in Cloud Storage through `gcsSource`). `crossClusterReplicationConfig`
answers the regional question: a `SECONDARY` in another region replicates
continuously from a `PRIMARY` and serves reads until it is promoted.

## What a destroy does

`deletionProtectionEnabled` defaults to true: a destroy fails until the
manifest flips it, and both engines send the value explicitly so a
manifest that never mentions it behaves the same everywhere.
`deletionPolicy: PREVENT` is a second, independent guard for a cache
whose loss would stampede the store behind it; `ABANDON` walks away from
a running (and billing) cluster. The maintenance window, when set, starts
on the hour -- Google exposes no finer setting.
