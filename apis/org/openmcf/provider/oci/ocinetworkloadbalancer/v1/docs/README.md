# OCI Network Load Balancer: Design Rationale and Research

## Introduction

The OciNetworkLoadBalancer component manages OCI's Layer 4 load balancer — the transport-layer counterpart to OciApplicationLoadBalancer (Layer 7). Where the L7 load balancer handles HTTP/HTTPS traffic with SSL termination, hostname routing, and rule sets, the NLB handles raw TCP/UDP traffic with source IP preservation and fully elastic bandwidth. This document explains the design decisions that shaped the NLB component.

## Why a Separate NLB Component

OCI has two distinct load balancer products:

1. **Application Load Balancer** (Layer 7) — `oci_load_balancer_load_balancer`
2. **Network Load Balancer** (Layer 4) — `oci_network_load_balancer_network_load_balancer`

These are entirely separate OCI services with different APIs, different Terraform/Pulumi resource types, different capabilities, and different pricing models. They share no configuration, no backends, and no listeners. Making them a single component would force a "mode" switch pattern that combines two unrelated APIs into one proto — increasing complexity without any ergonomic benefit.

The separation mirrors the OCI API boundary and matches what platform engineers expect from Terraform experience: `oci_load_balancer_*` resources vs `oci_network_load_balancer_*` resources.

## Why Bundle Backend Sets, Backends, and Listeners

The NLB component bundles all sub-resources into a single manifest, following the same philosophy as OciApplicationLoadBalancer. The rationale is identical:

1. **Backend sets are scoped to the NLB.** The Terraform resource `oci_network_load_balancer_backend_set` requires `network_load_balancer_id`. A backend set cannot exist independently.

2. **Listeners reference backend sets by name.** A listener's `default_backend_set_name` must match a backend set within the same NLB. Splitting listeners into a separate component would create fragile cross-resource name coupling.

3. **Atomic deployment is the common case.** Engineers deploying an NLB configure the full stack (NLB + backends + listeners) in one pass. The "I need a TCP load balancer for these backends" scenario is one manifest, one command.

4. **The Pulumi module manages creation ordering.** Backend sets are created first, then listeners depend on them via explicit `DependsOn` relationships. This ordering is an implementation detail that users should not manage.

## NLB vs L7 Load Balancer: Spec Complexity

The NLB spec is significantly simpler than the L7 load balancer:

| Aspect | NLB | L7 LB |
|--------|-----|-------|
| **Nested message types** | 6 | 17 |
| **Enums** | 3 | 3 |
| **Spec fields (top-level)** | 14 | ~20 |
| **SSL/TLS support** | None | Full (certificates, cipher suites) |
| **Routing complexity** | Port + protocol | URL path, hostname, headers, rule sets |

This simplicity is intentional — the NLB does one thing (L4 traffic distribution) and does it well. The spec reflects the focused nature of the underlying OCI service.

## Policy Enum: Tuple-Based Hashing

The NLB uses tuple-based hashing rather than the round-robin and least-connections policies available on the L7 load balancer. The proto enum uses lowercase values:

```
five_tuple  → FIVE_TUPLE  (src IP, src port, dst IP, dst port, protocol)
three_tuple → THREE_TUPLE (src IP, dst IP, protocol)
two_tuple   → TWO_TUPLE   (src IP, dst IP)
```

The Go module maps enum values to the uppercase strings expected by the OCI API:

```go
var policyMap = map[...Policy]string{
    five_tuple:  "FIVE_TUPLE",
    three_tuple: "THREE_TUPLE",
    two_tuple:   "TWO_TUPLE",
}
```

The three policies offer a spectrum of session stickiness: five-tuple is the most granular (different connections from the same client may go to different backends), while two-tuple is the stickiest (all traffic from a client IP goes to the same backend regardless of port or protocol).

## Health Check Protocol Breadth

The NLB supports 5 health check protocols (HTTP, HTTPS, TCP, UDP, DNS), compared to the L7 LB's 2 (HTTP, TCP). This broader set reflects the NLB's diverse backend types:

- **HTTP/HTTPS** — application-level health for web services behind an NLB
- **TCP** — port reachability for generic TCP services
- **UDP** — payload-based health checking for UDP backends (syslog, RADIUS, custom protocols)
- **DNS** — sends actual DNS queries and validates response codes, critical for DNS server backends

The DNS health check is particularly notable. A TCP port check on port 53 only confirms the DNS server is listening — it does not validate that the server can resolve queries. The DNS health check sends an A/AAAA/TXT query to a configurable domain and checks the response code (NOERROR, NXDOMAIN, etc.), providing true functional health validation.

