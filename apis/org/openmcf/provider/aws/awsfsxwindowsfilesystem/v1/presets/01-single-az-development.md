# Single-AZ Development FSx Windows

SINGLE_AZ_2 SSD file system with 32 GiB and 32 MB/s throughput joined to an AWS Managed Microsoft AD. The smallest and cheapest Windows file system configuration. Backups disabled for cost savings.

## When to Use

- Development and test environments for Windows-based applications
- Proof-of-concept deployments to validate SMB access patterns
- CI/CD pipelines needing a shared Windows file system temporarily
- Quick experiments where data loss is acceptable

## What It Configures

- **SINGLE_AZ_2** — Latest generation single-AZ deployment. Higher throughput ceiling than SINGLE_AZ_1
- **32 GiB SSD** — Minimum storage capacity. Sub-millisecond latency
- **32 MB/s throughput** — Minimum throughput tier. Sufficient for light workloads
- **AWS Managed AD** — Simplest AD integration. No credentials in the manifest
- **No backups** — `automatic_backup_retention_days: 0` disables daily backups

## What to Customize

- Replace placeholders: `subnet-0123456789abcdef0`, `sg-0123456789abcdef0`, `d-0123456789`
- Increase `storage_capacity_gib` (SSD: 32–65536 GiB) for more space
- Increase `throughput_capacity` (next tiers: 64, 128, 256) for faster I/O
- Set `automatic_backup_retention_days: 7` if you need backup protection
- Switch to `self_managed_active_directory` if using on-premises or EC2-hosted AD
- Add `aliases` for user-friendly DNS names
- Switch to `MULTI_AZ_1` when you need high availability
