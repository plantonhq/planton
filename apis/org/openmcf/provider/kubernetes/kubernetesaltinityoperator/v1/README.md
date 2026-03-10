# Overview

The **Altinity ClickHouse Operator Kubernetes API resource** is designed to deploy and manage the Altinity ClickHouse Operator within Kubernetes environments. This resource provides a consistent interface for deploying the operator that enables the management of ClickHouse database clusters on Kubernetes.

## Why We Created This API Resource

Deploying the Altinity ClickHouse Operator manually in Kubernetes can be complex, requiring specific knowledge of Helm charts, CRD installations, and operator configuration. The Altinity Operator Kubernetes API resource simplifies this process by offering a standardized, declarative approach to deploying the operator. This resource allows teams to:

- **Efficiently Deploy the Operator**: Simplify the process of deploying the Altinity ClickHouse Operator on Kubernetes by abstracting away the complexities of Helm chart configurations.
- **Enable ClickHouse Management**: Once deployed, the operator enables teams to easily deploy and manage ClickHouse clusters using the ClickHouseInstallation CRD.
- **Ensure Consistency**: Provide a standardized way to deploy the operator across different environments and clusters.
- **Optimize Resource Management**: Configure operator resource requirements to match your cluster's capacity.

## Key Features

### Kubernetes Cluster Integration

- **Target Cluster Configuration**: This resource integrates seamlessly with Planton's Kubernetes cluster credential management system, ensuring that the operator is deployed to the correct cluster.
- **Namespace Isolation**: The operator is deployed in a dedicated `kubernetes-altinity-operator` namespace for clean resource separation.

### Namespace Management

- **Flexible Namespace Creation**: Control whether the operator creates a new namespace or uses an existing one via the `create_namespace` flag
- **Default Behavior**: By default (`create_namespace: true`), a dedicated namespace is created for the operator
- **Existing Namespace Support**: Set `create_namespace: false` to deploy into a pre-existing namespace managed separately
- **Use Cases for External Namespace Management**:
  - Integration with centralized namespace provisioning systems
  - Namespaces with pre-configured ResourceQuotas and LimitRanges
  - Multi-tenant environments with strict namespace governance
  - Namespaces created by KubernetesNamespace resource with custom policies

### Operator Configuration

#### Resource Management

- **Configurable Resources**: Define CPU and memory allocation for the operator pod to ensure optimal performance. The recommended default values are:
  - **CPU Requests**: `100m`
  - **Memory Requests**: `256Mi`
  - **CPU Limits**: `1000m` (1 CPU core)
  - **Memory Limits**: `1Gi`

#### Automated CRD Installation

- **CRD Management**: The operator automatically installs the required ClickHouse Custom Resource Definitions (CRDs), including:
  - `ClickHouseInstallation` (CHI) - For deploying ClickHouse clusters
  - `ClickHouseOperatorConfiguration` - For operator configuration

#### Cluster-Wide Watching

- **Multi-Namespace Support**: The operator watches all namespaces by default, allowing ClickHouse clusters to be deployed in any namespace within the cluster.

### Helm Chart Based Deployment

- **Official Altinity Chart**: Deploys using the official Altinity Helm chart from `https://docs.altinity.com/clickhouse-operator/`
- **Version Control**: Uses a specific, tested version of the operator (currently 0.23.6) ensuring stability and predictability.
- **Atomic Deployments**: Helm releases are deployed atomically with automatic rollback on failure.

## Benefits

- **Simplicity**: This resource streamlines operator deployment, making it easier for DevOps teams to set up ClickHouse management capabilities in Kubernetes.
- **Foundation for ClickHouse**: Provides the necessary operator infrastructure that enables deploying actual ClickHouse clusters using the ClickHouseKubernetes workload resource.
- **Production Ready**: Configured with sensible defaults and production-grade deployment patterns including proper timeouts, cleanup on failure, and resource limits.
- **Consistent Deployment**: Ensures the operator is deployed consistently across all environments using infrastructure-as-code principles.

## Use Cases

The Altinity ClickHouse Operator is essential for teams that need to:

1. **Deploy ClickHouse Clusters**: Run analytical workloads requiring high-performance columnar databases
2. **Manage ClickHouse at Scale**: Operate multiple ClickHouse clusters across different namespaces
3. **Automate Database Operations**: Leverage the operator for automated cluster management, scaling, and updates
4. **Run Analytics Workloads**: Support data analytics, real-time dashboards, and OLAP workloads

## Next Steps

After deploying the Altinity Operator, you can:

1. Deploy ClickHouse clusters using the `ClickHouseKubernetes` workload resource
2. Create ClickHouse databases for your applications
3. Configure cluster topologies (shards, replicas, ZooKeeper)
4. Monitor operator and cluster health through Kubernetes tools

