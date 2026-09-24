# DigitalOcean Database Replica -- Operational Guide

What experience with this component teaches that the field reference cannot.

## Why region and size are required here

The upstream provider marks both optional ("inherit the primary's"), but reads them back unconditionally without computing them -- so an omitted value diffs on the next apply, and because region is create-only, that diff schedules a FULL REPLICA REPLACEMENT. This spec requires both fields: writing the primary's region and size explicitly is the same outcome with none of the landmine. When the primary resizes, revisit the replica's size too (it must stay >= the primary's).

## Tags replace the replica

Replica tags are create-only upstream. Changing the tag list REPLACES the replica: DigitalOcean seeds a fresh one from the primary and the old endpoint dies. No primary data is at risk, but read consumers see the endpoint churn and the reseed takes cluster-create time. Decide the tag set at birth (the modules add the standard resource-metadata tags automatically -- account for them when planning any future retag).

## Combined tags are capped at 255 characters -- and the Planton labels count

DigitalOcean caps a replica's combined tags -- every tag name joined by commas -- at 255 characters, exactly the rule its primary's create follows (measured 2026-09-17 on the replica endpoint: 255 passes, 256 fails with `422 combined tags cannot exceed 255 characters`, and the refused create leaves nothing behind). The six Planton label tags carry your `metadata.name` and `metadata.id`, so they alone cost about 133 characters plus those two lengths; a 47-character name spends the whole budget before a single `tags` entry. Both provisioners check the final tag set before anything renders and fail with the arithmetic (how many tags, how many characters, how much the labels used), so you learn the rule in seconds -- never as a replacement of a running replica. The two levers are a shorter `metadata.name`/`metadata.id` and fewer `tags`.

## Creation is as slow as a cluster

A replica seeds from the primary's backup chain; DigitalOcean even retries creation through 412 responses while a young primary's FIRST backup completes. Budget the same ~5 minutes you budget for a cluster, and expect brand-new primaries to add delay.

## The replica has its own firewall -- the primary's rules do not reach it

To DigitalOcean a replica is its own cluster: `replica_id` is a cluster UUID (`GET /v2/databases/{replica_id}` answers for it) with its own trusted-sources list, which starts EMPTY -- measured 2026-09-17 with a primary carrying one `ip_addr` rule and its fresh replica reading `rules: []`. A `DigitalOceanDatabaseFirewall` on the primary protects the primary alone; declare a second one per replica with `cluster` pointing at the replica (`valueFrom: {kind: DigitalOceanDatabaseReplica, name: <replica>, fieldPath: status.outputs.replica_id}`, or the UUID as a literal), or the replica stays reachable from anywhere with the primary's credentials. The replica also mints its own three DigitalOcean alert policies (CPU, memory, disk) when it comes online, exactly as the primary does.

## Cross-region replicas and VPCs

A cross-region replica joins the REPLICA region's VPC (the primary's VPC does not span regions) -- wire `vpc` to a VPC in `region`. Both are create-only; moving a replica between regions or networks is a replace.

## What a replica is NOT

Not automatic failover -- DigitalOcean offers manual console promotion, not managed HA (the primary's `node_count` standbys are the HA story). Not a backup -- replicas follow deletes and corruption faithfully; the primary's backups are the recovery path. Not writable -- writes go to the primary, always.

## Resize choreography

`size` and `storage_size_mib` change IN PLACE (the one update path), waited to "online". Grow storage with size when the slug's default would shrink below the current allocation -- storage can never decrease.

## What is deliberately NOT here

Promote-to-primary (exists in DigitalOcean's API but is not bridged by the provider -- a recorded absence to re-evaluate), replica-level users/databases (the primary owns them), and inline trusted-source rules (the replica's own rule set, which starts empty, is declared as a separate `DigitalOceanDatabaseFirewall` whose `cluster` references this replica -- see "The replica has its own firewall" above).
