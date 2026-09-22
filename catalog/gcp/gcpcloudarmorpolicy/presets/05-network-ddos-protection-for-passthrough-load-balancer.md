# Network DDoS Protection for a Passthrough Load Balancer

## Use Case

Protect a region's passthrough Network Load Balancers (and protocol forwarding rules and public-IP VMs) at the packet layer. A `CLOUD_ARMOR_NETWORK` policy turns on Google's network DDoS protection and filters packets on source ranges, protocols, ports, origin ASNs, and custom bytes read from the packet header -- the things a passthrough load balancer sees, since it never terminates HTTP.

## When to Use

- An internal or external passthrough NLB (the `INTERNAL` / `EXTERNAL` regional `GcpBackendService` behind a regional `GcpGlobalForwardingRule`)
- Protocol forwarding or a fleet of VMs with public IPs in one region
- The first step toward Cloud Armor Enterprise's advanced network DDoS protection (`ADVANCED_PREVIEW`, then `ADVANCED`, with `networkEdgeSecurityService` declared on one policy per region)

## What This Creates

- A REGIONAL Cloud Armor policy of type `CLOUD_ARMOR_NETWORK` in `us-central1`
- `STANDARD` network DDoS protection -- always on, free with the load balancer
- One user-defined field: two bytes at TCP offset 8, masked `0x8F00`
- Priority 100 (preview): allow TCP/443 from `10.10.0.0/16` whose signed field reads `0x8F00`
- Priority 200: deny(403) packets from two origin ASNs
- Priority 2147483647: the mandatory default rule, written as an empty `networkMatch` (every packet)

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `region` | `us-central1` | The region of the load balancers this policy protects. Immutable. |
| `ddosProtection` | `STANDARD` | `ADVANCED_PREVIEW` to observe Enterprise mitigations, `ADVANCED` to enforce them -- both need the project enrolled in Cloud Armor Enterprise and `networkEdgeSecurityService` declared. |
| `userDefinedFields` | one TCP field | Add fields for the bytes your protocol signs; up to 4 bytes each, names unique, referenced by name from rules. |
| `srcAsns` (priority 200) | two documentation ASNs | The autonomous systems you never serve; leave the list empty to drop the rule. |
| default rule `action` | `allow` | `deny(403)` makes the allowlist the guard. |

A network policy has no HTTP `match`, no labels, no Adaptive Protection, no redirect, and no header injection -- each is rejected before deploy. Without an Enterprise subscription, `ADVANCED` is accepted by the API and does nothing more than `STANDARD`; the module cannot check the subscription for you.
