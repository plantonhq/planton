# Persistent High Throughput FSx Lustre

PERSISTENT_2 SSD file system with 2400 GiB and 1000 MB/s/TiB throughput. Designed for production ML training, HPC simulations, and video rendering workloads that demand maximum I/O performance.

## When to Use

- Distributed ML training with multiple GPU instances reading training data simultaneously
- HPC workloads (CFD, weather modeling, molecular dynamics) that need sustained high throughput
- Video rendering farms processing large media files in parallel
- Any production workload requiring data durability with maximum storage performance

## What It Configures

- **PERSISTENT_2** — Latest generation persistent storage with data replication within the AZ
- **2400 GiB SSD** — Sub-millisecond latency, suitable for random I/O patterns
- **1000 MB/s/TiB throughput** — Maximum available tier. Aggregate: ~2340 MB/s for 2400 GiB
- **LZ4 compression** — Reduces storage consumption; transparent to applications
- **Automatic backups** — 7-day retention, daily at 04:00 UTC
- **Tags copied to backups** — Consistent tagging for cost allocation
- **AUTOMATIC metadata IOPS** — Scales with storage capacity; upgrade to USER_PROVISIONED if you need more
- **Sunday maintenance** — Weekly maintenance window at Sunday 03:00 UTC

## What to Customize

- Replace placeholders: `<subnet-id>`, `<security-group-id>`
- Increase `storage_capacity_gib` in increments of 2400 GiB for more capacity and aggregate throughput
- Add `kms_key_id` for customer-managed KMS key (compliance, audit trail)
- Switch metadata to `USER_PROVISIONED` with explicit `iops` for workloads with heavy file creation/listing
- Add `log_configuration` for CloudWatch audit logging
- Adjust `automatic_backup_retention_days` (max 90) based on recovery requirements
