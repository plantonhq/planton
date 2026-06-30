# Compliance SnapLock Volume

An ONTAP volume with SnapLock COMPLIANCE for immutable record retention. Files committed to this volume become Write Once Read Many (WORM) — they cannot be modified or deleted by anyone until their retention period expires.

## When to use

- SEC 17a-4(f) compliance for financial records
- HIPAA compliance for healthcare data
- FINRA regulatory record retention
- Any workload requiring tamper-proof, immutable storage

## Key settings

- **500 GB** initial size
- **SnapLock COMPLIANCE** — immutable, no escape hatch (not even AWS Support)
- **5-year default retention** (files without explicit retention get 5 years)
- **1-year minimum** / **10-year maximum** retention bounds
- **1-day autocommit** — files not modified for 24 hours are auto-committed to WORM
- **SNAPSHOT_ONLY** tiering — active data stays on SSD, only snapshots tier
