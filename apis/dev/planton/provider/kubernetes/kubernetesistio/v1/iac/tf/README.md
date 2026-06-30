# Kubernetes Istio - Terraform Module

This Terraform module deploys Istio service mesh on Kubernetes clusters using the official Istio Helm charts with automatic component orchestration and resource management.

## Overview

The module provides automated deployment of:

- **Complete Istio Stack**: Base CRDs, istiod control plane, and ingress gateway
- **Resource Configuration**: Tunable CPU and memory for the control plane
- **Proper Dependencies**: Automatic ordering of component installations
- **Namespace Isolation**: Dedicated namespaces for control plane and gateway
- **Production Ready**: Atomic Helm releases with automatic rollback

## Prerequisites

- Terraform >= 1.0
- Kubernetes cluster access (1.25+)
- kubectl configured
- Helm provider configured
- Sufficient cluster resources (see Resource Requirements)

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

Metadata for the Istio deployment.

```hcl
metadata = {
  name = "main-istio"
  id   = "unique-id"      # Optional
  org  = "my-org"         # Optional
  env  = "production"     # Optional
}
```

### `spec` (Required)

Specification for the Istio service mesh.

```hcl
spec = {
  namespace        = "istio-system"
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
}
```

## Module Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Namespace where Istio control plane is deployed (istio-system) |
| `service` | Name of the istiod service |
| `port_forward_command` | Command to port-forward to istiod for debugging |
| `kube_endpoint` | Kubernetes service endpoint for istiod |
| `ingress_endpoint` | Kubernetes service endpoint for ingress gateway |

## Usage Examples

### Basic Istio Deployment

Minimal configuration for development environments:

```hcl
module "istio" {
  source = "path/to/module"

  metadata = {
    name = "dev-istio"
  }

  spec = {
    namespace        = "istio-system"
    create_namespace = true

    container = {
      resources = {
        requests = {
          cpu    = "25m"
          memory = "64Mi"
        }
        limits = {
          cpu    = "500m"
          memory = "256Mi"
        }
      }
    }
  }
}

output "istio_namespace" {
  value = module.istio.namespace
}

output "port_forward_command" {
  value = module.istio.port_forward_command
}
```

### Production Istio Deployment

Standard production configuration:

```hcl
module "istio_prod" {
  source = "path/to/module"

  metadata = {
    name = "prod-istio"
    env  = "production"
    org  = "my-company"
  }

  spec = {
    namespace        = "istio-system"
    create_namespace = true

    container = {
      resources = {
        requests = {
          cpu    = "500m"
          memory = "512Mi"
        }
        limits = {
          cpu    = "2000m"
          memory = "2Gi"
        }
      }
    }
  }
}

# Enable automatic sidecar injection
resource "kubernetes_namespace_v1" "app" {
  metadata {
    name = "production-apps"
    labels = {
      "istio-injection" = "enabled"
    }
  }
  depends_on = [module.istio_prod]
}
```

### High-Availability Configuration

For large-scale production deployments:

```hcl
module "istio_ha" {
  source = "path/to/module"

  metadata = {
    name = "ha-istio"
    env  = "production"
    org  = "enterprise"
  }

  spec = {
    namespace        = "istio-system"
    create_namespace = true

    container = {
      resources = {
        requests = {
          cpu    = "1000m"
          memory = "1Gi"
        }
        limits = {
          cpu    = "4000m"
          memory = "8Gi"
        }
      }
    }
  }
}

# Outputs for monitoring
output "control_plane_namespace" {
  description = "Namespace for monitoring control plane pods"
  value       = module.istio_ha.namespace
}

output "istiod_service" {
  description = "Istiod service name for metrics scraping"
  value       = module.istio_ha.service
}

output "kube_endpoint" {
  description = "Internal cluster endpoint for service discovery"
  value       = module.istio_ha.kube_endpoint
}
```

### Complete Example with Application Deployment

