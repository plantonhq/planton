# BigQuery Destination

## Use Case

The one destination profile every stream in a region shares when it writes into BigQuery. The profile has no settings; each stream decides its datasets.

## When to Use

- Any stream that lands data in BigQuery
- A platform team offering one BigQuery destination to every application team

## What This Creates

- A BigQuery destination profile in `us-central1`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us-central1` | The region your streams run in. |
