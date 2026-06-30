# KubernetesSolrOperator

The **KubernetesSolrOperator** component deploys the Apache Solr Operator on a Kubernetes cluster, enabling automated management of SolrCloud clusters through Kubernetes custom resources.

## Overview

The Apache Solr Operator is the official, production-ready solution for running SolrCloud on Kubernetes. Originally developed at Bloomberg and donated to Apache, it automates the complete lifecycle management of Solr clusters including deployment, scaling, updates, backups, and monitoring.

### Key Features

- **Official Apache Project**: 100% open source (Apache License 2.0) with production-proven maturity
- **Automated Lifecycle Management**: Handles deployment, scaling, rolling updates, and recovery automatically
- **ZooKeeper Integration**: Manages ZooKeeper ensemble provisioning or connects to external clusters
- **Safe Scaling**: Automatic shard rebalancing when adding or removing nodes
- **Backup & Restore**: First-class backup/restore capabilities via Kubernetes CRDs
- **Production-Ready**: Battle-tested at Bloomberg (1000+ clusters, hundreds of machines)
- **Monitoring Integration**: Built-in Prometheus metrics exporter support
- **Advanced Configuration**: TLS, custom JVM options, ingress, topology awareness

### Use Cases

- **Search Infrastructure**: Production search services requiring high availability and scalability
- **Content Management Systems**: Large-scale content indexing and retrieval
- **E-commerce Search**: Product catalog search with real-time updates
- **Log Analytics**: Time-series data indexing and search (combined with other tools)
- **Knowledge Bases**: Enterprise search across documents and knowledge repositories

## Prerequisites

Before deploying the Solr Operator, ensure you have:

1. **Kubernetes Cluster**: Version 1.19+ with sufficient resources
2. **Kubernetes Credentials**: Valid credentials or credential ID for the target cluster
3. **Storage Class**: Dynamic provisioning enabled for persistent volumes
4. **Resource Capacity**: Operator requires minimal resources (defaults: 50m CPU, 100Mi memory)

### Optional

- **Cert-Manager**: For TLS certificate management in Solr clusters
- **Prometheus**: For monitoring and metrics collection
- **ExternalDNS**: For automatic DNS record management

## API Reference

### KubernetesSolrOperator

The main resource for deploying the Apache Solr Operator.

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesSolrOperator
metadata:
  name: <operator-name>
spec:
  targetCluster: # Optional - where to deploy
    clusterName: "<cluster-name>"
  namespace: # Required
    value: "solr-operator-system"
  create_namespace: false # Optional - whether to create the namespace
  container:
    resources: # Optional - defaults provided
      requests:
        cpu: 50m
        memory: 100Mi
      limits:
        cpu: 1000m
        memory: 1Gi
```

### Spec Fields

#### `targetCluster` (optional)

Specifies the target Kubernetes cluster for operator deployment.

- **`clusterName`** (string): Name of the Kubernetes cluster
- **`clusterKind`** (enum): Type of cluster (GKE, EKS, AKS, etc.)

If not provided, the operator deploys to the default cluster.

#### `namespace` (required)

Kubernetes namespace where the operator will be deployed. Default: `"solr-operator-system"`

#### `create_namespace` (optional)

Controls whether the component should create the namespace or use an existing one.

- **`true`**: Component creates the namespace if it doesn't exist
- **`false`** (default): Component uses an existing namespace (must exist before deployment)

**Use Cases:**
- Set to `true` when deploying to a new namespace that doesn't exist yet
- Set to `false` when namespace is managed separately (e.g., via KubernetesNamespace resource or external process)
- Set to `false` in multi-tenant scenarios where namespaces are pre-provisioned

**Default:** `false` (namespace must exist)

#### `container` (required)

Container configuration for the Solr Operator deployment.

- **`resources`** (object): CPU and memory allocations
  - **`requests`**: Minimum guaranteed resources
    - `cpu` (string): Default `"50m"`
    - `memory` (string): Default `"100Mi"`
  - **`limits`**: Maximum allowed resources
    - `cpu` (string): Default `"1000m"`
    - `memory` (string): Default `"1Gi"`

### Default Values

The operator ships with sensible defaults:

```yaml
container:
  resources:
    requests:
      cpu: "50m"        # Minimal baseline
      memory: "100Mi"   # Low memory footprint
    limits:
      cpu: "1000m"      # Can burst to 1 CPU
      memory: "1Gi"     # Generous memory limit
