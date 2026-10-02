# GcpVertexAiPersistentResource — Terraform Implementation

This directory contains the Terraform implementation for a Vertex AI persistent resource from the Planton spec: one `google_project_service` (API enablement), one guarded `data.google_project` lookup, and one `google_vertex_ai_persistent_resource`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the id defaulted from `metadata.name`, null-for-empty optionals, the peered network normalized to Google's project-number form, the `planton-ai_*` labels |
| `main.tf` | `google_project_service`, `data.google_project` (only when the network path carries a project ID), `google_vertex_ai_persistent_resource` |
| `outputs.tf` | `name`, `persistent_resource_id`, `location`, `state` |

## Send Posture

- **`name`** -- `spec.persistent_resource_id`, defaulting to `metadata.name` -- PARITY with the Pulumi module.
- **`network`** -- the self-link prefix stripped; when the project segment is not numeric, the number is read from `data.google_project` (one read, skipped for numeric paths and when no network is set).
- **Pools** -- `id` sent only when set (Optional+Computed); `replica_count` and the autoscaling bounds sent as decimal strings; machine fields sent only when set; `disk_spec` emitted only when the spec shapes it.
- **`resource_runtime_spec.service_account_spec`** -- emitted only when `enable_custom_service_account` is true (the spec lifts the two one-leaf wrappers).
- **`encryption_spec`**, **`psc_interface_config`**, **`reserved_ip_ranges`** -- emitted only when set.
- **`name` output** -- the resource's `id` (the provider's `name` attribute is the short id).

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
