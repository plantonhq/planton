# KubernetesService Terraform Module

## Overview

This Terraform module creates and manages Kubernetes Service resources. It provides feature parity with the Pulumi module, supporting all service types (ClusterIP, NodePort, LoadBalancer, ExternalName) and headless services.

## Usage

### Via OpenMCF CLI (Recommended)

```bash
# Plan changes
openmcf tofu plan --manifest manifest.yaml

# Apply changes
openmcf tofu apply --manifest manifest.yaml

# Destroy resources
openmcf tofu destroy --manifest manifest.yaml
```

### Standalone Usage

```hcl
module "kubernetes_service" {
  source = "./path/to/module"

  metadata = {
    name = "my-service"
  }

  spec = {
    namespace = "production"
    name      = "my-service"
    selector  = { app = "web" }
    ports = [{
      name        = "http"
      port        = 80
      target_port = "8080"
    }]
  }
}
```

## Variables

| Variable | Description | Type | Required |
|----------|-------------|------|----------|
| `metadata.name` | Resource name for OpenMCF tracking | string | Yes |
| `spec.namespace` | Target Kubernetes namespace | string | Yes |
| `spec.name` | Service name | string | Yes |
| `spec.type` | Service type (cluster_ip, node_port, load_balancer, external_name) | string | No (default: cluster_ip) |
| `spec.selector` | Label selector for target pods | map(string) | No |
| `spec.ports` | List of port configurations | list(object) | No |
| `spec.headless` | Create headless service (clusterIP: None) | bool | No (default: false) |
| `spec.external_dns_name` | DNS name for ExternalName services | string | No |
| `spec.external_traffic_policy` | External traffic policy (cluster, local) | string | No (default: cluster) |
| `spec.session_affinity` | Session affinity (none, client_ip) | string | No (default: none) |
| `spec.load_balancer_source_ranges` | IP ranges for LoadBalancer access control | list(string) | No |

## Outputs

| Output | Description |
|--------|-------------|
| `service_name` | The created service name |
| `namespace` | The namespace where the service was created |
| `type` | The service type |
| `cluster_ip` | The allocated cluster IP |
| `load_balancer_ingress` | LoadBalancer hostname or IP |
| `internal_dns_name` | Fully qualified internal DNS name |

## Requirements

- Terraform >= 1.0
- hashicorp/kubernetes provider >= 2.20
