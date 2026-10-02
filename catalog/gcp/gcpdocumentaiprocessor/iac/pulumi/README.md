# GcpDocumentAiProcessor — Pulumi Implementation

This directory contains the Pulumi implementation for a Document AI processor from the Planton spec: one `gcp.projects.Service` (API enablement), one `gcp.essentialcontacts.DocumentAiProcessor`, and an optional `gcp.essentialcontacts.DocumentAiProcessorDefaultVersion` (pulumi-gcp files Google's Document AI resources under the `essentialcontacts` package).

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `processor` |
| `module/locals.go` | Stack input holder |
| `module/processor.go` | Enables the API; creates the processor and the default version binding; exports the outputs |
| `module/outputs.go` | Output key constants (`name`, `processor_id`, `location`, `process_endpoint`) |

## Send Posture (parity with Terraform)

- **`DisplayName`** -- `spec.display_name`, defaulting to `metadata.name`.
- **`KmsKeyName`**, **`DeletionPolicy`** -- sent only when set.
- **`Version`** -- `{processor ID}/processorVersions/{default_version}`, only when `default_version` is set.
- **`name` output** -- the resource `ID()` (the full path); `processor_id` is the provider's short `Name`.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
