# Civo Kubernetes Cluster - Pulumi Module

This directory contains the Pulumi implementation for deploying managed Kubernetes clusters on Civo Cloud. It's designed to be invoked by the Planton CLI, but can also be used standalone for direct Pulumi workflows.

## Overview

The Pulumi module provisions:
- A managed Kubernetes cluster (K3s-based) on Civo
- Default node pool with specified size and node count
- Optional high-availability control plane
- Cluster outputs (ID, kubeconfig, API endpoint)

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) installed (v3.x)
- [Go](https://golang.org/dl/) 1.21 or later
- Civo account and API token
- Existing Civo network (VPC)
- `kubectl` installed for testing

## Quick Start (Standalone Usage)

### 1. Set up stack input

Create a `stack-input.json` file:

```json
{
  "target": {
    "apiVersion": "civo.planton.dev/v1",
    "kind": "CivoKubernetesCluster",
    "metadata": {
      "name": "dev-cluster",
      "id": "cikc-dev-123"
    },
    "spec": {
      "clusterName": "dev-k8s",
      "region": "lon1",
      "kubernetesVersion": "1.29.0+k3s1",
      "network": {
        "value": "network-id-here"
      },
      "autoUpgrade": true,
      "tags": ["environment:dev"],
      "defaultNodePool": {
        "size": "g4s.kube.small",
        "nodeCount": 1
      }
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

### 2. Initialize Pulumi stack

```bash
cd iac/pulumi

pulumi stack init dev
pulumi config set civo:region LON1
```

### 3. Deploy

```bash
pulumi config set --path stackInput --plaintext "$(cat stack-input.json)"
pulumi preview
pulumi up
```

### 4. Get kubeconfig

```bash
# Export kubeconfig from outputs
pulumi stack output kubeconfig --show-secrets > ~/.kube/civo-dev
export KUBECONFIG=~/.kube/civo-dev

# Test access
kubectl get nodes
kubectl get pods --all-namespaces
```

### 5. Destroy

```bash
pulumi destroy
```

**Warning:** This deletes the cluster and all workloads. Backup data first.

## Module Structure

```
iac/pulumi/
├── main.go              # Pulumi entrypoint
├── Pulumi.yaml          # Project configuration
├── Makefile             # Build and test commands
├── debug.sh             # Debug helper script
├── README.md            # This file
├── overview.md          # Architecture documentation
└── module/
    ├── main.go          # Module entry point (Resources function)
    ├── locals.go        # Local variables and initialization
    ├── outputs.go       # Output constants
    └── cluster.go       # Cluster provisioning logic
```

## Key Files

### `module/main.go`

Entry point invoked by Planton CLI:

```go
func Resources(
    ctx *pulumi.Context,
    stackInput *civokubernetesclusterv1.CivoKubernetesClusterStackInput,
) error
```

### `module/cluster.go`

Core provisioning logic:
- Creates `civo.KubernetesCluster` resource
- Configures default node pool
- Sets optional features (HA, auto-upgrade)
- Exports outputs (kubeconfig, API endpoint)

### `module/outputs.go`

Output constants:
- `cluster_id`
- `cluster_name`
- `kubeconfig`
- `api_endpoint`
- `master_ip`

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `cluster_id` | string | Civo cluster UUID |
| `cluster_name` | string | Cluster name |
| `kubeconfig` | string | Base64-encoded kubeconfig |
| `api_endpoint` | string | Kubernetes API URL |
| `master_ip` | string | Control plane IP |

Access outputs:

```bash
pulumi stack output cluster_id
pulumi stack output kubeconfig --show-secrets | base64 -d > ~/.kube/config
```

## Common Operations

### Scale Cluster

Update node count in stack input:

```json
{
  "spec": {
    "defaultNodePool": {
      "nodeCount": 5  # Scale from 3 to 5
    }
  }
}
```

```bash
pulumi up  # Adds 2 nodes
```

### Upgrade Kubernetes Version

```json
{
  "spec": {
    "kubernetesVersion": "1.30.0+k3s1"  # New version
  }
}
```

```bash
pulumi preview  # Review upgrade plan
pulumi up  # Apply upgrade
```

**Warning:** Kubernetes upgrades can cause downtime. Test in staging first.

### Add Tags

```json
{
  "spec": {
    "tags": ["environment:prod", "new-tag"]
  }
}
```

```bash
pulumi up  # Updates cluster metadata
```

## Integration with Planton

When using the Planton CLI:

1. CLI reads your `CivoKubernetesCluster` YAML
2. Converts to `CivoKubernetesClusterStackInput` protobuf
3. Invokes `module.Resources` function
4. Manages Pulumi stacks automatically
5. Returns outputs in standardized format

CLI workflow:

```bash
planton apply -f cluster.yaml
planton outputs civokubernetesclusters/dev-cluster
planton delete civokubernetesclusters/dev-cluster
```

## Development

### Build and test

```bash
make build
make test
make fmt
make lint
```

### Debug

```bash
export CIVO_TOKEN="your-token"
export DEBUG_STACK_INPUT="$(cat test-input.json)"
./debug.sh
```

## Troubleshooting

### Cluster creation timeout

**Symptom:** `pulumi up` hangs for > 5 minutes.

**Solutions:**
- Check Civo dashboard for cluster status
- Verify network has available IPs
- Check region capacity
- Review Civo quotas

### Kubeconfig not working

**Symptom:** `kubectl get nodes` fails with auth error.

**Solutions:**

```bash
# Re-export kubeconfig
pulumi stack output kubeconfig --show-secrets | base64 -d > ~/.kube/config

# Verify API endpoint is reachable
curl -k $(pulumi stack output api_endpoint)

# Check firewall allows port 6443
```

### State drift

If cluster modified manually:

```bash
pulumi refresh  # Sync state
pulumi preview  # Review differences
pulumi up       # Converge to desired state
```

## Best Practices

1. **Remote state** - Use Pulumi Cloud or S3 backend
2. **Separate stacks** - One per environment (dev/staging/prod)
3. **Protect production** - `pulumi stack change-secrets-provider`
4. **Version control** - Track stack files in Git
5. **Test upgrades** - Always test in lower environments first

## References

- [Pulumi Civo Provider](https://www.pulumi.com/registry/packages/civo/)
- [Civo Kubernetes API](https://www.civo.com/api/kubernetes)
- [K3s Documentation](https://docs.k3s.io/)
- [Architecture Overview](overview.md)

## Support

- Issues: [GitHub Issues](https://github.com/plantonhq/planton/issues)
- Pulumi Community: [Pulumi Slack](https://slack.pulumi.com/)
- Civo Support: support@civo.com

