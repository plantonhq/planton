# GcpModelArmorFloorSetting — Pulumi Implementation

This directory contains the Pulumi implementation for a Model Armor floor setting from the Planton spec: an optional `gcp.projects.Service` (API enablement, project floors) and one `gcp.modelarmor.Floorsetting`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `floorSetting` |
| `module/locals.go` | Stack input holder |
| `module/floor_setting.go` | Scope selection, API enablement for project floors, the filter configuration and service settings, the outputs |
| `module/outputs.go` | Output key constants (`name`, `parent`) |

## Send Posture (parity with Terraform)

- **`Parent`** -- from the one scope arm set; an empty scope uses the provider's project from `organizations.GetClientConfig`.
- **`Location`** -- `spec.location`, defaulting to `global`.
- **`InspectOnly` / `InspectAndBlock`** -- only the flag `enforcement_type` names.
- **`FloorSettingMetadata`** -- only when `enable_multi_language_detection` is true.
- **Destroy** -- the provider removes the floor from state only; Google keeps it.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
