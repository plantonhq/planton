# Structured Records

## Use Case

A store of structured JSON records -- products, tickets, people, articles as fields -- searched and recommended by their fields. Google infers the schema from the first import, so the store is ready the moment records arrive through the API, BigQuery, or Cloud Storage.

## When to Use

- A product catalog, a knowledge base of FAQs, a directory
- Anything that already lives as rows in BigQuery or JSON in a bucket
- The first Vertex AI Search store in a project

## What This Creates

- A `NO_CONTENT` data store in the `global` location, GENERIC vertical, enrolled in search and recommendation
- Google's default schema, inferred from the data

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `global` | `us` or `eu` to keep the data in one region; every engine over the store must match. |
| `solutionTypes` | search + recommendation | Drop recommendation if no recommendation engine will use the store; add `SOLUTION_TYPE_CHAT` for a chat engine. |
| `skipDefaultSchemaCreation` + `schema` | inferred | Declare your own JSON Schema when field types, facets, or key properties must be exact. |
| `kmsKeyName` | Google-managed | A `GcpKmsKey` in the store's location for customer-managed encryption. |
| `deletionPolicy` | `DELETE` | `PREVENT` once the store holds data you cannot re-import. |

Everything but the display name and the KMS key is immutable: a change replaces the store, documents included.
