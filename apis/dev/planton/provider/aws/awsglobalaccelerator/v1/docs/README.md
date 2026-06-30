# AWS Global Accelerator: Architecture, Trade-offs, and Design Rationale

## Introduction

AWS Global Accelerator is a networking service that improves the availability and performance of applications by routing traffic through the AWS global backbone network instead of the public internet. It provides two static anycast IPv4 addresses that serve as fixed entry points to applications hosted in one or more AWS regions.

This document provides the research foundation behind the Planton `AwsGlobalAccelerator` resource design. It covers the underlying architecture, trade-offs against alternatives, cost model, service limits, and the rationale for scoping v1 to standard accelerators.

## Architecture

### How Global Accelerator Works

Global Accelerator operates a network of anycast IP addresses advertised from AWS edge locations worldwide. The traffic flow is:

```
Client → Nearest AWS Edge Location → AWS Global Backbone → Regional Endpoint Group → Healthy Endpoint
```

1. **Client sends traffic to a static anycast IP**: The accelerator's two static IPs are advertised via BGP from all AWS edge locations. Any client in the world resolving or connecting to these IPs reaches the nearest edge location through standard internet routing.

2. **Edge ingress**: Traffic enters the AWS network at the geographically nearest edge location. AWS operates 100+ edge locations across 50+ cities worldwide (shared infrastructure with CloudFront).

3. **AWS backbone transit**: Instead of traversing the public internet (with its unpredictable routing, congestion, and packet loss), traffic travels over AWS's private global fiber network. This backbone provides consistent low latency, low jitter, and high throughput.

4. **Regional endpoint selection**: Global Accelerator evaluates endpoint health and proximity to route traffic to the optimal regional endpoint group. If the nearest region is unhealthy, traffic is automatically rerouted to the next-nearest healthy region.

5. **Endpoint delivery**: Within the selected region, traffic reaches the endpoint (ALB, NLB, EIP, or EC2 instance) based on endpoint weights.

### Anycast Routing

Anycast is the key networking concept behind Global Accelerator. In anycast routing, the same IP address is announced from multiple locations. Internet routers automatically direct traffic to the nearest announcement point based on BGP path selection.

Benefits of anycast for Global Accelerator:
- **No DNS dependency**: Failover doesn't depend on DNS TTL propagation. When an edge location or region fails, BGP converges within seconds and traffic automatically routes to the next-nearest edge.
- **Static entry point**: The two IPs never change, enabling IP-based whitelisting in firewalls, partner integrations, and regulatory environments where DNS-based resolution is unacceptable.
- **DDoS absorption**: Anycast naturally distributes attack traffic across all edge locations rather than concentrating it on a single origin.

### Resource Hierarchy

```
Global Accelerator (1)
├── Listener (1..N per accelerator)
│   ├── Port Ranges (1..10 per listener)
│   └── Endpoint Group (1..N per listener)
│       ├── Health Check Configuration
│       ├── Traffic Dial Percentage
│       ├── Port Overrides (0..10)
│       └── Endpoint (0..N per group)
│           ├── Endpoint ID (ALB ARN, NLB ARN, EIP, EC2 ID)
│           ├── Weight (0-255)
│           └── Client IP Preservation
```

This hierarchy is why the Planton resource bundles all three levels into a single manifest. An accelerator without listeners serves no purpose. A listener without endpoint groups drops all traffic. Separating them would force users to manage three resources for what is conceptually one deployment unit.

## Standard vs Custom Routing Accelerators

AWS offers two accelerator types:

### Standard Accelerator (covered in v1)

Standard accelerators perform **health-based, proximity-based routing**. Global Accelerator determines the optimal endpoint based on:
1. Geographic proximity of the client to edge locations
2. Health status of endpoints (unhealthy endpoints are bypassed)
3. Traffic dial percentage (for manual traffic distribution)
4. Endpoint weights (for load distribution within a group)

Standard accelerators account for ~95% of Global Accelerator deployments. They're the right choice for any workload that needs global load balancing with automatic failover.

