# Overview

The **Neo4j Kubernetes API resource** provides a consistent and streamlined interface for deploying and managing Neo4j database instances within Kubernetes environments. This resource simplifies the process of running Neo4j on Kubernetes, providing configurations for container resources, ingress settings, and Kubernetes credentials to ensure a smooth deployment.

## Why We Created This API Resource

Managing Neo4j in a Kubernetes environment can be challenging due to the complexity of setting up the database, configuring resources, and managing ingress traffic. We developed this API resource to:

- **Simplify Deployment**: Abstract the complexities of deploying Neo4j on Kubernetes, making it easier for users to set up and manage the database.
- **Ensure Consistency**: Standardize the deployment process across different environments and clusters, ensuring stability and performance.
- **Optimize Resource Management**: Provide the ability to fine-tune resource allocation (CPU and memory) for Neo4j containers.
- **Secure Access**: Allow easy ingress configuration for external access to Neo4j while maintaining control over security and access.

## Key Features

### Environment Integration

- **Environment Info**: Seamlessly integrates with our environment management system, deploying Neo4j within a specific environment.
- **Stack Job Settings**: Supports custom stack-update settings to ensure consistent and repeatable provisioning of Neo4j instances using infrastructure-as-code.

### Kubernetes Credential Management

- **Kubernetes Credential ID**: Specifies the Kubernetes credentials (`kubernetes_credential_id`) to be used for securely setting up and managing Neo4j within Kubernetes clusters.

### Neo4j Container Configuration

#### Resource Management

- **Container Resources**: Define CPU and memory limits to optimize the performance of the Neo4j container. Recommended default values are:
    - **CPU Requests**: `50m`
    - **Memory Requests**: `256Mi`
    - **CPU Limits**: `1`
    - **Memory Limits**: `1Gi`

This ensures that the Neo4j container can handle load efficiently while preventing overconsumption of resources.

### Ingress Configuration

- **Ingress Spec**: Configure ingress rules to expose the Neo4j service externally, allowing secure access to the database from outside the cluster. This includes options for handling public or private access based on organizational requirements.

### Namespace Management

- **Namespace Creation Control**: Use the `create_namespace` boolean flag to control whether the component creates the Kubernetes namespace or expects it to already exist.
  - Set to `true` to have the component create and manage the namespace
  - Set to `false` when the namespace is managed separately (e.g., by a KubernetesNamespace resource or external tooling)
- **Namespace Reference**: The `namespace` field uses `StringValueOrRef`, allowing you to specify the namespace name directly or reference a KubernetesNamespace resource

## Benefits

- **Simplified Deployment**: Reduces the complexities of setting up Neo4j on Kubernetes by providing a standardized and easy-to-use API resource.
- **Consistency**: Ensures all Neo4j deployments follow a consistent configuration across different environments.
- **Resource Optimization**: Allows for fine-tuning of container resource allocations, ensuring Neo4j performs optimally while avoiding overuse of cluster resources.
- **Secure Access**: Supports the configuration of ingress rules, enabling secure and controlled access to Neo4j instances deployed on Kubernetes.
