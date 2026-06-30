# Civo Kubernetes Node Pool - Pulumi Module

This directory contains the Pulumi implementation for managing Civo Kubernetes node pools. It's designed to be invoked by the Planton CLI, but can also be used standalone.

## Overview

The Pulumi module provisions:
- A node pool within an existing Civo Kubernetes cluster
- Configurable node count and instance size
- Optional autoscaling with min/max bounds
- Resource tagging for organization

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) (v3.x)
- [Go](https://golang.org/dl/) 1.21+
- Civo account and API token
- Existing Civo Kubernetes cluster
- `kubectl` configured

## Quick Start

### 1. Set up stack input

Create `stack-input.json`:

```json
{
  "target": {
    "apiVersion": "civo.planton.dev/v1",
    "kind": "CivoKubernetesNodePool",
    "metadata": {
      "name": "workers",
      "id": "ciknp-workers-123"
    },
    "spec": {
      "nodePoolName": "general-workers",
      "cluster": {
        "value": "prod-cluster"
      },
      "size": "g4s.kube.medium",
      "nodeCount": 3,
      "autoScale": true,
      "minNodes": 2,
      "maxNodes": 10,
      "tags": ["workload:general"]
    }
  },
  "providerConfig": {
    "credential": {
      "credentialType": "API_TOKEN",
      "apiToken": "YOUR_CIVO_API_TOKEN"
    }
  }
}
```

### 2. Deploy

```bash
cd iac/pulumi
pulumi stack init dev
pulumi config set --path stackInput "$(cat stack-input.json)"
pulumi up
```

### 3. Verify

```bash
kubectl get nodes
pulumi stack output node_pool_id
```

### 4. Scale

Update `nodeCount` or `maxNodes` and run:

```bash
pulumi up
```

## Module Structure

```
iac/pulumi/
├── main.go
├── Pulumi.yaml
├── Makefile
├── debug.sh
├── README.md        # This file
├── overview.md
└── module/
    ├── main.go
    ├── locals.go
    ├── outputs.go
    └── node_pool.go
```

## Stack Outputs

| Output | Description |
|--------|-------------|
| `node_pool_id` | Civo node pool UUID |

## Common Operations

### Add nodes

```json
{"spec": {"nodeCount": 5}}  // Increase from 3 to 5
```

```bash
pulumi up
```

### Change node size

```json
{"spec": {"size": "g4s.kube.large"}}  // Upgrade to larger
```

```bash
pulumi up  # Replaces nodes (downtime)
```

**Warning:** Changing size recreates nodes.

### Enable autoscaling

```json
{
  "spec": {
    "autoScale": true,
    "minNodes": 2,
    "maxNodes": 10
  }
}
```

```bash
pulumi up
```

## Integration with Planton

CLI handles Pulumi automatically:

```bash
planton apply -f nodepool.yaml
planton outputs civokubernetesnodepools/workers
planton delete civokubernetesnodepools/workers
```

## Troubleshooting

### Node pool creation fails

Check cluster exists:

```bash
civo kubernetes list
```

Verify network has capacity:

```bash
civo network show <network-id>
```

### Autoscaling not working

Verify cluster autoscaler installed:

```bash
kubectl get pods -n kube-system | grep autoscaler
```

Check pod resource requests:

```bash
kubectl describe pod <pending-pod>
```

## References

- [Pulumi Civo Provider](https://www.pulumi.com/registry/packages/civo/)
- [Civo API](https://www.civo.com/api/kubernetes)
- [Architecture Overview](overview.md)

## Support

- Issues: [GitHub](https://github.com/plantonhq/planton/issues)
- Civo: support@civo.com