```

These defaults are suitable for most deployments. Increase resources only if you're managing a very large number of Solr clusters.

## Architecture

```
┌─────────────────────────────────────┐
│   Kubernetes Cluster                │
│                                     │
│  ┌──────────────────────────────┐  │
│  │  Solr Operator Pod           │  │
│  │  - Watches SolrCloud CRDs    │  │
│  │  - Manages StatefulSets      │  │
│  │  - Handles Scaling/Updates   │  │
│  └──────────────────────────────┘  │
│             │                       │
│             ↓                       │
│  ┌──────────────────────────────┐  │
│  │  Custom Resource Definitions  │  │
│  │  - SolrCloud                 │  │
│  │  - SolrBackup                │  │
│  │  - SolrPrometheusExporter    │  │
│  └──────────────────────────────┘  │
│             │                       │
│             ↓                       │
│  ┌──────────────────────────────┐  │
│  │  Managed Resources           │  │
│  │  - StatefulSets (Solr Pods)  │  │
│  │  - Services                  │  │
│  │  - ConfigMaps                │  │
│  │  - PersistentVolumeClaims    │  │
│  └──────────────────────────────┘  │
└─────────────────────────────────────┘
```

### How It Works

1. **Operator Installation**: Deploys operator pod and registers Custom Resource Definitions (CRDs)
2. **Custom Resources**: Users create `SolrCloud` resources defining desired Solr clusters
3. **Reconciliation**: Operator continuously watches CRDs and creates/updates underlying Kubernetes resources
4. **Lifecycle Management**: Handles scaling, updates, backups automatically based on CRD specifications
5. **Monitoring**: Exposes metrics and status through Kubernetes API and Prometheus endpoints

## Installation Methods

### Pulumi

Deploy using Planton's Pulumi module:

```bash
cd iac/pulumi
pulumi up
```

See [iac/pulumi/README.md](iac/pulumi/README.md) for detailed Pulumi usage.

### Terraform

Deploy using Planton's Terraform module:

```bash
cd iac/tf
terraform init
terraform plan
terraform apply
```

See [iac/tf/README.md](iac/tf/README.md) for detailed Terraform usage.

## Post-Installation

After the operator is deployed, you can create SolrCloud clusters:

```yaml
apiVersion: solr.apache.org/v1beta1
kind: SolrCloud
metadata:
  name: example-solr
spec:
  replicas: 3
  solrImage:
    repository: solr
    tag: 8.11.3
  solrJavaMem: "-Xms4g -Xmx4g"
  zookeeperRef:
    provided:
      replicas: 3
  dataStorage:
    persistent:
      reclaimPolicy: Retain
      pvcTemplate:
        spec:
          resources:
            requests:
              storage: 20Gi
