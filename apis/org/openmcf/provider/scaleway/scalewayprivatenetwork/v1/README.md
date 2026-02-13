# Overview

The **Scaleway Private Network API Resource** provides a consistent and standardized interface for deploying and managing Private Networks on Scaleway. A Scaleway Private Network is a regional, Layer 2 network that lives inside a VPC and provides secure, private connectivity between Scaleway resources.

## Why Private Network Matters

In the Scaleway infrastructure model, Private Network is the **universal connector**. While a VPC is the logical container, resources do not attach directly to a VPC -- they attach to a Private Network within that VPC. This makes Private Network the most-referenced resource in the Scaleway resource graph:

- **Kapsule clusters** attach to a Private Network for pod-to-pod and pod-to-service communication.
- **RDB instances** (PostgreSQL/MySQL) attach to a Private Network for secure database access.
- **Redis clusters** and **MongoDB instances** attach to a Private Network for cache/document store access.
- **Load Balancers** attach to a Private Network for backend health checks and traffic forwarding.
- **Instances** attach to a Private Network via private NICs.
- **Public Gateways** attach to a Private Network to provide NAT and SSH bastion access.
- **Serverless Containers** attach to a Private Network for VPC-private function invocation.

Getting Private Network right is critical because every downstream resource depends on it.

## Purpose

This API resource streamlines the deployment and management of Scaleway Private Networks. By offering a unified interface, it reduces the complexity involved in setting up private connectivity, enabling users to:

- **Create Private Networks** inside an existing VPC with a single resource declaration.
- **Control IP Address Space** by specifying an IPv4 CIDR, or let Scaleway's IPAM auto-allocate one.
- **Enable Route Propagation** so resources in this network can discover and reach resources in other Private Networks within the same VPC.
- **Wire Dependencies** using `StringValueOrRef` for the VPC reference, enabling infra-chart composition.

## Key Features

- **Consistent Interface**: Aligns with the OpenMCF pattern for deploying cloud infrastructure across providers.
- **Cross-Resource References**: The `vpc_id` field uses `StringValueOrRef`, enabling both literal values and `valueFrom` references in infra charts.
- **IPAM Integration**: Optional IPv4 subnet specification -- omit it for automatic allocation, or specify a CIDR for explicit address planning.
- **IPv6 Support**: Optional dual-stack networking with one or more IPv6 subnets.
- **Route Propagation**: Control whether default VPC routes are visible to resources in this network.
- **Automatic Tagging**: Standard OpenMCF labels are applied as Scaleway tags for consistent resource management.
- **Infra-Chart Ready**: Exports `private_network_id` as the primary output for downstream resource wiring.

## How Scaleway Private Networks Work

Scaleway's networking model has a clear hierarchy:

1. **VPC** (Layer 0) -- Regional logical container. No IP addressing. Provides isolation boundary and optional inter-network routing.
2. **Private Network** (Layer 1) -- Regional Layer 2 network inside a VPC. This is where IP addressing happens. Resources attach here.
3. **Resources** (Layer 2+) -- Kapsule clusters, databases, instances, etc. They attach to one or more Private Networks.

Key characteristics:
- **Built-in DHCP**: Every Private Network includes automatic IP address management. Resources attached to the network receive IPs automatically.
- **Regional scope**: A Private Network exists within a single region and must match its parent VPC's region.
- **Multiple per VPC**: A VPC can contain multiple Private Networks, each with its own CIDR range.
- **Cross-network communication**: When the parent VPC has routing enabled and the Private Network has route propagation enabled, resources in different Private Networks can communicate.

## Critical Constraints

- **Region must match VPC**: The Private Network's region must be the same as its parent VPC. Creating a Private Network in a different region from its VPC will fail.
- **CIDR planning matters**: If multiple Private Networks share a VPC with routing enabled, their IPv4 CIDRs must not overlap. Plan address ranges carefully for multi-tier architectures.
- **Immutable region**: The region cannot be changed after creation. To relocate, create a new Private Network and migrate resources.

## Use Cases

- **Single-Tier Development**: One VPC, one Private Network, one Kapsule cluster. The simplest setup for development.
- **Multi-Tier Production**: One VPC with routing, separate Private Networks for application (Kapsule), database (RDB), and cache (Redis) tiers. Each tier has its own address space.
- **Managed Database Access**: A Private Network dedicated to RDB and Redis instances, referenced by application Kapsule clusters for secure database connectivity.
- **Public-Facing with NAT**: A Private Network with a Public Gateway attached for outbound NAT and SSH bastion access.

## Production Features

This resource provides complete support for production-grade Private Network deployments, including:

- **Address Space Control**: Specify IPv4 CIDR or rely on automatic IPAM allocation.
- **Dual-Stack Networking**: Optional IPv6 subnet support for workloads requiring IPv6 connectivity.
- **Route Propagation**: Enable default route propagation for inter-network communication within a VPC.
- **Automatic Labeling**: Standard OpenMCF labels applied as Scaleway tags for resource management and cost allocation.
- **Infrastructure as Code**: Full Pulumi (Go) and Terraform (HCL) implementations.
- **Infra-Chart Composability**: Designed as a Layer 1 connector that downstream resources reference via `StringValueOrRef`, and that itself references the Layer 0 VPC.
