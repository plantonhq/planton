# GcpMemorystoreInstance Guide

The judgment this guide protects: the new-generation Memorystore is
bought by TOPOLOGY (mode × node_type × shard_count), connected by PSC
automation that must be authorized BEFORE the instance exists, and
guarded by a deletion protection that defaults on. Most of its sharp
edges are ordering edges.

## The connection policy comes first, always

PSC is the only connectivity, and the automation refuses to place
endpoints without a `GcpServiceConnectionPolicy` for the
`gcp-memorystore` service class on each `pscAutoConnections` network in
this region. Deploy the policy before the instance and keep it alive
longer — deleting it mid-life strands the endpoints. In a chart, the
policy → instance edge is the load-bearing dependency.

## Buy capacity as topology

`CLUSTER_DISABLED` + 1 shard is a drop-in standalone for any client;
`CLUSTER` requires cluster-aware drivers and buys horizontal scale.
Node types tier cleanly: `SHARED_CORE_NANO`/`CUSTOM_PICO`/`CUSTOM_MICRO`/
`CUSTOM_MINI` are burstable dev/test metal; `STANDARD_SMALL`/`STANDARD_LARGE`
balance CPU and memory; `HIGHCPU_MEDIUM` leans compute;
`HIGHMEM_MEDIUM` through `HIGHMEM_2XLARGE` lean memory for large
keyspaces. `shardCount` and `replicaCount` resize IN PLACE — the mode
and node type do not, so pick the shape, then scale the counts.

## Persistence: RDB bounds loss, AOF bounds it tighter

RDB snapshots bound failover loss to the snapshot period; AOF
(`appendFsync: EVERY_SEC` is the sane middle) bounds it to about a
second at an I/O cost. `ALWAYS` is rarely worth it on a cache — if
every write is precious, the data belongs in a database with a cache in
front, not in the cache alone.

## TLS and the CA decision

`transitEncryptionMode: SERVER_AUTHENTICATION` turns on TLS;
`serverCaMode` decides which CA signs the server certificate — and it
is immutable. The per-instance Google CA (default) means every client
trusts a CA unique to this instance; `GOOGLE_MANAGED_SHARED_CA` lets a
fleet trust one CA; `CUSTOMER_MANAGED_CAS_CA` + `serverCaPool` puts the
chain under your Certificate Authority Service pool — the regulated-
environment answer, and the pool must live in the instance's region.

## ACLs are a shared policy, not per-instance rules

`aclPolicy` attaches a Memorystore ACL policy — Valkey users, key
patterns, and allowed commands authored once and shared by every
instance in the region — by its full resource name. Leave it empty for
the built-in default ACL (one user, full access, gated by
`authEnabled`). Attaching or swapping a policy is an in-place update,
and the instance's `isAclPolicyInSync` status says when the new rules
have reached every node.

## Maintenance and self-service patching

The weekly window pins day + hour (UTC) and always starts on the hour —
the API rejects finer start times, so stagger this cache against the
systems it fronts in whole-hour steps (Redis, by contrast, takes
minutes). `maintenanceVersion` is update-only and forward-only: set it
on an EXISTING instance to take a patch on your schedule; setting it at
create is rejected by the API.

## DR is a composition, not a flag

Cross-region replication deploys the PRIMARY first (listing its
secondaries by full resource path — another instance's `name` output),
then each SECONDARY pointing back. A secondary serves reads only until
promoted; the role exchange during a planned switchover is an in-place
update on both ends. Seeding (`gcsSource` XOR `managedBackupSource`)
happens at creation only — it is how you migrate INTO an instance, not
how you back one up (that is `automatedBackupConfig`).

## The two destroy guards do different jobs

`deletionProtectionEnabled` defaults TRUE: destroy fails until you
explicitly set it false and apply — the deliberate two-step. Dev
fixtures arm it false so teardown works; production presets keep it on.
`deletionPolicy` acts after that gate: `PREVENT` as a second
independent lock, `ABANDON` to leave the instance running (and billing)
under another management plane.
