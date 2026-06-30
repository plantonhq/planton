# Atlas MongoDB Terraform Module

This Terraform module provisions and manages Atlas MongoDB clusters based on the Planton Atlas MongoDB API specification.

## Overview

This module creates a production-ready Atlas MongoDB cluster with the following capabilities:

- **Multi-Cloud Support**: Deploy on AWS, GCP, or Azure
- **Flexible Topologies**: Support for REPLICASET, SHARDED, and GEOSHARDED cluster types
- **High Availability**: Configurable electable and read-only nodes with priority-based failover
- **Backup**: Automated cloud backup configuration
- **Auto-Scaling**: Optional disk auto-scaling
- **Resource Tagging**: Metadata-driven labels for organization and cost tracking

## Prerequisites

- Terraform >= 1.0
- Atlas MongoDB account with programmatic API access
- Atlas MongoDB Project ID (create via Atlas UI or API)
- Atlas MongoDB API Keys (Public and Private)

### Creating Atlas MongoDB API Keys

1. Log in to Atlas MongoDB
2. Navigate to Organization or Project Settings
3. Go to "Access Manager" → "API Keys"
4. Click "Create API Key"
5. Select appropriate permissions (typically "Project Owner" for full cluster management)
6. Save the Public Key and Private Key securely

## Usage

### Basic Usage

The Planton CLI automatically manages this module. Users typically don't interact with it directly.

```bash
# Deploy using Planton CLI
planton apply -f manifest.yaml
```

### Direct Terraform Usage (Advanced)

If you need to use this module directly with Terraform:

```hcl
module "atlas_mongodb" {
  source = "./path/to/this/module"

  # Authentication credentials
  atlasmongodb_credential = {
    public_key  = var.atlas_public_key
    private_key = var.atlas_private_key
  }

  # Resource metadata
  metadata = {
    name = "production-mongodb"
    id   = "mdbatl-prod-001"
    org  = "mycompany"
    env  = "production"
    labels = {
      team    = "platform"
      project = "core-services"
    }
  }

  # Cluster specification
  spec = {
    cluster_config = {
      project_id                  = "64f1a2b3c4d5e6f7g8h9i0j1"
      cluster_type                = "REPLICASET"
      electable_nodes             = 3
      priority                    = 7
      read_only_nodes             = 0
      cloud_backup                = true
      auto_scaling_disk_gb_enabled = true
      mongo_db_major_version      = "7.0"
      provider_name               = "AWS"
      provider_instance_size_name = "M30"
    }
  }
}
```

## Input Variables

### Required Variables

#### `atlasmongodb_credential`

Atlas MongoDB API authentication credentials.

```hcl
type = object({
  public_key  = string  # Atlas Public API Key
  private_key = string  # Atlas Private API Key (sensitive)
})
```

#### `metadata`

Resource metadata for identification and organization.

```hcl
type = object({
  name    = string                  # Resource name (used as cluster name)
  id      = optional(string)        # Unique resource ID
  org     = optional(string)        # Organization identifier
  env     = optional(string)        # Environment (dev, staging, prod)
  labels  = optional(map(string))   # Additional labels
  tags    = optional(list(string))  # Resource tags
  version = optional(object({ id = string, message = string }))
})
```

#### `spec`

Atlas MongoDB cluster specification.

```hcl
type = object({
  cluster_config = object({
    project_id                  = string  # Atlas Project ID (required)
    cluster_type                = string  # REPLICASET, SHARDED, or GEOSHARDED
    electable_nodes             = number  # Number of electable nodes (3, 5, or 7)
    priority                    = number  # Election priority (1-7, primary region is 7)
    read_only_nodes             = number  # Number of read-only nodes
    cloud_backup                = bool    # Enable cloud backups
    auto_scaling_disk_gb_enabled = bool   # Enable disk auto-scaling
    mongo_db_major_version      = string  # MongoDB version (4.4, 5.0, 6.0, 7.0)
    provider_name               = string  # Cloud provider (AWS, GCP, AZURE)
    provider_instance_size_name = string  # Instance size (M10, M20, M30, etc.)
  })
})
```

## Outputs

### Required Outputs (Proto-Defined)

- `id` - Cluster ID assigned by Atlas MongoDB
- `bootstrap_endpoint` - Primary connection string (SRV format)
- `crn` - Cluster resource name (cluster ID)
- `rest_endpoint` - Standard connection string

### Additional Outputs

- `connection_strings` - Complete connection strings object (all formats)
- `connection_string_standard` - Standard connection string
- `connection_string_standard_srv` - Standard SRV connection string (recommended)
- `connection_string_private` - Private endpoint connection string (if configured)
- `connection_string_private_srv` - Private endpoint SRV connection string (if configured)
- `cluster_name` - Cluster name
- `cluster_type` - Cluster type
- `state_name` - Current cluster state (IDLE, CREATING, UPDATING, etc.)
- `mongo_db_version` - MongoDB version running on the cluster
- `project_id` - Atlas Project ID

## Configuration Examples

### Development Cluster (M10)

