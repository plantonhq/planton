# Overview

The **Scaleway Public Gateway API Resource** provides a consistent and standardized interface for deploying and managing Public Gateways on Scaleway. A Scaleway Public Gateway is a managed network appliance that sits at the edge of a Private Network and provides NAT, SSH bastion, and port forwarding capabilities.

## Why Public Gateway Matters

In Scaleway's networking model, resources attached to a Private Network have no direct internet access by default. The Public Gateway is the standard way to grant outbound internet connectivity:

- **Kapsule clusters** need the gateway for pod internet access (pulling container images, reaching external APIs).
- **Instances** in a Private Network use the gateway's NAT to reach the internet without individual public IPs.
- **RDB instances** and **Redis clusters** may need outbound access for replication or external integrations.
- **Serverless Containers** in a VPC-attached Private Network route outbound traffic through the gateway.

The Public Gateway also serves as an **SSH bastion** -- a secure jump host for administrative access to resources that have no public IP.

## What This Resource Bundles

This is a **composite resource** that bundles three Scaleway resources into a single declarative unit:

| Scaleway Resource | Purpose | Always Created? |
|---|---|---|
| `scaleway_vpc_public_gateway_ip` | Dedicated public IPv4 (Flexible IP) | Yes |
| `scaleway_vpc_public_gateway` | The gateway appliance | Yes |
| `scaleway_vpc_gateway_network` | Attaches gateway to a Private Network | Yes |
| `scaleway_vpc_public_gateway_pat_rule` | Port forwarding rules | Only if `pat_rules` specified |

These resources are bundled because a gateway without an IP and network attachment is useless. Bundling them eliminates boilerplate and ensures correct dependency ordering.

**Not bundled** (deprecated): DHCP-related resources (`scaleway_vpc_public_gateway_dhcp`, `scaleway_vpc_public_gateway_dhcp_reservation`) were deprecated by Scaleway in favor of Private Network IPAM mode.

## Purpose

This API resource streamlines the deployment and management of Scaleway Public Gateways. By offering a unified interface, it reduces the complexity involved in setting up network edge services, enabling users to:

- **Enable NAT** for an entire Private Network with a single resource declaration.
- **Configure SSH bastion** access with IP allowlisting for secure remote administration.
- **Set up port forwarding** to expose specific services without public IPs on individual resources.
- **Wire dependencies** using `StringValueOrRef` for the Private Network reference, enabling infra-chart composition.

## Key Features

- **Consistent Interface**: Aligns with the Planton pattern for deploying cloud infrastructure across providers.
- **Cross-Resource References**: The `private_network_id` field uses `StringValueOrRef`, enabling both literal values and `valueFrom` references in infra charts.
- **Composite Resource**: Bundles IP + Gateway + GatewayNetwork into one declaration, eliminating manual wiring.
- **NAT Masquerade**: One toggle to enable outbound internet for all resources in the Private Network.
- **SSH Bastion**: Secure jump host with configurable port and IP allowlisting.
- **Port Forwarding**: PAT rules for inbound access to specific private services.
- **Reverse DNS**: Optional reverse DNS on the public IP for compliance and email deliverability.
- **Automatic Tagging**: Standard Planton labels are applied as Scaleway tags for consistent resource management.
- **Infra-Chart Ready**: Exports `gateway_id`, `public_ip_address`, `public_ip_id`, and `gateway_network_id` for downstream wiring.

## How Scaleway Public Gateways Work

Scaleway's networking model has a clear hierarchy, and the Public Gateway sits at the network edge:

1. **VPC** (Layer 0) -- Regional logical container. Provides isolation and optional inter-network routing.
2. **Private Network** (Layer 1) -- Regional Layer 2 network inside a VPC. Resources attach here.
3. **Public Gateway** (Layer 2) -- Zonal edge appliance attached to a Private Network. Provides the link between the private and public internet.

Key characteristics:
- **Zonal scope**: Gateways are deployed to a specific zone (e.g., `fr-par-1`) within a region. The zone must be within the same region as the attached Private Network.
- **Dedicated public IP**: Each gateway gets a Flexible IP. This IP is the single point of contact for all NAT, bastion, and port forwarding traffic.
- **Up to 8 attachments**: A single gateway can attach to up to 8 Private Networks (this resource manages one attachment; additional attachments can be managed separately).
- **Managed appliance**: Scaleway handles failover, updates, and maintenance.

## Critical Constraints

Understanding these constraints is essential for production deployments:

- **Zonal, not regional**: The gateway zone must match the Private Network's region. If the network is in `fr-par`, the gateway must be in `fr-par-1`, `fr-par-2`, etc.
- **VPC-GW-XL availability**: The high-bandwidth type is only available in Paris region zones (`fr-par-*`).
- **SMTP blocked by default**: Outbound port 25 is blocked unless `enable_smtp` is set to true.
- **Bastion requires SSH keys**: The SSH bastion proxies connections using SSH keys configured on the target instances, not separate bastion credentials.
- **PAT rules require known IPs**: Port forwarding rules need the target private IP, which is typically only known after the target resource is created.
- **One attachment per resource**: This Planton resource manages one Private Network attachment. For multi-network gateways, additional `GatewayNetwork` resources can be managed separately.

## Use Cases

- **Kapsule Environment with NAT**: Create a VPC, Private Network, and Public Gateway with masquerade. The gateway provides NAT for Kapsule pods to pull images and reach external APIs.
- **Secure Development Access**: Enable the SSH bastion with IP allowlisting for developers to SSH into instances in the Private Network without exposing them publicly.
- **Service Exposure via Port Forwarding**: Use PAT rules to expose a web server or database on specific ports, ideal for small deployments that don't need a full Load Balancer.
- **Email-Capable Infrastructure**: Enable SMTP and configure reverse DNS for resources that need to send email directly.

## Production Features

This resource provides complete support for production-grade gateway deployments, including:

- **NAT Masquerade**: Single toggle for outbound internet access across the Private Network.
- **SSH Bastion**: Auditable access point with IP allowlisting and configurable port.
- **Port Forwarding**: PAT rules for controlled inbound access to private services.
- **Reverse DNS**: Compliance and email deliverability support.
- **SMTP Control**: Explicit opt-in for outbound email traffic.
- **Automatic Labeling**: Standard Planton labels applied as Scaleway tags for resource management and cost allocation.
- **Infrastructure as Code**: Full Pulumi (Go) and Terraform (HCL) implementations.
- **Infra-Chart Composability**: Designed as a Layer 2 edge resource that references the Layer 1 Private Network via `StringValueOrRef` and exports outputs for diagnostics and DNS configuration.
