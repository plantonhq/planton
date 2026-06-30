# KubernetesTekton

Deploy Tekton Pipelines and Dashboard on Kubernetes using official release manifests.

## Overview

KubernetesTekton provides a manifest-based deployment of Tekton, ideal for users who prefer direct control over installation without the Tekton Operator's lifecycle management overhead.

**Key Features:**
- Deploy Tekton Pipelines using official release manifests
- Optional Tekton Dashboard deployment
- CloudEvents integration for pipeline notifications
- Dashboard ingress via Kubernetes Gateway API

## Comparison: KubernetesTekton vs KubernetesTektonOperator

| Feature | KubernetesTekton | KubernetesTektonOperator |
|---------|------------------|-------------------------|
| Deployment Method | Direct manifests (kubectl apply) | Operator with TektonConfig CRD |
| Complexity | Simpler, direct | More abstraction layers |
| Lifecycle Management | Manual upgrades | Automated by operator |
| Configuration | ConfigMap patches | TektonConfig CR |
| Best For | Simple deployments, debugging | Production with automation |

## Quick Start

### Minimal Example

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTekton
metadata:
  name: my-tekton
spec:
  pipeline_version: "latest"
```

### With Dashboard

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTekton
metadata:
  name: my-tekton
spec:
  pipeline_version: "v0.65.2"
  dashboard:
    enabled: true
    version: "v0.53.0"
```

### With CloudEvents and Dashboard Ingress

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTekton
metadata:
  name: my-tekton
spec:
  pipeline_version: "latest"
  dashboard:
    enabled: true
    version: "latest"
    ingress:
      enabled: true
      hostname: "tekton-dashboard.example.com"
  cloud_events:
    sink_url: "http://my-receiver.my-namespace.svc.cluster.local/tekton/events"
```

## Namespace Behavior

Tekton components are installed in the `tekton-pipelines` namespace, which is created automatically by the official Tekton manifests. **This namespace cannot be customized.**

## CloudEvents Integration

When `cloud_events.sink_url` is configured, Tekton sends CloudEvents for:
- TaskRun state changes (started, running, succeeded, failed)
- PipelineRun state changes

This enables external systems to react to pipeline events (e.g., update a UI, trigger notifications, start downstream processes).

## Dashboard Ingress

When enabled, the dashboard is exposed via Kubernetes Gateway API:
- TLS certificate via cert-manager
- HTTP → HTTPS redirect
- Routes traffic to `tekton-dashboard:9097`

**Prerequisites:**
- Istio ingress gateway installed
- cert-manager with ClusterIssuer matching your domain
- Gateway API CRDs installed

## Accessing the Dashboard

### Port Forward (No Ingress)

```bash
kubectl port-forward -n tekton-pipelines service/tekton-dashboard 9097:9097
```

Then open: http://localhost:9097

### Via Ingress

If dashboard ingress is enabled, access via the configured hostname (e.g., https://tekton-dashboard.example.com).

## Version Selection

### Pipeline Versions
See: https://github.com/tektoncd/pipeline/releases

### Dashboard Versions  
See: https://github.com/tektoncd/dashboard/releases

Use `"latest"` for the most recent stable release, or pin to a specific version (e.g., `"v0.65.2"`).
