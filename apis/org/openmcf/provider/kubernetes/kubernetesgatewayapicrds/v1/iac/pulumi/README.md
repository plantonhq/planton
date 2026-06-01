# KubernetesGatewayApiCrds Pulumi Module

This Pulumi module installs Kubernetes Gateway API CRDs on any Kubernetes cluster.

## Overview

The module applies the official Gateway API CRD manifests from the [kubernetes-sigs/gateway-api](https://github.com/kubernetes-sigs/gateway-api) releases to the target cluster.

## Prerequisites

- Go 1.21+
- Pulumi CLI
- Access to target Kubernetes cluster

## Installation

### Install Pulumi Plugins

```bash
make install-pulumi-plugins
```

### Build

```bash
make build
```

## Usage

### With OpenMCF CLI

```bash
openmcf pulumi up --manifest gateway-api-crds.yaml
```

### Direct Pulumi Usage

1. Set the stack input as an environment variable:

```bash
export STACK_INPUT_FILE_PATH=/path/to/manifest.yaml
```

2. Run Pulumi:

```bash
pulumi up
```

## Configuration

The module accepts configuration via the `KubernetesGatewayApiCrdsStackInput` protobuf message:

| Field | Description |
|-------|-------------|
| `target.spec.version` | Gateway API version to install (default: v1.2.1) |
| `target.spec.install_channel.channel` | standard or experimental (default: standard) |
| `provider_config` | Kubernetes provider configuration |

## Outputs

| Output | Description |
|--------|-------------|
| `installed_version` | The Gateway API version that was installed |
| `installed_channel` | The channel (standard/experimental) |
| `installed_manifest_url` | Full URL of the CRD bundle that was applied |

## Testing

Run a preview with the test manifest:

```bash
make test
```

## Module Structure

```
pulumi/
├── main.go           # Pulumi entrypoint
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and test automation
├── README.md         # This file
├── overview.md       # Architecture overview
└── module/
    ├── main.go       # Resource creation logic
    ├── locals.go     # Computed values
    ├── outputs.go    # Stack outputs
    └── vars.go       # Constants and URLs
```

## Troubleshooting

### CRDs Not Installing

Check that:
1. Kubernetes provider credentials are valid
2. The specified version exists in Gateway API releases
3. Cluster has network access to GitHub

### Permission Denied

Ensure the service account has cluster-admin or equivalent permissions to create CRDs.

## References

- [Gateway API Releases](https://github.com/kubernetes-sigs/gateway-api/releases)
- [Gateway API Documentation](https://gateway-api.sigs.k8s.io/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)
