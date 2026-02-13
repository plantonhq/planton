---
title: "Strimzi Kafka Operator"
description: "Strimzi Kafka Operator deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesstrimzikafkaoperator"
---

# Kubernetes Strimzi Kafka Operator

Deploys the Strimzi Kafka Operator on a Kubernetes cluster using the official Strimzi Helm chart (v0.42.0). Strimzi extends Kubernetes with custom resource definitions for managing Apache Kafka clusters, topics, users, connectors, and mirrors. Once installed, the operator watches all namespaces and lets you declaratively create and manage Kafka deployments as native Kubernetes resources.

## What Gets Created

When you deploy a KubernetesStrimziKafkaOperator resource, OpenMCF provisions:

- **Namespace** — created only when `createNamespace` is `true`; defaults to `strimzi-operator` if `namespace` is not specified
- **Strimzi Helm Release** — the `kubernetes-strimzi-kafka-operator` chart (v0.42.0) from `https://strimzi.io/charts/`, configured with `watchAnyNamespace: true`, which installs:
  - The Strimzi Cluster Operator Deployment with configurable CPU and memory limits
  - Custom Resource Definitions (CRDs) for Kafka, KafkaTopic, KafkaUser, KafkaConnect, KafkaMirrorMaker2, KafkaBridge, KafkaRebalance, and KafkaNodePool
  - ServiceAccount, ClusterRole, and ClusterRoleBinding for operator RBAC
  - ValidatingWebhookConfiguration for CRD validation
- **Kubernetes Labels** — standard OpenMCF labels (`resource-kind`, `resource-id`, `organization`, `environment`) are applied to the namespace and propagated to operator resources for consistent metadata

## Prerequisites

- **A Kubernetes cluster** with kubectl configured for access
- **Cluster-admin permissions** to install CRDs and ClusterRoles required by the operator
- **Kubernetes 1.25+** recommended for full Strimzi 0.42 compatibility

## Quick Start

Create a file `strimzi-operator.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesStrimziKafkaOperator
metadata:
  name: strimzi
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesStrimziKafkaOperator.strimzi
spec:
  namespace:
    value: strimzi-operator
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
openmcf apply -f strimzi-operator.yaml
```

This installs the Strimzi Kafka Operator in the `strimzi-operator` namespace with default resource allocations. After deployment, you can create Kafka, KafkaTopic, and KafkaUser custom resources in any namespace.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace for the operator deployment. Use `value` for a direct string or `valueFrom` to reference a KubernetesNamespace resource. Defaults to `strimzi-operator` when omitted. | Required |
| `container` | `object` | Container specification for the operator pod, including resource limits and requests. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | Create the namespace before deploying the operator. Set to `true` when the namespace does not already exist. |
| `container.resources.limits.cpu` | `string` | `"1000m"` | CPU limit for the Strimzi operator container. |
| `container.resources.limits.memory` | `string` | `"1Gi"` | Memory limit for the Strimzi operator container. |
| `container.resources.requests.cpu` | `string` | `"50m"` | CPU request for the Strimzi operator container. |
| `container.resources.requests.memory` | `string` | `"100Mi"` | Memory request for the Strimzi operator container. |

> **Note on `valueFrom`**: The `namespace` field is a `StringValueOrRef` type. You can provide a literal string with `value`, or use `valueFrom` to reference the output of another OpenMCF resource (e.g., a KubernetesNamespace). See the foreign key reference example below.

## Examples

### Minimal Operator Installation

Install the Strimzi Kafka Operator with defaults in the `strimzi-operator` namespace:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesStrimziKafkaOperator
metadata:
  name: strimzi
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesStrimziKafkaOperator.strimzi
spec:
  namespace:
    value: strimzi-operator
  createNamespace: true
  container: {}
```

### Operator with Increased Resources

Deploy the operator with higher resource allocations for clusters running many Kafka instances:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesStrimziKafkaOperator
metadata:
  name: strimzi-prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesStrimziKafkaOperator.strimzi-prod
spec:
  namespace:
    value: strimzi-operator
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

Install the operator in an OpenMCF-managed namespace using a `valueFrom` reference:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesStrimziKafkaOperator
metadata:
  name: strimzi-shared
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesStrimziKafkaOperator.strimzi-shared
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
| `namespace` | `string` | Kubernetes namespace where the Strimzi Kafka Operator is deployed |
| `service` | `string` | Name of the Kubernetes Service for the Strimzi operator (format: `{name}-kubernetes-strimzi-kafka-operator`) |
| `portForwardCommand` | `string` | Ready-to-run `kubectl port-forward` command for local access to the operator |
| `kubeEndpoint` | `string` | Cluster-internal endpoint (e.g., `strimzi.strimzi-operator.svc.cluster.local`) |
| `ingressEndpoint` | `string` | Public endpoint for external access, when configured |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — pre-create a namespace to reference via `valueFrom`
- [KubernetesKafka](/docs/catalog/kubernetes/kuberneteskafka) — deploy Kafka clusters managed by this operator
- [KubernetesHelmRelease](/docs/catalog/kubernetes/kuberneteshelmrelease) — alternative approach for deploying Helm charts with full value control
