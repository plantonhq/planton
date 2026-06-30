# DigitalOcean Volume - Terraform Module

This Terraform module deploys DigitalOcean Block Storage Volumes from Planton's protobuf-defined manifests.

## Overview

The module translates `DigitalOceanVolumeSpec` manifests into DigitalOcean Volume resources using Terraform's DigitalOcean provider. It handles:

- Block storage volume provisioning
- Region placement
- Filesystem pre-formatting (ext4, XFS, or none)
- Volume creation from snapshots
- Tagging and metadata

## Prerequisites

### 1. Terraform

Install Terraform 1.5 or later:

```bash
# macOS
brew install terraform

# Linux
wget https://releases.hashicorp.com/terraform/1.6.0/terraform_1.6.0_linux_amd64.zip
unzip terraform_1.6.0_linux_amd64.zip
sudo mv terraform /usr/local/bin/

# Verify installation
terraform version
```

### 2. DigitalOcean API Token

Get your token from: https://cloud.digitalocean.com/account/api/tokens

```bash
export DIGITALOCEAN_TOKEN="your-api-token"
```

## Module Structure

```
iac/tf/
├── variables.tf    # Input variable definitions
├── provider.tf     # Terraform and provider configuration
├── locals.tf       # Local variables and computed values
├── main.tf         # Volume resource
├── outputs.tf      # Module outputs
└── README.md       # This file
```

## Input Variables

### Required Variables

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | object | Resource metadata (name, org, env, labels, tags) |
| `spec` | object | Volume specification (see below) |

### Spec Object Structure

```hcl
spec = {
  volume_name     = string           # Name (lowercase letters, numbers, hyphens)
  description     = string           # Optional description (max 100 chars)
  region          = string           # DigitalOcean region (e.g., "nyc3", "sfo3")
  size_gib        = number           # Size in GiB (1-16000)
  filesystem_type = string           # "NONE", "EXT4", or "XFS"
  snapshot_id     = string           # Optional: Create from snapshot
  tags            = list(string)     # Optional: Tags for organization
}
```

## Usage

### Example: Minimal Development Volume

Create a `terraform.tfvars` file:

```hcl
metadata = {
  name = "dev-test-volume"
}

spec = {
  volume_name     = "dev-test-volume"
  description     = ""
  region          = "nyc3"
  size_gib        = 10
  filesystem_type = "NONE"
  snapshot_id     = ""
  tags            = ["env:dev"]
}
```

Deploy:

```bash
terraform init
terraform plan
terraform apply
```

### Example: Production Database Volume with XFS

```hcl
metadata = {
  name = "prod-db-data"
  labels = {
    app = "postgres"
    environment = "production"
  }
}

spec = {
  volume_name     = "prod-pg-data"
  description     = "PostgreSQL production database data volume"
  region          = "sfo3"
  size_gib        = 500
  filesystem_type = "XFS"  # Optimal for databases
  snapshot_id     = ""
  tags            = ["env:prod", "service:postgres", "tier:database"]
}
```

### Example: Volume from Snapshot

```hcl
spec = {
  volume_name     = "restored-db-data"
  description     = "Database volume restored from nightly backup"
  region          = "nyc3"  # Must match snapshot region
  size_gib        = 600     # Must be >= snapshot size
  filesystem_type = "XFS"
  snapshot_id     = "123456789"  # Replace with actual snapshot ID
  tags            = ["env:prod", "restored:true"]
}
```

## Outputs

| Output | Description |
|--------|-------------|
| `volume_id` | Volume UUID |
| `volume_urn` | Volume uniform resource name |
| `volume_name` | Volume name |
| `filesystem_type` | Filesystem type (ext4, xfs, or empty) |
| `filesystem_label` | Filesystem label |
| `droplet_ids` | List of attached Droplet IDs |
| `outputs` | Complete outputs object for cross-stack references |

Access outputs:

```bash
terraform output volume_id
terraform output -json outputs
```

Reference in other modules:

```hcl
data "terraform_remote_state" "volume" {
  backend = "s3"
  config = {
    bucket = "my-terraform-state"
    key    = "digitalocean-volume/terraform.tfstate"
    region = "us-east-1"
  }
}

# Use volume ID in volume attachment
resource "digitalocean_volume_attachment" "attach" {
  volume_id  = data.terraform_remote_state.volume.outputs.outputs.volume_id
  droplet_id = digitalocean_droplet.web.id
}
```

## State Management

### Local State (Development)

For testing and development, use local state:

```bash
terraform init
terraform apply
```

State is stored in `terraform.tfstate`.

### Remote State (Production)

For production, use remote state with locking:

#### Option 1: DigitalOcean Spaces (S3-Compatible)

```hcl
# backend.tf
terraform {
  backend "s3" {
    endpoint                    = "nyc3.digitaloceanspaces.com"
    region                      = "us-east-1"  # Dummy value (required by provider)
    bucket                      = "my-terraform-state"
    key                         = "digitalocean-volume/terraform.tfstate"
    skip_credentials_validation = true
    skip_metadata_api_check     = true
  }
}
```

Configure credentials:

```bash
export AWS_ACCESS_KEY_ID="your-spaces-key"
export AWS_SECRET_ACCESS_KEY="your-spaces-secret"
```

Initialize:

```bash
terraform init
```

#### Option 2: Terraform Cloud

```hcl
# backend.tf
terraform {
  backend "remote" {
    organization = "my-org"
    workspaces {
      name = "digitalocean-volume-prod"
    }
  }
}
```

## Workflow

### Initial Deployment

```bash
# Initialize Terraform
terraform init

# Validate configuration
terraform validate

# Preview changes
terraform plan

# Apply changes
terraform apply

# Save outputs
terraform output -json > outputs.json
```

