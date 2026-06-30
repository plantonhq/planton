# Kubernetes SigNoz

Deploys the SigNoz observability platform on Kubernetes using the official SigNoz Helm chart, providing unified logs, metrics, and traces through an OpenTelemetry-native stack with configurable SigNoz UI, OpenTelemetry Collector, self-managed or external ClickHouse database, optional Kubernetes Gateway API ingress for both the UI and OTel Collector endpoints, and custom Helm value overrides.

## What Gets Created

When you deploy a KubernetesSignoz resource, Planton provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **SigNoz Helm Release** — the full SigNoz stack (UI, API server, Ruler, Alertmanager, and Frontend) deployed via the `signoz` Helm chart from `https://charts.signoz.io` (chart version 0.52.0)
- **OpenTelemetry Collector** — a multi-replica data ingestion gateway accepting traces, metrics, and logs over gRPC (port 4317) and HTTP (port 4318)
- **Self-Managed ClickHouse** — an in-cluster ClickHouse deployment with configurable persistence, clustering (sharding and replication), and optional Zookeeper coordination; created only when `database.isExternal` is `false`
- **Zookeeper** — coordination service for distributed ClickHouse clusters; created only when `database.managedDatabase.zookeeper.isEnabled` is `true`
- **SigNoz UI Gateway and Routes** — a Kubernetes Gateway API Gateway, TLS Certificate (via cert-manager), HTTPS HTTPRoute, and HTTP-to-HTTPS redirect HTTPRoute for the SigNoz UI; created only when `ingress.ui.enabled` is `true`
- **OTel Collector Gateway and Routes** — a separate Gateway API Gateway, TLS Certificate, HTTPS HTTPRoute, and HTTP-to-HTTPS redirect HTTPRoute for the OpenTelemetry Collector HTTP endpoint; created only when `ingress.otelCollector.enabled` is `true`

## Prerequisites

- **Kubernetes credentials** configured via environment variables or Planton provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **A StorageClass** available in the cluster if enabling ClickHouse persistence (most managed Kubernetes clusters provide a default)
- **Istio ingress gateway** installed in the `istio-ingress` namespace if enabling ingress for the UI or OTel Collector
- **cert-manager** with a ClusterIssuer matching your ingress domain if enabling ingress
- **Gateway API CRDs** installed on the cluster if enabling ingress

## Quick Start

Create a file `signoz.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesSignoz
metadata:
  name: my-signoz
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesSignoz.my-signoz
spec:
  namespace: observability
  createNamespace: true
  database:
    isExternal: false
```

Deploy:

```shell
planton apply -f signoz.yaml
```

