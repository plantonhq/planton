# Overview

The **DigitalOcean VPC API Resource** provides a consistent and standardized interface for deploying and managing Virtual Private Cloud (VPC) networks on DigitalOcean. This resource simplifies the creation of isolated, private networks for your infrastructure, enabling secure communication between Droplets, Kubernetes clusters, and managed databases.

## Purpose

We developed this API resource to streamline the deployment and management of DigitalOcean VPCs. By offering a unified interface, it reduces the complexity involved in setting up private networks, enabling users to:

- **Easily Deploy VPCs**: Quickly provision VPCs with minimal configuration or auto-generated CIDR blocks.
- **VPC-First Infrastructure**: Create network isolation before deploying DOKS clusters, databases, and other resources.
- **Flexible IP Planning**: Choose between auto-generated /20 CIDR blocks (80% use case) or explicit IP address planning (20% use case).
- **Integrate Seamlessly**: VPCs automatically work with all DigitalOcean services (Droplets, DOKS, managed databases, load balancers).
- **Cost Efficiency**: Free internal VPC traffic and no additional VPC charges.

## Key Features

- **Consistent Interface**: Aligns with our existing APIs for deploying cloud infrastructure across providers.
- **Simplified Deployment**: Automates VPC provisioning with intelligent defaults.
- **Auto-Generated CIDR Blocks**: Optionally let DigitalOcean generate non-conflicting IP ranges automatically.
- **Explicit CIDR Control**: Supports custom IP ranges for production environments with specific IPAM requirements.
- **Regional Isolation**: VPCs are region-specific for compliance and data locality.
- **Production-Ready**: Built-in best practices for network planning and immutability awareness.

## Use Cases

- **Development Environments**: Quickly create isolated networks with auto-generated IP ranges.
- **Production Workloads**: Deploy VPC-native DOKS clusters and managed databases with explicit CIDR planning.
- **Multi-Environment Architecture**: Separate dev, staging, and production networks with non-overlapping IP ranges.
- **Security and Compliance**: Private network isolation for sensitive workloads.
- **Cost Optimization**: Free internal traffic within VPCs and between peered VPCs in the same datacenter.

## Critical Constraints

Understanding DigitalOcean VPC constraints is essential for production deployments:

- **Immutability**: IP address ranges cannot be changed after creation. Plan for growth or face costly migrations.
- **VPC-First Imperative**: DOKS clusters and load balancers must be created inside VPCs from day one (cannot be migrated later).
- **Regional Scope**: Each VPC is confined to a single region. Multi-region architectures require VPC peering.
- **No Overlapping CIDRs**: VPCs with overlapping IP ranges cannot be peered (even across regions).
- **Reserved Ranges**: DigitalOcean reserves specific CIDR blocks (10.244.0.0/16, 10.245.0.0/16) that cannot be used.

## Production Features

This resource provides complete support for production-grade VPC deployments, including:

- **Auto-Generated CIDR Blocks**: Let DigitalOcean handle IP allocation (recommended for dev/test).
- **Explicit CIDR Control**: Specify exact IP ranges for production environments.
- **Regional Deployment**: Deploy VPCs in any DigitalOcean region.
- **Description Field**: Document VPC purpose for team clarity.
- **Default VPC Configuration**: Optional setting for regional default VPC.
- **Immutability Protection**: Design enforces planning before deployment.
