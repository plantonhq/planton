# AWS Network Load Balancer: Architecture and Design Deep-Dive

This document provides an architecture-focused research and design reference for the AwsNetworkLoadBalancer Planton component. It covers Layer 4 load balancing fundamentals, NLB-specific behavior, operational constraints, and production patterns.

---

## Layer 4 vs Layer 7 Load Balancing

### OSI Model Context

Load balancers operate at different layers of the OSI model:

| Layer | Name | What It Inspects | Examples |
|-------|------|------------------|----------|
| **Layer 4** | Transport | IP, port, protocol (TCP/UDP) | NLB, HAProxy (TCP mode) |
| **Layer 7** | Application | HTTP headers, path, host, body | ALB, Nginx, HAProxy (HTTP mode) |

**NLB is Layer 4**: It forwards connections based solely on destination port and protocol. It does not inspect HTTP headers, paths, or request content. A client connecting to NLB on port 443 gets forwarded to a target's port—the NLB does not care whether the traffic is HTTPS, TLS-wrapped gRPC, or raw TCP.

**ALB is Layer 7**: It parses HTTP/HTTPS and can route based on path (`/api/*`), host (`api.example.com`), headers, or query strings. It supports redirects, fixed responses, and authentication actions.

### Implications for NLB

- **No path-based routing**: Cannot route `/api` to one target group and `/web` to another. Use a single target group per listener or put an ALB behind the NLB.
- **No HTTP-level features**: No redirects, no WAF integration at the NLB level, no Lambda targets.
- **Ultra-low latency**: No HTTP parsing means minimal processing overhead. NLB can handle millions of connections per second with sub-millisecond latency.
- **Protocol flexibility**: TCP, UDP, TLS, and TCP_UDP. Ideal for databases, game servers, IoT, custom protocols.

---

## Static IP Architecture (EIP Allocation per AZ)

### Why Static IPs Matter

ALBs receive a dynamic DNS name from AWS. The underlying IPs can change during scaling or maintenance. For many use cases—partner allowlisting, firewall rules, legacy integrations, DNS pinning—**static IPs are mandatory**.

### NLB Static IP Model

For **internet-facing** NLBs, you can assign one Elastic IP (EIP) per subnet mapping. Each subnet typically corresponds to one Availability Zone. The result:

- **AZ-1**: NLB node with static public IP `203.0.113.1`
- **AZ-2**: NLB node with static public IP `203.0.113.2`
- **AZ-3**: NLB node with static public IP `203.0.113.3` (if you add a third subnet)

These IPs **do not change** across NLB scaling events, instance replacements, or AWS maintenance. Clients can allowlist them permanently.

### Allocation Flow

1. Allocate EIPs (one per AZ) via `ec2 allocate-address` or an AwsElasticIp resource.
2. In the NLB subnet mapping, set `allocationId` to each EIP's allocation ID.
3. AWS attaches the EIP to the NLB node in that subnet.

### Internal NLBs: Private IP Pinning

For **internal** NLBs, you can optionally set `privateIpv4Address` per subnet mapping to pin the NLB node to a specific private IP within the subnet's CIDR. When omitted, AWS assigns a private IP automatically. Pinning is useful when downstream systems reference the NLB by IP.

---

## Cross-Zone Load Balancing Behavior

### Default: Disabled

Unlike ALB (where cross-zone is always on), NLB **defaults to cross-zone load balancing disabled**. This means:

- Traffic from clients in AZ-1 is routed **only** to targets in AZ-1.
- If AZ-1 has 2 targets and AZ-2 has 10 targets, AZ-1 targets receive roughly half the traffic from AZ-1 clients, regardless of target count.

### When to Enable

Set `crossZoneLoadBalancingEnabled: true` when:

- Target distribution across AZs is uneven and you want traffic proportional to target capacity.
- You need consistent per-target traffic distribution regardless of client AZ.

### Cost Consideration

With cross-zone disabled, traffic stays within the same AZ—no cross-AZ data transfer for NLB-to-target traffic. With cross-zone enabled, traffic may flow from an NLB node in AZ-1 to a target in AZ-2, incurring cross-AZ data transfer charges.

---

## Client IP Preservation: preserve_client_ip vs proxy_protocol_v2

### The Problem

By default, targets behind an NLB see the NLB's private IP as the source. Applications that need the original client IP (for logging, geo-routing, rate limiting, security) must use one of two mechanisms.

