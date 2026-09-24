# DigitalOcean Database User -- Operational Guide

What experience with this component teaches that the field reference cannot.

## One user per service, always

The cluster's built-in `doadmin` user works everywhere, which is exactly why production should not use it: one leaked credential exposes everything, and rotation breaks every consumer at once. Create one user per service; each gets its own server-generated password that rotates (by replacing the user) and revokes (by deleting it) independently.

## Passwords rotate by replacement

DigitalOcean generates passwords; there is no "set password" surface here. To rotate a credential, create a replacement user (new name), move the service to it, then delete the old user. Deleting a user revokes its access immediately.

## ACLs: declare everything, trust the manifest

Both provisioners record Kafka/OpenSearch ACLs only from the create response and never refresh them from the API afterward, so Planton cannot show you the live ACL state. Treat the manifest as the single source of truth and review permission changes in code review, not in the console. ACL edits apply in place (no replacement).

## Set `settings` by engine -- `{}` on PostgreSQL, absent on MySQL

What DigitalOcean stores for `settings` depends on the cluster's engine, and the provisioners cannot tell the engine from a cluster UUID -- so the manifest carries that knowledge (measured 2026-09-17 under the idempotency gate). Every PostgreSQL user comes back from the create with a settings object (`pg_allow_replication: false`), which the provisioners store as one empty settings block and never refresh; a PostgreSQL manifest without `settings` therefore proposes removing that block on its first re-plan -- a one-time server-side no-op, but a change your review has to explain away. Declare `settings: {}` on PostgreSQL users and the configuration mirrors what is stored from the first apply. A MySQL user never carries a settings object, and the API refuses a settings update on a MySQL cluster (`422 operation is not supported for this cluster type`), so a MySQL manifest that declared even `settings: {}` would fail every apply after the first: leave it out. Kafka and OpenSearch users declare their ACLs, which is the block's real purpose. The phantom diff and the never-refreshed read-back are reported upstream as [digitalocean/terraform-provider-digitalocean#1610](https://github.com/digitalocean/terraform-provider-digitalocean/issues/1610).

## MySQL auth plugin: leave it unset

Unset means DigitalOcean's `caching_sha2_password` -- the modern plugin. Set `mysql_native_password` only for clients too old to speak it, and treat that as a dated compatibility decision to revisit. Clearing the field later resets the user to the modern default in place.

## MongoDB quirk

MongoDB clusters return the user's password ONLY in the create response. The outputs still carry it (captured at create), but an imported MongoDB user has no recoverable password -- rotate by replacement instead.

## Serialized operations on busy clusters

User creates and deletes serialize per cluster (an API constraint the provider enforces with a lock). A chart creating ten users on one cluster deploys them one at a time -- expect linear time, not a hang.

## What is deliberately NOT here

In-database GRANTs for PostgreSQL/MySQL (data-plane SQL, not IaC), password values in specs (server-generated only), and per-user settings MongoDB exposes in the raw API but the provider does not bridge.
