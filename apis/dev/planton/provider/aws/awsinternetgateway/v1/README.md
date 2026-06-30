# Overview

The **AWS Internet Gateway API Resource** provides a consistent, standardized interface for deploying and managing an internet gateway and attaching it to an AWS Virtual Private Cloud (VPC). An internet gateway is the VPC's door to the public internet: a horizontally scaled, redundant, AWS-managed component that enables bidirectional IPv4 (and inbound/outbound IPv6 for dual-stack VPCs) traffic between the VPC and the internet.

## Purpose

This resource makes the internet gateway a first-class, independently composable building block rather than something bundled inside the VPC. By offering a unified interface, it lets you:

- **Compose public connectivity deliberately**: Create the gateway as its own graph node and attach it to exactly the VPC you intend, by reference or by literal id.
- **Wire it as a route target**: The gateway's id is exported so an `AwsSubnet` can send a default route (`0.0.0.0/0`, or `::/0` for IPv6) to it — the step that actually makes a subnet public.
- **Own its lifecycle**: The gateway is created, attached, and destroyed independently of the VPC, so it can be reasoned about, referenced, and managed on its own.

## Key Features

- **Single attachment**: Attaches to one VPC (`vpcId`), by reference to an `AwsVpc` or a literal vpc-id. A VPC can have at most one internet gateway attached at a time.
- **Updatable attachment**: Changing `vpcId` detaches the gateway from the old VPC and attaches it to the new one — the gateway itself is not replaced.
- **Consistent tagging**: Standard resource-identity tags are applied to the gateway.
- **Dual-engine parity**: Identical behavior and outputs from the Pulumi (Go) and Terraform (HCL) modules.

## How the Internet Gateway Fits the Network Model

- **Attachment is not exposure**: Attaching a gateway to a VPC does nothing on its own. A subnet becomes "public" only when its route table sends a default route to this gateway. The standard public-subnet recipe is this gateway plus an `AwsSubnet` whose route targets it (`targetType: internet_gateway`).
- **Internet gateway vs egress-only**: An internet gateway allows inbound and outbound traffic. For IPv6-only outbound access with no inbound exposure, use an egress-only internet gateway instead.
- **Internet gateway vs NAT gateway**: An internet gateway connects a VPC to the internet for resources with public IPs. A NAT gateway gives private-subnet instances outbound-only access and lives in a public subnet that itself routes to an internet gateway.

## Critical Constraints

- **VPC id**: Required. A VPC may have only one internet gateway attached at a time. Updatable — changing it re-attaches the gateway to a different VPC.
- **Region**: Required, must match the VPC's region.

## Use Cases

- **Public-facing VPCs**: Provide the internet path for public subnets hosting load balancers, bastions, or internet-facing services.
- **NAT egress topologies**: The internet path that public subnets (and the NAT gateways living in them) route through, enabling outbound access for private subnets.
- **Brownfield attachment**: Attach a Planton-managed gateway to a VPC created outside Planton by supplying its literal vpc-id.

## Production Features

- **Independent, referenceable gateway** exported as a route target for downstream subnets.
- **Consistent resource-identity tagging** applied to the gateway.
- **Infrastructure as Code** with full Pulumi (Go) and Terraform (HCL) implementations producing identical outputs.
- **Composability** as a foundational building block referenced by `AwsSubnet` routes via `StringValueOrRef`.