## Source IP Preservation and Transparent Mode

The NLB's killer feature is source IP preservation via `isPreserveSourceDestination`. When enabled:

1. The NLB automatically sets `skipSourceDestinationCheck` on its VNIC
2. Packets arrive at backends with the original client IP and destination IP intact
3. Backends (firewalls, IDS/IPS) see the true packet headers

Combined with `isSymmetricHashEnabled`, the NLB operates in "transparent mode" — a bump-in-the-wire topology where:

- Forward traffic: Client → NLB → Firewall (client IP preserved)
- Return traffic: Firewall → NLB → Client (hashed to the same firewall)

Symmetric hashing ensures that the forward and return paths for a given flow transit the same backend. Without symmetric hashing, return traffic might be sent to a different firewall VM, which would not have the connection state and would drop the packet. This is why `isSymmetricHashEnabled` requires `isPreserveSourceDestination` — symmetric hashing is only meaningful in transparent mode.

## Failover Design: Three Complementary Mechanisms

The NLB offers three failover mechanisms that can be combined:

1. **Instant Failover** (`isInstantFailoverEnabled`): When a backend becomes unhealthy, existing connections are immediately moved to a healthy backend rather than waiting for the TCP connection to time out. This reduces failover time from minutes (TCP timeout) to milliseconds (health check detection).

2. **TCP RST on Failover** (`isInstantFailoverTcpResetEnabled`): Instead of silently migrating connections, the NLB sends a TCP RST to the client. The client's TCP stack immediately detects the broken connection and can reconnect. This is faster than silent failover because the client does not wait for a read timeout.

3. **Fail-Open** (`isFailOpen`): When all backends are unhealthy, the NLB continues distributing traffic to all backends rather than returning errors. This is a deliberate trade-off: degraded service is better than no service. Critical for scenarios where health check failures may be transient (network blip) but backends are still functional.

These three mechanisms address different failure modes and can be combined per-backend-set for fine-grained control.

## What's Excluded and Why

### SSL Termination

The NLB does not support SSL termination. This is a fundamental characteristic of Layer 4 load balancing — the NLB forwards raw TCP/UDP packets without inspecting or modifying application-layer content. For SSL termination, use OciApplicationLoadBalancer (Layer 7).

### Content-Based Routing

No URL path routing, hostname matching, or header-based rules. The NLB routes solely based on listener port and protocol. All traffic on a given listener goes to one backend set.

### Multiple Subnets

The NLB deploys into a single subnet, unlike the L7 load balancer which can span multiple subnets for regional availability. This is a limitation of the OCI NLB service — the single-subnet design simplifies networking but limits availability zone distribution.

## What's Deferred

- **Defined Tags** — OCI defined tags require a pre-created tag namespace. Freeform tags (from OpenMCF labels) cover the majority of tagging use cases.
- **Reserved IP Management** — The NLB accepts pre-created reserved IP OCIDs but does not create them. Reserved IPs are managed by the OciPublicIp component.
- **Backend Set Policies Beyond Tuples** — OCI currently supports only tuple-based policies for the NLB. If new policies are added, the enum will be extended.

## Research Notes

### NLB vs L7 LB Performance Characteristics

| Metric | NLB | L7 LB |
|--------|-----|-------|
| Latency | Microseconds (no application inspection) | Milliseconds (HTTP parsing, SSL handshake) |
| Bandwidth | Fully elastic | Shape-dependent (10 Mbps - 8 Gbps flexible) |
| Connections/sec | Millions (wire-speed forwarding) | Thousands (application processing) |
| Source IP | Preserved in packet header | Available via X-Forwarded-For header only |

### Listener Port 0 and Protocol ANY

When `port` is set to 0 and `protocol` is `any`, the NLB acts as a catch-all for all IP traffic. This is the transparent mode configuration used for firewall appliances. Similarly, backends with `port: 0` accept traffic on whatever port the original packet was destined for.

### Proxy Protocol v2

PPv2 adds a binary header to the beginning of the TCP connection containing:
- Source IP and port
- Destination IP and port
- Protocol family (IPv4/IPv6)

This is useful when `isPreserveSourceDestination` cannot be used (e.g., when the NLB is not in the direct traffic path) but backends still need client identity. Popular backends like HAProxy, nginx, and Envoy support PPv2 natively.

### NLB Limits

| Resource | Limit | Notes |
|----------|-------|-------|
| NLBs per compartment | 50 (default) | Adjustable via service limit request |
| Backend sets per NLB | 1024 | |
| Backends per backend set | 512 | |
| Listeners per NLB | 256 | |
