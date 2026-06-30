# Kubernetes Harbor

Deploys Harbor cloud-native container registry on Kubernetes using the official Harbor Helm chart. Provisions separate Harbor Core, Portal, Registry, and Jobservice components with independent resource tuning. Supports self-managed or external PostgreSQL and Redis, multiple artifact storage backends (S3, GCS, Azure Blob, Alibaba OSS, filesystem), arbitrary Helm value overrides, and optional external access through Istio Gateway API ingress with TLS termination.

## What Gets Created

When you deploy a KubernetesHarbor resource, Planton provisions:

- **Kubernetes Namespace** — created if `createNamespace` is `true`
- **Harbor Helm Release** — the official `harbor` chart from `https://helm.goharbor.io`, which creates:
  - Harbor Core pod (API server, authentication, webhook) on port 80
  - Harbor Portal pod (web UI) on port 80
  - Harbor Registry pod (Docker/OCI registry backend) on port 5000
  - Harbor Jobservice pod (background job execution)
  - Kubernetes Services for each component for cluster-internal access
- **PostgreSQL** — either a self-managed in-cluster instance with persistent storage, or integration with an external PostgreSQL endpoint
- **Redis** — either a self-managed in-cluster instance, or integration with an external Redis endpoint (including Sentinel support)
- **Artifact Storage** — configured backend for container images and Helm charts (filesystem PVC, S3, GCS, Azure Blob, or Alibaba OSS)
- **Admin Credentials** — a Kubernetes Secret containing the Harbor admin password
- **Ingress Resources** (when `ingress.core.enabled` is `true`):
  - cert-manager Certificate for TLS, issued by a ClusterIssuer matching the ingress domain
  - Gateway API Gateway with HTTPS listener (port 443) and TLS termination
  - HTTPRoute for HTTPS traffic forwarding to the Harbor Core service

## Prerequisites

- **A Kubernetes cluster** with kubectl configured for access
- **Istio ingress gateway** installed (only if using ingress)
- **cert-manager** with a ClusterIssuer matching your ingress domain (only if using ingress)
- **Gateway API CRDs** installed in the cluster (only if using ingress)
- **An S3-compatible bucket, GCS bucket, or Azure Blob container** provisioned (only if using external object storage instead of filesystem)

## Quick Start

Create a file `harbor.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesHarbor
metadata:
  name: my-harbor
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesHarbor.my-harbor
spec:
  namespace:
    value: harbor-dev
  createNamespace: true
  database:
    isExternal: false
  cache:
    isExternal: false
  storage:
    type: filesystem
    filesystem:
      diskSize: "100Gi"
```

Deploy:

```shell
planton apply -f harbor.yaml
```

