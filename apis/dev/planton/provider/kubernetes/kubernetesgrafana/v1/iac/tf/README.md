# KubernetesGrafana Terraform Module

This Terraform module deploys Grafana on a Kubernetes cluster.

## Overview

This module provisions:
- A Kubernetes namespace for Grafana (optionally created or referenced)
- Grafana deployment via Helm chart
- Optional Ingress resources for external and internal access
- Service resources for internal communication

### Namespace Management

The module provides flexible namespace management through the `create_namespace` flag:

- **`create_namespace: true`**: The module creates a dedicated namespace with appropriate resource labels for tracking and organization. All Grafana resources are deployed into this created namespace, and the namespace is managed as part of the component's lifecycle.

- **`create_namespace: false`** (default): The module uses an existing namespace specified in the `namespace` field. The namespace must already exist before deployment. This is useful when deploying multiple components into a shared namespace or when namespace lifecycle is managed separately.

## Prerequisites

- Kubernetes cluster (>= 1.19)
- Terraform (>= 1.0)
- kubectl configured with access to the target cluster
- Helm provider configured

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
```

## Module Inputs

### `metadata` (Required)

Metadata for the Grafana deployment.

```hcl
metadata = {
  name = "my-grafana"      # Required: Name of the Grafana instance
  id   = "unique-id"       # Optional: Unique identifier
  org  = "my-org"          # Optional: Organization name
  env  = "dev"             # Optional: Environment (dev, staging, prod)
  labels = {               # Optional: Additional labels
    "team" = "platform"
  }
  tags = ["monitoring"]    # Optional: Tags
}
```

### `spec` (Required)

Specification for the Grafana deployment.

```hcl
spec = {
  namespace = "grafana"          # Kubernetes namespace name
  create_namespace = true        # Whether to create the namespace (default: false)
  
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
    dns_domain = "example.com"  # Will create grafana-<name>.example.com
  }
}
```

## Module Outputs

| Output | Description |
|--------|-------------|
| `namespace` | The Kubernetes namespace where Grafana is deployed |
| `service` | The Kubernetes service name for Grafana |
| `port_forward_command` | Command to port-forward to Grafana locally |
| `kube_endpoint` | Internal Kubernetes service FQDN |
| `external_hostname` | External URL for Grafana (if ingress enabled) |
| `internal_hostname` | Internal URL for Grafana (if ingress enabled) |

## Usage Examples

### Basic Deployment

```hcl
module "grafana" {
  source = "path/to/module"

  metadata = {
    name = "my-grafana"
    org  = "my-company"
    env  = "production"
  }

  spec = {
    container = {
      resources = {
        requests = {
          cpu    = "100m"
          memory = "256Mi"
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
```

### With Ingress Enabled

```hcl
module "grafana" {
  source = "path/to/module"

  metadata = {
    name = "my-grafana"
    org  = "my-company"
    env  = "production"
  }

  spec = {
    container = {
      resources = {
        requests = {
          cpu    = "100m"
          memory = "256Mi"
        }
        limits = {
          cpu    = "2000m"
          memory = "2Gi"
        }
      }
    }
    
    ingress = {
      is_enabled = true
      dns_domain = "example.com"
    }
  }
}

output "grafana_url" {
  value = module.grafana.external_hostname
}
```

### Development Environment

```hcl
module "grafana_dev" {
  source = "path/to/module"

  metadata = {
    name = "dev-grafana"
    env  = "dev"
  }

  spec = {
    container = {
      resources = {
        requests = {
          cpu    = "50m"
          memory = "100Mi"
        }
        limits = {
          cpu    = "500m"
          memory = "512Mi"
        }
      }
    }
    
    ingress = {
      is_enabled = false
      dns_domain = ""
    }
  }
}

# Use port-forwarding for local access
output "port_forward_cmd" {
  value = module.grafana_dev.port_forward_command
}
```

## Accessing Grafana

### With Ingress Disabled (Default)

Use port-forwarding to access Grafana locally:

```bash
# Get the port-forward command from outputs
terraform output port_forward_command

# Run the command
kubectl port-forward -n <namespace> service/<service-name> 8080:80

# Access Grafana at http://localhost:8080
```

Default credentials:
- Username: `admin`
- Password: `admin`

### With Ingress Enabled

Access Grafana via the external hostname:

```bash
# Get the external URL
terraform output external_hostname

# Open in browser (example)
# https://grafana-my-grafana.example.com
```

## Ingress Configuration

When ingress is enabled, the module creates:

1. **External Ingress**: 
   - Host: `grafana-<name>.<dns_domain>`
   - Ingress Class: `nginx`
   - Accessible from outside the cluster

2. **Internal Ingress**: 
   - Host: `grafana-<name>-internal.<dns_domain>`
   - Ingress Class: `nginx-internal`
   - Accessible only within the network

## Resource Defaults

The default resource allocations are:

- **CPU Request**: 50m
- **Memory Request**: 100Mi
- **CPU Limit**: 1000m (1 CPU)
- **Memory Limit**: 1Gi

Adjust these based on your workload requirements.

## Notes

- Grafana persistence is disabled by default. Data will be lost on pod restart.
- The module uses the official Grafana Helm chart version 8.7.0
- Admin credentials are set to `admin/admin` by default
- For production use, consider:
  - Enabling persistence
  - Changing default admin password
  - Configuring HTTPS/TLS for ingress
  - Setting up proper authentication (LDAP, OAuth, etc.)

## Troubleshooting

### Pod not starting

Check pod logs:
```bash
kubectl logs -n <namespace> -l app.kubernetes.io/name=grafana
```

### Ingress not working

Verify ingress controller is installed:
```bash
kubectl get pods -n ingress-nginx
```

Check ingress resource:
```bash
kubectl get ingress -n <namespace>
kubectl describe ingress -n <namespace> <ingress-name>
```

### Resource limits

If pods are being OOM killed, increase memory limits in the spec.

## Additional Resources

- [Grafana Official Documentation](https://grafana.com/docs/)
- [Grafana Helm Chart](https://github.com/grafana/helm-charts)
- [Kubernetes Documentation](https://kubernetes.io/docs/)

