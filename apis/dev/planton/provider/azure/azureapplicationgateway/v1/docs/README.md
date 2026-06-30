# Azure Application Gateway: From HTTP Routing Concepts to Production L7 Infrastructure

## Introduction

Azure Application Gateway occupies a specific and important niche in Azure's networking stack: it is the **Layer 7 (HTTP/HTTPS) load balancer** for applications that need traffic routing based on HTTP attributes -- host headers, URL paths, cookies -- rather than raw TCP/UDP port numbers.

If Azure Load Balancer is the traffic cop that directs packets based on IP and port, Application Gateway is the concierge that reads the HTTP request, examines the domain name and path, terminates SSL, inspects for attacks (with WAF), and routes the request to the correct backend service.

This document covers the Application Gateway landscape, explains why the Planton spec is shaped the way it is, documents the design decisions and corrections made during specification development, and provides the research backing for the 80/20 scoping choices.

## What is Azure Application Gateway

Azure Application Gateway is a managed, scalable, highly available Layer 7 load balancer. Its core capabilities:

1. **HTTP/HTTPS reverse proxy** -- terminates client connections and creates new connections to backends
2. **SSL/TLS termination** -- offloads SSL processing from backend servers, with certificates from Azure Key Vault
3. **Host-based routing** -- routes requests to different backend pools based on the `Host` header
4. **Cookie-based session affinity** -- routes repeat clients to the same backend via gateway-managed cookies
5. **Custom health probes** -- actively monitors backend health with configurable endpoints, intervals, and thresholds
6. **Web Application Firewall (WAF)** -- inspects HTTP traffic and blocks OWASP Top 10 attacks (SQL injection, XSS, etc.)
7. **Autoscaling** -- automatically adjusts instance count based on traffic load (v2 only)
8. **Zone redundancy** -- distributes instances across Azure Availability Zones (v2 only)

Application Gateway is **not** a CDN, not a global load balancer, and not a Layer 4 load balancer. Understanding its place in the Azure networking landscape is critical to using it correctly.

## The Deployment Landscape: Application Gateway vs. Alternatives

Azure offers multiple traffic management services. Choosing the wrong one is a common and expensive mistake. Here's the decision framework:

### Azure Application Gateway vs. Azure Load Balancer

| Aspect | Application Gateway | Azure Load Balancer |
|--------|-------------------|-------------------|
| **OSI Layer** | Layer 7 (HTTP/HTTPS) | Layer 4 (TCP/UDP) |
| **Routing** | Host, path, headers, cookies | IP + port only |
| **SSL termination** | Yes | No |
| **WAF** | Yes (WAF_v2 SKU) | No |
| **Session affinity** | Cookie-based | Source IP hash |
| **Health probes** | HTTP/HTTPS with path | TCP, HTTP, HTTPS |
| **Use case** | Web applications, APIs | Databases, non-HTTP services, internal TCP |
| **Cost** | Higher (L7 processing) | Lower (L4 forwarding) |

**Decision:** If your traffic is HTTP/HTTPS and you need any of: SSL termination, host routing, WAF, or cookie affinity -- use Application Gateway. For everything else (TCP services, UDP, non-HTTP protocols), use Load Balancer.

### Azure Application Gateway vs. Azure Front Door

| Aspect | Application Gateway | Azure Front Door |
|--------|-------------------|-----------------|
| **Scope** | Regional (single Azure region) | Global (anycast across edge locations) |
| **Latency optimization** | No (traffic goes to the region) | Yes (routes to nearest edge POP) |
| **SSL termination** | Yes | Yes |
| **WAF** | Yes | Yes (separate WAF policy) |
| **Caching** | No | Yes (CDN-like caching) |
| **Use case** | Regional L7 load balancing | Global L7 with CDN, acceleration |
| **Cost** | Per-instance-hour + data | Per-request + data transfer |

**Decision:** Use Application Gateway for regional traffic (single-region deployments, internal L7 routing). Use Front Door when you need global distribution, CDN caching, or multi-region failover.

### Azure Application Gateway vs. Azure API Management

