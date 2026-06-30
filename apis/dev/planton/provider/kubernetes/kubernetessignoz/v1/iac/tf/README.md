# SigNoz Kubernetes Terraform Module

This Terraform module deploys SigNoz, an open-source OpenTelemetry-native observability platform, on Kubernetes.

## Overview

This module provides a complete infrastructure-as-code solution for deploying SigNoz with support for:

- **Dual Database Modes**: Self-managed ClickHouse or external ClickHouse integration
- **High Availability**: Distributed ClickHouse clusters with sharding and replication
- **Scalable Ingestion**: Independent scaling of OpenTelemetry Collector replicas
- **Production-Ready**: Full support for Zookeeper coordination, persistent storage, and ingress configuration
- **Flexible Configuration**: Comprehensive Helm values customization

## Prerequisites

- Terraform >= 1.0
- Kubernetes cluster with kubectl access
- Helm 3.x
- Default StorageClass configured in the cluster (for persistent volumes)

## Provider Requirements

```hcl
terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.35"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.9"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
  }
}
```

## Namespace Management

The Terraform module provides flexible namespace management through the `create_namespace` variable:

- **When `create_namespace` is `true`**: The module creates the specified namespace using the Kubernetes provider. This is the recommended approach for new deployments.

- **When `create_namespace` is `false`**: The module uses an existing namespace. The namespace must already exist on the cluster, or the deployment will fail. This is useful when:
  - The namespace is managed by a separate Terraform module or process
  - Multiple components share the same namespace
  - Namespace creation is controlled by organizational policies

**Implementation Details**:
- Uses `count` parameter for conditional resource creation: `count = var.spec.create_namespace ? 1 : 0`
- Namespace name is always referenced via `local.namespace` (not the resource reference)
- Dependencies are handled conditionally using ternary operators in `depends_on`

**Best Practice**: Set `create_namespace` to `true` for most deployments to let the module manage its own namespace. Only set it to `false` when you have specific requirements for external namespace management.

## Module Structure

The module creates the following resources:

1. **Kubernetes Namespace** (Conditional): Created only when `create_namespace` is `true`
2. **SigNoz Helm Release**: Deploys SigNoz with configured components:
   - SigNoz binary (UI, API, Ruler, Alertmanager)
   - OpenTelemetry Collector
   - ClickHouse (self-managed mode) or connection to external ClickHouse
   - Zookeeper (for distributed ClickHouse clusters)

## Key Features

### Database Flexibility

- **Self-Managed ClickHouse**: Deploy and manage ClickHouse within the cluster
  - Single-node or distributed cluster configurations
  - Configurable persistence, resources, and disk size
  - Optional clustering with sharding and replication
  - Zookeeper coordination for HA deployments

- **External ClickHouse**: Connect to an existing ClickHouse instance
  - Reduced operational overhead
  - Ideal for centralized database management
  - Support for secure (TLS) connections

### Component Scaling

- **SigNoz Container**: Scale UI and API pods independently
- **OpenTelemetry Collector**: Scale ingestion capacity based on telemetry volume
- **ClickHouse**: Horizontal scaling through sharding (self-managed mode)

### Resource Management

- Configurable CPU and memory requests/limits for all components
- Custom container images support
- Persistent volume configuration for stateful components

## Usage

### Basic Deployment (Evaluation)

```hcl
module "signoz" {
  source = "./path/to/module"

  metadata = {
    name = "my-signoz"
    org  = "myorg"
    env  = "dev"
  }

  spec = {
    namespace        = "signoz-dev"
    create_namespace = true

    signoz_container = {
      replicas = 1
      resources = {
        requests = { cpu = "200m", memory = "512Mi" }
        limits   = { cpu = "1000m", memory = "2Gi" }
      }
    }

    otel_collector_container = {
      replicas = 2
      resources = {
        requests = { cpu = "500m", memory = "1Gi" }
        limits   = { cpu = "2000m", memory = "4Gi" }
      }
    }

    database = {
      is_external = false
      managed_database = {
        container = {
          replicas               = 1
          persistence_enabled = true
          disk_size              = "20Gi"
          resources = {
            requests = { cpu = "500m", memory = "1Gi" }
            limits   = { cpu = "2000m", memory = "4Gi" }
          }
        }
        cluster = {
          is_enabled = false
        }
        zookeeper = {
          is_enabled = false
        }
      }
    }
  }
}
```

### Production Deployment (High Availability)

See `examples.md` for comprehensive production deployment examples.

## Outputs

The module exports the following outputs:

- `namespace`: The Kubernetes namespace where SigNoz is deployed
- `signoz_service`: The SigNoz UI/API service name
- `otel_collector_service`: The OpenTelemetry Collector service name
- `kube_endpoint`: Internal cluster endpoint for SigNoz UI
- `otel_collector_grpc_endpoint`: Internal gRPC endpoint for telemetry ingestion
- `otel_collector_http_endpoint`: Internal HTTP endpoint for telemetry ingestion
- `port_forward_command`: Command to port-forward to SigNoz UI
- `external_hostname`: External hostname (if ingress enabled)
- `internal_hostname`: Internal hostname (if ingress enabled)
- `otel_collector_external_grpc_hostname`: External gRPC hostname (if ingress enabled)
- `otel_collector_external_http_hostname`: External HTTP hostname (if ingress enabled)
- `clickhouse_endpoint`: ClickHouse endpoint (self-managed mode only)
- `clickhouse_username`: ClickHouse username (self-managed mode only)
- `clickhouse_password_secret_name`: Kubernetes secret name for ClickHouse password
- `clickhouse_password_secret_key`: Key within the secret for ClickHouse password

## Important Considerations

### Storage Management

- Ensure your cluster has a default StorageClass for dynamic provisioning
- Monitor persistent volume usage to prevent outages
- Disk size cannot be modified after creation (Kubernetes limitation)

### High Availability

- Enable ClickHouse clustering with 2+ shards and 2+ replicas for production
- Deploy Zookeeper with odd number of nodes (3 or 5) for quorum
- Configure pod anti-affinity to spread replicas across nodes

### External ClickHouse

When using external ClickHouse:
- Ensure the external instance is accessible from the cluster
- A Zookeeper instance must be available for distributed cluster support
- The ClickHouse cluster must be configured with a distributed cluster named "cluster" (or update `cluster_name`)
- Provided credentials must have necessary permissions to create databases and tables

### Ingress Configuration

- Separate ingress resources may be needed for gRPC and HTTP OTel Collector endpoints
- gRPC requires `nginx.ingress.kubernetes.io/backend-protocol: "GRPC"` annotation
- Configure TLS certificates through cert-manager or Helm values

## Architecture

SigNoz consists of four main components:

1. **SigNoz Binary**: Consolidated service containing UI, API server, Ruler, and Alertmanager
2. **OpenTelemetry Collector**: Data ingestion and processing gateway
3. **ClickHouse**: High-performance columnar database for telemetry storage
4. **Zookeeper**: Coordination service for distributed ClickHouse (optional)

## Links

- [SigNoz Documentation](https://signoz.io/docs/)
- [SigNoz Helm Chart](https://github.com/SigNoz/charts)
- [OpenTelemetry](https://opentelemetry.io/)
- [ClickHouse Documentation](https://clickhouse.com/docs/)

## Support

For issues and questions:
- SigNoz: [GitHub Issues](https://github.com/SigNoz/signoz/issues)
- This module: Contact the module maintainers

