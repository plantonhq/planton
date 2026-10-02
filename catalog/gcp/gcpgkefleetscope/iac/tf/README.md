# GcpGkeFleetScope — Terraform Implementation

This directory contains the Terraform implementation for a fleet team scope from the Planton spec: `google_project_service` for the Fleet API, one `google_gke_hub_scope`, and the folded `google_gke_hub_namespace`, `google_gke_hub_scope_rbac_role_binding`, and `google_gke_hub_membership_binding` resources.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Project, scope ID default, attribution labels, children keyed by their IDs, membership names parsed |
| `main.tf` | `google_project_service`, the scope, namespaces, role bindings, cluster bindings |
| `outputs.tf` | `name`, `scope_id`, `uid` |

## Send Posture

- **Children** -- keyed by their own declared IDs, so editing one never renames or recreates another.
- **Namespace** -- the scope's short ID (URL) and full name (body), both from the created scope.
- **Cluster binding** -- the scope's project; `location` and `membership_id` split from the membership's full name -- PARITY with the Pulumi module's `parseMembershipName`.
- **Labels** -- attribution labels merged into every resource's `labels`, never into Kubernetes `namespace_labels`; empty `namespace_labels` are not sent.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
