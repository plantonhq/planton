# DigitalOcean Load Balancer - Pulumi Module

This Pulumi module deploys DigitalOcean Regional Load Balancers from Planton's protobuf-defined manifests.

## Overview

The module translates `DigitalOceanLoadBalancerSpec` manifests into DigitalOcean Load Balancer resources using Pulumi's DigitalOcean provider. It handles:

- Regional load balancer provisioning
- VPC network placement
- Tag-based or ID-based Droplet targeting
- Forwarding rules with SSL termination support
- Health check configuration
- Sticky sessions (optional)

## Architecture

See [overview.md](overview.md) for detailed module architecture and design decisions.

## Prerequisites

### 1. DigitalOcean Account and API Token

```bash
export DIGITALOCEAN_TOKEN="your-api-token"
```

Get your token from: https://cloud.digitalocean.com/account/api/tokens

### 2. Pulumi CLI

```bash
# macOS
brew install pulumi

# Linux
curl -fsSL https://get.pulumi.com | sh

# Windows
choco install pulumi
```

### 3. Go Runtime

The Pulumi program is written in Go. Ensure Go 1.21+ is installed:

```bash
go version
```

### 4. VPC and Droplets

- A DigitalOcean VPC must exist in the target region
- Droplets must be created and tagged (for tag-based targeting) or have known IDs

## Stack Configuration

### Required Configuration

Create a `Pulumi.<stack>.yaml` file:

```yaml
config:
  digitalocean:token: ${DIGITALOCEAN_TOKEN}
```

**Security Note:** Use environment variables or secret management for tokens. Never commit tokens to version control.

### Stack Input

The module expects a `DigitalOceanLoadBalancerStackInput` JSON file:

```json
{
  "target": {
    "kind": "DigitalOceanLoadBalancer",
    "metadata": {
      "name": "prod-web-lb"
    },
    "spec": {
      "load_balancer_name": "prod-web-lb",
      "region": "sfo3",
      "vpc": {
        "value": "vpc-123456"
      },
      "forwarding_rules": [
        {
          "entry_port": 443,
          "entry_protocol": "https",
          "target_port": 80,
          "target_protocol": "http",
          "certificate_name": "my-cert"
        }
      ],
      "health_check": {
        "port": 80,
        "protocol": "http",
        "path": "/healthz",
        "check_interval_sec": 10
      },
      "droplet_tag": "web-prod",
      "enable_sticky_sessions": true
    }
  },
  "provider_config": {
    "digitalocean_token": "${DIGITALOCEAN_TOKEN}"
  }
}
```

## Deployment Workflow

### Option 1: Using Planton CLI (Recommended)

```bash
# Create manifest
cat <<EOF > lb-manifest.yaml
apiVersion: digital-ocean.planton.dev/v1
kind: DigitalOceanLoadBalancer
metadata:
  name: prod-web-lb
spec:
  load_balancer_name: prod-web-lb
  region: sfo3
  vpc:
    value: "vpc-123456"
  forwarding_rules:
    - entry_port: 443
      entry_protocol: https
      target_port: 80
      target_protocol: http
      certificate_name: "my-cert"
  health_check:
    port: 80
    protocol: http
    path: "/healthz"
  droplet_tag: "web-prod"
EOF

# Deploy
planton pulumi up --manifest lb-manifest.yaml
```

### Option 2: Direct Pulumi Usage

```bash
# Initialize Pulumi stack
cd iac/pulumi
pulumi stack init prod

# Configure DigitalOcean token
pulumi config set digitalocean:token $DIGITALOCEAN_TOKEN --secret

# Set stack input (provide path to JSON file)
export STACK_INPUT_FILE="path/to/stack-input.json"

# Preview changes
pulumi preview

# Deploy
pulumi up
```

### Option 3: Using Makefile

```bash
cd iac/pulumi

# Deploy
make up STACK=prod

# Preview
make preview STACK=prod

# Destroy
make destroy STACK=prod
```

## Module Structure

```
iac/pulumi/
├── main.go                 # Pulumi program entrypoint
├── Pulumi.yaml             # Project configuration
├── Makefile                # Deployment automation
├── debug.sh                # Debugging helper script
├── README.md               # This file
├── overview.md             # Architecture documentation
└── module/
    ├── main.go             # Module orchestration
    ├── locals.go           # Local variables and labels
    ├── load_balancer.go    # Load balancer resource logic
    └── outputs.go          # Stack output definitions
```

## Outputs

The module exports the following stack outputs:

```go
// Stack Outputs
outputs := {
  "load_balancer_id": "lb-abc123",        // Load balancer UUID
  "ip": "203.0.113.42",                   // Public IP address
  "dns_name": "prod-web-lb"               // Name (DigitalOcean LBs don't have DNS)
}
```