This creates a Harbor instance with self-managed PostgreSQL (20Gi disk) and Redis (8Gi disk), filesystem-based artifact storage, and default resource allocations for all components. An admin user is created automatically with a generated password stored in a Kubernetes Secret.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace for the Harbor deployment. Use `value` for a direct string or `valueFrom` to reference a KubernetesNamespace resource. | Required |
| `database` | `object` | PostgreSQL configuration. Set `isExternal: false` for self-managed or `isExternal: true` with `externalDatabase` connection details. | Required |
| `cache` | `object` | Redis configuration. Set `isExternal: false` for self-managed or `isExternal: true` with `externalCache` connection details. | Required |
| `storage` | `object` | Artifact storage backend. Must include `type` (`filesystem`, `s3`, `gcs`, `azure`, or `oss`) and the matching configuration block. | Required; type-specific sub-object required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `createNamespace` | `bool` | `false` | Create the namespace if it does not exist. |
| `coreContainer.replicas` | `int32` | `1` | Number of Harbor Core pods. |
| `coreContainer.resources.limits.cpu` | `string` | `"1000m"` | CPU limit for Harbor Core. |
| `coreContainer.resources.limits.memory` | `string` | `"2Gi"` | Memory limit for Harbor Core. |
| `coreContainer.resources.requests.cpu` | `string` | `"200m"` | CPU request for Harbor Core. |
| `coreContainer.resources.requests.memory` | `string` | `"512Mi"` | Memory request for Harbor Core. |
| `coreContainer.image.repo` | `string` | chart default | Custom container image repository for Harbor Core. |
| `coreContainer.image.tag` | `string` | chart default | Custom container image tag for Harbor Core. |
| `portalContainer.replicas` | `int32` | `1` | Number of Harbor Portal pods. |
| `portalContainer.resources.limits.cpu` | `string` | `"500m"` | CPU limit for Harbor Portal. |
| `portalContainer.resources.limits.memory` | `string` | `"512Mi"` | Memory limit for Harbor Portal. |
| `portalContainer.resources.requests.cpu` | `string` | `"100m"` | CPU request for Harbor Portal. |
| `portalContainer.resources.requests.memory` | `string` | `"256Mi"` | Memory request for Harbor Portal. |
| `registryContainer.replicas` | `int32` | `1` | Number of Harbor Registry pods. |
| `registryContainer.resources.limits.cpu` | `string` | `"1000m"` | CPU limit for Harbor Registry. |
| `registryContainer.resources.limits.memory` | `string` | `"2Gi"` | Memory limit for Harbor Registry. |
| `registryContainer.resources.requests.cpu` | `string` | `"200m"` | CPU request for Harbor Registry. |
| `registryContainer.resources.requests.memory` | `string` | `"512Mi"` | Memory request for Harbor Registry. |
| `jobserviceContainer.replicas` | `int32` | `1` | Number of Harbor Jobservice pods. |
| `jobserviceContainer.resources.limits.cpu` | `string` | `"1000m"` | CPU limit for Harbor Jobservice. |
| `jobserviceContainer.resources.limits.memory` | `string` | `"1Gi"` | Memory limit for Harbor Jobservice. |
| `jobserviceContainer.resources.requests.cpu` | `string` | `"100m"` | CPU request for Harbor Jobservice. |
| `jobserviceContainer.resources.requests.memory` | `string` | `"256Mi"` | Memory request for Harbor Jobservice. |
| `database.managedDatabase.container.replicas` | `int32` | `1` | Number of self-managed PostgreSQL pods. |
| `database.managedDatabase.container.resources.limits.cpu` | `string` | `"1000m"` | CPU limit for self-managed PostgreSQL. |
| `database.managedDatabase.container.resources.limits.memory` | `string` | `"2Gi"` | Memory limit for self-managed PostgreSQL. |
| `database.managedDatabase.container.resources.requests.cpu` | `string` | `"200m"` | CPU request for self-managed PostgreSQL. |
| `database.managedDatabase.container.resources.requests.memory` | `string` | `"512Mi"` | Memory request for self-managed PostgreSQL. |
| `database.managedDatabase.container.persistenceEnabled` | `bool` | `true` | Enable persistent storage for PostgreSQL data. |
| `database.managedDatabase.container.diskSize` | `string` | `"20Gi"` | Persistent volume size for PostgreSQL. Cannot be changed after creation. |
| `database.externalDatabase.host` | `string` | -- | Hostname of the external PostgreSQL instance. Required when `database.isExternal` is `true`. |
| `database.externalDatabase.port` | `int32` | `5432` | Port for external PostgreSQL. |
| `database.externalDatabase.username` | `string` | -- | Username for external PostgreSQL authentication. |
| `database.externalDatabase.password` | `string` | -- | Password for external PostgreSQL authentication. |
| `database.externalDatabase.coreDatabase` | `string` | `"registry"` | Database name for Harbor Core. |
| `database.externalDatabase.clairDatabase` | `string` | `"clair"` | Database name for Clair vulnerability scanner. |
| `database.externalDatabase.notaryServerDatabase` | `string` | `"notary_server"` | Database name for Notary Server. |
| `database.externalDatabase.notarySignerDatabase` | `string` | `"notary_signer"` | Database name for Notary Signer. |
| `database.externalDatabase.useSsl` | `bool` | `false` | Enable SSL/TLS connection to external PostgreSQL. |
| `cache.managedCache.container.replicas` | `int32` | `1` | Number of self-managed Redis pods. |
| `cache.managedCache.container.resources.limits.cpu` | `string` | `"500m"` | CPU limit for self-managed Redis. |
| `cache.managedCache.container.resources.limits.memory` | `string` | `"512Mi"` | Memory limit for self-managed Redis. |
| `cache.managedCache.container.resources.requests.cpu` | `string` | `"100m"` | CPU request for self-managed Redis. |
| `cache.managedCache.container.resources.requests.memory` | `string` | `"256Mi"` | Memory request for self-managed Redis. |
| `cache.managedCache.container.persistenceEnabled` | `bool` | `true` | Enable persistent storage for Redis data. |
| `cache.managedCache.container.diskSize` | `string` | `"8Gi"` | Persistent volume size for Redis. |
| `cache.externalCache.host` | `string` | -- | Hostname of the external Redis instance. Required when `cache.isExternal` is `true`. |
| `cache.externalCache.port` | `int32` | `6379` | Port for external Redis. |
| `cache.externalCache.username` | `string` | -- | Username for external Redis (if ACLs are enabled). |
| `cache.externalCache.password` | `string` | -- | Password for external Redis. |
| `cache.externalCache.databaseIndex` | `int32` | `0` | Redis database index. |
| `cache.externalCache.useSentinel` | `bool` | `false` | Enable Redis Sentinel for high availability. |
| `cache.externalCache.sentinelMasterSet` | `string` | -- | Sentinel master set name. Required when `useSentinel` is `true`. |
| `storage.filesystem.diskSize` | `string` | -- | PVC size for filesystem storage (e.g., `"100Gi"`). Cannot be changed after creation. |
| `storage.filesystem.storageClass` | `string` | cluster default | Kubernetes StorageClass for the PVC. |
| `storage.s3.bucket` | `string` | -- | S3 bucket name. Required when `storage.type` is `s3`. |
| `storage.s3.region` | `string` | -- | AWS region for the S3 bucket. |
| `storage.s3.accessKey` | `string` | -- | AWS access key ID. |
| `storage.s3.secretKey` | `string` | -- | AWS secret access key. |
| `storage.s3.endpointUrl` | `string` | -- | Custom endpoint URL for S3-compatible services (e.g., MinIO). |
| `storage.s3.encrypt` | `bool` | `false` | Enable server-side encryption. |
| `storage.s3.secure` | `bool` | `false` | Use HTTPS connection to S3. |
| `storage.gcs.bucket` | `string` | -- | GCS bucket name. Required when `storage.type` is `gcs`. |
| `storage.gcs.keyData` | `string` | -- | Base64-encoded GCP service account key JSON. |
| `storage.gcs.chunkSize` | `int32` | `5242880` | Upload chunk size in bytes. |
| `storage.azure.accountName` | `string` | -- | Azure storage account name. Required when `storage.type` is `azure`. |
| `storage.azure.accountKey` | `string` | -- | Azure storage account key. |
| `storage.azure.container` | `string` | -- | Azure Blob container name. |
| `ingress.core.enabled` | `bool` | `false` | Enable external access to Harbor Core/Portal via Istio Gateway API with TLS. |
| `ingress.core.hostname` | `string` | -- | Full hostname for external Harbor access (e.g., `harbor.example.com`). Required when `ingress.core.enabled` is `true`. |
| `ingress.notary.enabled` | `bool` | `false` | Enable external access to the Notary service for image signing. |
| `ingress.notary.hostname` | `string` | -- | Full hostname for external Notary access. Required when `ingress.notary.enabled` is `true`. |
| `helmValues` | `map<string, string>` | `{}` | Additional Helm chart values for customization. See the [Harbor Helm chart values](https://github.com/goharbor/harbor-helm) for available options. |

## Examples

### Development Instance with Filesystem Storage

A minimal Harbor deployment using self-managed PostgreSQL, Redis, and local filesystem storage for development:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesHarbor
metadata:
  name: dev-harbor
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesHarbor.dev-harbor
spec:
  namespace:
    value: harbor-dev
  createNamespace: true
  coreContainer:
    replicas: 1
    resources:
      limits:
        cpu: "500m"
        memory: "1Gi"
      requests:
        cpu: "100m"
        memory: "256Mi"
  portalContainer:
    replicas: 1
    resources:
      limits:
        cpu: "250m"
        memory: "256Mi"
      requests:
        cpu: "50m"
        memory: "128Mi"
  registryContainer:
    replicas: 1
    resources:
      limits:
        cpu: "500m"
        memory: "1Gi"
      requests:
        cpu: "100m"
        memory: "256Mi"
  database:
    isExternal: false
    managedDatabase:
      container:
        persistenceEnabled: true
        diskSize: "10Gi"
  cache:
    isExternal: false
    managedCache:
      container:
        persistenceEnabled: true
        diskSize: "4Gi"
  storage:
    type: filesystem
    filesystem:
      diskSize: "50Gi"
```

### Production with S3 Storage and External Database

A deployment using an external PostgreSQL database, self-managed Redis, S3 object storage for artifacts, and increased replicas:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesHarbor
metadata:
  name: prod-harbor
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesHarbor.prod-harbor
spec:
  namespace:
    value: harbor-prod
  createNamespace: true
  coreContainer:
    replicas: 2
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
  portalContainer:
    replicas: 2
    resources:
      limits:
        cpu: "1000m"
        memory: "1Gi"
      requests:
        cpu: "200m"
        memory: "512Mi"
  registryContainer:
    replicas: 2
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
  jobserviceContainer:
    replicas: 2
    resources:
      limits:
        cpu: "2000m"
        memory: "2Gi"
      requests:
        cpu: "250m"
        memory: "512Mi"
  database:
    isExternal: true
    externalDatabase:
      host: harbor-db.us-east-1.rds.amazonaws.com
      port: 5432
      username: harbor_admin
      password: changeme-use-secret-manager
      coreDatabase: harbor_registry
      clairDatabase: harbor_clair
      notaryServerDatabase: harbor_notary_server
      notarySignerDatabase: harbor_notary_signer
      useSsl: true
  cache:
    isExternal: false
    managedCache:
      container:
        replicas: 1
        resources:
          limits:
            cpu: "1000m"
            memory: "1Gi"
          requests:
            cpu: "200m"
            memory: "512Mi"
        persistenceEnabled: true
        diskSize: "16Gi"
  storage:
    type: s3
    s3:
      bucket: harbor-artifacts-prod
      region: us-east-1
      accessKey: AKIAIOSFODNN7EXAMPLE
      secretKey: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
      encrypt: true
      secure: true
```

### Full-Featured with Ingress and GCS Storage

External HTTPS access with TLS, GCS artifact storage, external Redis with Sentinel, and Helm value overrides for Trivy vulnerability scanning:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesHarbor
metadata:
  name: main-harbor
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesHarbor.main-harbor
spec:
  namespace:
    value: harbor-production
  createNamespace: true
  coreContainer:
    replicas: 3
    resources:
      limits:
        cpu: "4000m"
        memory: "8Gi"
      requests:
        cpu: "1000m"
        memory: "2Gi"
  portalContainer:
    replicas: 2
  registryContainer:
    replicas: 3
    resources:
      limits:
        cpu: "4000m"
        memory: "8Gi"
      requests:
        cpu: "1000m"
        memory: "2Gi"
  jobserviceContainer:
    replicas: 2
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
  database:
    isExternal: true
    externalDatabase:
      host: harbor-postgres.internal.example.com
      port: 5432
      username: harbor
      password: changeme-use-secret-manager
      useSsl: true
  cache:
    isExternal: true
    externalCache:
      host: harbor-redis.internal.example.com
      port: 6379
      password: changeme-use-secret-manager
      useSentinel: true
      sentinelMasterSet: harbor-master
  storage:
    type: gcs
    gcs:
      bucket: harbor-artifacts-production
      keyData: ewogICJ0eXBlIjogInNlcnZpY2VfYWNjb3VudCIKfQ==
  ingress:
    core:
      enabled: true
      hostname: harbor.example.com
    notary:
      enabled: true
      hostname: notary.example.com
  helmValues:
    trivy.enabled: "true"
    trivy.ignoreUnfixed: "true"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where Harbor was created |
| `coreService` | `string` | Name of the Kubernetes Service for Harbor Core (e.g., `my-harbor-harbor-core`) |
| `portalService` | `string` | Name of the Kubernetes Service for Harbor Portal (e.g., `my-harbor-harbor-portal`) |
| `registryService` | `string` | Name of the Kubernetes Service for Harbor Registry (e.g., `my-harbor-harbor-registry`) |
| `jobserviceService` | `string` | Name of the Kubernetes Service for Harbor Jobservice (e.g., `my-harbor-harbor-jobservice`) |
| `portForwardCommand` | `string` | Ready-to-run `kubectl port-forward` command to access Harbor Portal locally on port 80 |
| `internalCoreEndpoint` | `string` | Cluster-internal endpoint for Harbor Core (e.g., `my-harbor-harbor-core.harbor-dev.svc.cluster.local:80`) |
| `internalRegistryEndpoint` | `string` | Cluster-internal endpoint for Harbor Registry (e.g., `my-harbor-harbor-registry.harbor-dev.svc.cluster.local:5000`) |
| `externalHostname` | `string` | External hostname when core ingress is enabled (e.g., `harbor.example.com`) |
| `registryExternalHostname` | `string` | External hostname for Docker registry access (docker login, pull, push) |
| `notaryExternalHostname` | `string` | External hostname for Notary image signing service when notary ingress is enabled |
| `adminUsername` | `string` | Harbor admin username (default: `admin`) |
| `adminPasswordSecret` | `KubernetesSecretKey` | Reference to the Kubernetes Secret containing the admin password (`name` = `{resourceName}-harbor-core`, `key` = `HARBOR_ADMIN_PASSWORD`) |
| `databaseEndpoint` | `string` | Cluster-internal PostgreSQL endpoint. Only populated when using self-managed database. |
| `databaseUsername` | `string` | PostgreSQL username. Only populated when using self-managed database. |
| `databasePasswordSecret` | `KubernetesSecretKey` | Reference to the Kubernetes Secret containing the PostgreSQL password. Only populated when using self-managed database. |
| `redisEndpoint` | `string` | Cluster-internal Redis endpoint. Only populated when using self-managed cache. |
| `redisPasswordSecret` | `KubernetesSecretKey` | Reference to the Kubernetes Secret containing the Redis password. Only populated when using self-managed cache. |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — pre-create a namespace to reference via `valueFrom`
- [KubernetesPostgres](/docs/catalog/kubernetes/kubernetespostgres) — deploy a standalone PostgreSQL instance as an external database for Harbor
- [KubernetesSecret](/docs/catalog/kubernetes/kubernetessecret) — manage secrets for external database or storage credentials
- [KubernetesHelmRelease](/docs/catalog/kubernetes/kuberneteshelmrelease) — deploy additional Helm charts alongside Harbor (e.g., monitoring exporters)
