# Kubernetes Tekton

Deploys Tekton Pipelines and optionally Tekton Dashboard on Kubernetes by applying official upstream release manifests directly, without requiring the Tekton Operator. This manifest-based approach gives direct control over component versions, is simpler to understand and debug, and supports optional CloudEvents integration for pipeline event notifications and external dashboard access through Istio Gateway API ingress with automatic TLS via cert-manager.

## What Gets Created

When you deploy a KubernetesTekton resource, Planton provisions:

- **Tekton Pipelines** — all resources from the official Tekton Pipeline release manifest including the `tekton-pipelines` namespace, CRDs (`Task`, `Pipeline`, `TaskRun`, `PipelineRun`, etc.), controllers, and webhook admission controllers
- **Tekton Dashboard** — the web UI for viewing and managing pipelines, tasks, and runs, deployed from the official Dashboard release manifest; created only when `dashboard.enabled` is `true`
- **CloudEvents ConfigMap Patch** — a patch to the `config-defaults` ConfigMap in the `tekton-pipelines` namespace that sets the `default-cloud-events-sink` key; created only when `cloudEvents.sinkUrl` is specified
- **TLS Certificate** — a cert-manager Certificate for the dashboard ingress hostname; created only when both `dashboard.enabled` and `dashboard.ingress.enabled` are `true`
- **Istio Gateway** — an external Gateway resource with HTTPS (port 443) and HTTP (port 80) listeners for the dashboard; created only when dashboard ingress is enabled
- **HTTP-to-HTTPS Redirect Route** — an HTTPRoute that redirects HTTP traffic to HTTPS with a 301 status code; created only when dashboard ingress is enabled
- **HTTPS Route** — an HTTPRoute that forwards HTTPS traffic to the Tekton Dashboard service on port 9097; created only when dashboard ingress is enabled

## Prerequisites

- **Kubernetes credentials** configured via environment variables or Planton provider config
- **Istio** with Gateway API support installed if enabling dashboard ingress
- **cert-manager** with a ClusterIssuer matching the ingress domain if enabling dashboard ingress with TLS

## Quick Start

Create a file `tekton.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTekton
metadata:
  name: my-tekton
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesTekton.my-tekton
spec:
  pipelineVersion: latest
```

Deploy:

```shell
planton apply -f tekton.yaml
```

This deploys the latest Tekton Pipelines release into the `tekton-pipelines` namespace. The namespace is created automatically by the upstream Tekton manifest.

## Configuration Reference

### Required Fields

All spec fields have sensible defaults. There are no strictly required fields beyond the standard `metadata` block.

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `pipelineVersion` | `string` | `latest` | Version of Tekton Pipelines to deploy. Maps to releases at https://github.com/tektoncd/pipeline/releases (e.g., `v0.65.2`, `v0.64.0`). |
| `dashboard.enabled` | `bool` | `false` | Enables deployment of the Tekton Dashboard web UI. |
| `dashboard.version` | `string` | `latest` | Version of Tekton Dashboard to deploy. Maps to releases at https://github.com/tektoncd/dashboard/releases (e.g., `v0.53.0`, `v0.52.0`). |
| `dashboard.ingress.enabled` | `bool` | `false` | Enables external access to the dashboard through Istio Gateway API with TLS termination and HTTP-to-HTTPS redirect. Requires `dashboard.enabled` to also be `true`. |
| `dashboard.ingress.hostname` | `string` | — | Full hostname for external access to the dashboard (e.g., `tekton.example.com`). Required when `dashboard.ingress.enabled` is `true`. |
| `cloudEvents.sinkUrl` | `string` | — | URL where CloudEvents will be sent for TaskRun and PipelineRun lifecycle events. Must be a valid HTTP or HTTPS URL (e.g., `http://my-service.my-namespace.svc.cluster.local/tekton/events`). |

## Examples

### Tekton Pipelines Only

A minimal deployment that installs just the Tekton Pipeline engine with a pinned version:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTekton
metadata:
  name: ci-tekton
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesTekton.ci-tekton
spec:
  pipelineVersion: v0.65.2
```

### Tekton with Dashboard

Tekton Pipelines and Dashboard deployed together, with the dashboard accessible inside the cluster via port-forward:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTekton
metadata:
  name: team-tekton
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.KubernetesTekton.team-tekton
spec:
  pipelineVersion: v0.65.2
  dashboard:
    enabled: true
    version: v0.53.0
```

After deployment, access the dashboard locally:

```shell
kubectl port-forward -n tekton-pipelines service/tekton-dashboard 9097:9097
```

### Tekton with Dashboard Ingress and CloudEvents

A full production setup with the dashboard exposed externally via TLS-terminated ingress and CloudEvents integration for pipeline notifications:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTekton
metadata:
  name: prod-tekton
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesTekton.prod-tekton
spec:
  pipelineVersion: v0.65.2
  dashboard:
    enabled: true
    version: v0.53.0
    ingress:
      enabled: true
      hostname: tekton-dashboard.example.com
  cloudEvents:
    sinkUrl: http://event-router.platform.svc.cluster.local/tekton/events
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Namespace where Tekton components are installed (always `tekton-pipelines`) |
| `pipeline_version` | `string` | Version of Tekton Pipelines that was deployed |
| `dashboard_version` | `string` | Version of Tekton Dashboard that was deployed; empty if dashboard is disabled |
| `dashboard_internal_endpoint` | `string` | Cluster-internal FQDN for the dashboard (format: `tekton-dashboard.tekton-pipelines.svc.cluster.local:9097`); empty if dashboard is disabled |
| `dashboard_external_hostname` | `string` | Public hostname for external access to the dashboard; only set when dashboard ingress is enabled |
| `port_forward_dashboard_command` | `string` | kubectl port-forward command for local access to the dashboard on port 9097; empty if dashboard is disabled |
| `cloud_events_sink_url` | `string` | CloudEvents sink URL configured for pipeline notifications; only set when `cloudEvents.sinkUrl` is specified |

## Related Components

- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — application deployments that use Tekton-built images or are triggered by Tekton pipelines
- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — namespaces for workloads that Tekton pipelines build and deploy into
- [KubernetesSecret](/docs/catalog/kubernetes/kubernetessecret) — secrets for Git credentials, container registry tokens, and other pipeline authentication needs
