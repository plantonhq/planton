# Three-Year Commitment

## Use Case

The deepest discount for long-lived capacity: 500 Enterprise Plus slots for three years, refused if another project in the organization already holds a commitment.

## When to Use

- Organizations centralizing BigQuery capacity in one admin project
- Stable, long-term analytics platforms

## What This Creates

- A three-year, 500-slot Enterprise Plus commitment with the single-admin-project guard

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `plan` | `THREE_YEAR` | Cannot be shortened later. |
| `enforceSingleAdminProjectPerOrg` | `true` | Keeps every commitment in one admin project. |
