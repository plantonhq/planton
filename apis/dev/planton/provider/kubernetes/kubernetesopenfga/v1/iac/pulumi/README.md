# OpenFGA Kubernetes Pulumi Module

## Key Features

### API-Resource Highlights

- **Standardized Structure**: Each API resource follows the Kubernetes API resource model, including `apiVersion`, `kind`, `metadata`, `spec`, and `status`, ensuring consistency and ease of integration.
- **Comprehensive Configuration**: The `OpenFgaKubernetesSpec` allows detailed configuration of the OpenFGA deployment, including container resources, replica counts, ingress settings, and datastore configurations.
- **Flexible Datastore Options**: Supports both MySQL and PostgreSQL as datastore engines, with configurable connection URIs tailored to the selected engine.
- **Ingress Management**: Integrated support for Istio ingress resources, enabling secure and manageable external access to the OpenFGA service.
- **Status Outputs**: Captures and exports essential deployment outputs such as namespace details, service endpoints, port-forwarding commands, and hostname configurations for both internal and external access.

### Module Features

- **Pulumi Integration**: Written in Go, the module seamlessly integrates with Pulumi, allowing developers to leverage familiar Pulumi workflows and tooling for infrastructure management.
- **Automated Namespace Creation**: Automatically creates and manages a dedicated Kubernetes namespace for the OpenFGA deployment, ensuring resource isolation and organizational clarity.
- **Helm Chart Deployment**: Utilizes Helm charts to install and configure the OpenFGA service, simplifying the deployment process and ensuring adherence to best practices.
- **Kubernetes Provider Configuration**: Dynamically sets up the Kubernetes provider using cluster credentials, facilitating deployments across different Kubernetes clusters and cloud providers.
- **Extensible and Modular**: Designed to be easily extensible, allowing for customization and integration with additional tools and services as needed.
- **Exported Outputs**: Provides detailed outputs that can be used for further automation, monitoring, and integration with other systems, enhancing operational efficiency.

## Installation

To install the OpenFGA Kubernetes Pulumi Module, follow the standard Pulumi module installation procedures. Ensure that you have Pulumi installed and configured in your development environment. The module is available on GitHub, and you can include it in your Pulumi projects by referencing the repository in your project’s dependencies.

## Configuration

The module requires an `OpenFgaKubernetesStackInput` specification to define the desired state of the OpenFGA deployment. This specification includes configurations for container resources, replicas, ingress settings, and datastore details. Additionally, Kubernetes cluster credentials must be provided to enable the module to interact with the target Kubernetes cluster.

### Key Configuration Parameters

- **Container Resources**: Define CPU and memory requests and limits for the OpenFGA containers.
- **Replicas**: Specify the number of replicas for the OpenFGA deployment to ensure high availability and scalability.
- **Ingress Settings**: Configure Istio ingress resources to manage external access to the OpenFGA service.
- **Datastore Configuration**: Choose between MySQL and PostgreSQL for the datastore engine and provide the corresponding connection URI.
- **Kubernetes Credentials**: Provide the necessary credentials to authenticate and interact with the target Kubernetes cluster.

## Usage

Refer to the example section for usage instructions.

## Outputs

Upon successful deployment, the module exports several key outputs that are essential for managing and accessing the OpenFGA service:

- **Namespace**: The name of the Kubernetes namespace where OpenFGA is deployed.
- **Service**: The Kubernetes service name for OpenFGA.
- **Port Forward Command**: A command to set up port-forwarding for local access to OpenFGA.
- **Kube Endpoint**: The internal Kubernetes endpoint for accessing OpenFGA.
- **External Hostname**: The public endpoint for accessing OpenFGA from outside the Kubernetes cluster.
- **Internal Hostname**: The private endpoint for accessing OpenFGA from within the Kubernetes cluster.

These outputs facilitate seamless integration with other systems and tools, enabling efficient operational management and monitoring of the OpenFGA deployment.

## Documentation

Comprehensive documentation for the API resources and module is available via [buf.build](https://buf.build). This documentation provides detailed insights into the API structure, configuration options, and operational guidelines, ensuring that developers can effectively utilize the module to deploy and manage OpenFGA services.

## Contributing

Contributions to the OpenFGA Kubernetes Pulumi Module are welcome. Please refer to the CONTRIBUTING.md file in the repository for guidelines on how to contribute, report issues, and suggest enhancements.

## License

This project is licensed under the [MIT License](LICENSE).