```hcl
# Deploy Istio
module "istio" {
  source = "path/to/module"

  metadata = {
    name = "main-istio"
    env  = "staging"
  }

  spec = {
    namespace        = "istio-system"
    create_namespace = true

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
  }
}

# Create application namespace with sidecar injection
resource "kubernetes_namespace_v1" "bookinfo" {
  metadata {
    name = "bookinfo"
    labels = {
      "istio-injection" = "enabled"
    }
  }
  depends_on = [module.istio]
}

# Deploy sample application
resource "kubernetes_deployment_v1" "productpage" {
  metadata {
    name      = "productpage"
    namespace = kubernetes_namespace_v1.bookinfo.metadata[0].name
  }

  spec {
    replicas = 1
    selector {
      match_labels = {
        app = "productpage"
      }
    }
    template {
      metadata {
        labels = {
          app = "productpage"
        }
      }
      spec {
        container {
          name  = "productpage"
          image = "docker.io/istio/examples-bookinfo-productpage-v1:1.18.0"
          port {
            container_port = 9080
          }
        }
      }
    }
  }
  depends_on = [module.istio]
}
```

## Resource Requirements

### Development Environment
- **CPU**: 500m+ available
- **Memory**: 256Mi+ available
- **Nodes**: 1+ node
- **Use Case**: Testing, local development

### Production Environment
- **CPU**: 2+ cores available
- **Memory**: 2Gi+ available
- **Nodes**: 3+ nodes (for HA)
- **Use Case**: Standard production workloads

### Enterprise Environment
- **CPU**: 8+ cores available
- **Memory**: 16Gi+ available
- **Nodes**: 5+ nodes across multiple zones
- **Use Case**: Large-scale, multi-tenant clusters

## Deployment Process

The module performs the following steps:

1. **Create Namespaces**
   - `istio-system` for control plane
   - `istio-ingress` for ingress gateway

2. **Deploy Istio Base**
   - Installs CRDs (Custom Resource Definitions)
   - Creates foundational resources
   - Timeout: 180 seconds

3. **Deploy Istiod Control Plane**
   - Installs control plane with configured resources
   - Configures Pilot with custom CPU/memory
   - Depends on base completion
   - Timeout: 180 seconds

4. **Deploy Ingress Gateway**
   - Creates gateway pods in istio-ingress namespace
   - Configures as ClusterIP service
   - Depends on istiod completion
   - Timeout: 180 seconds

5. **Export Outputs**
   - Service endpoints
   - Debug commands
   - Namespace information

## Verification

After deployment, verify the installation:

```bash
# Check control plane
kubectl get pods -n istio-system
kubectl get svc -n istio-system

# Check ingress gateway
kubectl get pods -n istio-ingress
kubectl get svc -n istio-ingress

# View Helm releases
helm list -n istio-system
helm list -n istio-ingress

# Check Istio version
kubectl get pods -n istio-system -l app=istiod -o jsonpath='{.items[0].spec.containers[0].image}'
```

Expected output:
```
# istio-system namespace
NAME                      READY   STATUS    RESTARTS   AGE
istiod-xxxxxxxxxx-xxxxx   1/1     Running   0          2m

# istio-ingress namespace
NAME                             READY   STATUS    RESTARTS   AGE
istio-gateway-xxxxxxxxxx-xxxxx   1/1     Running   0          1m
```

## Post-Deployment Configuration

### Enable Sidecar Injection

```bash
kubectl label namespace default istio-injection=enabled
kubectl label namespace production istio-injection=enabled
```

Or using Terraform:

```hcl
resource "kubernetes_namespace_v1" "production" {
  metadata {
    name = "production"
    labels = {
      "istio-injection" = "enabled"
    }
  }
  depends_on = [module.istio]
}
```

### Configure mTLS

Create PeerAuthentication for strict mTLS:

```hcl
resource "kubernetes_manifest" "peer_authentication" {
  manifest = {
    apiVersion = "security.istio.io/v1beta1"
    kind       = "PeerAuthentication"
    metadata = {
      name      = "default"
      namespace = module.istio.namespace
    }
    spec = {
      mtls = {
        mode = "STRICT"
      }
    }
  }
  depends_on = [module.istio]
}
```

### Deploy Gateway

```hcl
resource "kubernetes_manifest" "gateway" {
  manifest = {
    apiVersion = "networking.istio.io/v1beta1"
    kind       = "Gateway"
    metadata = {
      name      = "bookinfo-gateway"
      namespace = "default"
    }
    spec = {
      selector = {
        istio = "gateway"
      }
      servers = [
        {
          port = {
            number   = 80
            name     = "http"
            protocol = "HTTP"
          }
          hosts = ["*"]
        }
      ]
    }
  }
  depends_on = [module.istio]
}
```

## Namespace Management

The module manages two Istio namespaces:

- **istio-system**: Control plane namespace (istiod, base CRDs)
- **istio-ingress**: Ingress gateway namespace

