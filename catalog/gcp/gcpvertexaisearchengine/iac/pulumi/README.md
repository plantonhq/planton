# GcpVertexAiSearchEngine — Pulumi Implementation

This directory contains the Pulumi implementation for a Vertex AI Search
engine and its folded controls, serving config, widget config, and
assistants from the Planton spec: one or two `gcp.projects.Service` (API
enablement; the Dialogflow API on the chat arm), exactly one of
`gcp.discoveryengine.SearchEngine`, `ChatEngine`, or `RecommendationEngine`
(chosen by `spec.engine_type`), one `gcp.discoveryengine.Control` per
`spec.controls[]` entry, at most one `gcp.discoveryengine.ServingConfig`
and one `gcp.discoveryengine.WidgetConfig`, and one
`gcp.discoveryengine.Assistant` per `spec.assistants[]` entry.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, the engine, and its companions, then exports the outputs |
| `module/locals.go` | The resolved arm and its solution type, the engine id and display name defaulted from `metadata.name`, the collection id defaulted to `default_collection` |
| `module/engine.go` | Enables the APIs and dispatches to the arm; `createdEngine` is what the companions address |
| `module/search_engine.go` | The SEARCH arm: `SearchEngineConfig` always sent; `AppType`, `Features`, `KmsKeyName`, `KnowledgeGraphConfig` only when set |
| `module/chat_engine.go` | The CHAT arm: agent creation XOR link; the created agent read back from `ChatEngineMetadatas` |
| `module/recommendation_engine.go` | The RECOMMENDATION arm: the media config's levers only when set |
| `module/controls.go` | One control per entry with the derived `SolutionType`; conditions and the five actions |
| `module/serving_config.go` | The PATCH of `default_search`, depending on every control it lists |
| `module/widget_config.go` | The PATCH of `default_search_widget_config` |
| `module/assistants.go` | One assistant per entry |
| `module/outputs.go` | Output key constants (`name`, `engine_id`, `location`, `collection_id`, `engine_type`, `serving_config_name`, `widget_config_name`, `dialogflow_agent`, `control_names`, `assistant_names`) |

## Send Posture (parity with Terraform)

- **Arm selection** -- `spec.engine_type` empty resolves to SEARCH; exactly one engine resource is built.
- **`SearchEngineConfig`** -- always sent on the search arm, empty when the spec omits it (Google requires the block); its levers only when set.
- **Optional+Computed** (`Features`, `KnowledgeGraphConfig` and its bools) -- sent only when set so Google's defaults stay in charge and re-plan clean.
- **Companions** -- `EngineId` from the created engine, `CollectionId` from the resolved collection, `Location` from the spec; controls carry the arm's solution type; the serving config depends on the control resources, not only the engine; `DeletionPolicy` fanned to controls and assistants (the serving and widget configs have none).
- **`DialogflowAgent`** -- read from the chat engine's metadata list; an empty string on the other arms.
- **No labels** -- Discovery Engine resources carry none.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
