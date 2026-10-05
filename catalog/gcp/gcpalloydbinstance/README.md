# GCP AlloyDB Instance

Deploys an AlloyDB instance (`google_alloydb_instance`) on an existing cluster. Instances are first-class nodes for scaling read traffic (READ_POOL) or advanced PRIMARY/SECONDARY topologies — separate from the bundled primary in GcpAlloydbCluster.

## What Gets Created

When you deploy a GcpAlloydbInstance resource, Planton provisions:

- **AlloyDB API enablement** — `alloydb.googleapis.com` on the target project (idempotent; never disabled on destroy)
- **AlloyDB instance** — a `google_alloydb_instance` on the referenced cluster

## Prerequisites

- **An existing AlloyDB cluster** — referenced via `cluster` (a [GcpAlloydbCluster](/docs/catalog/gcp/gcpalloydbcluster) resource or a literal cluster resource path)
- **For READ_POOL** — a healthy PRIMARY instance on the cluster (bundled with GcpAlloydbCluster or otherwise)
- **GCP credentials** for the deploying principal — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set

## Quick Start

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpAlloydbInstance
metadata:
  name: orders-read-pool
spec:
  cluster:
    valueFrom:
      kind: GcpAlloydbCluster
      name: my-orders-cluster
      fieldPath: status.outputs.cluster_id
  instanceId: orders-read-pool
  readPoolConfig:
    nodeCount: 2
  cpuCount: 4
```

```shell
planton apply -f instance.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `cluster` | `StringValueOrRef` | Hosting cluster (ref → `cluster_id`). Immutable. |
| `instanceId` | `string` | Instance ID within the cluster. Immutable. |

### Key Optional Fields

| Field | Default | Description |
|-------|---------|-------------|
| `instanceType` | `READ_POOL` | `PRIMARY`, `READ_POOL`, or `SECONDARY`. Immutable. |
| `readPoolConfig.nodeCount` | — | Required for READ_POOL. Number of read nodes. |
| `cpuCount` / `machineType` | — | Mutually exclusive machine sizing. |
| `requireConnectors` / `sslMode` | — | Client connection hardening. |
| `activationPolicy` | `ALWAYS` | Stop/start lever: `NEVER` stops the instance (config and storage retained, compute billing stops). |
| `enablePublicIp` / `authorizedExternalNetworks` | — | Public IP arms (networks require public IP). |
| `pscInstanceConfig` | — | Private Service Connect settings. |

### Validation Rules

- `cpu_count` and `machine_type` are mutually exclusive.
- READ_POOL requires `read_pool_config.node_count >= 1`.
- `read_pool_config` applies to READ_POOL only.
- `authorized_external_networks` requires `enable_public_ip`.

## Outputs

| Output | Description |
|--------|-------------|
| `instance_name` | Fully qualified instance resource path |
| `ip_address` | Private IP connection endpoint |
| `state` | Current instance state (e.g. READY) |

## Deployment Methods

- Pulumi: [`iac/pulumi/README.md`](iac/pulumi/README.md)
- Terraform: [`iac/tf/README.md`](iac/tf/README.md)

## Related Kinds

- [GcpAlloydbCluster](/docs/catalog/gcp/gcpalloydbcluster) — the cluster this instance attaches to
- [GcpAlloydbUser](/docs/catalog/gcp/gcpalloydbuser) — per-application credentials on the same cluster

## Additional Resources

- [AlloyDB read pool instances](https://cloud.google.com/alloydb/docs/read-pool-overview)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
