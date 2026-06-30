# KubernetesGitlab Terraform Module

Terraform module for deploying GitLab on Kubernetes clusters.

## Overview

This Terraform module deploys GitLab to Kubernetes clusters. It provides flexible namespace management, creates necessary services, and optionally configures ingress resources for external access.

### Namespace Management

The module supports two namespace management strategies:

- **Automatic Creation** (`create_namespace = true`): The module creates a dedicated namespace with resource labels
- **Existing Namespace** (`create_namespace = false`): The module uses a pre-existing namespace that must be created beforehand

This flexibility allows you to manage namespaces according to your organization's requirements.

**Note:** This is a simplified implementation. For production GitLab deployments, we recommend using the official [GitLab Helm Chart](https://docs.gitlab.com/charts/) which provides comprehensive features including PostgreSQL, Redis, object storage integration, and GitLab Runner.

## Prerequisites

### Terraform Providers

This module requires the following providers:

```hcl
terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.23.0"
    }
  }
}
```

### Kubernetes Cluster Access

Configure the Kubernetes provider with access to your cluster:

```hcl
provider "kubernetes" {
  config_path = "~/.kube/config"
}
```

## Usage

### Basic Example (With Namespace Creation)

```hcl
module "gitlab" {
  source = "path/to/module"

  metadata = {
    name = "my-gitlab"
  }

  spec = {
    namespace        = "gitlab-instance"
    create_namespace = true  # Module creates the namespace

    container = {
      resources = {
        limits = {
          cpu    = "2000m"
          memory = "4Gi"
        }
        requests = {
          cpu    = "500m"
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
```

### Using Existing Namespace

```hcl
# First, create the namespace (or use an existing one)
resource "kubernetes_namespace" "shared" {
  metadata {
    name = "shared-services"
  }
}

module "gitlab_existing_ns" {
  source = "path/to/module"

  metadata = {
    name = "my-gitlab"
  }

  spec = {
    namespace        = "shared-services"
    create_namespace = false  # Use existing namespace

    container = {
      resources = {
        limits = {
          cpu    = "2000m"
          memory = "4Gi"
        }
        requests = {
          cpu    = "500m"
          memory = "1Gi"
        }
      }
    }

    ingress = {
      is_enabled = false
      dns_domain = ""
    }
  }

  depends_on = [kubernetes_namespace.shared]
}
```

### With Ingress

```hcl
module "gitlab_with_ingress" {
  source = "path/to/module"

  metadata = {
    name = "gitlab-prod"
    env  = "production"
  }

  spec = {
    container = {
      resources = {
        limits = {
          cpu    = "4000m"
          memory = "8Gi"
        }
        requests = {
          cpu    = "1000m"
          memory = "2Gi"
        }
      }
    }
    ingress = {
      enabled  = true
      hostname = "gitlab.example.com"
    }
  }
}
```

## Variables

### `metadata` (Required)

```hcl
metadata = {
  name = string           # Required: Resource name
  id = string            # Optional: Resource ID (defaults to name)
  org = string           # Optional: Organization
  env = string           # Optional: Environment
  labels = map(string)   # Optional: Additional labels
  tags = list(string)    # Optional: Tags
  version = object({     # Optional: Version info
    id = string
    message = string
  })
}
```

### `spec` (Required)

```hcl
spec = {
  namespace        = string  # Required: Namespace name
  create_namespace = bool    # Required: Whether to create namespace
  
  container = {              # Required
    resources = {
      limits = {
        cpu    = string      # e.g., "2000m"
        memory = string      # e.g., "4Gi"
      }
      requests = {
        cpu    = string      # e.g., "500m"
        memory = string      # e.g., "1Gi"
      }
    }
  }
  
  ingress = {               # Optional
    is_enabled = bool       # Enable external access
    hostname   = string     # External hostname
  }
}
```

**Namespace Management:**
- `namespace`: Name of the Kubernetes namespace for GitLab resources
- `create_namespace`: 
  - `true`: Module creates the namespace with appropriate labels
  - `false`: Module uses existing namespace (must exist before applying)

## Outputs

| Name | Description |
|------|-------------|
| `namespace` | Namespace where GitLab is deployed |
| `service_name` | Kubernetes service name |
| `service_fqdn` | Fully qualified domain name of the service |
| `port_forward_command` | Command to port-forward to GitLab |
| `ingress_hostname` | External hostname (if ingress enabled) |
| `internal_endpoint` | Internal Kubernetes endpoint |
| `external_endpoint` | External HTTPS endpoint (if ingress enabled) |

## Example Output Usage

```hcl
output "gitlab_endpoint" {
  value = module.gitlab.external_endpoint != null ? module.gitlab.external_endpoint : module.gitlab.internal_endpoint
}

output "port_forward" {
  value = module.gitlab.port_forward_command
}
```

## Production Deployment

For production GitLab deployments, consider using the official Helm chart with this module as a base:

```hcl
resource "helm_release" "gitlab" {
  name       = var.metadata.name
  namespace  = module.gitlab.namespace
  repository = "https://charts.gitlab.io/"
  chart      = "gitlab"
  version    = "7.0.0"

  values = [
    templatefile("${path.module}/gitlab-values.yaml", {
      hostname  = var.spec.ingress.hostname
      resources = var.spec.container.resources
    })
  ]

  depends_on = [module.gitlab]
}
```

## Resources Created

This module creates the following Kubernetes resources:

- **Namespace** (conditional): Created when `create_namespace = true`
  - Isolated namespace for GitLab with resource labels for tracking
- **Data Source** (conditional): References existing namespace when `create_namespace = false`
- **Service**: ClusterIP service for internal access
- **Ingress** (optional): External access with TLS when `ingress.is_enabled = true`

### Conditional Resource Creation

The module uses Terraform's `count` parameter to conditionally create resources:

```hcl
# Namespace is created only when create_namespace = true
resource "kubernetes_namespace" "gitlab" {
  count = var.spec.create_namespace ? 1 : 0
  # ...
}

# Existing namespace is referenced when create_namespace = false
data "kubernetes_namespace" "existing" {
  count = var.spec.create_namespace ? 0 : 1
  # ...
}
```

This ensures that only the appropriate resources are created based on your configuration.

## Networking

### Internal Access

GitLab is accessible within the cluster at:

```
http://<service-name>.<namespace>.svc.cluster.local:80
```

### External Access

When ingress is enabled, GitLab is accessible at:

```
https://<ingress-hostname>
```

TLS certificates are automatically provisioned using cert-manager.

## Port Forwarding

For local development and debugging:

```bash
kubectl port-forward -n <namespace> svc/<service-name> 80:80
```

Then access at `http://localhost:80`

## Troubleshooting

### Namespace Issues

**Error: Namespace already exists**

If you see an error about the namespace already existing:
- Set `create_namespace = false` to use the existing namespace
- Or delete the existing namespace if you want the module to create it

**Error: Namespace not found**

If you see an error about namespace not found when `create_namespace = false`:
- Create the namespace before applying Terraform
- Or set `create_namespace = true` to let the module create it

List namespaces:

```bash
kubectl get namespaces
```

Create namespace manually if needed:

```bash
kubectl create namespace <namespace-name>
```

### Service Issues

Check service status:

```bash
kubectl get svc -n <namespace>
kubectl describe svc <service-name> -n <namespace>
```

### Ingress Issues

Check ingress configuration:

```bash
kubectl get ingress -n <namespace>
kubectl describe ingress <ingress-name> -n <namespace>
```

Verify cert-manager certificate:

```bash
kubectl get certificate -n <namespace>
```

## Limitations

This is a simplified implementation focused on basic deployment. For production use:

- Use the official GitLab Helm chart for complete features
- Configure external PostgreSQL and Redis
- Set up object storage (S3, GCS, or Azure Blob)
- Configure GitLab Runner for CI/CD
- Enable monitoring and logging
- Configure backups and disaster recovery

## Reference

- [GitLab Charts Documentation](https://docs.gitlab.com/charts/)
- [GitLab Kubernetes Deployment Guide](https://docs.gitlab.com/ee/install/kubernetes/)
- [Terraform Kubernetes Provider](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs)

