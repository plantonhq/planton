# RedisKubernetes Pulumi Module

## Key Features

### Standardized API Resource Structure
- **apiVersion & kind**: Aligns with Kubernetes standards, ensuring familiarity and ease of integration.
- **metadata**: Facilitates resource identification and management through standard Kubernetes metadata fields.
- **spec**: Defines the desired state of the Redis deployment, including container specifications and ingress configurations.
- **status**: Provides real-time updates and outputs from the deployed infrastructure, enhancing visibility and monitoring.

### Comprehensive Redis Configuration
- **Namespace Management**: Control namespace creation through the `create_namespace` flag. When set to `true`, the module creates a dedicated namespace with proper labels. When `false`, it uses an existing namespace, allowing integration with pre-configured namespace policies and quotas.
- **Replicas Management**: Specify the number of Redis pods to ensure high availability and load distribution.
- **Resource Allocation**: Define CPU and memory resources for each Redis container to optimize performance and cost.
- **Persistence Control**: Toggle data persistence to ensure data durability across pod restarts, with configurable disk sizes for persistent volumes.
- **Ingress Specifications**: Configure ingress rules to manage external access to Redis services, enhancing security and accessibility.

### Seamless Cloud Integration
- **Multi-Cloud Support**: Deploy Redis instances across different cloud providers with ease, leveraging Pulumi's cloud-agnostic capabilities.
- **Kubernetes Credentials Management**: Securely manage Kubernetes cluster credentials, ensuring safe and efficient resource provisioning.
- **Automated Outputs Handling**: Capture and manage Pulumi outputs within the API resource status, providing essential information such as service endpoints and authentication details.

### Developer-Friendly CLI
- **Unified Deployment Command**: Utilize the `planton pulumi up --stack-input <api-resource.yaml>` command to deploy complex infrastructures effortlessly.
- **Default Module Configuration**: Automatically configure stack inputs using default Pulumi modules, reducing setup complexity for developers.
- **Git Integration**: Specify custom Pulumi modules via Git repository details, allowing for flexible and customized deployments.

### Robust Documentation and Validation
- **Buf.Build Documentation**: Access comprehensive documentation through buf.build, ensuring clear and accessible guidance for developers.
- **Validation Rules**: Implement validation rules within the API resource specifications to enforce correct configurations and prevent deployment errors.

## Installation

To install the RedisKubernetes Pulumi module, follow the standard Pulumi module installation procedures. Ensure that you have Pulumi and the necessary dependencies installed on your development machine.

## Usage

Refer to the example section for usage instructions.

## API Reference

### RedisKubernetesSpec
Defines the desired state of the Redis deployment, including container specifications and ingress settings.

- **kubernetes_credential_id**: Identifier for the Kubernetes cluster credentials to be used.
- **container**: Configuration for the Redis container, including replicas, resource allocations, and persistence settings.
- **ingress**: Ingress specifications for managing external access to Redis services.

### RedisKubernetesContainer
Specifies the container-level configurations for Redis.

- **replicas**: Number of Redis pods to deploy.
- **resources**: CPU and memory resource allocations for each Redis container.
- **persistence_enabled**: Flag to enable or disable data persistence.
- **disk_size**: Size of the persistent volume attached to each Redis pod.

### RedisKubernetesStackOutputs
Provides outputs from the deployed Redis infrastructure.

- **namespace**: Kubernetes namespace where Redis is deployed.
- **service**: Kubernetes service name for Redis.
- **port_forward_command**: Command to set up port-forwarding for local access.
- **kube_endpoint**: Internal Kubernetes endpoint for Redis.
- **external_hostname**: Public endpoint for external access (when ingress is enabled).
- **username**: Redis authentication username.
- **password_secret**: Kubernetes secret key for Redis password.

## Contributing

Contributions are welcome! Please refer to the contributing guidelines for more information on how to get involved.

## License

This project is licensed under the [MIT License](LICENSE).