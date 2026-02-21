# DigitalOcean Firewall

Manage cloud firewalls on DigitalOcean using a type-safe, protobuf-defined API with OpenMCF.

## Overview

**DigitalOceanFirewall** enables you to provision and manage stateful, network-edge firewalls that protect Droplets, Kubernetes clusters, and Load Balancers on DigitalOcean. Firewalls enforce a default-deny security model, blocking all traffic except explicitly allowed inbound and outbound rules.

## Why Use This Component?

- **Type-Safe Configuration**: Protobuf-based API with compile-time validation prevents invalid firewall configurations
- **Tag-Based Targeting**: Scale firewalls to thousands of Droplets using tag-based targeting (production standard)
- **Network-Edge Enforcement**: Traffic is blocked before it reaches your VMs, preventing resource-exhaustion attacks
- **Stateful Rules**: Define only initiating traffic rules; return traffic is automatically allowed
- **Free**: Cloud Firewalls are included with DigitalOcean at no additional cost

## Quick Start

### Production Web Tier Firewall

For production web servers behind a Load Balancer with restricted SSH access:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFirewall
metadata:
  name: prod-web-firewall
spec:
  name: prod-web-firewall
  tags:
    - prod-web-tier
  inbound_rules:
    # HTTPS from Load Balancer only
    - protocol: tcp
      port_range: "443"
      source_load_balancer_uids:
        - "lb-abc123"
    # HTTP for redirect to HTTPS
    - protocol: tcp
      port_range: "80"
      source_load_balancer_uids:
        - "lb-abc123"
    # SSH from office bastion only
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "203.0.113.10/32"
  outbound_rules:
    # Database access
    - protocol: tcp
      port_range: "5432"
      destination_tags:
        - prod-db-tier
    # External APIs and OS updates
    - protocol: tcp
      port_range: "443"
      destination_addresses:
        - "0.0.0.0/0"
        - "::/0"
    # DNS
    - protocol: udp
      port_range: "53"
      destination_addresses:
        - "0.0.0.0/0"
        - "::/0"
```

### Production Database Tier Firewall

For PostgreSQL databases accessible only by web tier and administrators:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFirewall
metadata:
  name: prod-db-firewall
spec:
  name: prod-db-firewall
  tags:
    - prod-db-tier
  inbound_rules:
    # PostgreSQL from web tier only
    - protocol: tcp
      port_range: "5432"
      source_tags:
        - prod-web-tier
    # SSH from office bastion
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "203.0.113.10/32"
  outbound_rules:
    # OS updates (locked to Ubuntu repos)
    - protocol: tcp
      port_range: "443"
      destination_addresses:
        - "91.189.88.0/21"
    # DNS (specific resolver)
    - protocol: udp
      port_range: "53"
      destination_addresses:
        - "1.1.1.1/32"
```

## Key Features

### Tag-Based Targeting (Production Standard)
- ✅ **Scalable**: No limit on number of Droplets per tag (vs. 10-Droplet limit for static IDs)
- ✅ **Auto-Scaling Friendly**: New Droplets with matching tags are automatically protected
- ✅ **Composable**: Apply multiple firewalls to a Droplet via multiple tags
- ✅ **Self-Documenting**: Tags like `web-tier`, `db-tier`, `prod` make configs readable

### Static Droplet IDs (Dev/Testing Only)
- ⚠️ **10-Droplet Maximum**: Static IDs hit hard limit at 10 resources
- ⚠️ **Manual Management**: Auto-scaling requires manual firewall updates
- ✅ **Acceptable for Dev**: Fine for manually-created development environments

### Rule Types and Capabilities
- **Protocols**: TCP, UDP, ICMP
- **Port Ranges**: Single ports (`"22"`), ranges (`"8000-9000"`), or all ports (`"1-65535"`)
- **IP-Based Sources/Destinations**: CIDR blocks (IPv4 and IPv6)
- **Resource-Based Sources/Destinations**: Droplet IDs, tags, Load Balancer UIDs, Kubernetes cluster IDs

## Configuration Reference

