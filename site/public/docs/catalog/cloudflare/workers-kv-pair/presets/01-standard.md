---
title: "Preset: Standard KV Entry"
description: "Seed a single configuration key into a Workers KV namespace, wired to the namespace by reference so it composes in an infra chart."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "workers-kv-pair"
componentTitle: "Workers KV Pair"
provider: "cloudflare"
icon: "package"
order: 1
---

# Preset: Standard KV Entry

Seed a single configuration key into a Workers KV namespace, wired to the
namespace by reference so it composes in an infra chart.

## When to use

- Seeding configuration or feature-flag keys that should be versioned in
  infrastructure (rather than written by the application at runtime).
- Pinning a value derived from another resource's output into KV.

## Key choices

- `namespaceId`: reference a `CloudflareKvNamespace` by name so the entry is
  created after the namespace and the dependency is explicit. A literal namespace
  ID also works.
- `value`: KV is general-purpose storage and values are not treated as secrets.
  Keep credentials out of KV — use a Worker `secret_text` binding or Cloudflare
  Secrets Store for those.
- `metadata`: optional JSON returned alongside the value on read.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
| `<kv-namespace-name>` | Name of the CloudflareKvNamespace to write into |
| `<key-name>` | The entry key (≤512 bytes) |
| `<value>` | The value to store (≤25 MiB) |
