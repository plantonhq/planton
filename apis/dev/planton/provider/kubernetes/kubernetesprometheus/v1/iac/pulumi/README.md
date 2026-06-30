# Prometheus Kubernetes Pulumi Module

## Key Features

### 1. **Kubernetes Provider Integration**
   The module creates a Kubernetes provider using the credentials specified in the `api-resource.yaml`, ensuring secure access to the target Kubernetes cluster. This integration automates the setup of all necessary Kubernetes resources for Prometheus, from namespace creation to service deployment.

### 2. **Prometheus Pod Deployment**
   The module provisions Prometheus pods based on the number of replicas defined in the `spec`. Developers can configure CPU and memory requests and limits to control the resource usage of each Prometheus pod. This ensures that the deployment can scale based on operational requirements while maintaining efficient resource consumption.

### 3. **Persistent Storage Options**
   The module supports enabling persistence for Prometheus data. When persistence is enabled, Prometheus' in-memory data is backed up to a persistent volume, ensuring that data is retained across pod restarts. The size of the persistent volume can be customized through the `spec`, and the backed-up data will be restored when Prometheus pods are restarted, ensuring the continuity of the monitoring system.

### 4. **Ingress Configuration**
   The module offers optional ingress configuration, which allows external access to Prometheus from outside the Kubernetes cluster. If ingress is enabled, developers can define the ingress class and host to expose Prometheus securely to external clients. This is especially useful for teams looking to integrate Prometheus monitoring across distributed environments.

### 5. **Pulumi Infrastructure-as-Code**
   This module uses Pulumi, providing developers with all the advantages of infrastructure-as-code, such as version control, automated rollbacks, and state management. This makes it easy to manage Prometheus deployments in a consistent and repeatable manner across multiple environments and cloud platforms.

### 6. **Modular and Scalable Design**
   The module is designed to be modular, allowing developers to easily customize and scale their Prometheus deployments. Whether it’s defining the number of replicas or tuning resource allocations, the module supports various configurations, making it suitable for a wide range of use cases—from small-scale monitoring setups to large, distributed environments.

### 7. **Pulumi Stack Outputs**
   Upon successful deployment, the module provides several outputs that are essential for managing and accessing the Prometheus instance:
   - **Namespace**: The namespace in which Prometheus is deployed, ensuring logical separation from other workloads.
   - **Service Name**: The Kubernetes service associated with Prometheus for internal access.
   - **Port Forward Command**: A command to set up port forwarding for local access when ingress is disabled.
   - **Kubernetes Endpoint**: An internal endpoint for accessing Prometheus within the cluster.
   - **External Hostname**: A public URL for accessing Prometheus from outside the cluster if ingress is enabled.
   - **Internal Hostname**: A private URL for internal access within the Kubernetes environment.

## Usage

Refer to the example section for usage instructions.

## Benefits

1. **Standardized API Resource Modeling**: The module uses Kubernetes-style API resource modeling, ensuring a consistent and familiar approach for developers working in Kubernetes environments. This allows for standardized deployment and configuration patterns across different teams and projects.
2. **Cloud-Agnostic**: The module works seamlessly across any Kubernetes cluster, whether hosted on AWS, GCP, Azure, or any other provider. This ensures flexibility in deployment, making it a robust solution for multi-cloud environments.
3. **Customizable Resource Allocation**: Developers can fine-tune the resource usage of Prometheus pods by adjusting CPU and memory settings, ensuring efficient use of cluster resources based on specific performance requirements.
4. **Persistent Data Storage**: With optional persistence enabled, the module ensures that Prometheus data is retained even after pod restarts, enhancing reliability for long-term monitoring.
5. **Infrastructure-as-Code with Pulumi**: Pulumi provides all the benefits of infrastructure-as-code, such as managing Prometheus configurations in version control, automated rollbacks, and consistent deployments across environments.

## Conclusion

The Prometheus Kubernetes Pulumi module offers a powerful, flexible, and standardized way to deploy Prometheus in any Kubernetes environment. By automating the provisioning of Prometheus resources, including pods, services, and ingress, the module reduces operational complexity and ensures that Prometheus instances can scale and adapt to varying workloads. Leveraging Pulumi’s infrastructure-as-code capabilities, this module makes managing complex Prometheus deployments intuitive, reliable, and repeatable across multi-cloud environments.