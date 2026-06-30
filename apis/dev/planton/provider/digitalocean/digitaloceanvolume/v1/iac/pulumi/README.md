# DigitalOcean Volume - Pulumi Module

This Pulumi module deploys DigitalOcean Block Storage Volumes from Planton's protobuf-defined manifests.

## Overview

The module translates `DigitalOceanVolumeSpec` manifests into DigitalOcean Volume resources using Pulumi's DigitalOcean provider. It handles:

- Block storage volume provisioning
- Region placement
- Filesystem pre-formatting (ext4, XFS, or none)
- Volume creation from snapshots
- Tagging and metadata

## Architecture

See [overview.md](overview.md) for detailed module architecture and design decisions.

## Prerequisites

### 1. DigitalOcean Account and API Token

```bash
export DIGITALOCEAN_TOKEN="your-api-token"
```

Get your token from: https://cloud.digitalocean.com/account/api/tokens

### 2. Pulumi CLI

```bash
# macOS
brew install pulumi

# Linux
curl -fsSL https://get.pulumi.com | sh

# Windows
choco install pulumi
```

### 3. Go Runtime

The Pulumi program is written in Go. Ensure Go 1.21+ is installed:

```bash
go version
```

## Stack Configuration

### Required Configuration

Create a `Pulumi.<stack>.yaml` file:

```yaml
config:
  digitalocean:token: ${DIGITALOCEAN_TOKEN}
```

**Security Note:** Use environment variables or secret management for tokens. Never commit tokens to version control.

### Stack Input

The module expects a `DigitalOceanVolumeStackInput` JSON file:

```json
{
  "target": {
    "kind": "DigitalOceanVolume",
    "metadata": {
      "name": "prod-db-data"
    },
    "spec": {
      "volume_name": "prod-pg-data",
      "description": "PostgreSQL production database data volume",
      "region": "sfo3",
      "size_gib": 500,
      "filesystem_type": "XFS",
      "tags": ["env:prod", "service:postgres", "tier:database"]
    }
  },
  "provider_config": {
    "digitalocean_token": "${DIGITALOCEAN_TOKEN}"
  }
}
```

## Deployment Workflow

### Option 1: Using Planton CLI (Recommended)

```bash
# Create manifest
cat <<EOF > volume-manifest.yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanVolume
metadata:
  name: prod-db-data
spec:
  volume_name: prod-pg-data
  description: "PostgreSQL production database data volume"
  region: sfo3
  size_gib: 500
  filesystem_type: XFS
  tags:
    - env:prod
    - service:postgres
    - tier:database
EOF

# Deploy
planton pulumi up --manifest volume-manifest.yaml
```

### Option 2: Direct Pulumi Usage

```bash
# Initialize Pulumi stack
cd iac/pulumi
pulumi stack init prod

# Configure DigitalOcean token
pulumi config set digitalocean:token $DIGITALOCEAN_TOKEN --secret

# Set stack input (provide path to JSON file)
export STACK_INPUT_FILE="path/to/stack-input.json"

# Preview changes
pulumi preview

# Deploy
pulumi up
```

### Option 3: Using Makefile

```bash
cd iac/pulumi

# Deploy
make up STACK=prod

# Preview
make preview STACK=prod

# Destroy
make destroy STACK=prod
```

## Module Structure

```
iac/pulumi/
├── main.go                 # Pulumi program entrypoint
├── Pulumi.yaml             # Project configuration
├── Makefile                # Deployment automation
├── debug.sh                # Debugging helper script
├── README.md               # This file
├── overview.md             # Architecture documentation
└── module/
    ├── main.go             # Module orchestration
    ├── locals.go           # Local variables and labels
    ├── volume.go           # Volume resource logic
    └── outputs.go          # Stack output definitions
```

## Outputs

The module exports the following stack outputs:

```go
// Stack Outputs
outputs := {
  "volume_id": "vol-abc123"  // Volume UUID
}
```

Access outputs:

```bash
pulumi stack output volume_id
```

## Environment Variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| `DIGITALOCEAN_TOKEN` | DigitalOcean API token | Yes | `dop_v1_abc...` |
| `STACK_INPUT_FILE` | Path to stack input JSON | Yes | `./stack-input.json` |
| `PULUMI_BACKEND_URL` | Pulumi state backend | No | `s3://my-pulumi-state` |

## Advanced Configuration

### Using DigitalOcean Spaces for State

```bash
# Configure Spaces backend
pulumi login s3://my-pulumi-state?endpoint=nyc3.digitaloceanspaces.com

# Deploy
pulumi up
```

### Filesystem Type Selection

```yaml
spec:
  filesystem_type: XFS  # or EXT4, or NONE
```

**Filesystem Recommendations:**
- **XFS**: Databases (PostgreSQL, MySQL), large files, high I/O workloads
- **EXT4**: General purpose, many small files, legacy compatibility
- **NONE**: Manual formatting required (advanced use cases)

### Creating Volume from Snapshot

```yaml
spec:
  snapshot_id: "123456789"
  size_gib: 600  # Must be >= snapshot size
  region: nyc3   # Must match snapshot region
```

**Important Notes:**
- Volume must be in the same region as snapshot
- Volume must be at least as large as the snapshot
- Snapshots cannot be copied between regions

### Volume Tagging

```yaml
spec:
  tags:
    - env:prod
    - service:postgres
    - cost-center:engineering
    - backup:daily
```

Tags enable:
- Cost allocation and tracking
- Resource organization
- Automated discovery
- Compliance reporting

## Troubleshooting

### "Volume not attaching" Error

**Symptom:** Cannot attach volume to Droplet

**Causes:**
1. Region mismatch (volume and Droplet in different regions)
2. Volume already attached to another Droplet

**Solutions:**

