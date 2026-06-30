# KubernetesOpenBao Pulumi Module

## Key Features

### Standardized API Resource Structure
- **apiVersion & kind**: Aligns with Kubernetes standards, ensuring familiarity and ease of integration.
- **metadata**: Facilitates resource identification and management through standard Kubernetes metadata fields.
- **spec**: Defines the desired state of the OpenBao deployment, including server, HA, injector, and ingress configurations.
- **status**: Provides real-time updates and outputs from the deployed infrastructure, enhancing visibility and monitoring.

### Comprehensive OpenBao Configuration
- **Namespace Management**: Control namespace creation through the `create_namespace` flag. When set to `true`, the module creates a dedicated namespace with proper labels. When `false`, it uses an existing namespace, allowing integration with pre-configured namespace policies and quotas.
- **Deployment Modes**: Support for both standalone (development/testing) and high-availability (production) deployment modes.
- **Raft Integrated Storage**: Built-in consensus protocol for HA deployments with automatic leader election and data replication.
- **Resource Allocation**: Define CPU and memory resources for OpenBao server containers to optimize performance and cost.
- **Persistent Storage**: Configurable persistent volume sizes for data storage across both standalone and HA modes.
- **Agent Injector**: Optional deployment of the mutating webhook for automatic secret injection into application pods.
- **Ingress Specifications**: Configure ingress rules to manage external access to OpenBao API and UI.

### Seamless Cloud Integration
- **Multi-Cloud Support**: Deploy OpenBao instances across different cloud providers with ease, leveraging Pulumi's cloud-agnostic capabilities.
- **Kubernetes Credentials Management**: Securely manage Kubernetes cluster credentials, ensuring safe and efficient resource provisioning.
- **Automated Outputs Handling**: Capture and manage Pulumi outputs within the API resource status, providing essential information such as service endpoints and secret references.

### Developer-Friendly CLI
- **Unified Deployment Command**: Utilize the `planton pulumi up --stack-input <api-resource.yaml>` command to deploy complex infrastructures effortlessly.
- **Default Module Configuration**: Automatically configure stack inputs using default Pulumi modules, reducing setup complexity for developers.
- **Git Integration**: Specify custom Pulumi modules via Git repository details, allowing for flexible and customized deployments.

### Robust Documentation and Validation
- **Buf.Build Documentation**: Access comprehensive documentation through buf.build, ensuring clear and accessible guidance for developers.
- **Validation Rules**: Implement validation rules within the API resource specifications to enforce correct configurations and prevent deployment errors.

## Installation

To install the KubernetesOpenBao Pulumi module, follow the standard Pulumi module installation procedures. Ensure that you have Pulumi and the necessary dependencies installed on your development machine.

## Usage

Refer to the example section for usage instructions.

## API Reference

### KubernetesOpenBaoSpec
Defines the desired state of the OpenBao deployment, including server, HA, injector, and ingress configurations.

- **target_cluster**: Selector for the target Kubernetes cluster.
- **namespace**: Kubernetes namespace for the deployment (supports value or reference).
- **create_namespace**: Flag to control namespace creation.
- **helm_chart_version**: Optional override for the Helm chart version (default: 0.23.3).
- **server_container**: Container specifications including replicas, resources, and storage.
- **high_availability**: HA configuration with Raft integrated storage.
- **ingress**: Ingress specifications for external access.
- **ui_enabled**: Flag to enable the OpenBao UI (default: true).
- **injector**: Agent Injector configuration for automatic secret injection.
- **tls_enabled**: Flag to enable TLS encryption.

### KubernetesOpenBaoServerContainer
Specifies the container-level configurations for the OpenBao server.

- **replicas**: Number of OpenBao server pods to deploy.
- **resources**: CPU and memory resource allocations.
- **data_storage_size**: Size of the persistent volume for data storage.

### KubernetesOpenBaoHighAvailability
Configures High Availability mode with Raft integrated storage.

- **enabled**: Flag to enable HA mode.
- **replicas**: Number of HA replicas (default: 3, recommended odd numbers).

### KubernetesOpenBaoInjector
Configures the OpenBao Agent Injector.

- **enabled**: Flag to enable the injector.
- **replicas**: Number of injector replicas (default: 1).

### KubernetesOpenBaoStackOutputs
Provides outputs from the deployed OpenBao infrastructure.

- **namespace**: Kubernetes namespace where OpenBao is deployed.
- **service**: Kubernetes service name for OpenBao.
- **port_forward_command**: Command to set up port-forwarding for local access.
- **kube_endpoint**: Internal Kubernetes endpoint for OpenBao.
- **external_hostname**: Public endpoint for external access (when ingress is enabled).
- **api_address**: Full API address for OpenBao.
- **cluster_address**: Cluster communication address (HA mode).
- **ha_enabled**: Boolean indicating if HA mode is enabled.
- **root_token_secret**: Reference to the root token Kubernetes secret.
- **unseal_keys_secret**: Reference to the unseal keys Kubernetes secret.

## Contributing

Contributions are welcome! Please refer to the contributing guidelines for more information on how to get involved.

## License

This project is licensed under the [MIT License](LICENSE).
