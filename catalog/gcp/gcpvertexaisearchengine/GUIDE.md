# GcpVertexAiSearchEngine Guide

This guide explains how a Vertex AI Search engine is shaped and the decisions behind the spec. For field-by-field detail see the README; for ready-made configurations see the presets.

The product has worn three names. The API is **Discovery Engine** (`discoveryengine.googleapis.com`), the search product is **Vertex AI Search**, and the console now groups it under **AI Applications**, with **Gemini Enterprise** as the assistant built on the same engines. This kind is the engine -- the "app" -- in all three vocabularies.

## One kind, three engines

Google's API has one `Engine` resource with a solution type; its provider splits it into three resources -- a search engine, a chat engine, a recommendation engine -- because each carries a different configuration block. This kind keeps the API's shape: `engineType` selects which resource the modules build (empty means `SEARCH`), the shared fields (`engineId`, `displayName`, `location`, `collectionId`, `dataStoreIds`, `industryVertical`, `commonConfig`) apply to all three, and each arm's own block sits at the spec's top level under the provider's own name:

- **SEARCH** -- `searchEngineConfig` (tier, add-ons, license tier; Google requires the block, so the module always sends it), plus `appType`, `disableAnalytics`, `features`, `kmsKeyName`, and `knowledgeGraphConfig`, which exist only on this resource.
- **CHAT** -- `chatEngineConfig`: exactly one of "create a Dialogflow CX agent" (`agentCreationConfig`) or "link one" (`dialogflowAgentToLink`), immutable.
- **RECOMMENDATION** -- `mediaRecommendationEngineConfig`: the model type, objective, training state, and feature config. This resource has no collection id and no KMS key, and takes exactly one data store; the spec walls all three.

Rules the spec enforces are the provider's: the chat block is required on CHAT and forbidden elsewhere, the search-only fields exist only on SEARCH, a chat engine is GENERIC, a recommendation engine is GENERIC or MEDIA. `engineType` is immutable; changing it replaces the engine.

## Data stores by reference

`dataStoreIds` lists `GcpVertexAiSearchDataStore` references (their `data_store_id` outputs). Every store must be in the engine's location and enrolled in the engine's solution -- `SOLUTION_TYPE_CHAT` for a chat engine, and so on. A search engine over several stores is a blended search; a recommendation engine takes one.

`collectionId` is `default_collection` -- where every `GcpVertexAiSearchDataStore` lives -- unless the engine searches stores a `GcpVertexAiSearchDataConnector` created, in which case it references the connector's `collection_id` output.

## Controls, and the serving config that applies them

A control is one rule that shapes results: `synonymsAction` treats terms as equivalent, `boostAction` reorders documents matching a filter (by a fixed amount or a curve over a numeric or freshness field), `filterAction` drops non-matching documents, `promoteAction` pins a link, `redirectAction` sends a query elsewhere. Each has exactly one action, an optional set of conditions (query terms, a regex, time windows), and an id.

Controls belong to one engine and nothing else references them, so they are folded as `controls[]`. A control takes effect only when the engine's **serving config** lists it: `servingConfig` carries five id lists, one per action, and the spec checks that every id names a declared control with that action -- the mistake of listing a boost control under filters is caught before Google sees it. Google creates the default serving config (`default_search`) with the engine; the provider PATCHes it and never deletes it, so `servingConfig` configures rather than owns. `solutionType` on every control is derived from the arm, because the provider hard-codes each engine's.

## The widget

`widgetConfig` configures Google's embeddable search widget and hosted web app: who may load it (`accessSettings` -- public from allowlisted domains, or authenticated, optionally through a workforce identity pool provider), the homepage shortcuts, the logo, and the UI (`interactionType` decides whether a generated answer or a conversation sits above the results; `dataStoreUiConfigs` map each store's fields to facets and result components; `generativeAnswerConfig` tunes the answer). Google creates the default widget config with a search engine; this block PATCHes it, and a widget config cannot be deleted, so destroy leaves it as configured.

## Assistants

`assistants[]` are Gemini Enterprise assistants on the engine: a content policy (banned phrases), Model Armor sanitization of prompts and responses (templates by resource name, fail open or closed), web grounding, and generation settings. They need Gemini Enterprise seats licensed on the project.

## Encryption, the chat agent, and destroy

`kmsKeyName` encrypts a search engine under a Cloud KMS key in its location (the chat and recommendation resources offer none). A chat engine that creates its agent does so through Google's service agent; the module enables the Dialogflow API so a fresh project works. `deletionPolicy` is fanned to the controls and assistants: `DELETE` (default), `PREVENT`, or `ABANDON`. Destroy the engine before its data stores -- Google refuses to delete a store an engine still uses -- and expect Google to reserve a deleted engine's id for a while.
