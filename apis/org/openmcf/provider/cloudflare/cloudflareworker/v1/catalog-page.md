# Cloudflare Worker

Deploy a Cloudflare Worker — serverless code that runs at the edge — together
with its resource bindings, routing, schedules, and runtime settings.

## What Gets Created

- A `cloudflare_workers_script` (the Worker), with all of its bindings and —
  when set — its uploaded static assets (Workers Static Assets).
- Optionally: a `cloudflare_workers_script_subdomain` (workers.dev), one
  `cloudflare_workers_custom_domain` per custom domain, one
  `cloudflare_workers_route` per route, and a `cloudflare_workers_cron_trigger`
  when schedules are set.

## Prerequisites

- A Cloudflare account ID.
- Code, assets, or both: a script source (inline `content`, or a pre-built bundle
  in an R2 bucket via `r2Bundle`) and/or a static-asset `directory` (`assets`).
  For the R2 path, R2 S3 credentials are supplied via the provider config.

## Configuration Reference

**Required**

- `accountId` — Cloudflare account ID.
- `workerName` — the Worker script name.
- Code, assets, or both: at most one script source — `content` (inline) or
  `r2Bundle` (`{bucket, path}`) — and/or `assets` (a built site directory).

**Optional**

- `assets` — Workers Static Assets: `directory` (the built site to upload),
  optional `bindingName` (expose as `env.<NAME>`), and `config`
  (`htmlHandling`, `notFoundHandling`, `headers`, `redirects`, and either
  `runWorkerFirst` or `runWorkerFirstRules`). With assets and no script source
  the Worker is a pure static site; with both it is a full-stack app.
- `compatibilityDate`, `compatibilityFlags`, `mainModule`.
- Bindings (grouped by type): `vars`, `secrets`, `kvNamespaces`, `r2Buckets`,
  `d1Databases`, `hyperdriveConfigs`, `services`, `queues`, `durableObjects`,
  `analyticsEngineDatasets`, `vectorizeIndexes`, `ai`, `versionMetadata`.
- Routing: `workersDev`, `customDomains`, `routes`.
- `schedules` (cron), `observability`, `placement`, `limits`, `logpush`,
  `tailConsumers`.

Cross-resource bindings (`kvNamespaces`, `r2Buckets`, `d1Databases`,
`hyperdriveConfigs`, `services`) accept either a literal id or a `valueFrom`
reference to the producing resource.

## Stack Outputs

| Output | Description |
|---|---|
| `script_id` | The deployed Worker script ID |
| `script_name` | The Worker script name (target of a service binding) |
| `custom_domain_hostnames` | Custom-domain hostnames attached to the Worker |
| `route_patterns` | Route patterns mapped to the Worker |

## Related Components

- CloudflareKvNamespace, CloudflareWorkersKvPair
- CloudflareD1Database
- CloudflareR2Bucket
- CloudflareHyperdriveConfig
- CloudflareDnsZone
