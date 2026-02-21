# AlicloudVswitch

Manages an Alibaba Cloud VSwitch (virtual subnet).

## Overview

A VSwitch is the subnet equivalent in Alibaba Cloud networking. It carves out a CIDR range within a VPC and pins it to a single availability zone. ECS instances, RDS databases, container clusters, NAT gateways, and most other VPC-aware resources are deployed into a VSwitch.

Each VSwitch belongs to exactly one VPC and one availability zone. The CIDR block must be a subset of the parent VPC's CIDR block.

### What Gets Created

- **VSwitch** -- an isolated subnet within a VPC, bound to a specific availability zone

### Immutable Fields

The following fields cannot be changed after creation. Modifying them causes the VSwitch to be destroyed and recreated:

- `vpcId` -- the parent VPC
- `zoneId` -- the availability zone
- `cidrBlock` -- the IPv4 CIDR range

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Alibaba Cloud region (e.g., `cn-hangzhou`, `cn-shanghai`, `us-west-1`) |
| `vpcId` | StringValueOrRef | VPC ID; can be a literal or a reference to an AlicloudVpc component |
| `zoneId` | string | Availability zone (e.g., `cn-hangzhou-a`, `cn-hangzhou-b`) |
| `cidrBlock` | string | IPv4 CIDR block within the VPC range (mask 16-29) |
| `vswitchName` | string | VSwitch name (1-128 chars) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | string | `""` | Human-readable description (1-256 chars) |
| `enableIpv6` | bool | `false` | Allocate an IPv6 CIDR block to this VSwitch |
| `ipv6CidrBlockMask` | int32 | `0` | IPv6 CIDR mask (0-255); selects a /64 from the VPC's /56 IPv6 allocation |
| `tags` | map | `{}` | Key-value tags applied to the VSwitch |

### IPv6 Requirements

IPv6 on a VSwitch requires the parent VPC to have IPv6 enabled (`enableIpv6: true` on the AlicloudVpc). When both are enabled, Alibaba Cloud assigns a /64 IPv6 CIDR block to the VSwitch from the VPC's /56 allocation, selected by `ipv6CidrBlockMask`.

### CIDR Planning Guidance

VSwitch CIDRs must fall within the parent VPC's CIDR. Common patterns:

| VPC CIDR | VSwitch CIDR | Addresses | Typical Use |
|----------|-------------|-----------|-------------|
| `10.0.0.0/8` | `10.0.0.0/24` | 256 | Single tier in one AZ |
| `10.0.0.0/8` | `10.0.0.0/20` | 4,096 | Large tier (Kubernetes node pool) |
| `172.16.0.0/12` | `172.16.1.0/24` | 256 | Medium deployment tier |
| `192.168.0.0/16` | `192.168.0.0/24` | 256 | Dev/test single subnet |

Avoid overlapping CIDR blocks between VSwitches in the same VPC.

## Stack Outputs

| Output | Description |
|--------|-------------|
| `vswitch_id` | The VSwitch ID, referenced by downstream components via StringValueOrRef |
| `vswitch_name` | The VSwitch name as created |
| `cidr_block` | The IPv4 CIDR block |
| `zone_id` | The availability zone |
| `ipv6_cidr_block` | The IPv6 CIDR block (only set when IPv6 is enabled) |

## Related Components

- **AlicloudVpc** -- the parent VPC that this VSwitch belongs to
- **AlicloudSecurityGroup** -- security groups applied to resources in this VSwitch
- **AlicloudNatGateway** -- provides outbound internet access for resources in private VSwitches
- **AlicloudEcsInstance** -- compute instances deployed into this VSwitch
- **AlicloudRdsInstance** -- database instances deployed into this VSwitch
- **AlicloudPolardbCluster** -- cloud-native database clusters in this VSwitch
- **AlicloudRedisInstance** -- cache instances in this VSwitch
- **AlicloudAckManagedCluster** -- Kubernetes clusters using this VSwitch for node placement
- **AlicloudApplicationLoadBalancer** -- ALB load balancers spanning this VSwitch
- **AlicloudNetworkLoadBalancer** -- NLB load balancers spanning this VSwitch
- **AlicloudNasFileSystem** -- NAS mount targets in this VSwitch
