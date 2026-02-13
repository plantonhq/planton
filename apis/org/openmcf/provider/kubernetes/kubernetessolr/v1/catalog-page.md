# Kubernetes Solr

Deploys an Apache Solr cluster on Kubernetes using the Solr Operator's SolrCloud custom resource, with a co-located ZooKeeper ensemble for cluster coordination, persistent storage for both Solr and ZooKeeper data, and optional external access through Istio Gateway API ingress with automatic TLS via cert-manager.

## What Gets Created

When you deploy a KubernetesSolr resource, OpenMCF provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **SolrCloud Custom Resource** — a Solr Operator SolrCloud object that manages a StatefulSet of Solr pods with configurable replicas, container image, JVM tuning, resource limits, and persistent data volumes
- **ZooKeeper Ensemble** — a co-located ZooKeeper cluster (via the Solr Operator's provided ZooKeeper reference) with configurable replicas, resource limits, and persistent storage
- **Solr Modules** — `jaegertracer-configurator` and `ltr` (Learning to Rank) modules enabled by default
- **TLS Certificate** — a cert-manager Certificate for the ingress hostnames, created only when ingress is enabled
- **Istio Gateway** — an external Gateway resource with HTTPS (port 443) and HTTP (port 80) listeners, created only when ingress is enabled
- **HTTP-to-HTTPS Redirect Route** — an HTTPRoute that redirects HTTP traffic to HTTPS with a 301 status code, created only when ingress is enabled
- **HTTPS Route** — an HTTPRoute that forwards HTTPS traffic to the Solr common service, created only when ingress is enabled

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **Solr Operator** installed on the target cluster (manages the SolrCloud custom resource lifecycle)
- **A StorageClass** available in the cluster for Solr and ZooKeeper persistent volumes (most managed Kubernetes clusters provide a default)
- **Istio** with Gateway API support installed if enabling ingress
- **cert-manager** with a ClusterIssuer matching the ingress domain if enabling ingress with TLS

## Quick Start

Create a file `solr.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSolr
metadata:
  name: my-solr
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesSolr.my-solr
spec:
  namespace: search
  createNamespace: true
```

Deploy:

```shell
openmcf apply -f solr.yaml
```

This creates a single-replica Solr 9.10.0 instance backed by a single-replica ZooKeeper ensemble, each with 1Gi persistent volumes and default resource limits (1000m CPU, 1Gi memory).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the Solr deployment. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `solrContainer.replicas` | `int32` | `1` | Number of Solr pods in the SolrCloud cluster. |
| `solrContainer.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation for each Solr pod. |
| `solrContainer.resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation for each Solr pod. |
| `solrContainer.resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU for each Solr pod. |
| `solrContainer.resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory for each Solr pod. |
| `solrContainer.diskSize` | `string` | `1Gi` | Size of the PersistentVolumeClaim attached to each Solr pod. Must be a valid Kubernetes quantity (e.g., `1Gi`, `10Gi`). |
| `solrContainer.image.repo` | `string` | `solr` | Container image repository for Solr. |
| `solrContainer.image.tag` | `string` | `9.10.0` | Container image tag for Solr. |
| `config.javaMem` | `string` | — | JVM memory settings for Solr (e.g., `-Xms512m -Xmx512m`). |
| `config.opts` | `string` | — | Custom Solr options (e.g., `-Dsolr.autoSoftCommit.maxTime=10000`). |
| `config.garbageCollectionTuning` | `string` | — | Solr GC tuning parameters (e.g., `-XX:SurvivorRatio=4 -XX:TargetSurvivorRatio=90 -XX:MaxTenuringThreshold=8`). |
| `zookeeperContainer.replicas` | `int32` | `1` | Number of ZooKeeper pods in the ensemble. |
| `zookeeperContainer.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation for each ZooKeeper pod. |
| `zookeeperContainer.resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation for each ZooKeeper pod. |
| `zookeeperContainer.resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU for each ZooKeeper pod. |
| `zookeeperContainer.resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory for each ZooKeeper pod. |
| `zookeeperContainer.diskSize` | `string` | `1Gi` | Size of the PersistentVolumeClaim attached to each ZooKeeper pod. Must be a valid Kubernetes quantity (e.g., `1Gi`, `10Gi`). |
| `ingress.enabled` | `bool` | `false` | Enables external access through Istio Gateway API with TLS termination and HTTP-to-HTTPS redirect. |
| `ingress.hostname` | `string` | — | Full hostname for external access (e.g., `solr.example.com`). Required when `ingress.enabled` is `true`. |

## Examples

### Development Solr with Minimal Resources

A lightweight single-node Solr instance for development and testing:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSolr
metadata:
  name: dev-solr
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesSolr.dev-solr
spec:
  namespace: dev
  createNamespace: true
  solrContainer:
    replicas: 1
    resources:
      limits:
        cpu: "500m"
        memory: "512Mi"
      requests:
        cpu: "100m"
        memory: "256Mi"
    diskSize: "1Gi"
    image:
      repo: solr
      tag: "9.10.0"
  zookeeperContainer:
    replicas: 1
    resources:
      limits:
        cpu: "250m"
        memory: "256Mi"
      requests:
        cpu: "50m"
        memory: "128Mi"
    diskSize: "1Gi"
```

### Production Solr Cluster with JVM Tuning

A multi-replica Solr cluster with increased resources, larger storage, and JVM tuning for production workloads:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSolr
metadata:
  name: prod-solr
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesSolr.prod-solr
spec:
  namespace: search
  solrContainer:
    replicas: 3
    resources:
      limits:
        cpu: "4000m"
        memory: "8Gi"
      requests:
        cpu: "1000m"
        memory: "4Gi"
    diskSize: "50Gi"
    image:
      repo: solr
      tag: "9.10.0"
  config:
    javaMem: "-Xms4g -Xmx4g"
    opts: "-Dsolr.autoSoftCommit.maxTime=10000"
    garbageCollectionTuning: "-XX:SurvivorRatio=4 -XX:TargetSurvivorRatio=90 -XX:MaxTenuringThreshold=8"
  zookeeperContainer:
    replicas: 3
    resources:
      limits:
        cpu: "1000m"
        memory: "2Gi"
      requests:
        cpu: "250m"
        memory: "512Mi"
    diskSize: "10Gi"
```

### Solr with External Ingress

Solr exposed outside the cluster via Istio Gateway with TLS termination and automatic certificate management:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSolr
metadata:
  name: shared-solr
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesSolr.shared-solr
spec:
  namespace: search
  solrContainer:
    replicas: 3
    resources:
      limits:
        cpu: "4000m"
        memory: "8Gi"
      requests:
        cpu: "1000m"
        memory: "4Gi"
    diskSize: "100Gi"
    image:
      repo: solr
      tag: "9.10.0"
  config:
    javaMem: "-Xms4g -Xmx4g"
  zookeeperContainer:
    replicas: 3
    resources:
      limits:
        cpu: "1000m"
        memory: "2Gi"
      requests:
        cpu: "250m"
        memory: "512Mi"
    diskSize: "10Gi"
  ingress:
    enabled: true
    hostname: solr.example.com
```

### Using Foreign Key References

Reference an OpenMCF-managed namespace instead of hardcoding the name:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesSolr
metadata:
  name: search
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesSolr.search
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: search-namespace
      field: spec.name
  solrContainer:
    replicas: 3
    diskSize: "50Gi"
  zookeeperContainer:
    replicas: 3
    diskSize: "10Gi"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where Solr is deployed |
| `service` | `string` | Kubernetes Service name for the SolrCloud common service (format: `{name}-solrcloud-common`) |
| `port_forward_command` | `string` | kubectl port-forward command for local access to the Solr UI on port 8080 |
| `kube_endpoint` | `string` | Cluster-internal FQDN (e.g., `my-solr-solrcloud-common.search.svc.cluster.local`) |
| `external_hostname` | `string` | Public hostname for external access, only set when ingress is enabled |
| `internal_hostname` | `string` | Internal hostname for private access (format: `internal-{ingress.hostname}`), only set when ingress is enabled |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — application deployments that consume Solr as a search backend
- [KubernetesRedis](/docs/catalog/kubernetes/kubernetesredis) — complementary caching layer often paired with Solr for search applications
