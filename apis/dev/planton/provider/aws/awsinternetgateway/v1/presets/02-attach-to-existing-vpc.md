# Attach to an Existing VPC (brownfield)

An internet gateway attached to a VPC that already exists outside Planton, by literal vpc-id. Use this when the VPC is managed elsewhere (an existing landing zone, another tool, or a hand-created network) but you want Planton to own the internet gateway.

## When to Use

- Adding public connectivity to a pre-existing, externally-managed VPC
- Incrementally adopting Planton in an account that already has networking
- Any case where you have a concrete `vpc-id` rather than a Planton `AwsVpc` resource

## Key Configuration Choices

- **Attach by literal id** (`vpcId.value`) — point directly at an existing VPC. This is the same field as the greenfield preset, using the literal arm of the reference instead of `valueFrom`.
- **Region** must match the existing VPC's region.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<vpc-id>` | The id of the existing VPC to attach to | AWS console / CLI (`aws ec2 describe-vpcs`) |

## Note

A VPC can have at most one internet gateway attached at a time. If the existing VPC already has one, detach or remove it before attaching this gateway.

## Related Presets

- **01-public-internet-gateway** — attach to a Planton-managed `AwsVpc` by reference (the greenfield path).
