# Kubernetes Elasticsearch

Deploys an Elasticsearch cluster on Kubernetes using the Elastic Cloud on Kubernetes (ECK) operator, with optional Kibana, configurable data persistence via PersistentVolumeClaims, and optional external access through Gateway API ingress with automatic TLS certificate provisioning and HTTP-to-HTTPS redirect.

## What Gets Created

When you deploy a KubernetesElasticsearch resource, Planton provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Elasticsearch Cluster** — an ECK-managed Elasticsearch custom resource (v8.15.0) with configurable replicas, resource limits, and node roles (master, data, ingest)
- **Persistent Volumes** — when `persistenceEnabled` is `true`, a PersistentVolumeClaim (`ReadWriteOnce`) is attached to each Elasticsearch pod for durable data storage
- **Kibana Instance** — when `kibana.enabled` is `true`, an ECK-managed Kibana custom resource connected to the Elasticsearch cluster with configurable replicas and resource limits
- **Password Secret** — an auto-generated Kubernetes Secret containing the `elastic` user password, created by the ECK operator
- **TLS Certificate** — when any ingress is enabled, a cert-manager Certificate resource with a ClusterIssuer derived from the ingress hostname domain
- **Gateway + HTTPRoutes (Elasticsearch)** — when Elasticsearch ingress is enabled, a Gateway API Gateway and HTTPRoute pair for HTTPS traffic plus an HTTP-to-HTTPS redirect route
- **Gateway + HTTPRoutes (Kibana)** — when Kibana ingress is enabled, a separate Gateway and HTTPRoute pair for Kibana with the same HTTPS and redirect setup

## Prerequisites

- **Kubernetes credentials** configured via environment variables or Planton provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **ECK operator** installed in the cluster (manages Elasticsearch and Kibana custom resources)
- **A StorageClass** available in the cluster if enabling persistence (most managed Kubernetes clusters provide a default)
- **Istio ingress gateway** running in the `istio-ingress` namespace if enabling ingress (used as the Gateway API implementation)
- **cert-manager** installed with a ClusterIssuer matching the ingress hostname domain if enabling ingress
- **external-dns** running in the cluster if enabling ingress with a hostname

## Quick Start

Create a file `elasticsearch.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesElasticsearch
metadata:
  name: my-es
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesElasticsearch.my-es
spec:
  namespace: search
  createNamespace: true
```

Deploy:

```shell
planton apply -f elasticsearch.yaml
```

This creates a single-node Elasticsearch 8.15.0 cluster with persistence enabled, a 1Gi PersistentVolumeClaim, default resource limits (1000m CPU, 1Gi memory), Kibana enabled with a single replica, and an auto-generated password stored in a Kubernetes Secret.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the Elasticsearch deployment. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `elasticsearch.container.replicas` | `int32` | `1` | Number of Elasticsearch pods to deploy. Each pod runs master, data, and ingest roles. |
| `elasticsearch.container.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation for each Elasticsearch pod. |
| `elasticsearch.container.resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation for each Elasticsearch pod. |
| `elasticsearch.container.resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU for each Elasticsearch pod. |
| `elasticsearch.container.resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory for each Elasticsearch pod. |
| `elasticsearch.container.persistenceEnabled` | `bool` | `true` | Enables persistent storage for Elasticsearch data. When enabled, data is persisted to a PersistentVolumeClaim and restored on pod restart. |
| `elasticsearch.container.diskSize` | `string` | `1Gi` | Size of the PersistentVolumeClaim attached to each Elasticsearch pod. Required when `persistenceEnabled` is `true`. Must be a valid Kubernetes quantity (e.g., `1Gi`, `10Gi`). Cannot be modified after creation. |
| `elasticsearch.ingress.enabled` | `bool` | `false` | Creates a Gateway API Gateway and HTTPRoutes exposing Elasticsearch on port 9200 with TLS termination. |
| `elasticsearch.ingress.hostname` | `string` | — | Hostname for external access (e.g., `elasticsearch.example.com`). Configured automatically via external-dns. Required when `elasticsearch.ingress.enabled` is `true`. |
| `kibana.enabled` | `bool` | `true` | Deploys a Kibana instance connected to the Elasticsearch cluster. |
| `kibana.container.replicas` | `int32` | `1` | Number of Kibana pods to deploy. |
| `kibana.container.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation for each Kibana pod. |
| `kibana.container.resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation for each Kibana pod. |
| `kibana.container.resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU for each Kibana pod. |
| `kibana.container.resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory for each Kibana pod. |
| `kibana.ingress.enabled` | `bool` | `false` | Creates a Gateway API Gateway and HTTPRoutes exposing Kibana on port 5601 with TLS termination. |
| `kibana.ingress.hostname` | `string` | — | Hostname for external Kibana access (e.g., `kibana.example.com`). Configured automatically via external-dns. Required when `kibana.ingress.enabled` is `true`. |

