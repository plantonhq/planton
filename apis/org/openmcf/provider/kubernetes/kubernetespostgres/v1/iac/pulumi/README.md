# PostgresKubernetes Pulumi Module

## Key Features

### API-Resource Highlights

- **Standardized Structure**: Adheres to the Kubernetes API resource model, including `apiVersion`, `kind`, `metadata`, `spec`, and `status`, ensuring consistency and ease of integration.
- **Unified API Approach**: Designed for the multi-cloud era, enabling seamless interaction with various cloud providers through a consistent API structure.
- **Comprehensive Configuration**: Supports detailed configurations for PostgreSQL deployments, including replica counts, resource allocations, disk sizes, and ingress settings.
- **Flexible Storage Options**: Allows specification of disk sizes with validation to ensure appropriate storage allocation for PostgreSQL instances.
- **Ingress Management**: Integrated support for customizable ingress settings, enabling secure external and internal access to the PostgreSQL service.

### Module Features

- **Pulumi Integration**: Developed in Go, the module integrates seamlessly with Pulumi, leveraging its powerful IaC capabilities for managing PostgreSQL infrastructure within Kubernetes.
- **Automated Namespace Creation**: Automatically creates and manages a dedicated Kubernetes namespace for the PostgreSQL deployment, ensuring resource isolation and organizational clarity.
- **Zalando PostgreSQL Operator Deployment**: Utilizes the Zalando PostgreSQL Operator to deploy and manage highly available PostgreSQL instances, ensuring reliability and scalability.
- **Kubernetes Provider Configuration**: Dynamically sets up the Kubernetes provider using cluster credentials, facilitating secure and efficient connections to target Kubernetes clusters.
- **Customizable Ingress Setup**: Supports the creation of external and internal load balancer services based on ingress configurations, enhancing accessibility and security.
- **Secret Management**: Manages Kubernetes secrets for PostgreSQL credentials, ensuring sensitive information is securely stored and easily accessible for authorized services.
- **Exported Outputs**: Provides detailed outputs such as namespace name, service endpoints, port-forwarding commands, and secret keys, enabling seamless integration with other tools and systems.
- **Extensible and Modular Design**: Designed to be easily extensible, allowing for the incorporation of additional functionalities and integrations as needed.
- **Consistent Deployment Pattern**: Utilizes a standardized pattern where API resources serve as inputs to Pulumi modules, promoting consistency and reducing complexity in infrastructure deployments.

## Installation

To install the PostgresKubernetes Pulumi Module, follow the standard Pulumi module installation procedures. Ensure that you have Pulumi installed and configured in your development environment. The module is available on GitHub, and you can include it in your Pulumi projects by referencing the repository in your project’s dependencies.

```shell
pulumi plugin install resource postgreskubernetes <version>
```

Replace `<version>` with the desired version of the module.

## Configuration

The module requires a `PostgresKubernetesStackInput` specification to define the desired state of the PostgreSQL deployment within a Kubernetes cluster. This specification includes configurations for PostgreSQL container resources, replica counts, disk sizes, and ingress settings. Additionally, Kubernetes cluster credentials must be provided to enable the module to interact with the target Kubernetes cluster.

### Key Configuration Parameters

- **Kubernetes Credentials**: Provide the necessary credentials to authenticate and interact with the target Kubernetes cluster by specifying the `kubernetes_credential_id`.
- **PostgreSQL Container Configuration**: Define the number of replicas, resource allocations (CPU and memory), and disk size for the PostgreSQL containers to ensure optimal performance and scalability.
- **User Configuration**: Declare PostgreSQL users/roles via the `users` list, where each user has a `name` and optional `flags` (e.g., `["createdb"]`). Users must be declared before being used as database owners.
- **Database Configuration**: Specify databases to create during cluster initialization via the `databases` map, where keys are database names and values are owner role names. Owner roles must be declared in `users`.
- **Ingress Settings**: Configure ingress specifications to manage external and internal access to the PostgreSQL service, including load balancer types and hostname configurations.
- **Pulumi Input**: Define Pulumi-specific configurations required for the stack-update, facilitating the integration between the API resource and Pulumi's deployment processes.
- **Target API-Resource**: Specify the `PostgresKubernetes` target, linking the API resource definition with the Pulumi module for deployment.

## Usage

Refer to example section for usage instructions.

## Outputs

Upon successful execution, the PostgresKubernetes module exports several key outputs that are essential for managing and accessing the PostgreSQL service:

- **Namespace**: The name of the Kubernetes namespace where PostgresKubernetes is deployed.
- **Service**: The Kubernetes service name created for PostgresKubernetes.
- **Port Forward Command**: A command to set up port-forwarding for local access to PostgreSQL, which is useful when ingress is disabled for security reasons.
- **Kube Endpoint**: The internal Kubernetes endpoint for accessing PostgreSQL from within the same Kubernetes cluster.
- **External Hostname**: The public endpoint for accessing PostgreSQL from outside the Kubernetes cluster.
- **Username Secret**: Kubernetes secret key for the PostgreSQL username.
- **Password Secret**: Kubernetes secret key for the PostgreSQL password.

These outputs facilitate seamless integration with other systems and tools, enabling efficient operational management and monitoring of the PostgreSQL deployment.

## Documentation

Comprehensive documentation for the API resources and module is available via [buf.build](https://buf.build). This documentation provides detailed insights into the API structure, configuration options, and operational guidelines, ensuring that developers can effectively utilize the module to deploy and manage PostgreSQL services within the Planton ecosystem.

## Contributing

Contributions to the PostgresKubernetes Pulumi Module are welcome. Please refer to the `CONTRIBUTING.md` file in the repository for guidelines on how to contribute, report issues, and suggest enhancements. Your contributions help improve the module and expand its capabilities to better serve the community.

## License

This project is licensed under the [MIT License](LICENSE).