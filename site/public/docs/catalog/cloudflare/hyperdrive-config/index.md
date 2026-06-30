---
title: "Hyperdrive Config"
description: "Hyperdrive Config deployment documentation"
icon: "package"
order: 100
componentName: "cloudflarehyperdriveconfig"
---

# Cloudflare Hyperdrive Config

Accelerate a Worker's access to a regional SQL database with connection pooling
and edge query caching.

## What Gets Created

- A `cloudflare_hyperdrive_config` pointing at your PostgreSQL or MySQL origin,
  with optional query caching and mutual-TLS.

## Prerequisites

- A Cloudflare account ID.
- A reachable origin database (public, or fronted by Cloudflare Access / a
  Cloudflare Tunnel).
- The database user's password, stored as a managed secret.

## Configuration Reference

**Required**

- `accountId` — Cloudflare account ID.
- `name` — config name.
- `origin.database`, `origin.scheme`, `origin.user`, `origin.password`.

**Optional**

- `origin.host`, `origin.port`, `origin.accessClientId`, `origin.accessClientSecret`.
- `origin.serviceId` — Workers VPC Service to egress through (mutually exclusive with `mtls`).
- `caching.disabled`, `caching.maxAge`, `caching.staleWhileRevalidate`.
- `mtls.caCertificateId`, `mtls.mtlsCertificateId`, `mtls.sslmode`.
- `originConnectionLimit`.

## Stack Outputs

| Output | Description |
|---|---|
| `hyperdrive_id` | The Hyperdrive config ID |
| `name` | The config name |

## Related Components

- CloudflareWorker
- CloudflareD1Database
- CloudflareR2Bucket
