# Kubernetes Tekton Operator

Deploys the Tekton Operator on Kubernetes to manage the lifecycle of Tekton components including Pipelines, Triggers, and Dashboard. The operator-based approach uses the official Tekton Operator release manifests and a TektonConfig custom resource to declaratively select which components to install, automatically choosing the correct operator profile (`lite`, `basic`, or `all`). Optional features include CloudEvents integration for pipeline event notifications and external dashboard access through Istio Gateway API ingress with automatic TLS via cert-manager.

## What Gets Created

When you deploy a KubernetesTektonOperator resource, Planton provisions:

- **Tekton Operator** — all resources from the official Tekton Operator release manifest installed into the fixed `tekton-operator` namespace, including the operator deployment, CRDs (`TektonConfig`, `TektonPipeline`, `TektonTrigger`, `TektonDashboard`, etc.), RBAC roles, and webhook configurations
- **TektonConfig Custom Resource** — a `TektonConfig` resource named `config` that tells the operator which components to install; the operator selects the `all` profile when Pipelines, Triggers, and Dashboard are all enabled, the `basic` profile when Pipelines and Triggers are enabled, or the `lite` profile otherwise
- **Tekton Pipelines** — the core CI/CD pipeline engine installed by the operator into the fixed `tekton-pipelines` namespace; created when `components.pipelines` is `true`
- **Tekton Triggers** — event-driven pipeline execution support for webhooks and external events; created when `components.triggers` is `true`
- **Tekton Dashboard** — the web UI for viewing and managing pipelines, tasks, and runs; created when `components.dashboard` is `true`
- **TLS Certificate** — a cert-manager Certificate for the dashboard ingress hostname in the `istio-ingress` namespace; created only when both `components.dashboard` and `dashboardIngress.enabled` are `true`
- **Istio Gateway** — an external Gateway resource with HTTPS (port 443) and HTTP (port 80) listeners for the dashboard; created only when dashboard ingress is enabled
- **HTTP-to-HTTPS Redirect Route** — an HTTPRoute that redirects HTTP traffic to HTTPS with a 301 status code; created only when dashboard ingress is enabled
- **HTTPS Route** — an HTTPRoute that forwards HTTPS traffic to the Tekton Dashboard service on port 9097; created only when dashboard ingress is enabled

## Prerequisites

- **Kubernetes credentials** configured via environment variables or Planton provider config
- **Cluster admin access** because the Tekton Operator installs cluster-scoped CRDs and RBAC resources
- **Istio** with Gateway API support installed if enabling dashboard ingress
- **cert-manager** with a ClusterIssuer matching the ingress domain if enabling dashboard ingress with TLS

## Quick Start

Create a file `tekton-operator.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTektonOperator
metadata:
  name: my-tekton-operator
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesTektonOperator.my-tekton-operator
spec:
  container:
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 512Mi
  components:
    pipelines: true
```

Deploy:

```shell
planton apply -f tekton-operator.yaml
```

This installs the Tekton Operator (default version v0.78.0), which in turn deploys Tekton Pipelines into the fixed `tekton-pipelines` namespace.

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `container.resources` | `object` | CPU and memory resource requests and limits for the operator container. Defaults: requests `100m` CPU / `128Mi` memory, limits `500m` CPU / `512Mi` memory. |
| `components` | `object` | Which Tekton components to install. At least one of `pipelines`, `triggers`, or `dashboard` must be `true`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `operatorVersion` | `string` | `v0.78.0` | Version of the Tekton Operator to deploy. Maps to releases at https://github.com/tektoncd/operator/releases. |
| `components.pipelines` | `bool` | `false` | Enable Tekton Pipelines for running CI/CD pipelines. |
| `components.triggers` | `bool` | `false` | Enable Tekton Triggers for event-driven pipeline execution via webhooks. |
| `components.dashboard` | `bool` | `false` | Enable Tekton Dashboard for a web-based UI to view and manage pipelines. |
| `dashboardIngress.enabled` | `bool` | `false` | Enable external access to the dashboard through Istio Gateway API with TLS termination and HTTP-to-HTTPS redirect. Requires `components.dashboard` to also be `true`. |
| `dashboardIngress.hostname` | `string` | — | Full hostname for external access to the dashboard (e.g., `tekton-dashboard.example.com`). The ClusterIssuer name is derived from the domain portion of the hostname. Required when `dashboardIngress.enabled` is `true`. |
| `cloudEventsSinkUrl` | `string` | — | URL where CloudEvents will be sent for TaskRun and PipelineRun lifecycle events. Configured as the `default-cloud-events-sink` in TektonConfig. Must be a valid HTTP or HTTPS URL (e.g., `http://my-receiver.my-namespace.svc.cluster.local/tekton/events`). |