### Core Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Firewall identifier (1-255 chars, unique per account) |
| `tags` | list(string) | No | Droplet tags to apply firewall to (max 5 tags) |
| `droplet_ids` | list(int64) | No | Droplet IDs to apply firewall to (max 10) |
| `inbound_rules` | list | No | Rules allowing traffic *to* Droplets |
| `outbound_rules` | list | No | Rules allowing traffic *from* Droplets |

### Inbound Rule Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `protocol` | string | Yes | `tcp`, `udp`, or `icmp` |
| `port_range` | string | No* | Port or range (e.g., `"443"`, `"8000-9000"`) |
| `source_addresses` | list(string) | No** | IPv4/IPv6 CIDR blocks |
| `source_droplet_ids` | list(int64) | No** | Droplet IDs |
| `source_tags` | list(string) | No** | Droplet tags |
| `source_kubernetes_ids` | list(string) | No** | Kubernetes cluster IDs |
| `source_load_balancer_uids` | list(string) | No** | Load Balancer UIDs |

\* Required for TCP/UDP; omit for ICMP  
\** At least one source field must be specified

### Outbound Rule Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `protocol` | string | Yes | `tcp`, `udp`, or `icmp` |
| `port_range` | string | No* | Port or range |
| `destination_addresses` | list(string) | No** | IPv4/IPv6 CIDR blocks |
| `destination_droplet_ids` | list(int64) | No** | Droplet IDs |
| `destination_tags` | list(string) | No** | Droplet tags |
| `destination_kubernetes_ids` | list(string) | No** | Kubernetes cluster IDs |
| `destination_load_balancer_uids` | list(string) | No** | Load Balancer UIDs |

\* Required for TCP/UDP; omit for ICMP  
\** At least one destination field must be specified

## Common Use Cases

### 1. Development Web Server (Permissive)

Simple firewall for development with open SSH, HTTP, HTTPS:

```yaml
spec:
  name: dev-web-firewall
  tags:
    - dev-web
  inbound_rules:
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "0.0.0.0/0"
    - protocol: tcp
      port_range: "80"
      source_addresses:
        - "0.0.0.0/0"
    - protocol: tcp
      port_range: "443"
      source_addresses:
        - "0.0.0.0/0"
  outbound_rules:
    - protocol: tcp
      port_range: "1-65535"
      destination_addresses:
        - "0.0.0.0/0"
    - protocol: udp
      port_range: "1-65535"
      destination_addresses:
        - "0.0.0.0/0"
```

**Warning**: Never use this pattern in production. Always lock down SSH and use Load Balancer UIDs for public traffic.

### 2. Management Firewall (Applied to All Instances)

Centralized SSH and monitoring access for all servers:

```yaml
spec:
  name: management-firewall
  tags:
    - all-instances
  inbound_rules:
    # SSH from office bastion
    - protocol: tcp
      port_range: "22"
      source_addresses:
        - "203.0.113.10/32"
    # Monitoring agent
    - protocol: tcp
      port_range: "9100"
      source_tags:
        - prometheus-server
  outbound_rules:
    # DNS
    - protocol: udp
      port_range: "53"
      destination_addresses:
        - "0.0.0.0/0"
    # HTTPS for OS updates
    - protocol: tcp
      port_range: "443"
      destination_addresses:
        - "0.0.0.0/0"
```

### 3. Multi-Tier Architecture

Compose firewalls by role:

**Web Tier**:
```yaml
tags:
  - prod-web-tier
inbound_rules:
  - protocol: tcp
    port_range: "443"
    source_load_balancer_uids:
      - "lb-abc123"
outbound_rules:
  - protocol: tcp
    port_range: "5432"
    destination_tags:
      - prod-db-tier
```

**Database Tier**:
```yaml
tags:
  - prod-db-tier
inbound_rules:
  - protocol: tcp
    port_range: "5432"
    source_tags:
      - prod-web-tier
outbound_rules: []  # No outbound allowed (maximum security)
```

## Best Practices

### Security

1. **Default Deny**: Cloud Firewalls block all traffic by default. Only add explicit "allow" rules.
2. **Never Expose Management Ports**: SSH (22), database ports (5432, 3306) should never allow `0.0.0.0/0` in production.
3. **Use Load Balancer UIDs**: For public services, allow traffic only from Load Balancer UIDs, not directly from `0.0.0.0/0`.
4. **Implement Outbound Rules**: High-security environments should use "default deny" outbound policies with explicit allowances.
5. **Check Both Firewalls**: Cloud Firewall (network edge) AND host firewall (`ufw`, `iptables`) are independent. Traffic must pass both.

