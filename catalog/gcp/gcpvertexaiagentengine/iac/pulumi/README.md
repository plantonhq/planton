# GcpVertexAiAgentEngine — Pulumi Implementation

This directory contains the Pulumi implementation for a Vertex AI Agent
Engine instance from the Planton spec: one `gcp.projects.Service` (API
enablement) and one `gcp.vertex.AiReasoningEngine`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `agent_engine` |
| `module/locals.go` | The display name defaulted from `metadata.name`; the `planton-ai_*` attribution labels |
| `module/agent_engine.go` | Enables the API; maps the top level, the agent `spec` (source, container, package, build, deployment); exports the outputs |
| `module/memory_bank.go` | Maps `context_spec.memory_bank_config`: generation, lookup, TTLs, structured schemas, per-scope customization with worked examples in Gemini's `Content.Part` shape |
| `module/outputs.go` | Output key constants (`name`, `reasoning_engine_id`, `location`, `create_time`, `update_time`) |

## Send Posture (parity with Terraform)

- **`Region`** -- the spec's `location` (the Vertex family's single word for the axis the provider calls `region`).
- **`DisplayName`** -- `spec.display_name`, defaulting to `metadata.name`.
- **Optional strings, booleans, maps** -- sent only when set; Optional+Computed instance and concurrency numbers only when set.
- **`SourceCodeSpec`** -- exactly one source and one build (proto-enforced), each emitted when set.
- **`BuildSpec`** -- `WorkerPool` only; `service_account` is an SDK gap held out on both engines.
- **Memory-bank example parts** -- every payload but `audio_transcription` (an SDK gap held out on both engines).
- **`name`** -- the resource id (the full path); `reasoning_engine_id` the provider's numeric `name`.
- **`DeletionPolicy`** -- sent only when set.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
