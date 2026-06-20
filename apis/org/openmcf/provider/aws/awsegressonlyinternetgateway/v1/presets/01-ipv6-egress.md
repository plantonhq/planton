# IPv6 Egress (greenfield)

An egress-only internet gateway attached to a Planton-managed dual-stack `AwsVpc` by reference. This is the canonical composition path: create an `AwsVpc` with an IPv6 CIDR, create this gateway pointing at it, then give your private subnets an `::/0` route to the gateway's id.

## When to Use

- Dual-stack VPCs where private instances need **outbound** IPv6 internet access but must not be reachable from the internet
- Cost-sensitive IPv6 egress (no per-hour or per-GB charge, unlike a NAT gateway)
- Any greenfield topology where the VPC is also managed in Planton

## Key Configuration Choices

- **Attach by reference** (`vpcId.valueFrom` -> `AwsVpc` `status.outputs.vpc_id`) — the platform resolves the VPC id at deploy time, so the gateway and VPC compose without hardcoding ids.
- **Region** must match the referenced VPC's region.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<awsvpc-name>` | The name of the dual-stack `AwsVpc` resource to attach to | The `metadata.name` of your `AwsVpc` |

## Next Step

Route IPv6 egress through this gateway: add an `AwsSubnet` with a `::/0` route whose `targetType` is `egress_only_internet_gateway` and whose `targetId` is this gateway's `egress_only_internet_gateway_id`.

## Related Presets

- **02-attach-to-existing-vpc** — attach to a VPC created outside Planton by literal vpc-id.
