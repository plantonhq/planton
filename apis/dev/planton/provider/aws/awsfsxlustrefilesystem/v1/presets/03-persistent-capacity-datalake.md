# Persistent Capacity Data Lake FSx Lustre

PERSISTENT_1 HDD file system with 6000 GiB and 12 MB/s/TiB throughput. Optimized for large-capacity, cost-effective storage where sequential throughput matters more than latency.

## When to Use

- Data lake staging: land raw data from S3, process with Spark/Flink, export results
- Genomics pipelines processing large sequencing datasets (BAM, FASTQ files)
- Log analysis and aggregation workloads with large sequential reads
- Archive-tier Lustre for datasets that are accessed periodically but need POSIX semantics
- Any workload where cost per GiB is the primary concern and latency is secondary

## What It Configures

- **PERSISTENT_1** — Persistent storage with data replication within the AZ. Only deployment type supporting HDD
- **6000 GiB HDD** — Minimum capacity for HDD storage. Significantly cheaper per GiB than SSD
- **12 MB/s/TiB throughput** — Lower tier. Aggregate: ~70 MB/s for 6000 GiB. Use 40 MB/s/TiB if more throughput is needed
- **LZ4 compression** — Reduces effective storage usage; especially beneficial for text-heavy data (logs, CSVs, JSON)
- **Automatic backups** — 14-day retention, daily at 02:00 UTC
- **Tags copied to backups** — Consistent tagging for cost allocation
- **Monday maintenance** — Weekly maintenance window at Monday 05:00 UTC

## What to Customize

- Replace placeholders: `<subnet-id>`, `<security-group-id>`
- Increase `storage_capacity_gib` in increments of 6000 GiB (e.g., 6000, 12000, 18000)
- Switch `per_unit_storage_throughput` to 40 MB/s/TiB for higher sequential throughput
- Add `kms_key_id` for customer-managed KMS key
- Add `log_configuration` for CloudWatch audit logging
- For S3 integration on PERSISTENT deployments, create a separate data repository association resource referencing the `file_system_id` output
