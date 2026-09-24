# GcpManagedKafkaConnectCluster — Terraform Implementation

This directory contains the Terraform implementation for a Managed Kafka Connect cluster from the Planton spec: one `google_project_service` (API enablement) and one `google_managed_kafka_connect_cluster`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the id defaulted from `metadata.name`, subnet self links trimmed, the attribution label merge |
| `main.tf` | `google_project_service`, `google_managed_kafka_connect_cluster` |
| `outputs.tf` | `name`, `connect_cluster_id`, `location` |

## Send Posture

- **`connect_cluster_id`** -- defaulting to `metadata.name`.
- **Capacity** -- the int64 counts sent as decimal strings.
- **`primary_subnet`** -- a `GcpSubnetwork` self link trimmed of the compute API prefix; `dns_domain_names` sent only when listed.
- **`additional_subnets`** -- not offered: deprecated by Google in favor of more network configs.
- **Labels** -- user labels merged under the attribution labels.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
