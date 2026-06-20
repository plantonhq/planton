---
title: "NAT Gateway"
description: "NAT Gateway deployment documentation"
icon: "package"
order: 100
componentName: "awsnatgateway"
---

# AWS NAT Gateway

Create a NAT gateway so private subnets can reach the internet (or other private networks) outbound-only. A public NAT gateway lives in a public subnet, is fronted by an Elastic IP, and is the route target a private subnet sends its default route to.

## What Gets Created

- An **EC2 NAT gateway** in the specified subnet.
- For a **public** gateway: the gateway is associated with the referenced Elastic IP allocation(s).
- For a **private** gateway: no Elastic IP is attached; AWS assigns private IPs from the subnet.

## Prerequisites

- An existing **AwsSubnet** (or a literal subnet-id) to place the gateway in. For a public gateway this must be a public subnet (routing to an internet gateway).
- For a public gateway, an **AwsElasticIp** (or a literal `eipalloc-` id) for the stable outbound address.

## Quick Start

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsNatGateway
metadata:
  name: main-nat
spec:
  region: us-west-2
  connectivityType: public
  subnetId:
    valueFrom:
      kind: AwsSubnet
      name: public-usw2a
      fieldPath: status.outputs.subnet_id
  allocationId:
    valueFrom:
      kind: AwsElasticIp
      name: nat-eip
      fieldPath: status.outputs.allocation_id
```

## Giving a Private Subnet Egress

A NAT gateway only provides egress once a subnet routes to it. Pair this gateway with an `AwsSubnet` whose default route targets it:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSubnet
metadata:
  name: private-usw2a
spec:
  region: us-west-2
  vpcId:
    value: vpc-0abc123
  availabilityZone: us-west-2a
  cidrBlock: 10.0.10.0/24
  routes:
    - destinationCidrBlock: 0.0.0.0/0
      targetType: nat_gateway
      targetId:
        value: nat-0abc123
```

## Configuration Reference

### Required

| Field | Description |
|---|---|
| `region` | AWS region (must match the subnet's region). |
| `connectivityType` | `public` (Elastic IP, internet egress) or `private` (no Elastic IP). |
| `subnetId` | The subnet to place the gateway in. Literal id or a reference to an `AwsSubnet`. |

### Public-gateway

| Field | Description |
|---|---|
| `allocationId` | Elastic IP allocation (required for public). Literal `eipalloc-` id or a reference to an `AwsElasticIp`. |
| `secondaryAllocationIds` | Additional Elastic IPs for very high-throughput egress. |

### Private-gateway

| Field | Description |
|---|---|
| `privateIp` | The private IPv4 address to assign (optional; AWS chooses if omitted). |
| `secondaryPrivateIpAddresses` / `secondaryPrivateIpAddressCount` | Additional private IPs (mutually exclusive). |

## Stack Outputs

| Output | Description |
|---|---|
| `nat_gateway_id` | The gateway's id — use this as a subnet route's `targetId`. |
| `public_ip` | The public IPv4 address of a public gateway (empty for private). |
| `private_ip` | The gateway's private IPv4 address within its subnet. |
| `network_interface_id` | The gateway's elastic network interface id. |
| `subnet_id` | The subnet the gateway lives in. |
| `region` | The region the gateway was created in. |

## Related Components

- **AwsSubnet** — both the placement of the gateway and the private subnets that route to it.
- **AwsElasticIp** — the stable outbound address for a public gateway.
- **AwsInternetGateway** — the internet path the public subnet (and thus the NAT gateway) routes through.
