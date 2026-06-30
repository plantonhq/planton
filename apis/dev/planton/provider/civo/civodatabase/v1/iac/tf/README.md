# Civo Database Terraform Module

This directory contains the Terraform implementation for provisioning Civo managed database instances.

## Overview

The Terraform module translates the `CivoDatabaseSpec` into Civo Database resources using the official `civo/civo` Terraform provider. It provides a declarative way to manage database lifecycle, configuration, and dependencies.

## Features

- **Database Engine Support**: MySQL and PostgreSQL with configurable versions
- **High Availability**: Support for up to 5 nodes (1 primary + 4 replicas)
- **Network Isolation**: Private network attachment for security
- **Firewall Integration**: Optional firewall rules for access control
- **Custom Storage**: Override default storage sizes
- **Tag Management**: Resource organization through tags
- **Comprehensive Outputs**: Export connection details for downstream usage

## Module Structure

```
iac/tf/
├── variables.tf    # Input variable definitions
├── locals.tf       # Local value computations
├── main.tf         # Database resource definition
├── outputs.tf      # Output value definitions
├── provider.tf     # Provider configuration
├── examples.md     # Usage examples
└── README.md       # This file
```

## Prerequisites

- Terraform >= 1.0
- Civo account with API access
- Civo API token (set via `CIVO_TOKEN` environment variable)
- Existing Civo network (private network)
- (Optional) Civo firewall for access control

## Quick Start

### Basic Usage

```hcl
module "database" {
  source = "path/to/civodatabase/v1/iac/tf"

  metadata = {
    name = "my-database"
    env  = "production"
  }

  spec = {
    db_instance_name = "prod-db"
    engine           = "postgres"
    engine_version   = "16"
    region           = "lon1"
    size_slug        = "g3.db.medium"
    network_id = {
      value = "net-12345678-abcd-1234-abcd-1234567890ab"
    }
  }
}
```

### With High Availability and Firewall

```hcl
module "database_ha" {
  source = "path/to/civodatabase/v1/iac/tf"

  metadata = {
    name = "prod-database"
    env  = "production"
    org  = "acme-corp"
  }

  spec = {
    db_instance_name = "production-db"
    engine           = "postgres"
    engine_version   = "16"
    region           = "lon1"
    size_slug        = "g3.db.large"
    replicas         = 2  # 3 total nodes (1 primary + 2 replicas)
    
    network_id = {
      value = "net-12345678-abcd-1234-abcd-1234567890ab"
    }
    
    firewall_ids = [
      {
        value = "fw-87654321-dcba-4321-dcba-0987654321ba"
      }
    ]
    
    storage_gib = 200
    
    tags = ["production", "backend", "critical"]
  }
}
```

## Input Variables

### `metadata`

Metadata for the resource, including name, environment, organization, and labels.

**Type**: `object`  
**Required**: Yes

**Schema**:
```hcl
metadata = {
  name    = string           # Required: Resource name
  id      = optional(string) # Optional: Resource ID
  org     = optional(string) # Optional: Organization
  env     = optional(string) # Optional: Environment (dev/staging/prod)
  labels  = optional(map(string)) # Optional: Key-value labels
  tags    = optional(list(string)) # Optional: Tags
  version = optional(object({ 
    id      = string
    message = string
  }))
}
```

**Example**:
```hcl
metadata = {
  name = "my-database"
  env  = "production"
  org  = "acme-corp"
  labels = {
    team = "backend"
    cost-center = "engineering"
  }
  tags = ["critical", "24x7"]
}
```

---

### `spec`

Database specification defining engine, configuration, and networking.

**Type**: `object`  
**Required**: Yes

**Schema**:
```hcl
spec = {
  # Required fields
  db_instance_name = string # Max 64 characters
  engine           = string # "mysql" or "postgres"
  engine_version   = string # Pattern: ^[0-9]+(\.[0-9]+)?$
  region           = string # Civo region (lon1, nyc1, fra1, phx1)
  size_slug        = string # Instance tier (g3.db.small, g3.db.medium, etc.)
  
  # Optional fields
  replicas    = optional(number, 0) # 0-4 replicas (default: 0)
  storage_gib = optional(number)    # Custom storage in GiB
  tags        = optional(list(string), []) # Resource tags
  
  # Network configuration (required)
  network_id = object({
    value = optional(string)
    value_from = optional(object({
      kind       = string
      env        = optional(string)
      name       = string
      field_path = string
    }))
  })
  
  # Firewall configuration (optional)
  firewall_ids = optional(list(object({
    value = optional(string)
    value_from = optional(object({
      kind       = string
      env        = optional(string)
      name       = string
      field_path = string
    }))
  })), [])
}
```

