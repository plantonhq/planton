# DigitalOcean Firewall - Pulumi Module

This Pulumi module provisions and manages stateful, network-edge firewalls for DigitalOcean Droplets using the `pulumi-digitalocean` provider.

## Overview

The module implements OpenMCF's `DigitalOceanFirewall` protobuf spec, providing a type-safe interface for:

- **Inbound Rules**: Control traffic allowed *to* Droplets
- **Outbound Rules**: Control traffic allowed *from* Droplets
- **Tag-Based Targeting**: Production-standard, scales infinitely
- **Static Droplet IDs**: Dev/testing only, max 10 Droplets
- **Resource-Aware Rules**: Support for Load Balancer UIDs, Kubernetes cluster IDs

## Module Structure

```
iac/pulumi/
├── main.go              # Pulumi entrypoint (calls module.Resources)
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build and deployment helpers
├── debug.sh             # Debug script for local testing
├── module/
│   ├── main.go          # Module entry point (Resources function)
│   ├── locals.go        # Local variables and context initialization
│   ├── firewall.go      # Firewall resource creation logic
│   └── outputs.go       # Output constant definitions
└── README.md            # This file
```

## Prerequisites

- **Pulumi CLI**: Install from [pulumi.com/docs/install](https://www.pulumi.com/docs/install/)
- **Go**: Version 1.21 or later
- **DigitalOcean API Token**: Set via environment variable `DIGITALOCEAN_TOKEN` or Pulumi config

## Quick Start

### 1. Set Up Environment

```bash
# Navigate to the Pulumi module directory
cd apis/org/openmcf/provider/digitalocean/digitaloceanfirewall/v1/iac/pulumi

# Set your DigitalOcean API token
export DIGITALOCEAN_TOKEN="your-digitalocean-api-token"
```

### 2. Configure Stack Input

Create a `stack-input.yaml` file with your firewall specification:

**Production Web Tier Firewall:**
```yaml
provider_config:
  credential_id: "digitalocean-prod-credential"

target:
  spec:
    name: prod-web-firewall
    tags:
      - prod-web-tier
    inbound_rules:
      - protocol: tcp
        port_range: "443"
        source_load_balancer_uids:
          - "lb-abc123"
      - protocol: tcp
        port_range: "22"
        source_addresses:
          - "203.0.113.10/32"
    outbound_rules:
      - protocol: tcp
        port_range: "5432"
        destination_tags:
          - prod-db-tier
      - protocol: tcp
        port_range: "443"
        destination_addresses:
          - "0.0.0.0/0"
          - "::/0"
      - protocol: udp
        port_range: "53"
        destination_addresses:
          - "0.0.0.0/0"
```

**Production Database Tier Firewall:**
```yaml
provider_config:
  credential_id: "digitalocean-prod-credential"

target:
  spec:
    name: prod-db-firewall
    tags:
      - prod-db-tier
    inbound_rules:
      - protocol: tcp
        port_range: "5432"
        source_tags:
          - prod-web-tier
      - protocol: tcp
        port_range: "22"
        source_addresses:
          - "203.0.113.10/32"
    outbound_rules:
      - protocol: tcp
        port_range: "443"
        destination_addresses:
          - "91.189.88.0/21"  # Ubuntu repos
      - protocol: udp
        port_range: "53"
        destination_addresses:
          - "1.1.1.1/32"  # Cloudflare DNS
```

### 3. Deploy

```bash
# Initialize Pulumi stack
pulumi stack init dev

# Preview changes
pulumi preview

# Deploy
pulumi up
```

### 4. Retrieve Outputs

```bash
# Get firewall ID
pulumi stack output firewall_id

# Get all outputs
pulumi stack output
```

## Stack Input Schema

### `provider_config` (Required)

DigitalOcean provider configuration.

```yaml
provider_config:
  credential_id: string  # ID of DigitalOcean credential in Planton
```

### `target.spec` (Required)

Firewall specification matching the protobuf `DigitalOceanFirewallSpec`.

```yaml
target:
  spec:
    name: string  # Firewall name (1-255 chars)

    # Inbound rules (traffic allowed *to* Droplets)
    inbound_rules:
      - protocol: string  # "tcp", "udp", or "icmp"
        port_range: string  # "22", "8000-9000", "1-65535"
        source_addresses: list[string]  # CIDR blocks
        source_droplet_ids: list[int]  # Droplet IDs
        source_tags: list[string]  # Droplet tags
        source_kubernetes_ids: list[string]  # K8s cluster IDs
        source_load_balancer_uids: list[string]  # LB UIDs

    # Outbound rules (traffic allowed *from* Droplets)
    outbound_rules:
      - protocol: string
        port_range: string
        destination_addresses: list[string]
        destination_droplet_ids: list[int]
        destination_tags: list[string]
        destination_kubernetes_ids: list[string]
        destination_load_balancer_uids: list[string]

    # Target Droplets (production: use tags, dev: use droplet_ids)
    tags: list[string]  # Production standard (unlimited, auto-scales)
    droplet_ids: list[int]  # Dev/testing only (max 10)
```

## Outputs

The module exports the following outputs via `ctx.Export()`:

| Output | Description |
|--------|-------------|
| `firewall_id` | DigitalOcean firewall ID |

Access outputs via:
```bash
pulumi stack output firewall_id
```

## Production Best Practices

### 1. Tag-Based Targeting (Production Standard)

**✅ Do:**
```yaml
spec:
  tags:
    - prod-web-tier
    - all-instances
  # No droplet_ids
```

**❌ Don't:**
```yaml
spec:
  droplet_ids:  # Anti-pattern for production
    - 123456
    - 234567
```

### 2. Never Expose Management Ports

**✅ Do:**
```yaml
inbound_rules:
  - protocol: tcp
    port_range: "22"
    source_addresses:
      - "203.0.113.10/32"  # Office bastion only
```

**❌ Don't:**
```yaml
inbound_rules:
  - protocol: tcp
    port_range: "22"
    source_addresses:
      - "0.0.0.0/0"  # NEVER in production!
```

### 3. Use Load Balancer UIDs for Public Services

**✅ Do:**
```yaml
inbound_rules:
  - protocol: tcp
    port_range: "443"
    source_load_balancer_uids:
      - "lb-abc123"
```

**❌ Don't:**
```yaml
inbound_rules:
  - protocol: tcp
    port_range: "443"
    source_addresses:
      - "0.0.0.0/0"  # Bypasses LB
```

## Multi-Tier Architecture Example

### Web Tier
```yaml
spec:
  name: web-tier-firewall
  tags:
    - web-tier
  inbound_rules:
    - protocol: tcp
      port_range: "443"
      source_load_balancer_uids:
        - "lb-prod"
  outbound_rules:
    - protocol: tcp
      port_range: "5432"
      destination_tags:
        - db-tier
```

### Database Tier
```yaml
spec:
  name: db-tier-firewall
  tags:
    - db-tier
  inbound_rules:
    - protocol: tcp
      port_range: "5432"
      source_tags:
        - web-tier
  outbound_rules:
    - protocol: tcp
      port_range: "443"
      destination_addresses:
        - "91.189.88.0/21"
```

## The "Double Firewall" Reality

DigitalOcean Cloud Firewalls operate at the **network edge**. Host-based firewalls (`ufw`) run **inside** the OS. Traffic must pass **both**.

**Traffic Flow:**
```
Internet → Cloud Firewall → Host Firewall → Application
```

**Solution**: Either configure both or disable host firewall:
```bash
sudo ufw disable
```

## Local Development

### Debug Script

Use the provided `debug.sh` script for local testing:

```bash
# Edit debug.sh with your stack input
./debug.sh
```

### Manual Testing

```bash
# Build
go build -o pulumi-main main.go

# Run
export DIGITALOCEAN_TOKEN="your-token"
pulumi up
```

## Troubleshooting

### Issue: Connection Times Out Despite Firewall Rule

**Cause**: Host firewall (`ufw`) inside Droplet is blocking traffic.

**Solution**: Check host firewall:
```bash
sudo ufw status verbose
```

### Issue: Auto-Scaling Creates Unprotected Droplets

**Cause**: Using static `droplet_ids` instead of tags.

**Solution**: Switch to tag-based targeting:
```yaml
spec:
  tags:
    - web-tier
  # Remove droplet_ids
```

## Pulumi Commands Reference

```bash
# Initialize stack
pulumi stack init <stack-name>

# Preview changes
pulumi preview

# Deploy
pulumi up

# View outputs
pulumi stack output

# Destroy resources
pulumi destroy

# View stack state
pulumi stack

# Export stack state
pulumi stack export
```

## Further Reading

- **Component Overview**: See [../../README.md](../../README.md)
- **Comprehensive Guide**: See [../../docs/README.md](../../docs/README.md)
- **Examples**: See [../../examples.md](../../examples.md)
- **Pulumi DigitalOcean Provider**: [Registry Docs](https://www.pulumi.com/registry/packages/digitalocean/)

## Support

For module-specific issues:
1. Enable debug logging: `export PULUMI_DEBUG_COMMANDS=true`
2. Check Pulumi logs: `pulumi up --logtostderr -v=9`
3. Validate DigitalOcean API connectivity:
   ```bash
   curl -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" https://api.digitalocean.com/v2/account
   ```

For component issues, see the main [README.md](../../README.md).

