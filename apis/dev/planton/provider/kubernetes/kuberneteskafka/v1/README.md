# Overview

The **Kafka Kubernetes API Resource** provides a standardized and efficient way to deploy Apache Kafka onto Kubernetes clusters. This API resource simplifies the deployment process by encapsulating all necessary configurations, enabling consistent and repeatable Kafka deployments across various environments.

## Purpose

Deploying Kafka on Kubernetes involves intricate configurations, including brokers, topics, storage, and resource management. The Kafka Kubernetes API Resource aims to:

- **Standardize Deployments**: Offer a consistent interface for deploying Kafka, reducing complexity and minimizing errors.
- **Simplify Configuration Management**: Centralize all deployment settings, making it easier to manage, update, and replicate configurations.
- **Enhance Scalability and Flexibility**: Allow granular control over various components like brokers, topics, and resources to meet specific requirements.

## Key Features

### Environment Configuration

- **Environment Info**: Tailor Kafka deployments to specific environments (development, staging, production) using environment-specific information.
- **Stack Job Settings**: Integrate with infrastructure-as-code (IaC) tools through stack-update settings for automated and repeatable deployments.

### Credential Management

- **Kubernetes Credential ID**: Specify credentials required to access and configure the target Kubernetes cluster securely.

### Kafka Cluster Configuration

- **Kafka Brokers**:
- **Replicas**: Define the number of Kafka broker instances. Recommended default is `1`.
- **Resources**: Allocate CPU and memory resources for each broker to optimize performance.
- **Disk Size**: Specify the disk size for each broker instance (e.g., `1Gi`). Defaults to `1Gi` if not provided.

- **Zookeeper**:
- **Replicas**: Set the number of Zookeeper instances. A minimum of `3` is recommended for high availability.
- **Resources**: Allocate CPU and memory resources for each Zookeeper instance.
- **Disk Size**: Specify the disk size for each Zookeeper instance (e.g., `1Gi`). Defaults to `1Gi` if not provided.

- **Schema Registry** (Optional):
- **Enable Schema Registry**: Toggle the deployment of the Schema Registry component.
- **Replicas**: Define the number of Schema Registry instances. Recommended default is `1`.
- **Resources**: Allocate CPU and memory resources for the Schema Registry.

### Kafka Topics

- **Kafka Topics**: Define a list of Kafka topics to be created upon deployment. For each topic:
- **Name**: Set the topic name, adhering to Kafka naming conventions.
- **Partitions**: Specify the number of partitions. Recommended default is `1`.
- **Replicas**: Define the replication factor. Recommended default is `1`.
- **Config**: Provide additional configurations like `cleanup.policy`, `retention.ms`, etc.

### Networking and Ingress

- **Ingress Configuration**: Set up Kubernetes Ingress resources to manage external access to Kafka brokers and other components.

### Additional Features

- **Kafka UI Deployment**: Optionally deploy a Kafka UI component for easier management and monitoring.
- **Custom Configurations**: Utilize maps for advanced configurations, allowing for fine-tuning of Kafka and Zookeeper settings.

### Namespace Management

The component provides flexible namespace management through the `create_namespace` flag:

- **Automatic Creation** (`create_namespace: true`): The module creates the namespace with appropriate labels and manages its lifecycle. This is the recommended approach for most use cases.
- **External Management** (`create_namespace: false`): Use an existing namespace created separately (e.g., via KubernetesNamespace component or other tooling). The module will deploy resources into the specified namespace without creating it.

This flexibility allows integration with existing namespace management practices, centralized governance policies, and multi-tenant cluster configurations.

## Benefits

- **Consistency Across Deployments**: Using a standardized API resource ensures deployments are predictable and maintainable.
- **Reduced Complexity**: Simplifies the deployment process by abstracting complex Kubernetes and Kafka configurations.
- **Scalability**: Easily scale brokers, partitions, and replicas to handle varying workloads.
- **Customization**: Enables detailed customization to fit specific use cases and performance requirements.

## Use Cases

- **Data Streaming Applications**: Deploy Kafka clusters to support real-time data streaming and processing applications.
- **Microservices Communication**: Use Kafka as a messaging backbone for microservices architecture.
- **Event Sourcing and CQRS**: Implement event-driven systems requiring reliable message storage and retrieval.
- **Log Aggregation**: Collect and manage logs from various services in a centralized Kafka cluster.