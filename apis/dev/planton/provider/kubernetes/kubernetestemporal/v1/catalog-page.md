# Kubernetes Temporal

Deploys a Temporal server cluster on Kubernetes using the official Temporal Helm chart, with support for Cassandra, PostgreSQL, or MySQL database backends (embedded or external), optional Temporal Web UI, Elasticsearch-based advanced visibility, Prometheus and Grafana monitoring, and external access through gRPC LoadBalancer services and Istio Gateway API ingress with automatic TLS via cert-manager.

## What Gets Created

When you deploy a KubernetesTemporal resource, Planton provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Temporal Helm Release** — installs the Temporal server (frontend, history, matching, worker services) from the official `go.temporal.io/helm-charts` repository with configurable chart version, database backend, dynamic config, history shards, and per-service replica/resource settings
- **Database Backend** — either an in-cluster Cassandra, MySQL, or PostgreSQL instance (managed by the Helm chart), or connection configuration for an external database with TLS enabled
- **Database Password Secret** — a Kubernetes Secret containing the external database password, created only when an external database is configured with a plain string password (skipped when using `secretRef`)
- **Schema Jobs** — automatic database schema creation, setup, and update jobs (can be disabled via `database.disableAutoSchemaSetup`)
- **Temporal Web UI** — enabled by default, can be disabled with `disableWebUi`
- **Monitoring Stack** — Prometheus, Grafana, and kube-prometheus-stack, deployed when `enableMonitoringStack` is `true` or when external Elasticsearch is configured
- **Elasticsearch** — embedded Elasticsearch when `enableEmbeddedElasticsearch` is `true`, or connection to an external Elasticsearch cluster for advanced visibility
- **Frontend gRPC LoadBalancer Service** — an external LoadBalancer Service exposing the Temporal frontend on port 7233, created only when frontend ingress is enabled with a `grpcHostname`
- **Frontend HTTP Ingress** — a cert-manager Certificate, Istio Gateway, and HTTPRoutes (HTTPS + HTTP-to-HTTPS redirect) for the frontend HTTP API on port 7243, created only when frontend ingress is enabled with an `httpHostname`
- **Web UI Ingress** — a cert-manager Certificate, Istio Gateway, and HTTPRoutes (HTTPS + HTTP-to-HTTPS redirect) for the Temporal Web UI on port 8080, created only when web UI ingress is enabled

## Prerequisites

- **Kubernetes credentials** configured via environment variables or Planton provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **An external database** accessible from the cluster when using PostgreSQL or MySQL backends without embedded mode (Cassandra can run in-cluster)
- **Istio** with Gateway API support installed if enabling frontend HTTP or web UI ingress
- **cert-manager** with a ClusterIssuer matching the ingress domain if enabling ingress with TLS
- **external-dns** configured if using frontend gRPC LoadBalancer ingress with automatic DNS

## Quick Start

Create a file `temporal.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTemporal
metadata:
  name: my-temporal
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesTemporal.my-temporal
spec:
  namespace: temporal
  createNamespace: true
  database:
    backend: cassandra
```

Deploy:

```shell
planton apply -f temporal.yaml
```

