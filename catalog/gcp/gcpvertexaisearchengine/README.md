# GCP Vertex AI Search Engine

A Vertex AI Search engine -- the app over one or more data stores -- on the Discovery Engine API (the console calls the product AI Applications / Gemini Enterprise). One kind builds one of three engines, chosen by `engineType`: a **search** engine (tiers, add-ons, a knowledge graph, the embeddable widget), a **chat** engine (creates or links a Dialogflow CX agent), or a **recommendation** engine (a generic or media model). Folded in, because each lives under exactly one engine: the serving controls and the serving config that applies them, the search widget's configuration, and the engine's Gemini Enterprise assistants.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `discoveryengine.googleapis.com` on the project, and `dialogflow.googleapis.com` on a CHAT engine (never disabled on destroy)
- **Engine** -- exactly one of `discovery_engine_search_engine`, `discovery_engine_chat_engine`, or `discovery_engine_recommendation_engine`
- **Controls** -- one `discovery_engine_control` per `controls[]` entry, keyed by `controlId`
- **Serving config** -- a PATCH of the engine's default serving config (`discovery_engine_serving_config`) when `servingConfig` is declared
- **Widget config** -- a PATCH of the engine's default widget config (`discovery_engine_widget_config`) when `widgetConfig` is declared
- **Assistants** -- one `discovery_engine_assistant` per `assistants[]` entry, keyed by `assistantId`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Discovery Engine admin permissions on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpVertexAiSearchDataStore`** -- at least one store in the engine's location, enrolled in the engine's solution (`dataStoreIds`).

### Optional Dependencies

- **`GcpVertexAiSearchDataConnector`** -- the collection a connector created, when the engine searches connector-synced stores (`collectionId`).
- **`GcpKmsKey`** -- customer-managed encryption for a search engine (`kmsKeyName`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiSearchEngine
metadata:
  name: product-search
spec:
  location: global
  dataStoreIds:
    - valueFrom:
        kind: GcpVertexAiSearchDataStore
        name: product-docs
        fieldPath: status.outputs.data_store_id
  searchEngineConfig:
    searchTier: SEARCH_TIER_ENTERPRISE
    searchAddOns:
      - SEARCH_ADD_ON_LLM
  controls:
    - controlId: synonyms-laptop
      displayName: Laptop synonyms
      synonymsAction:
        synonyms: [laptop, notebook]
  servingConfig:
    synonymsControlIds: [synonyms-laptop]
```

