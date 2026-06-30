# KubernetesElasticOperator

## Overview

**KubernetesElasticOperator** is a Planton component that deploys the Elastic Cloud on Kubernetes (ECK) operator to manage the Elastic Stack on Kubernetes clusters. ECK automates the deployment, provisioning, and orchestration of Elasticsearch, Kibana, APM Server, Enterprise Search, Beats, Elastic Agent, and Logstash.

The ECK operator extends the Kubernetes API with Custom Resource Definitions (CRDs) that enable declarative management of Elastic Stack components. It handles complex operational tasks including certificate management, rolling upgrades, scaling, and cross-cluster replication.

## Key Features

### Automated Lifecycle Management

The ECK operator provides automated operations for the entire Elastic Stack:

- **Certificate Management**: Automatic generation, rotation, and renewal of TLS certificates for secure inter-node and client communication
- **Rolling Upgrades**: Zero-downtime updates with automated orchestration of Elasticsearch cluster upgrades
- **Scaling Operations**: Automatic scaling of Elasticsearch clusters with proper data rebalancing
- **Configuration Management**: Dynamic configuration updates without manual intervention

### Operator Pattern Benefits

By using the Kubernetes Operator pattern, ECK provides:

- **Declarative Configuration**: Define desired state in YAML; ECK reconciles reality to match
- **Self-Healing**: Automatic recovery from node failures, pod crashes, and configuration drift
- **Consistent Deployments**: Same deployment process across development, staging, and production
- **Resource Efficiency**: Optimal resource allocation with Kubernetes-native scheduling

### Resource Management

The operator deployment can be configured with custom resource allocations:

- **CPU and Memory Limits**: Define maximum resources the operator can consume
- **Resource Requests**: Guaranteed baseline resources for operator pod
- **Default Configuration**: Pre-configured with production-ready defaults (50m CPU / 100Mi memory requests, 1000m CPU / 1Gi memory limits)

### Kubernetes Integration

ECK integrates seamlessly with Kubernetes ecosystem:

- **StatefulSets**: For stable Elasticsearch node identity and ordered deployment
- **Persistent Volumes**: For durable data storage across pod restarts
- **Services**: For stable network endpoints and load balancing
- **ConfigMaps & Secrets**: For configuration and credential management
- **RBAC**: For fine-grained access control to Kubernetes resources

### Multi-Cluster Management

Once installed, the ECK operator can manage Elastic Stack deployments:

- **Namespace Isolation**: Deploy multiple Elasticsearch clusters in different namespaces
- **Cluster-Wide Monitoring**: Single operator instance can manage multiple Elastic deployments
- **Resource Quotas**: Integration with Kubernetes resource quotas and limits
- **Network Policies**: Support for pod-to-pod network isolation

## Component Structure

### API Definition

The KubernetesElasticOperator API follows Planton's standard resource structure:

```
kubernetes-elastic-operator
├── api_version: "kubernetes.planton.dev/v1"
├── kind: "KubernetesElasticOperator"
├── metadata: CloudResourceMetadata
└── spec: KubernetesElasticOperatorSpec
    ├── target_cluster: KubernetesAddonTargetCluster
    ├── namespace: StringValueOrRef
    ├── create_namespace: bool
    └── container: KubernetesElasticOperatorSpecContainer
        └── resources: ContainerResources
            ├── limits: {cpu, memory}
            └── requests: {cpu, memory}
```

### Deployment Model

**Namespace**: Configurable (default: `elastic-system`)  
**Namespace Management**: Controlled by `create_namespace` flag
**Installation Method**: Helm chart from official Elastic repository  
**Operator Pod**: Single pod deployment (can be scaled for HA)  
**CRDs Installed**: Elasticsearch, Kibana, APM Server, Enterprise Search, Beats, Agent, Logstash

## Configuration

### Resource Specification

The `spec.container.resources` field controls the ECK operator pod's resource allocation:

```yaml
spec:
  container:
    resources:
      requests:
        cpu: "50m"      # Minimum guaranteed CPU
        memory: "100Mi"  # Minimum guaranteed memory
      limits:
        cpu: "1000m"    # Maximum CPU (1 core)
        memory: "1Gi"    # Maximum memory (1 GiB)
```

### Default Resources

If not specified, the operator uses production-ready defaults:

- **Requests**: 50m CPU, 100Mi memory (guaranteed baseline)
- **Limits**: 1000m CPU (1 core), 1Gi memory (prevents resource exhaustion)

These defaults are suitable for managing moderate-sized Elastic Stack deployments. Adjust based on:

- **Number of managed clusters**: More clusters = higher resource needs
- **Cluster sizes**: Larger Elasticsearch clusters require more operator resources
- **Update frequency**: Frequent updates increase operator workload

### Target Cluster Configuration

The `target_cluster` field specifies where to install the operator:

```yaml
spec:
  target_cluster:
    cluster_name: "my-gke-cluster"
  namespace:
    value: "elastic-system"
```

### Namespace Management

The `create_namespace` field controls whether the component creates the namespace or uses an existing one:

