# GcpVertexAiRagEngineConfig — Terraform Implementation

This directory contains the Terraform implementation for setting the tier
of Vertex AI RAG Engine's managed vector database from the Planton spec:
one `google_project_service` (API enablement) and one
`google_vertex_ai_rag_engine_config`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials
  or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback and the tier selector lists |
| `main.tf` | `google_project_service`, `google_vertex_ai_rag_engine_config` |
| `outputs.tf` | `name`, `location` |

## Send Posture

- **`region`** -- the spec's `location` -- PARITY with the Pulumi module.
- **`rag_managed_db_config`** -- exactly one of the `basic` / `scaled` / `unprovisioned` empty blocks, emitted by a `dynamic` block from the spec's `tier`.
- **`deletion_policy`** -- sent only when set. `DELETE` (the provider default) PATCHes the location to `UNPROVISIONED` on destroy, which deletes the managed database's data.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
