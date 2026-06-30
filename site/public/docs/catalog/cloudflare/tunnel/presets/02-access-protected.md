---
title: "Preset: Access-protected admin hostname"
description: "Publish an internal admin UI through the tunnel and require Cloudflare Access on every request, wired to an Access application you manage."
type: "preset"
rank: "02"
presetSlug: "02-access-protected"
componentSlug: "tunnel"
componentTitle: "Tunnel"
provider: "cloudflare"
icon: "package"
order: 2
---

# Preset: Access-protected admin hostname

Publish an internal admin UI through the tunnel and require Cloudflare Access on every
request, wired to an Access application you manage.

## When to use

- The service must only be reachable by authenticated, authorized users (an admin panel,
  internal dashboard, staging environment).

## Key choices

- `originRequest.access.audTag`: reference the `CloudflareZeroTrustAccessApplication`'s
  `aud` output so the app and the tunnel ingress it protects deploy as a connected graph.
- `originRequest.access.required: true`: deny any request that has not satisfied Access.
- `originRequest.access.teamName`: your Zero Trust organization (team) name.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
| `<zero-trust-org>` | Your Zero Trust organization (team) name |
| `<access-app-name>` | Name of the CloudflareZeroTrustAccessApplication protecting this hostname |
