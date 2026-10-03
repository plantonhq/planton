# Private NAT for Network Connectivity Center

This preset creates a PRIVATE-type NAT gateway that translates traffic between VPC networks attached as Network Connectivity Center spokes — the mechanism that lets two spokes with overlapping subnet ranges communicate.

## When to Use

- Hub-and-spoke topologies (Network Connectivity Center) where spoke VPCs have overlapping CIDR ranges — common after acquisitions or when teams provisioned networks independently
- Any spoke-to-spoke path that must not consume external IPs

## Prerequisites

1. The VPC is attached as a spoke to a Network Connectivity Center hub
2. A `GcpSubnetwork` with `purpose: PRIVATE_NAT` in the same region — its primary range provides the translated addresses

## Key Configuration Choices

- **`type: PRIVATE`** — no external IPs anywhere; the spec validation keeps `natIps` and `autoNetworkTier` empty because private NAT draws from subnetwork ranges instead
- **Rule-driven translation** — the rule matches traffic heading to the NCC hub (`nexthop.hub`) and translates it using the PRIVATE_NAT subnetwork's range
- **Explicit subnetwork scoping** — only the workload subnetwork's traffic is translated

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | GCP project ID | GCP Console or `GcpProject` outputs |
| `my-spoke-vpc` | Your spoke `GcpVpcNetwork` resource name | Your VPC manifest |
| `my-workload-subnet` | Your workload `GcpSubnetwork` name | Your subnetwork manifest |
| `my-private-nat-range` | Your PRIVATE_NAT-purpose `GcpSubnetwork` name | Your subnetwork manifest |
| `my-hub` | Your Network Connectivity Center hub name | NCC console or manifests |

## Related Presets

- **01-all-subnets-auto** — internet egress for a single VPC
- **02-static-ip-allowlisting** — stable public egress IPs

## Related Kinds

- [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — the PRIVATE_NAT-purpose range provider
- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — the spoke network
