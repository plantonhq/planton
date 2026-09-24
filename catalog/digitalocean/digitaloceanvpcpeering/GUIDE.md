# DigitalOcean VPC Peering -- Operational Guide

What experience with this component teaches that the field reference cannot.

## CIDR planning happens before the first VPC, not before the peering

DigitalOcean rejects peerings between VPCs with overlapping IP ranges, and a VPC's range is create-only -- discovering an overlap at peering time means REBUILDING one of the networks. If two environments might ever need to talk, give their VPCs disjoint ranges on day one.

## Deleting a peering is a traffic event

The moment the peering goes, cross-VPC routes vanish and everything that talked across it starts timing out -- databases in the data VPC, internal APIs, all of it. Trace what depends on the link before destroying it; there is no drain or grace period.

## All-or-nothing reachability -- trust accordingly

A peering exposes every private address in each VPC to the other. There is no per-CIDR filtering on the link itself; narrowing access is droplet-firewall work (the DigitalOceanFirewall kind) on each host. Peer networks that genuinely trust each other; for one-service access, prefer exposing that service properly instead of peering whole networks.

## No transit -- peer pairwise

DigitalOcean peering is non-transitive: A-B and B-C do not give A-C. A hub-VPC topology needs an explicit peering per spoke pair that must communicate. Each pair is its own instance of this kind -- charts compose them cleanly.

## Settling takes minutes, and deletes retry through 403s

Creates wait for ACTIVE and deletes ride out DigitalOcean's transient 403 responses while the peering settles (the provider handles both within 2-minute windows). Back-to-back create/destroy cycles on the same VPC pair can briefly collide with a peering still DELETING -- give teardowns a moment before re-peering.

## What is deliberately NOT here

Route filtering and per-CIDR rules (no provider surface -- droplet firewalls own per-host restriction); cross-region cost knobs (peerings are free, cross-region included); and `created_at` or `status` as outputs (the provider renders `created_at` in Go's non-standard time format, not RFC 3339; a status captured at apply time can only ever say ACTIVE and goes stale the moment DigitalOcean moves the peering -- read the lifecycle live from `GET /v2/vpcs/peerings/{id}` or the control panel).