### preserve_client_ip

When enabled, the NLB preserves the client's source IP in the IP header. Targets see the real client IP. This works for:

- **Instance targets**: Enabled by default.
- **IP targets**: Disabled by default (must be explicitly enabled).

**Limitation**: Only the IP is preserved. Port, protocol, and other connection metadata are not passed.

### proxy_protocol_v2

Proxy Protocol v2 is a header prepended to the connection that carries:

- Source IP and port
- Destination IP and port
- VPC endpoint ID (for PrivateLink)
- Connection metadata

**Requirements**:

- Targets must be configured to **parse** the Proxy Protocol header (e.g., Nginx `proxy_protocol` directive, HAProxy native support).
- If the target does not expect Proxy Protocol, it will misinterpret the first bytes of the connection as application data and fail.

**Use proxy_protocol_v2 when**: You need full connection metadata (e.g., client port, VPC endpoint ID) or when `preserve_client_ip` alone is insufficient.

**Use preserve_client_ip when**: You only need the client IP and your targets do not support Proxy Protocol.

---

## TLS Termination at NLB vs Pass-Through

### TLS Termination (protocol: TLS)

When the listener uses `protocol: TLS` with a `tls` configuration:

1. Client connects to NLB with TLS (e.g., TLS 1.2/1.3).
2. NLB terminates TLS using the ACM certificate.
3. NLB forwards **plaintext TCP** to targets.

**Benefits**: Offload certificate management and TLS decryption from application servers. Targets receive simple TCP.

**Target protocol**: Typically `TCP` (plaintext). The target group protocol is the NLB-to-target protocol, not the client-to-NLB protocol.

### Pass-Through (protocol: TCP)

When the listener uses `protocol: TCP`:

1. Client connects with raw TCP (or TLS inside TCP—NLB does not inspect).
2. NLB forwards bytes unchanged to targets.
3. Targets must handle TLS termination themselves if the client uses TLS.

**Use pass-through when**: You need end-to-end encryption (client to target) or the target requires the original TLS connection (e.g., for client certificates, mutual TLS).

---

## Health Check Model (TCP vs HTTP/HTTPS)

### TCP Health Checks (Default)

- NLB attempts a TCP connection to the target port.
- If the connection succeeds, the target is healthy.
- **Does not verify** application logic—a frozen app can still accept TCP connections.

### HTTP/HTTPS Health Checks

- NLB sends an HTTP GET to the specified `path`.
- Checks the response status code against `matcher` (e.g., `200-399`).
- **Verifies** that the application responds correctly.

**When to use HTTP/HTTPS**: For application-level health. A dedicated `/healthz` or `/api/health` endpoint that checks DB connectivity, cache, and dependencies is more reliable than TCP-only.

**NLB constraint**: For NLB target groups, `unhealthy_threshold` must equal `healthy_threshold` (unlike ALB). The Planton IaC modules enforce this.

---

## NLB Scaling Behavior (Flow-Based, Not Request-Based)

### Flow-Based Scaling

NLB scales based on **connection flows** (new connections per second, active connections), not HTTP requests. A single long-lived connection (e.g., WebSocket, gRPC stream, database connection) counts as one flow regardless of how many messages are exchanged.

### Implications

- **Long-lived connections**: NLB handles them efficiently. No per-request overhead.
- **Short-lived, high-request workloads**: If each HTTP request is a new connection, NLB scales with connection rate. For HTTP/1.1 with connection reuse, fewer flows.
- **UDP**: Each UDP "flow" is identified by 5-tuple (source IP, source port, dest IP, dest port, protocol). Stateless; no connection tracking in the traditional sense.

---

## Security Group Behavior (Optional, Can't Remove Once Added)

### Optional for NLB

Unlike ALB (where security groups are effectively required to control traffic), NLB can run **without** security groups. When omitted, the NLB accepts all traffic on configured listener ports. This is useful for:

- Internal NLBs where the VPC network ACLs and routing provide sufficient isolation.
- Simplifying configuration when the NLB is in a trusted network segment.

### Immutable Once Added

**Critical constraint**: Once you attach security groups to an NLB, you **cannot remove all of them**. At least one security group must remain. You can replace one SG with another, but you cannot go back to "no security groups."

