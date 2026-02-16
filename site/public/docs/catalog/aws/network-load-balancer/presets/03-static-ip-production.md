---
title: "Static IP Production NLB"
description: "This preset creates a production-grade internet-facing Network Load Balancer with Elastic IPs for static public IPs, TLS termination, Route53 DNS, HTTP health checks, cross-zone load balancing,..."
type: "preset"
rank: "03"
presetSlug: "03-static-ip-production"
componentSlug: "network-load-balancer"
componentTitle: "Network Load Balancer"
provider: "aws"
icon: "package"
order: 3
---

# Static IP Production NLB

This preset creates a production-grade internet-facing Network Load Balancer with Elastic IPs for static public IPs, TLS termination, Route53 DNS, HTTP health checks, cross-zone load balancing, deletion protection, connection termination, and client IP preservation. This is the comprehensive production preset.

## When to Use

- Production workloads requiring static public IPs for allowlisting, firewall rules, or DNS pinning
- Compliance or security requirements where IP addresses must not change across NLB scaling events
- Public APIs or services that need TLS termination, automatic DNS, and application-level health checks
- High-availability scenarios where cross-zone load balancing and connection termination are important

## Key Configuration Choices

- **Static IPs via Elastic IPs** — Each subnet mapping includes an `allocationId` to pin the NLB node to a specific Elastic IP; IPs do not change on redeploy or scaling
- **TLS termination** — Listener on port 443 with ACM certificate; NLB decrypts and forwards plaintext TCP to targets
- **Deletion protection** (`deleteProtectionEnabled: true`) — Prevents accidental deletion
- **Cross-zone load balancing** (`crossZoneLoadBalancingEnabled: true`) — Distributes traffic across all targets in all AZs; recommended when targets are unevenly distributed
- **HTTP health check** — Validates application readiness via HTTP path instead of TCP-only; `matcher` defines acceptable response codes
- **Connection termination** (`connectionTermination: true`) — NLB closes connections to deregistered targets when the deregistration delay expires; important for long-lived connections (WebSocket, gRPC streams)
- **Preserve client IP** (`preserveClientIp: true`) — Targets see the original client IP; useful for logging, rate limiting, and geo-aware logic
- **Route53 DNS** — Creates alias A records for the specified hostnames; no DNS management outside OpenMCF

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<nlb-name>` | Unique name for the NLB (lowercase, alphanumeric, hyphens) | Choose a descriptive name (e.g., `api-production`, `grpc-gateway`) |
| `<public-subnet-id-az1>` | Public subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<public-subnet-id-az2>` | Public subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<elastic-ip-allocation-id-az1>` | Allocation ID of an Elastic IP in the first AZ | Allocate an EIP in AWS EC2 console, then use its allocation ID |
| `<elastic-ip-allocation-id-az2>` | Allocation ID of an Elastic IP in the second AZ | Allocate an EIP in AWS EC2 console, then use its allocation ID |
| `<acm-certificate-arn>` | ARN of an ACM certificate covering your domain | AWS ACM console or `AwsCertManagerCert` status outputs |
| `<application-port>` | Port on targets to forward decrypted traffic to | Your application's listening port |
| `<health-check-path>` | HTTP path for health checks (e.g., `/health`, `/ready`) | Your application's health endpoint |
| `<route53-hosted-zone-id>` | ID of the Route53 hosted zone for your domain | AWS Route53 console or `AwsRoute53Zone` status outputs |
| `<your-domain.com>` | Domain name that should point to this NLB | Your domain (e.g., `api.example.com`) |

## Important Notes

- Elastic IPs must be allocated in the same region as the NLB and associated with the correct subnets (one per AZ)
- The `healthCheck.path` must be served by your application on the target port; otherwise targets will be marked unhealthy
- For NLB, `unhealthyThreshold` must equal `healthyThreshold`; both are set to 3 in this preset

## Common Additions

- Add `securityGroups` to restrict inbound traffic
- Add `dnsRecordClientRoutingPolicy: availability_zone_affinity` to reduce cross-zone traffic costs
- Use `valueFrom` references to `AwsCertManagerCert`, `AwsRoute53Zone`, or other OpenMCF resources for portability

## Related Presets

- **01-tcp-internal** — Use for internal-only microservice communication
- **02-tls-internet-facing** — Use when static IPs and DNS are not required
