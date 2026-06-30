# Civo Reserved IP Address

Reserve static public IPv4 addresses on Civo that persist independently of instance lifecycle.

## Overview

The `CivoIpAddress` resource provisions a persistent public IPv4 address in Civo Cloud. Unlike dynamic IPs that disappear when you delete an instance, reserved IPs remain in your account until explicitly deleted. This makes them ideal for services requiring stable network endpoints.

**Key features:**

- **Persistent addresses** - IPs survive instance deletion and recreation
- **Free when attached** - No monthly charges (unlike AWS, GCP, Azure)
- **Region-scoped** - Each IP belongs to a specific Civo region
- **Simple API** - Minimal configuration (just region + optional description)
- **Flexible attachment** - Attach to instances or load balancers

## Prerequisites

- Civo account with API access
- Civo API token ([get one here](https://dashboard.civo.com/security))
- Planton CLI installed

## Quick Start

### 1. Basic Reserved IP

Reserve a static IP in London region:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoIpAddress
metadata:
  name: web-server-ip
spec:
  description: Production web server static IP
  region: lon1
```

### 2. Load Balancer IP

Reserve an IP for Kubernetes LoadBalancer service:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoIpAddress
metadata:
  name: k8s-ingress-ip
spec:
  description: Kubernetes ingress LoadBalancer IP
  region: nyc1
```

### 3. API Gateway IP

Reserve an IP for API endpoints:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoIpAddress
metadata:
  name: api-gateway-ip
spec:
  description: API Gateway static endpoint
  region: fra1
```

## Deploy with Planton CLI

```bash
# Reserve the IP
planton apply -f reserved-ip.yaml

# Check status
planton get civoipaddresses

# View outputs (IP address, ID)
planton outputs civoipaddresses/web-server-ip
```

## Configuration Reference

### Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `description` | string | No | Human-readable label (max 100 chars) |
| `region` | enum | Yes | Civo region: `lon1`, `lon2`, `fra1`, `nyc1`, `phx1`, `mum1` |

### Available Regions

- `lon1` - London (UK)
- `lon2` - London 2 (UK) 
- `fra1` - Frankfurt (Germany)
- `nyc1` - New York (US)
- `phx1` - Phoenix (US)
- `mum1` - Mumbai (India)

## Stack Outputs

After provisioning, the following outputs are available:

- `reserved_ip_id` - Civo's unique IP identifier
- `ip_address` - The actual IPv4 address (e.g., `74.220.24.88`)
- `region` - Region where IP is reserved
- `created_at_rfc3339` - Timestamp of IP reservation

Access outputs via:

```bash
planton outputs civoipaddresses/your-ip-name
```

## Using Reserved IPs

### With Civo Instances

Attach at instance creation:

```bash
# Get the IP address from outputs
IP_ADDRESS=$(planton outputs civoipaddresses/web-server-ip --field ip_address)

# Create instance with reserved IP
civo instance create my-web-server \\
  --size g4s.small \\
  --reserved-ipv4 $IP_ADDRESS
```

### With Kubernetes LoadBalancer

Annotate your Service to request specific IP:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-ingress
  annotations:
    kubernetes.civo.com/loadbalancer-algorithm: round_robin
spec:
  type: LoadBalancer
  loadBalancerIP: "74.220.24.88"  # Your reserved IP
  ports:
    - port: 80
      targetPort: 80
  selector:
    app: nginx
```

### With Terraform

Reference the IP in Terraform:

```hcl
resource "civo_instance" "web" {
  hostname      = "web-server"
  size          = "g4s.small"
  reserved_ipv4 = "74.220.24.88"  # From outputs
}
```

## Best Practices

### 1. Descriptive Labels

Use clear descriptions to identify IP purpose:

```yaml
spec:
  description: "Prod Web LB - us-east (DO NOT DELETE)"
  region: nyc1
```

### 2. Region Planning

Reserve IPs in the same region as your resources:

- London instances → `lon1` or `lon2` IPs
- New York instances → `nyc1` IPs
- Frankfurt instances → `fra1` IPs

**Cross-region attachment fails** - you cannot attach a `lon1` IP to a `nyc1` instance.

### 3. DNS Integration

Update DNS records to point to reserved IPs:

```yaml
# After reserving IP, update DNS
apiVersion: civo.planton.dev/v1
kind: CivoDnsZone
metadata:
  name: example-zone
spec:
  domainName: example.com
  records:
    - name: "@"
      type: A
      values:
        - value: "74.220.24.88"  # Your reserved IP
      ttlSeconds: 3600
```

### 4. Lifecycle Management

Reserved IPs persist independently:

- Deleting an instance doesn't delete its IP
- You must explicitly delete IPs when no longer needed
- Unattached IPs count against your account quota

### 5. Cost Awareness

Civo's pricing model:

- **Attached IPs**: Free (no monthly charge)
- **Unattached IPs**: May incur charges (check current pricing)
- **Best practice**: Attach IPs promptly or delete if unused

### 6. Quota Management

Check your account quotas:

```bash
# View current quota usage
civo quota show
```

If you hit quota limits, delete unused IPs or request increase from Civo support.

## Common Use Cases

### High-Availability Failover

Reserve IP for primary/standby setup:

```yaml
# Reserve HA IP
apiVersion: civo.planton.dev/v1
kind: CivoIpAddress
metadata:
  name: ha-primary-ip
spec:
  description: HA Primary - move to standby on failure
  region: lon1
```

Move IP between instances during failover (via Civo CLI/API).

### Static Load Balancer Endpoint

Kubernetes services with predictable IPs:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoIpAddress
metadata:
  name: ingress-lb-ip
spec:
  description: Ingress controller LoadBalancer
  region: fra1
```

Configure DNS once, IP never changes even if pods restart.

### Firewall Whitelist Requirements

Some systems require static IPs for firewall rules:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoIpAddress
metadata:
  name: api-client-ip
spec:
  description: IP for accessing partner API (whitelisted)
  region: nyc1
```

Partner whitelists your IP, it remains constant across deployments.

## Troubleshooting

### Cannot Reserve IP (Quota Exceeded)

**Error**: `Quota exceeded for reserved IPs`

**Solution**:
1. Check usage: `civo quota show`
2. Delete unused IPs: `civo ip delete <ip-name>`
3. Request quota increase from Civo support

### IP Already Reserved

**Error**: `IP address already reserved`

**Solution**: The IP may exist from a previous deployment. Either:
- Import existing IP: `planton import civoipaddresses/my-ip`
- Choose different name/region

### Cannot Attach IP to Instance

**Error**: `Cannot attach IP - region mismatch`

**Solution**: Ensure instance and IP are in same region:

```bash
# Check IP region
planton get civoipaddresses/my-ip -o yaml | grep region

# Create instance in matching region
civo instance create --region LON1 ...
```

### IP Not Showing in Outputs

**Solution**:
1. Verify deployment: `planton get civoipaddresses`
2. Check stack status: `planton status civoipaddresses/my-ip`
3. Re-run: `planton apply -f reserved-ip.yaml`

## Limitations

### Regional Scope

- IPs are locked to creation region
- Cannot move IP between regions
- Plan multi-region deployments carefully

### IPv4 Only

- Civo currently provides IPv4 addresses only
- IPv6 not supported (as of 2025)

### Attachment Restrictions

- One IP per instance (cannot attach multiple IPs to single instance)
- IP must be detached before attaching to different instance

## More Information

- **Deep Dive** - See [docs/README.md](docs/README.md) for comprehensive research on deployment methods, cost analysis, and design decisions
- **Examples** - Check [examples.md](examples.md) for more IP reservation patterns
- **Pulumi Module** - See [iac/pulumi/README.md](iac/pulumi/README.md) for direct Pulumi usage
- **Civo IP API** - [Official API documentation](https://www.civo.com/api/ips)

## Support

- Issues & Feature Requests: [Planton GitHub](https://github.com/plantonhq/planton/issues)
- Civo Support: [support@civo.com](mailto:support@civo.com)
- Community: [Planton Discord](#)

## Related Resources

- `CivoComputeInstance` - Attach reserved IPs to instances
- `CivoKubernetesCluster` - Use reserved IPs for LoadBalancer services
- `CivoDnsZone` - Point DNS records to reserved IPs

