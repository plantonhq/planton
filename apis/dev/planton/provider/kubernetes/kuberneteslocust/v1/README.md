# Overview

The **Locust Kubernetes API resource** provides a consistent and streamlined interface for deploying and managing Locust-based load testing clusters within Kubernetes environments. This resource is designed to automate the configuration and deployment of Locust to run distributed load tests, allowing you to simulate traffic against target applications while ensuring scalability and flexibility.

## Why We Created This API Resource

Running distributed load tests using Locust in Kubernetes environments can be complex, requiring detailed configurations for containers, resource management, and test scripts. To simplify this process and provide a standardized approach, we developed this API resource. It allows you to:

- **Simplify Deployment**: Easily deploy a Locust load testing cluster without needing to configure the underlying Kubernetes resources manually.
- **Ensure Consistency**: Maintain consistent load testing setups across environments and clusters.
- **Enhance Flexibility**: Customize the behavior of load tests through Python test scripts and additional libraries.
- **Optimize Resource Management**: Efficiently allocate resources and scale Locust master and worker nodes as needed.

## Key Features

### Environment Integration

- **Environment Info**: Integrates seamlessly with our environment management system to deploy Locust within specific environments.
- **Stack Job Settings**: Supports custom stack-update settings for infrastructure-as-code deployments, ensuring consistent and repeatable provisioning processes.

### Kubernetes Credential Management

- **Kubernetes Credential ID**: Utilizes the specified Kubernetes credentials (`kubernetes_credential_id`) to ensure secure and authorized operations within Kubernetes clusters.

### Namespace Management

- **Namespace Configuration**: Specify the Kubernetes namespace where Locust will be deployed using the `namespace` field.
- **Namespace Creation Control**: Use the `create_namespace` flag to control whether the module should create the namespace or use an existing one:
  - **`create_namespace: true`**: The module creates the namespace with appropriate labels. Use this for new deployments or when you want the module to fully manage the namespace lifecycle.
  - **`create_namespace: false`**: The module uses an existing namespace without creating it. Use this when:
    - The namespace already exists in the cluster
    - Multiple deployments share the same namespace
    - Namespaces are managed centrally by cluster administrators
    - Using GitOps workflows where namespaces are managed separately

  **Important**: When `create_namespace: false`, you must ensure the namespace exists before deploying, otherwise the deployment will fail.

### Customizable Locust Deployment

#### Locust Master and Worker Configuration

- **Master and Worker Containers**: Configure CPU and memory resources for both the master and worker containers to optimize performance. Recommended defaults are:
    - **CPU Requests**: `50m`
    - **Memory Requests**: `256Mi`
    - **CPU Limits**: `1`
    - **Memory Limits**: `1Gi`
- **Replicas**: Define the number of replicas for both the master and worker containers to scale the load testing cluster based on your testing needs.

#### Ingress Configuration

- **Ingress Spec**: Configure ingress settings to expose the Locust web UI or API endpoints outside the cluster, allowing for external access and control over load tests.

### Load Test Configuration

#### Load Test Script

- **Main Python Script**: Provide the `main_py_content` which contains the Python code defining the behavior of the simulated users for the load test.
- **Library Files**: Add supporting Python files via `lib_files_content`, allowing the main script to reference additional classes or functions necessary for test execution.
- **Pip Packages**: Specify extra Python pip packages required for the test using `pip_packages`. This allows for flexible test execution by including additional dependencies.

### Helm Chart Customization

- **Helm Values**: Provide a map of key-value pairs (`helm_values`) to customize the Helm chart used to deploy the Locust cluster. This includes options for fine-tuning resource limits, environment variables, or version tags. For detailed options, refer to the [Helm chart documentation](https://github.com/deliveryhero/helm-charts/tree/master/stable/locust#values).

## Benefits

- **Simplified Deployment**: Abstracts the complexities of setting up a Locust load testing cluster on Kubernetes into an easy-to-use API resource.
- **Consistency**: Ensures all Locust deployments adhere to organizational standards for load testing environments.
- **Scalability**: Allows for easy scaling of Locust master and worker nodes to handle varying loads and testing requirements.
- **Customizability**: Enables custom load test scripts, library files, and Python package dependencies to suit specific test cases.
- **Resource Optimization**: Provides precise control over resource allocation for containers, optimizing performance and cost.
- **Flexibility**: Supports advanced ingress configurations and external access to the Locust web interface.
