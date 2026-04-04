# GCP Cloud SQL Terraform Module

## Overview

This Terraform module provides automated deployment and management of Google Cloud SQL instances for MySQL and PostgreSQL databases. It handles instance provisioning, network configuration, high availability setup, backup configuration, and database flag management.

## Features

- **Multiple Database Engines**: Support for MySQL and PostgreSQL with configurable versions
- **Flexible Instance Tiers**: From shared-core to high-memory configurations
- **Storage Configuration**: Configurable SSD storage from 10GB to 65TB
- **Network Security**: Private IP via VPC peering and/or public IP with authorized network restrictions
- **High Availability**: Optional regional HA with automatic failover capabilities
- **Automated Backups**: Configurable backup schedules with customizable retention periods
- **Database Flags**: Custom database configuration flags for fine-tuning
- **Resource Labeling**: Automatic labeling for resource organization and cost tracking

## Prerequisites

- Terraform 1.0 or later
- GCP account with appropriate permissions
- GCP project with Cloud SQL API enabled
- For private IP: VPC network with service networking peering configured

## Usage

### Basic MySQL Instance

```hcl
module "mysql_instance" {
  source = "./path/to/gcpcloudsql/v1/iac/tf"

  metadata = {
    name = "mysql-db"
  }

  spec = {
    project_id       = "my-gcp-project"
    region           = "us-central1"
    database_engine  = "MYSQL"
    database_version = "MYSQL_8_0"
    tier             = "db-n1-standard-1"
    storage_gb       = 10
    root_password    = "SecurePassword123!"
  }
}
```

### PostgreSQL with High Availability

```hcl
module "postgres_ha" {
  source = "./path/to/gcpcloudsql/v1/iac/tf"

  metadata = {
    name = "postgres-ha"
    org  = "myorg"
    env  = "production"
  }

  spec = {
    project_id       = "my-gcp-project"
    region           = "us-central1"
    database_engine  = "POSTGRESQL"
    database_version = "POSTGRES_15"
    tier             = "db-n1-standard-2"
    storage_gb       = 50

    high_availability = {
      enabled = true
      zone    = "us-central1-b"
    }

    backup = {
      enabled        = true
      start_time     = "03:00"
      retention_days = 7
    }

    root_password = "SecurePassword123!"
  }
}
```

### MySQL with Private IP

```hcl
module "mysql_private" {
  source = "./path/to/gcpcloudsql/v1/iac/tf"

  metadata = {
    name = "mysql-private"
  }

  spec = {
    project_id       = "my-gcp-project"
    region           = "us-central1"
    database_engine  = "MYSQL"
    database_version = "MYSQL_8_0"
    tier             = "db-n1-standard-1"
    storage_gb       = 10

    network = {
      vpc_id             = "projects/my-project/global/networks/my-vpc"
      private_ip_enabled = true
    }

    root_password = "SecurePassword123!"
  }
}
```

### PostgreSQL with Authorized Networks

```hcl
module "postgres_public" {
  source = "./path/to/gcpcloudsql/v1/iac/tf"

  metadata = {
    name = "postgres-public"
  }

  spec = {
    project_id       = "my-gcp-project"
    region           = "us-central1"
    database_engine  = "POSTGRESQL"
    database_version = "POSTGRES_15"
    tier             = "db-n1-standard-1"
    storage_gb       = 10

    network = {
      authorized_networks = [
        "203.0.113.0/24",
        "198.51.100.0/24"
      ]
    }

    root_password = "SecurePassword123!"
  }
}
```

### Production Setup with Custom Flags

