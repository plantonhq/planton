# AlicloudVpc

Manages an Alibaba Cloud Virtual Private Cloud (VPC).

## Overview

A VPC is the networking foundation for virtually every other Alibaba Cloud resource. It provides an isolated virtual network with its own CIDR block, virtual router, and system route table. VSwitches (subnets), security groups, NAT gateways, load balancers, database instances, and Kubernetes clusters are all deployed into a VPC.

### What Gets Created

- **VPC** -- an isolated virtual network with a primary IPv4 CIDR block
- **VRouter** -- automatically created with the VPC, manages route tables
- **System Route Table** -- the default route table for the VPC

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Alibaba Cloud region (e.g., `cn-hangzhou`, `cn-shanghai`, `us-west-1`) |
| `vpcName` | string | VPC name (1-128 chars; cannot start with `http://` or `https://`) |
| `cidrBlock` | string | Primary IPv4 CIDR block (e.g., `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | string | `""` | Human-readable description (1-256 chars) |
| `enableIpv6` | bool | `false` | Allocate an IPv6 CIDR block (/56) to the VPC |
| `resourceGroupId` | string | `""` | Resource group for organizational grouping |
| `tags` | map | `{}` | Key-value tags applied to the VPC |

### CIDR Block Guidance

Alibaba Cloud VPCs support private CIDR blocks from RFC 1918 ranges with a mask length of 8-28:

| Range | Typical Use |
|-------|-------------|
| `10.0.0.0/8` | Large deployments, many VSwitches |
| `172.16.0.0/12` | Medium deployments |
| `192.168.0.0/16` | Small deployments, dev/test environments |

Choose a CIDR that does not overlap with other VPCs if you plan to use VPC peering or CEN (Cloud Enterprise Network).

## Stack Outputs

| Output | Description |
|--------|-------------|
| `vpc_id` | The VPC ID, referenced by downstream components via StringValueOrRef |
| `vpc_name` | The VPC name as created |
| `cidr_block` | The primary IPv4 CIDR block |
| `router_id` | The virtual router ID |
| `route_table_id` | The system route table ID |

## Related Components

- **AlicloudVswitch** -- creates subnets within this VPC
- **AlicloudSecurityGroup** -- creates security groups bound to this VPC
- **AlicloudNatGateway** -- creates NAT gateways for outbound internet access
- **AlicloudAlbLoadBalancer** -- deploys ALB load balancers in this VPC
- **AlicloudNlbLoadBalancer** -- deploys NLB load balancers in this VPC
- **AlicloudAckManagedCluster** -- deploys Kubernetes clusters in this VPC
- **AlicloudRdsInstance** -- deploys database instances in this VPC (via VSwitch)
- **AlicloudPrivateZone** -- attaches private DNS zones to this VPC
