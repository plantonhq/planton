# Neo4j Kubernetes Pulumi Module

## Key Features

### Unified API Resource Model
The Neo4j Kubernetes Pulumi module adheres to a standardized API resource model, ensuring a consistent approach to managing infrastructure on Kubernetes. Each resource follows a familiar structure (`apiVersion`, `kind`, `metadata`, `spec`, and `status`), allowing developers to easily integrate and manage their Neo4j instances in line with Kubernetes best practices.

### Key Features of the API Resource:
- **Resource Configuration**: Developers can specify CPU and memory resource requests and limits for the Neo4j container. This ensures that the database operates efficiently and scales appropriately based on the available resources within the Kubernetes cluster.
- **Ingress Support**: The module includes optional ingress configuration, allowing the Neo4j instance to be accessed from outside the Kubernetes cluster. This is particularly useful for external clients or hybrid cloud architectures where external connectivity is required.
- **Namespace Isolation**: Each deployment of Neo4j is placed within its own Kubernetes namespace, ensuring that resources are logically isolated. This is especially beneficial in multi-tenant environments or when managing multiple instances of Neo4j in a single cluster.
- **Internal and External Access**: By configuring Kubernetes services, the module enables both internal (within-cluster) and external (outside-cluster) access to the Neo4j instance. Depending on the ingress configuration, the database can be exposed externally via a public endpoint.

### Key Features of the Pulumi Module:
- **Dynamic Resource Creation**: The module dynamically creates Kubernetes resources based on the specifications in the `Neo4jKubernetesStackInput`, including services, namespaces, and ingress (if enabled).
- **Kubernetes Provider Integration**: Leveraging the Pulumi Kubernetes provider, the module interacts with the cluster to create and manage resources such as Neo4j services and network configurations.
- **Port-Forwarding for Local Access**: When ingress is disabled, the module provides a port-forwarding command that developers can use to access the Neo4j instance locally via `kubectl`. This is useful for development and testing scenarios.
- **Status Outputs**: After the deployment, the module provides key output values such as the Kubernetes service name, internal and external endpoints, and the port-forwarding command. These outputs simplify the management and interaction with the Neo4j instance post-deployment.

## Status and Outputs

After deploying Neo4j using this module, the following outputs are provided to facilitate management and operations:
- **Namespace**: The Kubernetes namespace in which the Neo4j instance is deployed, allowing for resource isolation.
- **Service Name**: The Kubernetes service name for accessing the Neo4j instance within the cluster.
- **Port-Forwarding Command**: A command to enable port-forwarding for local access to the Neo4j instance, useful when ingress is not enabled.
- **Internal Endpoint**: The internal URL to access the Neo4j service within the Kubernetes cluster.
- **External Endpoint**: (If ingress is enabled) The external URL to access the Neo4j instance from outside the Kubernetes cluster.

## Usage

To deploy and manage a Neo4j Kubernetes instance using this module, create a YAML file representing the Neo4j Kubernetes resource. Use the CLI command `planton pulumi up --stack-input <api-resource.yaml>` to apply the configuration and provision the resources.

Refer to the example section for usage instructions.

## Requirements

- **Pulumi**: Ensure that Pulumi is installed and configured in your environment.
- **Kubernetes Provider**: The Kubernetes provider is required to manage and interact with the Kubernetes cluster.
- **Kubernetes Credential**: This credential is required to authenticate and interact with the target Kubernetes cluster where the Neo4j instance will be deployed.
- **Neo4j License**: If required, ensure that you have the necessary Neo4j license, depending on the environment and usage.

## Installation

You can install and use this Pulumi module by cloning the repository from GitHub and running the necessary Pulumi commands to set up the infrastructure. Make sure you have the correct dependencies installed and access to your Kubernetes cluster.

1. Clone the repository:
    ```bash
    git clone https://github.com/your-repo/planton-pulumi-neo4j-kubernetes.git
    cd planton-pulumi-neo4j-kubernetes
    ```

2. Install dependencies and initialize Pulumi:
    ```bash
    pulumi stack init <stack-name>
    pulumi config set <config-parameters>
    ```

## Contributing

We welcome contributions to improve and expand the functionality of this Pulumi module. If you encounter any issues or have suggestions for new features, feel free to open an issue or submit a pull request on the GitHub repository.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.
