# Elasticsearch Kubernetes Pulumi Module

This Pulumi module streamlines the deployment and management of Elasticsearch and Kibana on Kubernetes clusters. By
leveraging a standardized API resource definition, it enables developers to configure and deploy complex Elasticsearch
infrastructures with minimal effort. The module supports additional features like optional persistence, ingress
configurations, and resource customization, providing a comprehensive solution for search and analytics platforms.

## Key Features

- **Standardized API Resource**: Utilizes a consistent API structure with `apiVersion`, `kind`, `metadata`, `spec`, and
  `status`, simplifying resource definitions and management.

- **Customizable Elasticsearch Deployment**:
    - **Replicas**: Configure the number of Elasticsearch and Kibana replicas to meet scalability requirements.
    - **Resource Allocation**: Define CPU and memory resources for both Elasticsearch and Kibana containers.
    - **Persistence Options**: Optionally enable persistence for Elasticsearch data by attaching persistent volumes.

- **Kibana Integration**: Optionally deploy Kibana alongside Elasticsearch for data visualization and management.

- **Ingress Configuration**:
    - **External and Internal Access**: Supports ingress setup using Istio and the Gateway API for both external and
      internal clients.
    - **TLS Termination**: Automates TLS certificate provisioning and management using Cert-Manager.
    - **HTTPS Redirects**: Configures HTTP to HTTPS redirects for secure access.

- **Kubernetes Provider Integration**: Utilizes Kubernetes cluster credentials to set up providers, facilitating
  deployments across different cloud environments and clusters.

- **Pulumi Integration**: Written in Golang, the module leverages Pulumi for infrastructure as code, enabling seamless
  integration into existing workflows.

- **Outputs Captured in Status**: Pulumi outputs are captured in `status.outputs`, making it easier to retrieve
  deployment information such as service endpoints, credentials, and commands.

- **Resource Labeling and Annotation**: Supports adding custom labels and annotations to Kubernetes resources for better
  organization and management.

- **Scalability and Flexibility**: Easily scale the number of replicas and adjust resource limits to accommodate
  changing workloads and performance needs.

- **Security Features**: Integrates with Cert-Manager for automated TLS certificate provisioning and management,
  enhancing the security of your deployments.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
- [Module Components](#module-components)
    - [Namespace Management](#namespace-management)
    - [Elasticsearch Deployment](#elasticsearch-deployment)
    - [Kibana Deployment](#kibana-deployment)
    - [Ingress Configuration](#ingress-configuration)
- [Outputs](#outputs)
- [Contributing](#contributing)
- [License](#license)

## Prerequisites

- **Kubernetes Cluster**: A Kubernetes cluster with appropriate permissions.
- **Elasticsearch Operator**: The Elasticsearch Operator installed in the cluster.
- **Istio and Gateway API**: For ingress configurations, Istio and the Gateway API should be installed.
- **Cert-Manager**: Installed for certificate management.
- **Pulumi CLI**: Installed on your local machine.
- **Golang Environment**: Set up if you plan to modify the module code.

## Installation

Clone the repository containing the Pulumi module:

```bash
git clone https://github.com/your-org/elasticsearch-kubernetes-pulumi-module.git
```

Install the required dependencies:

```bash
cd elasticsearch-kubernetes-pulumi-module
go mod download
```

## Usage

Refer to [example](example.md) for usage instructions.

## Module Components

### Namespace Management

The module provides flexible namespace management through the `create_namespace` flag:

- **`create_namespace: true`** (recommended): The module will create the namespace with proper labels and configuration. Use this when you want the component to manage the full namespace lifecycle.

- **`create_namespace: false`**: The module will use an existing namespace. The namespace must already exist in the cluster. Use this when:
  - The namespace is managed separately (e.g., by another component or tool)
  - You're deploying multiple resources into a shared namespace
  - Namespace policies are managed centrally

**Important**: When `create_namespace: false`, ensure the namespace exists before deploying this component, or the deployment will fail.

### Elasticsearch Deployment

- **Elasticsearch Cluster**:
    - Deploys the specified number of Elasticsearch nodes with the provided resource allocations.
    - Supports optional data persistence by configuring persistent volume claims.
    - Configures Elasticsearch settings such as node roles and version.
- **Resource Customization**:
    - **Replicas**: Set the number of Elasticsearch replicas.
    - **Resources**: Define CPU and memory requests and limits for the Elasticsearch container.
    - **Persistence**: Enable or disable data persistence and specify disk size.

### Kibana Deployment

- **Optional Deployment**:
    - Kibana deployment can be enabled or disabled based on requirements.
- **Resource Customization**:
    - **Replicas**: Set the number of Kibana replicas.
    - **Resources**: Define CPU and memory requests and limits for the Kibana container.

### Ingress Configuration

- **External Access**:
    - Configures ingress resources for external access to Elasticsearch and Kibana.
    - Uses Istio and the Gateway API for ingress setup.
- **TLS Management**:
    - Automates TLS certificate provisioning using Cert-Manager.
    - Manages certificates and secrets within Kubernetes.
- **HTTPS Redirects**:
    - Sets up HTTP to HTTPS redirects to enforce secure connections.
- **Internal Access**:
    - Provides internal hostnames for access within the Kubernetes cluster.
- **Port Forwarding Commands**:
    - Generates commands for port-forwarding to access services from a developer's machine when ingress is disabled.

## Outputs

After deployment, the module provides several outputs:

- **Namespace**: The Kubernetes namespace where Elasticsearch and Kibana are deployed.
- **Elasticsearch Outputs**:
    - **Service Name**: Kubernetes service name for Elasticsearch.
    - **Port Forward Command**: Command to set up port-forwarding to Elasticsearch.
    - **Kubernetes Endpoint**: Internal endpoint for accessing Elasticsearch within the cluster.
    - **External Hostname**: Public URL for accessing Elasticsearch from outside the cluster.
    - **Internal Hostname**: Internal URL for accessing Elasticsearch from within the cluster.
    - **Credentials**:
        - **Username**: Elasticsearch username.
        - **Password Secret**: Kubernetes secret key where the password is stored.
- **Kibana Outputs**:
    - **Service Name**: Kubernetes service name for Kibana.
    - **Port Forward Command**: Command to set up port-forwarding to Kibana.
    - **Kubernetes Endpoint**: Internal endpoint for accessing Kibana within the cluster.
    - **External Hostname**: Public URL for accessing Kibana from outside the cluster.
    - **Internal Hostname**: Internal URL for accessing Kibana from within the cluster.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub.

## License

This project is licensed under the [MIT License](LICENSE).
