# GCP Vertex AI Search Engine

Stands up a Vertex AI Search app over your data stores: a search engine with Google-quality ranking, generated answers, and an embeddable widget; a chat engine that answers through a Dialogflow CX agent; or a recommendation engine that learns from user events. Tune results with synonyms, boosts, filters, promotions, and redirects, brand the widget, and add a Gemini Enterprise assistant with guardrails -- all in one block, on the Discovery Engine API behind the console's AI Applications.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `discoveryengine.googleapis.com` on the project (and `dialogflow.googleapis.com` for a chat engine)
- **Engine** -- exactly one of `discoveryengine.SearchEngine`, `ChatEngine`, or `RecommendationEngine`, chosen by `engineType`
- **Controls, serving config, widget config, assistants** -- one resource per declared control and assistant; PATCHes of the engine's default serving and widget configs

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Discovery Engine admin permissions on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **GcpVertexAiSearchDataStore** -- at least one store in the engine's location, enrolled in the engine's solution.

### Optional Dependencies

- **GcpVertexAiSearchDataConnector** -- for engines over connector-synced stores, referenced by `collectionId`.
- **GcpKmsKey** -- for customer-managed encryption of a search engine, referenced by `kmsKeyName`.

## Deploy

### Console

Open the deployment store, find **GCP Vertex AI Search Engine**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Search Standard** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiSearchEngine
metadata:
  name: product-search
  org: acme-corp
  env: prod
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
  widgetConfig:
    accessSettings:
      allowPublicAccess: true
      allowlistedDomains:
        - www.example.com
    uiSettings:
      interactionType: SEARCH_WITH_ANSWER
```

```shell
planton apply -f vertex-ai-search-engine.yaml
```

This creates an Enterprise-tier search engine with generated answers and a public widget embeddable on `www.example.com`. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, reference the data stores' `data_store_id` outputs from `dataStoreIds` and their `name` outputs from the engine's controls; reference a connector's `collection_id` from `collectionId` when the engine searches connector-synced stores.

## Key Configuration

These are the most important decisions when configuring an engine. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Which app** -- `engineType` picks search (the default), chat, or recommendation, and is immutable. Each arm has its own block: `searchEngineConfig` and the widget for search, `chatEngineConfig` for the Dialogflow agent, `mediaRecommendationEngineConfig` for the recommendation model.

**How results are shaped** -- `controls` declare synonyms, boosts, filters, promotions, and redirects, each keyed by id and active under its conditions; `servingConfig` says which of them the engine's default serving config applies. Editors tune results by editing the manifest, not the application.

**Who sees the widget** -- `widgetConfig.accessSettings` decides between an authenticated intranet widget (optionally through your own identity provider) and a public one embeddable from allowlisted domains; `uiSettings.interactionType` decides whether the widget generates an answer above the results.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVertexAiSearchDataStore** | `dataStoreIds[]` | `status.outputs.data_store_id` |
| **GcpVertexAiSearchDataStore** | `controls[].*Action.dataStore`, `widgetConfig.uiSettings.dataStoreUiConfigs[].name` | `status.outputs.name` |
| **GcpVertexAiSearchDataConnector** | `collectionId` | `status.outputs.collection_id` |
| **GcpKmsKey** | `kmsKeyName` | `status.outputs.key_id` |
| **GcpModelArmorTemplate** | `assistants[].customerPolicy.modelArmorConfig.userPromptTemplate`, `.responseTemplate` | `status.outputs.name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The engine's full resource name | Application configuration; the search and answer APIs |
| `engine_id` | The engine's id | SDK clients |
| `serving_config_name` | The default serving config's name | Search requests naming a serving config |
| `dialogflow_agent` | A chat engine's Dialogflow CX agent | Dialogflow integrations |
| `control_names`, `assistant_names` | The folded resources' names | Dashboards |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Search standard** -- The smallest search engine over one store. Start from the **Search Standard** preset.

**Enterprise controls and widget** -- Enterprise tier, generated answers, five serving controls, a branded public widget, CMEK. Start from the **Enterprise Controls and Widget** preset.

**Chat agent and assistant** -- A chat engine that creates its Dialogflow CX agent, with a guarded Gemini Enterprise assistant. Start from the **Chat Agent and Assistant** preset.

**Media recommendations** -- A recommended-for-you engine over a media catalog, optimizing watch time. Start from the **Media Recommendations** preset.

## Works With

- [**GCP Vertex AI Search Data Store**](/cloud-catalog/gcp-vertex-ai-search-data-store) -- the stores the engine reads
- [**GCP Vertex AI Search Data Connector**](/cloud-catalog/gcp-vertex-ai-search-data-connector) -- connector-synced stores in their own collection
- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- customer-managed encryption
- [**GCP Model Armor Template**](/cloud-catalog/gcp-model-armor-template) -- the safety templates assistants screen prompts and responses through