This creates a SigNoz instance with a single SigNoz replica (1000m CPU / 2Gi memory limits), two OTel Collector replicas (2000m CPU / 4Gi memory limits), and a self-managed single-node ClickHouse with 20Gi persistent storage. No ingress is configured; access the UI via port-forward using the `portForwardCommand` stack output.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace for the SigNoz deployment. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `database` | `object` | ClickHouse database configuration. Must specify either self-managed or external mode. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `signozContainer.replicas` | `int32` | `1` | Number of SigNoz (UI/API/Ruler/Alertmanager) pods. Must be at least 1. |
| `signozContainer.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation for each SigNoz pod. |
| `signozContainer.resources.limits.memory` | `string` | `2Gi` | Maximum memory allocation for each SigNoz pod. |
| `signozContainer.resources.requests.cpu` | `string` | `200m` | Minimum guaranteed CPU for each SigNoz pod. |
| `signozContainer.resources.requests.memory` | `string` | `512Mi` | Minimum guaranteed memory for each SigNoz pod. |
| `signozContainer.image.repo` | `string` | — | Custom container image repository for the SigNoz binary. |
| `signozContainer.image.tag` | `string` | — | Custom container image tag for the SigNoz binary. |
| `otelCollectorContainer.replicas` | `int32` | `2` | Number of OpenTelemetry Collector pods. Must be at least 1. |
| `otelCollectorContainer.resources.limits.cpu` | `string` | `2000m` | Maximum CPU allocation for each OTel Collector pod. |
| `otelCollectorContainer.resources.limits.memory` | `string` | `4Gi` | Maximum memory allocation for each OTel Collector pod. |
| `otelCollectorContainer.resources.requests.cpu` | `string` | `500m` | Minimum guaranteed CPU for each OTel Collector pod. |
| `otelCollectorContainer.resources.requests.memory` | `string` | `1Gi` | Minimum guaranteed memory for each OTel Collector pod. |
| `otelCollectorContainer.image.repo` | `string` | — | Custom container image repository for the OTel Collector. |
| `otelCollectorContainer.image.tag` | `string` | — | Custom container image tag for the OTel Collector. |
| `database.isExternal` | `bool` | `false` | When `true`, connects to an existing external ClickHouse instance instead of deploying one in-cluster. |
| `database.externalDatabase.host` | `string` | — | Hostname of the external ClickHouse instance. Required when `database.isExternal` is `true`. |
| `database.externalDatabase.httpPort` | `int32` | `8123` | HTTP port for the external ClickHouse instance. |
| `database.externalDatabase.tcpPort` | `int32` | `9000` | TCP port for the external ClickHouse native protocol. |
| `database.externalDatabase.clusterName` | `string` | `cluster` | Name of the distributed cluster in ClickHouse configuration. |
| `database.externalDatabase.isSecure` | `bool` | `false` | Whether to use TLS when connecting to the external ClickHouse instance. |
| `database.externalDatabase.username` | `string` | — | Username for authenticating to the external ClickHouse. Required when `database.isExternal` is `true`. |
| `database.externalDatabase.password` | `KubernetesSensitiveValue` | — | Password for the external ClickHouse. Supports `value` for a plain string or `secretRef` with `name` and `key` to reference an existing Kubernetes Secret. Required when `database.isExternal` is `true`. |
| `database.managedDatabase.container.replicas` | `int32` | `1` | Number of self-managed ClickHouse pods. Must be at least 1. |
| `database.managedDatabase.container.resources.limits.cpu` | `string` | `2000m` | Maximum CPU for each ClickHouse pod. |
| `database.managedDatabase.container.resources.limits.memory` | `string` | `4Gi` | Maximum memory for each ClickHouse pod. |
| `database.managedDatabase.container.resources.requests.cpu` | `string` | `500m` | Minimum guaranteed CPU for each ClickHouse pod. |
| `database.managedDatabase.container.resources.requests.memory` | `string` | `1Gi` | Minimum guaranteed memory for each ClickHouse pod. |
| `database.managedDatabase.container.persistenceEnabled` | `bool` | `true` | Enables persistent storage for ClickHouse data. |
| `database.managedDatabase.container.diskSize` | `string` | `20Gi` | Size of the PersistentVolumeClaim per ClickHouse pod. Required when `persistenceEnabled` is `true`. Must be a valid Kubernetes quantity (e.g., `20Gi`). Cannot be modified after creation. |
| `database.managedDatabase.container.image.repo` | `string` | — | Custom container image repository for ClickHouse. |
| `database.managedDatabase.container.image.tag` | `string` | — | Custom container image tag for ClickHouse. |
| `database.managedDatabase.cluster.isEnabled` | `bool` | `false` | Enables distributed cluster mode with sharding and replication for ClickHouse. |
| `database.managedDatabase.cluster.shardCount` | `int32` | — | Number of shards for distributed data storage. Must be at least 1 when clustering is enabled. |
| `database.managedDatabase.cluster.replicaCount` | `int32` | — | Number of replicas per shard for data redundancy. Must be at least 1 when clustering is enabled. |
| `database.managedDatabase.zookeeper.isEnabled` | `bool` | `false` | Enables Zookeeper deployment for distributed ClickHouse coordination. Must be `true` when clustering is enabled. |
| `database.managedDatabase.zookeeper.container.replicas` | `int32` | `1` | Number of Zookeeper pods. Use an odd number (3 or 5) for production quorum. |
| `database.managedDatabase.zookeeper.container.resources.limits.cpu` | `string` | `500m` | Maximum CPU for each Zookeeper pod. |
| `database.managedDatabase.zookeeper.container.resources.limits.memory` | `string` | `512Mi` | Maximum memory for each Zookeeper pod. |
| `database.managedDatabase.zookeeper.container.resources.requests.cpu` | `string` | `100m` | Minimum guaranteed CPU for each Zookeeper pod. |
| `database.managedDatabase.zookeeper.container.resources.requests.memory` | `string` | `256Mi` | Minimum guaranteed memory for each Zookeeper pod. |
| `database.managedDatabase.zookeeper.container.diskSize` | `string` | `8Gi` | Persistent volume size per Zookeeper pod. Must be a valid Kubernetes quantity. |
| `database.managedDatabase.zookeeper.container.image.repo` | `string` | — | Custom container image repository for Zookeeper. |
| `database.managedDatabase.zookeeper.container.image.tag` | `string` | — | Custom container image tag for Zookeeper. |
| `ingress.ui.enabled` | `bool` | `false` | Creates Gateway API resources for external SigNoz UI access with TLS termination and HTTP-to-HTTPS redirect. |
| `ingress.ui.hostname` | `string` | — | Hostname for external SigNoz UI access (e.g., `signoz.example.com`). Required when `ingress.ui.enabled` is `true`. |
| `ingress.otelCollector.enabled` | `bool` | `false` | Creates Gateway API resources for external OTel Collector HTTP endpoint access with TLS termination. |
| `ingress.otelCollector.hostname` | `string` | — | Hostname for external OTel Collector HTTP endpoint (e.g., `otel-ingest.example.com`). Required when `ingress.otelCollector.enabled` is `true`. |
| `helmValues` | `map<string, string>` | — | Additional key-value pairs passed to the SigNoz Helm chart for advanced customization. See [SigNoz Helm chart documentation](https://github.com/SigNoz/charts) for available options. |

> **Note on `namespace`:** The `namespace` field is a `StringValueOrRef`. You can provide a plain string value directly, or use `valueFrom` to reference the output of another Planton resource (e.g., a KubernetesNamespace).

## Examples

### Development SigNoz with Reduced Resources

A lightweight SigNoz instance for development and testing with smaller resource allocations and a single OTel Collector replica:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesSignoz
metadata:
  name: dev-signoz
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesSignoz.dev-signoz
spec:
  namespace: dev-observability
  createNamespace: true
  signozContainer:
    replicas: 1
    resources:
      limits:
        cpu: "500m"
        memory: "1Gi"
      requests:
        cpu: "100m"
        memory: "256Mi"
  otelCollectorContainer:
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
```