```bash
# Check volume region
doctl compute volume get <volume_id> --format Region

# Check Droplet region
doctl compute droplet get <droplet_id> --format Region

# Detach from other Droplet if needed
doctl compute volume-action detach <volume_id> --droplet-id <old_droplet_id>
```

### "Filesystem not visible" Issue

**Symptom:** Volume attached but not showing in `df -h`

**Cause:** Volume attached but not mounted

**Solution:**

```bash
# List block devices
lsblk

# Find volume (look for DO_Volume_name)
ls -l /dev/disk/by-id/

# Mount volume
mkdir -p /mnt/data
mount /dev/disk/by-id/scsi-0DO_Volume_name /mnt/data

# Add to /etc/fstab for persistence
echo "/dev/disk/by-id/scsi-0DO_Volume_name /mnt/data xfs defaults,nofail,discard,noatime 0 0" >> /etc/fstab
```

### Volume Resize Not Working

**Symptom:** Pulumi succeeds but volume size unchanged in OS

**Cause:** Two-step resize process (API + filesystem)

**Solution:**

```bash
# Step 1: Pulumi resizes at API level (automatic)
pulumi up

# Step 2: Expand filesystem (manual)
ssh droplet-ip

# For XFS
sudo xfs_growfs /mnt/mount-point

# For ext4
sudo resize2fs /dev/disk/by-id/scsi-0DO_Volume_name

# Verify
df -h
```

### State Drift

**Symptom:** Pulumi wants to recreate volume

**Causes:**
- Manual changes via DigitalOcean console
- Volume modified by another tool

**Solution:**

```bash
# Refresh state
pulumi refresh

# Import existing volume if needed
pulumi import digitalocean:index/volume:Volume volume_name vol-abc123
```

### Debug Mode

Enable debug output:

```bash
# Use debug script
./debug.sh

# Or manually
export PULUMI_DEBUG_COMMANDS=true
export PULUMI_DEBUG_GRPC=true
pulumi up --logtostderr --logflow -v=9 2>&1 | tee pulumi-debug.log
```

## Validation

### Pre-Deployment Validation

```bash
# Validate manifest
planton validate --manifest volume-manifest.yaml

# Preview infrastructure changes
pulumi preview
```

### Post-Deployment Testing

```bash
# Get volume ID
VOLUME_ID=$(pulumi stack output volume_id)

# Check volume status
doctl compute volume get $VOLUME_ID

# Verify volume exists
doctl compute volume list | grep $VOLUME_ID
```

## Cleanup

### Destroy Volume

```bash
# Using Planton CLI
planton pulumi destroy --manifest volume-manifest.yaml

# Using Pulumi directly
pulumi destroy
```

**Warning:** Destroying a volume **deletes all data**. Ensure you have backups (snapshots) before destroying.

### Create Snapshot Before Destroy

```bash
# Create snapshot
VOLUME_ID=$(pulumi stack output volume_id)
doctl compute volume-snapshot create \
  --volume-id $VOLUME_ID \
  --snapshot-name "backup-before-destroy-$(date +%Y%m%d)"

# Wait for snapshot to complete
doctl compute volume-snapshot list

# Then destroy
pulumi destroy
```

### Remove Stack

```bash
pulumi stack rm prod
```

## Production Checklist

Before deploying to production:

- [ ] Correct region selected (matches Droplet region)
- [ ] Appropriate filesystem type chosen (XFS for databases, ext4 for general)
- [ ] Volume size adequate (with 30% growth buffer)
- [ ] Snapshot strategy defined (frequency, retention)
- [ ] Monitoring configured (disk usage alerts)
- [ ] Tags applied for cost tracking
- [ ] Disaster recovery plan documented
- [ ] Tested in staging environment

## Best Practices

### 1. Always Use Pre-Formatted Filesystems

✅ **Recommended:**
```yaml
filesystem_type: XFS  # or EXT4
```

**Benefits:**
- Simpler mount scripts
- No risk of accidental formatting
- Consistent across deployments

❌ **Avoid:**
```yaml
filesystem_type: NONE
```

### 2. Plan for Growth

```yaml
size_gib: 600  # Not 500 - leave 20% buffer for growth
```

### 3. Use Snapshots for Backups

```bash
# Automated daily snapshots
0 2 * * * doctl compute volume-snapshot create \
  --volume-id $VOLUME_ID \
  --snapshot-name "daily-backup-$(date +%Y-%m-%d)"
```

### 4. Monitor Disk Usage

```bash
# Install monitoring agent on Droplet
curl -sSL https://repos.insights.digitalocean.com/install.sh | sudo bash

# Set up alerts for >80% usage
```

### 5. Test Restore Process

```bash
# Periodically test snapshot restore
pulumi up -f test-restore-manifest.yaml
# Verify data integrity
# Destroy test volume
```

## Cost Optimization

DigitalOcean Volume pricing: **$0.10/GB/month**

**Cost Examples:**
- 10 GB volume = $1/month
- 100 GB volume = $10/month
- 500 GB volume = $50/month
- 1 TB volume = $100/month

**Optimization Tips:**
1. Right-size volumes (don't over-provision)
2. Delete unused snapshots
3. Use compression for applicable workloads
4. Consider object storage (Spaces) for archival data

## Next Steps

- Review [overview.md](overview.md) for module architecture
- Check [../../examples.md](../../examples.md) for usage patterns
- Read [../../docs/README.md](../../docs/README.md) for best practices
- See [../../hack/manifest.yaml](../../hack/manifest.yaml) for test manifest

## Support

For issues or questions:
- Check [troubleshooting section](#troubleshooting)
- Review [DigitalOcean Volumes docs](https://docs.digitalocean.com/products/volumes/)
- Open an issue in the Planton repository

