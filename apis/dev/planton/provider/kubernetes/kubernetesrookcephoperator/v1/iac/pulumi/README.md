# KubernetesRookCephOperator Pulumi Module

This Pulumi module deploys the Rook Ceph Operator on a Kubernetes cluster.

## Overview

The module:
1. Optionally creates a namespace for the operator
2. Deploys the Rook Ceph Operator Helm chart
3. Configures CSI drivers based on spec

## Prerequisites

- Pulumi CLI installed
- Go 1.21+
- Kubernetes cluster access
- Pulumi Kubernetes plugin

## Installation

Install required Pulumi plugins:

```bash
make install-pulumi-plugins
```

## Usage

### With Planton CLI

```bash
planton pulumi up \
  --manifest manifest.yaml \
  --stack org/project/env
```

### Standalone

1. Set the stack input environment variable:

```bash
export STACK_INPUT=$(cat manifest.yaml | yq -o json)
```

2. Run Pulumi:

```bash
pulumi up --stack local
```

## Configuration

The module accepts a `KubernetesRookCephOperatorStackInput` with:

- `target`: The KubernetesRookCephOperator resource specification
- `provider_config`: Kubernetes provider credentials

### Key Spec Fields

| Field | Description | Default |
|-------|-------------|---------|
| `namespace` | Target namespace | Required |
| `create_namespace` | Create namespace if not exists | `false` |
| `operator_version` | Helm chart version | `v1.16.6` |
| `crds_enabled` | Let Helm manage CRDs | `true` |
| `container.resources` | Operator pod resources | See defaults |
| `csi.enable_rbd_driver` | Enable RBD CSI driver | `true` |
| `csi.enable_cephfs_driver` | Enable CephFS CSI driver | `true` |

## Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Namespace where operator is deployed |
| `helm_release_name` | Name of the Helm release |
| `webhook_service` | Webhook service name |

## Local Development

Build the module:

```bash
make build
```

Test with preview:

```bash
make test
```

## Debugging

Use the debug script:

```bash
./debug.sh
```

This sets up the environment and runs Pulumi preview.

## Resources Created

- **Namespace** (optional): If `create_namespace: true`
- **Helm Release**: Rook Ceph Operator chart

## Dependencies

- `pulumi-kubernetes` plugin v4.x
- Helm chart: `rook-ceph` from `https://charts.rook.io/release`

## Troubleshooting

### Operator Pod Not Starting

```bash
kubectl get pods -n rook-ceph -l app=rook-ceph-operator
kubectl logs -n rook-ceph -l app=rook-ceph-operator
```

### CRDs Not Installed

Ensure `crds_enabled: true` in spec, or install CRDs manually.

### Helm Release Issues

```bash
helm list -n rook-ceph
helm history <release-name> -n rook-ceph
```

## Additional Resources

- [Component README](../../README.md)
- [Examples](../../examples.md)
- [Rook Documentation](https://rook.io/docs/rook/latest/)