### Production SigNoz with Clustered ClickHouse and Ingress

A production-grade deployment with ClickHouse clustering (2 shards, 2 replicas), Zookeeper quorum, and external access for both the UI and OTel Collector:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesSignoz
metadata:
  name: prod-signoz
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesSignoz.prod-signoz
spec:
  namespace: observability
  signozContainer:
    replicas: 2
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
  otelCollectorContainer:
    replicas: 4
    resources:
      limits:
        cpu: "4000m"
        memory: "8Gi"
      requests:
        cpu: "1000m"
        memory: "2Gi"
  database:
    isExternal: false
    managedDatabase:
      container:
        replicas: 2
        resources:
          limits:
            cpu: "4000m"
            memory: "16Gi"
          requests:
            cpu: "1000m"
            memory: "4Gi"
        persistenceEnabled: true
        diskSize: "200Gi"
      cluster:
        isEnabled: true
        shardCount: 2
        replicaCount: 2
      zookeeper:
        isEnabled: true
        container:
          replicas: 3
          resources:
            limits:
              cpu: "500m"
              memory: "512Mi"
            requests:
              cpu: "100m"
              memory: "256Mi"
          diskSize: "10Gi"
  ingress:
    ui:
      enabled: true
      hostname: signoz.example.com
    otelCollector:
      enabled: true
      hostname: otel-ingest.example.com
