# AwsSubnet — Design Rationale and Research

This document explains why the `AwsSubnet` spec is shaped the way it is, the trade-offs behind its routing model, and what is intentionally deferred.

## Why the subnet is its own component

A subnet is a real, independently ownable network node: it has its own lifecycle, its own CIDR, its own routing, and it is referenced by name across compute, data, and load-balancing resources. Modeling it as a standalone component (rather than a list bundled inside the VPC) lets a topology be composed from intent — an author or an agent enumerates exactly the subnets they want, in exactly the AZs they want, and wires each downstream resource to a specific subnet via a foreign-key reference.

## The routing model: folded into the subnet

In AWS, whether a subnet is "public" or "private" is **not a property of the subnet** — it is a property of the route table associated with it. A subnet with a default route to an internet gateway is public; one whose default route is a NAT gateway is private; one with no internet route is isolated. Because of this, routing belongs with the subnet, and the spec offers three modes through two mutually-exclusive fields:

- **Inline `routes`** — a dedicated route table is created, owned by the subnet, populated with the rules, and associated with it. `route_table_id` is exported so downstream resources can reference the created table.
- **External `route_table_id`** — an existing route table is associated with the subnet (the subnet does not own its lifecycle).
- **Neither** — the subnet uses the VPC main route table; `route_table_id` is exported empty.

A message-level CEL rule enforces that `route_table_id` and `routes` are never both set. This mirrors the established `OciSubnet` precedent (inline `route_rules` vs `route_table_id`), keeping the platform's grain consistent across providers.

There is deliberately **no standalone `AwsRouteTable` component**. No provider in the catalog models a bare route table; route *targets* (internet/NAT/transit gateways, peering connections, endpoints, ENIs) are the standalone nodes, referenced from within each route. This keeps the resource graph free of trivial glue nodes.

### Route target shape

Each route carries a `target_type` enum plus a single `target_id`, rather than a wide set of typed target fields. This matches the AWS API (where the target attribute is type-dependent) while staying compact and following the OCI `destination_type` + `network_entity_id` grain. As internet, NAT, and egress-only gateways become first-class kinds, they are referenced by setting `target_type` and a `value_from` `target_id` — no schema change is required here.

## Public/private and launch behavior

`map_public_ip_on_launch` controls only whether instances get an auto-assigned public IPv4; it does not by itself make a subnet reachable (that is the route table's job). The resource-name DNS records and `private_dns_hostname_type_on_launch` follow the AWS launch-time DNS model; `resource-name` hostnames are required for IPv6-only subnets.

## Dual-stack IPv6

`ipv6_cidr_block` attaches an IPv6 /64 (carved from a VPC-associated IPv6 CIDR), `assign_ipv6_address_on_creation` controls auto-assignment, and `enable_dns64` turns on NAT64 resolution. Inline routes support an IPv6 destination and an egress-only internet gateway target for outbound-only IPv6.

## What's intentionally deferred

Kept out of the initial surface to preserve good taste; each is additive and can land later without breaking the schema:

- **IPv6-native subnets** (`ipv6_native`) — would make `cidr_block` optional and add cross-field rules; the IPv4/dual-stack path covers the overwhelming majority of real use.
- **IPAM-driven CIDRs** (`ipv4_ipam_pool_id` / netmask length, IPv6 equivalents) — for organizations using AWS IPAM to hand out address space.
- **Outpost and Local Zone specifics** (`outpost_arn`, `customer_owned_ipv4_pool`, `map_customer_owned_ip_on_launch`, `enable_lni_at_device_index`) — edge/hybrid placements.
- **`availability_zone_id`** as an alternative to the AZ name.
- **Live E2E for inline routing** — exercised once internet/NAT gateways are first-class kinds to serve as route targets; the routing logic is covered today by schema validation and module parity.
