# DigitalOcean Kubernetes Cluster - Pulumi Module

This Pulumi module provisions and manages DigitalOcean Kubernetes (DOKS) clusters based on the Planton DigitalOceanKubernetesCluster specification.

## Overview

The Pulumi implementation provides a production-ready infrastructure-as-code solution for deploying DOKS clusters with:
- Type-safe Go implementation
- Built-in secret management
- Multi-environment stack support
- Comprehensive error handling
- Automatic cluster configuration

## Prerequisites

1. **Pulumi CLI**: Version 3.x or higher
2. **Go**: Version 1.21 or higher
3. **DigitalOcean Account**: Active account with API access
4. **DigitalOcean API Token**: Personal access token with read/write permissions
5. **VPC**: Pre-existing VPC in the target region

## Installation

### Install Pulumi

```bash
# macOS
brew install pulumi

# Linux
curl -fsSL https://get.pulumi.com | sh

# Verify installation
pulumi version
```

### Initialize Pulumi Backend

```bash
# Use Pulumi Cloud (recommended)
pulumi login

# Or use local backend
pulumi login --local

# Or use S3
pulumi login s3://my-pulumi-state-bucket
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
# Create a new stack
pulumi stack init dev

# Set configuration values
pulumi config set cluster-name dev-cluster
pulumi config set region nyc1
pulumi config set vpc-uuid your-vpc-uuid
```

## Usage

### Project Structure

```
iac/pulumi/
├── main.go              # Pulumi entrypoint
├── Pulumi.yaml         # Pulumi project configuration
├── Makefile            # Helper commands
├── debug.sh            # Debug script
├── module/             # Cluster implementation
│   ├── main.go         # Module entrypoint
│   ├── cluster.go      # Cluster resource
│   ├── locals.go       # Local variables
│   └── outputs.go      # Stack outputs
├── README.md           # This file
└── overview.md         # Architecture documentation
```

### Deployment

#### 1. Navigate to Pulumi Directory

```bash
cd apis/dev/planton/provider/digitalocean/digitaloceankubernetescluster/v1/iac/pulumi
```

#### 2. Install Dependencies

```bash
go mod download
```

#### 3. Preview Changes

```bash
pulumi preview
```

This shows what resources will be created without making any changes.

#### 4. Deploy

```bash
pulumi up
```

Review the plan and confirm to deploy. Deployment takes 3-5 minutes.

#### 5. Verify Deployment

```bash
# Check stack outputs
pulumi stack output

# Get kubeconfig
pulumi stack output kubeconfig --show-secrets > kubeconfig.yaml
export KUBECONFIG=kubeconfig.yaml

# Verify cluster
kubectl get nodes
```

## Stack Outputs

The module exports the following outputs:

| Output | Description | Sensitive |
|--------|-------------|-----------|
| `cluster_id` | Cluster UUID | No |
| `kubeconfig` | Base64-encoded kubeconfig | Yes |
| `api_server_endpoint` | Kubernetes API URL | No |

### Accessing Outputs

```bash
# List all outputs
pulumi stack output

# Get specific output
pulumi stack output cluster_id

# Get sensitive output (kubeconfig)
pulumi stack output kubeconfig --show-secrets

# Export kubeconfig to file
pulumi stack output kubeconfig --show-secrets > ~/.kube/doks-config
```

## Multi-Environment Management

Pulumi stacks enable managing multiple environments (dev, staging, prod) with a single codebase.

### Create Environment Stacks

```bash
# Development stack
pulumi stack init dev
pulumi config set cluster-name dev-cluster
pulumi config set region nyc1
pulumi up

# Staging stack
pulumi stack init staging
pulumi config set cluster-name staging-cluster
pulumi config set region sfo3
pulumi up

# Production stack
pulumi stack init production
pulumi config set cluster-name prod-cluster
pulumi config set region nyc1
pulumi up
```

### Switch Between Stacks

