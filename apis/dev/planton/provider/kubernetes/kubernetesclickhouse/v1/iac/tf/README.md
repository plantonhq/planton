# ClickHouse Kubernetes Terraform Module

## Overview

This Terraform module deploys production-grade ClickHouse clusters on Kubernetes using the **Altinity ClickHouse Operator**. The operator provides enterprise-level features including automated upgrades, scaling, backup, and recovery.

## Key Features

- **Operator-Based Deployment**: Uses the Altinity ClickHouse Operator for production-grade cluster management
- **CRD-Native**: Deploys ClickHouseInstallation custom resources that the operator reconciles
- **Flexible Namespace Management**: Conditionally creates and manages a namespace or deploys into existing namespaces based on the `create_namespace` flag
- **Persistence**: Configurable persistent volumes for data storage with flexible sizing
- **Clustering**: Optional sharding and replication for distributed deployments
- **ZooKeeper Integration**: Auto-managed ZooKeeper for cluster coordination, or configure external ZooKeeper
- **Security**: Auto-generated URL-safe passwords stored in Kubernetes Secrets
- **Ingress**: Optional LoadBalancer service with custom hostname for external access
- **Version Control**: Pin specific ClickHouse versions for stability
- **Production-Ready**: Leverages official ClickHouse images from clickhouse.com

## Prerequisites

The **Altinity ClickHouse Operator** must be installed on your Kubernetes cluster before using this module. Install the operator using the `ClickhouseOperatorKubernetes` module.

## Namespace Management

This module provides flexible namespace management through the `create_namespace` flag:

- **`create_namespace = true`**: The module creates and manages the namespace with appropriate labels
- **`create_namespace = false`**: Deploy into an existing namespace (namespace must exist before applying)

**When to use `create_namespace = true`:**
- Component should own the namespace lifecycle
- Dedicated namespace for this ClickHouse deployment
- Want automatic label management

**When to use `create_namespace = false`:**
- Namespace is managed by another Terraform module or tool
- Deploying multiple components into a shared namespace
- Namespace has special configuration (quotas, policies) managed externally

## Usage

### Basic Standalone Example

```hcl
module "clickhouse" {
  source = "./path/to/this/module"

  metadata = {
    name = "my-clickhouse"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    namespace        = "clickhouse-prod"
    create_namespace = true
    cluster_name     = "production-analytics"
    version          = "24.8"
    
    container = {
      replicas               = 1
      persistence_enabled = true
      disk_size              = "50Gi"
      resources = {
        requests = {
          cpu    = "500m"
          memory = "1Gi"
        }
        limits = {
          cpu    = "2000m"
          memory = "4Gi"
        }
      }
    }
    
    ingress = {
      enabled  = true
      hostname = "clickhouse.example.com"
    }
  }
}
```

### With Distributed Clustering

```hcl
module "clickhouse_cluster" {
  source = "./path/to/this/module"

  metadata = {
    name = "clickhouse-cluster"
  }

  spec = {
    container = {
      replicas               = 3
      persistence_enabled = true
      disk_size              = "100Gi"
      resources = {
        requests = {
          cpu    = "1000m"
          memory = "2Gi"
        }
        limits = {
          cpu    = "4000m"
          memory = "8Gi"
        }
      }
    }
    cluster = {
      is_enabled    = true
      shard_count   = 3
      replica_count = 2
    }
  }
}
```

## Requirements

- Terraform >= 1.0
- Kubernetes cluster with kubectl access
- **Altinity ClickHouse Operator** installed on the cluster

## Providers

| Name | Version |
|------|---------|
| kubernetes | ~> 2.35 |
| random | ~> 3.5 |

## Password Security

The module generates a 20-character random password using **URL-safe special characters only** (`-` and `_`).

**Why URL-safe characters matter:**

Characters like `+`, `=`, `/`, `&`, `?`, `#` cause problems when passwords are used in URL-encoded connection strings like `tcp://host:port/?password=XXX`. The `+` character is particularly problematic as it's decoded as a space by URL parsers.

See: https://github.com/Altinity/clickhouse-operator/issues/1883

## Resources Created

- `kubernetes_namespace_v1.clickhouse_namespace` - Dedicated namespace
- `random_password.clickhouse_password` - Auto-generated URL-safe password
- `kubernetes_secret_v1.clickhouse_password` - Password secret
- `kubernetes_manifest.clickhouse_installation` - ClickHouseInstallation CRD
- `kubernetes_service_v1.ingress_external_lb` - LoadBalancer (if ingress enabled)

The operator then reconciles the ClickHouseInstallation to create:
- StatefulSets for ClickHouse pods
- Services for cluster communication
- ConfigMaps with ClickHouse configuration
- ZooKeeper (if clustering is enabled)

## Outputs

| Name | Description |
|------|-------------|
| namespace | The namespace in which ClickHouse is deployed |
| service | The base name of the ClickHouse service |
| kube_endpoint | Internal DNS name of the ClickHouse service |
| port_forward_command | Command to port-forward to ClickHouse |
| external_hostname | External hostname (if ingress enabled) |
| username | ClickHouse username |
| password_secret_name | Name of the password Secret |
| password_secret_key | Key within the password Secret |

## Testing

```bash
# Initialize
terraform init

# Plan
terraform plan -var-file=hack/terraform.tfvars

# Apply
terraform apply -var-file=hack/terraform.tfvars -auto-approve

# Destroy
terraform destroy -var-file=hack/terraform.tfvars -auto-approve
```

## License

Apache 2.0
