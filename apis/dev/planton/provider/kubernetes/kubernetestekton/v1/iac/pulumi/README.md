# KubernetesTekton Pulumi Module

This Pulumi module deploys Tekton Pipelines and Dashboard on Kubernetes using official release manifests.

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
module/
├── vars.go       # Configuration constants
├── locals.go     # Computed values and exports
├── main.go       # Orchestration
├── tekton.go     # Manifest deployment
├── config.go     # CloudEvents ConfigMap patch
├── ingress.go    # Gateway API resources
└── outputs.go    # Output documentation
```

## Usage

### Basic

```go
stackInput := &kubernetestektonv1.KubernetesTektonStackInput{
    Target: &kubernetestektonv1.KubernetesTekton{
        Metadata: &shared.CloudResourceMetadata{
            Name: "my-tekton",
        },
        Spec: &kubernetestektonv1.KubernetesTektonSpec{
            PipelineVersion: "latest",
            Dashboard: &kubernetestektonv1.KubernetesTektonDashboard{
                Enabled: true,
            },
        },
    },
}
```

### With CloudEvents

```go
Spec: &kubernetestektonv1.KubernetesTektonSpec{
    PipelineVersion: "latest",
    Dashboard: &kubernetestektonv1.KubernetesTektonDashboard{
        Enabled: true,
    },
    CloudEvents: &kubernetestektonv1.KubernetesTektonCloudEvents{
        SinkUrl: "http://receiver.ns.svc.cluster.local/events",
    },
},
```

### With Dashboard Ingress

```go
Spec: &kubernetestektonv1.KubernetesTektonSpec{
    PipelineVersion: "latest",
    Dashboard: &kubernetestektonv1.KubernetesTektonDashboard{
        Enabled: true,
        Version: "v0.53.0",
        Ingress: &kubernetestektonv1.KubernetesTektonDashboardIngress{
            Enabled:  true,
            Hostname: "tekton.example.com",
        },
    },
},
```

## Stack Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Always `tekton-pipelines` |
| `pipeline_version` | Deployed pipeline version |
| `dashboard_version` | Deployed dashboard version |
| `dashboard_internal_endpoint` | Cluster-internal dashboard URL |
| `dashboard_external_hostname` | External hostname (if ingress enabled) |
| `port_forward_dashboard_command` | kubectl port-forward command |
| `cloud_events_sink_url` | Configured CloudEvents sink |

## Deployment Order

1. **Tekton Pipelines** - Creates namespace, CRDs, controllers
2. **Tekton Dashboard** - Adds web UI (depends on pipelines)
3. **CloudEvents Config** - Patches config-defaults ConfigMap
4. **Dashboard Ingress** - Certificate, Gateway, HTTPRoutes

## Debugging

### Run with debug output

Uncomment the binary option in `Pulumi.yaml`:

```yaml
runtime:
  name: go
  options:
    binary: ./debug.sh
```

### Check deployed resources

```bash
# Pipeline controller
kubectl get pods -n tekton-pipelines -l app=tekton-pipelines-controller

# Dashboard
kubectl get pods -n tekton-pipelines -l app=tekton-dashboard

# CloudEvents config
kubectl get cm config-defaults -n tekton-pipelines -o yaml

# Ingress resources
kubectl get gateway,httproute -A
```

## Requirements

- Kubernetes cluster with kubectl access
- For ingress: Istio Gateway controller, cert-manager, Gateway API CRDs
