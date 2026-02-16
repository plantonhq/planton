# Multi-AZ High Availability FSx Windows

MULTI_AZ_1 SSD file system with 1000 GiB, 512 MB/s throughput, and automatic cross-AZ failover. Mission-critical configuration with DNS aliases, provisioned IOPS (100,000), full audit logging on both streams, and 30-day backup retention.

## When to Use

- Business-critical Windows file shares that cannot tolerate AZ-level downtime
- Finance, healthcare, or regulated industries requiring high availability and audit compliance
- Enterprise file servers being migrated from on-premises with DNS alias cutover
- Workloads needing consistent high IOPS independent of storage capacity

## What It Configures

- **MULTI_AZ_1** — Active/standby across two AZs. Automatic failover in under 30 seconds
- **1000 GiB SSD** — Sub-millisecond latency. Room for growth up to 65536 GiB
- **512 MB/s throughput** — High-performance tier. Can be scaled up after creation
- **AWS Managed AD** — Simplest HA setup. Automatic domain join
- **DNS aliases** — Two custom DNS names for user-friendly access and DFS namespace integration
- **100,000 provisioned IOPS** — USER_PROVISIONED mode for consistent performance under heavy load
- **Full audit logging** — Both file access and share access events logged at SUCCESS_AND_FAILURE level
- **Customer-managed KMS** — Encryption at rest with your own KMS key
- **30-day backup retention** — Daily automatic backups at 03:00 UTC with tags copied to backups
- **Sunday maintenance** — Weekly maintenance window at Sunday 02:00 UTC. MULTI_AZ maintains availability during maintenance via failover

## What to Customize

- Replace placeholders: two subnets (different AZs), preferred subnet, security group, KMS key ARN, AD directory ID
- Adjust `aliases` to match your DNS namespace (create CNAME records pointing to the `dns_name` output)
- Switch to `self_managed_active_directory` if using on-premises or EC2-hosted AD
- Adjust `disk_iops_configuration.iops` based on workload profiling (max 350,000)
- Increase `throughput_capacity` (next tiers: 1024, 2048, 4608) for extreme workloads
- Increase `storage_capacity_gib` as data grows
- Adjust `automatic_backup_retention_days` (max 90) based on compliance requirements
