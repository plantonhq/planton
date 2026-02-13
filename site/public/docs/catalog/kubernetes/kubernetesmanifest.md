---
title: "Manifest"
description: "Manifest deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesmanifest"
---

# Kubernetes Manifest

Deploys arbitrary Kubernetes YAML manifests -- single or multi-document -- through OpenMCF's lifecycle management, giving raw Kubernetes resources the same declarative apply/update/destroy workflow as any other OpenMCF component. The module handles namespace creation, multi-document ordering, and CRD dependency resolution automatically.

## What Gets Created

When you deploy a KubernetesManifest resource, OpenMCF provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **All resources defined in `manifestYaml`** — every Kubernetes resource in the provided YAML (single or multi-document) is applied through Pulumi's `yaml/v2.ConfigGroup`, which handles CRD ordering, multi-document splitting, and dependency tracking automatically

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **Valid Kubernetes manifest YAML** containing one or more resource definitions separated by `---`
- **CRD definitions available on the cluster** if the manifest references custom resource types

## Quick Start

Create a file `manifest.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesManifest
metadata:
  name: my-manifest
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesManifest.my-manifest
spec:
  namespace:
    value: my-namespace
  manifestYaml: |
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: my-config
    data:
      key: value
```

Deploy:

```shell
openmcf apply -f manifest.yaml
```

This applies the ConfigMap to the `my-namespace` namespace. Resources in the manifest that specify their own namespace use that namespace; resources without one use the namespace from `spec.namespace`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace for resources that do not specify their own namespace. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `manifestYaml` | `string` | Raw Kubernetes manifest YAML to deploy. Supports single or multi-document YAML separated by `---`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before applying the manifest. When `false`, the namespace must already exist. |

## Examples

### Basic ConfigMap

A single ConfigMap deployed through OpenMCF lifecycle management:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesManifest
metadata:
  name: app-config
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesManifest.app-config
spec:
  namespace:
    value: default
  manifestYaml: |
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: app-settings
    data:
      LOG_LEVEL: "info"
      MAX_CONNECTIONS: "100"
```

### Multi-Document Manifest with Namespace Creation

A RBAC setup deploying a ServiceAccount, Role, and RoleBinding in a newly created namespace:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesManifest
metadata:
  name: rbac-setup
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesManifest.rbac-setup
spec:
  namespace:
    value: monitoring
  createNamespace: true
  manifestYaml: |
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: prometheus
      namespace: monitoring
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: Role
    metadata:
      name: prometheus-reader
      namespace: monitoring
    rules:
      - apiGroups: [""]
        resources: ["pods", "services", "endpoints"]
        verbs: ["get", "list", "watch"]
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: RoleBinding
    metadata:
      name: prometheus-reader-binding
      namespace: monitoring
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: Role
      name: prometheus-reader
    subjects:
      - kind: ServiceAccount
        name: prometheus
        namespace: monitoring
```

### Full Application Stack with CRDs and Cross-Namespace Resources

A complete application manifest deploying a CronJob, a NetworkPolicy, and a PriorityClass (cluster-scoped), referencing a namespace managed by another OpenMCF resource:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesManifest
metadata:
  name: batch-stack
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesManifest.batch-stack
spec:
  targetCluster:
    clusterKind: GcpGkeCluster
    clusterName: prod-cluster
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: batch-ns
      field: spec.name
  manifestYaml: |
    apiVersion: scheduling.k8s.io/v1
    kind: PriorityClass
    metadata:
      name: batch-high-priority
    value: 1000000
    globalDefault: false
    description: "Priority class for batch jobs"
    ---
    apiVersion: networking.k8s.io/v1
    kind: NetworkPolicy
    metadata:
      name: batch-network-policy
    spec:
      podSelector:
        matchLabels:
          app: batch-worker
      policyTypes:
        - Ingress
        - Egress
      egress:
        - to:
            - namespaceSelector:
                matchLabels:
                  name: database
          ports:
            - protocol: TCP
              port: 5432
    ---
    apiVersion: batch/v1
    kind: CronJob
    metadata:
      name: data-sync
    spec:
      schedule: "0 */6 * * *"
      jobTemplate:
        spec:
          template:
            spec:
              priorityClassName: batch-high-priority
              containers:
                - name: sync
                  image: gcr.io/my-project/data-sync:v1.2.0
                  resources:
                    limits:
                      cpu: "2000m"
                      memory: "4Gi"
                    requests:
                      cpu: "500m"
                      memory: "1Gi"
              restartPolicy: OnFailure
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where the manifest resources were deployed |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — preferred for deploying containerized applications with built-in Service, ingress, and autoscaling support
- [KubernetesHelmRelease](/docs/catalog/kubernetes/kuberneteshelmrelease) — preferred when deploying packaged Helm charts rather than raw YAML