| Aspect | Application Gateway | API Management |
|--------|-------------------|---------------|
| **Focus** | Traffic routing + WAF | API lifecycle management |
| **Rate limiting** | No | Yes |
| **API versioning** | No | Yes |
| **Developer portal** | No | Yes |
| **Transformations** | No | Yes (request/response rewriting) |
| **Authentication** | No (passes through) | Yes (OAuth, JWT, API keys) |

**Decision:** Application Gateway is infrastructure (routing and security). API Management is a product (API governance, developer experience, rate limiting). They are often used together: Application Gateway in front of API Management for WAF protection.

### When Application Gateway is the Right Choice

Application Gateway is the sweet spot when:
1. You have a **single-region** deployment with HTTP/HTTPS traffic
2. You need **SSL termination** at the edge of your network
3. You want **host-based routing** to multiple backend services behind a single public IP
4. You need **WAF protection** against web exploits
5. Your backends are VMs, containers, App Services, or any HTTP endpoint

It's the standard L7 ingress point in Azure's hub-and-spoke network architecture, often deployed in the hub VNet as the public-facing entry point for spoke workloads.

## V1 vs V2 SKU: Why Only V2 is Supported

### V1 SKUs (Standard, WAF) -- Legacy

Application Gateway V1 was the original offering:
- Fixed instance sizes (Small, Medium, Large)
- No autoscaling
- No zone redundancy
- Limited to ~100 connections per instance
- Manual scaling requires redeployment
- No support for header rewrite rules

### V2 SKUs (Standard_v2, WAF_v2) -- Current

Application Gateway V2 was a ground-up redesign:
- **Autoscaling**: 0-125 instances based on traffic
- **Zone redundancy**: instances across Availability Zones
- **Improved performance**: 5x faster SSL offload, 5x connection draining improvement
- **Static VIP**: the public IP doesn't change during restarts
- **Header rewrites**: modify request/response headers
- **Key Vault integration**: native certificate management

### Why Planton Only Supports V2

The decision is straightforward:
1. **V1 is legacy** -- Microsoft recommends V2 for all new deployments
2. **V1 lacks autoscaling** -- the single most important operational feature
3. **V1 lacks zone redundancy** -- cannot meet production HA requirements
4. **V1 is more expensive per-connection** -- fewer connections per instance
5. **V2 has native Key Vault integration** -- the production certificate pattern

Planton validates `sku` against `Standard_v2` and `WAF_v2` only. Attempting to use `Standard` or `WAF` fails validation.

## Standard_v2 vs WAF_v2: Decision Framework

Both V2 SKUs share the same L7 load balancing capabilities. The only difference is WAF:

### Standard_v2
- General-purpose L7 load balancing
- SSL termination, host routing, health probes, autoscaling
- Lower cost (no WAF processing overhead)
- Use when: backends are on a private network with no public internet exposure, or when WAF is handled elsewhere (e.g., Azure Front Door in front)

### WAF_v2
- Everything in Standard_v2 plus Web Application Firewall
- OWASP 3.2 rule set (configurable)
- Detection mode (log only) or Prevention mode (block + log)
- Use when: the Application Gateway is the public-facing entry point and you need protection against web exploits

### Cost Impact

WAF adds ~20-30% to the per-instance cost. For most production workloads where the Application Gateway faces the internet, the security benefit far outweighs the cost. The Planton spec defaults `waf_enabled` to false but supports both SKUs so teams can choose based on their security requirements.

## Backend HTTP Settings: The Missing Piece

Backend HTTP settings are the most misunderstood part of Application Gateway configuration. They are **not** just "which port to talk to." They define the complete communication contract between the gateway and backend servers.

### What Backend HTTP Settings Control

1. **Port**: Which port to connect to on the backend (e.g., 80, 443, 8080)
2. **Protocol**: Http or Https (determines if gateway-to-backend is encrypted)
3. **Cookie-based affinity**: Whether repeat clients stick to the same backend
4. **Request timeout**: How long to wait for a backend response (1-86400 seconds)
5. **Health probe**: Which custom probe monitors backends using these settings
6. **Host header**: Override the Host header sent to backends

