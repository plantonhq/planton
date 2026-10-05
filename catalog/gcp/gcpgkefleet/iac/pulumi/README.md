# GcpGkeFleet — Pulumi Implementation

This directory contains the Pulumi implementation for a project's GKE fleet from the Planton spec: `gcp.projects.Service` for the Fleet API and one `gcp.gkehub.Fleet`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `fleet` |
| `module/locals.go` | IaC input holder |
| `module/fleet.go` | Project resolution, API enablement, the fleet, the outputs |
| `module/outputs.go` | Output key constants (`project_id`, `name`, `uid`) |

## Send Posture (parity with Terraform)

- **`Project`** -- the spec's project, or the provider's project from `organizations.GetClientConfig`.
- **`DisplayName`**, **`DeletionPolicy`**, and every cluster-default enum -- sent only when set.
- **Not sent** -- labels and the compliance posture default: `FleetArgs` and `FleetDefaultClusterConfigArgs` carry neither at pulumi-gcp v9.37.0 (the trigger to model them is pulumi-gcp v10 GA).

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
