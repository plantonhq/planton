# GcpVertexAiSearchEngine

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpVertexAiSearchEngineSpec defines a Vertex AI Search engine -- the app
over one or more data stores, on the Discovery Engine API behind the
console's AI Applications / Gemini Enterprise. One kind builds one of
three provider resources, chosen by engine_type:

  SEARCH (default)  `google_discovery_engine_search_engine`: a search
                    app with tiers, add-ons, a knowledge graph, and the
                    embeddable widget
  CHAT              `google_discovery_engine_chat_engine`: a chat app that
                    creates or links a Dialogflow CX agent
  RECOMMENDATION    `google_discovery_engine_recommendation_engine`: a
                    generic or media recommendation model

Folded in, because each lives under exactly one engine and nothing else
references it: the serving controls and the serving config that applies
them, the search widget's configuration, and the engine's assistants.
Every data store an engine uses must be in the engine's location and
enrolled in the engine's solution.

Immutable: engine_type, engine_id, location, collection_id,
industry_vertical, common_config, app_type, chat_engine_config.
Everything else updates in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiSearchEngine
metadata:
  name: product-search
spec:
  projectId:
    value: my-gcp-project
  # The same location as the data stores it searches.
  location: global
  # Empty means SEARCH; CHAT and RECOMMENDATION build the other two engines.
  engineType: SEARCH
  displayName: Product search
  dataStoreIds:
    - valueFrom:
        kind: GcpVertexAiSearchDataStore
        name: product-docs
        fieldPath: status.outputs.data_store_id
  industryVertical: GENERIC
  commonConfig:
    companyName: Acme
  # Enterprise tier with the LLM add-on: generative answers and summaries.
  searchEngineConfig:
    searchTier: SEARCH_TIER_ENTERPRISE
    searchAddOns:
      - SEARCH_ADD_ON_LLM
  # Serving controls, applied through the default serving config below.
  controls:
    - controlId: synonyms-laptop
      displayName: Laptop synonyms
      useCases:
        - SEARCH_USE_CASE_SEARCH
      synonymsAction:
        synonyms:
          - laptop
          - notebook
    - controlId: boost-docs
      displayName: Boost documentation
      useCases:
        - SEARCH_USE_CASE_SEARCH
      boostAction:
        dataStore:
          valueFrom:
            kind: GcpVertexAiSearchDataStore
            name: product-docs
            fieldPath: status.outputs.name
        filter: '(category: ANY("docs"))'
        fixedBoost: 0.5
  servingConfig:
    synonymsControlIds:
      - synonyms-laptop
    boostControlIds:
      - boost-docs
  # The embeddable widget: public, from the company's site, with a
  # generated answer above the results.
  widgetConfig:
    accessSettings:
      allowPublicAccess: true
      allowlistedDomains:
        - www.example.com
    uiSettings:
      interactionType: SEARCH_WITH_ANSWER
      resultDescriptionType: SNIPPET
      enableAutocomplete: true
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.engineType` | `string` |  |  |  |
| `spec.engineId` | `string` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.collectionId` | `string \| valueFrom` |  |  | GcpVertexAiSearchDataConnector (`status.outputs.collection_id`) |
| `spec.dataStoreIds` | `[]string \| valueFrom` | yes |  | GcpVertexAiSearchDataStore (`status.outputs.data_store_id`) |
| `spec.industryVertical` | `string` |  |  |  |
| `spec.commonConfig` | `GcpVertexAiSearchEngineCommonConfig` |  |  |  |
| `spec.commonConfig.companyName` | `string` |  |  |  |
| `spec.searchEngineConfig` | `GcpVertexAiSearchEngineSearchEngineConfig` |  |  |  |
| `spec.searchEngineConfig.searchTier` | `string` |  |  |  |
| `spec.searchEngineConfig.searchAddOns` | `[]string` |  |  |  |
| `spec.searchEngineConfig.requiredSubscriptionTier` | `string` |  |  |  |
| `spec.appType` | `string` |  |  |  |
| `spec.disableAnalytics` | `bool` |  |  |  |
| `spec.features` | `map<string, string>` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.knowledgeGraphConfig` | `GcpVertexAiSearchEngineKnowledgeGraphConfig` |  |  |  |
| `spec.knowledgeGraphConfig.enableCloudKnowledgeGraph` | `bool` |  |  |  |
| `spec.knowledgeGraphConfig.enablePrivateKnowledgeGraph` | `bool` |  |  |  |
| `spec.knowledgeGraphConfig.cloudKnowledgeGraphTypes` | `[]string` |  |  |  |
| `spec.knowledgeGraphConfig.featureConfig` | `GcpVertexAiSearchEngineKnowledgeGraphFeatureConfig` |  |  |  |
| `spec.knowledgeGraphConfig.featureConfig.disablePrivateKgAutoComplete` | `bool` |  |  |  |
| `spec.knowledgeGraphConfig.featureConfig.disablePrivateKgEnrichment` | `bool` |  |  |  |
| `spec.knowledgeGraphConfig.featureConfig.disablePrivateKgQueryUiChips` | `bool` |  |  |  |
| `spec.knowledgeGraphConfig.featureConfig.disablePrivateKgQueryUnderstanding` | `bool` |  |  |  |
| `spec.chatEngineConfig` | `GcpVertexAiSearchEngineChatEngineConfig` |  |  |  |
| `spec.chatEngineConfig.agentCreationConfig` | `GcpVertexAiSearchEngineAgentCreationConfig` |  |  |  |
| `spec.chatEngineConfig.agentCreationConfig.business` | `string` |  |  |  |
| `spec.chatEngineConfig.agentCreationConfig.defaultLanguageCode` | `string` | yes |  |  |
| `spec.chatEngineConfig.agentCreationConfig.timeZone` | `string` | yes |  |  |
| `spec.chatEngineConfig.agentCreationConfig.location` | `string` |  |  |  |
| `spec.chatEngineConfig.dialogflowAgentToLink` | `string` |  |  |  |
| `spec.chatEngineConfig.allowCrossRegion` | `bool` |  |  |  |
| `spec.mediaRecommendationEngineConfig` | `GcpVertexAiSearchEngineMediaRecommendationEngineConfig` |  |  |  |
| `spec.mediaRecommendationEngineConfig.type` | `string` |  |  |  |
| `spec.mediaRecommendationEngineConfig.optimizationObjective` | `string` |  |  |  |
| `spec.mediaRecommendationEngineConfig.optimizationObjectiveConfig` | `GcpVertexAiSearchEngineOptimizationObjectiveConfig` |  |  |  |
| `spec.mediaRecommendationEngineConfig.optimizationObjectiveConfig.targetField` | `string` | yes |  |  |
| `spec.mediaRecommendationEngineConfig.optimizationObjectiveConfig.targetFieldValueFloat` | `float` |  |  |  |
| `spec.mediaRecommendationEngineConfig.trainingState` | `string` |  |  |  |
| `spec.mediaRecommendationEngineConfig.engineFeaturesConfig` | `GcpVertexAiSearchEngineEngineFeaturesConfig` |  |  |  |
| `spec.mediaRecommendationEngineConfig.engineFeaturesConfig.mostPopularConfig` | `GcpVertexAiSearchEngineMostPopularConfig` |  |  |  |
| `spec.mediaRecommendationEngineConfig.engineFeaturesConfig.mostPopularConfig.timeWindowDays` | `int32` |  |  |  |
| `spec.mediaRecommendationEngineConfig.engineFeaturesConfig.recommendedForYouConfig` | `GcpVertexAiSearchEngineRecommendedForYouConfig` |  |  |  |
| `spec.mediaRecommendationEngineConfig.engineFeaturesConfig.recommendedForYouConfig.contextEventType` | `string` |  |  |  |
| `spec.controls` | `[]GcpVertexAiSearchEngineControl` |  |  |  |
| `spec.controls[].controlId` | `string` | yes |  |  |
| `spec.controls[].displayName` | `string` | yes |  |  |
| `spec.controls[].useCases` | `[]string` |  |  |  |
| `spec.controls[].conditions` | `[]GcpVertexAiSearchEngineControlCondition` |  |  |  |
| `spec.controls[].conditions[].queryTerms` | `[]GcpVertexAiSearchEngineQueryTerm` |  |  |  |
| `spec.controls[].conditions[].queryTerms[].value` | `string` | yes |  |  |
| `spec.controls[].conditions[].queryTerms[].fullMatch` | `bool` |  |  |  |
| `spec.controls[].conditions[].activeTimeRanges` | `[]GcpVertexAiSearchEngineActiveTimeRange` |  |  |  |
| `spec.controls[].conditions[].activeTimeRanges[].startTime` | `string` |  |  |  |
| `spec.controls[].conditions[].activeTimeRanges[].endTime` | `string` |  |  |  |
| `spec.controls[].conditions[].queryRegex` | `string` |  |  |  |
| `spec.controls[].boostAction` | `GcpVertexAiSearchEngineBoostAction` |  |  |  |
| `spec.controls[].boostAction.dataStore` | `string \| valueFrom` | yes |  | GcpVertexAiSearchDataStore (`status.outputs.name`) |
| `spec.controls[].boostAction.filter` | `string` | yes |  |  |
| `spec.controls[].boostAction.fixedBoost` | `float` |  |  |  |
| `spec.controls[].boostAction.interpolationBoostSpec` | `GcpVertexAiSearchEngineInterpolationBoostSpec` |  |  |  |
| `spec.controls[].boostAction.interpolationBoostSpec.fieldName` | `string` |  |  |  |
| `spec.controls[].boostAction.interpolationBoostSpec.attributeType` | `string` |  |  |  |
| `spec.controls[].boostAction.interpolationBoostSpec.interpolationType` | `string` |  |  |  |
| `spec.controls[].boostAction.interpolationBoostSpec.controlPoint` | `GcpVertexAiSearchEngineControlPoint` |  |  |  |
| `spec.controls[].boostAction.interpolationBoostSpec.controlPoint.attributeValue` | `string` |  |  |  |
| `spec.controls[].boostAction.interpolationBoostSpec.controlPoint.boostAmount` | `float` |  |  |  |
| `spec.controls[].filterAction` | `GcpVertexAiSearchEngineFilterAction` |  |  |  |
| `spec.controls[].filterAction.dataStore` | `string \| valueFrom` | yes |  | GcpVertexAiSearchDataStore (`status.outputs.name`) |
| `spec.controls[].filterAction.filter` | `string` | yes |  |  |
| `spec.controls[].promoteAction` | `GcpVertexAiSearchEnginePromoteAction` |  |  |  |
| `spec.controls[].promoteAction.dataStore` | `string \| valueFrom` | yes |  | GcpVertexAiSearchDataStore (`status.outputs.name`) |
| `spec.controls[].promoteAction.searchLinkPromotion` | `GcpVertexAiSearchEngineSearchLinkPromotion` | yes |  |  |
| `spec.controls[].promoteAction.searchLinkPromotion.title` | `string` | yes |  |  |
| `spec.controls[].promoteAction.searchLinkPromotion.uri` | `string` |  |  |  |
| `spec.controls[].promoteAction.searchLinkPromotion.document` | `string` |  |  |  |
| `spec.controls[].promoteAction.searchLinkPromotion.description` | `string` |  |  |  |
| `spec.controls[].promoteAction.searchLinkPromotion.imageUri` | `string` |  |  |  |
| `spec.controls[].promoteAction.searchLinkPromotion.enabled` | `bool` |  |  |  |
| `spec.controls[].redirectAction` | `GcpVertexAiSearchEngineRedirectAction` |  |  |  |
| `spec.controls[].redirectAction.redirectUri` | `string` | yes |  |  |
| `spec.controls[].synonymsAction` | `GcpVertexAiSearchEngineSynonymsAction` |  |  |  |
| `spec.controls[].synonymsAction.synonyms` | `[]string` | yes |  |  |
| `spec.servingConfig` | `GcpVertexAiSearchEngineServingConfig` |  |  |  |
| `spec.servingConfig.boostControlIds` | `[]string` |  |  |  |
| `spec.servingConfig.filterControlIds` | `[]string` |  |  |  |
| `spec.servingConfig.promoteControlIds` | `[]string` |  |  |  |
| `spec.servingConfig.redirectControlIds` | `[]string` |  |  |  |
| `spec.servingConfig.synonymsControlIds` | `[]string` |  |  |  |
| `spec.widgetConfig` | `GcpVertexAiSearchEngineWidgetConfig` |  |  |  |
| `spec.widgetConfig.accessSettings` | `GcpVertexAiSearchEngineWidgetAccessSettings` |  |  |  |
| `spec.widgetConfig.accessSettings.allowPublicAccess` | `bool` |  |  |  |
| `spec.widgetConfig.accessSettings.allowlistedDomains` | `[]string` |  |  |  |
| `spec.widgetConfig.accessSettings.enableWebApp` | `bool` |  |  |  |
| `spec.widgetConfig.accessSettings.languageCode` | `string` |  |  |  |
| `spec.widgetConfig.accessSettings.workforceIdentityPoolProvider` | `string` |  |  |  |
| `spec.widgetConfig.homepageSetting` | `GcpVertexAiSearchEngineWidgetHomepageSetting` |  |  |  |
| `spec.widgetConfig.homepageSetting.shortcuts` | `[]GcpVertexAiSearchEngineWidgetShortcut` |  |  |  |
| `spec.widgetConfig.homepageSetting.shortcuts[].title` | `string` |  |  |  |
| `spec.widgetConfig.homepageSetting.shortcuts[].destinationUri` | `string` |  |  |  |
| `spec.widgetConfig.homepageSetting.shortcuts[].iconUrl` | `string` |  |  |  |
| `spec.widgetConfig.uiBranding` | `GcpVertexAiSearchEngineWidgetUiBranding` |  |  |  |
| `spec.widgetConfig.uiBranding.logoUrl` | `string` |  |  |  |
| `spec.widgetConfig.uiSettings` | `GcpVertexAiSearchEngineWidgetUiSettings` |  |  |  |
| `spec.widgetConfig.uiSettings.interactionType` | `string` |  |  |  |
| `spec.widgetConfig.uiSettings.resultDescriptionType` | `string` |  |  |  |
| `spec.widgetConfig.uiSettings.defaultSearchRequestOrderBy` | `string` |  |  |  |
| `spec.widgetConfig.uiSettings.disableUserEventsCollection` | `bool` |  |  |  |
| `spec.widgetConfig.uiSettings.enableAutocomplete` | `bool` |  |  |  |
| `spec.widgetConfig.uiSettings.enableCreateAgentButton` | `bool` |  |  |  |
| `spec.widgetConfig.uiSettings.enablePeopleSearch` | `bool` |  |  |  |
| `spec.widgetConfig.uiSettings.enableQualityFeedback` | `bool` |  |  |  |
| `spec.widgetConfig.uiSettings.enableSafeSearch` | `bool` |  |  |  |
| `spec.widgetConfig.uiSettings.enableSearchAsYouType` | `bool` |  |  |  |
| `spec.widgetConfig.uiSettings.enableVisualContentSummary` | `bool` |  |  |  |
| `spec.widgetConfig.uiSettings.dataStoreUiConfigs` | `[]GcpVertexAiSearchEngineWidgetDataStoreUiConfig` |  |  |  |
| `spec.widgetConfig.uiSettings.dataStoreUiConfigs[].name` | `string \| valueFrom` | yes |  | GcpVertexAiSearchDataStore (`status.outputs.name`) |
| `spec.widgetConfig.uiSettings.dataStoreUiConfigs[].facetFields` | `[]GcpVertexAiSearchEngineWidgetFacetField` |  |  |  |
| `spec.widgetConfig.uiSettings.dataStoreUiConfigs[].facetFields[].field` | `string` | yes |  |  |
| `spec.widgetConfig.uiSettings.dataStoreUiConfigs[].facetFields[].displayName` | `string` |  |  |  |
| `spec.widgetConfig.uiSettings.dataStoreUiConfigs[].fieldsUiComponentsMap` | `[]GcpVertexAiSearchEngineWidgetFieldUiComponent` |  |  |  |
| `spec.widgetConfig.uiSettings.dataStoreUiConfigs[].fieldsUiComponentsMap[].uiComponent` | `string` | yes |  |  |
| `spec.widgetConfig.uiSettings.dataStoreUiConfigs[].fieldsUiComponentsMap[].field` | `string` | yes |  |  |
| `spec.widgetConfig.uiSettings.dataStoreUiConfigs[].fieldsUiComponentsMap[].deviceVisibility` | `[]string` |  |  |  |
| `spec.widgetConfig.uiSettings.dataStoreUiConfigs[].fieldsUiComponentsMap[].displayTemplate` | `string` |  |  |  |
| `spec.widgetConfig.uiSettings.generativeAnswerConfig` | `GcpVertexAiSearchEngineWidgetGenerativeAnswerConfig` |  |  |  |
| `spec.widgetConfig.uiSettings.generativeAnswerConfig.disableRelatedQuestions` | `bool` |  |  |  |
| `spec.widgetConfig.uiSettings.generativeAnswerConfig.ignoreAdversarialQuery` | `bool` |  |  |  |
| `spec.widgetConfig.uiSettings.generativeAnswerConfig.ignoreLowRelevantContent` | `bool` |  |  |  |
| `spec.widgetConfig.uiSettings.generativeAnswerConfig.ignoreNonAnswerSeekingQuery` | `bool` |  |  |  |
| `spec.widgetConfig.uiSettings.generativeAnswerConfig.imageSource` | `string` |  |  |  |
| `spec.widgetConfig.uiSettings.generativeAnswerConfig.languageCode` | `string` |  |  |  |
| `spec.widgetConfig.uiSettings.generativeAnswerConfig.maxRephraseSteps` | `int32` |  |  |  |
| `spec.widgetConfig.uiSettings.generativeAnswerConfig.modelPromptPreamble` | `string` |  |  |  |
| `spec.widgetConfig.uiSettings.generativeAnswerConfig.modelVersion` | `string` |  |  |  |
| `spec.widgetConfig.uiSettings.generativeAnswerConfig.resultCount` | `int32` |  |  |  |
| `spec.assistants` | `[]GcpVertexAiSearchEngineAssistant` |  |  |  |
| `spec.assistants[].assistantId` | `string` | yes |  |  |
| `spec.assistants[].displayName` | `string` | yes |  |  |
| `spec.assistants[].description` | `string` |  |  |  |
| `spec.assistants[].webGroundingType` | `string` |  |  |  |
| `spec.assistants[].customerPolicy` | `GcpVertexAiSearchEngineCustomerPolicy` |  |  |  |
| `spec.assistants[].customerPolicy.bannedPhrases` | `[]GcpVertexAiSearchEngineBannedPhrase` |  |  |  |
| `spec.assistants[].customerPolicy.bannedPhrases[].phrase` | `string` | yes |  |  |
| `spec.assistants[].customerPolicy.bannedPhrases[].matchType` | `string` |  |  |  |
| `spec.assistants[].customerPolicy.bannedPhrases[].ignoreDiacritics` | `bool` |  |  |  |
| `spec.assistants[].customerPolicy.modelArmorConfig` | `GcpVertexAiSearchEngineModelArmorConfig` |  |  |  |
| `spec.assistants[].customerPolicy.modelArmorConfig.userPromptTemplate` | `string` | yes |  |  |
| `spec.assistants[].customerPolicy.modelArmorConfig.responseTemplate` | `string` | yes |  |  |
| `spec.assistants[].customerPolicy.modelArmorConfig.failureMode` | `string` |  |  |  |
| `spec.assistants[].generationConfig` | `GcpVertexAiSearchEngineGenerationConfig` |  |  |  |
| `spec.assistants[].generationConfig.defaultLanguage` | `string` |  |  |  |
| `spec.assistants[].generationConfig.additionalSystemInstruction` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the engine lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

Where the engine lives: "global", "us", or "eu" -- the same location as
its data stores. Immutable.

- rule: {"required":true,"string":{"in":["global","us","eu"]}}

### spec.engineType

`string`

Which app this engine is. Empty or SEARCH builds a search engine; CHAT
a chat engine over a Dialogflow CX agent; RECOMMENDATION a
recommendation engine. Immutable: a change replaces the engine.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["SEARCH","CHAT","RECOMMENDATION"]}}

### spec.engineId

`string`

The engine's id -- 1-63 characters, RFC 1034 (lowercase letters,
digits, hyphens; starts with a letter). Defaults to metadata.name.
Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?$"}}

### spec.displayName

`string`

Human-readable name shown in the console (up to 1024 characters).
Defaults to metadata.name. Mutable.

- rule: {"string":{"maxLen":"1024"}}

### spec.collectionId

`string | valueFrom`

The collection the engine and its data stores live in. Empty is
Google's "default_collection", where every GcpVertexAiSearchDataStore
lives; a GcpVertexAiSearchDataConnector reference (or its literal
collection id) puts the engine over the data stores a connector syncs.
A RECOMMENDATION engine always lives in default_collection. Immutable.

- references: GcpVertexAiSearchDataConnector (`status.outputs.collection_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVertexAiSearchDataConnector, name: <that resource's name>, fieldPath: status.outputs.collection_id}} -- a bare string does not parse

