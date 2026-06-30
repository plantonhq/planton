# Overview

The **Grafana Kubernetes API resource** provides a consistent and streamlined interface for deploying and managing Grafana instances within Kubernetes environments as part of our cloud infrastructure. By abstracting the complexities of Grafana and Kubernetes configurations, this resource allows you to set up powerful visualization and monitoring tools effortlessly while ensuring consistency and compliance across different environments.

## Why We Created This API Resource

Monitoring and visualizing system metrics are crucial for maintaining the health and performance of applications. Deploying Grafana on Kubernetes can be complex due to various configuration options, resource management, and networking considerations. To simplify this process and promote a standardized approach, we developed this API resource. It enables you to:

- **Simplify Deployment**: Easily configure and deploy Grafana instances without dealing with low-level Kubernetes configurations.
- **Ensure Consistency**: Maintain uniform Grafana deployments across different environments and teams.
- **Enhance Productivity**: Reduce the time and effort required to set up monitoring tools, allowing your team to focus on application development.
- **Optimize Resource Utilization**: Efficiently manage resources to ensure optimal performance and cost-effectiveness.

## Key Features

### Environment Integration

- **Environment Info**: Integrates seamlessly with our environment management system to deploy Grafana within specific environments.
- **Stack Job Settings**: Supports custom stack-update settings for infrastructure-as-code deployments, ensuring consistent and repeatable provisioning processes.

### Kubernetes Credential Management

- **Kubernetes Credential ID**: Utilizes specified Kubernetes credentials to ensure secure and authorized operations within Kubernetes clusters.

### Customizable Grafana Deployment

#### Grafana Container Configuration

- **Container Resources**: Specify CPU and memory resources for the Grafana container to optimize performance according to your needs. Recommended defaults are:
    - **CPU Requests**: `50m`
    - **Memory Requests**: `256Mi`
    - **CPU Limits**: `1`
    - **Memory Limits**: `1Gi`
- **Resource Optimization**: Adjust resource allocations to match the demands of your monitoring workloads, ensuring efficient use of cluster resources.

#### Ingress Configuration

- **Ingress Spec**: Configure ingress settings to expose the Grafana service outside the cluster, including:
    - **Hostname**: Define the external URL through which Grafana will be accessible.
    - **TLS Settings**: Enable TLS to secure connections to Grafana.
    - **Ingress Annotations**: Customize ingress controller behavior with annotations (e.g., for specific ingress controllers like NGINX or Istio).

### High Availability and Scalability

- **Replicas**: While not explicitly specified, you can configure the number of replicas for the Grafana deployment to ensure high availability and handle increased load.

### Integration with Monitoring Tools

- **Data Sources**: Integrate Grafana with various data sources (e.g., Prometheus, Elasticsearch) to visualize metrics and logs from your applications and infrastructure.
- **Dashboards**: Deploy predefined or custom dashboards to monitor system health, performance metrics, and application-specific data.

## Benefits

- **Simplified Deployment**: Abstracts the complexities of deploying Grafana on Kubernetes into an easy-to-use API resource.
- **Consistency**: Ensures all Grafana deployments adhere to organizational standards for security, performance, and scalability.
- **Scalability**: Allows for easy scaling of Grafana services to handle varying workloads and user demands.
- **Resource Optimization**: Enables precise control over resource allocation for containers, optimizing performance and cost.
- **Flexibility**: Customize Grafana configurations to meet specific monitoring requirements without compromising best practices.
- **Enhanced Observability**: Provides a centralized platform for visualizing metrics and logs, improving system observability and facilitating proactive issue detection.

## Namespace Management

The Grafana Kubernetes component provides flexible namespace management through the `create_namespace` flag in the spec:

### Namespace Creation Control

- **`create_namespace: true`**: The component will create a dedicated namespace for Grafana with appropriate resource labels for tracking and organization. All Grafana resources (helm chart, ingress, etc.) will be deployed into this created namespace. The namespace will be managed as part of the component's lifecycle.

- **`create_namespace: false`** (default): The component will deploy Grafana resources into an existing namespace specified in the `namespace` field. The namespace must already exist before deployment. This is useful when you want to deploy multiple components into a shared namespace or when namespace creation is managed separately.

### When to Use Each Option

**Use `create_namespace: true` when:**
- You want the component to fully manage its namespace lifecycle
- You're deploying Grafana in isolation
- You want automatic cleanup when the component is destroyed

**Use `create_namespace: false` when:**
- The namespace already exists and is managed separately
- You're deploying multiple components into a shared namespace
- You have existing namespace policies or quotas that should not be modified
- You want to preserve the namespace when the component is destroyed
