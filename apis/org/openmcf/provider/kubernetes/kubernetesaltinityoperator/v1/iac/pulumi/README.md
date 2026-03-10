# KubernetesAltinityOperator Pulumi Module

## Key Features

### Standardized API Resource Structure
- **apiVersion & kind**: Aligns with Kubernetes standards, ensuring familiarity and ease of integration.
- **metadata**: Facilitates resource identification and management through standard Kubernetes metadata fields.
- **spec**: Defines the desired state of the operator deployment, including container resource specifications.
- **status**: Provides real-time updates and outputs from the deployed infrastructure, enhancing visibility and monitoring.

### Comprehensive Operator Configuration
- **Resource Allocation**: Define CPU and memory resources for the operator pod to optimize performance and cost.
- **Automated CRD Installation**: Automatically installs ClickHouse CRDs required for cluster management.
- **Namespace Management**: Creates a dedicated `kubernetes-altinity-operator` namespace for clean resource isolation.
- **Helm-Based Deployment**: Leverages the official Altinity Helm chart for reliable, version-controlled deployments.

### Seamless Kubernetes Integration
- **Multi-Cluster Support**: Deploy the operator to any Kubernetes cluster with proper credentials.
- **Kubernetes Credentials Management**: Securely manage cluster credentials through Planton's credential system.
- **Automated Outputs Handling**: Capture and manage Pulumi outputs within the API resource status, providing essential information such as namespace.

### Developer-Friendly CLI
- **Unified Deployment Command**: Utilize the `planton pulumi up --manifest <api-resource.yaml>` command to deploy the operator effortlessly.
- **Default Module Configuration**: Automatically configure stack inputs using default Pulumi modules, reducing setup complexity.
- **Git Integration**: Specify custom Pulumi modules via Git repository details for customized deployments.

### Production-Grade Deployment
- **Atomic Deployments**: Helm releases are atomic, ensuring all-or-nothing deployments with automatic rollback on failure.
- **Cleanup on Fail**: Automatically cleans up resources if deployment fails.
- **Configurable Timeouts**: 300-second timeout ensures sufficient time for operator initialization.
- **Resource Limits**: Enforces resource limits to prevent resource exhaustion.

## Installation

To use the KubernetesAltinityOperator Pulumi module, ensure that you have:
- Pulumi CLI installed
- Access to a Kubernetes cluster
- Valid Kubernetes cluster credentials configured in Planton

## Usage

Refer to the examples section for detailed usage instructions.

## Module Architecture

### Components

1. **Namespace Creation**: Creates the `kubernetes-altinity-operator` namespace with proper labels
2. **Helm Release**: Deploys the operator using the official Altinity Helm chart
3. **Resource Configuration**: Applies resource limits and requests from the spec
4. **Output Capture**: Exports the namespace to stack outputs for reference

### Helm Chart Details

- **Chart Name**: `altinity-clickhouse-operator`
- **Repository**: `https://docs.altinity.com/clickhouse-operator/`
- **Version**: `0.23.6` (configurable in vars.go)
- **Key Values**:
  - `operator.createCRD: true` - Automatically installs CRDs
  - `operator.resources` - Resource limits from spec

## API Reference

### KubernetesAltinityOperatorSpec
Defines the desired state of the operator deployment.

- **target_cluster**: Target Kubernetes cluster configuration with credential reference
- **container**: Container resource specifications for the operator pod

### KubernetesAltinityOperatorSpecContainer
Specifies the container-level configurations for the operator.

- **resources**: CPU and memory resource allocations
  - **requests**: Guaranteed resources (default: 100m CPU, 256Mi memory)
  - **limits**: Maximum resources (default: 1000m CPU, 1Gi memory)

### KubernetesAltinityOperatorStackOutputs
Provides outputs from the deployed operator infrastructure.

- **namespace**: Kubernetes namespace where the operator is deployed (always `kubernetes-altinity-operator`)

## Development

### Building the Module

```bash
cd apis/project/planton/provider/kubernetes/kubernetesaltinityoperator/v1/iac/pulumi
make build
```

### Local Testing

```bash
# Set up stack input
export PULUMI_STACK_INPUT=/path/to/manifest.yaml

# Run locally
./debug.sh
```

### Updating Dependencies

```bash
make update-deps
```

## Troubleshooting

### Operator Pod Not Starting

Check the operator pod logs:
```bash
kubectl logs -n kubernetes-altinity-operator -l app.kubernetes.io/name=altinity-clickhouse-operator
```

### CRDs Not Installed

Verify CRD installation:
```bash
kubectl get crds | grep clickhouse
```

If CRDs are missing, check the Helm values and ensure `operator.createCRD` is set to `true`.

### Resource Limits Too Low

If the operator is being OOMKilled or CPU throttled, increase the resource limits in the manifest.

## Contributing

Contributions are welcome! Please refer to the contributing guidelines for more information on how to get involved.

## License

This project is licensed under the [MIT License](LICENSE).

