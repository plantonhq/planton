---
title: "Solr Operator"
description: "Solr Operator deployment documentation"
icon: "package"
order: 100
componentName: "kubernetessolroperator"
---

# Kubernetes Solr Operator

Deploys the Apache Solr Operator on a Kubernetes cluster using its official Helm chart and CRD manifests. The operator manages the lifecycle of SolrCloud clusters, SolrBackup resources, and SolrPrometheusExporter instances through Kubernetes custom resources, enabling declarative management of Solr infrastructure.

## What Gets Created

When you deploy a KubernetesSolrOperator resource, OpenMCF provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Solr CRDs** — the full set of Solr Operator Custom Resource Definitions (SolrCloud, SolrBackup, SolrPrometheusExporter) downloaded from the official Apache Solr release artifacts for the specified operator version
- **Helm Release** — the `solr-operator` chart from `https://solr.apache.org/charts`, deployed with atomic install, cleanup-on-fail, and wait-for-jobs semantics to ensure a reliable rollout

The CRDs are applied before the Helm release to guarantee that the operator controller starts with all required resource definitions already registered in the cluster API server.

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **Network access** from the cluster to `https://solr.apache.org` for downloading CRD manifests and pulling the Helm chart

## Quick Start

Create a file `solr-operator.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSolrOperator
metadata:
  name: main
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesSolrOperator.main
spec:
  namespace: solr-system
  createNamespace: true
  container:
    resources:
      requests:
        cpu: "50m"
        memory: "100Mi"
      limits:
        cpu: "1000m"
        memory: "1Gi"
```

Deploy:

```shell
openmcf apply -f solr-operator.yaml
```

This installs the Solr Operator v0.9.1 into the `solr-system` namespace with default resource limits. Once the operator is running, you can create SolrCloud clusters using the [KubernetesSolr](/docs/catalog/kubernetes/solr) component.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the Solr Operator deployment. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `container` | `object` | Container specification for the operator pod, including resource requests and limits. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `operatorVersion` | `string` | `v0.9.1` | Version of the Apache Solr Operator to deploy. Must match a tag from the [solr-operator releases](https://github.com/apache/solr-operator/releases). |
| `container.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation for the operator pod. |
| `container.resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation for the operator pod. |
| `container.resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU for the operator pod. |
| `container.resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory for the operator pod. |

> **Note on `namespace`:** This field supports `valueFrom` for referencing outputs of other OpenMCF resources. When using `valueFrom`, specify the `kind`, `name`, and `field` of the source resource instead of a literal string value.

## Examples

### Development Operator with Default Settings

A minimal deployment suitable for development clusters where a single Solr Operator instance manages all SolrCloud resources:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSolrOperator
metadata:
  name: dev-operator
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesSolrOperator.dev-operator
spec:
  namespace: solr-system
  createNamespace: true
  operatorVersion: "v0.9.1"
  container:
    resources:
      requests:
        cpu: "50m"
        memory: "100Mi"
      limits:
        cpu: "500m"
        memory: "512Mi"
```

### Production Operator with Increased Resources

A production-grade deployment with higher resource limits to handle reconciliation of multiple SolrCloud clusters and frequent status updates:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSolrOperator
metadata:
  name: prod-operator
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesSolrOperator.prod-operator
spec:
  namespace: solr-system
  operatorVersion: "v0.9.1"
  container:
    resources:
      requests:
        cpu: "200m"
        memory: "256Mi"
      limits:
        cpu: "2000m"
        memory: "2Gi"
```

### Operator with Foreign Key Namespace Reference

Reference an OpenMCF-managed namespace instead of hardcoding the namespace name:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSolrOperator
metadata:
  name: search-operator
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesSolrOperator.search-operator
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: solr-system-namespace
      field: spec.name
  operatorVersion: "v0.9.1"
  container:
    resources:
      requests:
        cpu: "100m"
        memory: "128Mi"
      limits:
        cpu: "1000m"
        memory: "1Gi"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where the Solr Operator is deployed |
| `service` | `string` | Kubernetes Service name for the Solr Operator (format: `{name}-kubernetes-solr-operator`) |
| `portForwardCommand` | `string` | kubectl port-forward command for local access to the operator |
| `kubeEndpoint` | `string` | Cluster-internal FQDN for the operator service (e.g., `main-kubernetes-solr-operator.solr-system.svc.cluster.local`) |
| `ingressEndpoint` | `string` | Public endpoint for external access, when configured |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/namespace) — provides the target namespace via `valueFrom` reference
- [KubernetesSolr](/docs/catalog/kubernetes/solr) — deploys SolrCloud clusters that depend on the operator installed by this component
- [KubernetesCertManager](/docs/catalog/kubernetes/kubernetescertmanager) — manages TLS certificates for SolrCloud ingress when external access is configured
