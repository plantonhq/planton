# Overview

The **AWS Egress-Only Internet Gateway API Resource** provides a consistent, standardized interface for deploying and managing an egress-only internet gateway and attaching it to an AWS Virtual Private Cloud (VPC). An egress-only internet gateway is the IPv6 counterpart of a NAT gateway: a horizontally scaled, redundant, AWS-managed component that lets instances in a dual-stack VPC initiate **outbound** IPv6 traffic to the internet while AWS statefully blocks any unsolicited **inbound** IPv6 connections.

## Purpose

This resource makes the egress-only internet gateway a first-class, independently composable building block. By offering a unified interface, it lets you:

- **Give private IPv6 workloads outbound access**: Let instances with only IPv6 addresses reach the internet (package mirrors, APIs, telemetry) without being reachable from it.
- **Wire it as an IPv6 route target**: The gateway's id is exported so an `AwsSubnet` can send its IPv6 default route (`::/0`) to it.
- **Own its lifecycle**: The gateway is created and destroyed independently of the VPC, so it can be reasoned about, referenced, and managed on its own.

## Key Features

- **Single attachment**: Attaches to one VPC (`vpcId`), by reference to an `AwsVpc` or a literal vpc-id. The VPC should be dual-stack (have an IPv6 CIDR) for the gateway to be useful.
- **No charge**: Unlike a NAT gateway, an egress-only internet gateway has no per-hour or per-GB cost.
- **Consistent tagging**: Standard resource-identity tags are applied to the gateway.
- **Dual-engine parity**: Identical behavior and outputs from the Pulumi (Go) and Terraform (HCL) modules.

## How the Egress-Only Gateway Fits the Network Model

- **Attachment is not routing**: Attaching a gateway to a VPC does nothing on its own. A subnet uses it only when its route table sends the IPv6 default route (`::/0`) to this gateway. The standard recipe is this gateway plus an `AwsSubnet` whose route targets it (`targetType: egress_only_internet_gateway`).
- **Egress-only vs internet gateway**: An internet gateway allows bidirectional traffic; an egress-only internet gateway allows IPv6 outbound only, blocking unsolicited inbound. There is no IPv6 NAT — every IPv6 address is globally routable, so the egress-only gateway is what provides the "outbound-but-not-inbound" guarantee for IPv6.
- **Egress-only vs NAT gateway**: A NAT gateway provides the same outbound-only guarantee for **IPv4** (and costs per hour and per GB); an egress-only internet gateway does it for **IPv6** at no charge.

## Critical Constraints

- **VPC id**: Required and **immutable**. AWS attaches the gateway to the VPC at creation and provides no detach/re-attach API, so changing it replaces the gateway (ForceNew).
- **Region**: Required, must match the VPC's region.

## Use Cases

- **Private IPv6 subnets with outbound access**: Let dual-stack private instances pull updates and call external APIs over IPv6 without inbound exposure.
- **Cost-sensitive IPv6 egress**: Replace NAT-gateway charges for traffic that can use IPv6 end to end.
- **Brownfield attachment**: Add a Planton-managed egress-only gateway to a VPC created outside Planton by supplying its literal vpc-id.

## Production Features

- **Independent, referenceable gateway** exported as an IPv6 route target for downstream subnets.
- **Consistent resource-identity tagging** applied to the gateway.
- **Infrastructure as Code** with full Pulumi (Go) and Terraform (HCL) implementations producing identical outputs.
- **Composability** as a foundational building block referenced by `AwsSubnet` IPv6 routes via `StringValueOrRef`.
