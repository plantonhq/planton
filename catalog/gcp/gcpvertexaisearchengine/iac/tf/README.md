# GcpVertexAiSearchEngine — Terraform Implementation

This directory contains the Terraform implementation for a Vertex AI Search
engine and its folded controls, serving config, widget config, and
assistants from the Planton spec: `google_project_service` for the
Discovery Engine API (and, `count`-gated on the chat arm, the Dialogflow
API); exactly one of `google_discovery_engine_search_engine`,
`google_discovery_engine_chat_engine`, and
`google_discovery_engine_recommendation_engine` (`count`-gated by
`spec.engine_type`); `google_discovery_engine_control` per `spec.controls[]`
entry (`for_each` keyed by `control_id`); at most one
`google_discovery_engine_serving_config` and one
`google_discovery_engine_widget_config` (`count`); and
`google_discovery_engine_assistant` per `spec.assistants[]` entry
(`for_each` keyed by `assistant_id`).

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials
  or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the arm flags and derived solution type, the engine id / display name / collection defaults, null-for-empty optionals, `created_engine_id` through `one(concat(...))`, the control and assistant maps |
| `main.tf` | The API enablements, the three engine resources, `google_discovery_engine_control`, `google_discovery_engine_serving_config`, `google_discovery_engine_widget_config`, `google_discovery_engine_assistant` |
| `outputs.tf` | `name`, `engine_id`, `location`, `collection_id`, `engine_type`, `serving_config_name`, `widget_config_name`, `dialogflow_agent`, `control_names`, `assistant_names` |

## Send Posture

- **Arm selection** -- `local.engine_type` (`spec.engine_type`, or SEARCH when empty) gates the three engine resources with `count`; engine-level outputs read the one that exists through `one(concat(...))` -- PARITY with the Pulumi module.
- **`search_engine_config`** -- always present on the search arm (Google requires the block); its attributes `null` when the spec omits them.
- **Optional+Computed** (`features`, `knowledge_graph_config` and its bools) -- `null` / absent when unset so Google's defaults stay in charge.
- **Companions** -- `engine_id = local.created_engine_id`, `collection_id = local.collection_id`, `location` from the spec; controls carry `local.solution_type`; the serving config `depends_on` the controls; `deletion_policy` fanned to controls and assistants.
- **`dialogflow_agent`** -- from the chat engine's `chat_engine_metadata[0]`, an empty string on the other arms.
- **List outputs** -- in manifest order (comprehensions over the spec lists).
- **No labels** -- Discovery Engine resources carry none.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
