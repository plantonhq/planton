---
title: "Preset: Drop Catch-All Email Routing"
description: "Enable Email Routing and drop all mail that no explicit rule matched. Use with per-address `CloudflareEmailRoutingRule`s that forward only the addresses you care about; everything else is silently..."
type: "preset"
rank: "02"
presetSlug: "02-drop-catch-all"
componentSlug: "email-routing-zone"
componentTitle: "Email Routing (Zone)"
provider: "cloudflare"
icon: "package"
order: 2
---

# Preset: Drop Catch-All Email Routing

Enable Email Routing and drop all mail that no explicit rule matched. Use with
per-address `CloudflareEmailRoutingRule`s that forward only the addresses you
care about; everything else is silently discarded.

## When to use

- A domain that should accept mail only for a known set of addresses and discard
  the rest (reduces spam/backscatter).

## Key choices

- `catchAll.type: drop` — unmatched mail is discarded.
- Add `CloudflareEmailRoutingRule`s for each address you want delivered.

## Placeholders

| Placeholder | Description |
|---|---|
| `<zone-name>` | Name of the CloudflareDnsZone |
