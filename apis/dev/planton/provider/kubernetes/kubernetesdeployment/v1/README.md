# Overview

The **Microservice Kubernetes** API resource provides a standardized and streamlined way to deploy microservices onto Kubernetes clusters. This deployment module is designed to simplify the complex process of configuring and managing microservice deployments by encapsulating all necessary specifications into a single, cohesive resource.

## Purpose

Deploying microservices to Kubernetes can be a complex task involving numerous configurations for containers, networking, scaling, and more. The Microservice Kubernetes API resource aims to:

- **Standardize Deployments**: Offer a consistent interface for deploying microservices, reducing the learning curve and potential for errors.
- **Simplify Configuration Management**: Consolidate all deployment-related settings into one place, making it easier to manage and update configurations.
- **Enhance Flexibility**: Provide granular control over various deployment aspects to cater to diverse application requirements.

## Key Features

### Environment Configuration

- **Environment Info**: Tailor deployments to specific environments (development, staging, production) using environment-specific information.
- **Stack Job Settings**: Integrate with infrastructure-as-code (IaC) tools through stack-update settings for automated and repeatable deployments.

### Credential Management

- **Kubernetes Credential ID**: Specify credentials required to access and configure the target Kubernetes cluster securely.
- **Docker Credential ID**: Provide credentials for pulling container images from private Docker registries, ensuring secure and efficient image retrieval.

### Namespace Management

- **Namespace Configuration**: Specify the Kubernetes namespace where the microservice will be deployed.
- **Namespace Creation Control**: Use the `create_namespace` flag to control whether the module should create the namespace or use an existing one:
  - **`create_namespace: true`** (default): The module creates the namespace with appropriate labels. Use this for new deployments or when you want the module to fully manage the namespace lifecycle.
  - **`create_namespace: false`**: The module uses an existing namespace without creating it. Use this when:
    - The namespace already exists in the cluster
    - Multiple deployments share the same namespace
    - Namespaces are managed centrally by cluster administrators
    - Using GitOps workflows where namespaces are managed separately

  **Important**: When `create_namespace: false`, you must ensure the namespace exists before deploying, otherwise the deployment will fail.

### Version Control

- **Version Specification**: Deploy specific versions of microservices, such as the main branch or a particular merge request (`review-<id>`), facilitating controlled rollouts and testing.

### Container Specification

- **App Container Configuration**: Define the main application container, including:
- **Container Image**: Set the container image, which is computed based on the artifact store and code project path.
- **Resources**: Allocate CPU and memory resources to optimize performance and cost.
- **Environment Variables and Secrets**: Manage configuration data and sensitive information securely.
- **Ports**: Configure container and service ports, including network and application protocols.

- **Sidecar Containers**: Include additional sidecar containers to extend functionality, such as logging agents or proxies.

### Networking and Ingress

- **Ingress Configuration**: Set up Kubernetes Ingress resources to manage external access to the microservice, including hostname and path routing.

### Availability and Scaling

- **Replicas Management**: Define the minimum number of pod replicas to ensure availability.
- **Horizontal Pod Autoscaling (HPA)**:
- **Enable HPA**: Toggle autoscaling based on resource utilization.
- **CPU Utilization Target**: Set target CPU utilization percentage to trigger scaling.
- **Memory Utilization Target**: Specify memory usage thresholds for scaling decisions.

## Benefits

- **Consistency Across Deployments**: By using a standardized API resource, deployments become more predictable and maintainable.
- **Reduced Complexity**: Developers and DevOps teams can manage microservice deployments without dealing with intricate Kubernetes configurations directly.
- **Scalability**: Built-in support for autoscaling ensures that microservices can handle varying loads efficiently.
- **Security**: Securely manage sensitive information like credentials and secrets within the deployment specifications.

## Use Cases

- **Automated CI/CD Pipelines**: Integrate with continuous integration and deployment pipelines to automate microservice deployments upon code changes.
- **Multi-Environment Deployments**: Deploy microservices consistently across different environments with environment-specific configurations.
- **Microservice Updates and Rollbacks**: Easily update microservices to new versions or roll back to previous ones by changing the version specification.
- **Resource Optimization**: Adjust resource allocations and scaling policies to optimize cost and performance based on usage patterns.
