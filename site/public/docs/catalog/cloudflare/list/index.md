---
title: "List"
description: "List deployment documentation"
icon: "package"
order: 100
componentName: "cloudflarelist"
---

# Cloudflare List

Provision an account-scoped Cloudflare List — a reusable, named collection
referenced from rule expressions.

## What Gets Created

- A `cloudflare_list` (the container) of kind `ip`, `redirect`, `hostname`, or
  `asn`. Entries are added separately via `CloudflareListItem`.

## Prerequisites

- A Cloudflare account ID.

## Configuration Reference

**Required**

- `accountId` — Cloudflare account ID.
- `kind` — `ip`, `redirect`, `hostname`, or `asn` (immutable).
- `name` — name used in rule expressions (immutable).

**Optional**

- `description` — human-readable summary.

## Stack Outputs

| Output | Description |
|---|---|
| `list_id` | The list ID (referenced by list items) |
| `name` | The list name |
| `kind` | The list kind |

## Related Components

- CloudflareListItem
- CloudflareRuleset
