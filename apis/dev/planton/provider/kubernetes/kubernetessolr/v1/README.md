# Overview

The **Solr Kubernetes API resource** provides a structured way to deploy and manage Solr clusters in Kubernetes environments. It includes configurations for Solr, Zookeeper, and ingress management, allowing for a comprehensive setup of a Solr cluster that is optimized for scalability and performance.

## Purpose of the Solr Kubernetes API Resource

Deploying Solr in Kubernetes often involves various components, such as managing Solr configurations, Zookeeper instances, resource allocations, and persistence settings. This API resource simplifies that process by offering a well-defined structure for Solr deployment in Kubernetes, ensuring high availability, efficient resource management, and easy scaling.

## Key Features

### Solr Container Configuration

- **Replicas**: Configure the number of Solr pod replicas, with a recommended default of 1 for initial deployments.
- **Container Image**: Define the Solr container image, such as `solr:8.7.0`, for deployment.
- **Resource Allocation**: Solr container resources can be customized. The recommended default values are:
    - **CPU Requests**: `50m`
    - **Memory Requests**: `256Mi`
    - **CPU Limits**: `1`
    - **Memory Limits**: `1Gi`
- **Disk Size**: Allocate disk storage for persistent data. The default is `1Gi`, ensuring persistent data backup in case of restarts.

### Solr Configuration

- **JVM Memory Settings**: Set JVM memory configurations for Solr. The default is `"-Xms1g -Xmx3g"`.
- **Custom Solr Options**: Provide additional Solr options, such as `-Dsolr.autoSoftCommit.maxTime=10000`, to tune Solr performance.
- **Garbage Collection Tuning**: Customize the garbage collection settings for Solr, such as `-XX:SurvivorRatio=4 -XX:TargetSurvivorRatio=90`.

### Zookeeper Container Configuration

- **Replicas**: Configure the number of Zookeeper pod replicas, with a recommended default of 1.
- **Resource Allocation**: Customize Zookeeper's container resources. The recommended default values are:
    - **CPU Requests**: `50m`
    - **Memory Requests**: `256Mi`
    - **CPU Limits**: `1`
    - **Memory Limits**: `1Gi`
- **Disk Size**: Allocate disk storage for Zookeeper with a default value of `1Gi`.

### Ingress Configuration

- **Ingress Spec**: Use Kubernetes ingress configurations to expose the Solr service securely, enabling external access as needed.

### Namespace Management

The Solr Kubernetes component provides flexible namespace management through the `create_namespace` flag:

- **create_namespace: true** (default) - The module creates the namespace with proper labels and manages its lifecycle. This is ideal for dedicated Solr deployments where the namespace is fully managed by this component.

- **create_namespace: false** - The module uses an existing namespace. You're responsible for ensuring the namespace exists before deployment. This is useful when:
  - Multiple components share a namespace
  - Namespace is managed by a separate infrastructure module
  - Organization policies require centralized namespace management
  - Using a GitOps approach with namespace management in a separate layer

When using an existing namespace (`create_namespace: false`), ensure the namespace exists before deploying, otherwise the deployment will fail.

## Benefits

- **Simplified Deployment**: This API resource abstracts the complexities of deploying and managing Solr in Kubernetes, offering a straightforward approach.
- **Scalable and Resilient**: Built-in configuration options for replicas, resource management, and persistence ensure a highly available and scalable Solr cluster.
- **Data Persistence**: Persistent storage options guarantee that Solr data is securely backed up, reducing the risk of data loss during restarts or failures.
- **Customizable**: Fine-tune resource allocations, JVM settings, and garbage collection configurations to match your performance requirements.
- **Integrated Zookeeper**: Manage Zookeeper instances alongside Solr with similar configuration options for ease of use.
