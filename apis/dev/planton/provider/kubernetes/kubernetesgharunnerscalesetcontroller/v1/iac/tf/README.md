# KubernetesGhaRunnerScaleSetController Terraform Module

This directory contains the Terraform module for deploying the GitHub Actions Runner Scale Set Controller on Kubernetes.

## Prerequisites

- Terraform >= 1.0
- kubectl configured with cluster access
- Helm provider configured

## Usage

### With Planton CLI

```bash
planton tofu apply --manifest gha-controller.yaml
```

### Standalone Usage

```hcl
module "gha_controller" {
  source = "./path/to/module"

  metadata = {
    name = "arc-controller"
    org  = "my-org"
    env  = "production"
  }

  namespace        = "arc-system"
  create_namespace = true

  container = {
    resources = {
      requests = {
        cpu    = "100m"
        memory = "128Mi"
      }
      limits = {
        cpu    = "500m"
        memory = "512Mi"
      }
    }
  }

  flags = {
    log_level       = "info"
    log_format      = "json"
    update_strategy = "eventual"
  }
}
```

## Variables

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| `metadata` | Resource metadata | object | - | yes |
| `namespace` | Controller namespace | string | - | yes |
| `create_namespace` | Create namespace | bool | true | no |
| `helm_chart_version` | Chart version | string | "0.13.1" | no |
| `replica_count` | Replicas | number | 1 | no |
| `container` | Container config | object | {} | no |
| `flags` | Controller flags | object | {} | no |
| `metrics` | Metrics config | object | null | no |
| `image_pull_secrets` | Pull secrets | list(string) | [] | no |
| `priority_class_name` | Priority class | string | "" | no |

## Outputs

| Name | Description |
|------|-------------|
| `namespace` | Controller namespace |
| `release_name` | Helm release name |
| `chart_version` | Deployed chart version |
| `deployment_name` | Deployment name |
| `service_account_name` | Service account name |
| `metrics_endpoint` | Metrics endpoint (if enabled) |

## Examples

### High Availability Setup

```hcl
module "gha_controller_ha" {
  source = "./path/to/module"

  metadata = {
    name = "arc-controller"
  }

  namespace           = "arc-system"
  create_namespace    = true
  replica_count       = 3
  priority_class_name = "system-cluster-critical"

  container = {
    resources = {
      requests = { cpu = "200m", memory = "256Mi" }
      limits   = { cpu = "1000m", memory = "1Gi" }
    }
  }

  flags = {
    log_level                        = "info"
    log_format                       = "json"
    runner_max_concurrent_reconciles = 10
    update_strategy                  = "eventual"
  }
}
```

### With Metrics

```hcl
module "gha_controller_monitored" {
  source = "./path/to/module"

  metadata = {
    name = "arc-controller"
  }

  namespace        = "arc-system"
  create_namespace = true

  metrics = {
    controller_manager_addr = ":8080"
    listener_addr           = ":8080"
    listener_endpoint       = "/metrics"
  }
}
```

## Troubleshooting

### Helm Provider Configuration

Ensure your Helm provider is configured:

```hcl
provider "helm" {
  kubernetes {
    config_path = "~/.kube/config"
  }
}
```

### OCI Registry Access

The chart is hosted on ghcr.io. Ensure network access to:
- `ghcr.io/actions/actions-runner-controller-charts`

### Verifying Deployment

```bash
kubectl get pods -n arc-system
kubectl get crds | grep actions.github.com
```

