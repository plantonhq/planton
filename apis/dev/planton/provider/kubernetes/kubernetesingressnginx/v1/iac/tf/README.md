# Kubernetes Ingress NGINX - Terraform Module

This Terraform module deploys the official NGINX Ingress Controller on Kubernetes clusters with cloud-specific optimizations for GKE, EKS, and AKS.

## Overview

The module provides automatic configuration for:

- **Multi-Cloud Support**: GCP, AWS, and Azure
- **Load Balancer Setup**: Native cloud load balancer integration
- **Internal/External Control**: Toggle private vs public access
- **Namespace Management**: Create new or use existing namespaces
- **Static IP Assignment**: Cloud-specific static IP support
- **Security Integration**: Security groups, managed identities

## Prerequisites

- Terraform >= 1.0
- Kubernetes cluster access
- kubectl configured
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

Metadata for the ingress controller deployment.

```hcl
metadata = {
  name = "my-ingress"
  id   = "unique-id"
  org  = "my-org"
  env  = "production"
}
```

### `spec` (Required)

Specification for the ingress controller.

```hcl
spec = {
  chart_version = "4.11.1"  # Optional, uses default if not specified
  internal      = false     # false = external, true = internal
  
  # Cloud-specific configuration (choose one)
  gke = {
    static_ip_name       = "my-static-ip"
    subnetwork_self_link = "projects/my-project/regions/us-west1/subnetworks/private"
  }
  
  # OR
  eks = {
    additional_security_group_ids = ["sg-123", "sg-456"]
    subnet_ids                    = ["subnet-abc", "subnet-def"]
    irsa_role_arn_override        = "arn:aws:iam::123456789012:role/ingress-nginx"
  }
  
  # OR
  aks = {
    managed_identity_client_id = "12345678-1234-1234-1234-123456789012"
    public_ip_name             = "my-public-ip"
  }
}
```

## Module Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Kubernetes namespace where controller is deployed |
| `release_name` | Helm release name |
| `service_name` | Kubernetes service name for the controller |
| `service_type` | Service type (LoadBalancer) |

## Usage Examples

### Basic External Ingress

```hcl
module "ingress_nginx" {
  source = "path/to/module"

  metadata = {
    name = "basic-ingress"
  }

  spec = {
    chart_version = "4.11.1"
    internal      = false
  }
}
```

### GKE with Static IP

```hcl
module "gke_ingress" {
  source = "path/to/module"

  metadata = {
    name = "gke-ingress"
    env  = "production"
  }

  spec = {
    chart_version = "4.11.1"
    internal      = false
    gke = {
      static_ip_name = "prod-ingress-ip"
    }
  }
}

output "load_balancer_ip" {
  value = "Check GCP console for the static IP"
}
```

### EKS Internal with Subnets

```hcl
module "eks_internal_ingress" {
  source = "path/to/module"

  metadata = {
    name = "internal-ingress"
    env  = "production"
  }

  spec = {
    chart_version = "4.11.1"
    internal      = true
    eks = {
      subnet_ids = [
        "subnet-private-1a",
        "subnet-private-1b",
        "subnet-private-1c"
      ]
      additional_security_group_ids = [
        "sg-app-internal"
      ]
    }
  }
}
```

### AKS with Managed Identity

```hcl
module "aks_ingress" {
  source = "path/to/module"

  metadata = {
    name = "aks-ingress"
    env  = "production"
  }

  spec = {
    chart_version = "4.11.1"
    internal      = false
    aks = {
      managed_identity_client_id = "12345678-1234-1234-1234-123456789012"
      public_ip_name             = "prod-ingress-public-ip"
    }
  }
}
```

## Accessing Outputs

```hcl
output "controller_namespace" {
  value = module.ingress_nginx.namespace
}

output "controller_service" {
  value = module.ingress_nginx.service_name
}
```

## Verification

After deployment:

```bash
# Check namespace
kubectl get ns kubernetes-ingress-nginx

# Check Helm release
helm list -n kubernetes-ingress-nginx

# Check controller pods
kubectl get pods -n kubernetes-ingress-nginx

# Check load balancer service
kubectl get svc -n kubernetes-ingress-nginx

# Get load balancer IP/hostname
kubectl get svc -n kubernetes-ingress-nginx kubernetes-ingress-nginx-controller -o jsonpath='{.status.loadBalancer.ingress[0]}'
```

## Creating Ingress Resources

After the controller is deployed, create Ingress resources to route traffic:

```hcl
resource "kubernetes_ingress_v1" "example" {
  metadata {
    name      = "myapp-ingress"
    namespace = "default"
  }

  spec {
    ingress_class_name = "nginx"
    
    rule {
      host = "myapp.example.com"
      http {
        path {
          path      = "/"
          path_type = "Prefix"
          backend {
            service {
              name = "myapp"
              port {
                number = 80
              }
            }
          }
        }
      }
    }
  }
}
```

## Troubleshooting

### Load Balancer Pending

If the service stays in Pending state:

1. Check cloud provider quotas
2. Verify network configuration (subnets, security groups)
3. Review controller logs:
   ```bash
   kubectl logs -n kubernetes-ingress-nginx -l app.kubernetes.io/component=controller
   ```

### Static IP Not Assigned (GKE)

```bash
# Verify IP exists
gcloud compute addresses list

# Check if IP is global or regional (must match LB type)
gcloud compute addresses describe <ip-name> --global
```

### Security Group Issues (EKS)

```bash
# Verify security groups exist
aws ec2 describe-security-groups --group-ids <sg-id>

# Check rules allow traffic
aws ec2 describe-security-group-rules --filters Name=group-id,Values=<sg-id>
```

## Advanced Configuration

For advanced Helm chart configuration, you can extend the module or use the Helm provider directly with additional values.

## Additional Resources

- [Module Examples](./examples.md) - Terraform-specific examples
- [NGINX Ingress Controller Docs](https://kubernetes.github.io/ingress-nginx/)
- [Terraform Helm Provider](https://registry.terraform.io/providers/hashicorp/helm/latest/docs)
- [Terraform Kubernetes Provider](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs)

