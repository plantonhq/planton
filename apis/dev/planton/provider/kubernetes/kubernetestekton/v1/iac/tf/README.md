# KubernetesTekton Terraform Module

This Terraform module deploys Tekton Pipelines and Dashboard on Kubernetes using official release manifests.

## Overview

The module applies Tekton release YAMLs directly, similar to:

```bash
kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
kubectl apply --filename https://infra.tekton.dev/tekton-releases/dashboard/latest/release.yaml
```

## Features

- **Tekton Pipelines**: Core CI/CD pipeline engine
- **Tekton Dashboard**: Web UI for viewing pipelines (optional)
- **CloudEvents**: Pipeline event notifications (optional)
- **Dashboard Ingress**: Gateway API with TLS (optional)

## Module Structure

```
tf/
├── variables.tf  # Input variables
├── locals.tf     # Computed values
├── main.tf       # Manifest deployment
├── ingress.tf    # Gateway API resources
├── outputs.tf    # Module outputs
└── provider.tf   # Provider requirements
```

## Usage

### Basic

```hcl
module "tekton" {
  source = "./path/to/module"

  metadata = {
    name = "my-tekton"
  }

  spec = {
    pipeline_version = "latest"
  }
}
```

### With Dashboard

```hcl
module "tekton" {
  source = "./path/to/module"

  metadata = {
    name = "my-tekton"
    org  = "my-org"
    env  = "prod"
  }

  spec = {
    pipeline_version = "v0.65.2"
    dashboard = {
      enabled = true
      version = "v0.53.0"
    }
  }
}
```

### With CloudEvents and Ingress

```hcl
module "tekton" {
  source = "./path/to/module"

  metadata = {
    name = "my-tekton"
  }

  spec = {
    pipeline_version = "latest"

    dashboard = {
      enabled = true
      version = "latest"
      ingress = {
        enabled  = true
        hostname = "tekton.example.com"
      }
    }

    cloud_events = {
      sink_url = "http://receiver.ns.svc.cluster.local/events"
    }
  }
}
```

## Inputs

| Name | Description | Type | Default |
|------|-------------|------|---------|
| `metadata` | Resource metadata (name, org, env) | object | required |
| `spec.pipeline_version` | Tekton Pipelines version | string | `"latest"` |
| `spec.dashboard.enabled` | Enable Tekton Dashboard | bool | `false` |
| `spec.dashboard.version` | Tekton Dashboard version | string | `"latest"` |
| `spec.dashboard.ingress.enabled` | Enable dashboard ingress | bool | `false` |
| `spec.dashboard.ingress.hostname` | Dashboard hostname | string | `null` |
| `spec.cloud_events.sink_url` | CloudEvents receiver URL | string | `null` |

## Outputs

| Name | Description |
|------|-------------|
| `namespace` | Always `tekton-pipelines` |
| `pipeline_version` | Deployed pipeline version |
| `dashboard_version` | Deployed dashboard version |
| `dashboard_internal_endpoint` | Cluster-internal dashboard URL |
| `dashboard_external_hostname` | External hostname (if ingress enabled) |
| `port_forward_dashboard_command` | kubectl port-forward command |
| `cloud_events_sink_url` | Configured CloudEvents sink |

## Requirements

| Provider | Version |
|----------|---------|
| kubernetes | >= 2.0 |
| kubectl | >= 1.14 |
| http | >= 3.0 |

## Debugging

```bash
# Check deployed resources
kubectl get pods -n tekton-pipelines

# View ConfigMap
kubectl get cm config-defaults -n tekton-pipelines -o yaml

# Check ingress
kubectl get gateway,httproute -A
```
