# Search Standard

## Use Case

The smallest search app: a Standard-tier search engine over one data store, queried through the Discovery Engine API by your own application. No controls, no widget -- the engine ranks and returns results, your code renders them.

## When to Use

- The first search engine over a new data store
- An application that calls the search API itself and needs no hosted widget
- Keyword and semantic search without generated answers

## What This Creates

- A SEARCH engine in the `global` location over the referenced `GcpVertexAiSearchDataStore`, Standard tier, GENERIC vertical

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `dataStoreIds` | one store | Add more stores in the same location and collection to search them together (blended search). |
| `searchEngineConfig.searchTier` | Standard | `SEARCH_TIER_ENTERPRISE` for extractive answers, website enrichment, and the LLM add-on, at a higher per-query rate. |
| `controls` + `servingConfig` | none | Synonyms, boosts, filters, promotions, and redirects, applied through the serving config. |
| `widgetConfig` | none | Google's embeddable widget and hosted web app. |
| `engineType` | SEARCH | `CHAT` or `RECOMMENDATION` build the other two engines (see the other presets). |

The engine must live in the same location as its data stores, and the stores must be enrolled in `SOLUTION_TYPE_SEARCH`. Creating an engine is free; queries bill per thousand at the tier's rate.
