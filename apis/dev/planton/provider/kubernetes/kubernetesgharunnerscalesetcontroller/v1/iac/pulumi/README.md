# KubernetesGhaRunnerScaleSetController Pulumi Module

This directory contains the Pulumi module for deploying the GitHub Actions Runner Scale Set Controller on Kubernetes.

## Prerequisites

- Go 1.21+
- Pulumi CLI
- kubectl configured with cluster access
- Helm (for chart resolution)

## Usage

### With Planton CLI

```bash
planton pulumi up --manifest gha-controller.yaml --stack org/project/env
```

### Standalone Usage

1. Set up the stack input:

```bash
export STACK_INPUT_JSON=$(cat ../hack/manifest.yaml | yq -o=json)
```

2. Run Pulumi:

```bash
pulumi up --stack dev
```

## Module Structure

```
pulumi/
├── main.go           # Pulumi entrypoint
├── Pulumi.yaml       # Project configuration
├── Makefile          # Build and test automation
├── debug.sh          # Local development helper
└── module/
    ├── main.go       # Resource orchestration
    ├── locals.go     # Local value computation
    ├── controller.go # Helm release deployment
    ├── vars.go       # Constants and defaults
    └── outputs.go    # Export key names
```

## Configuration

The module accepts configuration through the `KubernetesGhaRunnerScaleSetControllerStackInput` message, which includes:

- **Target**: The complete resource specification
- **ProviderConfig**: Kubernetes provider configuration

## Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Namespace where controller is deployed |
| `release_name` | Helm release name |
| `chart_version` | Deployed chart version |
| `deployment_name` | Controller deployment name |
| `service_account_name` | Controller service account |
| `metrics_endpoint` | Metrics endpoint (if enabled) |

## Development

### Build

```bash
make build
```

### Test with Preview

```bash
make test
```

### Debug Script

For local development:

```bash
./debug.sh
```

## Troubleshooting

### Helm Chart Not Found

Ensure you have network access to `ghcr.io/actions/actions-runner-controller-charts`.

### CRDs Not Created

The Helm chart installs CRDs automatically. If they're missing, check:

```bash
kubectl get crds | grep actions.github.com
```

### Controller Not Starting

Check controller logs:

```bash
kubectl logs -n arc-system -l app.kubernetes.io/name=gha-runner-scale-set-controller
```