Access outputs:

```bash
pulumi stack output load_balancer_id
pulumi stack output ip
```

## Environment Variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| `DIGITALOCEAN_TOKEN` | DigitalOcean API token | Yes | `dop_v1_abc...` |
| `STACK_INPUT_FILE` | Path to stack input JSON | Yes | `./stack-input.json` |
| `PULUMI_BACKEND_URL` | Pulumi state backend | No | `s3://my-pulumi-state` |

## Advanced Configuration

### Using DigitalOcean Spaces for State

```bash
# Configure Spaces backend
pulumi login s3://my-pulumi-state?endpoint=nyc3.digitaloceanspaces.com

# Deploy
pulumi up
```

### Tag-Based Targeting (Recommended)

```yaml
spec:
  droplet_tag: "web-prod"
```

**Benefits:**
- Automatic backend discovery
- Works with autoscaling
- Enables blue-green deployments

### ID-Based Targeting (Static)

```yaml
spec:
  droplet_ids:
    - value: "386734086"
    - value: "386734087"
```

**Use only for:**
- Testing
- Manual setups
- Static infrastructure

### SSL Termination

```yaml
spec:
  forwarding_rules:
    - entry_port: 443
      entry_protocol: https
      target_port: 80
      target_protocol: http
      certificate_name: "my-le-cert"  # Use name, not ID!
```

**Certificate Setup:**
1. Upload certificate to DigitalOcean or use Let's Encrypt
2. Copy the certificate **name** (not ID)
3. Use in `certificate_name` field

**Why name instead of ID?**
- Certificate IDs change when Let's Encrypt auto-renews
- Names remain stable across renewals
- Prevents Pulumi state drift

## Troubleshooting

### "No healthy backends" Error

**Symptom:** Load balancer returns 503 Service Unavailable

**Causes:**
1. Health check misconfiguration (wrong port, path, or protocol)
2. Backend application not responding to health checks
3. Droplets not tagged correctly (if using tag-based targeting)
4. Droplets in different VPC or region

**Solutions:**

```bash
# Check Droplet tags
doctl compute droplet list --format ID,Name,Tags

# Test health check endpoint
ssh droplet-ip
curl http://localhost:80/healthz

# Check load balancer status
doctl compute load-balancer get lb-abc123
```

### "Certificate not found" Error

**Symptom:** Pulumi fails with "certificate does not exist"

**Cause:** Certificate name is incorrect or doesn't exist

**Solution:**

```bash
# List certificates
doctl compute certificate list

# Use the NAME column value, not ID
```

### State Drift with Droplet IDs

**Symptom:** Pulumi constantly wants to update `droplet_ids`

**Cause:** Tag-based targeting dynamically updates backend pool

**Solution:** This is expected behavior. Use tag-based targeting for production.

### Debug Mode

Enable debug output:

```bash
# Use debug script
./debug.sh

# Or manually
export PULUMI_DEBUG_COMMANDS=true
export PULUMI_DEBUG_GRPC=true
pulumi up --logtostderr --logflow -v=9 2>&1 | tee pulumi-debug.log
```

## Validation

### Pre-Deployment Validation

```bash
# Validate manifest
planton validate --manifest lb-manifest.yaml

# Preview infrastructure changes
pulumi preview
```

### Post-Deployment Testing

```bash
# Get load balancer IP
LB_IP=$(pulumi stack output ip)

# Test HTTP endpoint
curl http://$LB_IP

# Test HTTPS endpoint
curl https://$LB_IP

# Check health
curl http://$LB_IP/healthz
```

## Cleanup

### Destroy Load Balancer

```bash
# Using Planton CLI
planton pulumi destroy --manifest lb-manifest.yaml

# Using Pulumi directly
pulumi destroy
```

### Remove Stack

```bash
pulumi stack rm prod
```

## Production Checklist

Before deploying to production:

- [ ] VPC is created in target region
- [ ] Droplets are tagged correctly
- [ ] Certificate is uploaded (for HTTPS)
- [ ] Health check endpoint exists on backends
- [ ] Cloud Firewalls block direct Droplet access
- [ ] Sticky sessions configured (if needed)
- [ ] Monitoring and alerting configured
- [ ] DNS A record points to load balancer IP

## Next Steps

- Review [overview.md](overview.md) for module architecture
- Check [examples.md](../../examples.md) for usage patterns
- Read [docs/README.md](../../docs/README.md) for best practices
- See [hack/manifest.yaml](../../hack/manifest.yaml) for test manifest

## Support

For issues or questions:
- Check [troubleshooting section](#troubleshooting)
- Review [DigitalOcean Load Balancer docs](https://docs.digitalocean.com/products/networking/load-balancers/)
- Open an issue in the Planton repository

