# Preset: Single-AZ Production

**Use case**: Production workloads in a single availability zone where NFS performance, data compression, encryption, and daily backups are required.

## Configuration

- **Deployment type**: SINGLE_AZ_2 — latest generation with higher throughput ceiling
- **Storage**: 512 GiB — mid-range for production datasets
- **Throughput**: 640 MB/s — suitable for content serving, CI/CD, and application data
- **Compression**: ZSTD — best compression ratio, reduces storage costs 2-3x for compressible data
- **Record size**: 128 KiB (default) — balanced for mixed I/O patterns
- **NFS exports**: Open to all VPC clients with rw/crossmnt/no_root_squash
- **Encryption**: Customer-managed KMS key for compliance
- **Backups**: 7-day retention with daily 05:00 UTC window
- **Maintenance**: Sunday at 02:00 UTC

## When to use

- Production application data stores (CMS, web apps, analytics)
- CI/CD artifact storage and build caches
- Shared NFS storage for container workloads (EKS, ECS)
- Workloads that can tolerate single-AZ failure (stateless with rebuild capability)

## Cost considerations

SINGLE_AZ_2 costs roughly 40% less than MULTI_AZ_1 at equivalent specs. ZSTD compression can reduce effective storage costs by 50-70% for compressible data. The 7-day backup retention adds incremental storage cost for changed blocks only.
