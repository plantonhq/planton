# Overview

The **OCI VCN API Resource** provides a consistent and standardized interface for deploying and managing Virtual Cloud Networks on Oracle Cloud Infrastructure. A VCN is the foundational networking construct in OCI — an isolated virtual network within a compartment that supports multiple CIDR blocks, DNS resolution, IPv6, and optional gateway sub-resources.

## Purpose

This API resource streamlines the deployment and management of OCI VCNs and their tightly coupled gateways. By offering a unified interface, it reduces the complexity involved in setting up network foundations, enabling users to:

- **Create Network Foundations**: Provision VCNs as the Layer 0 building block for all OCI networking. Every subnet, load balancer, compute instance, and database in OCI lives inside a VCN.
- **Bundle Gateway Lifecycle**: Manage Internet, NAT, and Service Gateways alongside the VCN they serve, controlled by simple boolean toggles rather than separate resource definitions.
- **Enable Infra-Chart Composability**: VCNs export IDs for the VCN itself, default route table, default security list, default DHCP options, and each gateway — all available as `StringValueOrRef` targets for downstream resources.
- **Isolate Environments**: Separate development, staging, and production infrastructure with distinct VCNs and compartments.

## Key Features

- **Consistent Interface**: Aligns with the Planton pattern for deploying cloud infrastructure across providers.
- **Multiple CIDR Blocks**: A single VCN can span multiple non-overlapping IPv4 CIDR blocks (e.g., 10.0.0.0/16 and 172.16.0.0/16), unlike AWS VPCs which historically supported only one primary CIDR (secondary CIDRs added later as a separate operation).
- **IPv6 Support**: Optional Oracle-assigned /56 IPv6 GUA prefix for dual-stack workloads.
- **Toggle-Based Gateways**: Internet, NAT, and Service Gateways are declared as boolean fields on the VCN spec. No separate resource definitions, no cross-resource wiring — just flip the toggle.
- **Service Gateway Auto-Configuration**: The Service Gateway is automatically configured for all services in the Oracle Services Network. Users do not need to know service OCIDs, which vary by region.
- **DNS Resolution**: Optional DNS label creates a VCN domain (`<dnsLabel>.oraclevcn.com`) and enables hostname-based communication between resources.
- **Automatic Tagging**: Standard Planton freeform tags are applied to the VCN and all gateways (resource kind, resource ID, organization, environment, and user-defined labels from metadata).
- **Infra-Chart Ready**: Exports 7 stack outputs for downstream `StringValueOrRef` references from subnets, security groups, and other networking components.

## How OCI VCNs Differ from Other Providers

Understanding these differences is essential when coming from AWS or other cloud platforms:

- **Compartment Model**: OCI uses compartments for hierarchical resource isolation, unlike AWS which uses accounts. The `compartmentId` field is the first and most fundamental input for every OCI resource. Compartments can be nested, and IAM policies are scoped to compartments.
- **Multiple CIDRs at Creation**: OCI VCNs natively support multiple CIDR blocks as a first-class creation parameter. AWS VPCs require adding secondary CIDRs as a separate operation after creation.
- **Built-In Defaults**: OCI automatically creates a default route table, default security list, and default DHCP options with every VCN. These defaults are exported as stack outputs and can be used directly or replaced with custom equivalents.
- **Service Gateway**: Unique to OCI. Provides private, backbone-only access to OCI services (Object Storage, Autonomous Database, etc.) without traffic traversing the internet. AWS achieves similar functionality with VPC Endpoints, but those require per-service configuration. The OCI Service Gateway covers all services in one resource.
- **Region via Provider Config**: Unlike some providers where region is a per-resource field, OCI resources inherit their region from the provider configuration. This ensures all resources deployed together share a consistent region.

## Critical Constraints

- **CIDR Range Restrictions**: Each CIDR block must be between /16 and /30. Blocks within the same VCN must not overlap.
- **DNS Label Immutability**: Once a DNS label is set on a VCN, it cannot be changed. The label must be alphanumeric, start with a letter, and be at most 15 characters.
- **Service Gateway Scope**: The Service Gateway is automatically configured for all services in the Oracle Services Network. Individual service selection is not currently supported.
- **Gateway Naming**: Gateways are named using the VCN display name with a suffix (`-igw`, `-ngw`, `-sgw`). Custom gateway names are not supported.

## Use Cases

- **Foundation for OKE Environments**: Create a VCN with all three gateways, then layer OciSubnet (public + private) and OciContainerEngineCluster on top. The Internet Gateway serves the API endpoint and load balancers; the NAT Gateway gives worker nodes outbound access; the Service Gateway provides private access to OCI Container Registry.
- **Private Database Networking**: A VCN with only a Service Gateway — no internet access at all. OciAutonomousDatabase and OciDbSystem instances access OCI services privately while remaining completely isolated from the internet.
- **Multi-Tier Web Applications**: VCN with Internet and NAT Gateways. Public subnets host load balancers via the Internet Gateway; private subnets host application and database tiers with outbound access via the NAT Gateway.
- **Development Environments**: Minimal VCN with a single CIDR and no gateways. Quick to create, low complexity, suitable for isolated experimentation.

## Production Features

This resource provides complete support for production-grade VCN deployments, including:

- **Dual-Stack Networking**: IPv6 alongside IPv4 for forward-looking network architectures.
- **Gateway Toggles**: Enable or disable gateways declaratively without managing separate resources or cross-resource dependencies.
- **Freeform Tagging**: Standard Planton labels applied as OCI freeform tags for resource management, cost tracking, and compliance.
- **Infrastructure as Code**: Full Pulumi (Go) and Terraform (HCL) implementations producing identical outputs.
- **Infra-Chart Composability**: Designed as a Layer 0 foundation that downstream resources reference via `StringValueOrRef` for the VCN ID, gateway IDs, and default resource IDs.