## Examples

### Pipelines Only

A minimal deployment that installs Tekton Pipelines through the operator with a pinned version:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTektonOperator
metadata:
  name: ci-tekton-operator
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesTektonOperator.ci-tekton-operator
spec:
  operatorVersion: v0.78.0
  container:
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 512Mi
  components:
    pipelines: true
```

The operator installs with the `lite` profile and deploys only Tekton Pipelines into the `tekton-pipelines` namespace.

### Pipelines, Triggers, and Dashboard

A full Tekton stack with all three components, suitable for teams that need event-driven pipeline triggers and a web UI:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTektonOperator
metadata:
  name: team-tekton-operator
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.KubernetesTektonOperator.team-tekton-operator
spec:
  operatorVersion: v0.78.0
  container:
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 512Mi
  components:
    pipelines: true
    triggers: true
    dashboard: true
```

The operator installs with the `all` profile. After deployment, access the dashboard locally:

```shell
kubectl port-forward svc/tekton-dashboard -n tekton-pipelines 9097:9097
```

### Production with Dashboard Ingress and CloudEvents

A production setup with the dashboard exposed externally via TLS-terminated Istio Gateway ingress and CloudEvents integration for pipeline notifications:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTektonOperator
metadata:
  name: prod-tekton-operator
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesTektonOperator.prod-tekton-operator
spec:
  operatorVersion: v0.78.0
  container:
    resources:
      requests:
        cpu: 200m
        memory: 256Mi
      limits:
        cpu: "1"
        memory: 1Gi
  components:
    pipelines: true
    triggers: true
    dashboard: true
  dashboardIngress:
    enabled: true
    hostname: tekton-dashboard.example.com
  cloudEventsSinkUrl: http://event-router.platform.svc.cluster.local/tekton/events
```

This creates Certificate, Gateway, and HTTPRoute resources in addition to the full Tekton stack. The ClusterIssuer name is automatically derived from the hostname domain (`example.com` in this case).

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Namespace where Tekton components are installed (always `tekton-pipelines`). |
| `tektonConfigName` | `string` | Name of the TektonConfig custom resource created by the operator (always `config`). |
| `pipelinesControllerService` | `string` | Kubernetes service name for the Tekton Pipelines controller (`tekton-pipelines-controller`). Empty if pipelines component is not enabled. |
| `triggersControllerService` | `string` | Kubernetes service name for the Tekton Triggers controller (`tekton-triggers-controller`). Empty if triggers component is not enabled. |
| `dashboardService` | `string` | Kubernetes service name for the Tekton Dashboard (`tekton-dashboard`). Empty if dashboard component is not enabled. |
| `dashboardPortForwardCommand` | `string` | kubectl port-forward command for local access to the dashboard on port 9097. Empty if dashboard is not enabled. |

## Related Components

- [KubernetesTekton](/docs/catalog/kubernetes/kubernetestekton) — manifest-based Tekton deployment that applies Pipeline and Dashboard release YAMLs directly without the operator
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — application deployments that use Tekton-built images or are triggered by Tekton pipelines
- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — namespaces for workloads that Tekton pipelines build and deploy into
- [KubernetesSecret](/docs/catalog/kubernetes/kubernetessecret) — secrets for Git credentials, container registry tokens, and other pipeline authentication needs
- [KubernetesIstio](/docs/catalog/kubernetes/kubernetesistio) — Istio service mesh required for dashboard ingress via Gateway API
- [KubernetesCertManager](/docs/catalog/kubernetes/kubernetescertmanager) — cert-manager for automatic TLS certificate provisioning on dashboard ingress
