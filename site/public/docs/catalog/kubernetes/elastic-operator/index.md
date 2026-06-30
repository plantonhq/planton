---
title: "Elastic Operator"
description: "Elastic Operator deployment documentation"
icon: "package"
order: 100
componentName: "kuberneteselasticoperator"
---

# Kubernetes Elastic Operator

Deploys the Elastic Cloud on Kubernetes (ECK) operator via its official Helm chart. ECK extends Kubernetes with custom resource definitions for managing Elasticsearch, Kibana, APM Server, Enterprise Search, Beats, Elastic Agent, Elastic Maps Server, and Fleet Server. Once installed, you can declaratively create and manage Elastic Stack deployments as native Kubernetes resources.

## What Gets Created

When you deploy a KubernetesElasticOperator resource, Planton provisions:

- **Namespace** — created only when `createNamespace` is `true`; defaults to `elastic-system` if `namespace` is not specified
- **ECK Operator Helm Release** — the `eck-operator` chart (v2.14.0) from `https://helm.elastic.co`, which installs:
  - The ECK operator Deployment with configurable CPU and memory limits
  - Custom Resource Definitions (CRDs) for Elasticsearch, Kibana, APM Server, Beats, Elastic Agent, Enterprise Search, Elastic Maps Server, and Fleet Server
  - ServiceAccount, ClusterRole, and ClusterRoleBinding for operator RBAC
  - ValidatingWebhookConfiguration for CRD validation
  - Operator Service for webhook endpoints
- **Kubernetes Labels** — standard Planton labels (`resource`, `resource-name`, `resource-kind`, `resource-id`, `organization`, `environment`) are propagated as inherited labels so that all Elastic Stack resources created by the operator carry consistent metadata

## Prerequisites

- **A Kubernetes cluster** with kubectl configured for access
- **Cluster-admin permissions** to install CRDs and ClusterRoles required by the operator
- **Kubernetes 1.25+** recommended for full ECK 2.14 compatibility

## Quick Start

Create a file `elastic-operator.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesElasticOperator
metadata:
  name: eck
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesElasticOperator.eck
spec:
  namespace:
    value: elastic-system
  createNamespace: true
  container:
    resources:
      limits:
        cpu: "1000m"
        memory: "1Gi"
      requests:
        cpu: "50m"
        memory: "100Mi"
```

Deploy:

```shell
planton apply -f elastic-operator.yaml
```

This installs the ECK operator in the `elastic-system` namespace with default resource allocations. After deployment, you can create Elasticsearch and Kibana custom resources in any namespace.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace for the operator deployment. Use `value` for a direct string or `valueFrom` to reference a KubernetesNamespace resource. Defaults to `elastic-system` when omitted. | Required |
| `container` | `object` | Container specification for the operator pod, including resource limits and requests. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | Create the namespace before deploying the operator. Set to `true` when the namespace does not already exist. |
| `container.resources.limits.cpu` | `string` | `"1000m"` | CPU limit for the ECK operator container. |
| `container.resources.limits.memory` | `string` | `"1Gi"` | Memory limit for the ECK operator container. |
| `container.resources.requests.cpu` | `string` | `"50m"` | CPU request for the ECK operator container. |
| `container.resources.requests.memory` | `string` | `"100Mi"` | Memory request for the ECK operator container. |

## Examples

### Minimal Operator Installation

Install ECK with defaults in the `elastic-system` namespace:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesElasticOperator
metadata:
  name: eck
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesElasticOperator.eck
spec:
  namespace:
    value: elastic-system
  createNamespace: true
  container: {}
```

### Operator with Increased Resources

Deploy the operator with higher resource allocations for clusters running many Elastic Stack instances:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesElasticOperator
metadata:
  name: eck-prod
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesElasticOperator.eck-prod
spec:
  namespace:
    value: elastic-system
  createNamespace: true
  container:
    resources:
      limits:
        cpu: "2000m"
        memory: "2Gi"
      requests:
        cpu: "500m"
        memory: "512Mi"
```

### Operator in a Custom Namespace with Foreign Key Reference

Install the operator in an Planton-managed namespace using a `valueFrom` reference:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesElasticOperator
metadata:
  name: eck-shared
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.KubernetesElasticOperator.eck-shared
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: platform-operators
      field: spec.name
  container:
    resources:
      limits:
        cpu: "1000m"
        memory: "1Gi"
      requests:
        cpu: "100m"
        memory: "256Mi"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where the ECK operator is deployed |
| `service` | `string` | Name of the Kubernetes Service for the ECK operator (format: `{name}-eck-operator`) |
| `portForwardCommand` | `string` | Ready-to-run `kubectl port-forward` command for local access to the operator |
| `kubeEndpoint` | `string` | Cluster-internal endpoint (e.g., `eck-eck-operator.elastic-system.svc.cluster.local`) |
| `ingressEndpoint` | `string` | Public endpoint for external access, when configured |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/namespace) — pre-create a namespace to reference via `valueFrom`
- [KubernetesElasticsearch](/docs/catalog/kubernetes/elasticsearch) — deploy Elasticsearch clusters managed by this operator
- [KubernetesHelmRelease](/docs/catalog/kubernetes/helm-release) — alternative approach for deploying Helm charts with full value control
