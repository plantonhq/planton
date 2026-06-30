# DigitalOcean Database Cluster - Terraform Module

This Terraform module provisions and manages fully-managed database clusters on DigitalOcean using the official `digitalocean` provider.

## Overview

The module implements Planton's `DigitalOceanDatabaseCluster` protobuf spec, providing a declarative interface for deploying PostgreSQL, MySQL, Redis, and MongoDB clusters.

## Module Structure

```
iac/tf/
├── variables.tf    # Input variable definitions
├── locals.tf       # Local value transformations
├── main.tf         # Database cluster resource
├── outputs.tf      # Output exports
├── provider.tf     # Provider configuration
└── README.md       # This file
```

## Prerequisites

- **Terraform**: Version 1.0 or later
- **DigitalOcean API Token**: Set via environment variable `DIGITALOCEAN_TOKEN`
- **DigitalOcean Provider**: Version 2.x (auto-downloaded)

## Quick Start

### 1. Set Up Environment

```bash
cd apis/dev/planton/provider/digitalocean/digitaloceandatabasecluster/v1/iac/tf
export DIGITALOCEAN_TOKEN="your-digitalocean-api-token"
```

### 2. Create Configuration

**Development PostgreSQL:**
```hcl
module "dev_postgres" {
  source = "./path/to/module"

  metadata = {
    name = "dev-postgres"
    env  = "development"
  }

  spec = {
    cluster_name                = "dev-postgres"
    engine                      = "postgres"
    engine_version              = "16"
    region                      = "nyc3"
    size_slug                   = "db-s-1vcpu-1gb"
    node_count                  = 1
    enable_public_connectivity  = true
  }
}
```

**Production PostgreSQL (HA):**
```hcl
module "prod_postgres" {
  source = "./path/to/module"

  metadata = {
    name = "prod-postgres"
    org  = "acme-corp"
    env  = "production"
  }

  spec = {
    cluster_name               = "prod-postgres"
    engine                     = "postgres"
    engine_version             = "16"
    region                     = "nyc3"
    size_slug                  = "db-s-4vcpu-8gb"
    node_count                 = 3
    vpc = {
      value = "vpc-12345678"
    }
    storage_gib                = 200
    enable_public_connectivity = false
  }
}
```

### 3. Deploy

```bash
terraform init
terraform plan
terraform apply
```

### 4. Retrieve Outputs

```bash
terraform output cluster_id
terraform output host
terraform output -json | jq '.connection_uri.value'
```

## Module Variables

### `metadata` (Required)

Resource metadata.

### `spec` (Required)

Database cluster specification:

```hcl
spec = {
  cluster_name                = string  # Unique name (max 64 chars)
  engine                      = string  # "postgres", "mysql", "redis", "mongodb"
  engine_version              = string  # Major or major.minor (e.g., "16", "8.0")
  region                      = string  # DigitalOcean region
  size_slug                   = string  # Node size (db-s-2vcpu-4gb, etc.)
  node_count                  = number  # 1-3 nodes
  vpc                         = optional(object({ value = string }))
  storage_gib                 = optional(number)
  enable_public_connectivity  = optional(bool, false)
}
```

## Module Outputs

| Output | Type | Description |
|--------|------|-------------|
| `cluster_id` | string | DigitalOcean cluster UUID |
| `connection_uri` | string (sensitive) | Full connection string |
| `private_uri` | string (sensitive) | VPC-private connection string |
| `host` | string | Public hostname |
| `private_host` | string | VPC-private hostname |
| `port` | number | Database port |
| `database_name` | string | Default database name |
| `username` | string (sensitive) | Admin username |
| `password` | string (sensitive) | Admin password |

## Engine-Specific Notes

### PostgreSQL

**DigitalOcean API**: Uses "pg" as engine slug (module handles conversion).

**Critical Requirements:**
- ✅ Enable PgBouncer connection pool for production (separate resource)
- ✅ Connection limits are severe: 97 for 4GB RAM cluster

**Example with Connection Pool:**
```hcl
module "postgres_cluster" {
  # ... cluster config
}

resource "digitalocean_database_connection_pool" "app_pool" {
  cluster_id = module.postgres_cluster.cluster_id
  name       = "app-pool"
  mode       = "transaction"
  size       = 25
  db_name    = module.postgres_cluster.database_name
  user       = module.postgres_cluster.username
}
```

### MySQL

**Notes:**
- No SUPER privileges (cannot install plugins)
- Limited mysqldump access
- Use DigitalOcean backup system for migrations

### Redis

**Notes:**
- Redis 7+ is actually Valkey (SSPL licensing)
- No Redis Cluster mode (single-instance only)
- Use for caching, not primary storage

### MongoDB

**Notes:**
- Replica sets only (no sharding)
- Max 3 nodes
- Suitable for moderate-scale document stores

---

## Production Best Practices

### High Availability

```hcl
spec = {
  node_count = 3  # Primary + 2 standbys for maximum HA
}
```

### VPC Networking

```hcl
spec = {
  vpc = { value = var.vpc_uuid }
  enable_public_connectivity = false  # Force private connections
}
```

### Storage Planning

```hcl
spec = {
  storage_gib = 200  # Plan for 6-12 months growth
}
```

**Note**: Storage can only increase, never decrease.

### Firewall Configuration

```hcl
resource "digitalocean_database_firewall" "cluster" {
  cluster_id = module.cluster.cluster_id

  rule {
    type  = "vpc"
    value = var.vpc_uuid  # Allow entire VPC
  }

  rule {
    type  = "tag"
    value = "app-servers"  # Tag-based access
  }
}
```

---

## Troubleshooting

### Issue: "cluster name already exists"

**Solution**: Choose a unique name or delete existing cluster:
```bash
doctl databases list
doctl databases delete <cluster-uuid>
```

### Issue: Cluster stuck in "creating" state

**Cause**: Normal provisioning takes 10-15 minutes.

**Solution**: Wait for completion. Check status:
```bash
doctl databases get $(terraform output -raw cluster_id)
```

### Issue: Connection refused

**Cause**: Firewall rules not configured.

**Solution**: Add VPC CIDR or tags to firewall:
```hcl
resource "digitalocean_database_firewall" "allow_vpc" {
  cluster_id = module.cluster.cluster_id
  rule {
    type  = "vpc"
    value = "vpc-uuid-here"
  }
}
```

---

## Further Reading

- **Component Overview**: See [../../README.md](../../README.md)
- **Examples**: See [../../examples.md](../../examples.md)
- **Terraform Provider Docs**: [digitalocean_database_cluster](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/database_cluster)

