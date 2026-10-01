# GcpComputeImage — Pulumi Implementation

This directory contains the Pulumi implementation for a Compute Engine custom image from the Planton spec: `gcp.projects.Service` and one `gcp.compute.Image`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `image` |
| `module/locals.go` | Stack input, the image name default, attribution labels |
| `module/image.go` | API enablement, the image, the Secure Boot mapping, the outputs |
| `module/outputs.go` | Output key constants (`name`, `self_link`, `family`, `disk_size_gb`) |

## Send Posture (parity with Terraform)

- **`Name`** -- `image_name`, defaulting to `metadata.name`.
- **Optional+Computed** -- `DiskSizeGb`, `GuestOsFeatures`, `Licenses`, `StorageLocations`, and `ShieldedInstanceInitialState` only when declared.
- **`family` output** -- empty when unset, through a nil-safe `ApplyT`.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
