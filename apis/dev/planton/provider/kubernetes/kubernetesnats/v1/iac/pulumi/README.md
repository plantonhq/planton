# Pulumi Module for deploying NATS on Kubernetes

The NATS Kubernetes Pulumi module automates the deployment and management of a NATS cluster within Kubernetes. This
module simplifies deploying scalable, secure, and high-availability NATS clusters using a standardized YAML-based
configuration. Leveraging Pulumi for infrastructure-as-code, it ensures seamless integration, repeatable deployments,
and minimal operational overhead.

## Key Features

### 1. **Kubernetes Provider Integration**

The module automatically configures a Kubernetes provider using cluster credentials defined in the stack input, ensuring
secure communication and authentication.

### 2. **Namespace Management**

Automatically creates and manages a dedicated namespace for the NATS cluster, isolating the resources from other
Kubernetes workloads for enhanced security and resource organization.

### 3. **NATS Cluster Deployment**

Deploys the official NATS Helm chart (version 1.3.6), allowing easy customization of replicas, resource allocation (CPU
and memory), and JetStream persistence configurations.

### 4. **JetStream Support**

Offers optional JetStream configuration, enabling persistent messaging streams with configurable storage sizes for
robust data retention and recovery.

### 5. **Authentication Management**

Supports Bearer Token and Basic Authentication schemes, automatically provisioning Kubernetes secrets with randomly
generated credentials to secure access to the NATS server.

### 6. **TLS Security**

Automatically generates and provisions self-signed TLS certificates stored in Kubernetes secrets for secure, encrypted
communication within the NATS cluster.

### 7. **Ingress and External Access**

Supports external access via Kubernetes LoadBalancer or configurable ingress, providing flexibility for exposing NATS
services externally using DNS-based routing.

### 8. **Detailed Pulumi Stack Outputs**

Generates comprehensive deployment outputs including namespace details, internal and external URLs, TLS and
authentication secret information, JetStream domain, and metrics endpoints.

## Usage

### Prerequisites

* **Pulumi**: Ensure Pulumi is installed and configured for your Kubernetes environment.
* **Kubernetes Cluster**: Access to a Kubernetes cluster with valid credentials configured.
* **Planton CLI**: The Planton CLI must be installed for executing
  `planton pulumi up --stack-input <api-resource.yaml>`.

### Example Deployment

Refer to `examples.md` for practical examples of various NATS deployments using different authentication methods,
JetStream configurations, resource specifications, and ingress options.

### Deploying with Pulumi

```bash
planton pulumi up --stack-input <your-config.yaml>
```

## Outputs

After successful deployment, the module provides several outputs:

* **Namespace**: Kubernetes namespace used for deployment.
* **Client URLs**: Internal and external URLs for accessing the NATS service.
* **Auth and TLS Secret Details**: Secret names and keys for authentication and TLS.
* **JetStream Domain**: Configured domain for JetStream, if enabled.
* **Metrics Endpoint**: URL for accessing NATS metrics.

These outputs facilitate easy integration and management of the deployed NATS cluster.

## Benefits

* **Infrastructure-as-Code**: Version-controlled, repeatable, and automated deployments using Pulumi.
* **Security and Isolation**: Automated namespace and credential management ensure robust security.
* **Flexibility and Scalability**: Easily configurable for varying resource needs and secure communication.
* **Operational Efficiency**: Reduced manual effort in setting up and maintaining NATS clusters.

## Development

Use the provided Makefile to manage the module lifecycle:

```bash
make deps   # Install dependencies
make vet    # Run Go vet for linting
make fmt    # Format Go code
make build  # Run dependencies, linting, and formatting
```

### Debugging

For debugging purposes, use the provided debug script:

```bash
./debug.sh
```

This will build and launch the module with debugging enabled at port 2345.

## Conclusion

The NATS Kubernetes Pulumi module significantly simplifies the deployment and operational management of secure, scalable
NATS clusters. By utilizing standardized configurations and automated resource management, it enables developers and
DevOps teams to focus on higher-value tasks, ensuring efficient, repeatable, and secure NATS cluster management in
Kubernetes environments.
