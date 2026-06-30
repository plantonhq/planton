# DigitalOcean Kubernetes Node Pool - Pulumi Module

This Pulumi module provisions and manages node pools for DigitalOcean Kubernetes (DOKS) clusters based on the Planton DigitalOceanKubernetesNodePool specification.

## Overview

The Pulumi implementation provides a production-ready infrastructure-as-code solution for deploying DOKS node pools with:
- Type-safe Go implementation
- Built-in secret management
- Multi-environment stack support
- Comprehensive error handling
- Automatic configuration merging (metadata + spec labels)

## Prerequisites

1. **Pulumi CLI**: Version 3.x or higher
2. **Go**: Version 1.21 or higher
3. **DigitalOcean Account**: Active account with API access
4. **DigitalOcean API Token**: Personal access token with read/write permissions
5. **Existing DOKS Cluster**: Cluster to add node pool to

## Installation

```bash
# macOS
brew install pulumi

# Linux
curl -fsSL https://get.pulumi.com | sh

# Verify
pulumi version
```

## Configuration

### Set DigitalOcean Token

```bash
# Set as Pulumi secret (recommended)
pulumi config set digitalocean:token --secret YOUR_DO_TOKEN

# Or use environment variable
export DIGITALOCEAN_TOKEN="your-do-api-token"
```

### Configure Stack

```bash
# Create stack
pulumi stack init dev

# Set config values
pulumi config set pool-name dev-workers
pulumi config set cluster-id your-cluster-id
```

## Usage

### Project Structure

```
iac/pulumi/
├── main.go              # Pulumi entrypoint
├── Pulumi.yaml         # Project configuration
├── Makefile            # Helper commands
├── debug.sh            # Debug script
├── module/             # Node pool implementation
│   ├── main.go         # Module entrypoint
│   ├── node_pool.go    # Node pool resource
│   ├── locals.go       # Local variables
│   └── outputs.go      # Stack outputs
├── README.md           # This file
└── overview.md         # Architecture docs
```

### Deployment

```bash
# Navigate to directory
cd iac/pulumi

# Install dependencies
go mod download

# Preview changes
pulumi preview

# Deploy
pulumi up

# Verify
pulumi stack output node_pool_id
```

## Stack Outputs

| Output | Description | Sensitive |
|--------|-------------|-----------|
| `node_pool_id` | Node pool UUID | No |

### Accessing Outputs

```bash
# List all outputs
pulumi stack output

# Get specific output
pulumi stack output node_pool_id
```

## Multi-Environment Management

```bash
# Development
pulumi stack init dev
pulumi config set pool-name dev-workers
pulumi up

# Production
pulumi stack init prod
pulumi config set pool-name prod-workers
pulumi up

# Switch stacks
pulumi stack select dev
```

## Common Operations

### Update Node Count

Edit manifest and apply:

```bash
pulumi up
```

### Scale Pool

Update `nodeCount` in manifest:

```yaml
spec:
  nodeCount: 5  # Changed from 3
```

```bash
pulumi up
```

### Enable Autoscaling

```yaml
spec:
  autoScale: true
  minNodes: 3
  maxNodes: 10
```

### Add Labels

```yaml
spec:
  labels:
    workload: web
    tier: frontend
```

### Add Taints

```yaml
spec:
  taints:
    - key: dedicated
      value: backend
      effect: NoSchedule
```

## State Management

### Pulumi Cloud (Recommended)

```bash
# Login
pulumi login

# Deploy
pulumi up
```

### Self-Managed Backend

```bash
# S3
pulumi login s3://my-bucket?region=us-east-1

# Local
pulumi login --local
```

## Secrets Management

```bash
# Set secret
pulumi config set digitalocean:token --secret YOUR_TOKEN

# View secrets
pulumi config get digitalocean:token --show-secrets
```

## Debugging

```bash
# Enable debug output
pulumi up --logtostderr -v=9

# Use debug script
./debug.sh
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Deploy Node Pool

on:
  push:
    branches: [main]

jobs:
  pulumi:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - uses: pulumi/actions@v4
        with:
          command: up
          stack-name: production
        env:
          PULUMI_ACCESS_TOKEN: ${{ secrets.PULUMI_ACCESS_TOKEN }}
          DIGITALOCEAN_TOKEN: ${{ secrets.DIGITALOCEAN_TOKEN }}
```

## Best Practices

1. **Use Pulumi Cloud** - Automatic state locking and encryption
2. **Separate Stacks** - dev, staging, prod
3. **Version Control** - Commit Pulumi.yaml and stack configs
4. **Secret Management** - Use Pulumi config secrets, not plain text
5. **Resource Naming** - Use consistent conventions

## Troubleshooting

### "Cluster not found"

```bash
# Verify cluster exists
doctl kubernetes cluster list

# Update cluster ID in manifest
```

### "Invalid Droplet size"

```bash
# Check available sizes
doctl kubernetes options sizes

# Update spec.size
```

### "Taint validation failed"

**Effect must be one of:**
- NoSchedule
- PreferNoSchedule
- NoExecute

### State Conflict

```bash
# Cancel conflicting update
pulumi cancel

# Or export/import state
pulumi stack export > backup.json
pulumi stack import < backup.json
```

## Cost Optimization

**Development:**
```yaml
spec:
  size: s-1vcpu-2gb  # ~$12/node
  nodeCount: 2
```
**Cost:** ~$24/month

**Production:**
```yaml
spec:
  size: s-4vcpu-8gb  # ~$43/node
  nodeCount: 5
  autoScale: true
  minNodes: 3
  maxNodes: 10
```
**Cost:** ~$129-430/month (autoscaling)

## Cleanup

```bash
# Preview destruction
pulumi destroy --preview

# Destroy
pulumi destroy

# Remove stack
pulumi stack rm dev
```

## Module Features

- **Label Merging**: Automatically merges metadata labels with spec labels
- **Taint Support**: Full support for Kubernetes taints
- **Autoscaling**: Built-in autoscaling configuration
- **Error Handling**: Comprehensive error wrapping and reporting
- **Type Safety**: Go type checking at compile time

## Reference

- [Pulumi DigitalOcean Provider](https://www.pulumi.com/registry/packages/digitalocean/)
- [Pulumi Go SDK](https://www.pulumi.com/docs/reference/pkg/go/)
- [DOKS Documentation](https://docs.digitalocean.com/products/kubernetes/)
- [Module Architecture](./overview.md)

---

**Version:** 1.0.0  
**Last Updated:** 2025-11-16

