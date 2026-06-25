# CloudflareWorker

Deploy a Cloudflare Worker — JavaScript/TypeScript (or Wasm/Python) that runs on
Cloudflare's edge in V8 isolates — along with everything it needs to be useful:
resource bindings, routing, scheduled invocations, and runtime settings.

## Script source

A Worker carries executable code, static assets, or both. Provide at most one
script source (the two are mutually exclusive), optionally alongside `assets`:

- `content` — inline ES-module source. Best for small or generated scripts and
  for quick iteration.
- `r2Bundle` — `{bucket, path}` pointing at a pre-built bundle stored in an R2
  bucket. Best for CI/CD: build the bundle, upload it to R2, and reference it.

```yaml
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  workerName: edge-api
  content: |
    export default { async fetch(req, env) { return new Response("ok"); } }
```

## Static assets (static sites & full-stack apps)

Point `assets` at a built site directory to serve it directly from Cloudflare's
edge (Cloudflare Workers Static Assets — the converged successor to Cloudflare
Pages for the build-and-upload model). Use it three ways:

- **Pure static site / SPA** — set `assets` with no script source.
- **Full-stack app** — set `assets` *and* a script source; the script's
  Functions handle dynamic routes while everything else is served as an asset.
- **Programmatic access** — set `assets.bindingName` to expose the asset
  namespace to your script as `env.<NAME>` (e.g. `env.ASSETS.fetch(request)`).

```yaml
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  workerName: marketing-site
  assets:
    directory: dist            # your build output, uploaded at deploy
    config:
      notFoundHandling: single-page-application   # SPA fallback to /index.html
      htmlHandling: auto-trailing-slash
      headers: |
        /*
          X-Frame-Options: DENY
      redirects: |
        /old /new 301
      # runWorkerFirst: true            # run the script on every request, OR:
      # runWorkerFirstRules: ["/api/*", "!/api/docs/*"]
```

`config` is optional. `runWorkerFirst` (all paths) and `runWorkerFirstRules`
(per-path; `!`-prefixed rules take precedence) are mutually exclusive. A new
deploy with a changed `directory` ships a new version of the site — the same
desired-state flow as any other Worker change.

## Bindings (grouped by type)

Bindings expose resources to the Worker as JavaScript variables. They are grouped
by type, mirroring `wrangler.toml`, and each cross-resource binding accepts a
literal id or a `valueFrom` reference to the producing resource:

| Group | Binds | Reference kind |
|---|---|---|
| `vars` | plain-text variables (map) | — |
| `secrets` | secret values (managed-secret, JIT-resolved) | — |
| `kvNamespaces` | KV namespaces | CloudflareKvNamespace |
| `r2Buckets` | R2 buckets (+ optional jurisdiction) | CloudflareR2Bucket |
| `d1Databases` | D1 databases | CloudflareD1Database |
| `hyperdriveConfigs` | Hyperdrive configs | CloudflareHyperdriveConfig |
| `services` | other Workers (service bindings) | CloudflareWorker |
| `queues` | Queue producers (by name) | — |
| `durableObjects` | Durable Object namespaces | — |
| `analyticsEngineDatasets` | Analytics Engine datasets | — |
| `vectorizeIndexes` | Vectorize indexes | — |
| `ai` | Workers AI gateway | — |
| `versionMetadata` | deployed version metadata | — |

```yaml
  kvNamespaces:
    - name: CONFIG
      namespaceId:
        valueFrom: { kind: CloudflareKvNamespace, name: app-config, fieldPath: status.outputs.namespace_id }
  d1Databases:
    - name: DB
      databaseId:
        valueFrom: { kind: CloudflareD1Database, name: app-db, fieldPath: status.outputs.database_id }
  secrets:
    - name: API_KEY
      value: <managed-secret-reference>
```

## Routing

- `workersDev` — expose on `<name>.<account-subdomain>.workers.dev`.
- `customDomains` — managed hostnames with automatic TLS (Cloudflare infers the zone).
- `routes` — pattern-based routes within a zone (`{zoneId, pattern}`).

## Scheduling and runtime settings

- `schedules` — cron expressions invoking the Worker's scheduled handler.
- `observability` — Workers Logs (`enabled`, `headSamplingRate`).
- `placement` — Smart Placement (`mode: smart`).
- `limits` — `cpuMs` per invocation.
- `logpush`, `tailConsumers`.

## Outputs

| Output | Description |
|---|---|
| `script_id` | The deployed Worker script ID |
| `script_name` | The Worker script name (the target of a service binding) |
| `custom_domain_hostnames` | Custom-domain hostnames attached to the Worker |
| `route_patterns` | Route patterns mapped to the Worker |

## Secrets

`secrets[].value` is secret-by-default: provide a managed-secret reference,
resolved just-in-time at deploy. Plain configuration belongs in `vars`.

## Related components

- `CloudflareKvNamespace` / `CloudflareWorkersKvPair`, `CloudflareD1Database`,
  `CloudflareR2Bucket`, `CloudflareHyperdriveConfig`, `CloudflareDnsZone`.