```

### SigNoz with External ClickHouse

Connect SigNoz to an existing external ClickHouse instance instead of deploying one in-cluster. The password is referenced from a pre-existing Kubernetes Secret:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesSignoz
metadata:
  name: shared-signoz
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesSignoz.shared-signoz
spec:
  namespace: observability
  signozContainer:
    replicas: 2
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
  otelCollectorContainer:
    replicas: 3
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
      host: clickhouse.shared-infra.svc.cluster.local
      httpPort: 8123
      tcpPort: 9000
      clusterName: cluster
      isSecure: false
      username: signoz
      password:
        secretRef:
          name: clickhouse-credentials
          key: password
  ingress:
    ui:
      enabled: true
      hostname: signoz.example.com
```

### Using Foreign Key References

Reference an Planton-managed namespace instead of hardcoding the name:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesSignoz
metadata:
  name: team-signoz
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesSignoz.team-signoz
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: observability-namespace
      field: spec.name
  database:
    isExternal: false
    managedDatabase:
      container:
        persistenceEnabled: true
        diskSize: "50Gi"
```

### SigNoz with Custom Helm Values

Override additional Helm chart values for advanced customization, such as configuring retention policies:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesSignoz
metadata:
  name: custom-signoz
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.KubernetesSignoz.custom-signoz
spec:
  namespace: observability
  createNamespace: true
  database:
    isExternal: false
  helmValues:
    "signoz.alertmanager.enabled": "true"
    "queryService.replicaCount": "2"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where SigNoz is deployed |
| `signozService` | `string` | Kubernetes Service name for the SigNoz UI and API (format: `{name}-signoz`) |
| `otelCollectorService` | `string` | Kubernetes Service name for the OpenTelemetry Collector (format: `{name}-otel-collector`) |
| `portForwardCommand` | `string` | kubectl port-forward command for local access to SigNoz UI on port 8080 |
| `kubeEndpoint` | `string` | Cluster-internal FQDN for SigNoz UI (e.g., `my-signoz-signoz.observability.svc.cluster.local:8080`) |
| `externalHostname` | `string` | Public hostname for external SigNoz UI access, only set when `ingress.ui.enabled` is `true` |
| `internalHostname` | `string` | Internal hostname for VPC-internal SigNoz access |
| `otelCollectorGrpcEndpoint` | `string` | Cluster-internal FQDN for OTel Collector gRPC ingestion (e.g., `my-signoz-otel-collector.observability.svc.cluster.local:4317`) |
| `otelCollectorHttpEndpoint` | `string` | Cluster-internal FQDN for OTel Collector HTTP ingestion (e.g., `my-signoz-otel-collector.observability.svc.cluster.local:4318`) |
| `otelCollectorExternalGrpcHostname` | `string` | Public hostname for OTel Collector gRPC endpoint, only set when OTel Collector ingress is configured |
| `otelCollectorExternalHttpHostname` | `string` | Public hostname for OTel Collector HTTP endpoint, only set when `ingress.otelCollector.enabled` is `true` |
| `clickhouseEndpoint` | `string` | Cluster-internal ClickHouse endpoint (e.g., `my-signoz-clickhouse.observability.svc.cluster.local:8123`), only set when using self-managed ClickHouse |
| `clickhouseUsername` | `string` | ClickHouse username for authentication (always `admin`), only set when using self-managed ClickHouse |
| `clickhousePasswordSecret.name` | `string` | Name of the Kubernetes Secret containing the ClickHouse password (format: `{name}-clickhouse`), only set when using self-managed ClickHouse |
| `clickhousePasswordSecret.key` | `string` | Key within the ClickHouse password Secret (always `admin-password`), only set when using self-managed ClickHouse |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesClickHouse](/docs/catalog/kubernetes/kubernetesclickhouse) — standalone ClickHouse deployment that can be used as an external database for SigNoz
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — application deployments instrumented with OpenTelemetry SDKs that send telemetry to SigNoz
- [KubernetesIstio](/docs/catalog/kubernetes/kubernetesistio) — provides the Istio ingress gateway used by SigNoz Gateway API resources
- [KubernetesGatewayApiCrds](/docs/catalog/kubernetes/kubernetesgatewayapicrds) — installs the Gateway API CRDs required for SigNoz ingress configuration
