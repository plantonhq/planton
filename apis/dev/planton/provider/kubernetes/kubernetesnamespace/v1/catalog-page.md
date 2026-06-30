# Kubernetes Namespace

Deploys a Kubernetes namespace with optional resource quotas, LimitRanges, network policies, service mesh sidecar injection, and Pod Security Standards enforcement. Configuration is abstracted into T-shirt-sized resource profiles and declarative network isolation rules, so each namespace is created with resource limits and network isolation configured from the start.

## What Gets Created

When you deploy a KubernetesNamespace resource, Planton provisions:

- **Namespace** — a Kubernetes Namespace with merged labels (user-specified plus standard management labels) and annotations (user-specified plus service-mesh annotations when enabled)
- **ResourceQuota** — enforces CPU, memory, and object-count limits for the namespace, created only when a `resourceProfile` preset or custom quota is configured
- **LimitRange** — sets default CPU and memory requests/limits for containers that do not declare their own, created only when custom `defaultLimits` are provided
- **Ingress NetworkPolicy** — a default-deny ingress policy that allows traffic only from pods within the namespace and from explicitly listed namespaces, created only when `networkConfig.isolateIngress` is `true`
- **Egress NetworkPolicy** — a default-deny egress policy that allows DNS traffic to kube-system, traffic within the namespace, and traffic to explicitly listed CIDR blocks, created only when `networkConfig.restrictEgress` is `true`

## Prerequisites

- **Kubernetes credentials** configured via environment variables or Planton provider config
- **A running Kubernetes cluster** reachable from the deployment environment
- **A CNI plugin that supports NetworkPolicy** (e.g., Calico, Cilium) if using ingress isolation or egress restriction
- **A service mesh control plane** (Istio, Linkerd, or Consul Connect) already installed on the cluster if enabling service mesh injection

## Quick Start

Create a file `namespace.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesNamespace
metadata:
  name: my-namespace
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesNamespace.my-namespace
spec:
  name: my-namespace
```

Deploy:

```shell
planton apply -f namespace.yaml
```

This creates a bare namespace named `my-namespace` with standard Planton management labels. No quotas, network policies, or service mesh injection are applied.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `spec.name` | `string` | The Kubernetes namespace name. Used as `metadata.name` on the created Namespace. | 1–63 characters, valid DNS label (lowercase alphanumeric and hyphens, no leading/trailing hyphens) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `spec.targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `spec.targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `spec.labels` | `map<string, string>` | `{}` | Additional labels merged onto the Namespace. Standard management labels (`managed-by`, `resource`, `resource-kind`) are always added. |
| `spec.annotations` | `map<string, string>` | `{}` | Additional annotations merged onto the Namespace. Service-mesh injection annotations are added automatically when `serviceMeshConfig` is enabled. |
| `spec.resourceProfile.preset` | `enum` | — | T-shirt-sized resource profile. One of `small`, `medium`, `large`, `xlarge`. Mutually exclusive with `resourceProfile.custom`. See table below. |
| `spec.resourceProfile.custom.cpu.requests` | `string` | — | Total CPU requests quota (e.g., `"4"`, `"4000m"`). |
| `spec.resourceProfile.custom.cpu.limits` | `string` | — | Total CPU limits quota (e.g., `"8"`, `"8000m"`). |
| `spec.resourceProfile.custom.memory.requests` | `string` | — | Total memory requests quota (e.g., `"8Gi"`). |
| `spec.resourceProfile.custom.memory.limits` | `string` | — | Total memory limits quota (e.g., `"16Gi"`). |
| `spec.resourceProfile.custom.objectCounts.pods` | `int32` | — | Maximum number of pods. Must be >= 1. |
| `spec.resourceProfile.custom.objectCounts.services` | `int32` | — | Maximum number of services. Must be >= 1. |
| `spec.resourceProfile.custom.objectCounts.configmaps` | `int32` | — | Maximum number of ConfigMaps. Must be >= 1. |
| `spec.resourceProfile.custom.objectCounts.secrets` | `int32` | — | Maximum number of Secrets. Must be >= 1. |
| `spec.resourceProfile.custom.objectCounts.persistentVolumeClaims` | `int32` | — | Maximum number of PVCs. Must be >= 0. |
| `spec.resourceProfile.custom.objectCounts.loadBalancers` | `int32` | — | Maximum number of LoadBalancer services. Must be >= 0. |
| `spec.resourceProfile.custom.defaultLimits.defaultCpuRequest` | `string` | — | Default CPU request injected into containers without explicit values (e.g., `"100m"`). Triggers LimitRange creation. |
| `spec.resourceProfile.custom.defaultLimits.defaultCpuLimit` | `string` | — | Default CPU limit (e.g., `"1000m"`). |
| `spec.resourceProfile.custom.defaultLimits.defaultMemoryRequest` | `string` | — | Default memory request (e.g., `"128Mi"`). |
| `spec.resourceProfile.custom.defaultLimits.defaultMemoryLimit` | `string` | — | Default memory limit (e.g., `"512Mi"`). |
| `spec.networkConfig.isolateIngress` | `bool` | `false` | When `true`, creates a NetworkPolicy that denies all ingress except from the same namespace and from `allowedIngressNamespaces`. |
| `spec.networkConfig.restrictEgress` | `bool` | `false` | When `true`, creates a NetworkPolicy that denies all egress except DNS to kube-system, same-namespace traffic, and `allowedEgressCidrs`. |
| `spec.networkConfig.allowedIngressNamespaces` | `string[]` | `[]` | Namespace names permitted to send ingress traffic to this namespace. Only effective when `isolateIngress` is `true`. |
| `spec.networkConfig.allowedEgressCidrs` | `string[]` | `[]` | CIDR blocks that pods may reach (e.g., `"10.0.0.0/8"`). Only effective when `restrictEgress` is `true`. |
| `spec.networkConfig.allowedEgressDomains` | `string[]` | `[]` | DNS domains that pods may reach (e.g., `"api.stripe.com"`). Requires a CNI with DNS-based policy support (Calico or Cilium). |
| `spec.serviceMeshConfig.enabled` | `bool` | `false` | When `true`, adds mesh-specific sidecar injection annotations to the namespace. Requires `meshType` to be set. |
| `spec.serviceMeshConfig.meshType` | `enum` | — | Service mesh type. One of `istio`, `linkerd`, `consul`. Required when `serviceMeshConfig.enabled` is `true`. |
| `spec.serviceMeshConfig.revisionTag` | `string` | — | Istio revision tag (e.g., `"prod-stable"`, `"1-19-5"`). Istio-specific; sets `istio.io/rev` instead of the global `istio-injection` annotation. Max 63 characters. |
| `spec.podSecurityStandard` | `enum` | unspecified | Pod Security Standards enforcement level. One of `privileged`, `baseline`, `restricted`. Sets the `pod-security.kubernetes.io/enforce` label on the namespace. |