This creates a single-replica Temporal cluster backed by an in-cluster Cassandra node, with the Web UI enabled and default schema auto-setup.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the Temporal deployment. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `database.backend` | `enum` | Database backend for Temporal persistence. Valid values: `cassandra`, `postgresql`, `mysql`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `disableWebUi` | `bool` | `false` | Disables the Temporal Web UI. |
| `enableEmbeddedElasticsearch` | `bool` | `false` | Enables embedded Elasticsearch for advanced visibility. Ignored if external Elasticsearch is configured. |
| `enableMonitoringStack` | `bool` | `false` | Deploys Prometheus, Grafana, and kube-prometheus-stack for Temporal monitoring. |
| `cassandraReplicas` | `int32` | `1` | Number of Cassandra nodes. Only honored when the backend is `cassandra` and no external database is provided. |
| `version` | `string` | `0.62.0` | Version of the Temporal Helm chart to deploy (e.g., `0.62.0`). |
| `database.externalDatabase.host` | `string` | — | Hostname for the external database. Required when backend is `postgresql` or `mysql`. |
| `database.externalDatabase.port` | `int32` | — | Port for the external database. |
| `database.externalDatabase.username` | `string` | — | Username for the external database. |
| `database.externalDatabase.password` | `KubernetesSensitiveValue` | — | Password for the external database. Accepts `value` (plain string) or `secretRef` (reference to an existing Kubernetes Secret with `name` and `key`). |
| `database.databaseName` | `string` | `temporal` | Primary database or keyspace name. |
| `database.visibilityName` | `string` | `temporal_visibility` | Visibility database or keyspace name. |
| `database.disableAutoSchemaSetup` | `bool` | `false` | Disables automatic database schema creation. |
| `ingress.frontend.enabled` | `bool` | `false` | Enables external access to the Temporal frontend via gRPC LoadBalancer and optionally HTTP via Gateway API. |
| `ingress.frontend.grpcHostname` | `string` | — | Full hostname for gRPC access via LoadBalancer (e.g., `temporal-grpc.example.com`). Required when frontend ingress is enabled. |
| `ingress.frontend.httpHostname` | `string` | — | Full hostname for HTTP access via Gateway API (e.g., `temporal-http.example.com`). Optional; creates Gateway/HTTPRoute resources only if provided. |
| `ingress.webUi.enabled` | `bool` | `false` | Enables external access to the Temporal Web UI via Gateway API. |
| `ingress.webUi.hostname` | `string` | — | Full hostname for Web UI access (e.g., `temporal-ui.example.com`). Required when web UI ingress is enabled. |
| `externalElasticsearch.host` | `string` | — | Host address of an existing Elasticsearch cluster for advanced visibility. |
| `externalElasticsearch.port` | `int32` | — | Port for the external Elasticsearch cluster. |
| `externalElasticsearch.user` | `string` | — | Username for the external Elasticsearch cluster. |
| `externalElasticsearch.password` | `KubernetesSensitiveValue` | — | Password for the external Elasticsearch cluster. Accepts `value` (plain string) or `secretRef` (reference to an existing Kubernetes Secret with `name` and `key`). |
| `numHistoryShards` | `int32` | `512` | Number of history shards. This is **immutable** after initial deployment. Higher values enable better parallelism. Range: 1-16384. |
| `dynamicConfig.historySizeLimitError` | `int64` | `52428800` | Maximum workflow history size in bytes (50 MB default). Temporal terminates workflows exceeding this limit. Minimum: 1048576 (1 MB). |
| `dynamicConfig.historyCountLimitError` | `int64` | `51200` | Maximum number of events in workflow history. Minimum: 1000. |
| `dynamicConfig.historySizeLimitWarn` | `int64` | `10485760` | Warning threshold for history size in bytes (10 MB default). Minimum: 524288. |
| `dynamicConfig.historyCountLimitWarn` | `int64` | `10240` | Warning threshold for history event count. Minimum: 500. |
| `dynamicConfig.blobSizeLimitError` | `int64` | `2097152` | Maximum single payload size in bytes (2 MB default). Controls marker details, signal data, and activity I/O. Minimum: 1048576. |
| `dynamicConfig.blobSizeLimitWarn` | `int64` | `524288` | Warning threshold for payload size in bytes (512 KB default). Minimum: 262144. |
| `services.frontend.replicas` | `int32` | `1` | Number of frontend service replicas. Range: 1-100. |
| `services.frontend.resources.limits.cpu` | `string` | — | Maximum CPU for each frontend pod. |
| `services.frontend.resources.limits.memory` | `string` | — | Maximum memory for each frontend pod. |
| `services.frontend.resources.requests.cpu` | `string` | — | Minimum guaranteed CPU for each frontend pod. |
| `services.frontend.resources.requests.memory` | `string` | — | Minimum guaranteed memory for each frontend pod. |
| `services.history.replicas` | `int32` | `1` | Number of history service replicas (most resource-intensive). Range: 1-100. |
| `services.history.resources.limits.cpu` | `string` | — | Maximum CPU for each history pod. |
| `services.history.resources.limits.memory` | `string` | — | Maximum memory for each history pod. |
| `services.history.resources.requests.cpu` | `string` | — | Minimum guaranteed CPU for each history pod. |
| `services.history.resources.requests.memory` | `string` | — | Minimum guaranteed memory for each history pod. |
| `services.matching.replicas` | `int32` | `1` | Number of matching service replicas. Range: 1-100. |
| `services.matching.resources.limits.cpu` | `string` | — | Maximum CPU for each matching pod. |
| `services.matching.resources.limits.memory` | `string` | — | Maximum memory for each matching pod. |
| `services.matching.resources.requests.cpu` | `string` | — | Minimum guaranteed CPU for each matching pod. |
| `services.matching.resources.requests.memory` | `string` | — | Minimum guaranteed memory for each matching pod. |
| `services.worker.replicas` | `int32` | `1` | Number of worker service replicas. Range: 1-100. |
| `services.worker.resources.limits.cpu` | `string` | — | Maximum CPU for each worker pod. |
| `services.worker.resources.limits.memory` | `string` | — | Maximum memory for each worker pod. |
| `services.worker.resources.requests.cpu` | `string` | — | Minimum guaranteed CPU for each worker pod. |
| `services.worker.resources.requests.memory` | `string` | — | Minimum guaranteed memory for each worker pod. |

## Examples

### Development Temporal with In-Cluster Cassandra

