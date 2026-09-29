# Free Sandbox

This preset creates the zero-cost development sandbox: Azure's Free compute tier with the minimum storage, no high availability, and a single-address firewall rule for your dev machine.

## When to Use

- Learning the service or prototyping a data model
- CI fixtures and throwaway development databases

## Key Configuration Choices

- **`Free` tier** -- one per subscription (Azure's cap); rejects zone-redundant HA and sharding past one shard, both enforced at manifest time
- **32 GiB** -- the floor; storage grows in place later but never shrinks
- **Never let it become production** -- the tier ceiling arrives as throttling; the upgrade away from Free stages a tier-first update the provider performs itself

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-resource-group>` | The Planton name of your `AzureResourceGroup` resource | Planton console (or replace `valueFrom` with `value:` and a literal group name) |
| `<choose-a-password>` | A throwaway admin password (literal is acceptable for a sandbox) | Your choice |
| `198.51.100.7` | Your development machine's public IPv4 (`firewallRules[].startIpAddress` and `endIpAddress`) | `curl ifconfig.me` |

## Related Presets

- **01 Production Cluster** -- the dedicated-tier production shape with HA and Entra auth
