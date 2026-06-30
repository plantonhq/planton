---
title: "Workers KV Pair"
description: "Workers KV Pair deployment documentation"
icon: "package"
order: 100
componentName: "cloudflareworkerskvpair"
---

# Cloudflare Workers KV Pair

Write a single key-value entry into a Workers KV namespace as a managed,
composable resource.

## What Gets Created

- A `cloudflare_workers_kv` entry (key + value, optional JSON metadata) inside an
  existing KV namespace.

## Prerequisites

- A Cloudflare account ID.
- An existing Workers KV namespace (a `CloudflareKvNamespace`) to write into.

## Configuration Reference

**Required**

- `accountId` — Cloudflare account ID.
- `namespaceId` — KV namespace ID, or a reference to a CloudflareKvNamespace.
- `keyName` — the entry key (≤512 bytes).
- `value` — the value (≤25 MiB).

**Optional**

- `metadata` — arbitrary JSON returned with the value on read.

## Stack Outputs

| Output | Description |
|---|---|
| `key_name` | The entry's key name |
| `namespace_id` | The namespace ID written to |

## Related Components

- CloudflareKvNamespace
- CloudflareWorker
