# Preset: Edge API with Custom Domain

A production-shaped Worker: deployed from a CI-built bundle in R2, wired to KV and
D1 by reference, exposed on a managed custom domain, with observability on.

## When to use

- A real edge API or backend with configuration (KV), data (D1), and secrets.
- A CI/CD flow that builds the worker bundle and uploads it to an R2 bucket; the
  deploy then references that artifact.

## Key choices

- `r2Bundle`: points at the pre-built bundle object in R2. For small or inline
  scripts use `content` instead.
- `kvNamespaces` / `d1Databases`: bind other resources by reference so the graph
  expresses the dependency and ids resolve from upstream outputs.
- `secrets`: provide each value as a managed-secret reference; resolved
  just-in-time at deploy.
- `customDomains`: Cloudflare provisions and manages TLS and infers the zone from
  the hostname.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
| `<build-artifacts-bucket>` | R2 bucket holding the built worker bundle |
| `<path/to/worker-bundle.js>` | Object key of the bundle in that bucket |
| `<managed-secret-reference>` | Managed-secret reference for API_KEY |
| `<kv-namespace-name>` | CloudflareKvNamespace to bind as CONFIG |
| `<d1-database-name>` | CloudflareD1Database to bind as DB |
| `<api.example.com>` | Custom hostname to route to the Worker |
