---
title: "Neo4j"
description: "Neo4j deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesneo4j"
---

# Kubernetes Neo4j

Deploys a single-node Neo4j Community instance on Kubernetes using the official Neo4j Helm chart, with automatic admin password generation, optional persistent storage via PersistentVolumeClaims, configurable heap and page-cache memory settings, and optional external access through a LoadBalancer Service with external-dns integration.

## What Gets Created

When you deploy a KubernetesNeo4j resource, OpenMCF provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Helm Release (Neo4j)** — deploys a single-node Neo4j Community instance via the official Neo4j Helm chart (version 2025.03.0), with configurable resource limits, persistence, and memory tuning
- **Password Secret** — the Helm chart automatically generates a Kubernetes Secret named `{name}-auth` containing the admin password under the key `neo4j-password`
- **Kubernetes Service** — an internal ClusterIP Service named `{name}-neo4j` exposing the Bolt (7687) and HTTP (7474) ports
- **LoadBalancer Service** — created only when ingress is enabled, exposes Neo4j externally with an `external-dns.alpha.kubernetes.io/hostname` annotation for automatic DNS record creation
- **PersistentVolumeClaim** — provisions persistent storage for the Neo4j data directory using the cluster default StorageClass

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **A StorageClass** available in the cluster for persistent storage (most managed Kubernetes clusters provide a default)
- **external-dns** running in the cluster if enabling ingress with a hostname

## Quick Start

Create a file `neo4j.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesNeo4j
metadata:
  name: my-graph-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesNeo4j.my-graph-db
spec:
  namespace: graph
  createNamespace: true
```

Deploy:

```shell
openmcf apply -f neo4j.yaml
```

This creates a single-node Neo4j instance with default resource limits (1000m CPU, 1Gi memory), a 1Gi PersistentVolumeClaim, and an auto-generated admin password stored in a Kubernetes Secret.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the Neo4j deployment. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `container.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation for the Neo4j pod. |
| `container.resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation for the Neo4j pod. |
| `container.resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU for the Neo4j pod. |
| `container.resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory for the Neo4j pod. |
| `container.persistenceEnabled` | `bool` | `false` | Enables persistent storage for the Neo4j data directory. When enabled, database files survive pod restarts via a PersistentVolumeClaim. |
| `container.diskSize` | `string` | `1Gi` | Size of the PersistentVolumeClaim for Neo4j data. Must be a valid Kubernetes quantity (e.g., `1Gi`, `10Gi`). |
| `memoryConfig.heapMax` | `string` | — | Maximum Java heap size for Neo4j (e.g., `1Gi`, `512m`). If omitted, Neo4j uses its internal default or auto-detection. |
| `memoryConfig.pageCache` | `string` | — | Page cache size for on-disk data (e.g., `512m`). If omitted, Neo4j uses its internal default or auto-detection. |
| `ingress.enabled` | `bool` | `false` | Creates a LoadBalancer Service with external-dns annotations exposing Neo4j externally. |
| `ingress.hostname` | `string` | — | Hostname for external access (e.g., `neo4j.example.com`). Configured automatically via external-dns. Required when `ingress.enabled` is `true`. |

## Examples

### Development Neo4j without Persistence

A lightweight Neo4j instance for local development and testing with reduced resources and no persistent storage:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesNeo4j
metadata:
  name: dev-graph
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesNeo4j.dev-graph
spec:
  namespace: dev
  createNamespace: true
  container:
    persistenceEnabled: false
    resources:
      limits:
        cpu: "500m"
        memory: "512Mi"
      requests:
        cpu: "100m"
        memory: "256Mi"
```

### Production Neo4j with Tuned Memory

A production-grade Neo4j instance with larger disk, increased resource limits, and explicit heap and page-cache configuration:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesNeo4j
metadata:
  name: prod-graph
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesNeo4j.prod-graph
spec:
  namespace: production
  container:
    resources:
      limits:
        cpu: "4000m"
        memory: "8Gi"
      requests:
        cpu: "1000m"
        memory: "4Gi"
    persistenceEnabled: true
    diskSize: "50Gi"
  memoryConfig:
    heapMax: "4Gi"
    pageCache: "2Gi"
```

### Neo4j with External Access

Neo4j exposed outside the cluster via a LoadBalancer with automatic DNS management:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesNeo4j
metadata:
  name: shared-graph
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesNeo4j.shared-graph
spec:
  namespace: shared-services
  container:
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
    persistenceEnabled: true
    diskSize: "100Gi"
  memoryConfig:
    heapMax: "2Gi"
    pageCache: "1Gi"
  ingress:
    enabled: true
    hostname: neo4j.example.com
```

### Using Foreign Key References

Reference an OpenMCF-managed namespace instead of hardcoding the name:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesNeo4j
metadata:
  name: graph-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesNeo4j.graph-db
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: app-namespace
      field: spec.name
  container:
    persistenceEnabled: true
    diskSize: "20Gi"
  memoryConfig:
    heapMax: "1Gi"
    pageCache: "512m"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where Neo4j is deployed |
| `service` | `string` | Kubernetes Service name for the Neo4j instance (format: `{name}-neo4j`) |
| `bolt_uri_kube_endpoint` | `string` | Cluster-internal FQDN for Bolt connections (e.g., `my-graph-db-neo4j.graph.svc.cluster.local`) |
| `port_forward_command` | `string` | kubectl port-forward command for local access to the Neo4j Browser on port 7474 |
| `external_hostname` | `string` | Public hostname for external access, only set when ingress is enabled |
| `username` | `string` | Neo4j admin username (always `neo4j`) |
| `password_secret.name` | `string` | Name of the Kubernetes Secret containing the admin password (format: `{name}-auth`) |
| `password_secret.key` | `string` | Key within the password Secret (always `neo4j-password`) |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — application deployments that connect to Neo4j as a graph database backend
- [KubernetesExternalDns](/docs/catalog/kubernetes/kubernetesexternaldns) — manages DNS records for the LoadBalancer ingress hostname
