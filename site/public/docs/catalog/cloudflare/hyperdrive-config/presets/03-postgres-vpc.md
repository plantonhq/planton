---
title: "Preset: PostgreSQL Hyperdrive over a Workers VPC Service"
description: "A Hyperdrive config that reaches a private PostgreSQL origin by egressing through a Workers VPC Service, rather than dialing a publicly reachable host."
type: "preset"
rank: "03"
presetSlug: "03-postgres-vpc"
componentSlug: "hyperdrive-config"
componentTitle: "Hyperdrive Config"
provider: "cloudflare"
icon: "package"
order: 3
---

# Preset: PostgreSQL Hyperdrive over a Workers VPC Service

A Hyperdrive config that reaches a private PostgreSQL origin by egressing through
a Workers VPC Service, rather than dialing a publicly reachable host.

## When to use

- The origin database is private (not internet-reachable) and is reached over a
  Cloudflare Workers VPC Service.
- You want Hyperdrive's pooling and caching in front of a VPC-private database.

## Key choices

- `origin.serviceId`: the Workers VPC Service Hyperdrive egresses through. `host`
  is still the origin database host as seen from inside the VPC.
- No `mtls` block: TLS for a VPC Service origin is managed on the VPC Service, so
  `serviceId` and `mtls` are mutually exclusive (the spec rejects setting both).
- `originConnectionLimit`: raise from the default to handle higher concurrency on
  paid plans (5–100).

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
| `<database-name>` | Name of the origin database |
| `<database-user>` | Database user Hyperdrive authenticates as |
| `<private-database-host>` | Host of the origin database as reachable within the VPC |
| `<database-password>` | Password for the database user (managed secret) |
| `<workers-vpc-service-id>` | ID of the Workers VPC Service to egress through |
