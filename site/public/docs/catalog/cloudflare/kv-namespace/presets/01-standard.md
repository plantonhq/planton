---
title: "Standard KV Namespace"
description: "Creates a Workers KV namespace for key-value storage. KV provides globally replicated, eventually consistent storage for Workers. Only three fields: name, ttl, and description."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "kv-namespace"
componentTitle: "KV Namespace"
provider: "cloudflare"
icon: "package"
order: 1
---

# Standard KV Namespace

Creates a Workers KV namespace for key-value storage. KV provides globally replicated, eventually consistent storage for Workers. Only three fields: name, ttl, and description.

## When to Use

- Caching static assets or API responses at the edge
- Storing user sessions, feature flags, or config
- Key-value data accessed from Cloudflare Workers

## Key Configuration Choices

- **ttlSeconds** (`ttlSeconds: 0`) -- Default TTL for keys; 0 = never expire. Minimum 60 for expiring keys.
- **description** (`description`) -- Optional; helps document the namespace's purpose.
- **namespaceName** (`namespaceName`) -- Unique name within account; max 64 chars.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<namespace-name>` | Unique name for the KV namespace | Choose a descriptive name (e.g., app-cache, config-store) |
| `<short-description>` | Brief description of the namespace | e.g., "API response cache for production" |
