# MongoDB Kubernetes Pulumi Module

## Namespace Management

This module provides flexible namespace management through the `create_namespace` field in the spec:

- **`create_namespace: true`**: Creates a new namespace with resource labels for tracking
  - Ideal for new deployments and isolated environments
  - Namespace is automatically created before MongoDB resources
  
- **`create_namespace: false`**: Uses an existing namespace
  - Namespace must exist before applying this module
  - Suitable for environments with pre-configured policies, quotas, or RBAC
  - Allows sharing the namespace with other resources

## Key Features

### Operator-Based Deployment
The MongoDB Kubernetes module leverages the **Percona Server for MongoDB Operator** for production-grade database management. The operator provides enterprise-level features including automated upgrades, replica set management, backup capabilities, and self-healing.

### Unified API Resource Model
The MongoDB Kubernetes module adheres to a standard Kubernetes resource structure (`apiVersion`, `kind`, `metadata`, `spec`, and `status`), making it familiar to Kubernetes users. The `spec` defines critical aspects such as the MongoDB container, resource limits, persistence settings, and ingress configurations, providing granular control over the deployment.

### Key Features of the API Resource:
- **Deployment-Agnostic Design**: The specification is designed to work with different deployment implementations (Helm charts, operators, raw manifests). Currently implemented using Percona operator CRDs.
- **Replica Set Support**: Automatically configures MongoDB replica sets for high availability. The `replicas` field maps to Percona's replica set size.
- **Persistence**: The `persistence_enabled` flag allows you to enable or disable persistence for MongoDB data. When enabled, a persistent volume is attached to each MongoDB pod to store data, which is restored after restarts.
- **Resource Management**: Developers can define CPU and memory limits for MongoDB pods, ensuring that the cluster is configured with optimal resource utilization.
- **Ingress Support**: The module provides the ability to configure ingress for MongoDB clusters, enabling external access to the MongoDB service. This can be particularly useful for hybrid cloud environments or services requiring external data access.
- **Automated Password Generation**: A random password for MongoDB authentication is generated and securely stored in a Kubernetes secret, simplifying the process of credential management.

### Key Features of the Pulumi Module:
- **Dynamic Resource Creation**: The module dynamically creates Kubernetes resources based on the `MongodbKubernetesStackInput`, including namespaces, PerconaServerMongoDB CRDs, services, and persistent volumes.
- **Kubernetes Provider Integration**: The module uses Pulumi's Kubernetes provider to manage resources and interact with the Kubernetes cluster using provided credentials.
- **Operator-Based Management**: Deploys MongoDB using the Percona operator, which provides automated lifecycle management, failover, and recovery.
- **CRD-Based Deployment**: Creates `PerconaServerMongoDB` custom resources that the operator reconciles into running MongoDB clusters.
- **Status Outputs**: After resource creation, the module provides key outputs such as internal and external service endpoints, port forwarding commands, and the MongoDB username. These outputs are critical for accessing the MongoDB service and managing operations like backups or monitoring.

## Prerequisites

The **Percona Server for MongoDB Operator** must be installed on your Kubernetes cluster before deploying MongoDB instances. Use the `PerconaServerMongodbOperator` module to install the operator:

```bash
planton pulumi up --manifest percona-operator.yaml \
  --module-dir apis/project/planton/provider/kubernetes/perconaservermongodboperator/v1/iac/pulumi
```

The operator typically installs in the `mongodb-operator` namespace and watches all namespaces for `PerconaServerMongoDB` resources.

## Status and Outputs

The module provides the following outputs to simplify the operational management of the MongoDB Kubernetes cluster:
- **Namespace**: The Kubernetes namespace in which the MongoDB cluster is created, allowing for resource isolation.
- **Service Name**: The Kubernetes service name associated with the MongoDB cluster, which can be used for internal communication.
- **Port-Forwarding Command**: A Kubernetes command for setting up port-forwarding, enabling local access to the MongoDB cluster when ingress is disabled.
- **Internal and External Endpoints**: URLs for accessing MongoDB, both from within the Kubernetes cluster and externally if ingress is enabled.
- **MongoDB Credentials**: The MongoDB username and a reference to the Kubernetes secret containing the password are provided, ensuring secure access to the MongoDB service.

## Usage

