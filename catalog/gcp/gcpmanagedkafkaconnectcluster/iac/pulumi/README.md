# GcpManagedKafkaConnectCluster — Pulumi Implementation

This directory contains the Pulumi implementation for a Managed Kafka Connect cluster from the Planton spec: one `gcp.projects.Service` (API enablement) and one `gcp.managedkafka.ConnectCluster`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `connectCluster` |
| `module/locals.go` | The defaulted id and attribution labels |
| `module/connect_cluster.go` | Enables the API; maps the Connect cluster; exports the outputs |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`connect_cluster_id`** -- defaulting to `metadata.name`.
- **Capacity** -- the int64 counts sent as decimal strings.
- **`primary_subnet`** -- a `GcpSubnetwork` self link trimmed of the compute API prefix; `dns_domain_names` sent only when listed.
- **`additional_subnets`** -- not offered: deprecated by Google in favor of more network configs.
- **Labels** -- user labels merged under the attribution labels.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
