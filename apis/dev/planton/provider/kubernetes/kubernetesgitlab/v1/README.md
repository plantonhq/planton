# Overview

The GitLab Kubernetes API resource provides a consistent and streamlined interface for deploying and managing GitLab instances within Kubernetes environments as part of our cloud infrastructure. By abstracting the complexities of GitLab and Kubernetes configurations, this resource allows you to set up a robust DevOps platform effortlessly while ensuring consistency and compliance across different environments.

## Why We Created This API Resource

Deploying GitLab on Kubernetes can be complex due to the intricacies involved in configuring the application, managing resources, and ensuring secure access. To simplify this process and promote a standardized approach, we developed this API resource. It enables you to:

- **Simplify Deployment**: Easily configure and deploy GitLab instances without dealing with low-level Kubernetes and GitLab configurations.
- **Ensure Consistency**: Maintain uniform GitLab deployments across different environments and teams.
- **Enhance Productivity**: Reduce the time and effort required to set up GitLab, allowing your team to focus on development and collaboration.
- **Optimize Resource Utilization**: Efficiently manage resources to ensure optimal performance and cost-effectiveness.

## Key Features

### Cluster Targeting and Namespace Management

- **Target Cluster Selection**: Specify the Kubernetes cluster where GitLab should be deployed using the `target_cluster` field
- **Namespace Configuration**: Define the namespace for GitLab resources using the `namespace` field
- **Flexible Namespace Creation**: Control namespace creation with the `create_namespace` flag:
  - `true`: The module automatically creates the namespace with appropriate labels
  - `false`: Use an existing namespace (must be created beforehand)

### Environment Integration

- **Environment Info**: Seamlessly integrates with our environment management system to deploy GitLab within specific environments.
- **Stack Job Settings**: Supports custom stack-update settings for infrastructure-as-code deployments.

### Customizable GitLab Deployment

#### GitLab Container

- **Container Resources**: Specify CPU and memory resources for the GitLab container to optimize performance. Recommended defaults are:
    - CPU Requests: 50m
    - Memory Requests: 256Mi
    - CPU Limits: 1
    - Memory Limits: 1Gi

### Ingress Configuration

- **Ingress Spec**: Configure ingress settings to expose the GitLab service outside the cluster, including hostname, TLS settings, and ingress annotations.

## Namespace Management

The GitLab component provides flexible namespace management to suit different deployment scenarios:

### Automatic Namespace Creation (create_namespace: true)

When `create_namespace` is set to `true`, the module:
- Creates a dedicated namespace for GitLab resources
- Applies resource labels for tracking and organization
- Manages namespace lifecycle as part of the deployment

**Example:**
```yaml
spec:
  target_cluster:
    cluster_name: my-gke-cluster
  namespace:
    value: gitlab-prod
  create_namespace: true
```

### Using Existing Namespace (create_namespace: false)

When `create_namespace` is set to `false`:
- The namespace must exist before deployment
- GitLab resources will be created in the specified namespace
- Useful for shared namespaces or when namespace is managed externally

**Example:**
```yaml
spec:
  target_cluster:
    cluster_name: my-gke-cluster
  namespace:
    value: shared-services
  create_namespace: false
```

**Note:** When using an existing namespace, ensure it exists before deploying GitLab, otherwise the deployment will fail.

## Benefits

- **Simplified Deployment**: Abstracts the complexities of deploying GitLab on Kubernetes into an easy-to-use API.
- **Consistency**: Ensures all GitLab deployments adhere to organizational standards for security, performance, and scalability.
- **Scalability**: Allows for easy scaling of GitLab services to handle varying workloads.
- **Resource Optimization**: Enables precise control over resource allocation for containers, optimizing performance and cost.
- **Flexibility**: Customize GitLab configurations to meet specific project requirements without compromising best practices.
- **Enhanced Collaboration**: Provides a centralized platform for code hosting, CI/CD, and project management, improving team collaboration and productivity.