### Custom Routing Accelerator (excluded from v1)

Custom routing accelerators provide **deterministic, port-based routing** to specific EC2 instances. Instead of letting Global Accelerator choose the endpoint, the application controls which specific instance receives each connection by mapping listener port ranges to specific instance-port combinations.

Use case: Multiplayer gaming platforms that need to route a player to a specific game server instance. The game's matchmaking service assigns a player to a specific port on the accelerator, which deterministically maps to a specific EC2 instance.

**Why excluded from v1**: Custom routing accelerators have a fundamentally different API surface (subnet mappings instead of endpoint groups, port range mappings, VPC subnet requirements). They represent approximately 5% of Global Accelerator usage and would double the API surface area. They are better served as a separate `AwsGlobalAcceleratorCustomRouting` resource kind in v2.

## Cost Model

Global Accelerator pricing has two components:

### Fixed Cost

**$0.025 per hour per accelerator** (~$18/month). This covers the two static anycast IPs and the anycast routing infrastructure. The cost applies whether or not traffic is flowing.

Key implication: An idle accelerator still costs $18/month. Disabled accelerators (`enabled: false`) still incur this charge — the IPs are still allocated, they just don't resolve.

### Data Transfer Premium (DTC)

**Dominant Transfer Cost (DTC)** is charged per GB of data processed by the accelerator, varying by source region of the client:

| Client Region | Cost per GB |
|---------------|-------------|
| North America, Europe | $0.015 |
| Asia Pacific, Middle East, South Africa | $0.035 |
| Australia, India, Japan | $0.050 |
| South America | $0.070 |

This is **in addition to** the standard AWS data transfer charges for the underlying endpoints (ALB, NLB, etc.).

### Cost Comparison

For a typical production workload processing 1 TB/month with mostly North American clients:

| Component | Monthly Cost |
|-----------|-------------|
| Fixed hourly | ~$18 |
| DTC (1 TB × $0.015) | ~$15 |
| **Total GA premium** | **~$33/month** |

Compare this to:
- CloudFront: $0.085/GB for first 10 TB (but includes caching, so effective cost is often lower)
- Route53 latency routing: $0.50/million queries + health checks ($0.50-$0.75/endpoint/month), no data transfer premium

Global Accelerator is cost-effective when the performance improvement (reduced latency, instant failover) justifies the $33+/month premium over DNS-based alternatives.

## Health Check Behavior

Global Accelerator health checks have specific constraints that differ from ALB or Route53 health checks:

### Interval Constraint

AWS only supports **two health check intervals**: 10 seconds or 30 seconds. This is a hard API constraint — any other value is rejected. This is unusual compared to ALB target group health checks (which support 5–300 seconds in 1-second increments).

### Protocol Support

- **TCP**: Verifies that the endpoint's port is reachable (SYN/ACK). The fastest check but provides no application-level health validation.
- **HTTP**: Sends a GET request to the configured path and expects a 200 status code. Validates application-level health.
- **HTTPS**: Same as HTTP but over TLS. Useful when endpoints only accept encrypted traffic.

**No UDP health checks**: Even for UDP listeners, health checks must use TCP, HTTP, or HTTPS. A common pattern is to run a sidecar HTTP health endpoint on game servers or UDP services.

### Failover Timing

With a 10-second interval and threshold of 3:
- **Time to detect failure**: 30 seconds (3 × 10s intervals)
- **Time to failover**: 30 seconds + BGP convergence (~seconds) ≈ **~35 seconds total**

With a 30-second interval and threshold of 3:
- **Time to detect failure**: 90 seconds
- **Time to failover**: ~95 seconds total

Compare to Route53 health checks (minimum 10-second interval, failover depends on DNS TTL, typically 60–300 seconds).

### Health Check Source IPs

Global Accelerator health checks originate from the AWS edge network, not from the endpoint's region. Security groups on endpoints must allow health check traffic from the Global Accelerator IP ranges (published in the AWS IP ranges JSON file under the `GLOBALACCELERATOR` service).

