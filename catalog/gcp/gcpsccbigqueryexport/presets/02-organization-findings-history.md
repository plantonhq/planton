# Organization Findings History

## Use Case

Export every finding in the organization -- every create and update -- into the security team's data lake.

## When to Use

- Organization-level Security Command Center activation
- Long-term posture reporting and audits across all projects

## What This Creates

- An organization export of every finding to the dataset, with destroy blocked

Grant the export's `principal` output `roles/bigquery.dataEditor` on the dataset, and set table expiration on the dataset to bound storage.

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `scope.organizationId` | `123456789012` | Your organization's numeric ID. |
| `filter` | every finding | Narrow to active findings to cut volume. |