```bash
# List stacks
pulumi stack ls

# Select stack
pulumi stack select dev

# Get current stack
pulumi stack
```

## Common Operations

### Update Cluster Configuration

Edit your manifest YAML and rerun:

```bash
pulumi up
```

Pulumi will detect changes and update only affected resources.

### Scale Node Pool

Update the `nodeCount` in your manifest:

```yaml
spec:
  defaultNodePool:
    nodeCount: 5  # Changed from 3
```

```bash
pulumi up
```

### Enable Autoscaling

Update manifest to enable autoscaling:

```yaml
spec:
  defaultNodePool:
    autoScale: true
    minNodes: 3
    maxNodes: 10
```

```bash
pulumi up
```

### Upgrade Kubernetes Version

**Note:** The module ignores version changes to prevent drift during auto-upgrades. To manually upgrade:

1. Remove `IgnoreChanges` from `cluster.go`:
   ```go
   // Temporarily remove: pulumi.IgnoreChanges([]string{"version"}),
   ```

2. Update version in manifest:
   ```yaml
   spec:
     kubernetesVersion: "1.30"
   ```

3. Deploy:
   ```bash
   pulumi up
   ```

4. Restore `IgnoreChanges` after upgrade

### Add Control Plane Firewall

Update manifest:

```yaml
spec:
  controlPlaneFirewallAllowedIps:
    - "203.0.113.10/32"
    - "198.51.100.0/24"
```

```bash
pulumi up
```

## State Management

### Pulumi Cloud (Recommended)

Pulumi Cloud provides:
- Automatic state management
- Team collaboration
- State locking
- History and rollback
- Secrets encryption

```bash
# Login to Pulumi Cloud
pulumi login

# Deploy to cloud-backed state
pulumi up
```

### Self-Managed Backend

#### S3 Backend

```bash
# Configure S3 backend
pulumi login s3://my-bucket?region=us-east-1

# Deploy
pulumi up
```

#### Local Backend

```bash
# Use local filesystem
pulumi login --local

# State stored in ~/.pulumi
pulumi up
```

## Secrets Management

Pulumi provides built-in encrypted secrets:

```bash
# Set secret configuration
pulumi config set digitalocean:token --secret YOUR_TOKEN

# Secrets are encrypted in state
pulumi config

# View secret value
pulumi config get digitalocean:token --show-secrets
```

## Debugging

### Enable Debug Output

```bash
# Detailed logging
pulumi up --logtostderr -v=9

# Or use debug script
./debug.sh
```

### View Resource Details

```bash
# Show resource URNs
pulumi stack --show-urns

# Export entire state
pulumi stack export > stack.json
```

### Common Issues

#### 1. VPC Not Found

**Error:** `vpc not found`

**Solution:**
```bash
# List VPCs
doctl vpcs list

# Update manifest with correct VPC UUID
```

#### 2. Version Conflict

**Error:** `kubernetes version not supported`

**Solution:**
```bash
# Check available versions
doctl kubernetes options versions

# Update manifest with supported version
```

#### 3. Pulumi State Conflict

**Error:** `conflict: Another update is in progress`

**Solution:**
```bash
# Cancel conflicting update
pulumi cancel

# Or if update is stuck
pulumi stack export > backup.json
pulumi stack import < backup.json
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Deploy DOKS Cluster

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

### GitLab CI

```yaml
deploy:
  stage: deploy
  image: pulumi/pulumi
  script:
    - pulumi login $PULUMI_ACCESS_TOKEN
    - pulumi stack select production
    - pulumi up --yes
  variables:
    DIGITALOCEAN_TOKEN: $DIGITALOCEAN_TOKEN
  only:
    - main
```

## Cost Optimization

### Development Clusters

```yaml
spec:
  highlyAvailable: false      # Save $40/month
  defaultNodePool:
    size: s-1vcpu-2gb          # ~$12/node/month
    nodeCount: 2               # Minimal nodes
