# BigQuery Sink

## Use Case

Stream JSON messages from Kafka into BigQuery tables for analytics, one table per topic, created on first write.

## When to Use

- Analytics over event streams
- Landing Kafka data in a warehouse without a separate pipeline service

## What This Creates

- A BigQuery sink connector writing `orders` into the `orders` dataset

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `configs.defaultDataset` | `orders` | Grant the Managed Kafka service agent `roles/bigquery.dataEditor` on it. |
| `configs.autoCreateTables` | `true` | `false` when tables and schemas are managed elsewhere. |
