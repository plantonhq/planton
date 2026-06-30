# CivoVolume API

## Overview

The `CivoVolume` API provides a declarative way to provision and manage Civo Block Storage volumes using Planton. Civo Volumes are SSD-backed, network-attached block storage that can be attached to Civo Compute instances or dynamically provisioned by Civo Kubernetes clusters via the CSI driver.

This API abstracts the complexity of volume provisioning, attachment, and lifecycle management into a simple Protobuf-based specification. You declare what you want (volume name, size, region, filesystem type), and Planton handles the infrastructure provisioning via Pulumi or Terraform.

## API Structure

The `CivoVolume` resource follows the standard Planton API pattern:

```protobuf
message CivoVolume {
  string api_version = 1;                           // "civo.planton.dev/v1"
  string kind = 2;                                  // "CivoVolume"
  CloudResourceMetadata metadata = 3;               // Name, labels, description
  CivoVolumeSpec spec = 4;                          // Volume configuration
  CivoVolumeStatus status = 5;                      // Runtime outputs
}
```

## Specification Fields

### `CivoVolumeSpec`

Defines the desired state of your Civo block storage volume:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `volume_name` | `string` | Yes | Volume name (lowercase letters, numbers, hyphens; 1-64 chars) |
| `region` | `CivoRegion` | Yes | Civo region (e.g., `LON1`, `NYC1`, `FRA1`) |
| `size_gib` | `uint32` | Yes | Volume size in GiB (1-16,000) |
| `filesystem_type` | `CivoVolumeFilesystemType` | No | Desired filesystem (`NONE`, `EXT4`, `XFS`; default: `NONE`) |
| `snapshot_id` | `string` | No | Snapshot ID to create volume from (CivoStack only) |
| `tags` | `repeated string` | No | Organizational tags (must be unique) |

#### Volume Name Validation

The `volume_name` field enforces strict naming rules:

- **Length**: 1-64 characters
- **Characters**: Lowercase letters (a-z), numbers (0-9), hyphens (-)
- **Start**: Must start with a letter
- **End**: Must end with a letter or number
- **Pattern**: `^[a-z]([a-z0-9-]*[a-z0-9])?$`

**Valid Examples**:
- `db-data-volume`
- `prod-mysql-01`
- `app-storage-2024`

**Invalid Examples**:
- `DB-Volume` (uppercase)
- `db_volume` (underscores)
- `9db-volume` (starts with number)
- `-db-volume` (starts with hyphen)
- `db-volume-` (ends with hyphen)

#### Region Support

Civo Volumes are region-scoped and must be in the same region as any instance they attach to:

- `LON1` - London, United Kingdom
- `NYC1` - New York City, USA
- `FRA1` - Frankfurt, Germany
- `PHX1` - Phoenix, USA
- `SIN1` - Singapore

**Important**: Volumes cannot be moved between regions. To migrate data, create a new volume in the target region and copy data over the network or via a snapshot (CivoStack only).

#### Size Constraints

- **Minimum**: 1 GiB
- **Maximum**: 16,000 GiB (16 TiB)
- **Resizing**: Volumes can be expanded (offline only), but **never shrunk**
- **Cost**: Approximately $0.10 per GB-month

**Planning Tips**:
- Start smaller (you can always expand)
- A 500 GB volume costs ~$50/month
- Use multiple smaller volumes instead of one large volume for RAID0 striping if you need extreme IOPS

#### Filesystem Type

Specifies the desired filesystem format:

- `NONE` (default): Volume created unformatted (you format manually after attachment)
- `EXT4`: Recommended for most Linux workloads (stable, well-tested)
- `XFS`: Better for very large volumes (>1 TB) or high-concurrency workloads

**Important Limitation**: The Civo provider doesn't currently expose filesystem formatting during creation. This field is **informational only**. Volumes are created unformatted regardless of this setting. Use cloud-init or configuration management (Ansible, etc.) to format volumes after attachment.

#### Snapshot ID

Create a volume from an existing snapshot for data restoration or cloning:

**Limitation**: Snapshot functionality is **not available on public Civo cloud**. This parameter is reserved for CivoStack (private cloud) deployments. If specified on public Civo, it will be ignored with a warning.

**Alternative Backup Strategy**: Use application-level backups (e.g., `pg_dump`, `mysqldump`) or filesystem-level tools (`rsync`, `tar`) to export data to object storage.

#### Tags

Tags provide organizational metadata for volumes:

- Format: Letters, numbers, colons, dashes, underscores (max 64 chars per tag)
- Must be unique
- Use cases:
  - Environment tracking (`env:prod`, `env:staging`)
  - Team ownership (`team:backend`, `team:data`)
  - Compliance (`criticality:high`, `backup:daily`)

**Limitation**: The Civo Volume provider doesn't currently support tags. Tags in the spec are recorded in Planton metadata but not applied to the Civo resource. Use Civo labels (applied automatically by Planton) for resource organization in the Civo dashboard.

