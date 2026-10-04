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
- **Default Module Configuration**: Automatically configure IaC inputs using default Pulumi modules, reducing setup complexity.
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

1. **Namespace Creation**: Creates the namespace named in `spec.namespace`, with proper labels, when `spec.create_namespace` is true
2. **Helm Release**: Deploys the operator using the official Altinity Helm chart
3. **Resource Configuration**: Applies resource limits and requests from the spec
4. **Output Capture**: Exports the namespace to outputs for reference

### Helm Chart Details

- **Chart Name**: `altinity-clickhouse-operator`
- **Repository**: `https://docs.altinity.com/clickhouse-operator/`
- **Version**: `spec.chart_version` (default `0.27.2`)
- **Key Values**:
  - `operator.createCRD: true` - Automatically installs CRDs
  - `operator.resources` - Resource limits from spec

## API Reference

### KubernetesAltinityOperatorSpec
Defines the desired state of the operator deployment.

- **namespace**: Namespace to install the operator into (a literal name or a reference to a `KubernetesNamespace`)
- **create_namespace**: Create the namespace before installing, and delete it with the resource
- **chart_version**: Helm chart version (default `0.27.2`)
- **watch_namespaces**: Namespaces the operator watches (empty = its own namespace only)
- **namespace_scoped_rbac**: Namespace-scoped Roles/RoleBindings instead of cluster-wide RBAC
- **operator_credentials**: Credentials the operator uses to connect to every managed ClickHouse instance
- **metrics**: The metrics-exporter sidecar
- **crd_hook**: The hook job that applies the CRDs on install and upgrade
- **resources**: Operator container CPU and memory (empty = the chart defaults, no requests or limits)
- **service_monitor_enabled**, **node_selector**, **tolerations**, **image_pull_secrets**, **image**, **helm_values**: scheduling, image, and raw Helm value overrides

### KubernetesAltinityOperatorOutputs
Provides outputs from the deployed operator infrastructure.

- **namespace**: Kubernetes namespace where the operator is deployed
- **release_name**, **deployment_name**, **credentials_secret_name**, **metrics_endpoint**: the Helm release, the operator Deployment, the operator credentials Secret, and the metrics endpoint

## Development

### Building the Module

```bash
cd catalog/kubernetes/kubernetesaltinityoperator/iac/pulumi
go build .
```

### Local Testing

From this directory (it holds `Pulumi.yaml`, so the planton CLI runs this module):

```bash
planton pulumi preview --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack> --kube-context <context>
```

The CLI wraps the manifest into a `KubernetesAltinityOperatorIacInput` (under `target`) and hands it to the module through `IAC_INPUT_YAML_FILE`. The module also reads that input from the Pulumi config key `planton:iac-input` or from `IAC_INPUT_YAML` (YAML content).

### Updating Dependencies

The module's dependencies, including `github.com/pulumi/pulumi-kubernetes/sdk/v4`, are pinned in the repository's root `go.mod`.

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