### Common Patterns

**SSL Offload (most common):**
- Frontend: HTTPS (port 443)
- Backend HTTP settings: Http on port 80
- The gateway terminates SSL and talks plain HTTP to backends
- Simplest pattern, lowest backend overhead

**End-to-End SSL:**
- Frontend: HTTPS (port 443)
- Backend HTTP settings: Https on port 443
- Both hops encrypted -- required for compliance-sensitive workloads
- Backend must present a valid certificate

**Multi-Tenant Backends (App Service, etc.):**
- Set `pick_host_name_from_backend_address: true`
- The gateway sets the Host header to the backend's FQDN
- Required for Azure App Service which routes based on Host header

**Custom Host Header:**
- Set `host_name` to a specific value
- Useful when backends expect a particular Host header
- Mutually exclusive with `pick_host_name_from_backend_address`

### Why Backend HTTP Settings Are Separate from Routing Rules

Azure separates these concerns because:
- Multiple routing rules can share the same backend HTTP settings
- You can change how you talk to backends without changing routing logic
- Health probe configuration is per-settings, not per-pool

This is a key architectural insight: routing rules define **what goes where**, while backend HTTP settings define **how to talk to what's there**.

## SSL/TLS Termination Patterns

### Key Vault Integration (Production Pattern)

The recommended approach stores SSL certificates in Azure Key Vault:

1. Upload the PFX certificate to Key Vault
2. Create a user-assigned managed identity
3. Grant the identity GET permission on Key Vault certificates
4. Reference the Key Vault secret ID in the Application Gateway SSL certificate
5. Assign the identity to the Application Gateway via `identity_ids`

**Advantages:**
- Centralized certificate management
- Automatic renewal (when Key Vault auto-renewal is configured)
- No secrets in YAML manifests or IaC state files
- Audit trail via Key Vault access logs

### PFX Upload (Not Supported by Planton)

The alternative approach embeds a base64-encoded PFX certificate directly in the configuration, with a password. This is supported by Azure but **not by Planton** because:

1. Secrets in manifests are a security anti-pattern
2. PFX data bloats the YAML and state files
3. Manual certificate rotation is error-prone
4. No audit trail for certificate access

Planton's `AzureSslCertificate` only exposes `key_vault_secret_id`, enforcing the Key Vault pattern.

## Health Probe Best Practices

### Default Probe vs Custom Probe

Azure provides a default health probe that sends `GET /` to the backend port. This is insufficient for production because:
- `/` may return 200 even when the application is unhealthy
- No check of downstream dependencies (database, cache, etc.)
- Default interval (30s) may be too slow for critical services

### Custom Probe Design

A well-designed health probe endpoint should:

1. **Check downstream dependencies** -- database connectivity, cache availability, required APIs
2. **Return quickly** -- the probe itself should not do heavy work
3. **Use a dedicated path** -- `/health`, `/api/healthz`, `/ready` (not `/`)
4. **Return appropriate status codes** -- 200-399 is healthy, anything else is unhealthy
5. **Not require authentication** -- the gateway sends unauthenticated probe requests

### Interval and Threshold Tuning

| Setting | Default | Production Guidance |
|---------|---------|-------------------|
| `interval` | 30s | 15-30s for critical, 30-60s for non-critical |
| `timeout` | 30s | 10-15s (backend should respond fast to health checks) |
| `unhealthy_threshold` | 3 | 2-3 (balance between fast detection and avoiding flapping) |

**Key insight:** Setting `interval` to 15s with `unhealthy_threshold` to 3 means it takes 45 seconds to detect a failed backend. For faster detection, lower both values -- but at the cost of more probe traffic and potential false positives.

## Subnet Sizing Guidance

Application Gateway v2 requires a **dedicated subnet** with no other resources (no VMs, no other gateways, no NSGs that would block required ports).

### Sizing Formula

Each Application Gateway instance requires one private IP address from the subnet. The subnet needs:
- **Instance IPs**: 1 per instance (max 125 for autoscale)
- **Azure reserved**: 5 addresses per subnet (network, broadcast, gateway, 2 DNS)
- **Buffer**: headroom for scaling

