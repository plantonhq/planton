---
title: "Rook Ceph Operator"
description: "Rook Ceph Operator deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesrookcephoperator"
---

# Kubernetes Rook Ceph Operator

Deploys the Rook Ceph Operator on Kubernetes using the official Rook Helm chart (rook-ceph v1.16.6) with support for Ceph CSI RBD (block storage) and CephFS (file storage) drivers, optional NFS driver, configurable CSI host networking, CSI provisioner replica count, CRD management, operator container resource limits and requests, optional namespace creation, and atomic Helm rollback with cleanup-on-fail semantics.

## What Gets Created

When you deploy a KubernetesRookCephOperator resource, OpenMCF provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Helm Release (rook-ceph)** — deploys the Rook Ceph Operator from the `rook-ceph` chart at `https://charts.rook.io/release`, pinned to version v1.16.6 by default, with CRDs installed, atomic rollback enabled, cleanup-on-fail, wait-for-jobs, and a 5-minute timeout
- **CRDs** — Rook custom resource definitions (CephCluster, CephBlockPool, CephFilesystem, etc.) are installed by default via the Helm chart unless explicitly disabled
- **CSI Drivers** — Ceph CSI RBD and CephFS drivers are enabled by default, providing block and file storage support for PersistentVolumeClaims; NFS and CSI Addons drivers are available as optional add-ons

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **Cluster-level permissions** — the Rook operator requires RBAC privileges to manage CRDs, namespaces, pods, and storage resources across the cluster
- **Storage-capable nodes** — at least one node with raw block devices or directories available for Ceph OSDs (required when deploying a CephCluster after the operator is installed)

## Quick Start

Create a file `rook-ceph-operator.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesRookCephOperator
metadata:
  name: my-rook-ceph
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesRookCephOperator.my-rook-ceph
spec:
  namespace: rook-ceph
  createNamespace: true
  container:
    resources:
      requests:
        cpu: 200m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 512Mi
```

Deploy:

```shell
openmcf apply -f rook-ceph-operator.yaml
```

