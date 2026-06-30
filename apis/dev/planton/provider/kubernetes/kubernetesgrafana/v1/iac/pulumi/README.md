# Grafana Kubernetes Pulumi Module

## Key Features

- **Declarative Deployment**: The module follows a declarative approach using the `GrafanaKubernetes` API resource, allowing users to define their Grafana configuration, including container resources, namespace, and ingress, in a simple YAML format.
  
- **Kubernetes Native Integration**: The module automates the creation and management of Kubernetes resources such as namespaces, services, and ingress, ensuring that Grafana is deployed and configured seamlessly within the Kubernetes cluster.

- **Resource Customization**: Users can specify CPU and memory resource requests and limits for Grafana’s container, ensuring the deployment is optimized for their performance requirements.

- **Ingress Support**: The module provides optional ingress configuration, allowing users to expose Grafana to external networks. Ingress configurations can include custom hostnames and TLS settings for secure access.

- **Port-Forwarding for Local Access**: If ingress is disabled for security or internal deployment reasons, the module provides port-forwarding commands, allowing secure local access to Grafana via `kubectl`. This feature is useful when Grafana needs to be accessed internally without exposing it to the internet.

- **Pulumi Integration**: Built on top of Pulumi’s Go SDK, the module ensures that infrastructure is managed as code. Any changes made to the `GrafanaKubernetes` API resource will be automatically applied to the underlying Kubernetes infrastructure, simplifying updates and scaling operations.

- **Namespace Isolation**: The module provisions a Kubernetes namespace for Grafana, ensuring the deployment is isolated from other applications running in the cluster. It also reuses existing namespaces if specified.

- **Flexible Namespace Management**: The module provides control over namespace creation through the `create_namespace` flag:
  - When `create_namespace: true`, the module creates a dedicated namespace with appropriate resource labels for tracking and organization.
  - When `create_namespace: false` (default), the module uses an existing namespace, which must be created beforehand. This is useful for deploying multiple components into a shared namespace or when namespace lifecycle is managed separately.

- **Output Management**: The module captures and stores essential deployment details in `status.outputs`, such as:
  - The namespace where Grafana is deployed.
  - The service name that exposes Grafana within the Kubernetes cluster.
  - Commands for setting up port-forwarding when ingress is disabled.
  - Internal and external endpoints for accessing Grafana.
  - Ingress endpoints for public or internal access to Grafana.

## Usage

To deploy Grafana using this module, you need to define a `GrafanaKubernetes` API resource in YAML format that specifies the desired configuration. Once the YAML is created, the following command can be used to deploy the Grafana instance in your Kubernetes cluster:

```bash
planton pulumi up --stack-input <api-resource.yaml>
```

Refer to the **Examples** section for detailed usage instructions.

## Pulumi Integration

The module leverages Pulumi’s infrastructure-as-code capabilities to manage the lifecycle of Grafana within Kubernetes. By processing the `GrafanaKubernetes` API resource, it provisions all necessary Kubernetes components, such as services, namespaces, and ingress, based on the declarative configuration provided. Pulumi ensures that any updates to the API resource are automatically reflected in the Kubernetes infrastructure, reducing manual intervention and ensuring consistency.

### Key Pulumi Components

1. **Kubernetes Provider**: The module sets up the Kubernetes provider using the `kubernetes_credential_id` specified in the API resource, ensuring that resources are created in the correct cluster.

2. **Namespace Management**: A dedicated Kubernetes namespace is created (or reused) for the Grafana deployment, ensuring isolation and avoiding resource conflicts with other workloads in the cluster.

3. **Kubernetes Services**: The module provisions services that expose Grafana within the cluster and, optionally, to external networks based on ingress settings.

4. **Ingress Configuration**: If enabled, the module configures ingress resources to expose Grafana to external traffic. Custom hostnames and TLS settings can be specified for secure access from outside the Kubernetes cluster.

5. **Port-Forwarding**: For cases where ingress is disabled, the module generates port-forwarding commands, allowing developers to access Grafana locally without exposing it to the public internet. This is particularly useful in internal, secure deployments.

6. **Resource Requests and Limits**: Users can define CPU and memory resources for Grafana’s container, ensuring that it operates efficiently within the available cluster resources.

7. **Outputs**: After the deployment is completed, the module exports important details such as:
   - The namespace where Grafana is deployed.
   - The service name for accessing Grafana within the Kubernetes cluster.
   - Port-forwarding commands for accessing Grafana locally.
   - Internal and external endpoints for accessing Grafana.
   - Ingress endpoint for accessing Grafana externally (if enabled).

## Status and Monitoring

All deployment outputs, including service names, ingress endpoints, and port-forwarding commands, are captured and stored in the `status.outputs` field. This information allows administrators and developers to monitor the deployment and access essential details for managing and interacting with the Grafana instance.

## Conclusion

The `grafana-kubernetes-pulumi-module` offers a simplified and scalable solution for deploying Grafana in Kubernetes clusters. By adopting a declarative approach and leveraging Pulumi’s infrastructure-as-code capabilities, the module ensures that Grafana is deployed consistently across different environments. With features like resource customization, ingress configuration, and secure access options through port-forwarding, this module provides flexibility and control, making it suitable for both development and production deployments of Grafana.