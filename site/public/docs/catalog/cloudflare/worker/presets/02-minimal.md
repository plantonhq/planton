---
title: "Preset: Minimal Worker"
description: "The smallest deployable Worker: an inline script exposed on a workers.dev subdomain. No external bindings or custom domains."
type: "preset"
rank: "02"
presetSlug: "02-minimal"
componentSlug: "worker"
componentTitle: "Worker"
provider: "cloudflare"
icon: "package"
order: 2
---

# Preset: Minimal Worker

The smallest deployable Worker: an inline script exposed on a workers.dev
subdomain. No external bindings or custom domains.

## When to use

- A quick edge endpoint, a webhook receiver, or a starting point you will grow.
- Trying out the platform without first uploading a build artifact to R2.

## Key choices

- `content`: the inline ES-module source. For larger or CI-built workers, use the
  `r2_bundle` source instead (see the full-featured preset).
- `workersDev.enabled`: exposes the Worker at
  `<name>.<account-subdomain>.workers.dev`.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
