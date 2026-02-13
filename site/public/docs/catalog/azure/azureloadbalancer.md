---
title: "Load Balancer"
description: "Load Balancer deployment documentation"
icon: "package"
order: 100
componentName: "azureloadbalancer"
---

# AzureLoadBalancer -- Research & Design Documentation

## What is Azure Load Balancer?

Azure Load Balancer is a Layer 4 (transport layer) load balancing service that distributes incoming TCP and UDP traffic across healthy backend instances. It operates entirely at the network layer -- inspecting IP addresses and ports, not HTTP headers or URLs -- making it the right choice for high-throughput, protocol-agnostic traffic distribution.

Azure Load Balancer is a fully managed, zone-redundant service with built-in high availability. It is Azure's primary mechanism for distributing traffic within and into virtual networks.

## Deployment Landscape

### When to Use Azure Load Balancer

Azure Load Balancer is the right choice when you need:

1. **TCP/UDP load balancing** without application-layer inspection
2. **Internal service mesh routing** within a VNet (internal LB)
3. **High-performance traffic distribution** -- millions of flows per second
4. **HA ports** -- forward all ports and protocols for NVA or SQL AlwaysOn
5. **Ultra-low latency** -- no HTTP parsing overhead

### When NOT to Use Azure Load Balancer

Consider alternatives for these scenarios:

| Scenario | Better Alternative | Reason |
|----------|-------------------|--------|
| HTTP/HTTPS routing with path-based rules | **Application Gateway** | Layer 7 routing, SSL termination, WAF |
| Global traffic distribution | **Front Door** | Global anycast, CDN, WAF |
| DNS-based load balancing | **Traffic Manager** | Geographic routing, failover |
| Kubernetes ingress | **AKS Ingress Controller** | Native K8s integration |

### Layer 4 vs Layer 7 Comparison

| Feature | Load Balancer (L4) | Application Gateway (L7) |
|---------|-------------------|--------------------------|
| Protocol support | TCP, UDP, All | HTTP, HTTPS, WebSocket |
| Routing decisions | IP + Port | URL path, hostname, headers |
| SSL termination | No | Yes |
| WAF | No | Yes (WAF v2 SKU) |
| Performance | Higher throughput | Lower throughput |
| Latency | Lower | Higher (HTTP parsing) |
| Cost | Lower | Higher |
| Use case | Network traffic, DBs, NVAs | Web applications |

## SKU Analysis

### Why Standard Only

Azure offers three Load Balancer SKUs:

| Feature | Basic | Standard | Gateway |
|---------|-------|----------|---------|
| Status | **Retired Sept 2025** | Active | Active (niche) |
| Zone redundancy | No | Yes | N/A |
| Outbound rules | No | Yes | N/A |
| SLA | No | 99.99% | N/A |
| Backend pool size | 300 | 1000 | N/A |
| Health probes | TCP, HTTP | TCP, HTTP, HTTPS | N/A |
| HA ports | No | Yes | Yes |
| IP-based backend pools | No | Yes | N/A |
| Use case | Deprecated | All production | NVA chaining |

**Decision**: Hardcode Standard SKU in IaC modules. Basic is deprecated and non-viable for production. Gateway is for Network Virtual Appliance (NVA) chaining, which is an extremely specialized networking pattern that fewer than 1% of deployments need.

This follows the same approach as AzurePublicIp, which also hardcodes Standard.

## Frontend IP Configuration Model

The Azure Load Balancer's frontend is configured through a `frontend_ip_configuration` block embedded in the main resource. This is the IP address that clients connect to.

### Public Frontend

For internet-facing load balancers, the frontend references a public IP address:
- `public_ip_address_id` -- ARM ID of an AzurePublicIp resource
- The public IP must use Standard SKU (matches the LB SKU)
- Zone redundancy is inherited from the public IP's zone configuration

### Internal (Private) Frontend

For VNet-internal load balancers, the frontend references a subnet:
- `subnet_id` -- ARM ID of an AzureSubnet resource
- `private_ip_address` -- Optional static IP from the subnet's range
- `private_ip_address_allocation` -- "Static" or "Dynamic" (auto-derived from whether `private_ip_address` is set)

### Design Decision: No `is_internal` Boolean

The T02 spec originally included an `is_internal` boolean. We removed it because:

1. **Redundancy**: The mode is fully determined by `public_ip_id` vs `subnet_id`
2. **Contradiction risk**: `is_internal=true` + `public_ip_id` set = undefined behavior
3. **Simplicity**: One less field to understand and validate

## Backend Pool Membership Model

Azure offers two models for backend pool membership:

### NIC-Based (Default)

Backend instances associate with pools via their Network Interface Card (NIC). This is how AKS, VMSS, and individual VMs join pools. The LB resource itself only defines the pool name -- membership is managed externally.

### IP-Based