**Validation Rules**:
- `db_instance_name`: Must not exceed 64 characters
- `engine`: Must be either "mysql" or "postgres"
- `engine_version`: Must match pattern `^[0-9]+(\.[0-9]+)?$` (e.g., "16", "8.0", "14.10")
- `replicas`: Must be between 0 and 4 (inclusive)

**Example**:
```hcl
spec = {
  db_instance_name = "production-db"
  engine           = "postgres"
  engine_version   = "16"
  region           = "lon1"
  size_slug        = "g3.db.large"
  replicas         = 2
  storage_gib      = 200
  
  network_id = {
    value = "net-12345678"
  }
  
  firewall_ids = [
    { value = "fw-87654321" }
  ]
  
  tags = ["production", "backend"]
}
```

---

## Outputs

The module exports the following outputs for downstream resource wiring:

### `database_id`

**Description**: The unique ID of the created Civo database  
**Type**: `string`  
**Example**: `db-abc123def456`

---

### `database_name`

**Description**: The name of the database instance  
**Type**: `string`  
**Example**: `production-db`

---

### `dns_endpoint`

**Description**: The DNS endpoint for connecting to the database (recommended for HA)  
**Type**: `string`  
**Example**: `db-abc123.civo.com`

**Note**: Always use this endpoint for HA configurations. It automatically updates during failover.

---

### `host`

**Description**: The hostname/endpoint of the database (static)  
**Type**: `string`  
**Example**: `10.0.0.50`

---

### `port`

**Description**: The port number for database connections  
**Type**: `number`  
**Example**: `5432` (PostgreSQL) or `3306` (MySQL)

---

### `username`

**Description**: The master username for database authentication  
**Type**: `string` (sensitive)  
**Example**: `civo`

---

### `password`

**Description**: The master password for database authentication  
**Type**: `string` (sensitive)  
**Example**: `SecurePassword123!`

---

### `status`

**Description**: The current status of the database instance  
**Type**: `string`  
**Example**: `Active`, `Building`, `Failed`

---

### `network_id`

**Description**: The ID of the private network the database is attached to  
**Type**: `string`  
**Example**: `net-12345678`

---

### `firewall_id`

**Description**: The ID of the firewall rule attached to the database  
**Type**: `string`  
**Example**: `fw-87654321`

---

### `nodes`

**Description**: The total number of nodes in the database cluster (primary + replicas)  
**Type**: `number`  
**Example**: `3` (1 primary + 2 replicas)

---

## Usage Examples

### Accessing Outputs

```hcl
# Get DNS endpoint for application configuration
output "database_endpoint" {
  value = module.database.dns_endpoint
}

# Get credentials (sensitive)
output "database_credentials" {
  value = {
    username = module.database.username
    password = module.database.password
  }
  sensitive = true
}
```

### Using Outputs in Kubernetes Secret

```hcl
resource "kubernetes_secret" "database_credentials" {
  metadata {
    name = "db-credentials"
  }

  data = {
    hostname = module.database.dns_endpoint
    port     = tostring(module.database.port)
    username = module.database.username
    password = module.database.password
    database = "myapp"
  }
}
```

### Chaining with Other Resources

```hcl
# Create network
resource "civo_network" "main" {
  label  = "app-network"
  region = "lon1"
}

# Create firewall
resource "civo_firewall" "database" {
  name       = "db-firewall"
  network_id = civo_network.main.id
  region     = "lon1"

  ingress_rule {
    label    = "allow-postgres"
    protocol = "tcp"
    port     = "5432"
    cidr     = civo_network.main.cidr
    action   = "allow"
  }
}

# Create database with references
module "database" {
  source = "path/to/civodatabase/v1/iac/tf"

  metadata = {
    name = "app-database"
    env  = "production"
  }

  spec = {
    db_instance_name = "app-db"
    engine           = "postgres"
    engine_version   = "16"
    region           = "lon1"
    size_slug        = "g3.db.medium"
    replicas         = 1

    network_id = {
      value = civo_network.main.id
    }

    firewall_ids = [
      { value = civo_firewall.database.id }
    ]
  }
}
```

---

## Configuration Reference

### Available Regions

- `lon1` - London, United Kingdom
- `nyc1` - New York City, USA
- `fra1` - Frankfurt, Germany
- `phx1` - Phoenix, USA

### Available Instance Tiers

| Size Slug | vCPU | RAM | Storage | Price/month |
|-----------|------|-----|---------|-------------|
| `g3.db.small` | 2 | 4 GB | 40 GB | ~$43 |
| `g3.db.medium` | 4 | 8 GB | 80 GB | ~$87 |
| `g3.db.large` | 6 | 16 GB | 160 GB | ~$174 |
| `g3.db.xlarge` | 8 | 32 GB | 320 GB | ~$348 |
| `g3.db.2xlarge` | 10 | 64 GB | 640 GB | ~$695 |

