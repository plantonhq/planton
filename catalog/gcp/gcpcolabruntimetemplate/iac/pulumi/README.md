# GcpColabRuntimeTemplate — Pulumi Implementation

This directory contains the Pulumi implementation for a Colab Enterprise runtime template from the Planton spec: one `gcp.projects.Service` (API enablement) and one `gcp.colab.RuntimeTemplate`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `runtimeTemplate` |
| `module/locals.go` | The `planton-ai_*` attribution labels |
| `module/runtime_template.go` | Enables the API; maps the machine, disk, network, lifted wrappers, and software config; exports the outputs |
| `module/outputs.go` | Output key constants (`name`, `runtime_template_id`, `location`) |

## Send Posture (parity with Terraform)

- **`Name`** / **`DisplayName`** -- each defaulting to `metadata.name`.
- **Optional+Computed settings** -- sent only when set; `DiskSizeGb` as a decimal string.
- **Lifted wrappers** -- `IdleShutdownConfig`, `EucConfig`, `ShieldedVmConfig`, `EncryptionSpec` only when their spec field is set.
- **`Subnetwork`** -- a `GcpSubnetwork` self-link trimmed to its relative path.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
