# DigitalOcean Database Cluster -- Operational Guide

Judgment calls that matter when you run managed databases on DigitalOcean.

## Pick the engine slug, not the marketing name

The `engine` values are DigitalOcean's own API slugs: PostgreSQL is `pg`, never `postgres`. Redis and Valkey are two slugs for one caching product line, and DigitalOcean has finished the move: a new cluster with `engine: redis` is rejected at the API (measured 2026-09-16 — the error reads `region 'nyc3' is not valid`, DigitalOcean's way of saying the engine is offered in no region). Keep `redis` only for adopting a cluster that already exists; every new cache is `valkey`.

## The version must be one DigitalOcean offers today

`engineVersion` is checked against DigitalOcean's live offer list, not against a format. `GET /v2/databases/options` names the versions per engine, and a value not on it fails at create with `422 invalid cluster engine version` — a bare `"8"` for MySQL fails today because the only MySQL line offered is `"8.4"`. The list moves: PostgreSQL 15 leaves the offer in May 2027, and majors retire on a schedule DigitalOcean publishes in the same response. When a deploy fails with that error, read the options endpoint and raise the version; the presets in this component name the version they were verified against and the date.

## "cluster name is not available" means more than one thing

Cluster names are unique per account, and a real duplicate is refused with `422 cluster name is not available`. DigitalOcean gives the same 422 for a brand-new name in two other situations, measured on 2026-09-16: a create whose tag set the database service will not take (the identical request without tags succeeds; every one of Planton's label tags succeeds alone; all seven together — about 266 characters for a 47-character resource name whose `id` equals its `name` — fail every time; DigitalOcean documents no such budget), and any create in the few minutes after a failed create, tagged or not — a failed create can also leave a ghost cluster that answers 404, is missing from the list, and still counts as a member of its VPC. If you see this error on a name you know is free: wait a few minutes, then retry once before changing anything; if it persists, shorten the resource name or drop `spec.tags`, and when a VPC later refuses to delete, look at `GET /v2/vpcs/{id}/members` for a `do:dbaas:` URN with an empty name. The exact rule is still being characterized; until it is, treat long tag sets on long names as a risk.

## Every cluster comes with three alert policies you did not declare

When a cluster reaches `online`, DigitalOcean creates three monitoring alert policies for it — CPU, memory, and disk utilization above 90% over five minutes, emailing a team member — and they are not part of this resource: neither engine manages them, and deleting the cluster leaves them behind, still pointing at the deleted cluster's UUID. Expect them in `GET /v2/monitoring/alerts` (type `v1/dbaas/alerts/...`) and delete the ones whose cluster is gone; a DigitalOceanMonitorAlert manifest is the way to own the alerting you actually want.

## Node count is an engine decision

- **PostgreSQL / MySQL / MongoDB**: 1 node for dev, 3 for production failover. 2 buys a standby without quorum; most teams go straight to 3.
- **Redis / Valkey**: 1 node is normal — caches tolerate a failover gap. Add standbys only when cache warm-up is expensive.
- **Kafka**: 3 is the floor; DigitalOcean rejects less.
- **OpenSearch**: 1–15; go multi-node when the index must survive node loss rather than for query speed.

DigitalOcean enforces these server-side; the spec only enforces the universal minimum of 1 so new engine rules never require a contract change.

## Storage: plan for growth, not shrinkage

`storageGib` only ever grows. Two practical rules:

- Prefer `storageAutoscale` over hand-managed increments — DigitalOcean grows the disk at your threshold with a one-hour cooldown.
- If you grow `sizeSlug` while `storageGib` is unset, the cluster adopts the new slug's (larger) default storage automatically. A stale explicit `storageGib` smaller than the new slug's default is invalid — unset it when upsizing.

Note `storageAutoscale` currently deploys through the Terraform provisioner only; the Pulumi bridge rejects it loudly (no silent drop). Choose the provisioner accordingly or manage growth manually on Pulumi stacks.

## Upgrades are one-way and live

Raising `engineVersion` performs an in-place major upgrade on the running cluster. There is no downgrade, and no blue-green: take a backup-restore copy first if the application's compatibility is unproven. Region changes are similar — a live migration, not a recreate — expect elevated latency while it runs.

## VPC placement is create-only

`vpc` cannot be changed after the cluster exists. Decide network placement first; retrofitting means a new cluster plus a data migration (`backupRestore` gives you the copy).

## Restoring from backup

`backupRestore.databaseName` names the SOURCE cluster; omit `backupCreatedAt` to take the newest backup. The block acts only at creation and is never reported back by DigitalOcean — it is provisioning input, not ongoing configuration.

## Connection strings: public vs private

`connection_uri`/`host` traverse the public internet (TLS-required); `private_uri`/`private_host` resolve only inside the cluster's VPC. Applications in the same VPC should always use the private pair — lower latency and no public exposure. The default user's password is in both URIs; treat them as secrets.

## Eviction policy semantics

`evictionPolicy` values mirror Redis maxmemory policies with underscores (`allkeys_lru`, `volatile_ttl`, ...). Removing the field from a cluster that had one resets the policy to `noeviction` — it does not "keep the last value".

## What is deliberately NOT here

Users, logical databases, connection pools, read replicas, firewall (trusted-sources) rules, per-engine config parameters, Kafka topics, and log sinks are separate DigitalOcean resources with independent lifecycles. Manage them as their own resources rather than expecting cluster fields.