## BYOIP (Bring Your Own IP)

BYOIP allows organizations to use their own IPv4 address ranges with Global Accelerator instead of AWS-allocated anycast IPs.

### Use Cases

- **IP reputation**: Organizations that have built IP reputation (email deliverability, API integrations) and cannot change IPs.
- **Regulatory requirements**: Financial services, government, and healthcare organizations required to use addresses from their own allocated ranges.
- **Contractual obligations**: Partners or customers that have whitelisted specific IP ranges in their firewalls.
- **Migration**: Gradually moving traffic from on-premises infrastructure to AWS while maintaining the same IP addresses.

### Constraints

- Maximum 2 BYOIP addresses per accelerator.
- The IP range must be registered with AWS through the BYOIP onboarding process (requires ROA from the RIR).
- Changing BYOIP addresses is a **ForceNew** operation — the accelerator is destroyed and recreated.
- BYOIP addresses must be from a `/24` CIDR or larger that has been provisioned for use with Global Accelerator.

## Comparison with Alternatives

### Global Accelerator vs CloudFront

| Dimension | Global Accelerator | CloudFront |
|-----------|-------------------|------------|
| **OSI Layer** | Layer 4 (TCP/UDP) | Layer 7 (HTTP/HTTPS) |
| **Protocol support** | TCP, UDP | HTTP, HTTPS, WebSocket |
| **Caching** | No caching | Edge caching (primary feature) |
| **Static IPs** | Yes (2 anycast IPs) | No (uses DNS-based routing) |
| **Failover speed** | Seconds (BGP-based) | Minutes (DNS TTL-based) |
| **Use case** | Non-HTTP protocols, static IPs, instant failover | HTTP content delivery, caching, Lambda@Edge |
| **Cost model** | Fixed hourly + DTC/GB | Per-request + per-GB (no fixed cost) |

**Decision guidance**: Use CloudFront for HTTP workloads that benefit from caching. Use Global Accelerator for non-HTTP workloads, when static IPs are required, or when sub-minute failover is critical.

### Global Accelerator vs Route53 Latency-Based Routing

| Dimension | Global Accelerator | Route53 Latency Routing |
|-----------|-------------------|------------------------|
| **Routing mechanism** | Anycast BGP | DNS resolution |
| **Failover speed** | Seconds | Depends on DNS TTL (60–300s typical) |
| **Static IPs** | Yes | No |
| **Protocol** | Any TCP/UDP | Any (DNS is transparent) |
| **Health checks** | Built-in, 10/30s intervals | Separate Route53 health checks |
| **Cost** | $18/month + DTC | ~$1–2/month (queries + health checks) |
| **Traffic path** | AWS backbone end-to-end | Public internet after DNS resolution |

**Decision guidance**: Use Route53 latency routing when cost sensitivity is high and DNS-based failover timing is acceptable. Use Global Accelerator when you need static IPs, instant failover, or the performance benefits of the AWS backbone.

### When to Combine Them

A common production architecture combines all three:
- **Route53** for DNS management and domain routing
- **Global Accelerator** for static IP entry points and instant failover
- **CloudFront** as the endpoint behind Global Accelerator for HTTP caching

This architecture provides static IPs → instant failover → edge caching → origin ALBs.

## AWS Service Limits

| Resource | Default Limit | Adjustable |
|----------|--------------|------------|
| Accelerators per account | 20 | Yes (via support request) |
| Listeners per accelerator | 10 | No |
| Port ranges per listener | 10 | No |
| Endpoint groups per listener | Equals the number of AWS regions | No |
| Endpoints per endpoint group | 10 | Yes (up to 255 via support) |
| Port overrides per endpoint group | 10 | No |
| BYOIP addresses per accelerator | 2 | No |
| Static IPs per accelerator | 2 (IPv4) or 4 (dual-stack) | No |

### Important Limit Interactions

