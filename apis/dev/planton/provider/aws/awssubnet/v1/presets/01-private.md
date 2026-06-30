# Private Subnet

A private subnet whose default route points at a NAT gateway: instances can reach the internet outbound (patching, image pulls, external APIs) but cannot be reached from it. This is the standard placement for application tiers, EKS/ECS worker nodes, and internal services in a production VPC.

## When to Use

- Application backends, worker nodes, and internal APIs that need outbound internet but must not be publicly reachable
- The private half of a public/private VPC topology
- Any workload where instances should never receive a public IP

## Key Configuration Choices

- **NAT default route** (`0.0.0.0/0` via `nat_gateway`) — outbound-only internet access through a NAT gateway living in a public subnet. Creating `routes` makes this subnet own a dedicated route table.
- **No public IP on launch** — `mapPublicIpOnLaunch` is left at its default (`false`), so instances come up without a public IPv4.
- **CIDR** (`10.0.1.0/24`) — 256 addresses; pair it with a public subnet at `10.0.0.0/24` in the same VPC. Adjust to your address plan.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<vpc-id>` | The VPC to create the subnet in | `AwsVpc` status outputs (`vpc_id`), or reference an `AwsVpc` via `valueFrom` |
| `<nat-gateway-id>` | The NAT gateway to route outbound traffic through | The NAT gateway in the corresponding public subnet |

## Related Presets

- **02-public** — for subnets hosting internet-facing resources (load balancers, bastions)
- **03-isolated** — for data-tier subnets with no internet path at all