**Built-in resource profile sizes:**

| Preset | CPU Req/Limit | Memory Req/Limit | Pods | Services | ConfigMaps | Secrets | PVCs | LBs |
|--------|---------------|------------------|------|----------|------------|---------|------|-----|
| `small` | 2 / 4 | 4Gi / 8Gi | 20 | 10 | 50 | 50 | 5 | 2 |
| `medium` | 4 / 8 | 8Gi / 16Gi | 50 | 20 | 100 | 100 | 10 | 3 |
| `large` | 8 / 16 | 16Gi / 32Gi | 100 | 40 | 200 | 200 | 20 | 5 |
| `xlarge` | 16 / 32 | 32Gi / 64Gi | 200 | 80 | 400 | 400 | 40 | 10 |

## Examples

### Basic Namespace with a Resource Profile

A development namespace with the `small` resource profile to set guardrails on resource consumption:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesNamespace
metadata:
  name: dev-team-alpha
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesNamespace.dev-team-alpha
spec:
  name: dev-team-alpha
  labels:
    team: alpha
    cost-center: engineering
  resourceProfile:
    preset: small
```

### Network-Isolated Namespace with Pod Security

A staging namespace that locks down both ingress and egress, allows traffic from a shared `monitoring` namespace, permits outbound access to an internal subnet, and enforces the `baseline` Pod Security Standard:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesNamespace
metadata:
  name: staging-backend
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.KubernetesNamespace.staging-backend
spec:
  name: staging-backend
  labels:
    team: backend
    environment: staging
  resourceProfile:
    preset: medium
  networkConfig:
    isolateIngress: true
    restrictEgress: true
    allowedIngressNamespaces:
      - monitoring
      - istio-system
    allowedEgressCidrs:
      - "10.0.0.0/8"
  podSecurityStandard: baseline
```

### Full-Featured Production Namespace with Service Mesh and Custom Quotas

A production namespace with Istio sidecar injection pinned to a specific revision, custom resource quotas and default container limits, full network isolation, and the `restricted` Pod Security Standard:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesNamespace
metadata:
  name: prod-payments
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesNamespace.prod-payments
spec:
  name: prod-payments
  labels:
    team: payments
    compliance: pci-dss
  annotations:
    janitor/ttl: "never"
  targetCluster:
    clusterKind: GcpGkeCluster
    clusterName: prod-us-central1
  resourceProfile:
    custom:
      cpu:
        requests: "12"
        limits: "24"
      memory:
        requests: "24Gi"
        limits: "48Gi"
      objectCounts:
        pods: 150
        services: 30
        configmaps: 200
        secrets: 200
        persistentVolumeClaims: 50
        loadBalancers: 5
      defaultLimits:
        defaultCpuRequest: "100m"
        defaultCpuLimit: "2000m"
        defaultMemoryRequest: "128Mi"
        defaultMemoryLimit: "1Gi"
  networkConfig:
    isolateIngress: true
    restrictEgress: true
    allowedIngressNamespaces:
      - istio-system
      - monitoring
      - logging
    allowedEgressCidrs:
      - "10.0.0.0/8"
      - "172.16.0.0/12"
  serviceMeshConfig:
    enabled: true
    meshType: istio
    revisionTag: prod-stable
  podSecurityStandard: restricted
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Name of the created Kubernetes namespace |
| `namespaceId` | `string` | Fully qualified namespace identifier (same as `namespace`) |
| `resourceQuotasApplied` | `string` | `"true"` if ResourceQuota objects were created |
| `limitRangesApplied` | `string` | `"true"` if LimitRange objects were created |
| `networkPoliciesApplied` | `string` | `"true"` if NetworkPolicy objects were created |
| `serviceMeshEnabled` | `string` | `"true"` if the namespace is configured for automatic sidecar injection |
| `serviceMeshType` | `string` | Configured mesh type (`istio`, `linkerd`, `consul`, or empty) |
| `podSecurityStandard` | `string` | Applied Pod Security Standard level (`privileged`, `baseline`, `restricted`, or empty) |
| `labelsJson` | `string` | JSON representation of all labels applied to the namespace |
| `annotationsJson` | `string` | JSON representation of all annotations applied to the namespace |

## Related Components

- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — deploys containerized workloads into namespaces managed by this component
- [KubernetesPostgres](/docs/catalog/kubernetes/kubernetespostgres) — deploys PostgreSQL into a namespace, often co-deployed alongside application namespaces
- [KubernetesRedis](/docs/catalog/kubernetes/kubernetesredis) — deploys Redis into a namespace for caching and pub/sub workloads
