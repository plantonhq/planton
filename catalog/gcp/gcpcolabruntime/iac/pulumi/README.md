# GcpColabRuntime — Pulumi Implementation

This directory contains the Pulumi implementation for a Colab Enterprise runtime from the Planton spec: one `gcp.projects.Service` (API enablement) and one `gcp.colab.Runtime`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `runtime` |
| `module/locals.go` | Stack input holder |
| `module/runtime.go` | Enables the API; maps the runtime; exports the outputs |
| `module/outputs.go` | Output key constants (`name`, `runtime_id`, `location`) |

## Send Posture (parity with Terraform)

- **`Name`** -- `spec.runtime_id`, defaulting to `metadata.name`, always sent.
- **`NotebookRuntimeTemplateRef`** -- always set from `spec.runtime_template`.
- **`DesiredState`** -- sent only when set; **`AutoUpgrade`** as declared.
- **`Description`**, **`DeletionPolicy`** -- sent only when set.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
