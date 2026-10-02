# GcpManagedKafkaAcl — Terraform Implementation

This directory contains the Terraform implementation for a Managed Kafka ACL from the Planton spec: one `google_managed_kafka_acl`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the bare cluster id derived from a full path or bare id |
| `main.tf` | `google_managed_kafka_acl` |
| `outputs.tf` | `name`, `resource_type`, `resource_name`, `pattern_type` |

## Send Posture

- **`cluster`** -- the last path segment of the reference or literal.
- **`permission_type`, `host`** -- sent only when set; the provider defaults them to `ALLOW` and `*`.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
