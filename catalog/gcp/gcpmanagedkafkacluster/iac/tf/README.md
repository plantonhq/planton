# GcpManagedKafkaCluster — Terraform Implementation

This directory contains the Terraform implementation for a Managed Service for Apache Kafka cluster from the Planton spec: one `google_project_service` (API enablement) and one `google_managed_kafka_cluster`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the cluster id defaulted from `metadata.name`, subnet self links trimmed to Google's path form, null-for-empty optionals, the attribution label merge |
| `main.tf` | `google_project_service`, `google_managed_kafka_cluster` |
| `outputs.tf` | `name`, `cluster_id`, `location` |

## Send Posture

- **`cluster_id`** -- `spec.cluster_id`, defaulting to `metadata.name`.
- **Capacity** -- the int64 counts sent as decimal strings, as Google's API takes them.
- **Subnets** -- a `GcpSubnetwork` self link trimmed of `https://www.googleapis.com/compute/v1/`.
- **`broker_capacity_config`, `rebalance_config`** -- the lifted leaves, each block sent only when set.
- **`tls_config`** -- sent whenever the spec declares it, even empty (the clearing form); `trust_config` only when CA pools are listed.
- **`public_cluster_config`** -- never sent: an SDK gap at pulumi-gcp v9.37.0 (re-evaluated at pulumi-gcp v10 GA).
- **Labels** -- user labels merged under the platform attribution labels.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
