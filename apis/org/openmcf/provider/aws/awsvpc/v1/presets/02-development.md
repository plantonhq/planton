# Development VPC

This preset creates a minimal IPv4-only VPC with a `/16` CIDR and DNS enabled --
a clean foundation for development and testing networks. Add `AwsSubnet`
components (and, if outbound internet access is needed, an `AwsInternetGateway` or
`AwsNatGateway`) that reference this VPC to build out the environment.

## When to Use

- Development and testing environments
- Quick sandbox networks for prototyping
- Any case where IPv6 and secondary CIDRs are not (yet) needed

## Key Configuration Choices

- **/16 IPv4 CIDR** (`cidrBlock: 10.0.0.0/16`) -- ample room for development subnets
- **DNS enabled** -- DNS hostnames and support are on for service-discovery
  compatibility

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region code (e.g., `us-east-1`) | Your deployment region |

## Related Presets

- **01-production-dual-stack** -- a dual-stack (IPv4 + IPv6) production foundation