Backend instances are added by IP address (requires `virtual_network_id` on the pool). Used for cross-VNet or cross-region scenarios with Standard SKU.

### Design Decision: Names Only

We expose only pool names in the spec. This is the 80/20 approach because:
- Pool membership is always managed externally (VMSS, AKS, NIC binding)
- Defining members inline would create lifecycle coupling
- The LB definition is about routing topology, not instance management

## Health Probe Best Practices

### Protocol Selection

| Protocol | Use Case | Health Signal |
|----------|----------|---------------|
| Tcp | Any TCP service | Port open = healthy |
| Http | Web services | HTTP 200 OK = healthy |
| Https | Secure endpoints | HTTPS 200 OK = healthy |

### Tuning Parameters

- **`interval_in_seconds`** (default 15): Lower values detect failures faster but increase probe traffic. 5-10 seconds recommended for production.
- **`number_of_probes`** (default 2): Consecutive failures before marking unhealthy. 2-3 strikes is typical.
- **`request_path`** (Http/Https only): Should be a lightweight health endpoint, not a full page load. `/health` or `/ready` patterns recommended.

### Anti-Patterns

- Using the application's main page as `request_path` (too heavy, may timeout)
- Setting `interval_in_seconds` to 5 with `number_of_probes` = 1 (too aggressive, one slow response = failover)
- Not setting `request_path` for Http/Https probes (will probe `/` which may not exist)

## 80/20 Scoping Rationale

### Included (80% of Use Cases)

- **Standard SKU** -- only production-viable SKU
- **Single frontend** -- most LBs have one IP
- **Backend pools** -- named containers for external membership
- **Health probes** -- Tcp, Http, Https with tunable intervals
- **Load balancing rules** -- frontend-to-backend mapping with idle timeout
- **Floating IP** -- SQL AlwaysOn and HA clustering
- **Disable outbound SNAT** -- SNAT port exhaustion prevention

### Omitted (20% of Use Cases)

- **NAT rules** -- Port forwarding to individual VMs (use bastion or VPN instead)
- **Outbound rules** -- Explicit SNAT configuration (use NAT Gateway instead)
- **NAT pools** -- VMSS port range allocation (legacy pattern)
- **Multiple frontends** -- Multiple IPs on same LB (rare, add via portal)
- **Gateway SKU** -- NVA chaining (< 1% of deployments)
- **Global tier** -- Cross-region Standard LB (use Front Door instead)
- **IP-based backend pools** -- Cross-VNet membership (advanced networking)

### Why No NAT Rules?

NAT rules provide port forwarding (e.g., external port 50001 → VM1:22, port 50002 → VM2:22). In modern deployments:
- Use **Azure Bastion** for SSH/RDP access instead of NAT rules
- Use **VPN Gateway** for private network access
- NAT rules on Standard LB create management complexity

### Why No Outbound Rules?

Outbound rules control SNAT behavior for outbound internet access from backends. In modern deployments:
- Use **NAT Gateway** for outbound connectivity (dedicated, scalable, simpler)
- `disable_outbound_snat` on LB rules + NAT Gateway is the recommended pattern
- We expose `disable_outbound_snat` per rule to support this pattern

## Infra Chart Integration

### Enterprise Network Foundation

AzureLoadBalancer participates in the enterprise-network-foundation infra chart:

```
AzureResourceGroup
  └── AzureVpc
        ├── AzureSubnet (web)
        │     └── AzureNetworkSecurityGroup
        ├── AzureSubnet (app)
        │     └── AzureNetworkSecurityGroup
        └── AzureSubnet (lb)
              └── AzureLoadBalancer (internal)
                    ├── backend_pool → app VMs
                    └── health_probe → /health
AzurePublicIp
  └── AzureLoadBalancer (public)
        ├── backend_pool → web VMs
        └── health_probe → /health
```

### StringValueOrRef Dependencies

| Field | References | Default Kind |
|-------|-----------|-------------|
| `resource_group` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `public_ip_id` | AzurePublicIp | `status.outputs.public_ip_id` |
| `subnet_id` | AzureSubnet | `status.outputs.subnet_id` |

## Corrections from T02 Spec

Eight corrections were applied during deep provider research:

1. **Added `resource_group`** (StringValueOrRef) -- missing from T02, required per DD05
2. **Added `region`** (string) -- missing from T02, required per established pattern
3. **Dropped `sku` field** -- hardcoded to Standard (Basic retired, matches AzurePublicIp)
4. **Dropped `is_internal` boolean** -- redundant with public_ip_id/subnet_id presence
5. **Added `private_ip_address`** -- static IP for internal LB (enterprise requirement)
6. **Added `number_of_probes`** to health probe -- production tunable
7. **Used string+CEL for protocols** -- matches NSG pattern (provider-authentic values)
8. **Simplified outputs** -- single `backend_pool_id` instead of `repeated` (80/20)
