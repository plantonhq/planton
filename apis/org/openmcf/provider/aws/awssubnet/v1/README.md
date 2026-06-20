# Overview

The **AWS Subnet API Resource** provides a consistent, standardized interface for deploying and managing individual subnets within an AWS Virtual Private Cloud (VPC). A subnet is a contiguous range of IP addresses scoped to a single Availability Zone, and it is the placement target for nearly every AWS workload — EC2 instances, load balancers, RDS databases, EKS/ECS tasks, Lambda ENIs, and more.

## Purpose

This resource makes the subnet a first-class, independently composable building block rather than something bundled inside the VPC. By offering a unified interface, it lets you:

- **Lay out address space deliberately**: Carve a VPC's CIDR into purpose-specific subnets — public subnets for load balancers and bastions, private subnets for application tiers, isolated subnets for data stores.
- **Own routing per subnet**: Define route rules inline and let the subnet create and own its route table, reference an existing route table, or fall back to the VPC main route table. Three modes, two mutually-exclusive fields.
- **Reach the 90/10 surface**: Dual-stack IPv6, DNS64, resource-name DNS records, and launch-time public-IP assignment are all available, not just the surface knobs.
- **Compose topologies**: Every subnet exports its id, ARN, AZ, CIDR, and associated route table id as `StringValueOrRef` targets that downstream resources reference directly.

## Key Features

- **Single-AZ placement**: Each subnet lives in exactly one Availability Zone (`availabilityZone`), the AWS model. Span AZs by creating one `AwsSubnet` per zone.
- **Routing folded in**: Declare `routes` and a dedicated route table is created, populated, associated with the subnet, and its id exported — no separate route-table resource. Or set `routeTableId` to adopt an existing table. The two are mutually exclusive, enforced by schema validation.
- **Typed route targets**: Each route names a `targetType` (internet gateway, NAT gateway, transit gateway, VPC peering, VPC endpoint, network interface, egress-only internet gateway) and a `targetId`, mapping cleanly onto the AWS route attributes.
- **Dual-stack IPv6**: Optional `ipv6CidrBlock`, `assignIpv6AddressOnCreation`, and `enableDns64` for IPv6 and NAT64 workloads.
- **Launch-time behavior**: `mapPublicIpOnLaunch`, resource-name DNS A/AAAA records, and `privateDnsHostnameTypeOnLaunch` control how instances come up.
- **Consistent tagging**: Standard resource-identity tags are applied to the subnet and any created route table.
- **Dual-engine parity**: Identical behavior and outputs from the Pulumi (Go) and Terraform (HCL) modules.

## How AWS Subnets Differ from Other Providers

- **AZ-scoped, not regional**: An AWS subnet is always pinned to a single Availability Zone. (OCI subnets are regional by default; GCP subnetworks are regional.) High availability therefore means one subnet per AZ.
- **Public vs private is a routing property**: A subnet is "public" only because its route table has a default route to an internet gateway. There is no public/private flag — `mapPublicIpOnLaunch` only controls auto-assignment of public IPs, not reachability. This is why routing is folded into the subnet.
- **Explicit route-table association**: An AWS subnet implicitly uses the VPC main route table until an explicit association overrides it. This resource creates that association whenever `routes` or `routeTableId` is set.
- **CIDR within the VPC**: A subnet's `cidrBlock` must fall within the VPC's IPv4 CIDR (or a secondary CIDR) and not overlap a sibling. AWS reserves the first four and the last address in every subnet.

## Critical Constraints

- **CIDR block**: Required, must be within the VPC CIDR, must not overlap another subnet. Immutable — changing it replaces the subnet.
- **Availability zone**: Required and immutable. Changing it replaces the subnet.
- **VPC id**: Required and immutable. Changing it replaces the subnet.
- **Route table mutual exclusivity**: `routeTableId` and `routes` cannot both be set; the schema enforces this with a CEL rule.
- **One destination per route**: Each inline route sets exactly one of `destinationCidrBlock`, `destinationIpv6CidrBlock`, or `destinationPrefixListId`.

## Use Cases

- **Public load-balancer / bastion subnets**: `mapPublicIpOnLaunch: true` with an inline default route to an internet gateway.
- **Private application subnets**: No public IPs, with an inline default route to a NAT gateway for outbound-only internet access.
- **Isolated data-tier subnets**: No routes at all — the subnet stays on the VPC main route table with no internet path.
- **Dual-stack subnets**: An IPv6 CIDR plus an egress-only internet gateway route for outbound-only IPv6.

## Production Features

- **Per-subnet routing** with declarative rules for internet, NAT, transit, peering, endpoint, and egress-only targets.
- **Dual-stack networking** via IPv6 CIDR, DNS64, and IPv6 launch behavior.
- **Consistent resource-identity tagging** applied to the subnet and any created route table.
- **Infrastructure as Code** with full Pulumi (Go) and Terraform (HCL) implementations producing identical outputs.
- **Composability** as a foundational building block referenced by downstream resources via `StringValueOrRef`.
