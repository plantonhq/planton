# Managed Disaster Recovery

## Use Case

Capacity that survives a regional outage: an Enterprise Plus reservation in `us-central1` replicated to `us-east4`, ready to fail over together with replicated datasets.

## When to Use

- Business-critical reporting with a recovery-time objective
- Workloads paired with cross-region dataset replication

## What This Creates

- An Enterprise Plus reservation with a secondary replica in `us-east4`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `projectId`, assignee | `bq-admin-project`, `reporting-project` | Stand-ins: name your admin project and the project whose jobs use the slots, or reference GcpProjects. |
| `secondaryLocation` | `us-east4` | Pair it with the region your datasets replicate to. |
| `edition` | `ENTERPRISE_PLUS` | Required for managed disaster recovery. |
