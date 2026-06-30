# Overview

The **Prometheus Kubernetes API resource** is designed to manage and deploy Prometheus monitoring systems in Kubernetes environments. This resource simplifies the deployment, configuration, and scaling of Prometheus containers, ensuring that your Kubernetes-based monitoring infrastructure is efficient, scalable, and reliable.

## Why We Created This API Resource

Managing a Prometheus deployment within Kubernetes can be complex, especially when dealing with resource allocation, persistence, and ingress configurations. This API resource provides a consistent interface to:

- **Simplify Prometheus Deployment**: By offering a streamlined way to configure Prometheus instances in Kubernetes, users can deploy Prometheus with ease.
- **Ensure Resource Optimization**: Enables users to define CPU, memory, and persistence options for Prometheus to ensure optimal performance.
- **Enhance Data Persistence**: Provides built-in options for persistence, allowing Prometheus to store data reliably between restarts.
- **Ingress Management**: Facilitates easy setup of ingress rules to securely expose Prometheus services.

## Key Features

### Environment and Stack Integration

- **Environment Info**: Automatically integrates with Planton's environment management, ensuring that Prometheus is deployed within the appropriate context.
- **Stack Job Settings**: Supports configuration of stack-update settings to enable consistent infrastructure-as-code deployments.

### Kubernetes Credential Management

- **Kubernetes Credential ID**: The `kubernetes_credential_id` is required to securely manage the Kubernetes provider within the stack-update, ensuring a secure and authenticated deployment process.

### Namespace Management

The `create_namespace` field controls whether the module creates the Kubernetes namespace or expects it to already exist:

- **`create_namespace: true`** (recommended for new deployments): The module creates the namespace with appropriate labels and metadata. Use this when deploying Prometheus to a new, dedicated namespace.
- **`create_namespace: false`**: The module uses an existing namespace without creating it. Use this when:
  - The namespace already exists and is managed separately
  - Another team or process manages namespace creation
  - You're deploying multiple components to a shared namespace
  - You have specific namespace configuration requirements managed outside this component

**Example with namespace creation:**
```yaml
spec:
  target_cluster:
    cluster_name: "my-cluster"
  namespace:
    value: "prometheus"
  create_namespace: true  # Module creates the namespace
  container:
    replicas: 1
```

**Example with existing namespace:**
```yaml
spec:
  target_cluster:
    cluster_name: "my-cluster"
  namespace:
    value: "monitoring"  # Must already exist
  create_namespace: false  # Module uses existing namespace
  container:
    replicas: 1
```

### Prometheus Container Configuration

#### Resource Management

- **Replicas**: Configure the number of Prometheus replicas to ensure redundancy and high availability. The default recommended value is 1 replica.

- **Container Resources**: Fine-tune CPU and memory resources to optimize the performance of your Prometheus deployment. The recommended default values are:
    - **CPU Requests**: `50m`
    - **Memory Requests**: `256Mi`
    - **CPU Limits**: `1`
    - **Memory Limits**: `1Gi`

#### Persistence Options

- **Persistence Toggle**: Choose whether to enable persistence for Prometheus data. When enabled, data is stored in a persistent volume, ensuring that monitoring data is available even after a pod restart.

- **Disk Size**: Define the size of the persistent storage used by each Prometheus pod. If persistence is enabled, this field is mandatory. The disk size can be adjusted based on your monitoring data requirements, though the size cannot be modified after the StatefulSet is created.

### Ingress Configuration

- **Ingress Spec**: Configure ingress rules for Prometheus to expose the service securely to external clients or internal systems within the Kubernetes environment.

## Benefits

- **Ease of Deployment**: Simplifies the complex process of deploying Prometheus on Kubernetes, allowing developers and DevOps teams to quickly set up and manage monitoring.
- **Scalability**: The ability to define replicas and autoscale resources ensures that Prometheus can handle varying workloads and scale as needed.
- **Data Persistence**: Ensures that critical monitoring data is stored persistently and is accessible between pod restarts, reducing data loss risks.
- **Secure Access**: Enables easy setup of secure ingress rules to expose Prometheus to authorized users or systems.
