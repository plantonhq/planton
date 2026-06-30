---
title: "Preset: Destination Address"
description: "Register a destination mailbox that Email Routing rules and catch-alls can forward to. A verification email is sent to the address on creation."
type: "preset"
rank: "01"
presetSlug: "01-destination"
componentSlug: "email-routing-address"
componentTitle: "Email Routing Address"
provider: "cloudflare"
icon: "package"
order: 1
---

# Preset: Destination Address

Register a destination mailbox that Email Routing rules and catch-alls can forward
to. A verification email is sent to the address on creation.

## When to use

- Adding the real inbox a domain's mail should be forwarded to.

## Key choices

- `email` — the destination mailbox. The owner must click the verification link
  before the address can receive forwarded mail.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
| `<destination-email>` | The destination mailbox address |
