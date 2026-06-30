# ClickHouse Kubernetes Pulumi Module

## Overview

This module deploys production-grade ClickHouse clusters on Kubernetes using the **Altinity ClickHouse Operator**. The operator provides enterprise-level features including automated upgrades, scaling, backup, and recovery.

## Key Features

- **Operator-Based Deployment**  
  Uses the Altinity ClickHouse Operator for production-grade cluster management. The operator handles complex operations like rolling upgrades, scaling, and failure recovery automatically.

- **Standardized API Resource Model**  
  Provides an intuitive, deployment-agnostic API for defining ClickHouse clusters. Focus on what you need (resources, persistence, clustering) without worrying about underlying implementation details.

- **Automated CRD Management**  
  Automatically generates and applies ClickHouseInstallation custom resources. The operator reconciles these to create StatefulSets, Services, ConfigMaps, and other Kubernetes resources.

- **Flexible Cluster Topologies**  
  Deploy standalone instances for development or distributed clusters with configurable sharding and replication for production workloads.

- **Production-Grade Features**  
  - ClickHouse Keeper coordination (75% more efficient than ZooKeeper)
  - Rolling updates with zero downtime
  - Persistent storage with configurable sizes
  - Resource limits and requests for optimal performance
  - Built-in monitoring and metrics
  - Flexible coordination options (auto-managed Keeper, external Keeper, or ZooKeeper)

- **Security Best Practices**  
  Automatically generates secure random passwords stored in Kubernetes Secrets. Credentials never appear in manifests or version control.

- **Ingress Integration**  
  Optional LoadBalancer service with external DNS annotations for external access. Specify a custom hostname for your ClickHouse cluster (e.g., `clickhouse.example.com`).

- **Output Exports**  
  Exports namespace, service names, endpoints, credentials, and port-forward commands for easy integration and debugging.

## Usage

See [examples.md](examples.md) for usage details and step-by-step examples. In general:

1. Define a YAML resource describing your ClickHouse cluster using the **ClickHouseKubernetes** API.
2. Run:
   ```bash
   planton pulumi up --stack-input <your-clickhouse-file.yaml>
   ```

to apply the resource on your cluster.

## Prerequisites

The **Altinity ClickHouse Operator** must be installed on your Kubernetes cluster before deploying ClickHouse instances. Use the `ClickhouseOperatorKubernetes` module to install the operator:

```bash
planton pulumi up --stack-input clickhouse-operator.yaml \
  --module-dir apis/project/planton/provider/kubernetes/clickhouseoperatorkubernetes/v1/iac/pulumi
```

The operator typically installs in the `clickhouse-operator` namespace and watches all namespaces for ClickHouseInstallation resources.

## Getting Started

1. **Install the Operator** (one-time setup per cluster)  
   Deploy the Altinity ClickHouse Operator using the operator deployment module.

2. **Define Your ClickHouse Cluster**  
   Create a YAML specification with cluster name, resources, persistence, and clustering settings. See [examples.md](examples.md) for common configurations.

3. **Deploy the Cluster**  
   Execute `planton pulumi up --stack-input <clickhouse-spec.yaml>`. The module generates a ClickHouseInstallation CRD, and the operator creates all necessary Kubernetes resources.

4. **Verify Deployment**  
   Check that ClickHouse pods are running, services are created, and the cluster is accessible. Use the exported port-forward command for local testing.

## Module Architecture

1. **Initialization**  
   Reads your `ClickHouseKubernetesStackInput` (cluster credentials, resource definitions), initializes local variables, and prepares Kubernetes labels.

2. **Provider Setup**  
   Establishes a Pulumi Kubernetes Provider using the supplied cluster credentials.

3. **Namespace Management**  
   Conditionally creates or references a namespace based on the `create_namespace` flag:
   - When `create_namespace: true`: Creates a dedicated namespace with appropriate labels for resource tracking
   - When `create_namespace: false`: Uses an existing namespace (must exist before deployment)
   
   This flexibility allows you to either let the component manage the namespace lifecycle or deploy into a shared/externally-managed namespace.

4. **Secret Management**  
   Generates a cryptographically secure random password and stores it in a Kubernetes Secret for ClickHouse authentication.

5. **ClickHouseInstallation CRD Generation**  
   Builds a ClickHouseInstallation custom resource with:
   - Cluster topology (shards, replicas)
   - Resource allocations (CPU, memory)
   - Persistence configuration (disk size, storage class)
   - Coordination settings (ClickHouse Keeper or ZooKeeper references)
   - Security settings (password references)

6. **Coordination Service Setup** (Optional)  
   For auto-managed ClickHouse Keeper, the operator will create a ClickHouseKeeperInstallation resource automatically. For external coordination, references the provided nodes.

7. **CRD Application**  
   Applies the ClickHouseInstallation to Kubernetes. The Altinity operator watches for these resources and reconciles the actual state:
   - Creates StatefulSets for ClickHouse pods
   - Sets up Services for cluster communication
   - Configures ConfigMaps with ClickHouse settings
   - Manages ClickHouse Keeper (for auto-managed coordination)

8. **Ingress Service (Optional)**  
   If enabled, creates a LoadBalancer Service with external DNS annotations for public access.

9. **Output Exports**  
   Exports useful values: namespace, service names, endpoints, credentials, and port-forward commands.

## Benefits

- **Production-Ready**  
  Leverage the battle-tested Altinity operator used by enterprises worldwide. Benefit from years of operational experience built into the operator.

- **Simplified Operations**  
  The operator handles complex lifecycle operations: upgrades, scaling, backups, and recovery. You focus on your data, not Kubernetes complexity.

- **Cloud-Native Architecture**  
  CRD-based deployment follows Kubernetes best practices. Declarative configuration ensures reproducible deployments across environments.

- **Flexibility at Scale**  
  Start with a single node for development, seamlessly scale to distributed clusters with hundreds of nodes for production analytics.

- **Vendor Independence**  
  Uses official ClickHouse container images from clickhouse.com. No dependency on deprecated or commercial registries.

## Contributing

Contributions are always welcome! Please open an issue or submit a pull request in the main repository if you want to add features, fix bugs, or improve documentation.

## License

This project is licensed under the [MIT License](LICENSE). Feel free to adapt it for your internal workflows.
