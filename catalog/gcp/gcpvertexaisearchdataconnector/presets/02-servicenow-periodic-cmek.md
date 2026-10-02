# ServiceNow Periodic with CMEK

## Use Case

Index a ServiceNow instance's knowledge base, service catalog, and incidents into Vertex AI Search on a schedule, in the `us` multi-region, with every created data store encrypted under your own key. Incident titles and descriptions map to the search result's title and description; incidents are filtered to one knowledge base.

## When to Use

- Enterprise search or an assistant grounded on ServiceNow knowledge
- Regulated data that must stay in one region under a customer-managed key
- When you want the data indexed (fast, ranked search) rather than searched live

## What This Creates

- A data connector in `us` for the `servicenow` source that creates a collection with three data stores: knowledge base, catalog, incidents
- `DATA_INGESTION` mode with a daily full sync and six-hourly incremental syncs, key property mappings on incidents, an inclusion filter
- Encryption of every created store under a referenced `GcpKmsKey`; `PREVENT` as the destroy policy

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `params` | OAuth password grant | Your instance, service user, and the Secret Manager secrets holding the client secret and password. |
| `refreshInterval` / `incrementalRefreshInterval` | daily / six hours | Between 30 minutes and 7 days; equal values disable incremental sync. |
| `entities[].keyPropertyMappings` | title, description on incidents | Which source fields render as a result's title and description for each entity. |
| `entities[].params` | one knowledge base | Inclusion filters per entity, as compact JSON. |
| `kmsKeyName` | a `GcpKmsKey` reference | Remove for Google-managed encryption. |

Everything about the collection, the source, the location, the key, and each entity's name is immutable; the schedule, parameters, and filters update in place. The Discovery Engine service agent needs access to the secrets and the key.
