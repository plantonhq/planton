# Connect Workers

## Use Case

Kafka Connect workers for a team's pipelines: 3 vCPU and 12 GiB attached to the `events` cluster, on the same subnet as the brokers.

## When to Use

- Pub/Sub, BigQuery, or Cloud Storage sinks and sources
- A first Connect cluster before sizing for load

## What This Creates

- A 3-vCPU, 12 GiB Connect cluster for `events` with its workers on the `kafka` subnet

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `capacityConfig` | 3 vCPU, 12 GiB | Scale when connectors' tasks need more workers. |
| `networkConfigs` | `kafka` subnet | A private subnet in the region, reaching the systems connectors use. |