### Architecture

1. **Tag-Based Targeting for Production**: Use tags, not static Droplet IDs. Tags scale infinitely and auto-scale friendly.
2. **Compose Firewalls by Role**: Create focused firewalls (management, web-tier, db-tier) and apply multiple to each Droplet.
3. **Separation of Concerns**: Isolate management (SSH), public traffic (HTTPS), and internal communication (DB access) into separate firewalls.
4. **Least Privilege**: Database tier should be completely isolated from the internet. Only allow internal service communication.

### Operations

1. **Troubleshoot Both Layers**: When debugging connectivity issues, check Cloud Firewall (DigitalOcean dashboard) AND host firewall (`sudo ufw status`).
2. **Use Web Console for Host Access**: If SSH is blocked, use DigitalOcean's Web Console to access Droplet and check host firewall.
3. **Monitor Rule Count**: Maximum 50 rules per firewall. If you need more, split into multiple firewalls applied via tags.

## Validation Rules

The protobuf spec enforces these constraints at compile-time:

- `name`: Required, 1-255 characters
- `protocol`: Required, min 1 character
- Rules are optional (firewall can exist without rules, blocking all traffic)
- At least one source must be specified for inbound rules
- At least one destination must be specified for outbound rules

## Architecture: The "Double Firewall" Reality

DigitalOcean Cloud Firewalls operate at the **network edge**, outside your Droplets. Host-based firewalls (`ufw`, `iptables`) run **inside** the operating system. They are **independent** and **don't communicate**.

**Traffic Flow**:
```
Internet → Cloud Firewall (network edge) → Host Firewall (OS) → Application
```

**Common Pitfall**: You configure Cloud Firewall to allow port 443, but the connection times out because `ufw` inside the Droplet is still blocking it.

**Solution**: Either:
1. **Configure both firewalls** to allow the same traffic, OR
2. **Disable host firewall** and rely solely on Cloud Firewall:
   ```bash
   sudo ufw disable
   ```

For production, option 1 (defense in depth) is preferred, but option 2 (simplified operations) is common.

## Outputs

After successful provisioning, the following outputs are available:

| Output | Description |
|--------|-------------|
| `firewall_id` | DigitalOcean UUID for the firewall |

## Troubleshooting

### Connection Times Out Despite Firewall Rule

**Cause**: Host firewall (`ufw` or `iptables`) is blocking traffic inside the Droplet.

**Solution**: Check host firewall via Web Console:
```bash
sudo ufw status verbose
```

If blocking, either allow the port in `ufw` or disable host firewall entirely.

### Firewall Applied But Droplets Not Protected

**Cause**: Tags don't match. Firewall has `tags: ["web-tier"]` but Droplets are tagged `web-server`.

**Solution**: Verify tag consistency. In DigitalOcean dashboard, check Droplet's "Networking" tab to see all applied firewalls.

### Auto-Scaling Creates Unprotected Droplets

**Cause**: Using static `droplet_ids` instead of tags. New Droplets from auto-scaling don't have firewall applied.

**Solution**: Switch to tag-based targeting. Apply tags to auto-scaling template/launch configuration.

## Further Reading

- **Comprehensive Guide**: See [docs/README.md](./docs/README.md) for deep-dive coverage of deployment methods, IaC tool comparison, and production patterns
- **Examples**: See [examples.md](./examples.md) for copy-paste ready manifests for common scenarios
- **Pulumi Module**: See [iac/pulumi/README.md](./iac/pulumi/README.md) for standalone Pulumi usage
- **Terraform Module**: See [iac/tf/README.md](./iac/tf/README.md) for standalone Terraform usage

## Support

For issues, questions, or contributions, refer to the [OpenMCF documentation](https://openmcf.org) or file an issue in the repository.

---

**TL;DR**: Use tag-based targeting for production. Compose firewalls by role (management, web-tier, db-tier). Remember the "double firewall" trap—Cloud Firewall (network edge) and host firewall (`ufw`) are independent. Use Load Balancer UIDs for public services, never expose management ports to `0.0.0.0/0`.
