# Public Internet Gateway (greenfield)

An internet gateway attached to a Planton-managed `AwsVpc` by reference. This is the canonical composition path: create an `AwsVpc`, create this gateway pointing at it, then give your public subnets a default route to the gateway's id.

## When to Use

- Building a VPC and its public connectivity together in Planton
- The internet path for public subnets (load balancers, bastions) and for the NAT gateways that private subnets egress through
- Any greenfield topology where the VPC is also managed as an `AwsInternetGateway` sibling

## Key Configuration Choices

- **Attach by reference** (`vpcId.valueFrom` -> `AwsVpc` `status.outputs.vpc_id`) — the platform resolves the VPC id at deploy time, so the gateway and VPC compose without hardcoding ids.
- **Region** must match the referenced VPC's region.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<awsvpc-name>` | The name of the `AwsVpc` resource to attach to | The `metadata.name` of your `AwsVpc` |

## Next Step

Make a subnet public by routing to this gateway: add an `AwsSubnet` with a `0.0.0.0/0` route whose `targetType` is `internet_gateway` and whose `targetId` is this gateway's `internet_gateway_id`.

## Related Presets

- **02-attach-to-existing-vpc** — attach to a VPC created outside Planton by literal vpc-id.
