---
title: "Redis"
description: "Redis deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesredis"
---

# Kubernetes Redis

Deploys a Redis instance on Kubernetes using the Bitnami Helm chart in standalone architecture, with automatic password generation, optional data persistence via PersistentVolumeClaims, and optional external access through a LoadBalancer Service with external-dns integration.

## What Gets Created

When you deploy a KubernetesRedis resource, OpenMCF provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Random Password** — a 12-character password with mixed case, numbers, and special characters, generated automatically
- **Password Secret** — a Kubernetes Secret storing the base64-encoded password for Redis authentication
- **Helm Release (Bitnami Redis)** — deploys Redis in standalone architecture via the Bitnami Redis Helm chart, with configurable resource limits, replica count, and persistence settings
- **LoadBalancer Service** — created only when ingress is enabled, exposes Redis on port 6379 with an `external-dns.alpha.kubernetes.io/hostname` annotation for automatic DNS record creation

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **A StorageClass** available in the cluster if enabling persistence (most managed Kubernetes clusters provide a default)
- **external-dns** running in the cluster if enabling ingress with a hostname

## Quick Start

Create a file `redis.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesRedis
metadata:
  name: my-redis
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesRedis.my-redis
spec:
  namespace: cache
  createNamespace: true
```

Deploy:

```shell
openmcf apply -f redis.yaml
```

This creates a single-replica Redis instance with persistence enabled, a 1Gi PersistentVolumeClaim, default resource limits (1000m CPU, 1Gi memory), and a randomly generated password stored in a Kubernetes Secret.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the Redis deployment. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `container.replicas` | `int32` | `1` | Number of Redis replica pods. |
| `container.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation for each Redis pod. |
| `container.resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation for each Redis pod. |
| `container.resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU for each Redis pod. |
| `container.resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory for each Redis pod. |
| `container.persistenceEnabled` | `bool` | `true` | Enables persistent storage for Redis data. When enabled, in-memory data is backed up to a PersistentVolumeClaim and restored on pod restart. |
| `container.diskSize` | `string` | `1Gi` | Size of the PersistentVolumeClaim attached to each Redis pod. Required when `persistenceEnabled` is `true`. Must be a valid Kubernetes quantity (e.g., `1Gi`, `10Gi`). Cannot be modified after creation. |
| `ingress.enabled` | `bool` | `false` | Creates a LoadBalancer Service with external-dns annotations exposing Redis on port 6379. |
| `ingress.hostname` | `string` | — | Hostname for external access (e.g., `redis.example.com`). Configured automatically via external-dns. Required when `ingress.enabled` is `true`. |

## Examples

### Development Redis without Persistence

A lightweight Redis instance for development with persistence disabled and reduced resources:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesRedis
metadata:
  name: dev-redis
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesRedis.dev-redis
spec:
  namespace: dev
  createNamespace: true
  container:
    persistenceEnabled: false
    resources:
      limits:
        cpu: "500m"
        memory: "256Mi"
      requests:
        cpu: "100m"
        memory: "128Mi"
```

### Production Redis with Increased Storage

A production Redis instance with larger disk allocation and higher resource limits:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesRedis
metadata:
  name: prod-redis
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesRedis.prod-redis
spec:
  namespace: production
  container:
    replicas: 1
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
    persistenceEnabled: true
    diskSize: "20Gi"
```

### Redis with External Access

Redis exposed outside the cluster via a LoadBalancer with automatic DNS management:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesRedis
metadata:
  name: shared-redis
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesRedis.shared-redis
spec:
  namespace: shared-services
  container:
    replicas: 1
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
    persistenceEnabled: true
    diskSize: "50Gi"
  ingress:
    enabled: true
    hostname: redis.example.com
```

### Using Foreign Key References

Reference an OpenMCF-managed namespace instead of hardcoding the name:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesRedis
metadata:
  name: cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesRedis.cache
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: app-namespace
      field: spec.name
  container:
    persistenceEnabled: true
    diskSize: "10Gi"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where Redis is deployed |
| `service` | `string` | Kubernetes Service name for the Redis master (format: `{name}-master`) |
| `port_forward_command` | `string` | kubectl port-forward command for local access |
| `kube_endpoint` | `string` | Cluster-internal FQDN (e.g., `my-redis-master.cache.svc.cluster.local`) |
| `external_hostname` | `string` | Public hostname for external access, only set when ingress is enabled |
| `internal_hostname` | `string` | Internal hostname for VPC-internal access |
| `username` | `string` | Redis username (always `default`) |
| `password_secret.name` | `string` | Name of the Kubernetes Secret containing the Redis password |
| `password_secret.key` | `string` | Key within the password Secret (always `password`) |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — application deployments that consume Redis as a cache or session store
- [KubernetesExternalDns](/docs/catalog/kubernetes/kubernetesexternaldns) — manages DNS records for the LoadBalancer ingress hostname
