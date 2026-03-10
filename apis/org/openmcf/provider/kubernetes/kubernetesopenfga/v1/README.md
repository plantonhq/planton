# Overview

The **OpenFGA Kubernetes API resource** is designed to provide a streamlined interface for deploying OpenFGA, an open-source authorization system, in a Kubernetes environment. This API resource handles the configuration of the OpenFGA container, its resource allocation, ingress settings, and data store connection, ensuring a consistent and secure deployment across different environments.

## Why We Created This API Resource

Deploying OpenFGA in Kubernetes can involve complex configurations, especially when managing resources, ingress rules, and data store connections. This API resource simplifies the process by:

- **Simplifying Deployment**: Abstracts the complexity of Kubernetes configurations for OpenFGA.
- **Ensuring Consistency**: Provides a standardized way to deploy OpenFGA across multiple environments.
- **Optimizing Resource Management**: Allows for fine-tuning resource allocation (CPU and memory) for the OpenFGA container.
- **Enabling Flexible Data Store Options**: Supports both MySQL and PostgreSQL for OpenFGA data storage.

## Key Features

### Environment Integration

- **Environment Info**: Integrates with Planton's environment management system to deploy OpenFGA in a specific environment.
- **Stack Job Settings**: Supports stack-update settings to ensure consistent and repeatable deployments using infrastructure-as-code.

### Kubernetes Credential Management

- **Kubernetes Credential ID**: Specifies the Kubernetes credentials (`kubernetes_credential_id`) for securely deploying and managing the OpenFGA container within a Kubernetes cluster.

### OpenFGA Container Configuration

#### Resource Management

- **Container Resources**: Allows configuration of CPU and memory limits to optimize the performance of the OpenFGA container. Recommended default values:
    - **CPU Requests**: `50m`
    - **Memory Requests**: `256Mi`
    - **CPU Limits**: `1`
    - **Memory Limits**: `1Gi`

- **Replicas**: Define the number of replicas for the OpenFGA container, providing high availability and fault tolerance.

### Ingress Configuration

- **Ingress Spec**: Configures ingress rules to expose the OpenFGA service to external clients. Users specify the full hostname (e.g., `openfga.example.com`) for secure access to the OpenFGA API.

### Data Store Configuration

- **Engine**: Specifies the type of data store engine to use for OpenFGA, with support for both MySQL and PostgreSQL.
    - Supported engines: `"postgres"` and `"mysql"`

- **URI**: Defines the connection URI for the selected data store engine. The URI should be appropriately formatted based on the engine type (e.g., `mysql://user:password@host:port/database` or `postgres://user:password@host:port/database`).

### Namespace Management

The OpenFGA Kubernetes API resource provides flexible namespace management through the `create_namespace` flag:

- **create_namespace**: Boolean flag controlling namespace creation behavior
    - `true`: The module will create the specified namespace (recommended for most deployments)
    - `false`: The module expects the namespace to already exist and will use it

**When to use create_namespace: true**
- Fresh deployments where namespace doesn't exist
- Isolated deployments where you want the module to manage the full lifecycle
- Development/test environments
- Single-component namespaces

**When to use create_namespace: false**
- Namespace is managed by a separate KubernetesNamespace resource
- Namespace is pre-created by cluster administrators
- Multiple components sharing the same namespace
- Production environments with strict namespace management policies
- Organizations with centralized namespace governance

## Benefits

- **Simplified Deployment**: Reduces the complexities of deploying OpenFGA on Kubernetes with a standardized and easy-to-use API resource.
- **Consistency**: Ensures consistent deployment of OpenFGA across different environments and clusters.
- **Resource Optimization**: Allows for fine-tuning resource allocations to optimize performance and cost-efficiency.
- **Data Store Flexibility**: Supports both MySQL and PostgreSQL as data store engines, providing flexibility in database management.
- **Secure Access**: Configures ingress rules for secure and controlled access to the OpenFGA instance.
- **Flexible Namespace Management**: Control whether namespaces are created or managed externally, supporting various organizational policies.
