# OpenSearch

This preset creates a single-node OpenSearch 2.19 cluster with 100 GiB of storage and a weekend maintenance window -- a starting point for log analytics and full-text search.

## When to Use

- Centralized log storage and search
- Full-text search behind an application
- Dashboarding over semi-structured data (OpenSearch Dashboards ships with the cluster)

## Key Configuration Choices

- **OpenSearch 2.19** (`engine: opensearch`, `engineVersion: "2.19"`) -- DigitalOcean's default OpenSearch line at the time this preset was verified (2026-09-16; `3.3` and `3.6` were also offered). OpenSearch versions are major.minor here, and DigitalOcean rejects a version it no longer offers at create, so check `GET /v2/databases/options` if a deploy fails there. The cluster exposes OpenSearch Dashboards automatically; its connection details arrive as the `ui_*` stack outputs.
- **Single node** (`nodeCount: 1`) -- fine for logs you can re-ingest; OpenSearch on DigitalOcean scales to 15 nodes when the index must survive node loss.
- **Custom storage** (`storageGib: 100`) -- indexes are disk-hungry; storage can be grown later but never shrunk.
- **Maintenance window** (`maintenanceWindow`) -- pins updates to Saturday 03:00 UTC, away from weekday ingest peaks.

## Related Presets

- **04-kafka** -- A common upstream pipeline feeding OpenSearch
- **02-postgresql-dev** -- Use for relational data instead of search indexes
