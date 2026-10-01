# GcpGkeFleetFeature — Terraform Implementation

This directory contains the Terraform implementation for a GKE fleet feature from the Planton spec: `google_project_service` for the Fleet API and the feature's own API, one `google_gke_hub_feature`, and one `google_gke_hub_feature_membership` per per-cluster entry.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Project, location default, the feature-to-API table, attribution labels, per-cluster entries keyed by membership name |
| `main.tf` | The two API enablements, the feature with its spec and member defaults, the per-cluster entries |
| `outputs.tf` | `name` |

## Send Posture

- **`spec`** -- rendered only when a feature settings block is declared; the spec allows only the block of `feature`.
- **Policy Controller** -- one spec message rendered under each resource's names: `component` / `pod_toleration` / `bundle` on the feature, `component_name` / `pod_tolerations` / `bundle_name` on the per-cluster entry.
- **Config Sync** -- `sync_wait_secs` sent as the decimal string the provider takes, only when positive; presence-tracked bools only when set.
- **Optional+Computed** -- `clusterupgrade.post_conditions`, Policy Controller's audit interval, replica counts, affinity, monitoring, and template library sent only when declared.
- **Per-cluster entries** -- the feature's project; `membership` and `membership_location` split from the membership's full name -- PARITY with the Pulumi module's `parseMembershipName`.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