### spec.dataStoreIds

`[]string | valueFrom` · required

The data stores the engine answers from, as GcpVertexAiSearchDataStore
references (their data_store_id) or literal ids, all in the engine's
collection and location and enrolled in its solution (SOLUTION_TYPE_CHAT
for a chat engine). A recommendation engine takes at most one. Mutable.

- references: GcpVertexAiSearchDataStore (`status.outputs.data_store_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVertexAiSearchDataStore, name: <that resource's name>, fieldPath: status.outputs.data_store_id}} -- a bare string does not parse

### spec.industryVertical

`string`

The industry vertical, matching the data stores': GENERIC (Google's
default when empty), MEDIA (search or recommendation over media), or
HEALTHCARE_FHIR (search only). A chat engine is always GENERIC.
Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["GENERIC","MEDIA","HEALTHCARE_FHIR"]}}

### spec.commonConfig

`GcpVertexAiSearchEngineCommonConfig`

Metadata common to every engine type. Immutable.

### spec.commonConfig.companyName

`string`

The company, business, or entity the engine represents; improves
LLM features (summaries, answers) that mention it. Immutable.

### spec.searchEngineConfig

`GcpVertexAiSearchEngineSearchEngineConfig`

The search arm's tier, add-ons, and license tier (SEARCH only). Google
requires the block, so the module sends it even when omitted here.

### spec.searchEngineConfig.searchTier

`string`

SEARCH_TIER_STANDARD (Google's default when empty) or
SEARCH_TIER_ENTERPRISE, which unlocks website search with extractive
answers, structured data search enrichment, and the LLM add-on. The
Enterprise tier bills a higher per-query rate. Mutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["SEARCH_TIER_STANDARD","SEARCH_TIER_ENTERPRISE"]}}

### spec.searchEngineConfig.searchAddOns

`[]string`

Add-ons: SEARCH_ADD_ON_LLM turns on generative answers and summaries
(requires the Enterprise tier). Sent only when set.

- rule: {"repeated":{"items":{"string":{"in":["SEARCH_ADD_ON_LLM"]}}}}

### spec.searchEngineConfig.requiredSubscriptionTier

`string`

The Gemini Enterprise license tier a user needs to open this app.
Empty lets Google default. Sent only when set.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["SUBSCRIPTION_TIER_UNSPECIFIED","SUBSCRIPTION_TIER_SEARCH","SUBSCRIPTION_TIER_SEARCH_AND_ASSISTANT","SUBSCRIPTION_TIER_FRONTLINE_WORKER","SUBSCRIPTION_TIER_AGENTSPACE_STARTER","SUBSCRIPTION_TIER_AGENTSPACE_BUSINESS","SUBSCRIPTION_TIER_ENTERPRISE","SUBSCRIPTION_TIER_ENTERPRISE_EMERGING","SUBSCRIPTION_TIER_EDU","SUBSCRIPTION_TIER_EDU_PRO","SUBSCRIPTION_TIER_EDU_EMERGING","SUBSCRIPTION_TIER_EDU_PRO_EMERGING","SUBSCRIPTION_TIER_FRONTLINE_STARTER"]}}

### spec.appType

`string`

APP_TYPE_INTRANET marks a Gemini Enterprise (intranet) app; empty is a
standalone search app (SEARCH only). Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["APP_TYPE_INTRANET"]}}

### spec.disableAnalytics

`bool`

Do not record search analytics for this engine (SEARCH only).

### spec.features

`map<string, string>`

Feature opt-ins and opt-outs by name, each "FEATURE_STATE_ON" or
"FEATURE_STATE_OFF" -- e.g. "agent-sharing-without-admin-approval",
"disable-agent-sharing", "enable-end-user-sharing-with-groups" (SEARCH
only). Sent only when set.

- rule: {"map":{"values":{"string":{"in":["FEATURE_STATE_ON","FEATURE_STATE_OFF"]}}}}

### spec.kmsKeyName

`string | valueFrom`

Customer-managed encryption key protecting the engine (SEARCH only):
a GcpKmsKey reference or a literal
projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}.
Omit for Google-managed encryption. Mutable.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.knowledgeGraphConfig

`GcpVertexAiSearchEngineKnowledgeGraphConfig`

Cloud and private knowledge graphs (SEARCH only). Sent only when set.

### spec.knowledgeGraphConfig.enableCloudKnowledgeGraph

`bool` · optional (explicit presence)

Use Google's Cloud Knowledge Graph (public entities) for the engine.

### spec.knowledgeGraphConfig.enablePrivateKnowledgeGraph

`bool` · optional (explicit presence)

Build and use a private knowledge graph over the engine's data.

### spec.knowledgeGraphConfig.cloudKnowledgeGraphTypes

`[]string`

Cloud Knowledge Graph entity types to support.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.knowledgeGraphConfig.featureConfig

`GcpVertexAiSearchEngineKnowledgeGraphFeatureConfig`

Private knowledge graph features to switch off.

### spec.knowledgeGraphConfig.featureConfig.disablePrivateKgAutoComplete

`bool`

Do not use the private knowledge graph for query auto-complete.

### spec.knowledgeGraphConfig.featureConfig.disablePrivateKgEnrichment

`bool`

Do not enrich results with private knowledge graph entities.

### spec.knowledgeGraphConfig.featureConfig.disablePrivateKgQueryUiChips

`bool`

Do not show private knowledge graph entities as query UI chips.

### spec.knowledgeGraphConfig.featureConfig.disablePrivateKgQueryUnderstanding

`bool`

Do not use the private knowledge graph for query understanding.

### spec.chatEngineConfig

`GcpVertexAiSearchEngineChatEngineConfig`

The chat arm's agent (CHAT only; required there). Immutable.

- rule: a chat engine is exactly one of agent_creation_config or dialogflow_agent_to_link

### spec.chatEngineConfig.agentCreationConfig

`GcpVertexAiSearchEngineAgentCreationConfig`

Create a new Dialogflow CX agent for this engine.

### spec.chatEngineConfig.agentCreationConfig.business

`string`

The company, organization, or entity the agent represents; used in the
knowledge connector's LLM prompt.

### spec.chatEngineConfig.agentCreationConfig.defaultLanguageCode

`string` · required

The agent's default language as a language tag, e.g. "en".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.chatEngineConfig.agentCreationConfig.timeZone

`string` · required

The agent's IANA time zone, e.g. "America/New_York".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.chatEngineConfig.agentCreationConfig.location

`string`

The Dialogflow location the agent is created in, e.g. "global" or
"us-central1". Empty lets Google pick; a location other than the
engine's needs allow_cross_region.

### spec.chatEngineConfig.dialogflowAgentToLink

`string`

Link an existing Dialogflow CX agent:
projects/{project}/locations/{location}/agents/{agent}.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^projects/[a-zA-Z0-9-]+(?:/locations/[a-zA-Z0-9-]+)?/agents/[a-zA-Z0-9-]+$"}}

### spec.chatEngineConfig.allowCrossRegion

`bool`

Allow the agent and the engine to live in different locations
(Google's default requires the same location). Consumed once at
creation; Google does not read it back.

### spec.mediaRecommendationEngineConfig

`GcpVertexAiSearchEngineMediaRecommendationEngineConfig`

The recommendation arm's model (RECOMMENDATION only; MEDIA vertical).

### spec.mediaRecommendationEngineConfig.type

`string`

The recommendation model: "recommended-for-you", "others-you-may-like",
"more-like-this", or "most-popular-items". With optimization_objective
it decides how the engine trains and serves.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["recommended-for-you","others-you-may-like","more-like-this","most-popular-items"]}}

### spec.mediaRecommendationEngineConfig.optimizationObjective

`string`

What the model optimizes: "ctr" (click-through) or "cvr" (conversion,
e.g. watch time). Empty lets Google default by type (ctr for
recommended-for-you and others-you-may-like, and for more-like-this
and most-popular-items).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ctr","cvr"]}}

### spec.mediaRecommendationEngineConfig.optimizationObjectiveConfig

`GcpVertexAiSearchEngineOptimizationObjectiveConfig`

The threshold a cvr objective optimizes toward.

### spec.mediaRecommendationEngineConfig.optimizationObjectiveConfig.targetField

`string` · required

"watch-percentage" (a fraction in (0, 1]) or "watch-time" (seconds in
(0, 86400]).

- rule: {"required":true,"string":{"in":["watch-percentage","watch-time"]}}

### spec.mediaRecommendationEngineConfig.optimizationObjectiveConfig.targetFieldValueFloat

`float` · optional (explicit presence)

The threshold for the target field, e.g. 0.5 for half the media
watched, or 90 for ninety seconds.

### spec.mediaRecommendationEngineConfig.trainingState

`string`

TRAINING (Google's default at creation) or PAUSED. Training is part of
the engine's cost; pause it to hold a model still. Mutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["PAUSED","TRAINING"]}}

### spec.mediaRecommendationEngineConfig.engineFeaturesConfig

`GcpVertexAiSearchEngineEngineFeaturesConfig`

Feature config the chosen type needs.

- rule: engine_features_config is at most one of most_popular_config or recommended_for_you_config

### spec.mediaRecommendationEngineConfig.engineFeaturesConfig.mostPopularConfig

`GcpVertexAiSearchEngineMostPopularConfig`

For type "most-popular-items".

### spec.mediaRecommendationEngineConfig.engineFeaturesConfig.mostPopularConfig.timeWindowDays

`int32` · optional (explicit presence)

How many days of events the engine trains and predicts on. Required
for most-popular-items.

- rule: {"int32":{"gte":1}}

### spec.mediaRecommendationEngineConfig.engineFeaturesConfig.recommendedForYouConfig

`GcpVertexAiSearchEngineRecommendedForYouConfig`

For type "recommended-for-you".

### spec.mediaRecommendationEngineConfig.engineFeaturesConfig.recommendedForYouConfig.contextEventType

`string`

The event the engine is queried with at prediction time: "generic"
(view-item, media-play, media-complete) or "view-home-page" (those
plus the home-page view).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["generic","view-home-page"]}}

### spec.controls

`[]GcpVertexAiSearchEngineControl`

Serving controls on the engine, each keyed by control_id. A control
takes effect only when serving_config lists it.

- rule: a control is exactly one of boost_action, filter_action, promote_action, redirect_action, or synonyms_action

### spec.controls[].controlId

`string` · required

The control's id within the engine -- 1-63 characters, RFC 1034.
Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?$"}}

### spec.controls[].displayName

`string` · required

Human-readable name (up to 128 characters).

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.controls[].useCases

`[]string`

Where the control applies: SEARCH_USE_CASE_SEARCH (queries) and/or
SEARCH_USE_CASE_BROWSE (empty-query browsing).

- rule: {"repeated":{"items":{"string":{"in":["SEARCH_USE_CASE_SEARCH","SEARCH_USE_CASE_BROWSE"]}}}}

### spec.controls[].conditions

`[]GcpVertexAiSearchEngineControlCondition`

When the control is active; empty means always.

### spec.controls[].conditions[].queryTerms

`[]GcpVertexAiSearchEngineQueryTerm`

Terms the query must contain.

### spec.controls[].conditions[].queryTerms[].value

`string` · required

The term.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.controls[].conditions[].queryTerms[].fullMatch

`bool`

True requires the query to match the term exactly; false allows a
partial match.

### spec.controls[].conditions[].activeTimeRanges

`[]GcpVertexAiSearchEngineActiveTimeRange`

Time windows the control is active in (any one holding satisfies the
condition).

### spec.controls[].conditions[].activeTimeRanges[].startTime

`string`

Start of the window, e.g. "2026-11-20T00:00:00Z".

### spec.controls[].conditions[].activeTimeRanges[].endTime

`string`

End of the window.

### spec.controls[].conditions[].queryRegex

`string`

A regular expression the whole query must match.

### spec.controls[].boostAction

`GcpVertexAiSearchEngineBoostAction`

Reorder matching results.

- rule: a boost action is exactly one of fixed_boost or interpolation_boost_spec

### spec.controls[].boostAction.dataStore

`string | valueFrom` · required

The data store whose documents are boosted: a GcpVertexAiSearchDataStore
reference (its full name) or a literal
projects/{project}/locations/{location}/collections/{collection}/dataStores/{store}.

- references: GcpVertexAiSearchDataStore (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVertexAiSearchDataStore, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.controls[].boostAction.filter

`string` · required

The filter selecting the documents to boost, in Vertex AI Search filter
syntax, e.g. "(category: ANY(\"docs\"))".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.controls[].boostAction.fixedBoost

`float` · optional (explicit presence)

A fixed boost, -1 (bury) to 1 (promote).

- rule: {"float":{"lte":1,"gte":-1}}

### spec.controls[].boostAction.interpolationBoostSpec

`GcpVertexAiSearchEngineInterpolationBoostSpec`

A curve-based boost instead of a fixed one.

### spec.controls[].boostAction.interpolationBoostSpec.fieldName

`string`

The document field the curve reads.

### spec.controls[].boostAction.interpolationBoostSpec.attributeType

`string`

NUMERICAL (a numeric field) or FRESHNESS (a datetime field, boosted by
age). Sent only when set.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NUMERICAL","FRESHNESS"]}}

### spec.controls[].boostAction.interpolationBoostSpec.interpolationType

`string`

LINEAR is the only interpolation Google offers. Sent only when set.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["LINEAR"]}}

### spec.controls[].boostAction.interpolationBoostSpec.controlPoint

`GcpVertexAiSearchEngineControlPoint`

The curve's control point.

### spec.controls[].boostAction.interpolationBoostSpec.controlPoint.attributeValue

`string`

The attribute value at this point (a number for NUMERICAL, a duration
like "7d" or "24h" for FRESHNESS).

### spec.controls[].boostAction.interpolationBoostSpec.controlPoint.boostAmount

`float` · optional (explicit presence)

The boost at this point, -1 to 1.

- rule: {"float":{"lte":1,"gte":-1}}

### spec.controls[].filterAction

`GcpVertexAiSearchEngineFilterAction`

Drop non-matching results.

### spec.controls[].filterAction.dataStore

`string | valueFrom` · required

The data store the filter applies to (a GcpVertexAiSearchDataStore
reference or a literal full name).

- references: GcpVertexAiSearchDataStore (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVertexAiSearchDataStore, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.controls[].filterAction.filter

`string` · required

The filter results must match, in Vertex AI Search filter syntax.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.controls[].promoteAction

`GcpVertexAiSearchEnginePromoteAction`

Pin a link.

### spec.controls[].promoteAction.dataStore

`string | valueFrom` · required

The data store the promotion applies to (a GcpVertexAiSearchDataStore
reference or a literal full name).

- references: GcpVertexAiSearchDataStore (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVertexAiSearchDataStore, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.controls[].promoteAction.searchLinkPromotion

`GcpVertexAiSearchEngineSearchLinkPromotion` · required

The promoted link.

- rule: {"required":true}

### spec.controls[].promoteAction.searchLinkPromotion.title

`string` · required

The promoted link's title.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.controls[].promoteAction.searchLinkPromotion.uri

`string`

The URI to promote (website stores).

### spec.controls[].promoteAction.searchLinkPromotion.document

`string`

The document to promote (non-website stores), as a document resource
name.

### spec.controls[].promoteAction.searchLinkPromotion.description

`string`

A description shown with the link.

### spec.controls[].promoteAction.searchLinkPromotion.imageUri

`string`

An image shown with the link.

### spec.controls[].promoteAction.searchLinkPromotion.enabled

`bool`

Return the promotion on basic site search results.

### spec.controls[].redirectAction

`GcpVertexAiSearchEngineRedirectAction`

Redirect the query.

### spec.controls[].redirectAction.redirectUri

`string` · required

The URI to redirect to.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.controls[].synonymsAction

`GcpVertexAiSearchEngineSynonymsAction`

Treat terms as synonyms.

### spec.controls[].synonymsAction.synonyms

`[]string` · required

The synonyms, e.g. ["laptop", "notebook"].

- rule: {"repeated":{"minItems":"2","items":{"string":{"minLen":"1"}}}}

### spec.servingConfig

`GcpVertexAiSearchEngineServingConfig`

Which controls the engine's default serving config ("default_search",
which Google creates with a search engine) applies. Every id must name
a control above with the matching action.

### spec.servingConfig.boostControlIds

`[]string`

Ids of boost controls to apply.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.servingConfig.filterControlIds

`[]string`

Ids of filter controls to apply.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.servingConfig.promoteControlIds

`[]string`

Ids of promote controls to apply.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.servingConfig.redirectControlIds

`[]string`

Ids of redirect controls to apply.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.servingConfig.synonymsControlIds

`[]string`

Ids of synonyms controls to apply.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.widgetConfig

`GcpVertexAiSearchEngineWidgetConfig`

The embeddable search widget and hosted web app. Google creates a
search engine's default widget config with the engine; this PATCHes
it, and destroy leaves it in place.

### spec.widgetConfig.accessSettings

`GcpVertexAiSearchEngineWidgetAccessSettings`

Who may load the widget.

### spec.widgetConfig.accessSettings.allowPublicAccess

`bool`

Allow unauthenticated public access to the widget.

### spec.widgetConfig.accessSettings.allowlistedDomains

`[]string`

Domains allowed to embed the widget, e.g. "www.example.com".

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.widgetConfig.accessSettings.enableWebApp

`bool`

Serve Google's hosted web app for the engine.

### spec.widgetConfig.accessSettings.languageCode

`string`

UI language as a BCP 47 tag; Google defaults to "en-US".

### spec.widgetConfig.accessSettings.workforceIdentityPoolProvider

`string`

The workforce identity pool provider users sign in through:
locations/global/workforcePools/{pool}/providers/{provider}.

### spec.widgetConfig.homepageSetting

`GcpVertexAiSearchEngineWidgetHomepageSetting`

The homepage.

### spec.widgetConfig.homepageSetting.shortcuts

`[]GcpVertexAiSearchEngineWidgetShortcut`

Shortcuts shown on the homepage.

### spec.widgetConfig.homepageSetting.shortcuts[].title

`string`

The shortcut's title.

### spec.widgetConfig.homepageSetting.shortcuts[].destinationUri

`string`

Where the shortcut goes.

### spec.widgetConfig.homepageSetting.shortcuts[].iconUrl

`string`

The shortcut's icon URL.

### spec.widgetConfig.uiBranding

`GcpVertexAiSearchEngineWidgetUiBranding`

Branding.

### spec.widgetConfig.uiBranding.logoUrl

`string`

The logo image URL.

### spec.widgetConfig.uiSettings

`GcpVertexAiSearchEngineWidgetUiSettings`

UI behavior.

### spec.widgetConfig.uiSettings.interactionType

`string`

SEARCH_ONLY, SEARCH_WITH_ANSWER (a generated answer above results), or
SEARCH_WITH_FOLLOW_UPS (conversational). Sent only when set.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["SEARCH_ONLY","SEARCH_WITH_ANSWER","SEARCH_WITH_FOLLOW_UPS"]}}

### spec.widgetConfig.uiSettings.resultDescriptionType

`string`

SNIPPET or EXTRACTIVE_ANSWER under each result; empty shows none.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["SNIPPET","EXTRACTIVE_ANSWER"]}}

### spec.widgetConfig.uiSettings.defaultSearchRequestOrderBy

`string`

The default order of results (a SearchRequest orderBy expression).

### spec.widgetConfig.uiSettings.disableUserEventsCollection

`bool`

Do not collect user events from the widget.

### spec.widgetConfig.uiSettings.enableAutocomplete

`bool`

Offer query auto-complete.

### spec.widgetConfig.uiSettings.enableCreateAgentButton

`bool`

Show the "create agent" button.

### spec.widgetConfig.uiSettings.enablePeopleSearch

`bool`

Enable people search.

### spec.widgetConfig.uiSettings.enableQualityFeedback

`bool`

Collect result-quality feedback from end users.

### spec.widgetConfig.uiSettings.enableSafeSearch

`bool`

Enable safe search.

### spec.widgetConfig.uiSettings.enableSearchAsYouType

`bool`

Search as the user types.

### spec.widgetConfig.uiSettings.enableVisualContentSummary

`bool`

Visual content summaries on applicable searches (healthcare search).

### spec.widgetConfig.uiSettings.dataStoreUiConfigs

`[]GcpVertexAiSearchEngineWidgetDataStoreUiConfig`

Per-data-store result rendering.

### spec.widgetConfig.uiSettings.dataStoreUiConfigs[].name

`string | valueFrom` · required

The data store: a GcpVertexAiSearchDataStore reference (its full name)
or a literal full name.

- references: GcpVertexAiSearchDataStore (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVertexAiSearchDataStore, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.widgetConfig.uiSettings.dataStoreUiConfigs[].facetFields

`[]GcpVertexAiSearchEngineWidgetFacetField`

Fields offered as facets.

### spec.widgetConfig.uiSettings.dataStoreUiConfigs[].facetFields[].field

`string` · required

The registered field name, e.g. "category".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.widgetConfig.uiSettings.dataStoreUiConfigs[].facetFields[].displayName

`string`

The name end users see.

### spec.widgetConfig.uiSettings.dataStoreUiConfigs[].fieldsUiComponentsMap

`[]GcpVertexAiSearchEngineWidgetFieldUiComponent`

Which document field each result component shows.

### spec.widgetConfig.uiSettings.dataStoreUiConfigs[].fieldsUiComponentsMap[].uiComponent

`string` · required

The component: "title", "thumbnail", "url", "custom1", "custom2", or
"custom3".

- rule: {"required":true,"string":{"in":["title","thumbnail","url","custom1","custom2","custom3"]}}

### spec.widgetConfig.uiSettings.dataStoreUiConfigs[].fieldsUiComponentsMap[].field

`string` · required

The field shown in the component.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.widgetConfig.uiSettings.dataStoreUiConfigs[].fieldsUiComponentsMap[].deviceVisibility

`[]string`

Devices the component shows on: MOBILE and/or DESKTOP.

- rule: {"repeated":{"items":{"string":{"in":["MOBILE","DESKTOP"]}}}}

### spec.widgetConfig.uiSettings.dataStoreUiConfigs[].fieldsUiComponentsMap[].displayTemplate

`string`

A template around the value, e.g. "Price: {value}".

### spec.widgetConfig.uiSettings.generativeAnswerConfig

`GcpVertexAiSearchEngineWidgetGenerativeAnswerConfig`

The generated answer's settings.

### spec.widgetConfig.uiSettings.generativeAnswerConfig.disableRelatedQuestions

`bool`

Do not suggest related questions with the answer.

### spec.widgetConfig.uiSettings.generativeAnswerConfig.ignoreAdversarialQuery

`bool`

Skip answering queries classified as adversarial.

### spec.widgetConfig.uiSettings.generativeAnswerConfig.ignoreLowRelevantContent

`bool`

Skip answering when the results are not relevant to the query.

### spec.widgetConfig.uiSettings.generativeAnswerConfig.ignoreNonAnswerSeekingQuery

`bool`

Skip answering queries that do not seek an answer (navigational or
browsing queries).

### spec.widgetConfig.uiSettings.generativeAnswerConfig.imageSource

`string`

Where answer images come from: ALL_AVAILABLE_SOURCES, CORPUS_IMAGE_ONLY,
or FIGURE_GENERATION_ONLY. Sent only when set.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ALL_AVAILABLE_SOURCES","CORPUS_IMAGE_ONLY","FIGURE_GENERATION_ONLY"]}}

### spec.widgetConfig.uiSettings.generativeAnswerConfig.languageCode

`string`

The answer's language as a BCP 47 tag (experimental).

### spec.widgetConfig.uiSettings.generativeAnswerConfig.maxRephraseSteps

`int32` · optional (explicit presence)

How many times the query may be rephrased, 1-5 (Google defaults to 1).

- rule: {"int32":{"lte":5,"gte":1}}

### spec.widgetConfig.uiSettings.generativeAnswerConfig.modelPromptPreamble

`string`

Text placed before the prompt to steer the answer model.

### spec.widgetConfig.uiSettings.generativeAnswerConfig.modelVersion

`string`

The answer model version.

### spec.widgetConfig.uiSettings.generativeAnswerConfig.resultCount

`int32` · optional (explicit presence)

Top results the answer is generated from, up to 10.

- rule: {"int32":{"lte":10,"gte":1}}

### spec.assistants

`[]GcpVertexAiSearchEngineAssistant`

Gemini Enterprise assistants on the engine, keyed by assistant_id
(a Gemini Enterprise license on the project is Google's prerequisite).

### spec.assistants[].assistantId

`string` · required

The assistant's id within the engine, e.g. "default_assistant".
Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z0-9](?:[-a-z0-9_]{0,61}[a-z0-9])?$"}}

### spec.assistants[].displayName

`string` · required

Human-readable name (up to 128 characters).

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.assistants[].description

`string`

Notes shown on the configuration UI, not to end users.

### spec.assistants[].webGroundingType

`string`

WEB_GROUNDING_TYPE_DISABLED, WEB_GROUNDING_TYPE_GOOGLE_SEARCH, or
WEB_GROUNDING_TYPE_ENTERPRISE_WEB_SEARCH. Sent only when set.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["WEB_GROUNDING_TYPE_DISABLED","WEB_GROUNDING_TYPE_GOOGLE_SEARCH","WEB_GROUNDING_TYPE_ENTERPRISE_WEB_SEARCH"]}}

### spec.assistants[].customerPolicy

`GcpVertexAiSearchEngineCustomerPolicy`

Content policy.

### spec.assistants[].customerPolicy.bannedPhrases

`[]GcpVertexAiSearchEngineBannedPhrase`

Phrases the assistant refuses to engage with.

### spec.assistants[].customerPolicy.bannedPhrases[].phrase

`string` · required

The phrase.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.assistants[].customerPolicy.bannedPhrases[].matchType

`string`

SIMPLE_STRING_MATCH or WORD_BOUNDARY_STRING_MATCH. Sent only when set.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["SIMPLE_STRING_MATCH","WORD_BOUNDARY_STRING_MATCH"]}}

### spec.assistants[].customerPolicy.bannedPhrases[].ignoreDiacritics

`bool`

Ignore accents and umlauts when matching ("cafe" matches "café").

### spec.assistants[].customerPolicy.modelArmorConfig

`GcpVertexAiSearchEngineModelArmorConfig`

Model Armor sanitization.

### spec.assistants[].customerPolicy.modelArmorConfig.userPromptTemplate

`string` · required

The template applied to user prompts.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.assistants[].customerPolicy.modelArmorConfig.responseTemplate

`string` · required

The template applied to assistant responses.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.assistants[].customerPolicy.modelArmorConfig.failureMode

`string`

FAIL_OPEN (answer anyway when sanitization fails) or FAIL_CLOSED
(refuse). Sent only when set.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["FAIL_OPEN","FAIL_CLOSED"]}}

### spec.assistants[].generationConfig

`GcpVertexAiSearchEngineGenerationConfig`

Answer generation.

### spec.assistants[].generationConfig.defaultLanguage

`string`

Default answer language (ISO 639-1, e.g. "en"); empty auto-detects.

### spec.assistants[].generationConfig.additionalSystemInstruction

`string`

Text appended to the default system instruction.

### spec.deletionPolicy

`string`

What happens to the engine, its controls, and its assistants when this
resource is destroyed:
  "" / "DELETE" -- deleted (the widget config and serving config are
                   Google's own and stay with the engine's deletion)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- everything leaves management and stays in GCP
The data stores are separate resources and are not touched.

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `engine.chat_config_on_chat_arm`: chat_engine_config is required on a CHAT engine and applies only there
- `engine.recommendation_config_on_recommendation_arm`: media_recommendation_engine_config applies only to a RECOMMENDATION engine
- `engine.search_only_fields`: search_engine_config, app_type, disable_analytics, features, kms_key_name, and knowledge_graph_config exist only on a SEARCH engine
- `engine.recommendation_default_collection`: a RECOMMENDATION engine lives in default_collection; collection_id must be empty
- `engine.recommendation_one_data_store`: a RECOMMENDATION engine uses at most one data store
- `engine.industry_vertical_per_arm`: a CHAT engine is GENERIC; a RECOMMENDATION engine is GENERIC or MEDIA
- `engine.controls_unique_ids`: control_id must be unique within the engine
- `engine.assistants_unique_ids`: assistant_id must be unique within the engine
- `engine.serving_config_names_declared_controls`: every id in serving_config must name a control declared on this engine with the matching action

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVertexAiSearchEngine, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name of the engine: projects/{project}/locations/{location}/collections/{collection_id}/engines/{engine_id}. |
| `status.outputs.engine_id` | `string` | The engine's id -- what search and answer requests address the app by. |
| `status.outputs.location` | `string` | The engine's location (global, us, or eu). |
| `status.outputs.collection_id` | `string` | The collection the engine lives in (default_collection unless a connector's collection was named). |
| `status.outputs.engine_type` | `string` | Which app was built: SEARCH, CHAT, or RECOMMENDATION. |
| `status.outputs.serving_config_name` | `string` | Full resource name of the engine's default serving config when the spec configured it (empty otherwise). |
| `status.outputs.widget_config_name` | `string` | Full resource name of the widget config when the spec configured it (empty otherwise). |
| `status.outputs.dialogflow_agent` | `string` | The Dialogflow CX agent a CHAT engine answers through (projects/{project}/locations/{location}/agents/{agent}); empty on other engine types. |
| `status.outputs.control_names` | `[]string` | Full resource names of the controls declared on the engine, in manifest order. |
| `status.outputs.assistant_names` | `[]string` | Full resource names of the assistants declared on the engine, in manifest order. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.collectionId` | GcpVertexAiSearchDataConnector | `status.outputs.collection_id` |
| `spec.dataStoreIds` | GcpVertexAiSearchDataStore | `status.outputs.data_store_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.controls[].boostAction.dataStore` | GcpVertexAiSearchDataStore | `status.outputs.name` |
| `spec.controls[].filterAction.dataStore` | GcpVertexAiSearchDataStore | `status.outputs.name` |
| `spec.controls[].promoteAction.dataStore` | GcpVertexAiSearchDataStore | `status.outputs.name` |
| `spec.widgetConfig.uiSettings.dataStoreUiConfigs[].name` | GcpVertexAiSearchDataStore | `status.outputs.name` |

## See Also

- [Overview](../README.md)
