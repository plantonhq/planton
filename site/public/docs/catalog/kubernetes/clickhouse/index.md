---
title: "ClickHouse"
description: "ClickHouse deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesclickhouse"
---

# Kubernetes ClickHouse

Deploys a ClickHouse database on Kubernetes using the Altinity ClickHouse Operator, with automatic password generation, optional clustering with sharding and replication, configurable coordination via ClickHouse Keeper or ZooKeeper, persistent storage, and optional external access through a LoadBalancer Service with external-dns integration.

## What Gets Created

When you deploy a KubernetesClickHouse resource, Planton provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Random Password** — a 20-character password with mixed case, numbers, and URL-safe special characters, generated automatically
- **Password Secret** — a Kubernetes Secret storing the password for ClickHouse authentication (key: `admin-password`)
- **ClickHouseKeeperInstallation** — auto-managed ClickHouse Keeper for cluster coordination, created only when clustering is enabled with keeper coordination type (default)
- **ClickHouseInstallation** — the primary ClickHouse deployment managed by the Altinity operator, with configurable resource limits, persistence, cluster layout, logging, and version
- **LoadBalancer Service** — created only when ingress is enabled, exposes ClickHouse on HTTP port 8123 and native protocol port 9000 with an `external-dns.alpha.kubernetes.io/hostname` annotation for automatic DNS record creation

## Prerequisites

- **Kubernetes credentials** configured via environment variables or Planton provider config
- **Altinity ClickHouse Operator** installed in the `clickhouse-operator` namespace on the target cluster
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **A StorageClass** available in the cluster if enabling persistence (most managed Kubernetes clusters provide a default)
- **external-dns** running in the cluster if enabling ingress with a hostname

## Quick Start

Create a file `clickhouse.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesClickHouse
metadata:
  name: my-clickhouse
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesClickHouse.my-clickhouse
spec:
  namespace: analytics
  createNamespace: true
  logging:
    level: information
```

Deploy:

```shell
planton apply -f clickhouse.yaml
```

This creates a single-replica ClickHouse instance with version 24.8, persistence enabled, a 50Gi PersistentVolumeClaim, default resource limits (2000m CPU / 4Gi memory), requests (500m CPU / 1Gi memory), and a randomly generated password stored in a Kubernetes Secret.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the ClickHouse deployment. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `logging.level` | `enum` | Log level for the ClickHouse server. Valid values: `information`, `debug`, `trace`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `clusterName` | `string` | `metadata.name` | Identifier for the ClickHouseInstallation custom resource. Must be a valid DNS subdomain name (lowercase alphanumeric with hyphens). |
| `container.replicas` | `int32` | `1` | Number of ClickHouse replica pods. Ignored when clustering is enabled (use `cluster.shardCount` and `cluster.replicaCount` instead). Must be at least 1. |
| `container.resources.limits.cpu` | `string` | `2000m` | Maximum CPU allocation for each ClickHouse pod. |
| `container.resources.limits.memory` | `string` | `4Gi` | Maximum memory allocation for each ClickHouse pod. |
| `container.resources.requests.cpu` | `string` | `500m` | Minimum guaranteed CPU for each ClickHouse pod. |
| `container.resources.requests.memory` | `string` | `1Gi` | Minimum guaranteed memory for each ClickHouse pod. |
| `container.persistenceEnabled` | `bool` | `true` | Enables persistent storage for ClickHouse data. Strongly recommended for production use. |
| `container.diskSize` | `string` | `50Gi` | Size of the PersistentVolumeClaim attached to each ClickHouse pod. Required when `persistenceEnabled` is `true`. Must be a valid Kubernetes quantity (e.g., `50Gi`, `100Gi`). Cannot be easily modified after creation. |
| `version` | `string` | `24.8` | ClickHouse server version to deploy (e.g., `24.3`, `23.8`). Recommended to pin for production. |
| `ingress.enabled` | `bool` | `false` | Creates a LoadBalancer Service with external-dns annotations exposing ClickHouse on HTTP port 8123 and native port 9000. |
| `ingress.hostname` | `string` | — | Hostname for external access (e.g., `clickhouse.example.com`). Configured automatically via external-dns. Required when `ingress.enabled` is `true`. |
| `cluster.isEnabled` | `bool` | `false` | Enables distributed cluster mode with sharding and replication. When disabled, a single standalone instance is deployed. |
| `cluster.shardCount` | `int32` | — | Number of shards in the cluster. Each shard processes queries in parallel. Must be at least 1 when clustering is enabled. |
| `cluster.replicaCount` | `int32` | — | Number of replicas per shard. Provides data redundancy and high availability. Must be at least 1 when clustering is enabled. Typical values: 2-3. |
| `coordination.type` | `enum` | `keeper` | Coordination service type. Valid values: `keeper` (auto-managed ClickHouse Keeper, recommended), `external_keeper` (existing Keeper cluster), `external_zookeeper` (existing ZooKeeper cluster). Only relevant when clustering is enabled. |
| `coordination.keeperConfig.replicas` | `int32` | `1` | Number of ClickHouse Keeper replicas. Must be an odd number: 1, 3, or 5. Use 3 for production. Only used when `coordination.type` is `keeper`. |
| `coordination.keeperConfig.resources.limits.cpu` | `string` | `500m` | Maximum CPU for each Keeper pod. |
| `coordination.keeperConfig.resources.limits.memory` | `string` | `1Gi` | Maximum memory for each Keeper pod. |
| `coordination.keeperConfig.resources.requests.cpu` | `string` | `100m` | Minimum guaranteed CPU for each Keeper pod. |
| `coordination.keeperConfig.resources.requests.memory` | `string` | `256Mi` | Minimum guaranteed memory for each Keeper pod. |
| `coordination.keeperConfig.diskSize` | `string` | `10Gi` | Persistent volume size for each Keeper pod. Stores coordination metadata only. |
| `coordination.externalConfig.nodes` | `string[]` | — | List of external coordination nodes in `host:port` format. Required when `coordination.type` is `external_keeper` or `external_zookeeper`. |

