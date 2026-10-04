# Azure Endpoint

This preset adds a public Azure resource -- here a Public IP fronting a regional deployment -- as a Traffic Manager destination. Azure tracks the target's address itself: if the IP changes, the endpoint follows, and its region feeds Performance routing automatically.

## When to Use

- Regional Azure deployments (load balancers' public IPs, App Services) behind one Traffic Manager name
- Anything Azure-hosted -- prefer this over an external endpoint so the address and region stay tracked

## Key Configuration Choices

- **Both edges are references** -- the profile by name, and the target by an explicit `valueFrom` naming the kind (no kind dominates Azure endpoint targets: Public IPs, App Services, and more all steer)
- **The target needs a PUBLIC address** -- Standard-tier Public IPs steer; Basic does not
- **Explicit priority and weight** -- priority 10 (gaps leave room for tiers later; unset would let creation order own the failover plan), weight 100 (a round base for percentage math on Weighted profiles); whichever the profile's routing method ignores is harmless

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-azure-traffic-manager-profile-resource-name>` | The AzureTrafficManagerProfile component's resource name | Your Planton catalog |
| `<your-azure-public-ip-resource-name>` | The AzurePublicIp component fronting the regional deployment | Your Planton catalog |
