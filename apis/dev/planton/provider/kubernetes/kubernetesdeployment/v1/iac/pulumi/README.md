# Microservice Kubernetes Pulumi Module

## Key Features

- **Standardized API Resource Model**  
  Provides a unified way to define and deploy microservices on Kubernetes. By describing container images, resource
  allocations, environment variables, ports, and optional sidecars in a simple API resource, you ensure consistency
  across environments.

- **Automated Kubernetes Resource Creation**  
  Automatically creates Namespaces, Deployments, Services, and optional Ingress resources (via Gateway API and
  Cert-Manager) based on the provided specifications. Eliminates the need for hand-maintained YAML files.

- **Container Configuration**  
  Supports detailed specifications for both primary and sidecar containers, including their images, ports, lifecycle
  hooks, and resource limits/requests.

- **Ingress Integration**  
  When enabled, the module sets up Istio-based or Gateway API-based ingress, creating gateways, routes, and TLS
  certificates if requested, allowing for secure external or internal traffic routing.

- **Scalability & Availability**  
  Optionally configure minimum replicas and horizontal pod autoscaling (HPA) thresholds. This ensures your microservice
  can scale to meet demand while staying within resource budgets.

- **Secret Management**  
  Securely handles environment secrets via Kubernetes Secrets. Integrates with external providers (e.g., GCP Secret
  Manager) so sensitive information stays out of version control and container images.

- **Output Exports**  
  Exports useful values such as namespace, service name, internal service FQDN, and port-forward commands. These can be
  leveraged for further automation or debugging.

## Usage

See [example](example.md) for usage details and step-by-step examples. In general:

1. Define a YAML resource describing your microservice using the **KubernetesDeployment** API.
2. Run:
   ```bash
   planton pulumi up --stack-input <your-microservice-file.yaml>
   ```

to apply the resource on your cluster.

## Getting Started

1. **Craft Your Specification**  
   Include container info, environment variables, secrets, ports, and (optionally) ingress preferences. If you need
   sidecars, list them alongside your main container.

2. **Apply via CLI**  
   Execute `planton pulumi up --stack-input <microservice-spec.yaml>` (or your organization’s standard CLI command). The
   Pulumi module automatically compiles your specification into Kubernetes resources.

3. **Validate & Observe**  
   Check the logs of your microservice, confirm the Deployment and Service are created, and if ingress is enabled,
   verify external access or domain routing.

## Module Structure

1. **Initialization**  
   Reads your `KubernetesDeploymentStackInput` (containing cluster creds, Docker config, resource definitions), sets
   up local variables, and merges labels.

2. **Provider Setup**  
   Establishes a Pulumi Kubernetes Provider for your target cluster.

3. **Namespace Management**  
   Creates or uses a Kubernetes namespace to house all your microservice resources, controlled by the `create_namespace` flag:
   - **`create_namespace: true`**: The module creates the namespace with appropriate labels
   - **`create_namespace: false`**: The module uses an existing namespace (which must already exist in the cluster)

4. **Image Pull Secret (Optional)**  
   If Docker credentials (`docker_config_json`) are provided, creates a `kubernetes.io/dockerconfigjson` secret and
   configures it in the Deployment’s Pod spec.

5. **Deployment Configuration**  
   Generates the Deployment with the main container and any sidecars specified. Injects environment variables, secrets,
   lifecycle hooks, port configurations, and resource limits/requests.

6. **Service Configuration**  
   Creates a Kubernetes Service for internal cluster networking. Binds exposed ports, enabling other services to
   discover and communicate with your microservice.

7. **Ingress Setup (Optional)**  
   If requested, sets up Istio or Gateway-based routes, TLS certificates, or other ingress logic, providing external
   access or advanced networking features.

8. **Secret Management**  
   Creates a “main” Kubernetes Secret for storing your environment secrets. This allows sensitive credentials to remain
   secure and out of source code.

9. **Output Exports**  
   Publishes final references (e.g., namespace, service name, cluster endpoints), which can aid in post-deployment
   automation.

## Benefits

- **Simplified Deployment**  
  Focus on high-level configuration rather than writing raw Kubernetes manifests. Consistent patterns reduce the risk of
  misconfiguration.

- **Security & Compliance**  
  Minimizes exposure of secrets, enabling best practices for secret management, credential injection, and TLS
  provisioning.

- **Scalability**  
  Easily set minimum replicas or enable horizontal pod autoscaling for traffic spikes and resilience.

- **Extensibility**  
  The module is built on Pulumi’s Kubernetes provider. You can augment or override resources if your team needs advanced
  configurations (e.g., custom pod security policies).

## Contributing

Contributions are always welcome! Please open an issue or submit a pull request in the main repository if you want to
add features, fix bugs, or improve documentation.

## License

This project is licensed under the [MIT License](LICENSE). Feel free to adapt it for your internal workflows.