## Examples

### Development ClickHouse with Reduced Resources

A lightweight ClickHouse instance for development with smaller resource allocations:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesClickHouse
metadata:
  name: dev-clickhouse
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesClickHouse.dev-clickhouse
spec:
  namespace: dev
  createNamespace: true
  version: "24.8"
  container:
    replicas: 1
    resources:
      limits:
        cpu: "1000m"
        memory: "2Gi"
      requests:
        cpu: "250m"
        memory: "512Mi"
    persistenceEnabled: true
    diskSize: "10Gi"
  logging:
    level: information
```

### Production Clustered ClickHouse

A distributed ClickHouse cluster with 2 shards and 2 replicas per shard for high availability and horizontal scaling, using auto-managed ClickHouse Keeper with 3 replicas:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesClickHouse
metadata:
  name: prod-clickhouse
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesClickHouse.prod-clickhouse
spec:
  namespace: analytics
  version: "24.8"
  container:
    resources:
      limits:
        cpu: "4000m"
        memory: "16Gi"
      requests:
        cpu: "2000m"
        memory: "8Gi"
    persistenceEnabled: true
    diskSize: "200Gi"
  cluster:
    isEnabled: true
    shardCount: 2
    replicaCount: 2
  coordination:
    type: keeper
    keeperConfig:
      replicas: 3
      diskSize: "10Gi"
  logging:
    level: information
```

### ClickHouse with External Access

ClickHouse exposed outside the cluster via a LoadBalancer with automatic DNS management:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesClickHouse
metadata:
  name: shared-clickhouse
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesClickHouse.shared-clickhouse
spec:
  namespace: shared-services
  container:
    replicas: 1
    resources:
      limits:
        cpu: "2000m"
        memory: "8Gi"
      requests:
        cpu: "500m"
        memory: "2Gi"
    persistenceEnabled: true
    diskSize: "100Gi"
  ingress:
    enabled: true
    hostname: clickhouse.example.com
  logging:
    level: information
```

### Clustered ClickHouse with External ZooKeeper

A clustered deployment using an existing ZooKeeper ensemble shared with other services:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesClickHouse
metadata:
  name: analytics-clickhouse
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesClickHouse.analytics-clickhouse
spec:
  namespace: analytics
  version: "24.8"
  container:
    resources:
      limits:
        cpu: "4000m"
        memory: "16Gi"
      requests:
        cpu: "1000m"
        memory: "4Gi"
    persistenceEnabled: true
    diskSize: "500Gi"
  cluster:
    isEnabled: true
    shardCount: 4
    replicaCount: 2
  coordination:
    type: external_zookeeper
    externalConfig:
      nodes:
        - "zk-0.zk.shared-infra.svc.cluster.local:2181"
        - "zk-1.zk.shared-infra.svc.cluster.local:2181"
        - "zk-2.zk.shared-infra.svc.cluster.local:2181"
  logging:
    level: information
```

### Using Foreign Key References

Reference an Planton-managed namespace instead of hardcoding the name:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesClickHouse
metadata:
  name: events-clickhouse
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesClickHouse.events-clickhouse
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: analytics-namespace
      field: spec.name
  container:
    persistenceEnabled: true
    diskSize: "50Gi"
  logging:
    level: information
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where ClickHouse is deployed |
| `service` | `string` | Kubernetes Service name for the ClickHouse instance (format: `{name}`) |
| `port_forward_command` | `string` | kubectl port-forward command for local access on port 8123 |
| `kube_endpoint` | `string` | Cluster-internal FQDN (e.g., `my-clickhouse.analytics.svc.cluster.local:8123`) |
| `external_hostname` | `string` | Public hostname for external access, only set when ingress is enabled |
| `internal_hostname` | `string` | Internal hostname for VPC-internal access |
| `username` | `string` | ClickHouse username (always `default`) |
| `password_secret.name` | `string` | Name of the Kubernetes Secret containing the ClickHouse password (format: `{name}-password`) |
| `password_secret.key` | `string` | Key within the password Secret (always `admin-password`) |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/namespace) — provides the target namespace via `valueFrom` reference
- [KubernetesDeployment](/docs/catalog/kubernetes/deployment) — application deployments that query ClickHouse for analytics
- [KubernetesExternalDns](/docs/catalog/kubernetes/external-dns) — manages DNS records for the LoadBalancer ingress hostname
