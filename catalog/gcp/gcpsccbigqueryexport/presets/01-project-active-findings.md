# Project Active Findings

## Use Case

Keep one project's active, unmuted findings in BigQuery for a team dashboard.

## When to Use

- A project team that reports on its own security posture
- Trend reports on open findings

## What This Creates

- The Security Command Center API on the project
- An export of active, unmuted findings to the dataset

Grant the export's `principal` output `roles/bigquery.dataEditor` on the dataset.

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `dataset` | `scc_findings` | Reference your `GcpBigQueryDataset`. |
| `filter` | active, unmuted | Empty to keep the full history. |
