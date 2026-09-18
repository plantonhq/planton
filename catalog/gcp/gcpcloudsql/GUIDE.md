# GcpCloudSql Guide

The judgment this guide protects: a Cloud SQL instance is a stateful
singleton whose name burns for a week after deletion and whose hardest
properties — region, CMEK, disk type, private-network attachment — are
immutable. Almost everything else changes in place, but "in place" on a
database usually means a restart. Decide the immutables deliberately,
schedule the restarts, and treat the data-loss levers as the real
configuration surface.

## Edition and tier are cost policy, not sizing

ENTERPRISE is the default for a reason; ENTERPRISE_PLUS buys the 99.99%
HA SLA, the data cache, 35-day PITR reach, and near-zero-downtime
maintenance — pay for it where an outage minute costs more than the
instance does. The `tier` string is mutable (a restart, not a
replacement), so start small and resize on evidence. The data cache and
PSC write-endpoint DNS are Plus-only; the spec's validation says so
before the API would.

## HA is a backup setting wearing an availability badge

`availabilityType: REGIONAL` only works because automated backups (and
binary logs on MySQL) feed the standby — which is why the spec refuses
REGIONAL without them. Decide recovery FIRST: `transactionLogRetentionDays`
bounds how far PITR reaches, `retainedBackups` bounds how many dailies
exist, `retainBackupsOnDelete` and `finalBackup` decide what survives the
instance itself. HA protects against zone loss; PITR protects against
yesterday's bad migration. They are different insurance policies — most
production instances want both.

## Connectivity: pick the path before the VPC hardens

Private IP requires the VPC to already carry a service networking
connection (compose `GcpGlobalAddress` + `GcpServiceNetworkingConnection`
first — the provider fails fast, and a failed create still burns the
name). Private network attachment is one-way: it can be set or changed,
never removed. And the attachment outlives the instance: after a
private-IP instance is deleted, Cloud SQL's producer releases its hold
on the service networking connection asynchronously — live-measured at
over 40 minutes in 2026-08 — so a same-pass teardown of instance +
connection + VPC must either budget for that lag or abandon the
connection and retire the whole network (the connection kind's GUIDE
covers both shapes). PSC is the modern alternative when peering topology
is the problem — with `autoDnsEnabled`, consumers resolve the instance
by name instead of tracking endpoint IPs. A public IP with an empty
`authorizedNetworks` list is safer than it looks: only the Auth Proxy and
IAM-authenticated connectors can use it.

## Scale reads with the right resource

Three read-scaling shapes, in order of operational weight: a read
REPLICA is its own `GcpCloudSql` node (`masterInstanceName`) you place,
size, and even promote (`instanceType: CLOUD_SQL_INSTANCE`, clearing the
master reference in the same change); a read POOL
(`instanceType: READ_POOL_INSTANCE`) is one endpoint over `nodeCount`
interchangeable nodes — no per-node identity, and `readPoolAutoScale`
sizes it for you (while it does, `nodeCount` is the autoscaler's field,
not yours); cross-region DR is `failoverDrReplicaName` on the primary,
which names the replica a switchover promotes. Replicas for control,
pools for elasticity, DR pairing for continuity. On a replica,
`replicationLagMaxSeconds` (300 to 31536000) makes it recreate itself
once it lags the primary past the threshold for five minutes — the
self-healing answer to a replica that fell too far behind to catch up.
Two input-only opt-ins ride major-version work:
`includeReplicasForMajorVersionUpgrade` upgrades replicas in place with
the primary when `databaseVersion` moves majors, and
`switchTransactionLogsToCloudStorageEnabled` moves PITR transaction logs
off the data disk into Cloud Storage for longer retention. Cloud SQL
acts on both and stores neither. `enforceNewSqlNetworkArchitecture` is
the irreversible opt-in older projects need before PSC and outbound
network attachments; set it once and leave it.

## Restores are triggers, not state

`clone`, `restoreBackupContext`, `pointInTimeRestoreContext`, and
`backupdrBackup` describe where data COMES FROM — and on an existing
instance, changing them RUNS the restore. That is the declarative idiom
for an imperative act: powerful in a recovery runbook, dangerous in a
template. Keep restore blocks out of copy-paste manifests; add them
deliberately, watch the operation complete, and remove or freeze them
after.

## Destroy stance

Layers, from softest to hardest: `deletionPolicy: PREVENT` fails a
destroying plan; `deletionProtection` makes the engines refuse;
`deletionProtectionEnabled` makes GCP itself reject deletion from every
surface including the console. `ABANDON` is the release valve — the
instance leaves IaC management and keeps running. Under all of it,
`retainBackupsOnDelete` and `finalBackup` decide whether a deletion is
recoverable at all. Production wants both guards on, backups retained,
and the final backup armed — turned off only in the change that means to
destroy.

## On the diagram

The instance is the data tier's hub: it consumes `GcpProject`,
`GcpVpcNetwork` (private IP), `GcpKmsKey` (CMEK), and optionally another
`GcpCloudSql` (as a replica of it); its `instance_name` output is what
every `GcpCloudSqlDatabase`, `GcpCloudSqlUser`, and replica references.
An instance with no database/user nodes attached is usually a manifest
that forgot its companions.

## Pairs well with

- `GcpCloudSqlDatabase` — one per application; databases outlive
  deployments.
- `GcpCloudSqlUser` — one per service, IAM-typed where possible.
- `GcpServiceNetworkingConnection` + `GcpGlobalAddress` — the private-IP
  prerequisite chain, created before the instance.
- `GcpKmsKey` — CMEK, decided at create time (immutable).
- `GcpCloudSql` — the primary, when this node is its replica.
