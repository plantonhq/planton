# Overview

The **OCI Subnet API Resource** provides a consistent and standardized interface for deploying and managing subnets within Oracle Cloud Infrastructure Virtual Cloud Networks. A subnet is a contiguous range of IP addresses within a VCN that serves as the deployment target for compute instances, load balancers, databases, and container workloads. Subnets control network access through public/private designation, route table association, and security list binding.

## Purpose

This API resource streamlines the deployment and management of OCI subnets with optional inline route tables. By offering a unified interface, it reduces the complexity of network segmentation and routing configuration, enabling users to:

- **Segment Network Address Space**: Carve a VCN's CIDR blocks into purpose-specific subnets — public subnets for load balancers, private subnets for application tiers, isolated subnets for databases.
- **Control Public/Private Access**: Toggle `prohibitPublicIpOnVnic` and `prohibitInternetIngress` to enforce private-only networking at the subnet level, independent of security rules.
- **Own Route Table Lifecycle**: Define route rules inline and let the subnet manage its own dedicated route table, or reference an existing route table, or inherit the VCN's default. Three modes, one field.
- **Enable Infra-Chart Composability**: Subnets export the subnet ID, domain name, virtual router IP/MAC, and associated route table ID — all available as `StringValueOrRef` targets for downstream resources like OciComputeInstance and OciContainerEngineCluster.

## Key Features

- **Consistent Interface**: Aligns with the Planton pattern for deploying cloud infrastructure across providers.
- **Regional by Default**: Subnets span all availability domains unless `availabilityDomain` is explicitly set. Regional subnets are recommended for most workloads as they simplify high-availability architecture.
- **Inline Route Table**: Declare `routeRules` directly on the subnet spec and a dedicated route table is created, named `{displayName}-rt`, and associated automatically. No separate resource definition required.
- **Three Route Table Modes**: Inline rules (creates new), external reference (uses existing), or neither (inherits VCN default). Mutual exclusivity between `routeTableId` and `routeRules` is enforced by schema validation.
- **Dual-Stack IPv6**: Optional IPv6 CIDR block for subnets within IPv6-enabled VCNs.
- **DNS Resolution**: Optional DNS label creates a subnet domain (`<dnsLabel>.<vcnDnsLabel>.oraclevcn.com`) for hostname-based communication between resources.
- **Automatic Tagging**: Standard Planton freeform tags are applied to the subnet and any created route table (resource kind, resource ID, organization, environment, and user-defined labels from metadata).
- **Security List Binding**: Associate up to 5 security lists with a subnet. OCI recommends Network Security Groups for new deployments, but security lists remain supported for legacy configurations.

## How OCI Subnets Differ from Other Providers

Understanding these differences is essential when coming from AWS or other cloud platforms:

- **Regional vs AZ-Scoped**: OCI subnets are regional by default — a single subnet spans all availability domains in a region. AWS subnets are always scoped to a single Availability Zone. This means an OCI deployment needs fewer subnets to achieve high availability across ADs.
- **Route Table Association**: In OCI, a subnet is associated with exactly one route table at creation time. In AWS, subnets have an implicit association with the VPC's main route table that can be overridden with an explicit association. Planton supports all three OCI patterns: inline rules, explicit reference, or VCN default.
- **Security Lists vs NACLs**: OCI security lists are the subnet-level firewall, analogous to AWS Network ACLs. However, OCI also offers Network Security Groups (NSGs), which are per-VNIC and stateful — more comparable to AWS Security Groups. OCI recommends NSGs for new deployments.
- **CIDR Within VCN**: A subnet's CIDR must fall within one of the parent VCN's CIDR blocks. Unlike AWS where subnet CIDRs are carved from a single VPC CIDR (or secondary CIDRs), OCI subnets can be drawn from any of the VCN's multiple CIDR blocks.
- **Public/Private Is Explicit**: In OCI, the private nature of a subnet is controlled by `prohibitPublicIpOnVnic` and `prohibitInternetIngress` — explicit boolean fields. In AWS, a subnet's "public" status is implicit, determined by whether its route table has a route to an Internet Gateway.

## Critical Constraints

- **CIDR Block**: Must fall within one of the parent VCN's CIDR blocks. Must not overlap with any other subnet in the same VCN.
- **DNS Label Immutability**: Once set, a DNS label cannot be changed. Must be alphanumeric, start with a letter, and be at most 15 characters.
- **Security List Limit**: OCI enforces a maximum of 5 security lists per subnet.
- **Route Table Mutual Exclusivity**: `routeTableId` and `routeRules` cannot both be set. The schema enforces this with a CEL validation rule.
- **IPv6 Requires VCN Support**: The `ipv6CidrBlock` field is only valid when the parent VCN has IPv6 enabled via `isIpv6Enabled`.

## Use Cases

- **OKE Node Pool Subnets**: Private subnets with inline route rules pointing to the NAT Gateway (outbound image pulls) and Service Gateway (private OCI service access). OKE node pools reference these subnets for worker node placement.
- **Load Balancer Subnets**: Public subnets with DNS labels for load balancers that need direct internet-facing access. Typically placed in the same VCN as the OKE cluster subnets.
- **Database Tier Isolation**: Private subnets with no internet access at all — `prohibitPublicIpOnVnic: true`, `prohibitInternetIngress: true`, and route rules limited to the Service Gateway for OCI service access.
- **Development Flat Network**: A single regional subnet with the VCN's default route table and default security list. Minimal configuration for rapid prototyping.

## Production Features

This resource provides complete support for production-grade subnet deployments, including:

- **Regional High Availability**: Spans all availability domains by default, enabling workload placement across ADs without creating per-AD subnets.
- **Inline Routing**: Dedicated route tables per subnet with declarative route rules for NAT, Internet, Service, and DRG gateways.
- **Network Isolation**: Independent control over public IP assignment and internet ingress at the subnet level.
- **Freeform Tagging**: Standard Planton labels applied as OCI freeform tags for resource management, cost tracking, and compliance.
- **Infrastructure as Code**: Full Pulumi (Go) and Terraform (HCL) implementations producing identical outputs.
- **Infra-Chart Composability**: Designed as a Layer 1 building block that downstream resources reference via `StringValueOrRef` for the subnet ID and virtual router metadata.
