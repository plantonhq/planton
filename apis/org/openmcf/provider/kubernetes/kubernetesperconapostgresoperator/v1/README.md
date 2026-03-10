# Overview

The **Percona Operator for PostgreSQL Kubernetes API resource** is designed to deploy and manage the Percona Operator for PostgreSQL within Kubernetes environments. This resource provides a consistent interface for deploying the operator that enables the management of PostgreSQL database clusters on Kubernetes.

## Why We Created This API Resource

Deploying the Percona Operator for PostgreSQL manually in Kubernetes can be complex, requiring specific knowledge of Helm charts, CRD installations, and operator configuration. The Percona Postgresql Operator API resource simplifies this process by offering a standardized, declarative approach to deploying the operator. This resource allows teams to:

- **Efficiently Deploy the Operator**: Simplify the process of deploying the Percona Operator for PostgreSQL on Kubernetes by abstracting away the complexities of Helm chart configurations.
- **Enable PostgreSQL Management**: Once deployed, the operator enables teams to easily deploy and manage PostgreSQL clusters with high availability, automated backups, and disaster recovery capabilities.
- **Ensure Consistency**: Provide a standardized way to deploy the operator across different environments and clusters.
- **Optimize Resource Management**: Configure operator resource requirements to match your cluster's capacity.

## Key Features

### Kubernetes Cluster Integration

- **Target Cluster Configuration**: This resource integrates seamlessly with Planton's Kubernetes cluster credential management system, ensuring that the operator is deployed to the correct cluster.
- **Namespace Isolation**: The operator is deployed in a dedicated namespace for clean resource separation.

### Namespace Management

This component provides flexible namespace management through the `create_namespace` flag:

- **Automatic Namespace Creation** (`create_namespace: true`): The module creates and manages the namespace for you. This is ideal for:
  - New deployments where the namespace doesn't exist
  - Self-contained deployments where you want the module to handle all resources
  - Development and testing environments

- **Use Existing Namespace** (`create_namespace: false`): The module expects the namespace to already exist in the cluster. This is useful for:
  - Environments where namespace creation is controlled by platform teams
  - Multi-tenant clusters with pre-configured namespaces
  - Organizations with strict namespace governance policies
  - GitOps workflows where namespaces are managed separately

**Important**: When `create_namespace: false`, ensure the specified namespace exists before deploying the operator, otherwise the deployment will fail.

### Operator Configuration

#### Resource Management

- **Configurable Resources**: Define CPU and memory allocation for the operator pod to ensure optimal performance. The recommended default values are:
  - **CPU Requests**: `100m`
  - **Memory Requests**: `256Mi`
  - **CPU Limits**: `1000m` (1 CPU core)
  - **Memory Limits**: `1Gi`

#### Automated CRD Installation

- **CRD Management**: The operator automatically installs the required PostgreSQL Custom Resource Definitions (CRDs), including:
  - `PerconaPGCluster` - For deploying PostgreSQL clusters with high availability
  - `PerconaPGBackup` - For managing backups
  - `PerconaPGRestore` - For restore operations

#### Cluster-Wide Watching

- **Multi-Namespace Support**: The operator watches all namespaces by default, allowing PostgreSQL clusters to be deployed in any namespace within the cluster.

### Helm Chart Based Deployment

- **Official Percona Chart**: Deploys using the official Percona Helm chart from `https://percona.github.io/percona-helm-charts/`
- **Version Control**: Uses a specific, tested version of the operator ensuring stability and predictability.
- **Atomic Deployments**: Helm releases are deployed atomically with automatic rollback on failure.

## Benefits

- **Simplicity**: This resource streamlines operator deployment, making it easier for DevOps teams to set up PostgreSQL management capabilities in Kubernetes.
- **Foundation for PostgreSQL**: Provides the necessary operator infrastructure that enables deploying actual PostgreSQL clusters using the PostgreSQLKubernetes workload resource.
- **Production Ready**: Configured with sensible defaults and production-grade deployment patterns including proper timeouts, cleanup on failure, and resource limits.
- **Consistent Deployment**: Ensures the operator is deployed consistently across all environments using infrastructure-as-code principles.

## Use Cases

The Percona Operator for PostgreSQL is essential for teams that need to:

1. **Deploy PostgreSQL Clusters**: Run relational database workloads with ACID compliance and complex queries
2. **Manage PostgreSQL at Scale**: Operate multiple PostgreSQL clusters across different namespaces
3. **Automate Database Operations**: Leverage the operator for automated cluster management, scaling, backups, and updates
4. **High Availability**: Deploy PostgreSQL with automatic failover and disaster recovery capabilities
5. **Enterprise Workloads**: Support mission-critical applications requiring robust relational database capabilities

## Next Steps

After deploying the Percona Operator, you can:

1. Deploy PostgreSQL clusters using the `PostgreSQLKubernetes` workload resource
2. Create PostgreSQL databases for your applications
3. Configure high availability with automated failover
4. Set up automated backups and point-in-time recovery
5. Monitor operator and cluster health through Kubernetes tools

