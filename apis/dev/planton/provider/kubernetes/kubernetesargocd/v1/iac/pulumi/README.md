# ArgoCD Kubernetes Pulumi Module

## Key Features

- **Declarative GitOps Approach**: This module leverages ArgoCD’s declarative GitOps capabilities, allowing for continuous deployment automation and version control directly from Git repositories.
  
- **Unified API Structure**: The module follows the Kubernetes-like resource modeling with fields like `apiVersion`, `kind`, `metadata`, and `spec`. This ensures a consistent structure that is easy to understand and reuse across different environments and infrastructures.

- **Kubernetes Native Deployment**: The module optionally provisions namespaces (based on the `create_namespace` flag), and automatically provisions services and ingress for ArgoCD. It integrates directly with your Kubernetes cluster, ensuring that all necessary resources are created and managed within the cluster.

- **Ingress Support**: This module supports Kubernetes ingress configuration for exposing ArgoCD to external traffic. Users can easily configure custom hostnames and TLS settings to expose the ArgoCD UI to the internet or internal networks.

- **Port-Forwarding Commands**: When ingress is disabled for security reasons, the module provides port-forwarding commands that allow developers to access the ArgoCD UI from their local environment using Kubernetes `kubectl` commands. This is particularly useful for internal, secure access.

- **Secure Resource Management**: The module includes provisions for securing the deployment of ArgoCD, including setting up Kubernetes services, ensuring secure internal and external access, and managing secrets and sensitive configurations.

- **Declarative Resource Specification**: The module uses declarative YAML definitions for deploying ArgoCD. Developers simply define the desired state in the `ArgocdKubernetes` API resource, and the module handles the provisioning, management, and scaling of the necessary resources.

- **Seamless Integration with Planton CLI**: The module is fully integrated with Planton's CLI, making it easy to deploy infrastructure by running a single command. The default Pulumi module is used if no specific module is provided, streamlining the deployment process for both simple and advanced use cases.

- **Comprehensive Outputs**: The module captures important output details, including:
  - Namespace in which ArgoCD is deployed.
  - Kubernetes service details.
  - Commands for port-forwarding to access ArgoCD when ingress is disabled.
  - Internal and external endpoints for accessing ArgoCD.
  - Public and private hostnames for accessing ArgoCD from within and outside the Kubernetes cluster.

## Usage

To deploy the `argocd-kubernetes-pulumi-module`, create an `ArgocdKubernetes` YAML file specifying the desired configuration for ArgoCD. Once the YAML is created, you can use the following command to apply the configuration and provision the ArgoCD instance in your Kubernetes cluster:

```bash
planton pulumi up --stack-input <api-resource.yaml>
```

Refer to the **Examples** section for detailed usage instructions.

## Pulumi Integration

This module is built using Pulumi’s Go SDK and integrates seamlessly with Kubernetes. The module processes the `ArgocdKubernetes` API resource and converts it into the required Kubernetes resources, ensuring that everything is set up in a declarative and repeatable manner. Pulumi manages the lifecycle of the resources, ensuring that updates, scaling, and deletions are handled consistently.

### Key Pulumi Components

1. **Kubernetes Provider**: The module uses the Kubernetes provider configured via the provided `kubernetes_credential_id`, ensuring that resources are created in the specified Kubernetes cluster.

2. **Namespace Management**: The module can automatically create a Kubernetes namespace for the ArgoCD deployment, or use an existing namespace based on the `create_namespace` flag, ensuring proper resource isolation and flexibility in namespace governance.

3. **Kubernetes Services**: Kubernetes services are created for ArgoCD, making it accessible within the cluster and optionally exposing it to the external world via ingress.

4. **Ingress Setup**: The module provides built-in support for Kubernetes ingress, allowing users to define custom hostnames and configure TLS certificates. Ingress ensures that the ArgoCD UI can be accessed securely from external networks.

5. **Port-Forwarding**: For cases where ingress is not enabled, the module outputs commands to set up port-forwarding, allowing local access to ArgoCD using `kubectl`.

6. **Security and Access Control**: The module takes care of managing access to ArgoCD, including securing Kubernetes service accounts and setting up proper RBAC configurations.

7. **Outputs**: The module exports key details, including:
   - The namespace in which ArgoCD is deployed.
   - Kubernetes service name for accessing ArgoCD.
   - Port-forwarding commands for secure internal access.
   - Public and internal endpoints for accessing ArgoCD within and outside the Kubernetes cluster.
   - Hostnames for external access to ArgoCD.
