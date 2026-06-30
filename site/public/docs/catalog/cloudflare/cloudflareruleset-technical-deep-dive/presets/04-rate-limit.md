---
title: "Preset: Rate limiting rules"
description: "Throttle abusive traffic with the `http_ratelimit` phase. Each rule counts requests per a set of characteristics (buckets) over a period and applies its action when the threshold is exceeded."
type: "preset"
rank: "04"
presetSlug: "04-rate-limit"
componentSlug: "cloudflareruleset-technical-deep-dive"
componentTitle: "CloudflareRuleset — Technical Deep Dive"
provider: "cloudflare"
icon: "package"
order: 4
---

# Preset: Rate limiting rules

Throttle abusive traffic with the `http_ratelimit` phase. Each rule counts requests
per a set of characteristics (buckets) over a period and applies its action when the
threshold is exceeded.

## When to use

- Protect login/auth endpoints from credential stuffing.
- Cap per-API-key or per-IP request rates on an API.

## Key choices

- `ratelimit.characteristics`: the buckets requests are counted against (e.g.
  `ip.src`, a header value). Always include `cf.colo.id` with `ip.src` for accurate
  per-PoP counting.
- `ratelimit.period` / `requests_per_period`: the window and threshold.
- `ratelimit.mitigation_timeout`: how long the action stays in effect once triggered.
- `ratelimit.requests_to_origin`: count only requests that reach the origin (not
  cache hits).
- `action`: what to do when over the limit — `block`, `managed_challenge`, `js_challenge`, etc.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-zone-id>` | The zone the rate-limit ruleset applies to |
