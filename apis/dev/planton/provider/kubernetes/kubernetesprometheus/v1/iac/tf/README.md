# Prometheus Kubernetes Terraform Module

## Key Features

### 1. **Namespace Management**

The module provides flexible namespace management through the `create_namespace` configuration option:

- **`create_namespace = true`** (recommended for new deployments): The module creates the namespace with appropriate labels and metadata. Use this when deploying Prometheus to a new, dedicated namespace.
- **`create_namespace = false`**: The module uses an existing namespace without creating it. Use this when:
  - The namespace already exists and is managed separately
  - Another team or process manages namespace creation
  - You're deploying multiple components to a shared namespace
  - You have specific namespace configuration requirements managed outside this module

**Example with namespace creation:**
```hcl
module "prometheus" {
  source = "..."
  
  spec = {
    namespace = "prometheus"
    create_namespace = true  # Module creates the namespace
    container = { ... }
  }
}
```

**Example with existing namespace:**
```hcl
module "prometheus" {
  source = "..."
  
  spec = {
    namespace = "monitoring"  # Must already exist
    create_namespace = false  # Module uses existing namespace
    container = { ... }
  }
}
```

### 2. **Kubernetes Provider Integration**
   The module creates a Kubernetes namespace foundation using the credentials configured in your Terraform environment. This integration automates the setup of necessary Kubernetes resources for Prometheus deployment, establishing a secure and organized environment for your monitoring stack.

### 3. **Prometheus Pod Deployment Foundation**
   The module provisions the namespace and configuration foundation for Prometheus pods based on the number of replicas defined in the `spec`. Developers can configure CPU and memory requests and limits to control the resource usage of each Prometheus pod. This ensures that the deployment can scale based on operational requirements while maintaining efficient resource consumption.

### 4. **Persistent Storage Options**
   The module supports configuration for enabling persistence for Prometheus data. When persistence is enabled in the spec, Prometheus' in-memory data can be backed up to a persistent volume, ensuring that data is retained across pod restarts. The size of the persistent volume can be customized through the `spec`, and the backed-up data will be restored when Prometheus pods are restarted.

### 5. **Ingress Configuration**
   The module offers optional ingress configuration, which allows external access to Prometheus from outside the Kubernetes cluster. If ingress is enabled, developers can define the DNS domain to expose Prometheus securely to external clients. This is especially useful for teams looking to integrate Prometheus monitoring across distributed environments.

### 6. **Terraform Infrastructure-as-Code**
   This module uses Terraform, providing developers with all the advantages of infrastructure-as-code, such as version control, state management, and declarative configuration. This makes it easy to manage Prometheus deployments in a consistent and repeatable manner across multiple environments and cloud platforms.

### 7. **Modular and Scalable Design**
   The module is designed to be modular, allowing developers to easily customize and scale their Prometheus deployments. Whether it's defining the number of replicas or tuning resource allocations, the module supports various configurations, making it suitable for a wide range of use cases—from small-scale monitoring setups to large, distributed environments.

### 8. **Terraform Outputs**
   Upon successful deployment, the module provides several outputs that are essential for managing and accessing the Prometheus instance:
   - **Namespace**: The namespace in which Prometheus is deployed, ensuring logical separation from other workloads.
   - **Service Name**: The Kubernetes service associated with Prometheus for internal access.
   - **Port Forward Command**: A command to set up port forwarding for local access when ingress is disabled.
   - **Kubernetes Endpoint**: An internal endpoint for accessing Prometheus within the cluster.
   - **External Hostname**: A public URL for accessing Prometheus from outside the cluster if ingress is enabled.
   - **Internal Hostname**: A private URL for internal access within the Kubernetes environment.

## Prerequisites

- Terraform >= 1.0
- Kubernetes cluster (any provider: AWS EKS, GCP GKE, Azure AKS, etc.)
- Kubernetes provider configured in your Terraform environment
- kube-prometheus-stack Helm chart (for actual Prometheus deployment)

## Usage

This Terraform module creates the namespace foundation for Prometheus deployment. The actual Prometheus server deployment is expected to be managed via the **kube-prometheus-stack Helm chart** or the Prometheus Operator.

Refer to the examples section for usage instructions.

## Benefits

1. **Standardized API Resource Modeling**: The module uses Kubernetes-style API resource modeling, ensuring a consistent and familiar approach for developers working in Kubernetes environments. This allows for standardized deployment and configuration patterns across different teams and projects.
2. **Cloud-Agnostic**: The module works seamlessly across any Kubernetes cluster, whether hosted on AWS, GCP, Azure, or any other provider. This ensures flexibility in deployment, making it a robust solution for multi-cloud environments.
3. **Customizable Resource Allocation**: Developers can fine-tune the resource usage of Prometheus pods by adjusting CPU and memory settings through the spec variable, ensuring efficient use of cluster resources based on specific performance requirements.
4. **Persistent Data Storage Support**: With optional persistence configuration, the module supports Prometheus data retention even after pod restarts, enhancing reliability for long-term monitoring.
5. **Infrastructure-as-Code with Terraform**: Terraform provides all the benefits of infrastructure-as-code, such as managing Prometheus configurations in version control, state tracking, and consistent deployments across environments.
6. **Declarative Configuration**: Using Terraform's declarative syntax makes it easy to understand and maintain Prometheus infrastructure configurations.

## Architecture

The module creates:
- **Kubernetes Namespace**: Dedicated namespace for Prometheus resources with proper labels
- **Output Values**: Connection strings, endpoints, and commands for accessing Prometheus

The module is designed to work with the **kube-prometheus-stack** Helm chart, which provides:
- Prometheus server with operator-managed configuration
- Grafana for visualization
- Alertmanager for alert routing
- ServiceMonitors for auto-discovery of scrape targets
- PrometheusRules for recording and alerting rules
- Node exporters and kube-state-metrics for comprehensive monitoring

## Deployment Workflow

1. Apply this Terraform module to create the namespace foundation
2. Deploy the kube-prometheus-stack Helm chart into the created namespace
3. Configure ServiceMonitors and PrometheusRules as needed
4. Access Prometheus via the endpoints provided in the module outputs

## Conclusion

The Prometheus Kubernetes Terraform module offers a standardized, cloud-agnostic way to prepare your Kubernetes environment for Prometheus deployment. By automating the namespace provisioning and configuration management, the module reduces operational complexity and ensures that Prometheus instances can be deployed consistently across different environments. Leveraging Terraform's infrastructure-as-code capabilities, this module makes managing Prometheus infrastructure intuitive, reliable, and repeatable across multi-cloud environments.