```hcl
module "postgres_production" {
  source = "./path/to/gcpcloudsql/v1/iac/tf"

  metadata = {
    name = "postgres-production"
    org  = "myorg"
    env  = "production"
  }

  spec = {
    project_id       = "my-gcp-project"
    region           = "us-central1"
    database_engine  = "POSTGRESQL"
    database_version = "POSTGRES_15"
    tier             = "db-n1-highmem-4"
    storage_gb       = 100

    network = {
      vpc_id             = "projects/my-project/global/networks/production-vpc"
      private_ip_enabled = true
    }

    high_availability = {
      enabled = true
      zone    = "us-central1-c"
    }

    backup = {
      enabled        = true
      start_time     = "02:00"
      retention_days = 30
    }

    database_flags = {
      max_connections = "200"
      shared_buffers  = "262144"
    }

    root_password = "ProductionSecurePassword123!"
  }
}
```

## Inputs

| Name | Description | Type | Required | Default |
|------|-------------|------|----------|---------|
| metadata | Resource metadata including name, id, org, env | object | Yes | - |
| spec | Cloud SQL instance specification | object | Yes | - |

### Spec Object

| Field | Description | Type | Required | Default |
|-------|-------------|------|----------|---------|
| project_id | GCP project ID | string | Yes | - |
| region | GCP region | string | Yes | - |
| database_engine | Database engine (MYSQL or POSTGRESQL) | string | Yes | - |
| database_version | Database version (e.g., MYSQL_8_0, POSTGRES_15) | string | Yes | - |
| tier | Instance machine type | string | Yes | - |
| storage_gb | Storage size in GB | number | Yes | - |
| root_password | Root password | string | Yes | - |
| network | Network configuration | object | No | {} |
| high_availability | HA configuration | object | No | {} |
| backup | Backup configuration | object | No | {} |
| database_flags | Custom database flags | map(string) | No | {} |

## Outputs

| Name | Description |
|------|-------------|
| instance_name | Name of the Cloud SQL instance |
| connection_name | Full connection name (project:region:instance) |
| private_ip | Private IP address (if enabled) |
| public_ip | Public IP address |
| self_link | GCP resource self link |

## Backend Configuration

This module is designed to work with local backend for development and testing. For production use, configure a remote backend:

```hcl
terraform {
  backend "gcs" {
    bucket = "my-terraform-state-bucket"
    prefix = "gcpcloudsql/production"
  }
}
```

## Deployment Commands

### Initialize Terraform
```shell
terraform init
```

### Plan Changes
```shell
terraform plan -var-file="variables.tfvars"
```

### Apply Configuration
```shell
terraform apply -var-file="variables.tfvars" -auto-approve
```

### Destroy Resources
```shell
terraform destroy -var-file="variables.tfvars" -auto-approve
```

## Security Considerations

- **Root Password**: Store passwords in a secure secrets management system, not in version control
- **Private IP**: Recommended for production workloads to avoid public internet exposure
- **Authorized Networks**: When using public IP, restrict access to known CIDR blocks
- **Encryption**: All data is encrypted at rest and in transit by default
- **IAM**: Use GCP IAM for access control to the Cloud SQL instance

## Cost Optimization

- **Instance Tier**: Choose appropriate tier based on workload requirements
- **Storage**: Start with minimum required storage and enable automatic increase
- **High Availability**: Only enable for production workloads requiring high uptime
- **Backup Retention**: Balance retention period with storage costs
- **Development Instances**: Use shared-core instances (db-f1-micro, db-g1-small) for non-production

## Troubleshooting

### Private IP Connection Issues
Ensure VPC peering is configured:
```shell
gcloud services vpc-peerings list --network=my-vpc
```

### Connection Timeouts
Check authorized networks configuration and firewall rules.

### Backup Failures
Verify backup configuration and check Cloud SQL instance logs in GCP Console.

## References

- [GCP Cloud SQL Documentation](https://cloud.google.com/sql/docs)
- [Terraform Google Provider](https://registry.terraform.io/providers/hashicorp/google/latest/docs)
- [MySQL Configuration Flags](https://cloud.google.com/sql/docs/mysql/flags)
- [PostgreSQL Configuration Flags](https://cloud.google.com/sql/docs/postgres/flags)

## Support

For support, please contact support@planton.ai.

## License

This project is licensed under the MIT License.

