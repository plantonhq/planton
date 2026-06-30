---
title: "List Item"
description: "List Item deployment documentation"
icon: "package"
order: 100
componentName: "cloudflarelistitem"
---

# Cloudflare List Item

Write a single entry into a Cloudflare List, matching the parent list's kind.

## What Gets Created

- A `cloudflare_list_item` (one IP/CIDR, ASN, hostname, or redirect) inside an
  existing list.

## Prerequisites

- A Cloudflare account ID.
- An existing `CloudflareList` to write into.

## Configuration Reference

**Required**

- `accountId` — Cloudflare account ID.
- `listId` — list ID, or a reference to a `CloudflareList`.
- exactly one of `ip` / `asn` / `hostname` / `redirect`.

**Optional**

- `comment` — informative summary.

## Stack Outputs

| Output | Description |
|---|---|
| `item_id` | The list item's identifier |
| `list_id` | The list written to |

## Related Components

- CloudflareList
- CloudflareRuleset
