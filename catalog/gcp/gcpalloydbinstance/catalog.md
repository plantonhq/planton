# GCP AlloyDB Instance

Deploys a standalone AlloyDB instance (`google_alloydb_instance`) attached to an existing cluster. Read pools scale read traffic independently of the bundled primary in GcpAlloydbCluster. PRIMARY and SECONDARY types exist for advanced topologies; presets target READ_POOL.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **AlloyDB API enablement** (`alloydb.googleapis.com`) on the target project (never disabled on destroy)
- **AlloyDB Instance** — a compute node within an existing cluster (READ_POOL, PRIMARY, or SECONDARY), sized by `cpuCount` XOR `machineType`, with read-pool node count, optional public IP with authorized CIDR ranges, per-instance PSC configuration, database flags, and query insights as declared in the spec

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** — an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GcpAlloydbCluster** — the cluster this instance joins. Reference its `status.outputs.cluster_id`, or pass the full cluster resource path directly.
- **A GCP project** — only needed as an override when the cluster's project differs from the provider default.

## Deploy

### Console

Open the deployment store, find **GCP AlloyDB Instance**, and click **Deploy**. Start from the **Basic Read Pool** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpAlloydbInstance
metadata:
  name: orders-read-pool
  org: acme-corp
  env: prod
spec:
  cluster:
    value: "projects/acme-prod-12345/locations/us-central1/clusters/orders-db"
  instanceId: orders-read-pool
  readPoolConfig:
    nodeCount: 2
  cpuCount: 4
```

```shell
planton apply -f alloydb-instance.yaml
```

This attaches a two-node, 4-CPU read pool to the existing cluster — two nodes spread across zones, so the pool survives a zone outage. An Infra Job tracks the provisioning in real time.

### InfraChart

Wire the instance to a cluster deployed in the same InfraPipeline:

```yaml
spec:
  cluster:
    valueFrom:
      kind: GcpAlloydbCluster
      name: orders-db
      fieldPath: status.outputs.cluster_id
```

## Key Configuration

These are the most important decisions when configuring an AlloyDB instance. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Instance type** — Blank or `READ_POOL` is the common read-scaling shape. `PRIMARY` and `SECONDARY` exist for advanced topologies where the cluster does not bundle the compute node.

**Read pool sizing** — `readPoolConfig.nodeCount` sets how many read nodes serve traffic, and it alone decides availability: 1 node is zonal, 2+ nodes spread across zones for regional read HA. Leave `availabilityType` empty on read pools — the API derives it from the node count and does not store a sent value (`availabilityType` is for PRIMARY/SECONDARY instances).

**Machine sizing** — Set `cpuCount` (2, 4, 8, 16, 32, 64, 96, 128) or `machineType` explicitly — never both.

**Connection security** — `requireConnectors: true` enforces IAM-based auth via AlloyDB Auth Proxy. `sslMode: ENCRYPTED_ONLY` requires TLS.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpAlloydbCluster** | `cluster` | `status.outputs.cluster_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `instance_name` | Fully qualified instance resource name (`projects/{p}/locations/{l}/clusters/{c}/instances/{i}`) | AlloyDB Auth Proxy connections, monitoring dashboards |
| `ip_address` | Private IP of the instance | Application read connection strings (port 5432) |
| `state` | Instance state (`READY`, `CREATING`) | Deployment validation, health checks |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic read pool** — single-node ZONAL pool for dev/staging read offload. Start from **Read Pool Basic**.

**HA read pool** — REGIONAL pool with two nodes for production read scaling. Start from **Read Pool HA**.

**Production read pool** — connectors enforced, TLS-only, query insights enabled. Start from **Read Pool Production**.

## Works With

- [**GCP AlloyDB Cluster**](/infra-catalog/gcp-alloydb-cluster) — parent cluster this instance attaches to
- [**GCP AlloyDB User**](/infra-catalog/gcp-alloydb-user) — application credentials on the same cluster
- [**GCP Project**](/infra-catalog/gcp-project) — optional project override when the cluster lives elsewhere
