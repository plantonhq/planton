# Attach to an Existing VPC (brownfield)

An egress-only internet gateway attached to a dual-stack VPC that already exists outside Planton, by literal vpc-id. Use this when the VPC is managed elsewhere (an existing landing zone, another tool, or a hand-created network) but you want Planton to own the IPv6 egress gateway.

## When to Use

- Adding IPv6 outbound connectivity to a pre-existing, externally-managed dual-stack VPC
- Incrementally adopting Planton in an account that already has networking
- Any case where you have a concrete `vpc-id` rather than a Planton `AwsVpc` resource

## Key Configuration Choices

- **Attach by literal id** (`vpcId.value`) — point directly at an existing VPC. This is the same field as the greenfield preset, using the literal arm of the reference instead of `valueFrom`.
- **Region** must match the existing VPC's region.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<vpc-id>` | The id of the existing dual-stack VPC to attach to | AWS console / CLI (`aws ec2 describe-vpcs`) |

## Note

The VPC must have an IPv6 CIDR for the egress-only gateway to be useful. Changing the attachment later replaces the gateway (the attachment is immutable).

## Related Presets

- **01-ipv6-egress** — attach to a Planton-managed `AwsVpc` by reference (the greenfield path).
