# ScalewayLoadBalancer

The **ScalewayLoadBalancer** resource provides a declarative way to provision and manage Scaleway Load Balancers through OpenMCF. It is a **composite resource** that bundles a Flexible IP, the Load Balancer appliance, backend server pools, frontend listeners, and optional TLS certificates into a single manifest.

## What It Represents

A [Scaleway Load Balancer](https://www.scaleway.com/en/load-balancer/) is a managed Layer 4/7 traffic distribution appliance. It sits in front of backend servers and distributes incoming connections across them based on configurable rules.

## Bundled Terraform Resources

Applying a single `ScalewayLoadBalancer` manifest creates up to 5 Scaleway resource types:

| Terraform Resource | Purpose |
|---|---|
| `scaleway_lb_ip` | Dedicated Flexible IPv4 with independent lifecycle |
| `scaleway_lb` | The Load Balancer appliance |
| `scaleway_lb_backend` | Backend server pool(s) with health checks |
| `scaleway_lb_frontend` | Frontend listener(s) that route traffic to backends |
| `scaleway_lb_certificate` | TLS certificate(s) for HTTPS frontends |

## Key Features

### Named Backends and Frontends

Backends and frontends are **named** entities. Frontends reference backends by name (`backend_name` field), creating a clear, self-documenting relationship:

```yaml
backends:
  - name: web
    serverIps: ["10.0.1.5", "10.0.1.6"]
    forwardPort: 80
    forwardProtocol: http
frontends:
  - name: http
    inboundPort: 80
    backendName: web    # ← references the "web" backend
```

### Private Network Integration

The LB can be attached to a `ScalewayPrivateNetwork` via the `private_network_id` field (using `StringValueOrRef`). When attached, the LB receives a private IP and can reach backend servers on their private addresses -- the recommended production topology.

### TLS Certificate Management

Certificates support both **Let's Encrypt** (auto-provisioned and auto-renewed) and **custom PEM** chains. Frontends reference certificates by name:

```yaml
certificates:
  - name: my-cert
    letsencrypt:
      commonName: example.com
frontends:
  - name: https
    inboundPort: 443
    backendName: web
    certificateNames: ["my-cert"]
```

### Configurable Health Checks

Each backend has its own health check configuration supporting TCP, HTTP, and HTTPS probes with configurable intervals, timeouts, and retry counts.

### Load-Balancing Algorithms

Supported algorithms per backend: `roundrobin` (default), `leastconn`, and `first` (active-passive).

## Upstream Dependencies (What This Resource Needs)

| Dependency | Field | Purpose |
|---|---|---|
| `ScalewayPrivateNetwork` | `spec.private_network_id` | Attach LB to a private network for backend connectivity |

## Downstream Dependents (What References This Resource)

| Dependent | Output Used | Purpose |
|---|---|---|
| `ScalewayDnsRecord` | `status.outputs.lb_ip_address` | Create DNS A records pointing to the LB |

## Stack Outputs

| Output | Description |
|---|---|
| `lb_id` | The zoned ID of the Load Balancer |
| `lb_ip_address` | The public IPv4 address (for DNS records, monitoring) |
| `lb_ip_id` | The Flexible IP resource ID |

## LB Types

| Type | Bandwidth | Use Case |
|---|---|---|
| `LB-S` | Up to 400 Mbps | Development, small apps |
| `LB-GP-M` | Up to 4 Gbps | General-purpose production |
| `LB-GP-L` | Up to 8 Gbps | High-traffic applications |
| `LB-GP-XL` | Up to 10 Gbps | Maximum throughput |

## What's Not Included (Deferred)

The following Scaleway LB features are intentionally deferred to future versions:

- **Routes** (`scaleway_lb_route`) -- Host/path/SNI-based traffic routing
- **ACLs** (`scaleway_lb_acl`) -- Access control lists on frontends
- **IPv6** -- Flexible IPv6 addresses
- **Multiple IPs** -- Attaching multiple Flexible IPs to a single LB

These can be added as new optional fields in future spec versions without breaking changes.

## References

- [Scaleway Load Balancer Documentation](https://www.scaleway.com/en/docs/network/load-balancer/)
- [Scaleway Load Balancer Concepts](https://www.scaleway.com/en/docs/network/load-balancer/concepts/)
- [Terraform scaleway_lb Resource](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/lb)
- [Scaleway LB Types and Pricing](https://www.scaleway.com/en/pricing/?tags=network)
