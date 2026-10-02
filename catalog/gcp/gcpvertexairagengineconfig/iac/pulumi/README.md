# GcpVertexAiRagEngineConfig — Pulumi Implementation

This directory contains the Pulumi implementation for setting the tier of
Vertex AI RAG Engine's managed vector database from the Planton spec: one
`gcp.projects.Service` (API enablement) and one
`gcp.vertex.AiRagEngineConfig`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `rag_engine_config` |
| `module/locals.go` | Carries the stack input (a singleton Google names; no derived name, no labels) |
| `module/rag_engine_config.go` | Enables the API and maps the spec's tier enum to the matching empty block |
| `module/outputs.go` | Output key constants (`name`, `location`) |

## Send Posture (parity with Terraform)

- **`Region`** -- the spec's `location` (the Vertex family's single word for the axis the provider calls `region`).
- **`RagManagedDbConfig`** -- exactly one of `Basic`, `Scaled`, `Unprovisioned`, chosen by the spec's `tier`; the blocks carry no settings.
- **`DeletionPolicy`** -- sent only when set. `DELETE` (the provider default) unprovisions the location on destroy, which deletes the managed database's data.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
