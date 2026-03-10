# Overview

The **Percona Operator for MySQL Kubernetes API resource** is designed to deploy and manage the Percona Operator for MySQL within Kubernetes environments. This resource provides a consistent interface for deploying the operator that enables the management of MySQL database clusters on Kubernetes.

## Why We Created This API Resource

Deploying the Percona Operator for MySQL manually in Kubernetes can be complex, requiring specific knowledge of Helm charts, CRD installations, and operator configuration. The Percona Server MySQL Operator API resource simplifies this process by offering a standardized, declarative approach to deploying the operator. This resource allows teams to:

- **Efficiently Deploy the Operator**: Simplify the process of deploying the Percona Operator for MySQL on Kubernetes by abstracting away the complexities of Helm chart configurations.
- **Enable MySQL Management**: Once deployed, the operator enables teams to easily deploy and manage MySQL clusters with group replication, asynchronous replication, and high availability using the PerconaServerMySQL CRD.
- **Ensure Consistency**: Provide a standardized way to deploy the operator across different environments and clusters.
- **Optimize Resource Management**: Configure operator resource requirements to match your cluster's capacity.

## Key Features

### Kubernetes Cluster Integration

- **Target Cluster Configuration**: This resource integrates seamlessly with Planton's Kubernetes cluster credential management system, ensuring that the operator is deployed to the correct cluster.
- **Namespace Isolation**: The operator is deployed in a dedicated namespace for clean resource separation.

### Namespace Management

The component provides flexible namespace management through the `create_namespace` flag:

#### Automatic Namespace Creation (`create_namespace: true`)

When enabled, the component automatically creates the specified namespace before deploying the operator. This is the recommended approach for new deployments.

**Use cases:**
- Initial operator installation in a new cluster
- Development and testing environments
- Simplified deployment where namespace creation is delegated to the component

**Example:**
```yaml
spec:
  namespace:
    value: percona-mysql-operator
  create_namespace: true
```

#### Using Existing Namespace (`create_namespace: false`)

When disabled, the component expects the namespace to already exist in the cluster. The operator will be deployed into the existing namespace without attempting to create it.

**Use cases:**
- Namespaces managed separately (e.g., by platform team or GitOps)
- Namespaces with pre-configured policies, quotas, or network policies
- Multi-component deployments where namespace is shared
- Environments with strict RBAC where component lacks namespace creation permissions

**Prerequisites:**
- Namespace must exist before operator deployment
- Namespace must be accessible with provided credentials
- Verify namespace exists: `kubectl get namespace <namespace-name>`

**Example:**
```yaml
spec:
  namespace:
    value: percona-mysql-operator
  create_namespace: false
```

**Important:** If `create_namespace` is set to `false` and the namespace doesn't exist, the deployment will fail.

### Operator Configuration

#### Resource Management

- **Configurable Resources**: Define CPU and memory allocation for the operator pod to ensure optimal performance. The recommended default values are:
  - **CPU Requests**: `100m`
  - **Memory Requests**: `256Mi`
  - **CPU Limits**: `1000m` (1 CPU core)
  - **Memory Limits**: `1Gi`

#### Automated CRD Installation

- **CRD Management**: The operator automatically installs the required MySQL Custom Resource Definitions (CRDs), including:
  - `PerconaServerMySQL` - For deploying MySQL clusters with group replication or asynchronous replication
  - `PerconaServerMySQLBackup` - For managing backups
  - `PerconaServerMySQLRestore` - For restore operations

#### Cluster-Wide Watching

- **Multi-Namespace Support**: The operator watches all namespaces by default, allowing MySQL clusters to be deployed in any namespace within the cluster.

### Helm Chart Based Deployment

- **Official Percona Chart**: Deploys using the official Percona Helm chart from `https://percona.github.io/percona-helm-charts/`
- **Version Control**: Uses a specific, tested version of the operator ensuring stability and predictability.
- **Atomic Deployments**: Helm releases are deployed atomically with automatic rollback on failure.

## Benefits

- **Simplicity**: This resource streamlines operator deployment, making it easier for DevOps teams to set up MySQL management capabilities in Kubernetes.
- **Foundation for MySQL**: Provides the necessary operator infrastructure that enables deploying actual MySQL clusters using the MySQLKubernetes workload resource.
- **Production Ready**: Configured with sensible defaults and production-grade deployment patterns including proper timeouts, cleanup on failure, and resource limits.
- **Consistent Deployment**: Ensures the operator is deployed consistently across all environments using infrastructure-as-code principles.

## Use Cases

The Percona Operator for MySQL is essential for teams that need to:

1. **Deploy MySQL Clusters**: Run relational database workloads with ACID compliance
2. **Manage MySQL at Scale**: Operate multiple MySQL clusters across different namespaces
3. **Automate Database Operations**: Leverage the operator for automated cluster management, scaling, backups, and updates
4. **Run Application Workloads**: Support web applications, microservices, and transactional workloads
5. **High Availability**: Deploy MySQL with group replication for automatic failover and high availability

## Next Steps

After deploying the Percona Operator, you can:

1. Deploy MySQL clusters using the `MySQLKubernetes` workload resource
2. Create MySQL databases for your applications
3. Configure group replication or asynchronous replication clusters
4. Set up automated backups and monitoring
5. Monitor operator and cluster health through Kubernetes tools

