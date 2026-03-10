# Overview

The **Redis Kubernetes API resource** is designed to manage and deploy Redis databases within Kubernetes environments. This resource provides a consistent interface for configuring Redis instances, including resource management, data persistence, and ingress setup, ensuring that your Redis deployment is both scalable and reliable.

## Why We Created This API Resource

Deploying Redis in a Kubernetes environment can be complex, especially when dealing with persistent storage, resource allocation, and ensuring high availability. The Redis Kubernetes API resource simplifies this process by offering a comprehensive set of configuration options tailored for Kubernetes environments. This resource allows teams to:

- **Efficiently Deploy Redis**: Simplify the process of deploying Redis on Kubernetes by abstracting away the complexities of Kubernetes configurations.
- **Ensure Data Persistence**: Offer built-in options for data persistence to ensure Redis in-memory data is securely backed up.
- **Optimize Resource Management**: Provide fine-grained control over resource allocation for optimal Redis performance.
- **Ingress Management**: Easily configure secure ingress routes to expose Redis services for internal or external access.

## Key Features

### Environment and Stack Integration

- **Environment Info**: This resource integrates seamlessly with Planton’s environment management system, ensuring that Redis instances are deployed in the appropriate environment.
- **Stack Job Settings**: Stack job settings ensure that Redis instances are deployed using consistent infrastructure-as-code approaches.

### Kubernetes Credential Management

- **Kubernetes Credential ID**: The `kubernetes_credential_id` is required to authenticate and securely manage the Kubernetes provider used during Redis deployment.

### Redis Container Configuration

#### Resource Management

- **Replicas**: Define the number of Redis replicas for high availability and redundancy. The recommended default value is 1.

- **Container Resources**: Configure CPU and memory allocation to ensure optimal Redis performance. The recommended default values are:
    - **CPU Requests**: `50m`
    - **Memory Requests**: `256Mi`
    - **CPU Limits**: `1`
    - **Memory Limits**: `1Gi`

#### Persistence Options

- **Persistence Toggle**: Enable or disable persistence for Redis data. When persistence is enabled, in-memory data is stored in a persistent volume, ensuring data continuity even after pod restarts.

- **Disk Size**: Define the size of the persistent storage attached to each Redis pod. If persistence is enabled, this field is mandatory, and the specified disk size is used to store Redis data across pod restarts.

### Namespace Management

- **Flexible Namespace Control**: The `create_namespace` flag provides fine-grained control over namespace creation:
  - **When `true` (Recommended for most use cases)**: The module creates a dedicated namespace with appropriate resource labels for tracking and organization. This is ideal for new deployments, testing, and isolated environments.
  - **When `false`**: The module uses an existing namespace without attempting to create it. The namespace must exist before deployment. This is suitable for environments with pre-configured namespaces that have existing ResourceQuotas, LimitRanges, NetworkPolicies, or RBAC policies.
- **Use Cases**:
  - Set to `true`: New deployments, development environments, testing scenarios
  - Set to `false`: Production environments with governed namespaces, multi-tenant clusters, compliance-driven deployments

### Ingress Configuration

- **Ingress Hostname Control**: Configure a LoadBalancer service with external-dns annotations to expose the Redis service. Users have full control over the ingress hostname, allowing custom DNS patterns like `redis.example.com` or `cache-prod.company.com`. When ingress is enabled, the hostname field is required and the system creates a LoadBalancer service with the specified hostname configured via external-dns.

## Benefits

- **Simplicity**: This resource streamlines Redis deployment, making it easier for DevOps teams to set up and manage Redis in Kubernetes.
- **High Availability**: By defining replicas and resource limits, Redis instances can be highly available and resilient.
- **Data Reliability**: With persistence enabled, Redis data is backed up to a persistent volume, ensuring that data is not lost during restarts or failures.
- **Secure Access**: Securely configure ingress rules to control access to Redis services.
