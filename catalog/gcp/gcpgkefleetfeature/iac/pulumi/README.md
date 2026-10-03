# GcpGkeFleetFeature — Pulumi Implementation

This directory contains the Pulumi implementation for a GKE fleet feature from the Planton spec: `gcp.projects.Service` for the Fleet API and the feature's own API, one `gcp.gkehub.Feature`, and one `gkehub.FeatureMembership` per per-cluster entry.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `feature` |
| `module/locals.go` | IaC input, attribution labels, the label merge |
| `module/feature.go` | The feature-to-API table, the feature, the spec and member-default builders, the per-cluster entries, the outputs |
| `module/outputs.go` | Output key constant (`name`) |

## Send Posture (parity with Terraform)

- **Builders** -- one per SDK path, because the SDK gives the fleet default and the per-cluster entry different types; the per-cluster builder sends `ComponentName` and `BundleName`.
- **Optional values** -- `optionalString`, `optionalBool`, `optionalInt`, and `syncWaitSecs` send a value only when the spec sets it.
- **Per-cluster entries** -- logical name `<name>-membership-<location>-<id>`, parsed from the membership's full name.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
