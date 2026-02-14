# Azure Front Door Profile

## Overview

The **AzureFrontDoorProfile** component provisions an Azure Front Door profile with endpoints, origin groups, origins, and routes, providing a global CDN and application delivery network that combines HTTP load balancing, SSL offloading, caching, and application acceleration in a single service.

Azure Front Door uses Microsoft's global edge network to create fast, secure, and widely scalable web applications. Traffic enters at the nearest edge location, is routed to the optimal origin based on health probes and latency, and can be cached at the edge to reduce origin load. A profile bundles endpoints (public entry points), origin groups (pools of backends), origins (individual backends), and routes (URL-to-backend mappings) because these five resources form a single coherent routing unit that cannot serve traffic independently.

## When to Use This Component

Use AzureFrontDoorProfile when you need:

- **Global CDN acceleration** for static and dynamic web content served from edge locations worldwide
- **Multi-origin load balancing** with health-probe-based failover across regions or backend services
- **SSL offloading** at the edge to terminate TLS and forward traffic to origins over HTTP or HTTPS
- **URL-based routing** directing different URL paths (e.g., `/api/*`, `/static/*`) to different backend pools
- **Caching and compression** at the edge to reduce origin load and improve end-user latency
- **Private link to origins** (Premium SKU) to connect to backends without exposing them to the public internet
- **Application acceleration** for latency-sensitive APIs and SPAs that benefit from anycast routing

## SKU Tiers

| Tier | Caching | Compression | Private Link | WAF | Bot Manager | SLA |
|------|---------|-------------|--------------|-----|-------------|------|
| Standard_AzureFrontDoor | Yes | Yes | No | No | No | 99.99% |
| Premium_AzureFrontDoor | Yes | Yes | Yes | Enhanced | Yes | 99.99% |

**Recommendation**: Use **Standard_AzureFrontDoor** for most production workloads (CDN, SSL offloading, multi-origin routing). Use **Premium_AzureFrontDoor** only when you need private link connectivity to backends, enhanced WAF integration, or Bot Manager.

## Key Configuration

### Profile-Level Settings

- **`name`**: Globally unique profile name (2-46 characters). **ForceNew**.
- **`sku`**: SKU tier (`Standard_AzureFrontDoor` or `Premium_AzureFrontDoor`). Defaults to Standard. **ForceNew**.
- **`response_timeout_seconds`**: How long Front Door waits for an origin response (16-240 seconds). Default: 120.

### Endpoint Settings

- **`name`**: Endpoint name, generates a public hostname (`{name}-{hash}.z01.azurefd.net`). **ForceNew**.
- **`enabled`**: Whether the endpoint is accepting traffic. Default: true.

### Origin Group Settings

- **`name`**: Origin group name, referenced by routes.
- **`session_affinity_enabled`**: Sticky sessions via cookies. Default: true.
- **`load_balancing`**: Sample size, successful samples required, additional latency tolerance.
- **`health_probe`**: Protocol, path, request type, interval. Omit to disable probing.

### Origin Settings

- **`host_name`**: Backend hostname (e.g., `myapp.azurewebsites.net`).
- **`origin_host_header`**: Host header sent to the origin. Critical for multi-tenant backends.
- **`priority`**: Active-passive failover (1-5, lower is preferred). Default: 1.
- **`weight`**: Traffic distribution within same priority (1-1000). Default: 500.
- **`private_link`**: Premium-only private connectivity to the origin.

### Route Settings

- **`endpoint_name`**: Which endpoint receives traffic for this route.
- **`origin_group_name`**: Which origin group serves this route.
- **`patterns_to_match`**: URL patterns (e.g., `/*`, `/api/*`).
- **`cache`**: Optional caching with query string behavior, compression, and content type filtering.

## Deliberately Omitted

The following are deliberately omitted from this component to keep it focused (80/20 rule):

- **Custom domains** (separate resource with DNS validation and certificate lifecycle)
- **WAF/firewall policies** (separate resource with complex rule structure)
- **Rule sets** (advanced URL rewriting and header manipulation)
- **Security policies** (WAF-to-profile association)
- **Secrets** (custom domain certificates)

## Outputs

| Output | Description |
|--------|-------------|
| `profile_id` | Azure Resource Manager ID of the Front Door profile |
| `profile_name` | The profile name |
| `resource_guid` | Front Door service GUID |
| `endpoint_ids` | Map of endpoint names to Azure resource IDs |
| `endpoint_hostnames` | Map of endpoint names to generated hostnames (*.azurefd.net) |

## Infra Chart Usage

AzureFrontDoorProfile is typically a **mid-tier resource** in infra chart DAGs. Its endpoint hostnames are consumed by AzureDnsRecord components for custom domain CNAME setup, and applications reference it for CDN configuration.

```yaml
# In a DNS record infra chart:
spec:
  records:
    - name: cdn
      recordType: CNAME
      values:
        - valueFrom:
            kind: AzureFrontDoorProfile
            name: "{{ values.env }}-cdn"
            fieldPath: status.outputs.endpoint_hostnames.main-endpoint
```

## Azure Documentation

- [Azure Front Door overview](https://learn.microsoft.com/en-us/azure/frontdoor/front-door-overview)
- [Front Door SKU comparison](https://learn.microsoft.com/en-us/azure/frontdoor/standard-premium/tier-comparison)
- [Front Door routing architecture](https://learn.microsoft.com/en-us/azure/frontdoor/front-door-routing-architecture)
- [Private Link with Front Door](https://learn.microsoft.com/en-us/azure/frontdoor/private-link)
- [Caching with Front Door](https://learn.microsoft.com/en-us/azure/frontdoor/front-door-caching)
