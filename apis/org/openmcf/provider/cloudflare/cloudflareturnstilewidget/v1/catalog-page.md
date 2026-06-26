# Cloudflare Turnstile Widget

Provision a Cloudflare Turnstile widget — a privacy-preserving CAPTCHA
alternative — and expose its site key and (sensitive) secret key.

## What Gets Created

- A `cloudflare_turnstile_widget` with a public site key and a secret key.

## Prerequisites

- A Cloudflare account ID.

## Configuration Reference

**Required**

- `accountId` — Cloudflare account ID.
- `name` — human-readable widget name.
- `domains` — domains the widget may run on (≥1).
- `mode` — `non-interactive`, `invisible`, or `managed`.

**Optional**

- `clearanceLevel`, `botFightMode`, `ephemeralId`, `offlabel` (Enterprise),
  `region` (`world`/`china`, immutable).

## Stack Outputs

| Output | Description |
|---|---|
| `sitekey` | Public site key |
| `secret` | Secret key for `/siteverify` (sensitive) |
| `created_on` | Creation timestamp |
| `modified_on` | Last-modified timestamp |

## Related Components

- CloudflareWorker
