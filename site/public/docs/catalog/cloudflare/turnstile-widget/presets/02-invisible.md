---
title: "Preset: Invisible Widget"
description: "An `invisible` Turnstile widget that runs without any visible challenge UI for most users — ideal when you want frictionless protection on high-traffic flows."
type: "preset"
rank: "02"
presetSlug: "02-invisible"
componentSlug: "turnstile-widget"
componentTitle: "Turnstile Widget"
provider: "cloudflare"
icon: "package"
order: 2
---

# Preset: Invisible Widget

An `invisible` Turnstile widget that runs without any visible challenge UI for
most users — ideal when you want frictionless protection on high-traffic flows.

## When to use

- Background bot protection where a visible challenge would hurt conversion.

## Key choices

- `mode: invisible` — no visible widget unless a challenge is strictly needed.
- Still verify the token server-side with the `secret` output.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
| `<widget-name>` | Human-readable widget name |
| `<domain>` | A domain the widget runs on |
