# Keycloak Kubernetes Pulumi Module

## Key Features

- **API Resource-Driven Deployment:** The module follows the `KeycloakKubernetes` API resource model, which uses a standard Kubernetes-style API resource with `apiVersion`, `kind`, `metadata`, `spec`, and `status`. This approach ensures a familiar structure for managing Keycloak deployments, allowing for easy integration with other Kubernetes-based systems.

- **Container Resource Management:** The `KeycloakKubernetes` API resource allows for fine-grained control over the Keycloak container’s resource limits, including CPU and memory requests and limits. This ensures that the Keycloak deployment is appropriately sized for the environment and the workload, enabling both performance and cost efficiency.

- **Namespace and Service Creation:** The module automatically creates a dedicated Kubernetes namespace for the Keycloak instance, ensuring proper isolation of resources. Additionally, it creates the necessary Kubernetes services to allow internal and, optionally, external communication with the Keycloak instance.

- **Ingress Configuration:** The module supports the optional deployment of an ingress controller to expose the Keycloak service externally. This feature allows you to manage Keycloak either internally within the Kubernetes cluster or externally via a public or private endpoint. You can configure hostnames and TLS settings through the `IngressSpec` to ensure secure access.

- **Port-Forwarding for Secure Local Access:** If ingress is not enabled, the module provides a `port_forward_command`, allowing developers to access Keycloak from their local machine securely through Kubernetes port forwarding. This ensures that even without external exposure, Keycloak can be accessed for development and debugging purposes.

- **Output Management:** The module captures and exports key output information about the deployment, such as:
  - The Kubernetes namespace where Keycloak is deployed.
  - The service name of the Keycloak instance.
  - Internal and external hostnames for accessing Keycloak.
  - A command for setting up port forwarding when ingress is disabled.
  
  These outputs are stored in `status.outputs`, ensuring that developers have easy access to essential connection details post-deployment.

## Usage

Refer to the **example section** for detailed usage instructions on how to configure the API resource and use this Pulumi module.

## Inputs

The following key inputs are supported by the module from the `KeycloakKubernetes` API resource:

- **kubernetes_credential_id**: (Required) The Kubernetes cluster credentials used to authenticate and deploy resources on the target cluster.

- **container**: (Required) Defines the resource configuration for the Keycloak container, including CPU and memory requests and limits. This ensures the container is deployed with sufficient resources to handle the expected load.

- **ingress**: (Optional) Configures an ingress controller to expose Keycloak externally, allowing access to the service from outside the Kubernetes cluster. The ingress configuration includes options for defining the host, paths, and TLS settings for secure access.

### Namespace Management

- **namespace**: (Required) The Kubernetes namespace where Keycloak will be deployed. Can be a literal value or reference to a KubernetesNamespace resource.

- **create_namespace**: (Optional, default: false) Controls whether the module should create the namespace:
  - `true`: The module will create the namespace specified in the `namespace` field
  - `false`: The namespace must already exist; the module will only use it

**Example with namespace creation:**

```yaml
spec:
  target_cluster:
    cluster_name: "prod-gke"
  namespace:
    value: "auth-keycloak"
  create_namespace: true
  container:
    resources:
      requests:
        cpu: "100m"
        memory: "256Mi"
      limits:
        cpu: "1000m"
        memory: "1Gi"
```

**Example using existing namespace:**

```yaml
spec:
  target_cluster:
    cluster_name: "prod-gke"
  namespace:
    value: "shared-auth"
  create_namespace: false  # Assumes 'shared-auth' namespace already exists
  container:
    resources:
      requests:
        cpu: "100m"
        memory: "256Mi"
      limits:
        cpu: "1000m"
        memory: "1Gi"
```

## Outputs

The module provides several key outputs for managing and accessing the deployed Keycloak instance:

- **namespace**: The Kubernetes namespace where the Keycloak instance is deployed.
- **service**: The name of the Kubernetes service associated with Keycloak.
- **port_forward_command**: A command for setting up port forwarding to access Keycloak from a local machine when ingress is disabled.
- **kube_endpoint**: The internal Kubernetes endpoint for accessing Keycloak from within the cluster.
- **external_hostname**: The public hostname for accessing Keycloak from external clients when ingress is enabled.
- **internal_hostname**: The internal hostname for accessing Keycloak from within the Kubernetes network.

## Benefits

This Pulumi module provides a standardized and declarative way to manage Keycloak deployments on Kubernetes. It abstracts away the complexities of infrastructure setup, such as creating namespaces, configuring container resources, and managing ingress or service exposure. Developers can focus on configuring the Keycloak deployment using a simple YAML specification, while the module handles all underlying Kubernetes resources.

By exporting key outputs and enabling both internal and external access to Keycloak, the module offers flexibility for various use cases, including secure internal authentication, external client access, and local development environments. The module's ability to scale with the resource requirements of Keycloak ensures that deployments can grow with your infrastructure needs, making it a suitable solution for development, staging, and production environments.