To deploy and manage a MongoDB Kubernetes cluster using this module, create a YAML file representing the MongoDB Kubernetes resource. Use the CLI command `planton pulumi up --stack-input <api-resource.yaml>` to apply the configuration and provision the resources.

Refer to the example section for usage instructions.

## Requirements

- **Pulumi**: Ensure that Pulumi is installed and configured for your environment.
- **Kubernetes Provider**: The Kubernetes provider is required to manage resources on the Kubernetes cluster where MongoDB will be deployed.
- **Percona Operator**: The Percona Server for MongoDB Operator must be installed on your cluster before deploying MongoDB instances.
- **Kubernetes Credential**: This is required to authenticate and interact with the target Kubernetes cluster.

## Architecture

The deployment follows a modern, operator-based architecture:

```
┌─────────────────────────────────────────────────────────────┐
│  Kubernetes Cluster                                         │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐ │
│  │  mongodb-operator namespace                           │ │
│  │  ┌─────────────────────────────────────────────────┐  │ │
│  │  │  Percona MongoDB Operator                       │  │ │
│  │  │  (watches PerconaServerMongoDB CRDs)            │  │ │
│  │  └─────────────────────────────────────────────────┘  │ │
│  └───────────────────────────────────────────────────────┘ │
│                                                             │
│  ┌───────────────────────────────────────────────────────┐ │
│  │  Application Namespace (e.g., my-mongodb)             │ │
│  │                                                         │ │
│  │  1. Pulumi Module Creates:                             │ │
│  │  ┌─────────────────────────────────────────────────┐  │ │
│  │  │  PerconaServerMongoDB CRD                        │  │ │
│  │  │  (replica set config, resources, persistence)    │  │ │
│  │  └─────────────────────────────────────────────────┘  │ │
│  │  ┌─────────────────────────────────────────────────┐  │ │
│  │  │  Secret (auto-generated password)                │  │ │
│  │  └─────────────────────────────────────────────────┘  │ │
│  │                                                         │ │
│  │  2. Operator Reconciles to Create:                     │ │
│  │  ┌─────────────────────────────────────────────────┐  │ │
│  │  │  StatefulSets (MongoDB pods)                     │  │ │
│  │  │  ├─ Pod with persistent volumes                  │  │ │
│  │  │  ├─ Pod with persistent volumes                  │  │ │
│  │  │  └─ ...                                          │  │ │
│  │  └─────────────────────────────────────────────────┘  │ │
│  │  ┌─────────────────────────────────────────────────┐  │ │
│  │  │  Services (cluster communication)                │  │ │
│  │  └─────────────────────────────────────────────────┘  │ │
│  │  ┌─────────────────────────────────────────────────┐  │ │
│  │  │  ConfigMaps (MongoDB configuration)              │  │ │
│  │  └─────────────────────────────────────────────────┘  │ │
│  │  ┌─────────────────────────────────────────────────┐  │ │
│  │  │  LoadBalancer Service (optional ingress)         │  │ │
│  │  └─────────────────────────────────────────────────┘  │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Benefits

- **Production-Ready**: Leverage the battle-tested Percona operator used by enterprises worldwide.
- **Simplified Operations**: The operator handles complex lifecycle operations: upgrades, scaling, backups, and recovery.
- **High Availability**: Automatic replica set configuration ensures database availability even during node failures.
- **Automated Failover**: The operator automatically promotes replicas when primaries fail.
- **Declarative Management**: Define desired state, operator ensures actual state matches.
- **Self-Healing**: Operator automatically recovers from failures.
- **Cloud-Native Architecture**: CRD-based deployment follows Kubernetes best practices.

## Installation

You can install this Pulumi module from GitHub by cloning the repository and running the required Pulumi commands to set up the infrastructure. Ensure you have all necessary dependencies installed, including Pulumi and Kubernetes access.

1. Clone the repository:
    ```bash
    git clone https://github.com/plantonhq/planton.git
    cd planton/apis/project/planton/provider/kubernetes/mongodbkubernetes/v1/iac/pulumi
    ```

2. Install dependencies and initialize Pulumi:
    ```bash
    pulumi stack init <stack-name>
    pulumi config set <config-parameters>
    ```

## Contributing

Contributions are welcome to enhance the functionality of this Pulumi module. If you encounter any issues or want to suggest new features, feel free to submit a pull request or open an issue on the GitHub repository.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.
