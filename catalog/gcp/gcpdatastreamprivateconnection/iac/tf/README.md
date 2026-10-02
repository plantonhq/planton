# GcpDatastreamPrivateConnection — Terraform Implementation

This directory contains the Terraform implementation for a Datastream private connection from the Planton spec: one `google_project_service` (API enablement) and one `google_datastream_private_connection`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the id and display name defaulted from `metadata.name`, null-for-empty optionals, the merged labels |
| `main.tf` | `google_project_service`, `google_datastream_private_connection` with one dynamic connectivity block |
| `outputs.tf` | `name`, `private_connection_id` |

## Send Posture

- **`private_connection_id`, `display_name`** -- defaulting to `metadata.name`.
- **Connectivity** -- exactly one of `vpc_peering_config` or `psc_interface_config` is emitted.
- **Labels** -- the spec's labels merged under the platform attribution labels.
- **`deletion_policy`** -- sent only when set, leaving the provider's `FORCE` default otherwise.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