A lightweight single-node Temporal instance backed by embedded Cassandra for local development and testing:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTemporal
metadata:
  name: dev-temporal
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesTemporal.dev-temporal
spec:
  namespace: temporal-dev
  createNamespace: true
  database:
    backend: cassandra
  cassandraReplicas: 1
```

### Production Temporal with External PostgreSQL

A production-grade Temporal cluster using an external PostgreSQL database, increased history limits for large workflows, and tuned service replicas:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTemporal
metadata:
  name: prod-temporal
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesTemporal.prod-temporal
spec:
  namespace: temporal
  database:
    backend: postgresql
    externalDatabase:
      host: temporal-db.internal.example.com
      port: 5432
      username: temporal
      password:
        secretRef:
          name: temporal-db-credentials
          key: password
    databaseName: temporal
    visibilityName: temporal_visibility
  numHistoryShards: 512
  dynamicConfig:
    historySizeLimitError: 104857600
    historyCountLimitError: 102400
    blobSizeLimitError: 10485760
    blobSizeLimitWarn: 5242880
  services:
    frontend:
      replicas: 2
      resources:
        requests:
          cpu: "200m"
          memory: "512Mi"
        limits:
          cpu: "1000m"
          memory: "2Gi"
    history:
      replicas: 3
      resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
        limits:
          cpu: "2000m"
          memory: "4Gi"
    matching:
      replicas: 2
      resources:
        requests:
          cpu: "200m"
          memory: "512Mi"
    worker:
      replicas: 1
      resources:
        requests:
          cpu: "100m"
          memory: "256Mi"
```

### Temporal with Full Ingress, Monitoring, and External Elasticsearch

Temporal exposed externally via gRPC LoadBalancer and Istio Gateway API for both the frontend HTTP API and Web UI, with monitoring and external Elasticsearch for advanced visibility:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTemporal
metadata:
  name: platform-temporal
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesTemporal.platform-temporal
spec:
  namespace: temporal
  database:
    backend: postgresql
    externalDatabase:
      host: temporal-db.internal.example.com
      port: 5432
      username: temporal
      password:
        secretRef:
          name: temporal-db-credentials
          key: password
  enableMonitoringStack: true
  externalElasticsearch:
    host: elasticsearch.internal.example.com
    port: 9200
    user: elastic
    password:
      secretRef:
        name: es-credentials
        key: password
  ingress:
    frontend:
      enabled: true
      grpcHostname: temporal-grpc.example.com
      httpHostname: temporal-http.example.com
    webUi:
      enabled: true
      hostname: temporal-ui.example.com
  services:
    frontend:
      replicas: 3
      resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
        limits:
          cpu: "2000m"
          memory: "4Gi"
    history:
      replicas: 3
      resources:
        requests:
          cpu: "1000m"
          memory: "2Gi"
        limits:
          cpu: "4000m"
          memory: "8Gi"
    matching:
      replicas: 2
    worker:
      replicas: 2
```

### Using Foreign Key References

Reference an Planton-managed namespace instead of hardcoding the name:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTemporal
metadata:
  name: my-temporal
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesTemporal.my-temporal
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: temporal-namespace
      field: spec.name
  database:
    backend: postgresql
    externalDatabase:
      host: temporal-db.internal.example.com
      port: 5432
      username: temporal
      password:
        value: my-dev-password
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where Temporal is deployed |
| `frontendServiceName` | `string` | Kubernetes Service name for the Temporal frontend (format: `{name}-frontend`) |
| `uiServiceName` | `string` | Kubernetes Service name for the Temporal Web UI (format: `{name}-web`) |
| `portForwardFrontendCommand` | `string` | kubectl port-forward command for local access to the Temporal frontend on port 7233 |
| `portForwardUiCommand` | `string` | kubectl port-forward command for local access to the Temporal Web UI on port 8080 |
| `frontendEndpoint` | `string` | Cluster-internal FQDN for the frontend (e.g., `my-temporal-frontend.temporal.svc.cluster.local:7233`) |
| `webUiEndpoint` | `string` | Cluster-internal FQDN for the Web UI (e.g., `my-temporal-web.temporal.svc.cluster.local:8080`) |
| `externalFrontendHostname` | `string` | External hostname for the frontend, only set when frontend ingress is enabled with a gRPC hostname |
| `externalUiHostname` | `string` | External hostname for the Web UI, only set when web UI ingress is enabled |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesPostgres](/docs/catalog/kubernetes/kubernetespostgres) — deploy an in-cluster PostgreSQL instance as the Temporal database backend
- [KubernetesElasticsearch](/docs/catalog/kubernetes/kuberneteselasticsearch) — deploy an in-cluster Elasticsearch instance for Temporal advanced visibility
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — application deployments that use Temporal as a workflow orchestration backend
