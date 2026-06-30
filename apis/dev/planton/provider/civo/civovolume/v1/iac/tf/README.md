# Terraform Module for CivoVolume

## Overview

This Terraform module provisions Civo block storage volumes based on the `CivoVolumeSpec` specification. It creates volumes with specified name, region, and size configuration.

## Features

- **Volume Provisioning**: Creates Civo Volumes with validated parameters
- **Region Support**: Supports all Civo regions (LON1, NYC1, FRA1, PHX1, SIN1)
- **Size Configuration**: 1-16,000 GiB range
- **Labels Integration**: Automatically applies Planton labels for resource tracking
- **Output Mapping**: Maps Civo Volume attributes to stack outputs

## Requirements

- **Terraform**: >= 1.0
- **Civo Provider**: >= 1.0
- **Authentication**: `CIVO_TOKEN` environment variable

## Usage

### Basic Example

```hcl
module "civo_volume" {
  source = "./iac/tf"

  metadata = {
    name = "prod-db-data"
    id   = "civol-abc123"
    org  = "my-org"
    env  = "prod"
  }

  spec = {
    volume_name = "prod-db-data"
    region      = "LON1"
    size_gib    = 100
  }
}
```

### With Optional Fields

```hcl
module "civo_volume" {
  source = "./iac/tf"

  metadata = {
    name = "app-storage"
    id   = "civol-xyz789"
    org  = "acme"
    env  = "staging"
  }

  spec = {
    volume_name     = "app-storage"
    region          = "FRA1"
    size_gib        = 50
    filesystem_type = "EXT4"  # Informational only
    snapshot_id     = ""      # Not supported on public Civo
    tags            = ["env:staging", "app:web"]  # Informational only
  }
}
```

## Inputs

### `metadata`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | Yes | Resource name |
| `id` | `string` | No | Resource ID |
| `org` | `string` | No | Organization name |
| `env` | `string` | No | Environment name |
| `labels` | `map(string)` | No | Additional labels |
| `tags` | `list(string)` | No | Additional tags |

### `spec`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `volume_name` | `string` | Yes | Volume name (lowercase, alphanumeric + hyphens) |
| `region` | `string` | Yes | Civo region (LON1, NYC1, FRA1, etc.) |
| `size_gib` | `number` | Yes | Volume size in GiB (1-16000) |
| `filesystem_type` | `string` | No | Desired filesystem (NONE, EXT4, XFS) - **informational only** |
| `snapshot_id` | `string` | No | Snapshot ID to restore from - **not supported on public Civo** |
| `tags` | `list(string)` | No | Organizational tags - **not applied to Civo resource** |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `volume_id` | `string` | The unique identifier of the created volume |
| `attached_instance_id` | `string` | ID of the instance the volume is attached to (empty if unattached) |
| `device_path` | `string` | Device path on the instance (empty - not exposed by provider) |
| `volume_name` | `string` | The name of the volume |
| `size_gib` | `number` | The size of the volume in GiB |
| `region` | `string` | The region where the volume was created |

## Provider Configuration

The module uses the Civo provider. Configure it via environment variable:

```bash
export CIVO_TOKEN="your-civo-api-token"
terraform init
terraform plan
terraform apply
```

Or using Terraform variables:

```hcl
provider "civo" {
  token  = var.civo_token
  region = var.spec.region
}
```

## Limitations

### 1. Filesystem Formatting

**Limitation**: The Civo provider doesn't expose filesystem formatting during volume creation.

**Field Affected**: `spec.filesystem_type`

**Behavior**: Field is **informational only**. Volumes are created unformatted regardless of this setting.

**Workaround**: Format volumes manually after attachment:
```bash
ssh root@<instance-ip>
mkfs.ext4 /dev/vdb  # or mkfs.xfs
mkdir -p /data
mount /dev/vdb /data
```

### 2. Snapshots

**Limitation**: Snapshot functionality is not available on public Civo cloud (only CivoStack).

**Field Affected**: `spec.snapshot_id`

**Behavior**: If specified, the value is ignored. Volume is created empty.

**Workaround**: Implement application-level backups (`pg_dump`, `mysqldump`, `rsync`) to object storage.

### 3. Tags

**Limitation**: The Civo Volume provider doesn't support tags.

**Field Affected**: `spec.tags`

**Behavior**: Tags are **informational only**. Not applied to the Civo resource.

**Note**: Planton labels (in `locals.tf`) are tracked internally but also not applied to the Civo resource, as the provider doesn't support labels/tags.

### 4. Device Path

**Limitation**: The Civo provider doesn't expose the device path after attachment.

**Output Affected**: `device_path`

**Behavior**: Output is empty.

**Workaround**: SSH into the instance and identify the device using `lsblk` or `/dev/disk/by-id/`.

## Post-Deployment Steps

After Terraform creates the volume:

1. **Attach the volume**:
   ```bash
   civo volume attach <volume-name> <instance-id>
   ```

2. **Format the volume** (first-time):
   ```bash
   ssh root@<instance-ip>
   mkfs.ext4 /dev/vdb  # or mkfs.xfs
   ```

3. **Mount the volume**:
   ```bash
   mkdir -p /data
   mount /dev/vdb /data
   echo "/dev/vdb /data ext4 defaults,nofail 0 2" >> /etc/fstab
   ```

## Examples

### Development Volume (Small)

```hcl
module "dev_volume" {
  source = "./iac/tf"

  metadata = {
    name = "dev-test-vol"
    env  = "dev"
  }

  spec = {
    volume_name = "dev-test-vol"
    region      = "FRA1"
    size_gib    = 10
  }
}
```

### Production Database Volume (Large)

```hcl
module "prod_db_volume" {
  source = "./iac/tf"

  metadata = {
    name = "prod-db-data"
    id   = "civol-prod-001"
    org  = "acme-corp"
    env  = "prod"
  }

  spec = {
    volume_name     = "prod-db-data"
    region          = "LON1"
    size_gib        = 1000
    filesystem_type = "XFS"  # Informational (format manually)
    tags            = ["env:prod", "criticality:high", "backup:daily"]
  }
}
```

## Validation

Before applying, validate the configuration:

```bash
terraform fmt      # Format code
terraform validate # Validate syntax
terraform plan     # Preview changes
```

## Cleanup

To destroy the volume:

```bash
# Detach volume first (if attached)
civo volume detach <volume-id>

# Destroy with Terraform
terraform destroy
```

**Warning**: Volume deletion is **irreversible**. Ensure backups exist before destroying production volumes.

## Related Documentation

- **API Reference**: [../../README.md](../../README.md)
- **Examples**: [../../examples.md](../../examples.md)
- **Research**: [../../docs/README.md](../../docs/README.md)
- **Civo Provider Docs**: [registry.terraform.io/providers/civo/civo](https://registry.terraform.io/providers/civo/civo/latest/docs)
- **Civo API**: [civo.com/api/volumes](https://www.civo.com/api/volumes)

## Support

For issues or questions:
- **Planton**: [github.com/plantonhq/planton/issues](https://github.com/plantonhq/planton/issues)
- **Civo Support**: [civo.com/support](https://www.civo.com/support)