### Automatic Namespace Creation

Set `create_namespace = true` to automatically create both namespaces:

```hcl
module "istio" {
  source = "path/to/module"

  metadata = {
    name = "prod-istio"
  }

  spec = {
    namespace        = "istio-system"
    create_namespace = true  # Creates both namespaces automatically

    container = {
      resources = {
        requests = {
          cpu    = "500m"
          memory = "512Mi"
        }
        limits = {
          cpu    = "2000m"
          memory = "2Gi"
        }
      }
    }
  }
}
```

### Using Existing Namespaces

Set `create_namespace = false` to use pre-existing namespaces:

```hcl
module "istio" {
  source = "path/to/module"

  metadata = {
    name = "prod-istio"
  }

  spec = {
    namespace        = "istio-system"
    create_namespace = false  # Use existing namespaces

    container = {
      resources = {
        requests = {
          cpu    = "500m"
          memory = "512Mi"
        }
        limits = {
          cpu    = "2000m"
          memory = "2Gi"
        }
      }
    }
  }
}
```

**Prerequisites when using existing namespaces:**
- Both `istio-system` and `istio-ingress` namespaces must already exist
- Namespaces should have appropriate labels and RBAC configured

**Create namespaces manually if needed:**

```bash
kubectl create namespace istio-system
kubectl create namespace istio-ingress
```

Or using Terraform:

```hcl
resource "kubernetes_namespace_v1" "istio_system" {
  metadata {
    name = "istio-system"
  }
}

resource "kubernetes_namespace_v1" "istio_ingress" {
  metadata {
    name = "istio-ingress"
  }
}

module "istio" {
  source = "path/to/module"
  
  # ... configuration ...
  
  depends_on = [
    kubernetes_namespace_v1.istio_system,
    kubernetes_namespace_v1.istio_ingress
  ]
}
```

## Troubleshooting

### Namespace Not Found Error

If you encounter "namespace not found" errors:

1. Verify `create_namespace` setting:
   ```bash
   # If create_namespace = false, check namespaces exist
   kubectl get namespace istio-system istio-ingress
   ```

2. Solutions:
   - Set `create_namespace = true` to let the module create namespaces
   - Or manually create the namespaces before running `terraform apply`

### Control Plane Not Starting

Check pod status and events:

```bash
kubectl get pods -n istio-system
kubectl describe pod -n istio-system <istiod-pod-name>
kubectl logs -n istio-system -l app=istiod
```

### Helm Release Failed

Check Helm release status:

```bash
helm status base -n istio-system
helm status istiod -n istio-system
helm status gateway -n istio-ingress
```

### Resource Issues

Check cluster resource availability:

```bash
kubectl top nodes
kubectl describe node <node-name>
```

### Debugging with Port-Forward

Use the exported port-forward command:

```bash
# Get the command from Terraform output
terraform output port_forward_command

# Execute it
kubectl port-forward -n istio-system svc/istiod 15014:15014

# Access debug interface
curl http://localhost:15014/debug
```

## Upgrading

To upgrade Istio to a newer version:

1. Update the chart version in `locals.tf`
2. Run Terraform plan to preview changes:
   ```bash
   terraform plan
   ```
3. Apply the upgrade:
   ```bash
   terraform apply
   ```

The module's atomic release configuration ensures automatic rollback on failure.

## Cleanup

To remove Istio:

```bash
terraform destroy
```

Note: This removes all Istio components. Ensure no applications depend on the service mesh before destroying.

## Best Practices

1. **Start with Conservative Resources**: Begin with standard production values and scale based on metrics
2. **Monitor Control Plane**: Set up monitoring for istiod CPU and memory usage
3. **Use Namespaces**: Keep control plane and applications in separate namespaces
4. **Enable mTLS**: Configure strict mTLS for production environments
5. **Label for Injection**: Use namespace labels for automatic sidecar injection
6. **Version Pinning**: Use specific chart versions in production (configured in locals.tf)
7. **Test Upgrades**: Always test Istio upgrades in staging before production

## Additional Resources

- [Istio Official Documentation](https://istio.io/latest/docs/)
- [Istio Helm Charts](https://github.com/istio/istio/tree/master/manifests/charts)
- [Terraform Helm Provider](https://registry.terraform.io/providers/hashicorp/helm/latest/docs)
- [Terraform Kubernetes Provider](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs)
- [Component Overview](../pulumi/overview.md) - Detailed architecture documentation

