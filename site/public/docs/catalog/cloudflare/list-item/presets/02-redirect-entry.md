---
title: "Preset: Bulk Redirect Entry"
description: "Add a single source→target redirect to a `redirect`-kind `CloudflareList`. A redirect ruleset resolves the list with `from_list`."
type: "preset"
rank: "02"
presetSlug: "02-redirect-entry"
componentSlug: "list-item"
componentTitle: "List Item"
provider: "cloudflare"
icon: "package"
order: 2
---

# Preset: Bulk Redirect Entry

Add a single source→target redirect to a `redirect`-kind `CloudflareList`. A
redirect ruleset resolves the list with `from_list`.

## When to use

- Adding one URL redirect to a Bulk Redirect list managed as data.

## Key choices

- `redirect.sourceUrl` / `targetUrl`: the match and destination URLs.
- `redirect.statusCode`: 301 (default), 302, 307, or 308.
- `redirect.preserveQueryString` / `includeSubdomains` / `preservePathSuffix` /
  `subpathMatching`: matching/behavior flags (default false).

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
| `<redirect-list-name>` | Name of the redirect-kind CloudflareList |
| `<source-url>` | URL to match (e.g. `example.com/old`) |
| `<target-url>` | Destination URL |