### Recommendations

| Scenario | Subnet CIDR | Usable IPs | Max Instances |
|----------|------------|------------|---------------|
| Production (autoscale) | /24 | 251 | 125 |
| Production (fixed, small) | /26 | 59 | 54 |
| Development | /27 | 27 | 22 |

**The safe default is /24.** It supports the maximum 125 instances with plenty of room. Smaller subnets save IP space but risk running out during traffic spikes.

### NSG Requirements

If an NSG is attached to the Application Gateway subnet, it must allow:
- **Inbound**: ports 65200-65535 (Azure infrastructure communication)
- **Inbound**: client traffic ports (80, 443, or custom)
- **Outbound**: internet access (for CRL checks, Azure telemetry)

Blocking ports 65200-65535 is the #1 cause of "Application Gateway stuck in failed state" support tickets.

## 80/20 Scoping Rationale

### What's Included (the 80%)

The Planton AzureApplicationGatewaySpec captures the configuration that 80% of Application Gateway deployments need:

1. **SKU selection** (Standard_v2, WAF_v2) -- every deployment needs this
2. **Capacity / Autoscale** -- fixed or dynamic instance count
3. **Backend address pools** -- where traffic goes (FQDN and/or IP)
4. **Backend HTTP settings** -- how to talk to backends (port, protocol, affinity, timeout, probes)
5. **HTTP listeners** -- where traffic enters (port, protocol, host name, SSL cert)
6. **Request routing rules** -- connecting listeners to backends (Basic type)
7. **Health probes** -- custom backend health monitoring
8. **SSL certificates** -- Key Vault integration for HTTPS
9. **WAF configuration** -- enable/disable, Detection/Prevention mode
10. **HTTP/2 support** -- performance optimization
11. **Identity integration** -- user-assigned identity for Key Vault access

### What's Excluded (the 20%) and Why

**Path-based routing (URL path maps):**
- Adds significant spec complexity (nested URL path map -> path rules -> backend mappings)
- Host-based routing covers the primary use case (different services on different domains)
- Path-based routing is typically needed for monolith-to-microservice migration patterns
- Planned for v2 enhancement

**Redirect configurations:**
- HTTP-to-HTTPS redirect, external redirect, listener redirect
- Common but adds 3 more sub-resource types with cross-references
- Can be handled at the application level or via a separate redirect rule resource

**Rewrite rules (header/URL rewrite):**
- Modifying request/response headers and URL paths
- Powerful but complex (conditions + actions + rule sets)
- Advanced use case typically needed for legacy backend compatibility

**Custom error pages:**
- Returning custom HTML for 403/502 errors
- Nice to have but rarely critical for initial deployments

**Private Link configuration:**
- Exposing the gateway via Private Endpoint
- Advanced networking scenario for fully private architectures

**Mutual TLS (mTLS):**
- Client certificate authentication
- Enterprise/compliance use case, not common in initial deployments

**Trusted root certificates:**
- Custom CA certificates for end-to-end SSL with self-signed backend certs
- Needed only for end-to-end SSL with non-public CAs

### The Inclusion Principle

The spec includes everything needed to deploy a functional, production-ready Application Gateway that:
- Terminates SSL from Key Vault certificates
- Routes traffic based on host headers to multiple backend pools
- Monitors backend health with custom probes
- Protects against web exploits with WAF
- Scales automatically based on traffic

This is the complete L7 ingress story for most enterprise deployments.

## Infra Chart Integration

### Enterprise Network Foundation

In the enterprise-network-foundation infra chart, AzureApplicationGateway serves as the L7 ingress point:

```
Internet
  │
  ▼
AzurePublicIp (Standard SKU, Static)
  │
  ▼
AzureApplicationGateway (WAF_v2, autoscale)
  │  ├── Listener: HTTPS on 443 (wildcard cert from Key Vault)
  │  ├── WAF: Prevention mode
  │  └── Routes to backend pools
  │
  ▼
Backend services (in spoke VNets)
```