## Status and Outputs

### `CivoVolumeStatus`

After provisioning, the `status` field contains runtime information:

```protobuf
message CivoVolumeStatus {
  CivoVolumeStackOutputs outputs = 1;
}
```

### `CivoVolumeStackOutputs`

| Field | Type | Description |
|-------|------|-------------|
| `volume_id` | `string` | Unique identifier (UUID) for the volume |
| `attached_instance_id` | `string` | ID of the instance the volume is attached to (empty if unattached) |
| `device_path` | `string` | Device path on the instance (e.g., `/dev/vdb`) - available after attachment |

These outputs are exported after successful provisioning and can be consumed by applications or other infrastructure resources.

## Quick Start

### Minimal Example

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoVolume
metadata:
  name: test-data
spec:
  volumeName: test-data
  region: FRA1
  sizeGib: 10
```

This creates a 10 GiB unformatted volume in Frankfurt.

### Production Database Volume

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoVolume
metadata:
  name: prod-db-data
  description: Production PostgreSQL data volume
spec:
  volumeName: prod-db-data
  region: LON1
  sizeGib: 1000
  filesystemType: XFS
  tags:
    - env:prod
    - criticality:high
    - backup:daily
```

This creates a 1 TiB volume in London. Note: You must format it as XFS manually after attachment, as `filesystemType` is informational.

## Volume Lifecycle

### 1. Create

Provision a volume with a name, size, and region. The volume starts **unattached** and **unformatted**.

```bash
# Apply the CivoVolume manifest
planton apply -f volume.yaml
```

### 2. Attach

Attach the volume to a running Civo instance. This can be done via:

- **Civo Console**: Manually attach in the dashboard
- **Civo CLI**: `civo volume attach <volume-id> <instance-id>`
- **Terraform/Pulumi**: `civo_volume_attachment` resource (for advanced setups)
- **Kubernetes CSI**: Automatic attachment via `PersistentVolumeClaim` (K8s only)

After attachment, the OS sees a new block device (e.g., `/dev/vdb`, `/dev/sda`).

### 3. Format and Mount

**First-time setup** (after attachment):

```bash
# SSH into the instance
ssh root@<instance-ip>

# Identify the new device
lsblk
# Example output: vdb    8:16   0  10G  0 disk

# Format with ext4 (or xfs)
mkfs.ext4 /dev/vdb
# OR
mkfs.xfs /dev/vdb

# Create mount point
mkdir -p /data

# Mount the volume
mount /dev/vdb /data

# Verify
df -h /data

# Make mount persistent (add to /etc/fstab)
echo "/dev/vdb /data ext4 defaults,nofail 0 2" >> /etc/fstab
```

**Automation with Cloud-Init**:

```yaml
#cloud-config
runcmd:
  - mkfs.ext4 /dev/vdb
  - mkdir -p /data
  - mount /dev/vdb /data
  - echo "/dev/vdb /data ext4 defaults,nofail 0 2" >> /etc/fstab
```

### 4. Detach (Safe Removal)

Before detaching, **always unmount** to avoid data corruption:

```bash
# Unmount the volume
umount /data

# Remove from /etc/fstab
sed -i '/\/dev\/vdb/d' /etc/fstab

# Now safe to detach via Civo console or CLI
```

### 5. Resize (Expansion Only)

Volumes can be expanded (but **never shrunk**):

```bash
# 1. Detach the volume (or stop the instance)
civo volume detach <volume-id>

# 2. Resize via Civo (console or CLI)
civo volume resize <volume-id> --size 20

# 3. Reattach and expand the filesystem
mount /dev/vdb /data

# For ext4:
resize2fs /dev/vdb

# For xfs:
xfs_growfs /data

# Verify
df -h /data
```

**In Kubernetes**: Resizing is automatic if the CSI driver supports online expansion (it does for Civo).

### 6. Delete

Permanently destroy the volume (after detachment):

```bash
civo volume delete <volume-id>
```

**Warning**: Deletion is **irreversible**. Ensure backups exist before deleting production volumes.

## Kubernetes Integration

On Civo Kubernetes clusters, the `civo-volume` StorageClass uses the Civo CSI driver to provision volumes on-demand.

### Automatic Provisioning with PVC

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mysql-data
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: civo-volume
  resources:
    requests:
      storage: 20Gi
```

The CSI driver automatically:
1. Creates a 20 GiB Civo Volume
2. Attaches it to the node running the pod
3. Formats it (ext4 by default)
4. Mounts it into the container

**Use Case**: Perfect for StatefulSets (databases, Kafka, etc.).

### Manual Pre-Provisioning

For more control, pre-create the volume with Planton, then reference it in a PersistentVolume:

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: prod-db-volume
spec:
  capacity:
    storage: 100Gi
  accessModes:
    - ReadWriteOnce
  csi:
    driver: civo.csi.k8s.io
    volumeHandle: <volume-id-from-stack-outputs>
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: prod-db-data
spec:
  volumeName: prod-db-volume
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 100Gi
```

