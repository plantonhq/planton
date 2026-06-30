---
title: "Preset: Managed Widget"
description: "The recommended default: a `managed` Turnstile widget where Cloudflare chooses the challenge dynamically. Best general-purpose protection for forms."
type: "preset"
rank: "01"
presetSlug: "01-managed"
componentSlug: "turnstile-widget"
componentTitle: "Turnstile Widget"
provider: "cloudflare"
icon: "package"
order: 1
---

# Preset: Managed Widget

The recommended default: a `managed` Turnstile widget where Cloudflare chooses the
challenge dynamically. Best general-purpose protection for forms.

## When to use

- Default choice for protecting a login/signup/contact form.

## Key choices

- `mode: managed` — Cloudflare decides whether to show an interactive challenge.
- `domains` — list every domain (and `localhost` for local development).
- The `secret` output wires into a Worker that calls `/siteverify`.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
| `<widget-name>` | Human-readable widget name |
| `<domain>` | A domain the widget runs on |
