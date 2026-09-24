# GcpManagedKafkaCluster — Pulumi Implementation

This directory contains the Pulumi implementation for a Managed Service for Apache Kafka cluster from the Planton spec: one `gcp.projects.Service` (API enablement) and one `gcp.managedkafka.Cluster`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `cluster` |
| `module/locals.go` | The defaulted cluster id, trimmed subnets, and attribution labels |
| `module/cluster.go` | Enables the API; maps the cluster; exports the outputs |
| `module/outputs.go` | Output key constants (`name`, `cluster_id`, `location`) |

## Send Posture (parity with Terraform)

- **`cluster_id`** -- `spec.cluster_id`, defaulting to `metadata.name`.
- **Capacity** -- the int64 counts sent as decimal strings, as Google's API takes them.
- **Subnets** -- a `GcpSubnetwork` self link trimmed of `https://www.googleapis.com/compute/v1/`.
- **`broker_capacity_config`, `rebalance_config`** -- the lifted leaves, each block sent only when set.
- **`tls_config`** -- sent whenever the spec declares it, even empty (the clearing form); `trust_config` only when CA pools are listed.
- **`public_cluster_config`** -- never sent: an SDK gap at pulumi-gcp v9.37.0 (re-evaluated at pulumi-gcp v10 GA).
- **Labels** -- user labels merged under the platform attribution labels.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
