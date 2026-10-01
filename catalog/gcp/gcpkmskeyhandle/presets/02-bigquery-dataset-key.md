# BigQuery Dataset Key

## Use Case

Get a default customer-managed key for a BigQuery dataset in the US multi-region; tables, models, and query results in the dataset use it.

## When to Use

- An analytics dataset that must use CMEK
- A dataset in a multi-region location

## What This Creates

- The Cloud KMS API on the project
- A key handle for `bigquery.googleapis.com/Dataset` in `us`, and through it an HSM key; reference `status.outputs.kms_key` from the dataset's `kmsKeyName`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us` | Match the dataset's location (`eu`, or a region). |
| `projectId` | `my-gcp-project` | Reference the dataset's `GcpProject`. |
