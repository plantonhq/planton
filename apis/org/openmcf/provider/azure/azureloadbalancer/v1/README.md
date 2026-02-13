# AzureLoadBalancer

Azure Load Balancer is a Layer 4 (TCP/UDP) network load balancer that distributes incoming traffic across healthy backend instances. It operates at the transport layer, making routing decisions based on IP address and port combinations without inspecting application-layer content.

## When to Use

Use AzureLoadBalancer when you need:

- **Layer 4 load balancing** -- TCP/UDP traffic distribution without HTTP inspection
- **High availability** -- Automatic health checks remove unhealthy backends from rotation
- **Internal service routing** -- Private load balancing within a VNet for microservice architectures
- **Public internet ingress** -- Internet-facing traffic distribution to backend pools
- **HA ports** -- Forward all ports and protocols (protocol "All") for NVA or SQL AlwaysOn scenarios

For Layer 7 (HTTP/HTTPS) load balancing with path-based routing, SSL termination, or WAF, use **AzureApplicationGateway** instead.

## Key Configuration

### Public vs Internal

The LB mode is determined by which frontend field you set -- there is no separate `is_internal` flag:

- **`public_ip_id`** -- Creates a public (internet-facing) load balancer using the referenced AzurePublicIp
- **`subnet_id`** -- Creates an internal (private VNet) load balancer in the specified subnet

Exactly one must be set. For internal LBs, you can optionally specify a `private_ip_address` for a static IP (otherwise Azure allocates dynamically).

### Standard SKU

Standard SKU is hardcoded. Basic SKU was retired by Azure in September 2025 and lacks zone redundancy, outbound rule support, and SLA guarantees. This matches the AzurePublicIp component which also hardcodes Standard.

### Backend Pools

Backend pools define named containers for backend instances. Only the pool name is configured here. Actual instance membership -- VMs, VMSS instances, or NICs -- is managed through separate mechanisms (AKS node pools, VMSS configurations, or NIC-to-pool bindings). This keeps the LB lifecycle clean and independent.

### Health Probes

Health probes check backend availability at configurable intervals. Supported protocols:

- **Tcp** -- Simple TCP connection check (port open = healthy)
- **Http** -- HTTP GET request, expects 200 OK
- **Https** -- HTTPS GET request, expects 200 OK

After `number_of_probes` consecutive failures (default 2), the backend is removed from rotation. Healthy backends are automatically re-added.

### Load Balancing Rules

Each rule maps a frontend port/protocol to a backend pool and health probe. The frontend IP configuration name is auto-derived from the LB name and does not need to be specified.

Advanced options per rule:
- **`enable_floating_ip`** -- Direct Server Return for SQL AlwaysOn and HA clustering
- **`disable_outbound_snat`** -- Prevent SNAT port exhaustion when using NAT Gateway for outbound

## Outputs

| Output | Description |
|--------|-------------|
| `lb_id` | Azure Resource Manager ID |
| `lb_name` | Load Balancer name |
| `frontend_ip_address` | Frontend IP (public or private) |
| `frontend_ip_configuration_id` | Frontend config ID (for NAT rule association) |
| `backend_pool_id` | First (default) backend pool ID |

## Infra Chart Usage

AzureLoadBalancer is a component in the **enterprise-network-foundation** infra chart, providing Layer 4 traffic distribution alongside AzureApplicationGateway (Layer 7), NSGs, and public IPs.

## Related Resources

- **AzurePublicIp** -- Provides the public IP for public load balancers
- **AzureSubnet** -- Provides the subnet for internal load balancers
- **AzureApplicationGateway** -- Layer 7 alternative with HTTP routing and WAF
- **AzureDnsRecord** -- Create DNS records pointing to the frontend IP
