# Network DDoS Protection for a Passthrough Load Balancer

## Use Case

Protect a region's passthrough Network Load Balancers (and protocol forwarding rules and public-IP VMs) at the packet layer. A `CLOUD_ARMOR_NETWORK` policy turns on Google's network DDoS protection and filters packets on source ranges, protocols, ports, origin ASNs, and custom bytes read from the packet header -- the things a passthrough load balancer sees, since it never terminates HTTP.

## When to Use

- An internal or external passthrough NLB (the `INTERNAL` / `EXTERNAL` regional `GcpBackendService` behind a regional `GcpGlobalForwardingRule`)
- Protocol forwarding or a fleet of VMs with public IPs in one region
- A project enrolled in Cloud Armor Enterprise, in a region where advanced network DDoS protection is already on

## Before You Start

Packet rules and user-defined fields are a Cloud Armor Enterprise feature. Google refuses this policy at create in a region without advanced network DDoS protection ("Creating network security policies with custom rules requires advanced network DDoS protection to be enabled in the region"). Turn it on first with a separate `GcpCloudArmorPolicy` in the same region: `type: CLOUD_ARMOR_NETWORK`, `ddosProtectionConfig.ddosProtection: ADVANCED` (or `ADVANCED_PREVIEW` to observe first), and `networkEdgeSecurityService` declared (one per region per project). Google says not to reuse that policy for rules, so this preset is the second policy. In a region on `STANDARD` protection only, a network policy carries no rules and no user-defined fields, and Google adds its default rule.

## What This Creates

- A REGIONAL Cloud Armor policy of type `CLOUD_ARMOR_NETWORK` in `us-central1`, the region's packet filter
- One user-defined field: two bytes at TCP offset 8, masked `0x8F00`
- Priority 100 (preview): allow TCP/443 from `10.10.0.0/16` whose signed field reads `0x8F00`
- Priority 200: deny(403) packets from two origin ASNs
- Priority 2147483647: the mandatory default rule, written as an empty `networkMatch` (every packet)

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `region` | `us-central1` | The region of the load balancers this policy protects. Immutable. |
| `userDefinedFields` | one TCP field | Add fields for the bytes your protocol signs; up to 4 bytes each, names unique, referenced by name from rules. |
| `srcAsns` (priority 200) | two documentation ASNs | The autonomous systems you never serve. To drop the rule, delete it: an empty network match matches every packet, so the rule with no ASNs would deny everything. |
| default rule `action` | `allow` | `deny(403)` makes the allowlist the guard. |

A network policy has no HTTP `match`, no labels, no Adaptive Protection, no redirect, and no header injection -- each is rejected before deploy. The module cannot check the Enterprise subscription or the region's enrollment for you; Google's refusal at create is the check.
