# Preset: Basic Analytics Dataset

## When to Use

Use this preset when you need a straightforward BigQuery dataset for analytics
workloads with default settings. This is the simplest configuration, suitable
for development, prototyping, or workloads without specific compliance requirements.

## What It Creates

- A BigQuery dataset in the US multi-region
- Google-managed encryption (default)
- 7-day time travel window (default)
- Default project-level access (owners/editors/viewers)
- Logical storage billing (default)

## Configuration

| Field | Value | Notes |
|-------|-------|-------|
| Location | US | Multi-regional for maximum availability |
| Encryption | Google-managed | Default, no CMEK |
| Time Travel | 168 hours (7 days) | Default maximum |
| Billing Model | LOGICAL | Default, charges per logical bytes |
| Access | Default | Project owners/editors/viewers |

## How to Use

1. Replace `<project-id>` with your GCP project ID
2. Replace `<your_dataset_id>` with a descriptive name (letters, numbers, underscores only)
3. Optionally change `location` to a specific region for data residency

## Downstream Usage

Reference this dataset's outputs from other resources:

```yaml
projectId:
  valueFrom:
    kind: GcpBigQueryDataset
    name: my-analytics-dataset
    fieldPath: status.outputs.project
```
