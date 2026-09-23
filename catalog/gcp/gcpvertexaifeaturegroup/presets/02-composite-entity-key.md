# Composite Entity Key

## Use Case

Features whose entity is a pair -- a merchant and a product, a user and a store -- read from a BigQuery table keyed by two columns, with one feature reading a column whose name carries its unit.

## When to Use

- Ranking and recommendation features keyed by more than one id
- Sources whose column names do not match the feature names models expect
- Feature groups other teams' online stores depend on (`PREVENT`)

## What This Creates

- A feature group `merchant_product_features` in `us-central1` over the literal source `my-gcp-project.features.merchant_product_daily` (the modules add `bq://`), keyed by `merchant_id` and `product_id`
- Features `click_rate` and `revenue`, the latter reading the `revenue_usd` column
- `deletionPolicy: PREVENT`, fanned to both features

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `bigQuery.inputUri` | a literal table | A `GcpBigQueryTable` reference when the table is in the same chart. |
| `bigQuery.entityIdColumns` | `merchant_id`, `product_id` | Your key columns. |
| `features[].versionColumnName` | `revenue_usd` on `revenue` | The column a feature reads when it is named differently. |
| `deletionPolicy` | `PREVENT` | `DELETE` for a throwaway group. |
