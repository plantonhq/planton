# Overview

The **Keycloak Kubernetes API resource** provides a consistent and streamlined interface for deploying and managing Keycloak instances within Kubernetes environments as part of our cloud infrastructure. By abstracting the complexities of Keycloak and Kubernetes configurations, this resource enables you to set up powerful identity and access management (IAM) capabilities effortlessly, while ensuring consistency and compliance across different environments.

## Why We Created This API Resource

Deploying Keycloak on Kubernetes can be challenging due to the various configuration, resource management, and networking requirements. To simplify this process and promote a standardized approach, we developed this API resource. It enables you to:

- **Simplify Deployment**: Easily configure and deploy Keycloak instances without dealing with low-level Kubernetes and Keycloak configurations.
- **Ensure Consistency**: Maintain uniform IAM deployments across different environments.
- **Enhance Security**: Integrate a robust, centralized identity provider for authentication and authorization.
- **Optimize Resource Utilization**: Efficiently manage resources to ensure optimal performance and cost-effectiveness.

## Key Features

### Environment Integration

- **Environment Info**: Integrates seamlessly with our environment management system to deploy Keycloak within specific environments.
- **Stack Job Settings**: Supports custom stack-update settings for infrastructure-as-code deployments, ensuring consistent and repeatable provisioning processes.

### Kubernetes Credential Management

- **Kubernetes Credential ID**: Utilizes specified Kubernetes credentials (`kubernetes_credential_id`) to ensure secure and authorized operations within Kubernetes clusters.

### Customizable Keycloak Deployment

#### Keycloak Container Configuration

- **Container Resources**: Specify CPU and memory resources for the Keycloak container to optimize performance. Recommended defaults are:
    - **CPU Requests**: `50m`
    - **Memory Requests**: `256Mi`
    - **CPU Limits**: `1`
    - **Memory Limits**: `1Gi`
- **Resource Optimization**: Adjust resource allocations to match the demands of your identity management workloads, ensuring efficient use of cluster resources.

#### Ingress Configuration

- **Ingress Spec**: Configure ingress settings to expose the Keycloak service outside the cluster, including:
    - **Hostname**: Define the external URL through which Keycloak will be accessible.
    - **TLS Settings**: Enable TLS to secure connections to Keycloak.
    - **Ingress Annotations**: Customize ingress controller behavior with annotations (e.g., for specific ingress controllers like NGINX or Istio).

### Namespace Management

- **Namespace Specification**: Define the target Kubernetes namespace for the Keycloak deployment using the `namespace` field
- **Namespace Creation Control**: Use the `create_namespace` boolean flag to control namespace lifecycle:
  - `true`: The infrastructure module will create the namespace
  - `false`: The namespace must already exist and will only be referenced

This provides flexibility for different deployment scenarios:
- **Isolated deployments**: Create dedicated namespaces per Keycloak instance
- **Shared namespaces**: Deploy multiple components into pre-existing namespaces
- **External namespace management**: Use tools like Terraform or ArgoCD to manage namespace lifecycle separately

### High Availability and Scalability

- **Replicas**: Although not explicitly specified in the spec, you can configure the number of replicas for the Keycloak deployment to ensure high availability and handle increased load.
- **Persistence**: Configure persistent storage for Keycloak data to ensure configurations, user data, and session states are retained across pod restarts.

## Benefits

- **Simplified Deployment**: Abstracts the complexities of deploying Keycloak on Kubernetes into an easy-to-use API resource.
- **Consistency**: Ensures all Keycloak deployments adhere to organizational standards for security, performance, and scalability.
- **Scalability**: Allows for easy scaling of Keycloak services to handle varying workloads and identity management needs.
- **Resource Optimization**: Enables precise control over resource allocation for containers, optimizing performance and cost.
- **Enhanced Security**: Provides a centralized identity and access management solution for securing your applications and services.
- **Flexibility**: Customize Keycloak configurations to meet specific identity and access management requirements.
