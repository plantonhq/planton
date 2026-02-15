---
title: "Minimal Worker"
description: "Bare minimum Cloudflare Worker with only the script bundle. No KV bindings, DNS routes, or env vars. Use when deploying a Worker that will be attached to routes or configured elsewhere (e.g., via..."
type: "preset"
rank: "02"
presetSlug: "02-minimal"
componentSlug: "worker"
componentTitle: "Worker"
provider: "cloudflare"
icon: "package"
order: 2
---

# Minimal Worker

Bare minimum Cloudflare Worker with only the script bundle. No KV bindings, DNS routes, or env vars. Use when deploying a Worker that will be attached to routes or configured elsewhere (e.g., via Wrangler or dashboard).

## When to Use

- Initial Worker deployment; add DNS/routes later
- Workers invoked by Cron Triggers or Queues only
- Simplest possible Worker manifest

## Key Configuration Choices

- **scriptBundle only** (`scriptBundle`) -- R2 bucket and path to the pre-built Worker bundle.
- **No dns** (`dns` omitted) -- Worker runs but has no route; attach via dashboard or separate config.
- **No kvBindings** (`kvBindings` omitted) -- Add when Worker needs KV storage.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cloudflare-account-id>` | Cloudflare account ID | Dashboard → Overview → Account ID |
| `<worker-name>` | Worker name | Descriptive name (e.g., hello-world) |
| `<r2-bucket-name>` | R2 bucket containing the script bundle | CloudflareR2Bucket or upload target |
| `<script-bundle-path>` | Path to bundle in R2 | Your build output (e.g., dist/worker.js) |

## Related Presets

- **01-api-with-custom-domain** -- Use when you need KV bindings, DNS, and env vars
