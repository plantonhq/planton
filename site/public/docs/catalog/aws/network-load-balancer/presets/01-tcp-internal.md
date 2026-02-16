---
title: "TCP Internal NLB"
description: "This preset creates an internal Network Load Balancer for microservice-to-microservice communication within a VPC. It uses plain TCP (no TLS termination), spans two subnets for high availability, and..."
type: "preset"
rank: "01"
presetSlug: "01-tcp-internal"
componentSlug: "network-load-balancer"
componentTitle: "Network Load Balancer"
provider: "aws"
icon: "package"
order: 1
---

# TCP Internal NLB

This preset creates an internal Network Load Balancer for microservice-to-microservice communication within a VPC. It uses plain TCP (no TLS termination), spans two subnets for high availability, and is not accessible from the internet.

## When to Use

- Load balancing traffic between internal microservices (e.g., gRPC, Redis, databases)
- Service mesh or API gateway backends that only need VPC-internal access
- Development or staging environments where TLS termination is handled by the application
- Any scenario where you need Layer 4 load balancing without public exposure

## Key Configuration Choices

- **Internal** (`internal: true`) — NLB is created in private subnets and receives a private DNS name; not reachable from the internet
- **TCP protocol** — Pass-through TCP; no TLS termination at the NLB; targets handle encryption if needed
- **Two subnets** — Minimum for cross-AZ high availability; use private subnets in different Availability Zones
- **Simple target group** — Single TCP port; targets (instances, IPs) are registered by ECS, EKS, or auto-scaling groups

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<nlb-name>` | Unique name for the NLB (lowercase, alphanumeric, hyphens) | Choose a descriptive name (e.g., `api-internal`, `grpc-backend`) |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<listener-port>` | Port the NLB accepts traffic on (e.g., 80, 8080, 9090) | Your application's exposed port |
| `<target-port>` | Port on targets to forward traffic to | Usually matches listener port for TCP pass-through |

## Common Additions

- Add `securityGroups` with an `AwsSecurityGroup` reference to restrict inbound traffic
- Add `crossZoneLoadBalancingEnabled: true` if targets are unevenly distributed across AZs
- Add `preserveClientIp: true` on the target group if backends need the original client IP
- Use `valueFrom` references to `AwsVpc` subnets instead of literal subnet IDs for portability

## Related Presets

- **02-tls-internet-facing** — Use when the NLB must be internet-facing with TLS termination
- **03-static-ip-production** — Use when you need static public IPs, DNS, and production hardening
