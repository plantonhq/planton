# Production Cluster

This preset creates a Default-mode production cluster: dedicated M30 compute, zone-redundant high availability, MongoDB 8.0, and both authentication worlds enabled so applications onboard through Entra grants while the native administrator stays break-glass.

## When to Use

- The primary database for an application or service
- Lift-and-shift MongoDB workloads that need the real engine (change streams, transactions, aggregations)

## Key Configuration Choices

- **M30 with `ZoneRedundantPreferred`** -- the smallest general-purpose tier that supports zone-redundant HA; Free/M25 cannot
- **`MicrosoftEntraID` in the auth methods** -- required BEFORE any AzureMongoClusterUser grant can bind an app identity
- **Password by reference** -- rotation becomes a secret-store change; the username is create-only, so choose it deliberately
- **One shard** -- vertical scaling covers most workloads; shard count is create-only, so shard only with evidence

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-resource-group>` | The Planton name of your `AzureResourceGroup` resource | Planton console (or replace `valueFrom` with `value:` and a literal group name) |
| `<your-password-secret>` | The Planton name of the `AzureKeyVaultSecret` holding the admin password | Planton console |
| `203.0.113.0` / `203.0.113.255` | Your office/VPN egress IPv4 range (`firewallRules[].startIpAddress` / `endIpAddress`) | Your network team |
| `my-org-orders-db` | The cluster's global hostname label | Your naming convention -- org-prefixed |

## Related Presets

- **02 Free Sandbox** -- the zero-cost development sandbox
