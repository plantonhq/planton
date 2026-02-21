---
title: "Web Tier NSG"
description: "This preset creates a Network Security Group for internet-facing resources such as load balancers, web servers, and API gateways. Inbound traffic is restricted to HTTP (80) and HTTPS (443) from..."
type: "preset"
rank: "01"
presetSlug: "01-web-tier"
componentSlug: "network-security-group"
componentTitle: "Network Security Group"
provider: "oci"
icon: "package"
order: 1
---

# Web Tier NSG

This preset creates a Network Security Group for internet-facing resources such as load balancers, web servers, and API gateways. Inbound traffic is restricted to HTTP (80) and HTTPS (443) from anywhere, plus ICMP Path MTU Discovery. All outbound traffic is allowed. This covers the majority of public-facing OCI deployments and is the recommended starting point for any resource that serves web traffic.

## When to Use

- Load balancers that terminate TLS and serve HTTPS traffic from the public internet
- Web servers or reverse proxies that handle HTTP/HTTPS requests directly
- API gateways that expose REST or gRPC endpoints to external clients
- Any VNIC that needs to accept inbound web traffic while restricting all other ports

## Key Configuration Choices

- **HTTPS ingress** (`protocol: tcp`, port 443 from `0.0.0.0/0`) -- The primary rule. All modern web traffic should use TLS. This allows inbound HTTPS from any source.
- **HTTP ingress** (`protocol: tcp`, port 80 from `0.0.0.0/0`) -- Allows plain HTTP for redirect-to-HTTPS flows or legacy clients. Remove this rule if your architecture enforces HTTPS-only at a layer above (e.g., a CDN or WAF).
- **ICMP Path MTU Discovery** (`protocol: icmp`, type 3 code 4 from `0.0.0.0/0`) -- OCI best practice for any public-facing resource. Without this rule, TCP connections can silently stall when packets exceed the path MTU. This is especially important for resources behind load balancers.
- **All outbound** (`protocol: all` to `0.0.0.0/0`) -- Web-tier resources typically need to reach backend services, databases, external APIs, and OCI services. Restricting egress is uncommon for web-tier NSGs.
- **Stateful rules** (`stateless` not set, defaults to `false`) -- All rules are stateful, meaning return traffic is automatically allowed. This is the standard for web workloads and matches OCI's default behavior.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the NSG will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<vcn-ocid>` | OCID of the VCN this NSG belongs to | OCI Console > Networking > VCNs, or `OciVcn` status outputs (`vcnId`) |

## Related Presets

- **02-private-backend** -- Use instead for resources that should only accept traffic from within the VCN (databases, app servers, internal microservices)
- **03-development** -- Use instead for dev/test environments where all traffic should be permitted