**Resource dependency chain:**
1. `AzureResourceGroup` -- contains all resources
2. `AzureSubnet` -- dedicated /24 subnet for the gateway
3. `AzurePublicIp` -- Standard SKU, Static allocation
4. `AzureUserAssignedIdentity` -- identity for Key Vault access
5. `AzureKeyVault` + certificate -- SSL certificate storage
6. `AzureApplicationGateway` -- references all of the above

**DNS integration:**
The public IP's `ip_address` output (from AzurePublicIp, not from AzureApplicationGateway) is used by AzureDnsRecord to create A records. The Application Gateway itself does not export the public IP because it's a separate resource.

## Corrections from T02 Specification

During the specification development process, 10 corrections were identified and applied to the initial T02 design:

### Correction 1: SKU as String, Not Enum
**Original:** Custom protobuf enum for SKU values.
**Corrected:** Plain string with CEL validation (`"Standard_v2"`, `"WAF_v2"`).
**Rationale:** Provider authenticity -- use Azure's exact API values as strings rather than inventing enum names. Consistent with how other Azure resources handle mode/tier values.

### Correction 2: Protocol Values Match Azure API
**Original:** Lowercase protocol values (`"http"`, `"https"`).
**Corrected:** Azure-native casing (`"Http"`, `"Https"`).
**Rationale:** The Pulumi/Terraform Azure providers expect these exact strings. Using different casing would require a mapping layer.

### Correction 3: WAF Mode as String, Not Enum
**Original:** Custom enum for WAF mode.
**Corrected:** String with CEL validation (`"Detection"`, `"Prevention"`).
**Rationale:** Same provider authenticity reasoning as SKU.

### Correction 4: Cookie Affinity as String, Not Bool
**Original:** Boolean `cookie_based_affinity` field.
**Corrected:** String field with values `"Enabled"` / `"Disabled"`.
**Rationale:** Azure API uses these exact string values. A boolean would require a mapping layer and lose the Azure-native feel.

### Correction 5: Frontend Ports Auto-Derived
**Original:** Explicit `frontend_ports` array in the spec.
**Corrected:** Frontend ports auto-derived from listener definitions by IaC modules.
**Rationale:** Frontend ports are an Azure implementation detail. Every listener needs a port -- having users define them separately and then reference by name is unnecessary indirection. The IaC modules create named port objects (as `"{listener_name}-port"`) automatically.

### Correction 6: Gateway IP Configuration Auto-Derived
**Original:** Explicit `gateway_ip_configuration` in the spec.
**Corrected:** Auto-derived by IaC modules from the resource name and subnet.
**Rationale:** There is exactly one gateway IP configuration per Application Gateway that maps to the subnet. It's deterministic and should be hidden.

### Correction 7: Frontend IP Configuration Auto-Derived
**Original:** Explicit `frontend_ip_configuration` in the spec.
**Corrected:** Auto-derived by IaC modules from the public IP reference.
**Rationale:** Same as gateway IP configuration -- exactly one public frontend IP config, deterministic from the `public_ip_id`.

### Correction 8: Rule Priority is Required
**Original:** Optional priority field on routing rules.
**Corrected:** Required field with range 1-20000.
**Rationale:** Application Gateway v2 (API version 2021-08-01+) requires a unique priority for every routing rule. Omitting it causes deployment failures.

### Correction 9: Health Probe Host is Optional
**Original:** Required `host` field on health probes.
**Corrected:** Optional field (defaults to backend IP/FQDN when empty).
**Rationale:** Most probes work fine with the backend's own address as the Host header. The `host` override is only needed when backends require a specific Host header for routing.

### Correction 10: Public IP is Not Exported
**Original:** `public_ip_address` in stack outputs.
**Corrected:** Removed from outputs; only `app_gateway_id` and `app_gateway_name` exported.
**Rationale:** The public IP is a separate AzurePublicIp resource with its own outputs. Duplicating it in the Application Gateway outputs creates ambiguity about the source of truth. DNS records should reference the AzurePublicIp's `ip_address` output directly.

## Capacity Planning

### Fixed Capacity vs Autoscale

