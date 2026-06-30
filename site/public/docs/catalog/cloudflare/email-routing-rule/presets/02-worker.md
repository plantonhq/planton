---
title: "Preset: Route to an Email Worker"
description: "Hand mail matching a recipient to an Email Worker for custom processing (parsing, webhooks, storage, auto-responses)."
type: "preset"
rank: "02"
presetSlug: "02-worker"
componentSlug: "email-routing-rule"
componentTitle: "Email Routing Rule"
provider: "cloudflare"
icon: "package"
order: 2
---

# Preset: Route to an Email Worker

Hand mail matching a recipient to an Email Worker for custom processing (parsing,
webhooks, storage, auto-responses).

## When to use

- Programmatic email handling beyond simple forwarding.

## Key choices

- `action.type: worker` with `worker` referencing a `CloudflareWorker` (an Email
  Worker).
- Combine with a `literal` matcher to scope which mail the Worker receives.

## Placeholders

| Placeholder | Description |
|---|---|
| `<zone-name>` | Name of the CloudflareDnsZone |
| `<rule-name>` | Descriptive rule name |
| `<matched-address>` | The recipient address to match |
| `<email-worker-name>` | Name of the CloudflareWorker (Email Worker) |
