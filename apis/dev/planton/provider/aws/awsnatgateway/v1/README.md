# Overview

The **AWS NAT Gateway API Resource** provides a consistent, standardized interface for deploying and managing an AWS NAT (network address translation) gateway. A NAT gateway gives instances in a private subnet outbound network access while keeping them unreachable from inbound connections — the standard way to let private workloads reach the internet (or other private networks) without exposing them.

## Purpose

This resource makes the NAT gateway a first-class, independently composable building block rather than something bundled inside the VPC. By offering a unified interface, it lets you:

- **Compose egress deliberately**: Create the gateway as its own graph node in exactly the subnet you intend, by reference or by literal id.
- **Wire it as a route target**: The gateway's id is exported so an `AwsSubnet` can send a default route (`0.0.0.0/0`) to it — the step that actually gives a private subnet outbound access.
- **Bind a stable IP by reference**: A public gateway's Elastic IP is referenced (`allocationId` -> `AwsElasticIp`), never embedded, so the address has its own lifecycle and can be reasoned about as a first-class node.
- **Own its lifecycle**: The gateway is created and destroyed independently of the VPC and the subnet, so it can be reasoned about, referenced, and managed on its own.

## Key Features

- **Public and private connectivity**: `connectivityType: public` (Elastic IP, internet egress) or `private` (no Elastic IP, egress to peered/transit/VPN networks only).
- **EIP by reference**: A public gateway binds a stable outbound address via `allocationId`, referencing an `AwsElasticIp` or a literal `eipalloc-` id. Never embeds an IP allocation.
- **High-throughput options**: Additional Elastic IPs (`secondaryAllocationIds`) for public gateways or additional private IPs (`secondaryPrivateIpAddresses` / `secondaryPrivateIpAddressCount`) for private gateways, to widen the source-port range under heavy egress.
- **Consistent tagging**: Standard resource-identity tags are applied to the gateway.
- **Dual-engine parity**: Identical behavior and outputs from the Pulumi (Go) and Terraform (HCL) modules.

## How the NAT Gateway Fits the Network Model

- **Placement vs routing**: A NAT gateway lives in one subnet but serves others. A public gateway must live in a **public** subnet (one that routes to an internet gateway); the **private** subnets that need egress send their default route to the gateway. Creating the gateway does nothing on its own until a subnet routes to it.
- **NAT gateway vs internet gateway**: An internet gateway connects a VPC to the internet for resources with public IPs (inbound and outbound). A NAT gateway gives private-subnet instances **outbound-only** access and itself depends on an internet gateway being present for the public-subnet path.
- **NAT gateway vs egress-only internet gateway**: A NAT gateway is for IPv4. For IPv6-only outbound access, use an egress-only internet gateway instead.

## Critical Constraints

- **Connectivity type**: Required, `public` or `private`. A public gateway requires an `allocationId`; a private gateway forbids any Elastic IP and is the only kind that accepts private-IP addressing.
- **Subnet**: Required and immutable — changing it replaces the gateway.
- **Region**: Required, must match the subnet's region.

## Use Cases

- **Private-subnet internet egress**: The canonical pattern — private subnets route `0.0.0.0/0` to a public NAT gateway in a public subnet for outbound access to package repos, APIs, and AWS services.
- **High-availability egress**: One NAT gateway per availability zone, each in that zone's public subnet, so zonal failures do not sever egress.
- **Private inter-network egress**: A private NAT gateway for outbound communication to other VPCs or on-premises networks without any internet exposure.

## Production Features

- **Independent, referenceable gateway** exported as a route target for downstream subnets.
- **Composable Elastic IP** referenced rather than embedded, preserving the IP's own lifecycle.
- **Consistent resource-identity tagging** applied to the gateway.
- **Infrastructure as Code** with full Pulumi (Go) and Terraform (HCL) implementations producing identical outputs.