## Examples

### Development Elasticsearch without Persistence

A lightweight Elasticsearch instance for development with persistence disabled, Kibana enabled, and reduced resources:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesElasticsearch
metadata:
  name: dev-es
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesElasticsearch.dev-es
spec:
  namespace: dev
  createNamespace: true
  elasticsearch:
    container:
      replicas: 1
      persistenceEnabled: false
      resources:
        limits:
          cpu: "500m"
          memory: "512Mi"
        requests:
          cpu: "100m"
          memory: "256Mi"
  kibana:
    enabled: true
    container:
      replicas: 1
      resources:
        limits:
          cpu: "500m"
          memory: "512Mi"
        requests:
          cpu: "100m"
          memory: "256Mi"
```

### Production Elasticsearch with Increased Storage

A production Elasticsearch cluster with multiple replicas, larger disk allocation, and higher resource limits:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesElasticsearch
metadata:
  name: prod-es
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesElasticsearch.prod-es
spec:
  namespace: production
  elasticsearch:
    container:
      replicas: 3
      resources:
        limits:
          cpu: "4000m"
          memory: "8Gi"
        requests:
          cpu: "1000m"
          memory: "4Gi"
      persistenceEnabled: true
      diskSize: "100Gi"
  kibana:
    enabled: true
    container:
      replicas: 2
      resources:
        limits:
          cpu: "2000m"
          memory: "2Gi"
        requests:
          cpu: "500m"
          memory: "1Gi"
```

### Elasticsearch with External Ingress

Elasticsearch and Kibana exposed outside the cluster via Gateway API with TLS termination and automatic DNS:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesElasticsearch
metadata:
  name: shared-es
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesElasticsearch.shared-es
spec:
  namespace: shared-services
  elasticsearch:
    container:
      replicas: 3
      resources:
        limits:
          cpu: "4000m"
          memory: "8Gi"
        requests:
          cpu: "1000m"
          memory: "4Gi"
      persistenceEnabled: true
      diskSize: "200Gi"
    ingress:
      enabled: true
      hostname: elasticsearch.example.com
  kibana:
    enabled: true
    container:
      replicas: 2
      resources:
        limits:
          cpu: "2000m"
          memory: "2Gi"
        requests:
          cpu: "500m"
          memory: "1Gi"
    ingress:
      enabled: true
      hostname: kibana.example.com
```

### Using Foreign Key References

Reference an Planton-managed namespace instead of hardcoding the name:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesElasticsearch
metadata:
  name: search
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesElasticsearch.search
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: search-namespace
      field: spec.name
  elasticsearch:
    container:
      replicas: 3
      persistenceEnabled: true
      diskSize: "50Gi"
  kibana:
    enabled: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where Elasticsearch is deployed |
| `elasticsearch.service` | `string` | Kubernetes Service name for Elasticsearch (format: `{name}-es-http`) |
| `elasticsearch.portForwardCommand` | `string` | kubectl port-forward command for local access on port 9200 |
| `elasticsearch.kubeEndpoint` | `string` | Cluster-internal FQDN (e.g., `my-es-es-http.search.svc.cluster.local`) |
| `elasticsearch.externalHostname` | `string` | Public hostname for external access, only set when Elasticsearch ingress is enabled |
| `elasticsearch.username` | `string` | Elasticsearch username (always `elastic`) |
| `elasticsearch.passwordSecret.name` | `string` | Name of the Kubernetes Secret containing the Elasticsearch password (format: `{name}-es-elastic-user`) |
| `elasticsearch.passwordSecret.key` | `string` | Key within the password Secret (always `elastic`) |
| `kibana.service` | `string` | Kubernetes Service name for Kibana (format: `{name}-kb-http`) |
| `kibana.portForwardCommand` | `string` | kubectl port-forward command for local Kibana access on port 5601 |
| `kibana.kubeEndpoint` | `string` | Cluster-internal FQDN for Kibana (e.g., `my-es-kb-http.search.svc.cluster.local`) |
| `kibana.externalHostname` | `string` | Public hostname for external Kibana access, only set when Kibana ingress is enabled |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — application deployments that query Elasticsearch for search or analytics
- [KubernetesSecret](/docs/catalog/kubernetes/kubernetessecret) — manage additional secrets consumed by Elasticsearch clients
