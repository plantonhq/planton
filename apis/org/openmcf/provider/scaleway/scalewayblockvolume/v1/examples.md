# Scaleway Block Storage Volume Examples

## Table of Contents

1. [Minimal Volume (Development)](#1-minimal-volume-development)
2. [High-Performance Database Volume](#2-high-performance-database-volume)
3. [Volume from Snapshot (Clone/Restore)](#3-volume-from-snapshot-clonerestore)
4. [Large Archive Volume](#4-large-archive-volume)
5. [Multi-Environment Volumes](#5-multi-environment-volumes)

---

## 1. Minimal Volume (Development)

The simplest block volume: 10 GB with standard performance for development and testing.

**Use case:** Developer scratch space, CI/CD build artifacts, temporary data.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayBlockVolume
metadata:
  name: dev-scratch
  org: acme
  env: dev
spec:
  zone: fr-par-1
  size_gb: 10
  performance_tier: sbs_5k
```

**Deploy:**
```bash
openmcf pulumi up --manifest dev-scratch.yaml
```

**What gets created:**
- 1x `scaleway_block_volume` (10 GB, SBS 5K, fr-par-1)

**After attaching to an Instance:**
```bash
# Format the volume
sudo mkfs.ext4 /dev/vdb

# Mount it
sudo mkdir -p /mnt/scratch
sudo mount /dev/vdb /mnt/scratch

# Persist across reboots
echo '/dev/vdb /mnt/scratch ext4 defaults 0 2' | sudo tee -a /etc/fstab
```

---

## 2. High-Performance Database Volume

A 500 GB volume with 15,000 IOPS for production databases.

**Use case:** PostgreSQL, MySQL, MongoDB data directories where low latency and high throughput are critical.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayBlockVolume
metadata:
  name: prod-postgres-data
  org: acme
  env: production
spec:
  zone: fr-par-1
  size_gb: 500
  performance_tier: sbs_15k
```

**Notes:**
- Choose `sbs_15k` for database workloads -- the 3x IOPS improvement significantly reduces query latency.
- Ensure the target Instance has >= 3 GiB/s block bandwidth (check Instance type specs).
- Use XFS for large database volumes (better performance with large files):
  ```bash
  sudo mkfs.xfs /dev/vdb
  ```

---

## 3. Volume from Snapshot (Clone/Restore)

Create a volume from an existing Block Storage snapshot for disaster recovery or environment cloning.

**Use case:** Restore from backup, clone production data to staging, create test fixtures.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayBlockVolume
metadata:
  name: staging-db-clone
  org: acme
  env: staging
spec:
  zone: fr-par-1
  size_gb: 500
  performance_tier: sbs_5k
  snapshot_id: "fr-par-1/b5a6d3e8-1234-5678-9abc-def012345678"
```

**Notes:**
- The snapshot must be in the same zone as the target volume.
- `size_gb` must be >= the source volume's size.
- Staging can use `sbs_5k` (cheaper) even if production uses `sbs_15k`.
- The volume is immediately usable -- no need to re-format (filesystem is preserved from snapshot).

---

## 4. Large Archive Volume

A 5 TB volume for data archival, media storage, or log retention.

**Use case:** Application logs, media files, backup storage, data lake staging.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayBlockVolume
metadata:
  name: archive-storage
  org: acme
  env: production
spec:
  zone: nl-ams-1
  size_gb: 5120
  performance_tier: sbs_5k
```

**Notes:**
- Maximum volume size is 10,240 GB (10 TB).
- For archival data, `sbs_5k` is cost-effective -- IOPS are less important than capacity.
- Consider using XFS for volumes > 1 TB (better scalability than ext4 for large filesystems).

---

## 5. Multi-Environment Volumes

A pattern for creating volumes across environments with consistent naming and sizing.

### Development
```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayBlockVolume
metadata:
  name: app-data-dev
  org: acme
  env: dev
spec:
  zone: fr-par-1
  size_gb: 20
  performance_tier: sbs_5k
```

### Staging
```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayBlockVolume
metadata:
  name: app-data-staging
  org: acme
  env: staging
spec:
  zone: fr-par-1
  size_gb: 100
  performance_tier: sbs_5k
```

### Production
```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayBlockVolume
metadata:
  name: app-data-prod
  org: acme
  env: production
spec:
  zone: fr-par-1
  size_gb: 500
  performance_tier: sbs_15k
```

**Pattern notes:**
- Dev and staging use `sbs_5k` for cost savings.
- Production uses `sbs_15k` for performance.
- All environments use the same zone for consistency.
- Size scales with environment (20 -> 100 -> 500 GB).

---

## Advanced Patterns

### Volume Resize Workflow

To increase a volume's size:

1. **Update the manifest** with the new `size_gb` value (must be larger than current):
   ```yaml
   spec:
     size_gb: 1000  # was 500
   ```

2. **Apply the change:**
   ```bash
   openmcf pulumi up --manifest volume.yaml
   ```

3. **Grow the filesystem inside the OS** (while mounted):
   ```bash
   # For ext4
   sudo resize2fs /dev/vdb

   # For XFS
   sudo xfs_growfs /mnt/data
   ```

### Performance Tier Change

To change from `sbs_5k` to `sbs_15k` (or vice versa):

1. **Update the manifest:**
   ```yaml
   spec:
     performance_tier: sbs_15k  # was sbs_5k
   ```

2. **Apply the change:**
   ```bash
   openmcf pulumi up --manifest volume.yaml
   ```

The tier change is applied in-place -- no volume recreation, no data loss, no detach required.

---

## Best Practices

### Filesystem Choice
- **ext4**: General purpose, good default. Best for volumes < 1 TB.
- **XFS**: Better for large files, large volumes (> 1 TB), and database workloads.

### Naming Convention
Use a consistent naming pattern: `{purpose}-{environment}` (e.g., `postgres-data-prod`, `app-logs-staging`).

### Zone Strategy
Plan zones before creating volumes. A volume cannot be moved between zones -- you would need to snapshot, create in the new zone, and restore.

### Backup Strategy
Create regular snapshots using `scaleway_block_snapshot` (Terraform directly or future OpenMCF kind) for disaster recovery. Snapshots are incremental and cost-effective.

### Cost Optimization
- Use `sbs_5k` for non-latency-sensitive workloads (60-70% of use cases).
- Reserve `sbs_15k` for databases and real-time systems.
- Right-size volumes -- you can always increase, never decrease.

---

## Troubleshooting

### Volume Not Visible After Creation
Block volumes are zonal. Ensure you're checking the correct zone in the Scaleway console. The volume ID includes the zone prefix (e.g., `fr-par-1/...`).

### Cannot Attach to Instance
- Verify the volume and Instance are in the **same Availability Zone**.
- Verify the volume is in "available" state (not attached to another Instance).
- Check the Instance hasn't reached the 15-volume attachment limit.

### Terraform Plan Shows Destruction
If `terraform plan` shows the volume being destroyed and recreated, check:
- You haven't changed the `zone` (triggers recreation).
- You haven't decreased `size_gb` (the provider will error, but check the plan).

### Filesystem Not Growing After Resize
After increasing `size_gb` via IaC:
1. Verify the block device shows the new size: `lsblk`
2. If partitioned, grow the partition: `sudo growpart /dev/vdb 1`
3. Grow the filesystem: `sudo resize2fs /dev/vdb1` (ext4) or `sudo xfs_growfs /mnt/data` (XFS)

---

## References

- [Scaleway Block Storage Documentation](https://www.scaleway.com/en/docs/block-storage/)
- [Scaleway Block Storage API](https://www.scaleway.com/en/developers/api/block/)
- [Terraform scaleway_block_volume](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/block_volume)
