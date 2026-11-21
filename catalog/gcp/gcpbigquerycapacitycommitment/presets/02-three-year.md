# Three-Year Commitment

## Use Case

The deepest discount for long-lived capacity: 500 Enterprise Plus slots for three years that end at the term (`renewalPlan: NONE`).

## When to Use

- Organizations centralizing BigQuery capacity in one admin project
- Stable, long-term analytics platforms

## What This Creates

- A three-year, 500-slot Enterprise Plus commitment that Google removes at the end of its term, with `ABANDON` on destroy and on any replacing change

Google's single-admin-project guard (`enforceSingleAdminProjectPerOrg`) is left out: the pinned provider never sends it to Google, so it would guard nothing, and changing it later replaces the commitment. Once Google removes the commitment at the end of the term, remove this block: a deploy of the unchanged manifest after the term buys a new one.

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `plan` | `THREE_YEAR` | Cannot be shortened later. |
| `projectId` | `bq-admin-project` | A stand-in: name your BigQuery admin project, or reference a GcpProject. |
| `renewalPlan` | `NONE` | `THREE_YEAR` renews it for another three years; keep `plan` equal to `renewalPlan`. |