### Supported Database Engines

- **MySQL**: Versions 5.7, 8.0
- **PostgreSQL**: Versions 12, 13, 14, 15, 16, 17 (beta)

---

## Best Practices

### 1. Always Use Private Networks

Deploy databases in a Civo network to prevent public internet access:

```hcl
spec = {
  # ... other fields ...
  network_id = {
    value = civo_network.main.id
  }
}
```

### 2. Apply Firewall Rules

Use firewalls with least-privilege access control:

```hcl
resource "civo_firewall" "database" {
  # Allow only from Kubernetes node CIDR
  ingress_rule {
    cidr = "10.0.0.0/16"  # Kubernetes node range
    port = "5432"
  }
}
```

### 3. Use DNS Endpoint for HA

For high availability setups, always use `dns_endpoint` in connection strings:

```hcl
connection_string = "postgresql://${module.database.username}:${module.database.password}@${module.database.dns_endpoint}:${module.database.port}/myapp"
```

### 4. Tag Resources

Apply consistent tags for organization and cost tracking:

```hcl
spec = {
  # ... other fields ...
  tags = ["production", "backend", "team-alpha"]
}
```

### 5. Right-Size Instances

Start with smaller tiers and scale up based on actual usage:

- **Development**: `g3.db.small` (no replicas)
- **Staging**: `g3.db.medium` (1 replica)
- **Production**: `g3.db.large` or higher (2+ replicas)

---

## Troubleshooting

### Issue: "Error: Resource Already Exists"

**Symptom**: Terraform fails with "database name already exists" error.

**Solution**: Civo database names must be unique per region. Either:
1. Change `db_instance_name` to a unique value
2. Import the existing database: `terraform import module.database.civo_database.this <database_id>`

---

### Issue: "Error: Invalid Network ID"

**Symptom**: Terraform fails with "network not found" error.

**Solution**: Verify the network exists and is in the same region as the database:
```bash
civo network ls --region=lon1
```

---

### Issue: "Error: Firewall Not Found"

**Symptom**: Terraform fails with "firewall not found" error.

**Solution**: Ensure the firewall exists and is associated with the correct network:
```bash
civo firewall ls --region=lon1
```

---

### Issue: "Error: Validation Failed"

**Symptom**: Terraform fails with input validation error.

**Common Causes**:
- `db_instance_name` exceeds 64 characters
- `engine` is not "mysql" or "postgres"
- `engine_version` doesn't match pattern (e.g., "16.2.1" instead of "16")
- `replicas` exceeds 4

**Solution**: Review validation rules in `variables.tf` and correct the input.

---

## Terraform Commands

### Initialize Module

```bash
terraform init
```

### Validate Configuration

```bash
terraform validate
```

### Plan Changes

```bash
terraform plan
```

### Apply Changes

```bash
terraform apply
```

### Destroy Resources

```bash
terraform destroy
```

### Show Outputs

```bash
terraform output
```

---

## State Management

### Remote State

Store Terraform state remotely for team collaboration and safety:

```hcl
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "civo-database/prod/terraform.tfstate"
    region = "us-east-1"
  }
}
```

### State Locking

Use state locking to prevent concurrent modifications:

```hcl
terraform {
  backend "s3" {
    # ... s3 config ...
    dynamodb_table = "terraform-lock"
  }
}
```

---

## Security Considerations

### Credential Management

1. **Never hardcode API tokens**: Use environment variables
   ```bash
   export CIVO_TOKEN="your-api-token"
   ```

2. **Store passwords securely**: Use Terraform Cloud or encrypted backends

3. **Mark outputs as sensitive**: Prevent credential leakage in logs
   ```hcl
   output "password" {
     value     = module.database.password
     sensitive = true
   }
   ```

### Network Security

1. **Always use private networks**: Never deploy databases without network isolation
2. **Restrict firewall rules**: Use least-privilege CIDR blocks
3. **Disable public access**: Civo databases in private networks are not publicly accessible

---

## Additional Resources

- **Civo Terraform Provider**: [registry.terraform.io/providers/civo/civo](https://registry.terraform.io/providers/civo/civo)
- **Civo API Documentation**: [civo.com/api/databases](https://www.civo.com/api/databases)
- **Examples**: See [`examples.md`](examples.md) for more usage examples
- **Parent README**: See [`../../README.md`](../../README.md) for component overview
- **Research Documentation**: See [`../../docs/README.md`](../../docs/README.md) for architectural patterns