```

Apply with `kubectl apply -f solrcloud.yaml`.

## Validation

The KubernetesSolrOperator spec includes built-in validation rules:

- **api_version**: Must be exactly `"kubernetes.planton.dev/v1"`
- **kind**: Must be exactly `"KubernetesSolrOperator"`
- **metadata**: Required with valid `name` field
- **spec.container**: Required field
- **Resource values**: Must be valid Kubernetes resource quantities (e.g., "100m", "1Gi")

Invalid configurations will be rejected with clear error messages.

## Outputs

After deployment, the operator provides these outputs:

- **namespace**: Kubernetes namespace where operator is deployed
- **operator_version**: Version of the Solr Operator installed
- **crds_installed**: List of Custom Resource Definitions registered

Access outputs via your IaC tool (Pulumi `pulumi stack output`, Terraform `terraform output`).

## Resource Requirements

### Operator Pod

| Resource | Request | Limit  |
|----------|---------|--------|
| CPU      | 50m     | 1000m  |
| Memory   | 100Mi   | 1Gi    |

The operator itself is lightweight. Resource usage scales with the number of managed Solr clusters, but even managing 100+ clusters typically stays well within these limits.

### Storage

The operator pod itself requires no persistent storage. Solr clusters managed by the operator will require PersistentVolumes according to their individual `SolrCloud` specifications.

## Monitoring

The operator exposes Prometheus metrics at `/metrics`:

- **solr_operator_reconcile_duration_seconds**: Time spent reconciling resources
- **solr_operator_reconcile_errors_total**: Count of reconciliation errors
- **solr_operator_solrcloud_managed**: Number of SolrCloud instances managed

Configure your Prometheus instance to scrape the operator's metrics endpoint for operational visibility.

## Troubleshooting

### Operator Pod Not Starting

```bash
# Check pod status
kubectl get pods -n <namespace>

# View pod logs
kubectl logs -n <namespace> <operator-pod-name>

# Check resource constraints
kubectl describe pod -n <namespace> <operator-pod-name>
```

### CRDs Not Installed

```bash
# List Custom Resource Definitions
kubectl get crds | grep solr

# Expected CRDs:
# - solrclouds.solr.apache.org
# - solrbackups.solr.apache.org
# - solrprometheusexporters.solr.apache.org
```

If CRDs are missing, the operator installation may have failed. Check operator pod logs.

### SolrCloud Resources Not Reconciling

```bash
# Check SolrCloud status
kubectl get solrcloud -n <namespace>

# View detailed status
kubectl describe solrcloud <solrcloud-name> -n <namespace>

# Check operator logs for reconciliation errors
kubectl logs -n <operator-namespace> <operator-pod-name>
```

## Best Practices

1. **Resource Allocation**: Start with defaults and increase only if needed
2. **Single Operator**: Deploy one operator per cluster (manages multiple SolrClouds)
3. **Namespace Isolation**: Deploy operator in dedicated namespace (e.g., `solr-operator-system`)
4. **Version Pinning**: Specify exact operator version for reproducible deployments
5. **Monitoring**: Enable Prometheus metrics for operational visibility
6. **Backup Testing**: Regularly test restore procedures for managed Solr clusters

## Security Considerations

- **RBAC**: Operator requires cluster-wide permissions to manage CRDs and resources
- **Network Policies**: Apply network policies to restrict operator pod communication if needed
- **Secrets**: Operator doesn't store credentials; SolrCloud resources may reference secrets
- **TLS**: Enable TLS for Solr clusters using cert-manager integration

## Upgrade Strategy

To upgrade the operator:

1. Check release notes for breaking changes
2. Update the operator deployment (via Pulumi/Terraform)
3. Monitor operator logs during rollout
4. Verify existing SolrCloud clusters remain healthy
5. Test with canary SolrCloud before rolling out widely

## Additional Resources

- **Apache Solr Operator GitHub**: https://github.com/apache/solr-operator
- **Official Documentation**: https://apache.github.io/solr-operator/
- **Solr on Kubernetes Guide**: https://solr.apache.org/operator/
- **Research Documentation**: [docs/README.md](docs/README.md) - Deep dive into deployment patterns
- **Examples**: [examples.md](examples.md) - Practical deployment scenarios

## Support

For issues with:
- **Operator Deployment**: Check this README and [iac/pulumi/README.md](iac/pulumi/README.md) or [iac/tf/README.md](iac/tf/README.md)
- **SolrCloud Resources**: Refer to Apache Solr Operator documentation
- **Planton**: File issues on Planton repository

## License

The Apache Solr Operator is licensed under Apache License 2.0. This component (KubernetesSolrOperator) is part of Planton and follows Planton's licensing terms.

