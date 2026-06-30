# Overview

The **Percona Operator for MongoDB Kubernetes API resource** is designed to deploy and manage the Percona Operator for MongoDB within Kubernetes environments. This resource provides a consistent interface for deploying the operator that enables the management of MongoDB database clusters on Kubernetes.

## Why We Created This API Resource

Deploying the Percona Operator for MongoDB manually in Kubernetes can be complex, requiring specific knowledge of Helm charts, CRD installations, and operator configuration. The Percona Server Mongodb Operator API resource simplifies this process by offering a standardized, declarative approach to deploying the operator. This resource allows teams to:

- **Efficiently Deploy the Operator**: Simplify the process of deploying the Percona Operator for MongoDB on Kubernetes by abstracting away the complexities of Helm chart configurations.
- **Enable MongoDB Management**: Once deployed, the operator enables teams to easily deploy and manage MongoDB replica sets and sharded clusters using the PerconaServerMongoDB CRD.
- **Ensure Consistency**: Provide a standardized way to deploy the operator across different environments and clusters.
- **Optimize Resource Management**: Configure operator resource requirements to match your cluster's capacity.

## Key Features

### Kubernetes Cluster Integration

- **Target Cluster Configuration**: This resource integrates seamlessly with Planton's Kubernetes cluster credential management system, ensuring that the operator is deployed to the correct cluster.
- **Flexible Namespace Management**: Choose whether to create a new namespace or use an existing one, providing flexibility for different deployment scenarios.

### Operator Configuration

#### Resource Management

- **Configurable Resources**: Define CPU and memory allocation for the operator pod to ensure optimal performance. The recommended default values are:
  - **CPU Requests**: `100m`
  - **Memory Requests**: `256Mi`
  - **CPU Limits**: `1000m` (1 CPU core)
  - **Memory Limits**: `1Gi`

#### Automated CRD Installation

- **CRD Management**: The operator automatically installs the required MongoDB Custom Resource Definitions (CRDs), including:
  - `PerconaServerMongoDB` - For deploying MongoDB replica sets and sharded clusters
  - `PerconaServerMongoDBBackup` - For managing backups
  - `PerconaServerMongoDBRestore` - For restore operations

#### Cluster-Wide Watching

- **Multi-Namespace Support**: The operator watches all namespaces by default, allowing MongoDB clusters to be deployed in any namespace within the cluster.

### Namespace Management

The module provides flexible namespace handling through the `create_namespace` flag:

- **Create New Namespace (`create_namespace: true`)**: The module will create a dedicated namespace for the operator. Use this when deploying to a fresh environment or when you want the operator isolated in its own namespace.

- **Use Existing Namespace (`create_namespace: false`)**: The module will use an existing namespace specified in the `namespace` field. Use this when:
  - The namespace already exists in your cluster
  - You have pre-configured namespace with specific policies, quotas, or RBAC
  - You want to deploy the operator alongside other resources in a shared namespace
  - You're using a GitOps workflow where namespaces are managed separately

**Important**: When using an existing namespace (`create_namespace: false`), ensure the namespace exists before deploying the operator, otherwise the deployment will fail.

### Helm Chart Based Deployment

- **Official Percona Chart**: Deploys using the official Percona Helm chart from `https://percona.github.io/percona-helm-charts/`
- **Version Control**: Uses a specific, tested version of the operator ensuring stability and predictability.
- **Atomic Deployments**: Helm releases are deployed atomically with automatic rollback on failure.

## Benefits

- **Simplicity**: This resource streamlines operator deployment, making it easier for DevOps teams to set up MongoDB management capabilities in Kubernetes.
- **Foundation for MongoDB**: Provides the necessary operator infrastructure that enables deploying actual MongoDB clusters using the MongoDBKubernetes workload resource.
- **Production Ready**: Configured with sensible defaults and production-grade deployment patterns including proper timeouts, cleanup on failure, and resource limits.
- **Consistent Deployment**: Ensures the operator is deployed consistently across all environments using infrastructure-as-code principles.

## Use Cases

The Percona Operator for MongoDB is essential for teams that need to:

1. **Deploy MongoDB Clusters**: Run document-oriented workloads requiring flexible schema databases
2. **Manage MongoDB at Scale**: Operate multiple MongoDB clusters across different namespaces
3. **Automate Database Operations**: Leverage the operator for automated cluster management, scaling, backups, and updates
4. **Run Application Workloads**: Support web applications, microservices, and data-intensive workloads

## Next Steps

After deploying the Percona Operator, you can:

1. Deploy MongoDB clusters using the `MongoDBKubernetes` workload resource
2. Create MongoDB databases for your applications
3. Configure replica sets and sharded clusters
4. Set up automated backups and monitoring
5. Monitor operator and cluster health through Kubernetes tools

