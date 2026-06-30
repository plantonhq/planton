# Scratch Development FSx Lustre

SCRATCH_2 SSD file system with 1200 GiB — the smallest and cheapest Lustre configuration. No data replication, no backups. Temporary storage for fast processing.

## When to Use

- Development and test environments for HPC or ML workloads
- Short-lived data processing jobs (ETL, batch transforms, video rendering)
- CI/CD pipelines that need a high-performance scratch space
- Quick experiments where data loss is acceptable

## What It Configures

- **SCRATCH_2** — Temporary storage with no replication. Higher burst throughput than SCRATCH_1 (200 MB/s/TiB baseline, burst to 1300 MB/s/TiB)
- **1200 GiB SSD** — Minimum storage capacity. Sub-millisecond latency
- **No backups** — Scratch file systems do not support automatic backups
- **No S3 integration** — Add `import_path`/`export_path` if you need to load data from S3

## What to Customize

- Replace placeholders: `<subnet-id>`, `<security-group-id>`
- Increase `storage_capacity_gib` in increments of 2400 GiB (e.g., 1200, 3600, 6000)
- Add `import_path` and `export_path` for S3 data pipeline workflows
- Add `data_compression_type: LZ4` to reduce storage consumption for compressible data
- Switch to `PERSISTENT_2` when you need data durability or backups
