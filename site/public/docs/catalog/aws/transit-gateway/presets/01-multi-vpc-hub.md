---
title: "Multi-VPC Hub"
description: "Production Transit Gateway connecting application and shared-services VPCs with full-mesh routing. This is the most common Transit Gateway pattern, replacing complex VPC peering meshes with a..."
type: "preset"
rank: "01"
presetSlug: "01-multi-vpc-hub"
componentSlug: "transit-gateway"
componentTitle: "Transit Gateway"
provider: "aws"
icon: "package"
order: 1
---

# Multi-VPC Hub

Production Transit Gateway connecting application and shared-services VPCs with full-mesh routing. This is the most common Transit Gateway pattern, replacing complex VPC peering meshes with a centralized hub.

## When to Use

- Connecting 2+ VPCs that all need to communicate with each other
- Replacing VPC peering when you exceed the practical peering limit (~10 connections)
- Centralizing network traffic flow through a hub for simplified routing
- Production environments requiring DNS resolution across VPCs

## Key Configuration Choices

- **defaultRouteTableAssociation: true** -- all attachments are automatically associated with the default route table
- **defaultRouteTablePropagation: true** -- VPC CIDR blocks are automatically propagated, enabling full-mesh connectivity
- **dnsSupport: true** -- instances can resolve each other's private DNS hostnames across VPCs
- **vpnEcmpSupport: true** -- ready for future VPN connectivity with ECMP load balancing
- **Two VPC attachments** -- minimum production pattern with multi-AZ subnets

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<app-vpc-id>` | Application VPC ID | AwsVpc status.outputs.vpc_id |
| `<app-private-subnet-az1>` | App VPC private subnet in AZ1 | AwsSubnet status.outputs.subnet_id |
| `<app-private-subnet-az2>` | App VPC private subnet in AZ2 | AwsSubnet status.outputs.subnet_id |
| `<shared-services-vpc-id>` | Shared services VPC ID | AwsVpc status.outputs.vpc_id |
| `<shared-private-subnet-az1>` | Shared VPC private subnet in AZ1 | AwsSubnet status.outputs.subnet_id |
| `<shared-private-subnet-az2>` | Shared VPC private subnet in AZ2 | AwsSubnet status.outputs.subnet_id |

## Related Presets

- **02-single-vpc-development** -- simplified single-VPC setup for dev/test
- **03-hub-and-spoke-firewall** -- adds a firewall inspection VPC with appliance mode
