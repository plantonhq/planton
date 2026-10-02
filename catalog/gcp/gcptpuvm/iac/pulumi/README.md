# GcpTpuVm — Pulumi Implementation

This directory contains the Pulumi implementation for a Cloud TPU VM from the Planton spec: one `gcp.projects.Service` (API enablement) and one `gcp.tpu.V2Vm`. pulumi-gcp serves Google's beta-only TPU resource from its single provider; no second provider is involved.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `tpuVm` |
| `module/locals.go` | The `planton-ai_*` attribution labels |
| `module/tpu_vm.go` | Enables the API; maps the accelerator, network interfaces, identity, scheduling, disks, and Secure Boot; exports the outputs |
| `module/outputs.go` | Output key constants (`name`, `node_id`, `zone`) |

## Send Posture (parity with Terraform)

- **`Name`** -- `spec.node_id`, defaulting to `metadata.name`.
- **`AcceleratorType`** -- only when set.
- **`NetworkConfig` vs `NetworkConfigs`** -- one interface or several.
- **Compute references** -- trimmed to relative paths.
- **`ShieldedInstanceConfig`** -- only when `enable_secure_boot` is true.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
