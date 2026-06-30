# Kubernetes Keycloak - Terraform Module

This Terraform module deploys Keycloak identity and access management on Kubernetes with production-ready configurations.

## Overview

The module provides a simplified interface for deploying Keycloak with:

- **Production-Ready Setup**: StatefulSet-based deployment (avoiding anti-patterns)
- **Resource Management**: Configurable CPU and memory allocations
- **Ingress Support**: Optional external access with DNS configuration
- **Day 2 Operations**: Built-in support for clustering and high availability

## Prerequisites

- Terraform >= 1.0
- Kubernetes cluster (1.19+)
- kubectl configured
- Sufficient cluster resources for Keycloak and PostgreSQL

## Required Providers

```hcl
terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.0"
    }
  }
}

provider "kubernetes" {
  # Configure kubernetes provider
}

provider "helm" {
  kubernetes {
    # Helm uses kubernetes provider configuration
  }
}
```

## Module Inputs

### `metadata` (Required)

Metadata for the Keycloak deployment.

```hcl
metadata = {
  name = "main-keycloak"
  id   = "unique-id"      # Optional
  org  = "my-org"         # Optional
  env  = "production"     # Optional
}
```

### `spec` (Required)

Specification for the Keycloak deployment.

```hcl
spec = {
  target_cluster = {
    cluster_name = "prod-gke-cluster"
  }
  namespace = {
    value = "keycloak-prod"
  }
  create_namespace = true
  container = {
    resources = {
      requests = {
        cpu    = "50m"
        memory = "100Mi"
      }
      limits = {
        cpu    = "1000m"
        memory = "1Gi"
      }
    }
  }
  ingress = {
    is_enabled = true
    dns_domain = "keycloak.example.com"
  }
}
```

#### Namespace Management

The `spec.namespace` and `spec.create_namespace` fields control namespace behavior:

```hcl
spec = {
  target_cluster = {
    cluster_name = "prod-gke-cluster"
  }
  namespace = {
    value = "keycloak-prod"
  }
  create_namespace = true  # Module creates namespace
  
  # ... other spec fields
}
```

**Namespace Control:**

- When `create_namespace = true`: Module creates the namespace
- When `create_namespace = false`: Namespace must already exist
- The namespace name is specified in `spec.namespace.value`

This allows flexibility in namespace lifecycle management, enabling use of either module-managed or externally-managed namespaces.

## Module Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Kubernetes namespace where Keycloak is deployed |
| `service` | Name of the Keycloak service |
| `port_forward_command` | Command to port-forward to Keycloak for local access |
| `kube_endpoint` | Internal Kubernetes endpoint |
| `external_hostname` | Public endpoint (if ingress enabled) |
| `internal_hostname` | Internal VPC endpoint (if ingress enabled) |

## Usage Example

### Basic Deployment

```hcl
module "keycloak" {
  source = "path/to/module"

  metadata = {
    name = "my-keycloak"
  }

  spec = {
    target_cluster = {
      cluster_name = "dev-gke-cluster"
    }
    namespace = {
      value = "keycloak-dev"
    }
    create_namespace = true
    container = {
      resources = {
        requests = {
          cpu    = "50m"
          memory = "100Mi"
        }
        limits = {
          cpu    = "1000m"
          memory = "1Gi"
        }
      }
    }
    ingress = {
      is_enabled = false
      dns_domain = ""
    }
  }
}

output "keycloak_namespace" {
  value = module.keycloak.namespace
}
```

### With Ingress Enabled

```hcl
module "keycloak_public" {
  source = "path/to/module"

  metadata = {
    name = "prod-keycloak"
    env  = "production"
  }

  spec = {
    target_cluster = {
      cluster_name = "prod-gke-cluster"
    }
    namespace = {
      value = "keycloak-prod"
    }
    create_namespace = true
    container = {
      resources = {
        requests = {
          cpu    = "200m"
          memory = "512Mi"
        }
        limits = {
          cpu    = "2000m"
          memory = "2Gi"
        }
      }
    }
    ingress = {
      is_enabled = true
      dns_domain = "auth.company.com"
    }
  }
}

output "keycloak_url" {
  value = module.keycloak_public.external_hostname
}
```

## Deployment Approach

This module follows the Keycloak Operator pattern by using production-ready configurations that avoid common anti-patterns:

### What This Module Does

1. **Creates Namespace**: Dedicated namespace for Keycloak and its dependencies
2. **Sets Up Configuration**: Prepares locals and outputs for Keycloak deployment
3. **Manages Resources**: Defines resource limits and requests

### Deployment Pattern

Per the research documentation, this module is designed to work with the Bitnami Helm chart which provides:

- **StatefulSet** (not Deployment) for proper stateful workload handling
- **PostgreSQL backend** with persistence
- **JDBC-ping** for Kubernetes-native clustering
- **Health probes** and readiness checks
- **Production security** defaults

This avoids the "split-brain" anti-pattern of using plain Deployment resources for Keycloak.

## Current Implementation Status

**Note**: This module currently creates the namespace and configuration foundation. The full Helm chart deployment would be added in the main.tf file to complete the implementation. See main.tf for details on the recommended Helm chart approach.

## Verification

After deployment, verify your Keycloak instance:

```bash
# Check namespace and pods
kubectl get pods -n keycloak-<name>

# Port-forward for local access
kubectl port-forward -n keycloak-<name> svc/keycloak-<name> 8080:8080

# Access Keycloak admin console
open http://localhost:8080
```

## Best Practices

1. **Resource Allocation**: Start with recommended defaults and adjust based on load
2. **Ingress Configuration**: Use ingress for production deployments
3. **High Availability**: Deploy multiple replicas for production
4. **Database**: Ensure PostgreSQL has sufficient storage and resources
5. **Security**: Use proper DNS and TLS configuration
6. **Monitoring**: Set up monitoring for Keycloak metrics

## Additional Resources

- [Keycloak Official Documentation](https://www.keycloak.org/documentation)
- [Bitnami Keycloak Helm Chart](https://github.com/bitnami/charts/tree/main/bitnami/keycloak)
- [Module Research Documentation](../docs/README.md)
- [Terraform Examples](examples.md)

