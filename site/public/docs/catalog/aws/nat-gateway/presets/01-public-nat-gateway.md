---
title: "Public NAT Gateway (greenfield)"
description: "A public NAT gateway placed in a Planton-managed public `AwsSubnet` and fronted by a Planton-managed `AwsElasticIp`, both by reference. This is the canonical egress path: create a public subnet..."
type: "preset"
rank: "01"
presetSlug: "01-public-nat-gateway"
componentSlug: "nat-gateway"
componentTitle: "NAT Gateway"
provider: "aws"
icon: "package"
order: 1
---

# Public NAT Gateway (greenfield)

A public NAT gateway placed in a Planton-managed public `AwsSubnet` and fronted by a Planton-managed `AwsElasticIp`, both by reference. This is the canonical egress path: create a public subnet (routing to an internet gateway), allocate an Elastic IP, create this gateway pointing at both, then give your private subnets a default route to the gateway's id.

## When to Use

- Giving private-subnet workloads outbound internet access (package repos, third-party APIs, AWS service endpoints)
- High-availability egress: one public NAT gateway per availability zone, each in that zone's public subnet
- Any greenfield topology where the subnet and Elastic IP are also managed in Planton

## Key Configuration Choices

- **Public connectivity** (`connectivityType: public`) — the gateway is fronted by an Elastic IP and provides internet egress.
- **Compose by reference** — `subnetId` resolves from an `AwsSubnet` and `allocationId` from an `AwsElasticIp`, so the gateway composes with its dependencies without hardcoding ids.
- **Region** must match the referenced subnet's region.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<public-subnet-name>` | The name of the public `AwsSubnet` to place the gateway in | The `metadata.name` of your public `AwsSubnet` |
| `<elastic-ip-name>` | The name of the `AwsElasticIp` to use as the outbound address | The `metadata.name` of your `AwsElasticIp` |

## Next Step

Give a private subnet egress by routing to this gateway: add an `AwsSubnet` with a `0.0.0.0/0` route whose `targetType` is `nat_gateway` and whose `targetId` is this gateway's `nat_gateway_id`.

## Related Presets

- **02-private-nat-gateway** — a private gateway (no Elastic IP) for egress to peered/transit/on-premises networks.
