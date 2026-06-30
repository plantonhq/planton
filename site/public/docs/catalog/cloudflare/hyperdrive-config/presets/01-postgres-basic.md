---
title: "Preset: Basic PostgreSQL Hyperdrive"
description: "A Hyperdrive config that pools and caches connections to a regional PostgreSQL database, ready to bind to a Worker."
type: "preset"
rank: "01"
presetSlug: "01-postgres-basic"
componentSlug: "hyperdrive-config"
componentTitle: "Hyperdrive Config"
provider: "cloudflare"
icon: "package"
order: 1
---

# Preset: Basic PostgreSQL Hyperdrive

A Hyperdrive config that pools and caches connections to a regional PostgreSQL
database, ready to bind to a Worker.

## When to use

- A Worker needs to query a PostgreSQL database with low latency from anywhere.
- You want connection pooling (so the Worker doesn't exhaust the database's
  connection slots) plus edge caching of read queries.

## Key choices

- `origin.host` / `origin.port`: the reachable address of the database. Must be
  reachable from Cloudflare's network (a public address or one fronted by
  Cloudflare Access — see the mTLS / Access preset).
- `origin.password`: provide as a managed-secret reference; it is resolved
  just-in-time at deploy and never stored in plaintext.
- `caching`: leave enabled with the defaults (60s max age, 15s
  stale-while-revalidate) for read-heavy workloads; set `disabled: true` for
  strongly write-consistent workloads.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
| `<database-name>` | Name of the origin database |
| `<database-user>` | Database user Hyperdrive authenticates as |
| `<database-host>` | Hostname/IP of the origin database |
| `<database-password>` | Password for the database user (managed secret) |

## Binding it to a Worker

Reference this config from a `CloudflareWorker` `hyperdrive_configs` binding:

```yaml
hyperdriveConfigs:
  - name: DB
    configId:
      valueFrom:
        kind: CloudflareHyperdriveConfig
        name: app-prod-pg
        fieldPath: status.outputs.hyperdrive_id
```
