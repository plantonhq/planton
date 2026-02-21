# Overview

The **OCI Network Load Balancer API Resource** provides a consistent interface for deploying and managing Layer 4 load balancers on Oracle Cloud Infrastructure. Unlike the Layer 7 OciApplicationLoadBalancer (which handles HTTP/HTTPS with SSL termination, hostname routing, and rule sets), the Network Load Balancer operates at the transport layer, distributing TCP, UDP, and mixed-protocol traffic with fully elastic bandwidth and native source IP preservation.

## Purpose

This API resource streamlines the deployment of OCI Network Load Balancers and their associated sub-resources. By bundling the NLB, backend sets, backends, and listeners into a single manifest, it eliminates the multi-resource orchestration that would otherwise be required:

- **Layer 4 Traffic Distribution**: Distribute TCP, UDP, TCP+UDP, or all IP traffic across backend servers using tuple-based hashing policies.
- **Source IP Preservation**: Maintain the original client IP address in packet headers — critical for firewalls, intrusion detection systems, and logging pipelines that must see the true client identity.
- **Transparent Mode**: Deploy NLBs as bump-in-the-wire appliances with symmetric hashing for firewall and security appliance topologies.
- **Elastic Bandwidth**: No shape configuration or bandwidth sizing. The NLB scales automatically to handle traffic, unlike the L7 load balancer which requires shape selection.
- **Infra-Chart Composability**: The NLB OCID and assigned IP addresses are exported as stack outputs for DNS record configuration and monitoring integration.

## Key Features

- **Consistent Interface**: Aligns with the OpenMCF pattern for deploying cloud infrastructure across providers.
- **Tuple-Based Load Balancing**: Three policy options — five-tuple (src IP, src port, dst IP, dst port, protocol), three-tuple (src IP, dst IP, protocol), and two-tuple (src IP, dst IP) — matching traffic stickiness to workload requirements.
- **Comprehensive Health Checking**: Supports HTTP, HTTPS, TCP, UDP, and DNS health check protocols. DNS health checks enable monitoring DNS server backends by sending queries and validating response codes.
- **Advanced Failover**: Instant failover moves existing connections to healthy backends immediately. Fail-open keeps traffic flowing to degraded backends when all are unhealthy. TCP RST signaling gives clients immediate reconnection cues.
- **Active-Standby Support**: Designate backup backends that receive traffic only when all primary backends fail. Combine with operationally-active-preferred routing for active-standby topologies.
- **Proxy Protocol v2**: Optional PPv2 support prepends connection metadata to the TCP stream, allowing backends to see client information when source IP preservation is not possible.
- **IPv6 and Dual-Stack**: Supports IPV4, IPV6, and IPV4_AND_IPV6 configurations with optional static IPv6 address assignment.
- **Automatic Tagging**: Standard OpenMCF freeform tags applied to the NLB for resource tracking, cost allocation, and compliance.

## How OCI NLB Differs from OCI Application Load Balancer

| Aspect | Network Load Balancer (NLB) | Application Load Balancer (L7) |
|--------|---------------------------|-------------------------------|
| **OSI Layer** | Layer 4 (transport) | Layer 7 (application) |
| **Protocols** | TCP, UDP, TCP+UDP, ANY | HTTP, HTTPS, HTTP/2, gRPC, TCP |
| **Bandwidth** | Fully elastic (no shape) | Shape-based (fixed or flexible) |
| **Source IP** | Preserved by default | Not preserved (X-Forwarded-For header) |
| **SSL Termination** | Not supported | Full SSL offload with certificate management |
| **Routing** | Port and protocol only | URL path, hostname, headers, rule sets |
| **Subnets** | Single subnet | Multiple subnets |
| **Health Checks** | HTTP, HTTPS, TCP, UDP, DNS | HTTP, TCP |
| **Use Case** | Raw TCP/UDP, firewalls, DNS, gaming | Web apps, APIs, microservices |

## Critical Constraints

- **Single Subnet**: The NLB deploys into exactly one subnet. Changing the subnet after creation forces resource recreation.
- **Layer 4 Only**: No SSL termination, no content-based routing, no hostname matching. For Layer 7 features, use OciApplicationLoadBalancer.
- **Symmetric Hash Requires Preservation**: `isSymmetricHashEnabled` is only valid when `isPreserveSourceDestination` is also enabled. The NLB must operate in transparent mode for symmetric hashing.
- **Backend Identification**: Each backend needs either an `ipAddress` or a `targetId` (compute instance OCID). When using `targetId`, OCI resolves the IP automatically.
- **Reserved IP Immutability**: Reserved public IPs assigned at creation cannot be changed without recreating the NLB.

## Use Cases

- **TCP Service Load Balancing**: Distribute database connections, gRPC services, or custom TCP protocols across backend servers with consistent hashing for session affinity.
- **Firewall and Security Appliances**: Deploy NLBs in transparent mode with source IP preservation and symmetric hashing as bump-in-the-wire traffic inspection points.
- **DNS Server Load Balancing**: Distribute DNS queries across multiple DNS servers with DNS-based health checks that validate actual query responses.
- **UDP Workloads**: Balance real-time media, gaming, IoT, and syslog traffic where the Layer 7 load balancer's TCP-only protocol support is insufficient.
- **Internal Service Mesh Entry Point**: Private NLBs as internal entry points for service-to-service communication where low latency and source IP visibility are required.

## Production Features

This resource provides complete support for production-grade NLB deployments, including:

- **Instant Failover**: Immediate connection migration to healthy backends with optional TCP RST signaling for fast client-side recovery.
- **Fail-Open Mode**: Prevent total service outage by continuing to distribute traffic even when all health checks are failing.
- **Weighted Traffic Distribution**: Control traffic proportions across backends using weights for gradual rollouts or capacity-proportional distribution.
- **Drain Mode**: Gracefully remove backends from rotation — existing connections complete while new traffic routes elsewhere.
- **Backup Backends**: Active-standby topologies where backup backends activate only during primary backend failures.
- **Freeform Tagging**: Standard OpenMCF labels applied as OCI freeform tags for resource management and cost tracking.
- **Infrastructure as Code**: Full Pulumi (Go) and Terraform (HCL) implementations producing identical resource topology and outputs.