**Fixed capacity** (`capacity` field):
- Predictable cost (you always pay for N instances)
- Suitable when traffic patterns are well-known and stable
- Simpler to reason about
- Default: 2 instances

**Autoscale** (`autoscale` field):
- Cost-efficient (scale down during low traffic)
- Handles traffic spikes automatically
- Recommended for production workloads with variable traffic
- `min_capacity`: minimum instances (0 for scale-to-zero, 2+ for production HA)
- `max_capacity`: cap to control costs (2-125)

**Guidance:**
- Production: use autoscale with `min_capacity: 2` and `max_capacity` based on expected peak
- Development: use `capacity: 1` for cost savings (no HA, acceptable for dev)
- Cost-sensitive production: use autoscale with `min_capacity: 2` and conservative `max_capacity`

### Instance Sizing

Each Application Gateway v2 instance provides approximately:
- 10 Gbps throughput
- ~200 concurrent persistent connections per compute unit
- Multiple compute units per instance

For most web workloads, 2-5 instances with autoscale handles significant traffic. Instances scale horizontally, not vertically -- more instances handle more traffic linearly.

## Security Considerations

### WAF Rule Sets

WAF_v2 uses OWASP Core Rule Set (CRS) 3.2 by default:
- **SQL injection** detection and prevention
- **Cross-site scripting (XSS)** blocking
- **Remote code execution** prevention
- **Local file inclusion** blocking
- **HTTP protocol violations** detection

### Detection vs Prevention Mode

**Detection mode:**
- Logs all WAF rule matches
- Does **not** block traffic
- Use during initial deployment to identify false positives
- Review logs, create exclusions for legitimate traffic
- Then switch to Prevention

**Prevention mode:**
- Blocks requests that match WAF rules
- Logs blocked requests
- Production default
- May require rule exclusions for APIs that legitimately use patterns that match WAF rules (e.g., SQL-like syntax in query parameters)

### Identity and Key Vault

The Application Gateway's identity chain for SSL certificates:
1. Create `AzureUserAssignedIdentity`
2. Grant the identity `Key Vault Secrets User` role on the Key Vault (or GET permission via access policy)
3. Assign the identity to the Application Gateway via `identity_ids`
4. Reference certificates via `key_vault_secret_id`

This avoids storing any certificate material in YAML manifests or IaC state.

## Troubleshooting Common Issues

### "Application Gateway stuck in Failed/Updating state"
**Cause:** NSG blocking ports 65200-65535 on the gateway subnet.
**Fix:** Add inbound allow rule for ports 65200-65535 from GatewayManager service tag.

### "Backend health shows Unknown"
**Cause:** Health probe cannot reach backends (NSG, routing, backend not listening).
**Fix:** Verify backend is accessible from the gateway subnet. Check probe path returns 200-399.

### "SSL certificate error during provisioning"
**Cause:** Identity doesn't have GET permission on Key Vault, or certificate doesn't exist.
**Fix:** Verify Key Vault access policy / RBAC. Verify the `key_vault_secret_id` URL is correct.

### "Provisioning takes 20+ minutes"
**Expected:** Application Gateway v2 provisioning is inherently slow (10-20 minutes). Plan for this in CI/CD pipelines. Updates to existing gateways (adding listeners, rules) are faster (~5 minutes).

## Conclusion

Azure Application Gateway fills the Layer 7 load balancing niche in Azure's networking stack. It's the right tool when you need HTTP-aware routing, SSL termination, and optional WAF protection for regional deployments.

The Planton spec captures the complete configuration surface for production L7 ingress: SKU selection, backend pools with custom HTTP settings and health probes, host-based routing via listeners and rules, SSL from Key Vault, WAF protection, and autoscaling. The auto-derivation of gateway IP, frontend IP, and frontend port configurations eliminates Azure's internal naming complexity while preserving full control over the routing topology.

The excluded features (path-based routing, redirects, rewrite rules, mTLS) serve the remaining 20% of use cases and are planned for future versions. The current spec handles the enterprise-network-foundation infra chart's L7 ingress requirements completely.
