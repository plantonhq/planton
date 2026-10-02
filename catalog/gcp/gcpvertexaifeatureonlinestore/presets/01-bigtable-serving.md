# Bigtable Serving

## Use Case

Serve registered customer features to production models from a small autoscaled Bigtable store, refreshed every night from the feature group's BigQuery source.

## When to Use

- The first online store in a project
- Feature sets that grow large or see bursty traffic
- Features that change daily rather than by the second

## What This Creates

- An online store `serving_store` in `us-central1` on a managed Bigtable instance scaling between one and three nodes toward 60 percent CPU
- A feature view `customer_view` serving three features from the referenced `GcpVertexAiFeatureGroup`, synced at 02:00 New York time

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `bigtable.autoScaling.minNodeCount` | `1` | The committed floor; each node bills around the clock. |
| `bigtable.autoScaling.maxNodeCount` | `3` | The ceiling (at most ten times the floor). |
| `featureViews[].featureRegistrySource` | one group, three features | The features your model serves. |
| `featureViews[].syncConfig.cron` | nightly | More often for fresher features; `continuous: true` to stream. |
| `deletionPolicy` | `DELETE` | `PREVENT` once production traffic depends on the store. |
