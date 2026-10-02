# GcpGkeFleetMembership — Terraform Implementation

This directory contains the Terraform implementation for an explicit fleet membership from the Planton spec: `google_project_service` for the Fleet API and one `google_gke_hub_membership`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Project, ID and location defaults, the cluster's resource link, attribution labels |
| `main.tf` | `google_project_service`, `google_gke_hub_membership` |
| `outputs.tf` | `name`, `membership_id`, `location` |

## Send Posture

- **`membership_id`** -- defaults to `metadata.name`; **`location`** to `global`.
- **`endpoint.gke_cluster.resource_link`** -- the cluster's ID prefixed with `//container.googleapis.com/` (left as is when already in that form) -- PARITY with the Pulumi module's `gkeClusterResourceLink`.
- **`authority.issuer`** -- sent only when set.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
