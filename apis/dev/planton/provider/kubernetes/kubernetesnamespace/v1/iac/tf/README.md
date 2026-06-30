# Kubernetes Namespace - Terraform Module

## Overview

This Terraform module creates production-ready Kubernetes namespaces with resource quotas, network policies, and service mesh integration following the "Namespace-as-a-Service" pattern.

## Features

- **Resource Quotas**: CPU, memory, and object count limits
- **LimitRanges**: Default container resource requests/limits
- **Network Policies**: Ingress/egress isolation with explicit allows
- **Service Mesh Integration**: Istio, Linkerd, or Consul sidecar injection
- **Pod Security Standards**: Kubernetes-native security enforcement

## Usage

```hcl
module "namespace" {
  source = "path/to/module"

  metadata = {
    name = "my-namespace"
    org  = "myorg"
    env  = "prod"
  }

  spec = {
    name = "my-namespace"
    
    labels = {
      team        = "platform"
      environment = "production"
    }

    resource_profile = {
      preset = "large"
    }

    network_config = {
      isolate_ingress            = true
      restrict_egress            = true
      allowed_ingress_namespaces = ["istio-system"]
      allowed_egress_cidrs       = ["10.0.0.0/8"]
    }

    service_mesh_config = {
      enabled      = true
      mesh_type    = "istio"
      revision_tag = "prod-stable"
    }

    pod_security_standard = "baseline"
  }
}
```

## Resource Profiles

### Preset Profiles

- **small**: 2-4 CPU, 4-8Gi memory (dev/test)
- **medium**: 4-8 CPU, 8-16Gi memory (staging)
- **large**: 8-16 CPU, 16-32Gi memory (production)
- **xlarge**: 16-32 CPU, 32-64Gi memory (high-scale)

### Custom Quotas

```hcl
resource_profile = {
  custom = {
    cpu = {
      requests = "10"
      limits   = "20"
    }
    memory = {
      requests = "20Gi"
      limits   = "40Gi"
    }
    object_counts = {
      pods     = 50
      services = 20
    }
  }
}
```

## Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Created namespace name |
| `resource_quotas_applied` | Whether quotas were configured |
| `limit_ranges_applied` | Whether default limits were set |
| `network_policies_applied` | Whether policies were created |
| `service_mesh_enabled` | Mesh injection status |
| `service_mesh_type` | Configured mesh type |
| `pod_security_standard` | Security enforcement level |

## Requirements

- Terraform >= 1.0
- Kubernetes provider >= 2.20
- Kubernetes cluster with appropriate CNI for NetworkPolicy support

## Examples

See [../../examples.md](../../examples.md) for complete YAML manifest examples that can be used with `planton tofu apply`.

## References

- [Kubernetes Namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/)
- [Resource Quotas](https://kubernetes.io/docs/concepts/policy/resource-quotas/)
- [Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)

