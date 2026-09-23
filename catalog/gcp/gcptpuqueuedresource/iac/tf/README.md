# GcpTpuQueuedResource — Terraform Implementation

This directory contains the Terraform implementation for a Cloud TPU queued resource from the Planton spec: one `google_project_service` (API enablement, GA provider) and one `google_tpu_v2_queued_resource` on `google-beta`, attached under the recorded admission in `pkg/providerparity/admissions/google-beta.yaml`.

## Provider

- **Providers**: `hashicorp/google` and `hashicorp/google-beta`, both `~> 8.3` (one release line)
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block with both providers and why the beta channel is attached |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the request id defaulted from `metadata.name`, the node parent |
| `main.tf` | `data.google_client_config` (no project in the spec only), `google_project_service`, `google_tpu_v2_queued_resource` (`provider = google-beta`) |
| `outputs.tf` | `name`, `queued_resource_id`, `zone` |

## Send Posture

- **`name`** -- `spec.queued_resource_id`, defaulting to `metadata.name` -- PARITY with the Pulumi module.
- **`tpu.node_spec[].parent`** -- derived as `projects/{project}/locations/{zone}`; the project is the spec's, or the provider's from `google_client_config` (provider configuration, no API call).
- **Node optionals** -- `node_id`, `accelerator_type`, `description`, `queue_count` sent only when set; the subnetwork self-link trimmed to its relative path.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
