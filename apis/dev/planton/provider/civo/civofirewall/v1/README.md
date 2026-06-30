# Civo Firewall

Manage network security for Civo instances and Kubernetes clusters using declarative firewall rules.

## Overview

The `CivoFirewall` resource provides a declarative way to configure network security on Civo. It allows you to define inbound (ingress) and outbound (egress) rules that control traffic to and from your instances, creating a stateful firewall that automatically allows return traffic.

**Key features:**

- **Stateful firewall** - Return traffic automatically allowed
- **Default-deny inbound** - Only explicitly allowed traffic gets through
- **Protocol support** - TCP, UDP, and ICMP protocols
- **Port ranges** - Single ports or ranges (e.g., `80`, `8000-9000`)
- **CIDR-based rules** - Control traffic from specific IP blocks
- **Instance tags** - Auto-apply firewalls to tagged instances
- **Network scoped** - Firewalls belong to a specific Civo network (VPC)

## Prerequisites

- Civo account with API access
- Civo API token ([get one here](https://dashboard.civo.com/security))
- Existing Civo network (VPC) - use `CivoVpc` resource
- Planton CLI installed

## Quick Start

### 1. Basic Web Server Firewall

Allow HTTP/HTTPS from anywhere and SSH from office IP:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoFirewall
metadata:
  name: web-server-firewall
spec:
  name: web-server-fw
  networkId:
    value: "your-network-id-here"  # Or reference a CivoVpc
  inboundRules:
    - protocol: tcp
      portRange: "22"
      cidrs:
        - "203.0.113.10/32"  # Office IP
      action: allow
      label: SSH from office
    - protocol: tcp
      portRange: "80"
      cidrs:
        - "0.0.0.0/0"  # Anywhere
      action: allow
      label: HTTP
    - protocol: tcp
      portRange: "443"
      cidrs:
        - "0.0.0.0/0"
      action: allow
      label: HTTPS
  tags:
    - web-server
```

### 2. Database Firewall

Restrict access to database ports from application tier only:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoFirewall
metadata:
  name: database-firewall
spec:
  name: database-fw
  networkId:
    value: "your-network-id-here"
  inboundRules:
    - protocol: tcp
      portRange: "5432"
      cidrs:
        - "10.0.1.0/24"  # App tier subnet
      action: allow
      label: PostgreSQL from app tier
    - protocol: tcp
      portRange: "22"
      cidrs:
        - "203.0.113.10/32"  # Bastion host
      action: allow
      label: SSH from bastion
  tags:
    - database
    - postgresql
```

### 3. Kubernetes Cluster Firewall

Open ports required for Kubernetes clusters:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoFirewall
metadata:
  name: k8s-cluster-firewall
spec:
  name: k8s-cluster-fw
  networkId:
    value: "your-network-id-here"
  inboundRules:
    - protocol: tcp
      portRange: "6443"
      cidrs:
        - "0.0.0.0/0"
      action: allow
      label: Kubernetes API server
    - protocol: tcp
      portRange: "30000-32767"
      cidrs:
        - "0.0.0.0/0"
      action: allow
      label: Kubernetes NodePort range
    - protocol: tcp
      portRange: "80"
      cidrs:
        - "0.0.0.0/0"
      action: allow
      label: HTTP ingress
    - protocol: tcp
      portRange: "443"
      cidrs:
        - "0.0.0.0/0"
      action: allow
      label: HTTPS ingress
  tags:
    - kubernetes
    - cluster
```

## Deploy with Planton CLI

```bash
# Create the firewall
planton apply -f firewall.yaml

# Check status
planton get civofirewalls

# View outputs (firewall ID)
planton outputs civofirewalls/web-server-firewall
```

## Configuration Reference

### Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique name for the firewall (per Civo account) |
| `networkId` | StringValueOrRef | Yes | Network (VPC) to create firewall in |
| `inboundRules` | array | No | List of inbound (ingress) rules |
| `outboundRules` | array | No | List of outbound (egress) rules |
| `tags` | array[string] | No | Instance tags to auto-apply this firewall |

### Inbound Rule Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `protocol` | string | Yes* | Protocol: `tcp`, `udp`, or `icmp` |
| `portRange` | string | No | Port or port range (e.g., `80`, `8000-9000`). Empty for ICMP |
| `cidrs` | array[string] | No | Source CIDR blocks (default: `0.0.0.0/0`) |
| `action` | string | No | Rule action: `allow` or `deny` (default: `allow`) |
| `label` | string | No | Human-readable description |

*Protocol must match pattern: `^(tcp|udp|icmp)$` (lowercase only)

### Outbound Rule Fields

Same as inbound rules, but controls traffic **from** instances to destinations.

## Stack Outputs

After provisioning, the following outputs are available:

- `firewall_id` - Civo's unique firewall identifier

Access outputs via:

```bash
planton outputs civofirewalls/your-firewall-name
```

## Common Use Cases

### Allow Ping (ICMP)

```yaml
inboundRules:
  - protocol: icmp
    portRange: ""  # Empty for ICMP
    cidrs:
      - "0.0.0.0/0"
    action: allow
    label: Allow ping
```

### Restrict SSH to Specific IPs

```yaml
inboundRules:
  - protocol: tcp
    portRange: "22"
    cidrs:
      - "203.0.113.10/32"  # Office IP 1
      - "198.51.100.20/32"  # Office IP 2
      - "192.0.2.30/32"     # VPN IP
    action: allow
    label: SSH from trusted IPs
```

### Port Range for Custom Applications

```yaml
inboundRules:
  - protocol: tcp
    portRange: "8000-9000"
    cidrs:
      - "10.0.0.0/8"
    action: allow
    label: Application port range
```

### Restrict Outbound Traffic

By default, all outbound traffic is allowed. To restrict:

```yaml
outboundRules:
  - protocol: tcp
    portRange: "443"
    cidrs:
      - "0.0.0.0/0"
    action: allow
    label: HTTPS only
  - protocol: tcp
    portRange: "53"
    cidrs:
      - "0.0.0.0/0"
    action: allow
    label: DNS
  - protocol: udp
    portRange: "53"
    cidrs:
      - "0.0.0.0/0"
    action: allow
    label: DNS
```

### Reference Another Resource's Network

Instead of hardcoding network ID, reference a `CivoVpc`:

```yaml
spec:
  name: my-firewall
  networkId:
    valueFrom:
      kind: CivoVpc
      name: my-vpc
      env: production
      fieldPath: "status.outputs.network_id"
```

## Best Practices

### 1. Default-Deny Philosophy

- Start with **no inbound rules** (implicit deny-all)
- Add only what's explicitly needed
- This prevents accidentally exposing services

### 2. Least Privilege Access

- Restrict SSH to office IPs or VPN ranges, not `0.0.0.0/0`
- Database ports should only allow traffic from app tier subnets
- Use `/32` CIDR notation for single IPs

### 3. Use Descriptive Labels

```yaml
inboundRules:
  - protocol: tcp
    portRange: "443"
    label: "HTTPS from CloudFlare CDN"  # Clear purpose
```

Labels help you understand rule intent months later.

### 4. Organize by Function

Create separate firewalls for different roles:
- `web-server-fw` for web servers
- `database-fw` for databases
- `k8s-cluster-fw` for Kubernetes
- `bastion-fw` for jump hosts

### 5. Use Tags for Auto-Assignment

```yaml
tags:
  - web-server
  - production
```

Any instance created with these tags automatically gets this firewall applied.

### 6. Test Before Production

1. Create firewall with restrictive rules
2. Test access from expected sources
3. Verify denied access from unexpected sources
4. Apply to production instances

### 7. Document Your Rules

Use labels and commit firewall configs to Git with descriptive commit messages:

```bash
git commit -m "Add office IP 198.51.100.20 to web-server-fw for new team member"
```

## Security Considerations

### Stateful Firewall Behavior

Civo firewalls are **stateful**:
- Inbound rule allowing port 80 automatically allows return traffic
- You don't need outbound rules for response packets
- This prevents common "broken return traffic" issues

### Default Behaviors

- **Inbound**: Default deny (only explicitly allowed traffic passes)
- **Outbound**: Default allow (unless you add outbound rules)

### Dangerous Patterns to Avoid

❌ **Don't**: Allow SSH from anywhere

```yaml
- protocol: tcp
  portRange: "22"
  cidrs: ["0.0.0.0/0"]  # BAD: Exposes SSH to brute force
```

✅ **Do**: Restrict to known IPs

```yaml
- protocol: tcp
  portRange: "22"
  cidrs: ["203.0.113.10/32"]  # GOOD: Only office IP
```

❌ **Don't**: Leave database ports open to internet

```yaml
- protocol: tcp
  portRange: "5432"
  cidrs: ["0.0.0.0/0"]  # BAD: Database exposed
```

✅ **Do**: Restrict to application tier

```yaml
- protocol: tcp
  portRange: "5432"
  cidrs: ["10.0.1.0/24"]  # GOOD: Only app subnet
```

### Audit Your Rules Regularly

```bash
# List all firewalls
planton get civofirewalls

# Review specific firewall
planton get civofirewalls/web-server-firewall -o yaml
```

Look for:
- Rules allowing `0.0.0.0/0` on sensitive ports
- Orphaned rules from decommissioned services
- Overly broad port ranges

## Limitations

Civo firewalls are designed for simplicity and cover most use cases, but have some limitations:

- **No advanced routing** - No geo-based or latency-based rules
- **No IDS/IPS** - Basic packet filtering only, no intrusion detection
- **No rate limiting** - Cannot throttle connections at firewall level
- **No layer 7 filtering** - Operates at network layer (ports/protocols), not application layer (HTTP paths, headers)

For advanced features, consider:
- **Cloudflare** - DDoS protection, WAF, rate limiting
- **AWS WAF** - Application-layer filtering (if using hybrid cloud)
- **Application-level firewalls** - nginx rate limiting, fail2ban

## Troubleshooting

### Connection Refused / Timeout

1. Verify firewall has rule for the port:

```bash
planton get civofirewalls/your-firewall -o yaml
```

2. Check instance is using correct firewall:

```bash
civo instance show <instance-name>
```

3. Verify service is actually listening:

```bash
# SSH to instance and check
sudo netstat -tlnp | grep :80
```

### Can't SSH After Rule Change

If you accidentally removed SSH access:

1. Use Civo web console to access instance
2. Fix firewall via CLI or dashboard
3. Test SSH access before closing console

**Prevention**: Always test SSH rule changes from a separate session before closing your current one.

### Rules Not Taking Effect

Firewall changes may take 10-30 seconds to propagate. Wait briefly and retry.

If still not working:
- Verify rule syntax (protocol lowercase, valid port range)
- Check CIDR notation is correct (`203.0.113.10/32` not `203.0.113.10`)
- Ensure firewall is attached to the correct network

### Too Many Rules

Civo doesn't document a hard limit, but 50+ rules per firewall may cause performance issues. If you have many rules:

1. Consolidate similar rules using port ranges
2. Split into multiple firewalls by function
3. Use broader CIDR blocks where appropriate (e.g., `/24` instead of multiple `/32`)

## More Information

- **Deep Dive** - See [docs/README.md](docs/README.md) for comprehensive research on deployment methods, IaC comparisons, and design decisions
- **Examples** - Check [examples.md](examples.md) for more real-world firewall patterns
- **Pulumi Module** - See [iac/pulumi/README.md](iac/pulumi/README.md) for direct Pulumi usage
- **Civo Firewall API** - [Official API documentation](https://www.civo.com/api/firewalls)

## Support

- Issues & Feature Requests: [Planton GitHub](https://github.com/plantonhq/planton/issues)
- Civo Support: [support@civo.com](mailto:support@civo.com)
- Community: [Planton Discord](#)

## Related Resources

- `CivoVpc` - Create networks for your firewalls
- `CivoComputeInstance` - Instances protected by firewalls
- `CivoKubernetesCluster` - Kubernetes clusters with firewall protection