This creates the Rook Ceph Operator in the `rook-ceph` namespace with default CSI drivers enabled and CRDs installed.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the Rook Ceph Operator deployment. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `container` | `object` | Container specifications for the operator deployment. | Required |
| `container.resources` | `object` | CPU and memory resource requests and limits for the operator container. Defaults: requests `200m` CPU / `128Mi` memory, limits `500m` CPU / `512Mi` memory. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `operatorVersion` | `string` | `v1.16.6` | Helm chart version for the Rook Ceph Operator. Must match a valid [Rook release tag](https://github.com/rook/rook/releases). The `v` prefix is stripped automatically before passing to Helm. |
| `crdsEnabled` | `bool` | `true` | Whether the Helm chart should install and update CRDs. Only set to `false` if managing CRDs independently. |
| `csi.enableRbdDriver` | `bool` | `true` | Enable the Ceph CSI RBD (block storage) driver. |
| `csi.enableCephfsDriver` | `bool` | `true` | Enable the Ceph CSI CephFS (file storage) driver. |
| `csi.disableCsiDriver` | `bool` | `false` | Disable the CSI driver entirely. Set to `true` to use an external CSI driver instead. |
| `csi.enableCsiHostNetwork` | `bool` | `true` | Enable host networking for CSI CephFS and RBD nodeplugins. May be necessary when the SDN does not provide access to an external cluster or when there is significant drop in read/write performance. |
| `csi.provisionerReplicas` | `int32` | `2` | Number of replicas for the CSI provisioner deployment. |
| `csi.enableCsiAddons` | `bool` | `false` | Enable CSI Addons for additional CSI functionality such as volume replication and reclaim space. |
| `csi.enableNfsDriver` | `bool` | `false` | Enable the NFS CSI driver for NFS storage support via Ceph NFS gateways. |

> **Note on `valueFrom`**: The `namespace` field is a `StringValueOrRef` type. You can provide a literal string value directly, or use `valueFrom` to reference the output of another OpenMCF resource. See the foreign key reference example below.

## Examples

### Default Deployment with All CSI Drivers

Deploy the Rook Ceph Operator with default settings, enabling both RBD and CephFS drivers:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesRookCephOperator
metadata:
  name: rook-default
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesRookCephOperator.rook-default
spec:
  namespace: rook-ceph
  createNamespace: true
  container:
    resources:
      requests:
        cpu: 200m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 512Mi
```

This deploys the operator at version v1.16.6 with CRDs installed, RBD and CephFS CSI drivers enabled, host networking on, and 2 CSI provisioner replicas. After the operator is running, you can create CephCluster and CephBlockPool resources in the same namespace.

### Production with Custom Resources and NFS

Deploy a production-grade Rook Ceph Operator with increased resource limits, NFS driver, and CSI Addons enabled:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesRookCephOperator
metadata:
  name: rook-prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesRookCephOperator.rook-prod
spec:
  namespace: rook-ceph
  createNamespace: true
  operatorVersion: "v1.16.6"
  container:
    resources:
      requests:
        cpu: 500m
        memory: 256Mi
      limits:
        cpu: "1"
        memory: 1Gi
  csi:
    enableRbdDriver: true
    enableCephfsDriver: true
    enableNfsDriver: true
    enableCsiAddons: true
    enableCsiHostNetwork: true
    provisionerReplicas: 3
```

This deploys the operator with higher CPU and memory allocations suitable for production clusters, enables the NFS driver for NFS-backed PersistentVolumes, activates CSI Addons for features like volume replication, and runs 3 CSI provisioner replicas for high availability.

### Block Storage Only with External CRD Management

Deploy the operator with only the RBD (block storage) driver and manage CRDs externally:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesRookCephOperator
metadata:
  name: rook-block-only
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesRookCephOperator.rook-block-only
spec:
  namespace: rook-ceph
  createNamespace: true
  crdsEnabled: false
  container:
    resources:
      requests:
        cpu: 200m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 512Mi
  csi:
    enableRbdDriver: true
    enableCephfsDriver: false
    enableNfsDriver: false
    enableCsiHostNetwork: false
    provisionerReplicas: 1
```

This deploys a minimal operator for environments that only need Ceph block storage. CRD management is disabled (assumes CRDs are applied separately), CephFS and NFS drivers are turned off, host networking is disabled, and a single provisioner replica reduces resource consumption in staging or development environments.

### Using Foreign Key References

Reference an OpenMCF-managed namespace instead of hardcoding values:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesRookCephOperator
metadata:
  name: rook-with-ref
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesRookCephOperator.rook-with-ref
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: rook-ceph-ns
      field: spec.name
  createNamespace: false
  container:
    resources:
      requests:
        cpu: 200m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 512Mi
```

This example references an OpenMCF-managed namespace rather than embedding a literal value. The `createNamespace` flag is set to `false` because the referenced KubernetesNamespace resource manages the namespace lifecycle.

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where the Rook Ceph Operator is deployed |
| `helmReleaseName` | `string` | Name of the Helm release for the Rook Ceph Operator |
| `webhookService` | `string` | Kubernetes service name for the Rook Ceph Operator webhook (format: `{name}-rook-ceph-operator`) |
| `portForwardCommand` | `string` | Command to set up port-forwarding for accessing operator metrics (e.g., `kubectl port-forward svc/rook-ceph-operator -n rook-ceph 9443:9443`) |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesHelmRelease](/docs/catalog/kubernetes/kuberneteshelmrelease) — alternative for deploying Helm charts with custom configurations
- [KubernetesStatefulSet](/docs/catalog/kubernetes/kubernetesstatefulset) — deploy stateful workloads that consume Ceph-backed PersistentVolumeClaims provisioned by the operator
- [KubernetesPrometheus](/docs/catalog/kubernetes/kubernetesprometheus) — collect metrics from the Rook Ceph Operator and Ceph cluster components
- [KubernetesGrafana](/docs/catalog/kubernetes/kubernetesgrafana) — visualize Ceph cluster health, OSD performance, and storage utilization using Rook-provided dashboards
