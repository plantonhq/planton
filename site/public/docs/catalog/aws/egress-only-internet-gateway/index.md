---
title: "Egress-Only Internet Gateway"
description: "Egress-Only Internet Gateway deployment documentation"
icon: "package"
order: 100
componentName: "awsegressonlyinternetgateway"
---

# AWS Egress-Only Internet Gateway

Create an egress-only internet gateway and attach it to an AWS VPC. An egress-only internet gateway is the IPv6 counterpart of a NAT gateway — it lets dual-stack instances make **outbound** IPv6 connections to the internet while AWS blocks unsolicited **inbound** ones, at no charge.

## What Gets Created

- An **EC2 egress-only internet gateway**, attached to the specified VPC at creation.
- When `vpcId` changes on a later apply: the gateway is **replaced** (the attachment is immutable — AWS has no detach/re-attach API).

## Prerequisites

- An existing dual-stack **AwsVpc** (or a literal vpc-id) with an IPv6 CIDR to attach to.

## Quick Start

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsEgressOnlyInternetGateway
metadata:
  name: main-eigw
spec:
  region: us-west-2
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: my-vpc
      fieldPath: status.outputs.vpc_id
```

## Routing IPv6 Egress Through the Gateway

An egress-only gateway only provides connectivity once a subnet routes to it. Pair this gateway with an `AwsSubnet` whose IPv6 default route targets it:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsSubnet
metadata:
  name: private-usw2a
spec:
  region: us-west-2
  vpcId:
    value: vpc-0abc123
  availabilityZone: us-west-2a
  cidrBlock: 10.0.0.0/24
  routes:
    - destinationIpv6CidrBlock: ::/0
      targetType: egress_only_internet_gateway
      targetId:
        value: eigw-0abc123
```

## Configuration Reference

### Required

| Field | Description |
|---|---|
| `region` | AWS region (must match the VPC's region). |
| `vpcId` | The VPC to attach the gateway to. Literal id or a reference to an `AwsVpc`. Immutable — changing it replaces the gateway. |

## Stack Outputs

| Output | Description |
|---|---|
| `egress_only_internet_gateway_id` | The gateway's id — use this as a subnet IPv6 route's `targetId`. |
| `vpc_id` | The id of the VPC the gateway is attached to. |
| `region` | The region the gateway was created in. |

## Related Components

- **AwsVpc** — the dual-stack network the gateway attaches to.
- **AwsSubnet** — routes its IPv6 default route (`::/0`) to this gateway for outbound IPv6.
- **AwsNatGateway** — the IPv4 equivalent (outbound-only access for private IPv4 subnets).
