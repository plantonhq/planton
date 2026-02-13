# Overview

The **Scaleway VPC API Resource** provides a consistent and standardized interface for deploying and managing Virtual Private Cloud (VPC) networks on Scaleway. A Scaleway VPC is a regional, logical container that groups Private Networks, enabling network isolation and optional inter-network routing.

## Purpose

This API resource streamlines the deployment and management of Scaleway VPCs. By offering a unified interface, it reduces the complexity involved in setting up isolated network environments, enabling users to:

- **Create Network Containers**: Provision VPCs as the foundation layer for all Scaleway networking.
- **Enable Inter-Network Routing**: Allow Private Networks within the same VPC to communicate with each other, essential for multi-tier architectures.
- **Isolate Environments**: Separate development, staging, and production infrastructure with distinct VPCs.
- **Integrate with Infra Charts**: VPCs serve as the Layer 0 foundation in Scaleway infra charts (kapsule-environment, serverless-environment, database-stack).

## Key Features

- **Consistent Interface**: Aligns with the OpenMCF pattern for deploying cloud infrastructure across providers.
- **Regional Scope**: VPCs are confined to a single Scaleway region (fr-par, nl-ams, pl-waw).
- **Private Network Routing**: Optional routing toggle enables communication between Private Networks in the same VPC.
- **Custom Routes Propagation**: Advanced networking support for VPN gateways and network appliances.
- **Automatic Tagging**: Standard OpenMCF labels are applied as Scaleway tags for consistent resource management.
- **Infra-Chart Ready**: Exports `vpc_id` for downstream `StringValueOrRef` references from Private Networks, Kapsule clusters, and other resources.

## How Scaleway VPCs Differ from Other Providers

Unlike DigitalOcean or AWS VPCs, Scaleway VPCs do **not** define IP ranges or CIDR blocks at the VPC level. A Scaleway VPC is purely a logical grouping container. IP address planning happens at the **Private Network** level (ScalewayPrivateNetwork), where CIDR blocks are assigned to individual networks.

This means:
- **No CIDR planning required** when creating a VPC.
- **IP conflicts are managed** at the Private Network level, not the VPC level.
- **VPCs are lightweight** -- they're essentially named containers with an optional routing toggle.

## Critical Constraints

Understanding these constraints is essential for production deployments:

- **One-Way Routing Toggle**: `enable_routing` allows Private Networks in the VPC to communicate. Once enabled, it **cannot be disabled**. This is enforced by the Scaleway API. Plan your routing requirements before creation.
- **One-Way Custom Routes Propagation**: `enable_custom_routes_propagation` advertises custom routes between Private Networks. Once enabled, it **cannot be deactivated**. Only enable this if you have specific advanced networking requirements.
- **Regional Scope**: Each VPC is confined to a single region. Multi-region architectures require separate VPCs per region.
- **Immutable Region**: The region cannot be changed after creation. To move to a different region, you must create a new VPC and migrate resources.

## Use Cases

- **Foundation for Kapsule Environments**: Create a VPC with routing enabled, then attach a Kapsule cluster and RDB instance via separate Private Networks that can communicate through the VPC.
- **Development Environments**: Minimal VPC with no routing -- just a logical container for a single Private Network.
- **Multi-Tier Production**: VPC with routing enabled for application, database, and cache tiers in separate Private Networks.
- **Multi-Region Architecture**: Separate VPCs per region, each containing their own set of Private Networks and resources.

## Production Features

This resource provides complete support for production-grade VPC deployments, including:

- **Routing Control**: Enable or disable inter-Private-Network routing at VPC creation time.
- **Custom Routes**: Support for advanced networking scenarios with custom route propagation.
- **Automatic Labeling**: Standard OpenMCF labels applied as Scaleway tags for resource management, cost allocation, and compliance.
- **Infrastructure as Code**: Full Pulumi (Go) and Terraform (HCL) implementations.
- **Infra-Chart Composability**: Designed as a Layer 0 foundation that downstream resources reference via `StringValueOrRef`.
