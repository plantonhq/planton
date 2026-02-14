# Custom Mode VPC with Global Routing

This preset creates a VPC in custom subnet mode with global dynamic routing. Global routing enables Cloud Routers to advertise routes across all regions in the VPC, which is required for multi-region deployments or hybrid connectivity via Cloud VPN / Cloud Interconnect.

## When to Use

- Multi-region deployments where workloads need cross-region route visibility
- Hybrid connectivity setups using Cloud VPN or Cloud Interconnect
- Shared VPC networks that span multiple regions

## Key Configuration Choices

- **Custom mode** (`autoCreateSubnetworks: false`) -- subnets managed explicitly via `GcpSubnetwork`
- **Global routing** (`routingMode: GLOBAL`) -- routes learned by any Cloud Router are propagated to all regions
- **No Private Services Access** -- add it separately if needed for managed service private IPs

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the VPC will be created | GCP Console or `GcpProject` outputs |
| `<your-vpc-name>` | Name for this VPC network (1-63 chars, lowercase) | Choose a descriptive name (e.g., `global-vpc`) |

## Related Presets

- **01-custom-mode-regional** -- Use for single-region deployments with Private Services Access
