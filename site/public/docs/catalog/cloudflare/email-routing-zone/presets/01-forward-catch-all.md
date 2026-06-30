---
title: "Preset: Forward-All Email Routing"
description: "Enable Email Routing on a zone and forward all otherwise-unmatched mail to a verified destination address. The simplest \"catch everything → my inbox\" setup."
type: "preset"
rank: "01"
presetSlug: "01-forward-catch-all"
componentSlug: "email-routing-zone"
componentTitle: "Email Routing (Zone)"
provider: "cloudflare"
icon: "package"
order: 1
---

# Preset: Forward-All Email Routing

Enable Email Routing on a zone and forward all otherwise-unmatched mail to a
verified destination address. The simplest "catch everything → my inbox" setup.

## When to use

- A domain that should funnel all inbound mail to one (or a few) real mailboxes.

## Key choices

- `catchAll.type: forward` with `forwardTo` listing verified destination
  addresses (each must be a verified `CloudflareEmailRoutingAddress`).
- Add `CloudflareEmailRoutingRule`s for per-address routing that should take
  precedence over the catch-all.

## Placeholders

| Placeholder | Description |
|---|---|
| `<zone-name>` | Name of the CloudflareDnsZone |
| `<destination-email>` | A verified destination email address |
