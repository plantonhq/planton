# Private NAT Gateway

A private NAT gateway (no Elastic IP) placed in an `AwsSubnet` by reference. A private gateway provides outbound access to other private networks — peered VPCs, a transit gateway, or an on-premises network over VPN/Direct Connect — with no internet exposure at all.

## When to Use

- Outbound communication from a private subnet to other VPCs or on-premises networks, where internet access is explicitly not wanted
- Overlapping-CIDR scenarios where source addresses must be translated without going to the internet
- Compliance environments that forbid any internet egress path

## Key Configuration Choices

- **Private connectivity** (`connectivityType: private`) — no Elastic IP is attached. AWS assigns a private IP from the subnet (override with `privateIp`, or add more with `secondaryPrivateIpAddresses` / `secondaryPrivateIpAddressCount`).
- **Compose by reference** — `subnetId` resolves from an `AwsSubnet`, so the gateway composes with its subnet without hardcoding ids.
- **Region** must match the referenced subnet's region.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<subnet-name>` | The name of the `AwsSubnet` to place the gateway in | The `metadata.name` of your `AwsSubnet` |

## Next Step

Route traffic destined for the remote private network through this gateway: add an `AwsSubnet` route whose `targetType` is `nat_gateway` and whose `targetId` is this gateway's `nat_gateway_id`.

## Related Presets

- **01-public-nat-gateway** — a public gateway (with an Elastic IP) for internet egress.