**When `create_namespace: true` (default for new deployments):**
- The component creates the specified namespace with appropriate labels
- Namespace lifecycle is managed by this component
- Useful for new installations

**When `create_namespace: false`:**
- The component assumes the namespace already exists
- You must ensure the namespace is created beforehand
- Useful when namespace is managed by another component or tool
- Useful when namespace has specific configurations (quotas, policies) managed externally

Example with namespace creation:

```yaml
spec:
  namespace:
    value: "elastic-system"
  create_namespace: true
```

Example using existing namespace:

```yaml
spec:
  namespace:
    value: "my-existing-namespace"
  create_namespace: false
```

> **Note:** If `create_namespace: false`, ensure the namespace exists before deploying the operator, or the Helm release will fail.

## Usage Patterns

### Basic Installation

Deploy ECK operator with default configuration:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesElasticOperator
metadata:
  name: eck-operator
spec:
  target_cluster:
    cluster_name: "my-gke-cluster"
  namespace:
    value: "elastic-system"
  container:
    resources:
      requests:
        cpu: "50m"
        memory: "100Mi"
      limits:
        cpu: "1000m"
        memory: "1Gi"
```

### High-Availability Configuration

For production environments managing multiple large clusters:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesElasticOperator
metadata:
  name: eck-operator-ha
spec:
  target_cluster:
    cluster_name: "production-gke-cluster"
  namespace:
    value: "elastic-system"
  container:
    resources:
      requests:
        cpu: "200m"
        memory: "512Mi"
      limits:
        cpu: "2000m"
        memory: "2Gi"
```

### Development/Testing

Minimal resource allocation for development environments:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesElasticOperator
metadata:
  name: eck-operator-dev
spec:
  target_cluster:
    cluster_name: "dev-gke-cluster"
  namespace:
    value: "elastic-system"
  container:
    resources:
      requests:
        cpu: "25m"
        memory: "64Mi"
      limits:
        cpu: "500m"
        memory: "512Mi"
```

## Post-Installation

After deploying the ECK operator, you can create Elastic Stack resources:

### Example: Elasticsearch Cluster

```yaml
apiVersion: elasticsearch.k8s.elastic.co/v1
kind: Elasticsearch
metadata:
  name: quickstart
  namespace: default
spec:
  version: 8.11.0
  nodeSets:
  - name: default
    count: 3
    config:
      node.store.allow_mmap: false
```

### Example: Kibana Instance

```yaml
apiVersion: kibana.k8s.elastic.co/v1
kind: Kibana
metadata:
  name: quickstart
  namespace: default
spec:
  version: 8.11.0
  count: 1
  elasticsearchRef:
    name: quickstart
```

## Benefits

### Operational Efficiency

- **Reduced Complexity**: No manual certificate management, upgrade orchestration, or scaling procedures
- **Faster Deployments**: Deploy complete Elastic Stack in minutes instead of hours/days
- **Lower Maintenance**: Operator handles routine operational tasks automatically
- **Consistency**: Same deployment process across all environments

### Reliability

- **Self-Healing**: Automatic recovery from failures
- **Zero-Downtime Updates**: Rolling upgrades without service interruption
- **Data Safety**: Proper data rebalancing during scaling operations
- **Certificate Rotation**: Automatic renewal prevents expiration-related outages

### Scalability

- **Multi-Cluster Support**: Single operator manages multiple Elasticsearch clusters
- **Efficient Resource Use**: Kubernetes-native scheduling and resource allocation
- **Horizontal Scaling**: Easy scale-out for growing data volumes
- **Namespace Isolation**: Secure multi-tenancy support

### Security

- **TLS by Default**: All communication secured with automatically managed certificates
- **RBAC Integration**: Fine-grained access control using Kubernetes RBAC
- **Secret Management**: Credentials stored in Kubernetes Secrets
- **Network Policies**: Support for pod-level network isolation

## Documentation

For detailed information, see:

- **Research Documentation**: [docs/README.md](docs/README.md) - Comprehensive guide on ECK deployment patterns, comparison of installation methods, and production best practices
- **Examples**: [examples.md](examples.md) - Practical usage examples for different scenarios
- **Pulumi Module**: [iac/pulumi/README.md](iac/pulumi/README.md) - Pulumi-specific implementation details
- **Terraform Module**: [iac/tf/README.md](iac/tf/README.md) - Terraform-specific implementation details

## Version Information

- **ECK Operator Version**: 2.14.0
- **Helm Chart Repository**: https://helm.elastic.co
- **Chart Name**: eck-operator
- **Namespace**: elastic-system
- **Supported Elastic Stack Versions**: 7.x and 8.x

## Support

For issues, questions, or contributions:

- **GitHub**: [plantonhq/planton](https://github.com/plantonhq/planton)
- **Documentation**: [Planton Docs](https://planton.dev)
- **Elastic ECK Documentation**: [Elastic ECK Docs](https://www.elastic.co/guide/en/cloud-on-k8s/current/index.html)

