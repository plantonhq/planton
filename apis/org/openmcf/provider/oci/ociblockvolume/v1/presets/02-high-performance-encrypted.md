# High Performance Encrypted

This preset creates a 200 GB OCI Block Volume at the Higher Performance tier (20 VPUs/GB) with customer-managed KMS encryption, performance-based autotune that scales up to Ultra High Performance under load, a backup policy for scheduled snapshots, and a cross-region replica for disaster recovery. This is the configuration for database data files, transaction logs, and latency-sensitive application workloads.

## When to Use

- Database data files (Oracle, MySQL, PostgreSQL) where consistent low-latency I/O is critical
- Transaction log volumes where write latency directly impacts application throughput
- Latency-sensitive workloads that benefit from automatic performance scaling during peak load
- Production data volumes requiring customer-managed encryption for compliance and cross-region DR for business continuity

## Key Configuration Choices

- **200 GB at Higher Performance** (`sizeInGbs: 200`, `vpusPerGb: 20`) -- provides 75 IOPS/GB (15,000 total IOPS) and 600 KB/s/GB throughput. This tier offers a meaningful performance uplift over Balanced (60 IOPS/GB) for database workloads where every millisecond of I/O latency matters. Increase `sizeInGbs` for more aggregate IOPS (IOPS scales linearly with volume size).
- **Performance-based autotune** (`autotuneType: performance_based`, `maxVpusPerGb: 40`) -- OCI dynamically increases VPUs beyond the baseline 20 up to 40 when sustained workload demand is detected, and scales back down during quiet periods. This handles traffic spikes without manual intervention while keeping costs lower than permanently provisioning Ultra High Performance (30+ VPUs/GB).
- **KMS encryption** (`kmsKeyId`) -- encrypts the volume with a customer-managed AES key in OCI Vault. All data written to the volume is encrypted at rest, and the key can be rotated independently of the volume. Required by most compliance frameworks for database storage. Use an `OciKmsKey` created from the 01-aes-256-hsm-auto-rotation preset.
- **Backup policy** (`backupPolicyId`) -- same as the Balanced preset. For database volumes, Gold policy (daily + weekly + monthly + yearly) is recommended for maximum recovery point coverage.
- **Cross-region replica** (`blockVolumeReplicas`) -- asynchronously replicates the volume to a different availability domain (can be in a different region). In a disaster scenario, the replica can be promoted to a standalone volume in the target region. RPO depends on replication lag (typically minutes). For multi-region DR, place the replica in a different region's AD.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the volume will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<availability-domain>` | Availability domain for the source volume (e.g., `Uocm:US-ASHBURN-AD-1`) | `oci iam availability-domain list` CLI command, or OCI Console > region selector |
| `<kms-key-ocid>` | OCID of the KMS encryption key for the volume | `OciKmsKey` status outputs (`keyId`), or OCI Console > Identity & Security > Vault > Keys |
| `<backup-policy-ocid>` | OCID of the backup policy to assign (Gold recommended for databases) | `oci bv volume-backup-policy list` CLI command, or OCI Console > Block Storage > Backup Policies |
| `<replica-availability-domain>` | Availability domain for the DR replica (e.g., `Uocm:US-PHOENIX-AD-1`) | `oci iam availability-domain list` in the target region |

## Related Presets

- **01-balanced-with-backup** -- Use instead for general application data where Balanced performance is sufficient and cross-region DR is not needed