- **One endpoint group per region per listener**: You cannot have two endpoint groups in the same region within the same listener. To serve more endpoints in a region, add them to the single endpoint group (up to the 10/255 endpoint limit).
- **Listener port ranges cannot overlap**: Within an accelerator, no two listeners can have overlapping port ranges for the same protocol.
- **Cross-account endpoints**: Endpoints must be in the same AWS account as the accelerator (cross-account support is not available as of v1 design).

## Planton Design Decisions

### Why Bundle the Full Hierarchy

An accelerator without listeners is a pair of static IPs doing nothing. A listener without endpoint groups drops all traffic. The three-level hierarchy (accelerator → listener → endpoint group) is always deployed together and managed as a unit. Splitting them into three Planton resources would:
1. Force users to manage three manifests for one logical unit
2. Create ordering dependencies (accelerator must exist before listener, listener before endpoint group)
3. Provide no practical benefit since the levels are never independently useful

### Why names Instead of Auto-Generated Keys

Listeners and endpoint groups use user-provided `name` fields as keys in output maps (`listener_arns`, `endpoint_group_arns`). This design enables stable `valueFrom` references. If outputs were keyed by index or auto-generated IDs, downstream resources would break when the list order changes.

### Why String Enums Instead of Proto Enums

Fields like `protocol`, `client_affinity`, and `health_check_protocol` use validated strings rather than protobuf enums. This follows the Planton convention of using CEL validation on string fields, which:
1. Produces clearer error messages ("protocol must be 'TCP' or 'UDP'" vs "invalid enum value 3")
2. Serializes naturally in YAML (`protocol: TCP` instead of `protocol: 1`)
3. Avoids protobuf's zero-value enum behavior (where the default value is the first enum entry, not necessarily the desired default)

## v2 Roadmap

The following features are candidates for future versions of the AwsGlobalAccelerator resource:

### Custom Routing Accelerators

A separate `AwsGlobalAcceleratorCustomRouting` resource kind for deterministic port-based routing to specific EC2 instances. This requires a fundamentally different spec structure with subnet mappings and port range definitions.

### Cross-Account Endpoint Support

When AWS adds cross-account endpoint support, the spec will need to accept endpoint ARNs from other accounts and handle the necessary cross-account permissions.

### WAF Integration

Global Accelerator doesn't natively support AWS WAF (which is Layer 7), but when combined with ALB endpoints, WAF can be attached to the ALBs. A future version could provide a convenience field that automatically attaches a WAF Web ACL to all ALB endpoints.

### Custom DNS Integration

First-class Route53 integration similar to the `AwsAlb` resource's `dns` field. This would automatically create Route53 alias records pointing custom domains to the accelerator's DNS name using the well-known hosted zone ID (`Z2BJ6XQ5FK7U4H`).

### Tagging Support

AWS Global Accelerator supports resource tagging. A future version should expose a `tags` field for cost allocation, access control, and organizational metadata.

### Multi-Protocol Listeners

A convenience pattern for services that need both TCP and UDP on the same ports (e.g., DNS servers on port 53). Currently requires two separate listener entries; a future version could add a `TCP_UDP` protocol option if AWS adds native support.

## References

- [AWS Global Accelerator Developer Guide](https://docs.aws.amazon.com/global-accelerator/latest/dg/what-is-global-accelerator.html)
- [Global Accelerator FAQ](https://aws.amazon.com/global-accelerator/faqs/)
- [Global Accelerator Pricing](https://aws.amazon.com/global-accelerator/pricing/)
- [AWS Global Network](https://aws.amazon.com/about-aws/global-infrastructure/global_network/)
- [Global Accelerator API Reference](https://docs.aws.amazon.com/global-accelerator/latest/api/Welcome.html)
- [AWS IP Address Ranges (for health check source IPs)](https://ip-ranges.amazonaws.com/ip-ranges.json)
- [BYOIP for Global Accelerator](https://docs.aws.amazon.com/global-accelerator/latest/dg/using-byoip.html)
- [Standard vs Custom Routing](https://docs.aws.amazon.com/global-accelerator/latest/dg/introduction-how-it-works.html)
