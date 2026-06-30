# Kubernetes MongoDB

Deploys a MongoDB instance on Kubernetes using the Percona Server for MongoDB operator, with automatic password generation, configurable replica sets, optional data persistence via PersistentVolumeClaims, and optional external access through a LoadBalancer Service with external-dns integration.

## What Gets Created

When you deploy a KubernetesMongodb resource, Planton provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Random Password** — a 12-character password with mixed case, numbers, and special characters, generated automatically
- **Password Secret** — a Kubernetes Secret storing the generated password for MongoDB authentication
- **PerconaServerMongoDB CRD** — deploys MongoDB via the Percona Server for MongoDB operator with a configurable replica set, resource limits, and optional persistent storage
- **LoadBalancer Service** — created only when ingress is enabled, exposes MongoDB on port 27017 with an `external-dns.alpha.kubernetes.io/hostname` annotation for automatic DNS record creation

## Prerequisites

- **Kubernetes credentials** configured via environment variables or Planton provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **Percona Server for MongoDB operator** installed in the cluster (the operator manages the PerconaServerMongoDB custom resource)
- **A StorageClass** available in the cluster if enabling persistence (most managed Kubernetes clusters provide a default)
- **external-dns** running in the cluster if enabling ingress with a hostname

## Quick Start

Create a file `mongodb.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesMongodb
metadata:
  name: my-mongodb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesMongodb.my-mongodb
spec:
  namespace: database
  createNamespace: true
```

Deploy:

```shell
planton apply -f mongodb.yaml
```

This creates a single-replica MongoDB instance with persistence enabled, a 1Gi PersistentVolumeClaim, default resource limits (1000m CPU, 1Gi memory), and a randomly generated password stored in a Kubernetes Secret.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the MongoDB deployment. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `container.replicas` | `int32` | `1` | Number of MongoDB replica set members. |
| `container.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation for each MongoDB pod. |
| `container.resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation for each MongoDB pod. |
| `container.resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU for each MongoDB pod. |
| `container.resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory for each MongoDB pod. |
| `container.persistenceEnabled` | `bool` | `true` | Enables persistent storage for MongoDB data. When enabled, data is persisted to a PersistentVolumeClaim and survives pod restarts. |
| `container.diskSize` | `string` | `1Gi` | Size of the PersistentVolumeClaim attached to each MongoDB pod. Required when `persistenceEnabled` is `true`. Must be a valid Kubernetes quantity (e.g., `1Gi`, `10Gi`). Cannot be modified after creation. |
| `ingress.enabled` | `bool` | `false` | Creates a LoadBalancer Service with external-dns annotations exposing MongoDB on port 27017. |
| `ingress.hostname` | `string` | — | Hostname for external access (e.g., `mongodb.example.com`). Configured automatically via external-dns. Required when `ingress.enabled` is `true`. |
| `helmValues` | `map<string, string>` | — | Additional key-value pairs passed to the Helm chart for further customization. See the [Bitnami MongoDB chart documentation](https://artifacthub.io/packages/helm/bitnami/mongodb) for available options. |

## Examples

### Development MongoDB without Persistence

A lightweight MongoDB instance for development with persistence disabled and reduced resources:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesMongodb
metadata:
  name: dev-mongodb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesMongodb.dev-mongodb
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
        memory: "128Mi"
```

### Production MongoDB with Increased Storage

A production MongoDB instance with a larger replica set, more disk, and higher resource limits:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesMongodb
metadata:
  name: prod-mongodb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesMongodb.prod-mongodb
spec:
  namespace: production
  container:
    replicas: 3
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
    persistenceEnabled: true
    diskSize: "50Gi"
```

### MongoDB with External Access

MongoDB exposed outside the cluster via a LoadBalancer with automatic DNS management:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesMongodb
metadata:
  name: shared-mongodb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesMongodb.shared-mongodb
spec:
  namespace: shared-services
  container:
    replicas: 3
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
    persistenceEnabled: true
    diskSize: "100Gi"
  ingress:
    enabled: true
    hostname: mongodb.example.com
```

### Using Foreign Key References

Reference an Planton-managed namespace instead of hardcoding the name:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesMongodb
metadata:
  name: app-mongodb
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesMongodb.app-mongodb
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: app-namespace
      field: spec.name
  container:
    persistenceEnabled: true
    diskSize: "20Gi"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where MongoDB is deployed |
| `service` | `string` | Kubernetes Service name for the MongoDB instance (format: `{name}`) |
| `port_forward_command` | `string` | kubectl port-forward command for local access |
| `kube_endpoint` | `string` | Cluster-internal FQDN (e.g., `my-mongodb.database.svc.cluster.local`) |
| `external_hostname` | `string` | Public hostname for external access, only set when ingress is enabled |
| `internal_hostname` | `string` | Internal hostname for cluster-internal access |
| `username` | `string` | MongoDB admin username (always `databaseAdmin`) |
| `password_secret.name` | `string` | Name of the Kubernetes Secret containing the MongoDB password (format: `{name}-password`) |
| `password_secret.key` | `string` | Key within the password Secret (always `MONGODB_DATABASE_ADMIN_PASSWORD`) |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — application deployments that connect to MongoDB as a data store
- [KubernetesExternalDns](/docs/catalog/kubernetes/kubernetesexternaldns) — manages DNS records for the LoadBalancer ingress hostname