### Updates

```bash
# Modify terraform.tfvars or spec

# Preview changes
terraform plan

# Apply updates
terraform apply
```

### Destruction

```bash
# Preview destruction
terraform plan -destroy

# Destroy infrastructure
terraform destroy
```

## Best Practices

### 1. Always Use Pre-Formatted Filesystems

✅ **Recommended:**
```hcl
filesystem_type = "XFS"  # or "EXT4"
```

**Benefits:**
- Simpler mount scripts
- No risk of accidental formatting
- Consistent across deployments

❌ **Avoid:**
```hcl
filesystem_type = "NONE"
```

### 2. Choose the Right Filesystem

**Use XFS for:**
- Databases (PostgreSQL, MySQL, MongoDB)
- Large files (videos, backups)
- High I/O workloads

**Use ext4 for:**
- General-purpose applications
- Many small files
- Legacy compatibility needs

### 3. Plan for Region Constraints

- Volume and Droplet **must be in same region**
- Snapshots **cannot be copied between regions**
- For DR, use manual file-level replication or managed services

### 4. Use Descriptive Names and Tags

✅ **Good:**
```hcl
spec = {
  volume_name = "prod-postgres-data-2024"
  description = "Primary PostgreSQL data volume for production cluster"
  tags        = ["env:prod", "db:postgres", "cluster:primary", "backup:hourly"]
}
```

❌ **Avoid:**
```hcl
spec = {
  volume_name = "vol1"
}
```

### 5. Tag for Cost Allocation

```hcl
spec = {
  tags = [
    "env:prod",
    "project:customer-portal",
    "cost-center:engineering",
    "owner:platform-team"
  ]
}
```

Enables cost tracking and organization-wide reporting.

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

**Symptom:** Terraform succeeds but volume size unchanged in OS

**Cause:** Two-step resize process (API + filesystem)

**Solution:**

```bash
# Step 1: Terraform resizes at API level (automatic)
terraform apply

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

**Symptom:** Terraform wants to recreate volume

**Causes:**
- Manual changes via DigitalOcean console
- Volume modified by another tool

**Solution:**

```bash
# Refresh state
terraform refresh

# Import existing volume if needed
terraform import digitalocean_volume.this vol-abc123
```

### Provider Authentication Errors

**Symptoms:** "Invalid token" or "Unauthorized"

**Solution:**

```bash
# Verify token is set
echo $DIGITALOCEAN_TOKEN

# Test token with doctl
doctl auth init --access-token $DIGITALOCEAN_TOKEN
doctl account get
```

## Advanced Configuration

### Multiple Environments with Workspaces

```bash
# Create dev workspace
terraform workspace new dev

# Create prod workspace
terraform workspace new prod

# Switch to prod
terraform workspace select prod

# Apply with environment-specific variables
terraform apply -var-file="prod.tfvars"
```

### Volume Attachment

The module creates volumes but doesn't attach them. Use a separate attachment resource:

```hcl
resource "digitalocean_volume_attachment" "attach" {
  volume_id  = digitalocean_volume.this.id
  droplet_id = digitalocean_droplet.web.id
}
```

### Automated Snapshots

```bash
# Create snapshot schedule (external to Terraform)
# Add to cron:
0 2 * * * doctl compute volume-snapshot create \
  --volume-id $(terraform output -raw volume_id) \
  --snapshot-name "backup-$(date +\%Y-\%m-\%d)"
```

### Custom Backend Configuration

```bash
# Initialize with backend config file
terraform init -backend-config=backend-prod.hcl

# backend-prod.hcl
endpoint = "nyc3.digitaloceanspaces.com"
bucket   = "my-terraform-state"
key      = "prod/digitalocean-volume/terraform.tfstate"
```

## Integration with Planton

This Terraform module can be used standalone or as part of Planton's manifest-driven workflow:

```bash
# Convert manifest to Terraform variables
planton convert --manifest volume.yaml --output tfvars

# Deploy with Terraform
terraform apply -var-file=volume.tfvars
```

**Note:** Planton primarily uses Pulumi for deployments. This Terraform module provides an alternative for teams preferring Terraform.

## Validation

### Pre-Deployment

```bash
# Format code
terraform fmt -recursive

# Validate configuration
terraform validate

# Security scan
tfsec .

# Preview changes
terraform plan
```

### Post-Deployment

```bash
# Get volume ID
VOLUME_ID=$(terraform output -raw volume_id)

# Check volume status
doctl compute volume get $VOLUME_ID

# Verify volume exists
doctl compute volume list | grep $VOLUME_ID
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
- [ ] Remote state backend configured
- [ ] State locking enabled (Terraform Cloud or DynamoDB)

## Cost Estimation

DigitalOcean Volume pricing: **$0.10/GB/month**

**Cost Examples:**
- 10 GB volume = $1/month
- 100 GB volume = $10/month
- 500 GB volume = $50/month
- 1 TB volume = $100/month

Use `terraform-cost-estimation` or Infracost for detailed cost analysis:

```bash
# Install Infracost
brew install infracost

# Generate cost estimate
infracost breakdown --path .
```

## Next Steps

- Review [../../docs/README.md](../../docs/README.md) for architecture and best practices
- Check [../../examples.md](../../examples.md) for usage patterns
- See [../pulumi/README.md](../pulumi/README.md) for Pulumi alternative
- See [../../hack/manifest.yaml](../../hack/manifest.yaml) for test manifest

## Support

For issues or questions:
- Check [troubleshooting section](#troubleshooting)
- Review [Terraform DigitalOcean Provider docs](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/volume)
- Review [DigitalOcean Volumes docs](https://docs.digitalocean.com/products/volumes/)
- Open an issue in the Planton repository

