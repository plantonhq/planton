# Standard KV Namespace

Creates a Workers KV namespace for key-value storage. KV provides globally replicated, eventually consistent storage for Workers. The namespace is created under the given Cloudflare account; the remaining fields are the name, ttl, and description.

## When to Use

- Caching static assets or API responses at the edge
- Storing user sessions, feature flags, or config
- Key-value data accessed from Cloudflare Workers

## Key Configuration Choices

- **ttlSeconds** (`ttlSeconds: 0`) -- Default TTL for keys; 0 = never expire. Minimum 60 for expiring keys.
- **description** (`description`) -- Optional; helps document the namespace's purpose.
- **namespaceName** (`namespaceName`) -- Unique name within account; max 64 chars.
- **accountId** (`accountId`) -- The Cloudflare account that owns the namespace.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<namespace-name>` | Unique name for the KV namespace | Choose a descriptive name (e.g., app-cache, config-store) |
| `<cloudflare-account-id>` | The Cloudflare account ID (32 hex characters) | Cloudflare dashboard -> account home (right sidebar), or `wrangler whoami` |
| `<short-description>` | Brief description of the namespace | e.g., "API response cache for production" |
