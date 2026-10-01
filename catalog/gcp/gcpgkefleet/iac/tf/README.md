# GcpGkeFleet — Terraform Implementation

This directory contains the Terraform implementation for a project's GKE fleet from the Planton spec: `google_project_service` for the Fleet API and one `google_gke_hub_fleet`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | The project (provider project when empty), optional strings as null |
| `main.tf` | `data.google_client_config` (empty project only), `google_project_service`, `google_gke_hub_fleet` |
| `outputs.tf` | `project_id`, `name`, `uid` |

## Send Posture

- **`project`** -- the spec's project, or the provider's project from `google_client_config` (provider configuration, no API call), so the `project_id` output is always concrete -- PARITY with the Pulumi module's `GetClientConfig`.
- **`display_name`**, **`deletion_policy`**, and every cluster-default enum -- sent only when set.
- **`default_cluster_config`** -- rendered only when a posture or Binary Authorization block is declared; each policy binding name becomes one `policy_bindings` block.
- **Not sent** -- `labels` and `default_cluster_config.compliance_posture_config` (SDK gaps at pulumi-gcp v9.37.0; both engines send the same arguments).

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