```

**Cost:** ~$24/month

### Production Clusters

```yaml
spec:
  highlyAvailable: true       # $40/month (may be waived)
  defaultNodePool:
    size: s-4vcpu-8gb          # ~$43/node/month
    nodeCount: 5
    autoScale: true
    minNodes: 5
    maxNodes: 10
```

**Cost:** ~$255-$470/month (depending on autoscaling)

## Cleanup

### Destroy Stack

```bash
# Preview destruction
pulumi destroy --preview

# Destroy resources (requires confirmation)
pulumi destroy

# Auto-confirm (use with caution)
pulumi destroy --yes
```

### Remove Stack

```bash
# Destroy resources first
pulumi destroy --yes

# Remove stack from state
pulumi stack rm dev
```

## Development Workflow

### Local Testing

```bash
# Install dependencies
go mod tidy

# Build
go build -o pulumi-main .

# Run tests
go test ./module/...

# Deploy to dev stack
pulumi stack select dev
pulumi up
```

### Module Development

The module is organized into:

- **main.go**: Pulumi program entrypoint, initializes stack
- **module/main.go**: Module entry, orchestrates resource creation
- **module/cluster.go**: Cluster resource implementation
- **module/locals.go**: Local variables and data processing
- **module/outputs.go**: Stack outputs definition

See `overview.md` for detailed architecture documentation.

## Helper Scripts

### Makefile

```bash
# Deploy
make deploy

# Destroy
make destroy

# Preview
make preview

# Get outputs
make outputs
```

### Debug Script

```bash
# Run with debug logging
./debug.sh
```

## Best Practices

1. **Use Pulumi Cloud for Team Collaboration**
   - Automatic state locking
   - Encrypted secrets
   - Audit history

2. **Separate Stacks Per Environment**
   - dev, staging, prod stacks
   - Environment-specific configuration
   - Isolated state

3. **Version Control Configuration**
   - Commit Pulumi.yaml
   - Commit stack config (encrypted secrets are safe)
   - Use `.gitignore` for sensitive local files

4. **Implement RBAC**
   - Use Pulumi Cloud teams
   - Role-based stack access
   - Separate service accounts for CI/CD

5. **Monitor State Size**
   - Large states slow operations
   - Consider splitting into multiple stacks
   - Regular state cleanup

## Troubleshooting

### Cannot Access Cluster

**Symptom:** kubectl commands timeout

**Possible Causes:**
- Control plane firewall enabled without your IP
- Kubeconfig not exported

**Solution:**
```bash
# Get your IP
curl ifconfig.me

# Add to manifest controlPlaneFirewallAllowedIps
# Rerun pulumi up

# Export kubeconfig
pulumi stack output kubeconfig --show-secrets > ~/.kube/config
export KUBECONFIG=~/.kube/config
```

### Autoscaling Not Working

**Symptom:** Nodes don't scale despite pod pressure

**Solution:**
- Verify `autoScale: true` in manifest
- Check `minNodes` and `maxNodes` are set
- Ensure pod resource requests are defined

```bash
# Check autoscaler
kubectl logs -n kube-system -l app=cluster-autoscaler
```

## Reference

- [Pulumi DigitalOcean Provider](https://www.pulumi.com/registry/packages/digitalocean/)
- [Pulumi Go SDK Documentation](https://www.pulumi.com/docs/reference/pkg/go/)
- [DOKS Documentation](https://docs.digitalocean.com/products/kubernetes/)
- [Module Architecture](./overview.md)

## Support

For issues or questions:
- Planton: [GitHub Issues](https://github.com/plantonhq/planton/issues)
- Pulumi Community: [Slack](https://slack.pulumi.com/)
- DigitalOcean Support: [Support Portal](https://www.digitalocean.com/support)

---

**Version:** 1.0.0  
**Last Updated:** 2025-11-16

