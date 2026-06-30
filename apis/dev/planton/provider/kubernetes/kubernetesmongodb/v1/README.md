# Overview

The **Mongodb Kubernetes API Resource** provides a standardized and efficient way to deploy MongoDB onto Kubernetes clusters. This API resource simplifies the deployment process by encapsulating all necessary configurations, enabling consistent and repeatable MongoDB deployments across various environments.

## Namespace Management

This component provides flexible namespace management through the `create_namespace` field in the spec:

- **`create_namespace: true`**: Creates a new namespace with resource labels for tracking
  - Ideal for new deployments and isolated environments
  - Namespace is automatically created before MongoDB resources
  
- **`create_namespace: false`**: Uses an existing namespace
  - Namespace must exist before applying this component
  - Suitable for environments with pre-configured policies, quotas, or RBAC
  - Allows sharing the namespace with other resources

## Purpose

Deploying MongoDB on Kubernetes involves complex configurations, including resource management, storage persistence, and environment settings. The Mongodb Kubernetes API Resource aims to:

- **Standardize Deployments**: Offer a consistent interface for deploying MongoDB, reducing complexity and minimizing errors.
- **Simplify Configuration Management**: Centralize all deployment settings, making it easier to manage, update, and replicate configurations.
- **Enhance Flexibility and Scalability**: Allow granular control over various components like replicas, resources, and persistence to meet specific requirements.

## Key Features

### Environment Configuration

- **Environment Info**: Tailor MongoDB deployments to specific environments (development, staging, production) using environment-specific information.
- **Stack Job Settings**: Integrate with infrastructure-as-code (IaC) tools through stack-update settings for automated and repeatable deployments.

### Credential Management

- **Kubernetes Credential ID**: Specify credentials required to access and configure the target Kubernetes cluster securely.

### MongoDB Container Configuration

- **Replicas**: Define the number of MongoDB pod instances. Recommended default is `1`.
- **Resources**: Allocate CPU and memory resources for the MongoDB container to optimize performance.
- **Persistence**:
- **Enable Persistence**: Toggle data persistence for MongoDB using `persistenceEnabled`. When enabled, data is stored in a persistent volume, allowing data to survive pod restarts.
- **Disk Size**: Specify the size of the persistent volume attached to each MongoDB pod (e.g., `1Gi`). This is mandatory if persistence is enabled.

### Helm Chart Customization

- **Helm Values**: Provide a map of key-value pairs for additional customization options via the MongoDB Helm chart. This allows for:
- Customizing resource limits
- Setting environment variables
- Specifying version tags
- For detailed options, refer to the [MongoDB Helm Chart values.yaml](https://artifacthub.io/packages/helm/bitnami/mongodb)

### Networking and Ingress

- **Ingress Configuration**: Set up external access to MongoDB by enabling ingress and specifying a custom hostname. When enabled, creates a LoadBalancer service with external-dns annotations for automatic DNS configuration.

## Benefits

- **Consistency Across Deployments**: Using a standardized API resource ensures deployments are predictable and maintainable.
- **Reduced Complexity**: Simplifies the deployment process by abstracting complex Kubernetes and Helm configurations.
- **Scalability and Flexibility**: Easily adjust replicas and resources to handle varying workloads and performance requirements.
- **Data Persistence**: Optionally enable data persistence to ensure data durability across pod restarts and failures.
- **Customization**: Enables detailed customization through Helm values to fit specific use cases.

## Use Cases

- **Database Services for Applications**: Deploy MongoDB as the database backend for applications running on Kubernetes.
- **Microservices Architecture**: Use MongoDB in a microservices environment where each service may require its own database instance.
- **Data Persistence and Backups**: Ensure data durability and facilitate backup strategies by enabling persistence.
- **Development and Testing Environments**: Quickly spin up MongoDB instances for development or testing purposes with environment-specific configurations.
