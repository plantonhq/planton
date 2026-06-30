# DigitalOcean VPC - Pulumi Module

This Pulumi module provisions and manages DigitalOcean Virtual Private Cloud (VPC) networks based on the Planton DigitalOceanVpc specification.

## Overview

The Pulumi implementation provides a production-ready infrastructure-as-code solution for deploying DigitalOcean VPCs with:
- Type-safe Go implementation
- Built-in secret management
- Multi-environment stack support
- 80/20 principle: auto-generated CIDR support
- Comprehensive error handling

## Prerequisites

1. **Pulumi CLI**: Version 3.x or higher
2. **Go**: Version 1.21 or higher
3. **DigitalOcean Account**: Active account with API access
4. **DigitalOcean API Token**: Personal access token with read/write permissions

## Installation

```bash
# macOS
brew install pulumi

# Linux
curl -fsSL https://get.pulumi.com | sh

# Verify
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

## Usage

### Project Structure

```
iac/pulumi/
├── main.go              # Pulumi entrypoint
├── Pulumi.yaml         # Project configuration
├── Makefile            # Helper commands
├── debug.sh            # Debug script
├── module/             # VPC implementation
│   ├── main.go         # Module entrypoint
│   ├── vpc.go          # VPC resource
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
pulumi stack output vpc_id
```

## Examples

### Example 1: Auto-Generated CIDR (80% Use Case)

**Manifest:**
```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanVpc
metadata:
  name: dev-vpc
spec:
  region: nyc3
  # ip_range_cidr omitted - auto-generation
```

**Deploy:**
```bash
pulumi up

# Get auto-generated IP range
pulumi stack output ip_range
# Example: 10.116.0.0/20
```

### Example 2: Explicit CIDR (20% Use Case)

**Manifest:**
```yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanVpc
metadata:
  name: prod-vpc
spec:
  description: "Production VPC"
  region: nyc1
  ipRangeCidr: "10.101.0.0/16"
```

**Deploy:**
```bash
pulumi up

# Verify CIDR
pulumi stack output ip_range
# Output: 10.101.0.0/16 (exactly as specified)
```

## Stack Outputs

| Output | Description | Sensitive |
|--------|-------------|-----------|
| `vpc_id` | VPC UUID | No |

**Note:** Additional outputs (URN, IP range, is_default) are available via DigitalOcean API but not currently exported by this module.

### Accessing Outputs

```bash
# Get VPC ID
pulumi stack output vpc_id

# Use in another stack
pulumi stack output vpc_id --stack dev-vpc > vpc-id.txt
```

## Multi-Environment Management

Create separate stacks for each environment:

```bash
# Development
pulumi stack init dev
pulumi config set vpc-name dev-vpc
pulumi config set region nyc3
pulumi up

# Staging
pulumi stack init staging
pulumi config set vpc-name staging-vpc
pulumi config set region sfo3
pulumi up

# Production
pulumi stack init prod
pulumi config set vpc-name prod-vpc
pulumi config set region nyc1
pulumi up
```

## Common Operations

### Get VPC Information

```bash
# Get VPC ID
pulumi stack output vpc_id

# Get all outputs
pulumi stack output

# Export to JSON
pulumi stack export > vpc-state.json
```

### Update VPC Description

Edit manifest to change description:

```yaml
spec:
  description: "Updated description text"
```

```bash
pulumi up
```

**Note:** IP range is immutable and cannot be changed.

### Use VPC in Other Resources

```bash
# Get VPC ID
VPC_ID=$(pulumi stack output vpc_id)

# Use in cluster manifest
cat > cluster-manifest.yaml << EOF
spec:
  vpc:
    value: "$VPC_ID"
EOF
```

### Stack References (Cross-Stack)

```go
// In another Pulumi program
vpcStack := pulumi.NewStackReference(ctx, "org/dev-vpc/dev", nil)
vpcId := vpcStack.GetOutput(pulumi.String("vpc_id"))

// Use in cluster
cluster := digitalocean.NewKubernetesCluster(ctx, "cluster", &Args{
    VpcUuid: vpcId,
})
```

## State Management

### Pulumi Cloud (Recommended)

```bash
pulumi login
pulumi up
```

**Features:**
- Automatic state locking
- Encrypted secrets
- Team collaboration
- Audit history

### Self-Managed Backend

```bash
# S3
pulumi login s3://my-bucket?region=us-east-1

# Local
pulumi login --local
```

## Secrets Management

```bash
# Set secret configuration
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

# Export stack state
pulumi stack export > debug-state.json
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Deploy VPC

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

1. **Use Auto-Generation for Dev**
   - No planning overhead
   - Fast iteration

2. **Explicit CIDRs for Production**
   - Document in IPAM
   - Plan for growth (/16 recommended)

3. **Separate Stacks Per Environment**
   - dev, staging, prod stacks
   - Isolated state

4. **VPC-First Deployment**
   - Create VPCs before clusters
   - Avoid costly migrations

5. **Non-Overlapping Ranges**
   - Enable future VPC peering
   - Support VPN connectivity

## Troubleshooting

### "CIDR block overlaps"

```bash
# List existing VPCs
doctl vpcs list

# Choose non-overlapping range
```

### "Cannot delete VPC"

```bash
# VPC must be empty
VPC_ID=$(pulumi stack output vpc_id)
doctl vpcs resources get $VPC_ID

# Delete resources first
```

### "Invalid region"

```bash
# List valid regions
doctl vpcs list-regions
```

## Cost Information

**VPCs are free on DigitalOcean.**
- No monthly charges
- Free internal traffic
- Free VPC peering (same datacenter)

## Cleanup

```bash
# Preview destruction
pulumi destroy --preview

# Destroy VPC (must be empty of resources first)
pulumi destroy

# Remove stack
pulumi stack rm dev
```

## Reference

- [Pulumi DigitalOcean Provider](https://www.pulumi.com/registry/packages/digitalocean/)
- [DigitalOcean VPC API](https://docs.digitalocean.com/reference/api/api-reference/#tag/VPCs)
- [Module Architecture](./overview.md)
- [Planton Documentation](https://docs.planton.dev/)

---

**Version:** 1.0.0  
**Last Updated:** 2025-11-16

