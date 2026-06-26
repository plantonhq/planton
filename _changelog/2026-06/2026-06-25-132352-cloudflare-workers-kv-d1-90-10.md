# Cloudflare Workers / KV / D1 family to 90/10 on provider v5

**Date**: June 25, 2026
**Type**: Enhancement
**Components**: API Definitions, Provider Framework, IAC (Terraform + Pulumi), Resource Management

## Summary

Raised the Cloudflare Workers/KV/D1 family to deep ("90/10") coverage on provider
v5 and forged two new composable kinds. `CloudflareWorker` was rewritten around
grouped, foreign-keyed binding lists (the wrangler.toml grain) plus folded
routing, cron, and runtime settings; `CloudflareKvNamespace` and
`CloudflareD1Database` were enriched and corrected; and `CloudflareHyperdriveConfig`
and `CloudflareWorkersKvPair` were added as first-class kinds. Both engines move
together, and the work was validated to a live `tofu apply`/`destroy` against a
real Cloudflare account.

## What's New

### CloudflareWorker — rewritten (kind 1803)

- **Script source** is now a oneof: inline `content` or an R2 `r2_bundle`
  artifact (the old R2-only bundle model).
- **Bindings are grouped, typed lists** — `vars`, `secrets`, `kv_namespaces`,
  `r2_buckets`, `d1_databases`, `hyperdrive_configs`, `services`, `queues`,
  `durable_objects`, `analytics_engine_datasets`, `vectorize_indexes`, `ai`,
  `version_metadata` — each cross-resource binding a `StringValueOrRef` to the
  producing kind (KV/R2/D1/Hyperdrive/Worker). The IaC layer flattens these into
  the provider's single discriminated `bindings` array.
- **Routing folds onto the worker**: `workers_dev` (workers.dev subdomain),
  `custom_domains` (managed TLS hostnames), and `routes` (zone patterns),
  replacing the old dummy-AAAA-record + route hack.
- **Cron, observability, placement, limits, logpush, tail_consumers** folded on.
- Removed the dead `env{}` model and the disabled Secrets-API uploader; secrets
  are now `secret_text` bindings (`StringValueOrRef + (sensitive)`, JIT-resolved).
- Outputs: `script_id`, `script_name`, `custom_domain_hostnames`, `route_patterns`.

### CloudflareHyperdriveConfig — forged (kind 1810, `cfhyp`)

Account-scoped connection pooler + edge cache for a regional SQL database. Models
`origin` (database/scheme/user/host/port, with `password` and
`access_client_secret` as `StringValueOrRef + (sensitive)`), `caching`, `mtls`
(`sslmode` a validated string), and `origin_connection_limit` (≥5). A Worker binds
it via `hyperdrive_configs`.

### CloudflareWorkersKvPair — forged (kind 1809, `cfkvp`)

A single key-value entry in a KV namespace, as a first-class composable resource —
distinct from runtime application data. `namespace_id` is a `StringValueOrRef` to
`CloudflareKvNamespace`. Lets infrastructure seed configuration keys (including
values derived from other resources' outputs).

### CloudflareKvNamespace / CloudflareD1Database — enriched

- KV: dropped the dead `ttl_seconds`/`description` (not on the v5 resource);
  added the `supports_url_encoding` output.
- D1: `read_replication.mode` is now an enum (`auto`/`disabled`); dropped the
  phantom `connection_string` output; tightened `account_id` to 32-hex.

## Engine parity: SDK upgraded, full parity (no deferrals, no proto `reserved`)

An initial pass deferred four v5 fields because the then-pinned Pulumi Cloudflare
SDK (v6.10.1) did not expose them, and marked them with proto `reserved`. That was
the wrong instinct: the proto is the future-proof source of truth and must not be
held back by an engine SDK lag, and this codebase does not use `reserved` for
backward-compatibility. Corrected by **upgrading `pulumi-cloudflare/sdk/v6` to
v6.17.0** (transitively `pulumi/sdk/v3` → v3.242.0; `go 1.26` already satisfies the
toolchain), which exposes all of them. Now modeled in the proto and honored by
**both** engines at full parity:

- D1 `jurisdiction` (eu/fedramp, mutually exclusive with region)
- worker service-binding `entrypoint`
- worker `limits.subrequests`
- worker custom-domain `zone_id` (FK → CloudflareDnsZone)
- DNS-record `private_routing` (the same upgrade unlocked this slice-2 deferral too)

Every `reserved` marker added during the deferral was removed and the affected
specs renumbered contiguously, leaving zero engine-driven deferrals and zero
`reserved` cruft in the Cloudflare provider. The durable principle is captured in
the project's `coding-guidelines/0004-engine-parity-and-no-proto-reserved.md`.

## Validation

`make protos` (incl. the Java compile gate) green; `go build ./...`; all five
component spec tests pass; `pkg/outputs` conformance extended and green (the parity
guard); `pkg/secretcoverage` gate green (new sensitive fields covered, stale
`CloudflareWorker:spec.env.secrets` baseline entry removed); `tofu validate` on all
five modules against the real v5 provider; all five Pulumi entrypoints build the
release way. **Live `tofu apply` + `destroy`** against a real Cloudflare account
created and tore down a KV namespace + KV pair, a D1 database, and a Worker binding
both (workers.dev enabled) with zero leftover resources. After the SDK upgrade and
field restoration, re-validated: `make protos`, the affected spec tests, conformance
and secret-coverage gates, `tofu validate` of the worker/d1/dns-record modules with
the restored attributes, and a live `tofu apply`/`destroy` of a D1 database with
`jurisdiction: eu` (a restored field, confirmed provisioning on v6.17.0).

## Known follow-ups

- A live `tofu apply` for `CloudflareHyperdriveConfig` needs a reachable origin
  database (Hyperdrive verifies connectivity at create); the module is
  `tofu validate`-clean and plan-ready.
- The deep `docs/README.md` research essays for the worker/KV/D1 components retain
  some pre-rewrite field names; the user-facing `README.md` and `catalog-page.md`
  are current.