```yaml
apiVersion: atlas.planton.dev/v1
kind: AtlasMongodb
metadata:
  name: dev-mongodb
spec:
  clusterConfig:
    projectId: "your-project-id"
    clusterType: REPLICASET
    electableNodes: 3
    priority: 7
    readOnlyNodes: 0
    cloudBackup: false
    autoScalingDiskGbEnabled: true
    mongoDbMajorVersion: "7.0"
    providerName: AWS
    providerInstanceSizeName: M10
```

### Production Cluster (M30 with Read Replicas)

```yaml
apiVersion: atlas.planton.dev/v1
kind: AtlasMongodb
metadata:
  name: prod-mongodb
spec:
  clusterConfig:
    projectId: "your-project-id"
    clusterType: REPLICASET
    electableNodes: 5
    priority: 7
    readOnlyNodes: 2
    cloudBackup: true
    autoScalingDiskGbEnabled: true
    mongoDbMajorVersion: "7.0"
    providerName: AWS
    providerInstanceSizeName: M30
```

### Sharded Cluster (High Throughput)

```yaml
apiVersion: atlas.planton.dev/v1
kind: AtlasMongodb
metadata:
  name: sharded-mongodb
spec:
  clusterConfig:
    projectId: "your-project-id"
    clusterType: SHARDED
    electableNodes: 7
    priority: 7
    readOnlyNodes: 0
    cloudBackup: true
    autoScalingDiskGbEnabled: true
    mongoDbMajorVersion: "7.0"
    providerName: GCP
    providerInstanceSizeName: M50
```

## Cluster Sizing Guide

| Tier | RAM | vCPU | Storage | Use Case |
|------|-----|------|---------|----------|
| M10  | 2 GB | 2 | 10-32 GB | Development, Testing |
| M20  | 4 GB | 2 | 20-64 GB | Small Production |
| M30  | 8 GB | 2 | 40-128 GB | Production (Recommended Entry) |
| M40  | 16 GB | 4 | 80-256 GB | Medium Production |
| M50  | 32 GB | 8 | 160-512 GB | Large Production |
| M60  | 64 GB | 16 | 320-1024 GB | High-Load Production |

**Note:** M0, M2, M5 are shared tiers and not suitable for production workloads.

## Important Considerations

### Cluster Type Selection

- **REPLICASET**: Standard high-availability configuration. Use for most workloads.
- **SHARDED**: Horizontal scaling for high throughput. More complex to manage.
- **GEOSHARDED**: Global distribution with location-aware reads/writes. Requires M30+.

### Node Configuration

- **Electable Nodes**: Must be 3, 5, or 7 (odd numbers for election consensus)
- **Priority**: Primary region should have priority 7, other regions decrease by 1
- **Read-Only Nodes**: Scale read capacity without affecting write elections

### Backup Strategy

- **Cloud Backup**: Recommended for all production clusters
- **Snapshot Frequency**: Default is every 6 hours (configurable via Atlas UI)
- **Retention**: Default is 2 days (configurable)
- For point-in-time recovery, enable Continuous Cloud Backups in Atlas UI

### Security Best Practices

1. **Never commit API keys to version control**
2. Use environment variables or secure secret management (Vault, AWS Secrets Manager)
3. Configure network access via IP Access Lists or Private Endpoints
4. Create database users with least-privilege access
5. Enable encryption at rest (available M10+)
6. Use VPC Peering or Private Endpoints for production (not implemented in this basic module)

## Limitations

This module currently implements a basic single-region configuration. Advanced features not yet implemented:

- Multi-region replication specs
- VPC Peering
- Private Endpoints (AWS PrivateLink, Azure Private Link, GCP PSC)
- Custom backup schedules
- Analytics nodes
- Advanced cluster settings (pit_enabled, bi_connector, etc.)

For these advanced features, additional configuration will be needed.

## Troubleshooting

### Error: "project_id not found"

Ensure the Project ID is correct and your API keys have access to the project.

### Error: "Invalid cluster tier"

- M0, M2, M5 require the legacy `atlasmongodb_cluster` resource
- This module uses `atlasmongodb_advanced_cluster` for M10+ tiers
- Check `provider_instance_size_name` is M10 or higher

### Cluster stuck in CREATING state

- Atlas cluster provisioning takes 7-10 minutes
- Check Atlas UI for detailed status
- Verify your Atlas subscription has available capacity

### Connection string issues

- Connection strings are only available after cluster is IDLE
- Use `connection_string_standard_srv` for modern drivers
- Add database users and network access rules in Atlas UI

## Provider Documentation

- [Atlas MongoDB Terraform Provider](https://registry.terraform.io/providers/mongodb/atlasmongodb/latest/docs)
- [Atlas MongoDB Advanced Cluster Resource](https://registry.terraform.io/providers/mongodb/atlasmongodb/latest/docs/resources/advanced_cluster)
- [Atlas MongoDB Documentation](https://www.mongodb.com/docs/atlas/)

## Support

For issues specific to this module, contact your Planton administrator.

For Atlas MongoDB questions, refer to:
- [Atlas MongoDB Support](https://www.mongodb.com/cloud/atlas/support)
- [MongoDB Community Forums](https://www.mongodb.com/community/forums/)

