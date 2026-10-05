# All-Subnets Auto-Allocated NAT

This preset creates a Cloud Router with a NAT gateway that covers all subnets in the region using automatically allocated external IPs. This is the simplest and most common Cloud NAT configuration, ideal for giving private GKE nodes or Compute Engine VMs outbound internet access.

## When to Use

- Private GKE clusters that need internet egress for container image pulls
- Compute Engine VMs without external IPs that need outbound connectivity
- Any VPC where all subnets in a region should share NAT egress

## Key Configuration Choices

- **All subnets covered** — `subnetworks` is empty, so NAT applies to every subnetwork in the region, primary and secondary ranges
- **Auto-allocated IPs** — `natIps` is empty, so GCP provisions and scales the external IP pool itself; the IPs can change over time, so use the static-IP preset when third parties allowlist your egress
- **VPC by reference** — `vpcSelfLink` resolves the `GcpVpcNetwork` node's self link
- **Error-only logging** (`logFilter: ERRORS_ONLY`) — logs port exhaustion and connection failures without the volume of full translation logging

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | GCP project ID | GCP Console or `GcpProject` outputs |
| `my-app-vpc` | Your `GcpVpcNetwork` resource name | Your VPC manifest |

## Related Presets

- **02-static-ip-allowlisting** — stable egress IPs for partner allowlisting or compliance
- **03-private-nat** — NAT between VPC networks (Network Connectivity Center spokes)

## Related Kinds

- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — the network the router attaches to
- [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — scope NAT to specific subnetworks when needed
