# GitLab Kubernetes Pulumi Module
## Key Features

- **Declarative Deployment**: The module uses a declarative YAML-based configuration to define all aspects of the GitLab deployment. Users specify the `GitlabKubernetes` API resource, and the module handles the provisioning automatically.

- **Kubernetes Native**: The module is tightly integrated with Kubernetes, automating the creation of namespaces, services, and containers required for running GitLab. It sets up these resources based on the provided configuration, ensuring that GitLab is deployed consistently across environments.

- **Resource Customization**: Users can configure CPU and memory resource requests and limits for GitLab, ensuring that the containerized GitLab deployment is optimized for their specific performance and resource consumption needs.

- **Ingress Support**: The module includes optional ingress support, allowing GitLab to be exposed to external traffic. Users can specify custom hostnames and enable TLS for secure external access to the GitLab instance.

- **Port-Forwarding for Secure Access**: When ingress is disabled, the module provides port-forwarding commands that enable secure local access to GitLab. This allows developers to interact with the GitLab UI from their local machine without exposing it to the public internet.

- **Pulumi-Driven Infrastructure Management**: Built using Pulumi’s Go SDK, this module automates the lifecycle of GitLab infrastructure within Kubernetes. The integration with Pulumi ensures that changes to the API resource are automatically reflected in the Kubernetes infrastructure, simplifying updates and scaling operations.

- **Flexible Namespace Management**: The module provides flexible namespace management through the `create_namespace` flag:
  - When `create_namespace: true`, the module automatically creates a dedicated namespace with appropriate labels
  - When `create_namespace: false`, the module uses an existing namespace (which must be created beforehand)
  - This ensures GitLab deployments are isolated from other workloads while supporting different namespace management strategies

- **Comprehensive Output Management**: After provisioning, the module provides essential information such as:
  - The Kubernetes namespace where GitLab is deployed.
  - The Kubernetes service name that exposes GitLab within the cluster.
  - Commands for setting up port-forwarding to access GitLab locally.
  - Internal and external endpoints for accessing GitLab.
  - Ingress endpoint (if ingress is enabled) for external access to GitLab.

## Usage

To deploy the `gitlab-kubernetes-pulumi-module`, create a `GitlabKubernetes` YAML file that defines the desired configuration for GitLab. Once the YAML is prepared, the following command can be used to apply the configuration and deploy the GitLab instance in your Kubernetes cluster:

```bash
planton pulumi up --stack-input <api-resource.yaml>
```

Refer to the **Examples** section for detailed usage instructions.

## Namespace Management

The module supports flexible namespace management through the `create_namespace` configuration:

### Creating a New Namespace

Set `create_namespace: true` in your API resource to have the module create the namespace automatically:

```yaml
spec:
  target_cluster:
    cluster_name: my-gke-cluster
  namespace:
    value: gitlab-prod
  create_namespace: true
```

The module will:
- Create a new Kubernetes namespace with the specified name
- Apply resource labels for tracking (resource_id, resource_kind, organization, environment)
- Manage the namespace as part of the Pulumi stack lifecycle

### Using an Existing Namespace

Set `create_namespace: false` to deploy into an existing namespace:

```yaml
spec:
  target_cluster:
    cluster_name: my-gke-cluster
  namespace:
    value: shared-services
  create_namespace: false
```

**Important:** The namespace must exist before running `pulumi up`, otherwise the deployment will fail.

This is useful when:
- The namespace is managed by a separate process or team
- Multiple applications share the same namespace
- Namespace policies or quotas are pre-configured

## Pulumi Integration

The module is built on top of Pulumi’s Go SDK, providing deep integration with Kubernetes. It processes the `GitlabKubernetes` API resource and translates it into the necessary Kubernetes resources, such as services, namespaces, and ingress configurations. Pulumi manages the lifecycle of the resources, ensuring that the GitLab deployment can be easily updated, scaled, or removed by modifying the API resource YAML file.

### Key Pulumi Components

1. **Kubernetes Provider**: The module configures the Kubernetes provider using the `kubernetes_credential_id` provided in the API resource, ensuring that all resources are deployed in the correct Kubernetes cluster.

2. **Namespace Management**: The module provides conditional namespace creation based on the `create_namespace` flag:
   - **Automatic Creation** (`create_namespace: true`): Creates a new namespace with resource labels for tracking and organization
   - **Existing Namespace** (`create_namespace: false`): Uses a pre-existing namespace, which must be created before deployment
   - This provides flexibility for different namespace management strategies while ensuring proper isolation

3. **Kubernetes Services**: The module provisions a Kubernetes service to expose the GitLab instance either within the cluster or externally, depending on the configuration.

4. **Ingress Configuration**: If ingress is enabled, the module configures ingress resources to expose GitLab to external traffic. This includes support for custom hostnames and TLS certificates to secure access.

5. **Port-Forwarding**: For cases where ingress is disabled, the module outputs port-forwarding commands, allowing developers to access the GitLab instance locally using `kubectl`.

6. **Resource Management**: The module allows developers to specify resource requests and limits for GitLab, ensuring that the deployment is appropriately sized for the workload.

7. **Output Management**: After the deployment, the module exports critical information such as:
   - The namespace where GitLab is deployed.
   - The service name for accessing GitLab within the Kubernetes cluster.
   - Port-forwarding commands to access GitLab when ingress is disabled.
   - Internal and external endpoints for accessing GitLab.
   - Ingress endpoint (if enabled) for external access to GitLab.

## Status and Monitoring

The outputs from the Pulumi deployment are captured in the `status.outputs` field. These outputs include details about the Kubernetes service, ingress endpoints, and port-forwarding commands. This information provides an easy way for administrators and developers to monitor the status of the GitLab deployment and access critical information for managing the instance.

## Conclusion

The `gitlab-kubernetes-pulumi-module` simplifies the process of deploying and managing GitLab within Kubernetes clusters. By leveraging Planton's unified API structure and Pulumi's infrastructure-as-code capabilities, the module ensures that GitLab is consistently deployed across different environments. With built-in support for resource customization, ingress configuration, and secure access options, this module offers flexibility and control for deploying GitLab in both development and production environments.