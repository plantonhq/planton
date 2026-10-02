# BigQuery Features

## Use Case

Register the customer features a team already computes in BigQuery -- age, tenure, recent orders -- so training pipelines and an online store read one definition of each.

## When to Use

- The first feature group in a project
- Features computed by a scheduled BigQuery job into one table
- Any source with a single entity ID column

## What This Creates

- A feature group `customer_features` in `us-central1` over the referenced `GcpBigQueryTable`, keyed by `customer_id`
- Three registered features: `age`, `tenure_days`, `orders_last_30d`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `featureGroupId` | `customer_features` | Your group's id (underscores, no hyphens). |
| `bigQuery.inputUri` | a `GcpBigQueryTable` reference | Your source; it needs a `feature_timestamp` TIMESTAMP column. |
| `bigQuery.entityIdColumns` | `customer_id` | The column(s) that identify an entity. |
| `features` | three features | The columns your models use. |
