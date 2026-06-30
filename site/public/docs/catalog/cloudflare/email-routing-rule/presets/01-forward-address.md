---
title: "Preset: Forward an Address"
description: "Forward mail sent to a specific recipient (e.g. `support@`) to one or more verified destination mailboxes."
type: "preset"
rank: "01"
presetSlug: "01-forward-address"
componentSlug: "email-routing-rule"
componentTitle: "Email Routing Rule"
provider: "cloudflare"
icon: "package"
order: 1
---

# Preset: Forward an Address

Forward mail sent to a specific recipient (e.g. `support@`) to one or more
verified destination mailboxes.

## When to use

- Routing role addresses to real inboxes.

## Key choices

- `matchers`: a `literal` matcher on `field: to` with the matched address.
- `action.type: forward` with `forwardTo` destination addresses (each must be a
  verified `CloudflareEmailRoutingAddress`).

## Placeholders

| Placeholder | Description |
|---|---|
| `<zone-name>` | Name of the CloudflareDnsZone |
| `<rule-name>` | Descriptive rule name |
| `<matched-address>` | The recipient address to match (e.g. support@example.com) |
| `<destination-email>` | The verified destination mailbox |
