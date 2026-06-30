---
title: "Internet Gateway"
description: "Internet Gateway deployment documentation"
icon: "package"
order: 100
componentName: "awsinternetgateway"
---

# AWS Internet Gateway

Create an internet gateway and attach it to an AWS VPC. An internet gateway is the VPC's door to the public internet — the route target a public subnet sends its default route to so that internet-facing resources can be reached and reach out.

## What Gets Created

- An **EC2 internet gateway**, attached to the specified VPC.
- When `vpcId` changes on a later apply: the gateway is **re-attached** to the new VPC (it is not recreated).

## Prerequisites

- An existing **AwsVpc** (or a literal vpc-id) to attach to. A VPC can have only one internet gateway attached at a time.

## Quick Start

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsInternetGateway
metadata:
  name: main-igw
spec:
  region: us-west-2
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: my-vpc
      fieldPath: status.outputs.vpc_id
```

## Making a Subnet Public

An internet gateway only provides connectivity once a subnet routes to it. Pair this gateway with an `AwsSubnet` whose default route targets it:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsSubnet
metadata:
  name: public-usw2a
spec:
  region: us-west-2
  vpcId:
    value: vpc-0abc123
  availabilityZone: us-west-2a
  cidrBlock: 10.0.0.0/24
  mapPublicIpOnLaunch: true
  routes:
    - destinationCidrBlock: 0.0.0.0/0
      targetType: internet_gateway
      targetId:
        value: igw-0abc123
```

## Configuration Reference

### Required

| Field | Description |
|---|---|
| `region` | AWS region (must match the VPC's region). |
| `vpcId` | The VPC to attach the gateway to. Literal id or a reference to an `AwsVpc`. |

## Stack Outputs

| Output | Description |
|---|---|
| `internet_gateway_id` | The gateway's id — use this as a subnet route's `targetId`. |
| `internet_gateway_arn` | The gateway's ARN. |
| `vpc_id` | The id of the VPC the gateway is attached to. |
| `region` | The region the gateway was created in. |

## Related Components

- **AwsVpc** — the network the gateway attaches to.
- **AwsSubnet** — routes a default route to this gateway to become public.
- **AwsElasticIp** — a stable public IP for a NAT gateway in a public subnet that itself routes here.