## Deployment Workflow

1. **Define the resource**: Create a YAML manifest with your volume specification
2. **Apply via Planton CLI**: Use `planton apply` to provision
3. **Pulumi/Terraform provisions the volume**:
   - Creates Civo Volume in specified region
   - Exports volume ID and metadata
4. **Attach to instance or use in Kubernetes** (manual or via CSI driver)
5. **Format and mount** (first-time setup)
6. **Consume outputs**: Applications retrieve volume ID and attachment info from status

## Performance Characteristics

- **Storage Type**: SSD-backed, network-attached
- **Network**: 10 Gbps
- **Performance Tier**: Single tier (no gp3 vs io2 complexity)
- **IOPS/Throughput**: Baseline suitable for most workloads (comparable to DigitalOcean Volumes or AWS gp3)

**For Extreme IOPS**: Use LVM or RAID0 to stripe multiple volumes.

**Volume vs. Instance Storage**:
- **Volumes**: Persistent data (databases, user uploads, logs). Data survives instance deletion.
- **Instance Storage**: Ephemeral data (OS, caches, temp files). Faster if local, but lost on instance termination.

## Best Practices

### Security

1. **Never expose raw block devices to untrusted users**: Use filesystem permissions to control access
2. **Encrypt sensitive data at the application layer**: Civo encrypts volumes at rest, but you control the keys
3. **Regularly backup critical data**: Use application-level backups (no native snapshots on public cloud)

### Cost Optimization

1. **Right-size**: Start with the minimum size you need (you can expand later)
2. **Delete orphaned volumes**: Volumes remain (and keep billing) after instance deletion. Audit regularly.
3. **Use IaC for tracking**: Terraform/Pulumi state files help remember what exists
4. **Compress data**: Databases like PostgreSQL support tablespace compression

### Data Protection

1. **Application-Level Backups**: Use `pg_dump`, `mysqldump`, or similar tools to export to object storage
2. **Filesystem Backups**: Use `rsync`, `tar`, or `duplicity` to copy volume data on a schedule
3. **Volume Cloning**: Manually attach both volumes to a helper instance and use `dd` or `rsync` to clone
4. **Test Restores**: Regularly verify backups are valid by performing test restores

## Limitations and Known Issues

1. **Snapshot Functionality**: Not available on public Civo cloud (only CivoStack)
2. **Filesystem Formatting**: Not exposed by Civo provider (must format manually after attachment)
3. **Tags**: Not supported by Civo Volume provider (recorded in metadata only)
4. **Region Migration**: Volumes cannot be moved between regions (must copy data manually)
5. **Multi-Attach**: Volumes can only attach to one instance at a time (no shared storage)
6. **Resizing**: Only expansion supported (no shrinking)
7. **Offline Resize**: Volume must be detached (or pods stopped) during resize

## Related Documentation

- **Examples**: See [examples.md](./examples.md) for real-world scenarios
- **Research**: See [docs/README.md](./docs/README.md) for deployment methods deep dive
- **IaC Implementation**: 
  - Pulumi: [iac/pulumi/](./iac/pulumi/)
  - Terraform: [iac/tf/](./iac/tf/)
- **Civo Docs**: [Civo Volumes Documentation](https://www.civo.com/docs/compute/instance-volumes)

## Support and Troubleshooting

### Common Issues

**Issue**: Volume not visible after attachment
- **Solution**: SSH into the instance and run `lsblk` to identify the device. It may be `/dev/vdb`, `/dev/vdc`, etc.

**Issue**: Mount fails with "wrong fs type" error
- **Solution**: Volume is unformatted. Run `mkfs.ext4 /dev/vdb` (or `mkfs.xfs`) before mounting.

**Issue**: Cannot detach volume (in use)
- **Solution**: Unmount the volume first: `umount /data`. Ensure no processes are accessing the mount point.

**Issue**: Resize doesn't increase filesystem size
- **Solution**: After resizing the volume, run `resize2fs /dev/vdb` (ext4) or `xfs_growfs /data` (xfs) to expand the filesystem.

**Issue**: Data loss after instance termination
- **Solution**: Volumes are independent of instances. Ensure you detach and preserve volumes before deleting instances. Check volume status in Civo dashboard.

### Getting Help

- **Planton Issues**: [GitHub Issues](https://github.com/plantonhq/planton/issues)
- **Civo Support**: [Civo Support Portal](https://www.civo.com/support)
- **Community**: [Planton Discussions](https://github.com/plantonhq/planton/discussions)

## Version History

- **v1**: Initial release with volume provisioning, region selection, size configuration, and filesystem type field

