# CloudflareCustomHostnameFallbackOrigin — Research & Design Notes

## Purpose

The fallback origin (`cloudflare_custom_hostname_fallback_origin`) is the default
backend for a Cloudflare for SaaS zone: every custom hostname in the zone routes here
unless it sets its own `custom_origin_server`. It is required before custom hostnames
can serve traffic.

## Why this is its own kind

The fallback origin is a zone-level singleton — there is exactly one per zone, and
the zone id is the resource id. It is shared by every custom hostname in the zone, so
it cannot be folded into any single CloudflareCustomHostname. Folding it into
CloudflareDnsZone would couple general DNS-zone management to a SaaS-specific feature
and bloat that spec. A small dedicated kind that CloudflareCustomHostname declares a
`depends_on` relationship to is the clean shape.

## Composition

`zone_id` references CloudflareDnsZone. `origin` is a backend endpoint modeled as a
StringValueOrRef (no default_kind, like a load balancer pool's origin address) so it
can reference another resource's output, though it must resolve to a record within
the SaaS zone.

## Engine parity

Both engines create the same resource and emit the same outputs (`status`,
`created_at`, `updated_at`, `errors`). No `PARITY-EXCEPTION` is required at
pulumi-cloudflare v6.17.0 / provider v5.

## Gotchas

- One fallback origin per zone; the `origin` must be a hostname within the SaaS zone.
- Status transitions through `pending_deployment` before `active`.
