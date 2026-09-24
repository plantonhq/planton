# DigitalOcean Database Read Replica

Creates a single-node read-only replica of a DigitalOcean managed database cluster -- in the primary's region for read scaling or in a different region for geographically local reads -- with optional VPC placement and custom storage. Replicas follow PostgreSQL and MySQL primaries and bill hourly like a second single-node cluster of their size slug, from creation, regardless of read traffic. Size and storage grow in place; every other field -- including tags -- replaces the replica with a fresh seed from the primary, so the placement decisions below are effectively one-way doors.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Read-Only Replica** -- a full single-node managed database of the configured size, continuously following the primary
- **VPC Network Attachment** -- configured only when `vpc` is provided; places the replica's private endpoint in the named VPC (the REPLICA region's VPC for cross-region replicas)
- **Custom Storage** -- configured only when `storageSizeMib` is provided; must stay at or above the primary's storage
- **DigitalOcean Tags** -- your `tags` plus resource metadata tags applied automatically -- note replica tags are create-only upstream (a retag replaces the replica), and DigitalOcean caps the combined tags (joined by commas) at 255 characters with the metadata tags counting toward it; both provisioners fail fast with the arithmetic before anything is created

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **A DigitalOceanDatabaseCluster** -- the primary (PostgreSQL or MySQL), referenced by name (or an existing cluster's UUID as a literal).

### DigitalOcean Account

- **Budget for a second node** -- the replica bills hourly like a single-node cluster of its slug, from creation, regardless of read traffic.
- **A VPC in the replica's region** (only for private networking) -- cross-region replicas join the REPLICA region's VPC, not the primary's; reference a DigitalOceanVpc Cloud Resource or pass a VPC UUID.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Database Read Replica**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Same-Region Read Replica** preset in the [Presets](#presets) tab to place a matching-size replica beside its primary.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseReplica
metadata:
  name: orders-read-replica
  org: acme-corp
  env: prod
spec:
  cluster:
    valueFrom:
      kind: DigitalOceanDatabaseCluster
      name: orders-postgres
      fieldPath: status.outputs.cluster_id
  replicaName: orders-read-replica
  region: nyc3
  size: db-s-1vcpu-1gb
```

```shell
planton apply -f do-database-replica.yaml
```

This creates a same-region read replica of the referenced cluster with its own read-only endpoint. A Stack Job tracks the provisioning in real time.

### InfraChart

For a cross-region replica with private networking, wire both the primary and the replica-region VPC via ValueFromRef:

```yaml
spec:
  cluster:
    valueFrom:
      kind: DigitalOceanDatabaseCluster
      name: orders-postgres
      fieldPath: status.outputs.cluster_id
  replicaName: orders-eu-replica
  region: ams3
  size: db-s-1vcpu-2gb
  vpc:
    valueFrom:
      kind: DigitalOceanVpc
      name: ams-private-network
      fieldPath: status.outputs.vpc_id
```

The InfraPipeline resolves the dependency graph, deploys the cluster and VPC first, then seeds the replica from the primary.

## Key Configuration

These are the most important decisions when configuring a read replica. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Region and size are required by design** -- The upstream provider marks both optional ("inherit the primary's") but reads them back without computing them, so an omitted value drifts on the next apply -- and because region is create-only, that drift schedules a full replica replacement. This spec requires both fields: "inherit from the primary" is expressed by writing the primary's region and size explicitly, same outcome with none of the landmine.

**Size floor and resize** -- `size` must be at least the primary's slug (API-enforced). Growing it resizes the replica in place; shrinking is not supported by DigitalOcean. When the primary resizes, revisit the replica's size too -- it must stay at or above the primary's.

**Storage** -- `storageSizeMib` must stay at or above the primary's storage, or replication can fail when the primary outgrows the replica's disk. It grows in place together with size changes; DigitalOcean rounds and enforces per-slug bounds server-side. Grow storage alongside size when the new slug's default would fall below the current allocation -- storage never decreases.

**Tags replace the replica** -- Replica tags are create-only upstream: changing the list replaces the replica, DigitalOcean seeds a fresh one from the primary, and the old endpoint dies. No primary data is at risk, but read consumers see the endpoint churn and the reseed takes cluster-create time. Decide the tag set at birth, and account for the resource-metadata tags the modules add automatically when planning any future retag.

**VPC placement** -- A cross-region replica joins the REPLICA region's VPC (the primary's VPC does not span regions), so wire `vpc` to a VPC in `region`. Both fields are create-only; moving a replica between regions or networks is a replace. When `vpc` is unset, DigitalOcean uses the target region's default VPC.

**What a replica is not** -- Not automatic failover: DigitalOcean offers manual console promotion, not managed HA (the primary's standby nodes are the HA story). Not a backup: replicas follow deletes and corruption faithfully; the primary's backups are the recovery path. Not writable: writes go to the primary, always.

**Creation takes cluster time** -- A replica seeds from the primary's backup chain; budget the same creation time you budget for a cluster, and expect brand-new primaries to add delay while their first backup completes (the modules retry through that window automatically).

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanDatabaseCluster** | `cluster` | `status.outputs.cluster_id` |
| **DigitalOceanVpc** (optional) | `vpc` | `status.outputs.vpc_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `replica_id` | The replica's own UUID (DigitalOcean addresses reads and deletes by (cluster, name); the UUID serves resizes and cross-references) | API operations, monitoring |
| `host` / `private_host` | Public and private-network hostnames of the replica endpoint | Read-side application connection strings |
| `port` | Port the replica listens on | Read-side application connection strings |
| `database` / `user` | The default database and user served by the replica | Application configuration |
| `password` | The default user's password (secret) | Application authentication |
| `uri` / `private_uri` | Full connection URIs including credentials (secrets) | Read-side application configuration wired as secrets |

`cluster_id` and `replica_name` are also echoed for addressing and verification -- replica reads and deletes go through that pair.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Same-region read scaling** -- a replica beside its primary, matching region and size, taking reports, search indexers, and read APIs off the primary's back. The primary's slug is the natural size floor. Start from the **Same-Region Read Replica** preset.

**Cross-region reads with VPC placement** -- a replica in a second region joined to that region's VPC, giving remote readers local latency and a manually promotable warm copy -- at the cost of a second node's bill and create-only placement. Start from the **Cross-Region Replica with VPC Placement** preset.

## Works With

- [**DigitalOcean Database Cluster**](/cloud-catalog/digital-ocean-database-cluster) -- the primary this replica follows, wired via the `cluster` reference
- [**DigitalOcean VPC**](/cloud-catalog/digital-ocean-vpc) -- private networking in the replica's region, wired via the `vpc` reference