**Planning**: If you might want to run without SGs initially, do not add them. Add SGs only when you are committed to using them long-term.

---

## Subnet Mapping Constraints (Can Only Add, Not Remove)

### Add-Only Semantics

AWS allows **adding** subnet mappings to an NLB but **does not support removing** them. If you initially deploy with subnets in AZ-1 and AZ-2, you can add AZ-3 later. You cannot remove AZ-1 or AZ-2.

**Planning**: Start with the minimum set of subnets you need. Prefer two for HA; add more only when you are sure you need them.

---

## Connection Draining and Deregistration

### deregistration_delay_seconds

When a target is deregistered (e.g., during a deployment or scale-in), the NLB waits this many seconds before fully removing it. During this period:

- **In-flight connections** are allowed to complete.
- **New connections** are not sent to the draining target.

**Default**: 300 seconds. For long-lived connections (WebSocket, gRPC, DB), ensure your application's graceful shutdown period is **less than** this delay so connections can drain before the target is terminated.

### connection_termination

When enabled, the NLB **actively closes** connections to deregistered targets when the deregistration delay expires, instead of waiting for the client to close. Use this for:

- Long-lived connections that may not close naturally.
- Faster, predictable cleanup during deployments.

---

## DNS and Alias Record Patterns

### NLB DNS Name

AWS assigns a DNS name like `my-nlb-abc123.elb.us-east-1.amazonaws.com`. This name resolves to the NLB's IPs (one per AZ). The exact IPs can change for internet-facing NLBs **without** Elastic IPs; with Elastic IPs, the resolved IPs are stable.

### Route53 Alias Records

When `dns.enabled` is true, Planton creates Route53 **alias** A records pointing your hostnames to the NLB's DNS name. Alias records:

- Work at the zone apex (e.g., `example.com`), unlike CNAME.
- Have no charge for alias queries.
- Can evaluate target health when configured.

### dns_record_client_routing_policy

Controls how DNS resolvers route clients to NLB nodes:

- **any_availability_zone** (default): Client may reach any AZ. Best for general use.
- **availability_zone_affinity**: Client is routed to the AZ of the resolver. Reduces cross-zone traffic; best when targets are evenly distributed.
- **partial_availability_zone_affinity**: 85% stay in resolver's AZ, 15% spill over. Balances affinity with availability.

---

## NLB-in-Front-of-ALB Pattern

### Architecture

```
Internet → NLB (static IPs, Layer 4) → ALB (Layer 7 routing) → Targets
```

### How It Works

1. Create an NLB with static Elastic IPs per AZ.
2. Create an ALB with listeners and target groups as usual.
3. Create an NLB target group with `targetType: alb`.
4. Register the ALB as a target of the NLB target group.
5. Clients connect to NLB's static IPs; NLB forwards to ALB; ALB performs path/host-based routing.

### Use Cases

- **Static IP + Layer 7 routing**: Partners need allowlisted IPs, but you also need path-based routing, WAF, or Lambda targets.
- **Hybrid entry point**: Single NLB with static IPs fronting multiple ALBs (e.g., different path prefixes routed to different ALBs via NLB listeners on different ports).

### Target Group Configuration

```yaml
targetGroup:
  port: 443
  protocol: TCP
  targetType: alb
```

The ALB's DNS name or IP is registered as the target. Ensure the ALB is in the same VPC and that security groups allow NLB → ALB traffic.

---

## Summary: Key Design Decisions

| Topic | NLB Behavior | Recommendation |
|-------|--------------|----------------|
| **Static IPs** | Elastic IP per subnet mapping | Use for allowlisting, firewall rules, DNS pinning |
| **Cross-zone** | Default off | Enable only when target distribution is uneven |
| **Client IP** | preserve_client_ip or proxy_protocol_v2 | Use preserve_client_ip for IP only; proxy_protocol_v2 for full metadata |
| **TLS** | Terminate at NLB or pass-through | Terminate for simplicity; pass-through for end-to-end encryption |
| **Health checks** | TCP, HTTP, or HTTPS | Use HTTP/HTTPS for application-level health |
| **Security groups** | Optional; immutable once added | Add only when committed; cannot remove all |
| **Subnet mappings** | Add-only | Start minimal; add AZs only when needed |
| **Deregistration** | connection_termination for long-lived | Enable for WebSocket, gRPC, DB workloads |