```shell
planton apply -f vertex-ai-search-engine.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | `global`, `us`, or `eu` -- the same as the data stores'. Immutable. |
| `dataStoreIds` | `[]StringValueOrRef` | The stores the engine reads (`GcpVertexAiSearchDataStore` references or literal ids); a recommendation engine takes exactly one. Mutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `engineType` | `string` | `SEARCH` | `SEARCH`, `CHAT`, or `RECOMMENDATION`. Immutable. |
| `engineId` | `string` | `metadata.name` | RFC 1034 id. Immutable. |
| `displayName` | `string` | `metadata.name` | Console name; mutable. |
| `collectionId` | `StringValueOrRef` | `default_collection` | A `GcpVertexAiSearchDataConnector` reference for connector-built collections. Immutable; always default for RECOMMENDATION. |
| `industryVertical` | `string` | `GENERIC` | `GENERIC`, `MEDIA`, `HEALTHCARE_FHIR` (CHAT is GENERIC only; RECOMMENDATION is GENERIC or MEDIA). Immutable. |
| `commonConfig.companyName` | `string` | none | The company the engine represents. Immutable. |
| `searchEngineConfig` | `object` | Standard tier | SEARCH only: `searchTier`, `searchAddOns` (`SEARCH_ADD_ON_LLM`), `requiredSubscriptionTier`. Always sent (Google requires the block). |
| `appType`, `disableAnalytics`, `features`, `kmsKeyName`, `knowledgeGraphConfig` | | | SEARCH only: `APP_TYPE_INTRANET`, analytics off, feature opt-ins (`FEATURE_STATE_ON` / `OFF`), CMEK, the Cloud and private knowledge graphs. |
| `chatEngineConfig` | `object` | | CHAT only, required there: exactly one of `agentCreationConfig { business, defaultLanguageCode, timeZone, location }` or `dialogflowAgentToLink` (a `GcpDialogflowCxAgent` reference or a literal agent name); `allowCrossRegion`. Immutable. |
| `mediaRecommendationEngineConfig` | `object` | | RECOMMENDATION only: `type`, `optimizationObjective`, `optimizationObjectiveConfig`, `trainingState`, `engineFeaturesConfig`. |
| `controls[]` | `[]object` | none | `controlId`, `displayName`, `useCases`, `conditions[]`, and exactly one of `boostAction`, `filterAction`, `promoteAction`, `redirectAction`, `synonymsAction`. |
| `servingConfig` | `object` | none | `boostControlIds`, `filterControlIds`, `promoteControlIds`, `redirectControlIds`, `synonymsControlIds` -- each id a declared control with that action. |
| `widgetConfig` | `object` | none | `accessSettings`, `homepageSetting`, `uiBranding`, `uiSettings` (interaction type, result descriptions, per-store facets and components, the generated answer). |
| `assistants[]` | `[]object` | none | `assistantId`, `displayName`, `webGroundingType`, `customerPolicy { bannedPhrases, modelArmorConfig }`, `generationConfig`. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`; fanned to the controls and assistants. |

### Validation Rules

- `chatEngineConfig` is required on CHAT and forbidden elsewhere; `mediaRecommendationEngineConfig` is RECOMMENDATION only; the search-only fields exist only on SEARCH.
- A RECOMMENDATION engine lives in `default_collection`, has no KMS key, and takes at most one data store.
- A control has exactly one action; a boost is exactly one of `fixedBoost` or `interpolationBoostSpec`; control and assistant ids are unique.
- Every id in `servingConfig` names a declared control with the matching action.
- Tiers, subscription tiers, verticals, use cases, interaction types, and the other value lists are Google's.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/collections/{collection}/engines/{id}` |
| `engine_id` | `string` | The engine's id |
| `location` | `string` | The engine's location |
| `collection_id` | `string` | The collection the engine lives in |
| `engine_type` | `string` | `SEARCH`, `CHAT`, or `RECOMMENDATION` |
| `serving_config_name` | `string` | The default serving config's full name when configured |
| `widget_config_name` | `string` | The widget config's full name when configured |
| `dialogflow_agent` | `string` | The Dialogflow CX agent a CHAT engine answers through |
| `control_names` | `[]string` | The controls' full names, in manifest order |
| `assistant_names` | `[]string` | The assistants' full names, in manifest order |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **`engineType` is immutable.** Changing it replaces the engine; so do `engineId`, `location`, `collectionId`, `industryVertical`, `commonConfig`, `appType`, and the whole `chatEngineConfig`.
- **The serving and widget configs are Google's own.** Both exist with the engine; this block PATCHes them, and destroy leaves the widget config as configured (Google never deletes it).
- **A chat engine creates a Dialogflow CX agent** in your project (unless it links one); the module enables the Dialogflow API. Whether the created agent outlives the engine at destroy is Google's behavior, not the module's.
- **Assistants are a Gemini Enterprise feature** and need licensed seats on the project.
- **Queries bill per thousand** at the tier's rate; generated answers add the LLM add-on rate. Creating the engine is free.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpVertexAiSearchDataStore** -- the stores the engine reads
- **GcpVertexAiSearchDataConnector** -- a connector-built collection of stores
- **GcpKmsKey** -- customer-managed encryption for a search engine
- **GcpModelArmorTemplate** -- the templates an assistant screens prompts and responses through

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
