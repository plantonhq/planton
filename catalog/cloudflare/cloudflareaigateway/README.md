# Cloudflare AI Gateway

## Overview

`CloudflareAiGateway` manages one AI Gateway: the control plane Cloudflare puts in front of AI model traffic. Requests to model providers route through the gateway's endpoint, which applies caching, rate limiting, retries, logging, guardrails, data-loss prevention, spend limits, and dynamic request routing -- all configured here.

The gateway's `gateway_id` is its URL slug and is create-only: renaming replaces the gateway and changes the endpoint URL every client calls. Dynamic routes are managed as part of this resource; a route's elements graph is create-only at the provider, so editing a graph recreates that route object (never the gateway itself).

## Key Features

- **Response caching** -- serve repeated prompts from cache (`cache_ttl`), with invalidation on config updates
- **Rate limiting** -- request budgets per fixed or sliding window
- **Retries** -- constant/linear/exponential backoff toward the model provider
- **Guardrails** -- per-hazard-category FLAG/BLOCK controls on prompts and responses
- **DLP** -- screen prompts and responses against the account's DLP profiles
- **Spend limits** -- cost budgets over windows, scoped by model, provider, or request metadata
- **Dynamic routing** -- named routing graphs (conditional, percentage, rate, model nodes) requests address by name
- **BYO keys** -- provider API keys read from the account Secrets Store (`store_id`)
- **Telemetry** -- OpenTelemetry export and Logpush integration

## Use Cases

**Ideal for:**

- Putting cost caps and guardrails in front of every LLM call an organization makes
- Routing free-tier traffic to cheap models and paid traffic to smart ones (dynamic routes)
- Caching repeated prompts to cut provider bills

**Not ideal for:**

- Running the models themselves -- the gateway fronts providers (Workers AI, OpenAI, Anthropic, and kin)
- Storing the provider keys -- that is `CloudflareSecretsStore` / `CloudflareSecretsStoreSecret`

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | The Cloudflare account (32-hex). |
| `gateway_id` | string | Yes | The user-chosen URL slug. Create-only: renaming replaces the gateway and changes the endpoint URL. |
| `cache_invalidate_on_update` | bool | Yes | Whether a config update invalidates cached responses. |
| `cache_ttl` | int | Yes | Cache TTL in seconds (0 disables caching). |
| `collect_logs` | bool | Yes | Whether request/response logs are collected. |
| `rate_limiting_interval` | int | Yes | Rate-limit window in seconds (0 disables). |
| `rate_limiting_limit` | int | Yes | Requests allowed per window (0 disables). |

### Optional Fields (selected)

| Field | Type | Description |
|-------|------|-------------|
| `rate_limiting_technique` | string | `fixed` or `sliding`. |
| `retry` | object | `backoff` (constant/linear/exponential), `delay` (<= 5000ms), `max_attempts` (1-5). |
| `log_management` | object | `max_records` (10k-10M) + `strategy` (STOP_INSERTING / DELETE_OLDEST). |
| `authentication` | bool | Require the cf-aig-authorization token on every request. |
| `logpush` / `logpush_public_key` | bool / string | Push logs to the account's Logpush destination, optionally encrypted. |
| `zdr` | bool | Zero Data Retention -- never store request/response bodies. |
| `workers_ai_billing_mode` | string | Only `postpaid` today. |
| `store_id` | StringValueOrRef | The Secrets Store holding BYO provider keys. |
| `dlp` | object | DLP screening: default action, profiles, per-policy rows. |
| `guardrails` | object | Per-hazard-code FLAG/BLOCK controls for prompt and response sides. |
| `otel[]` | list | OpenTelemetry export destinations (authorization is sensitive in this spec). |
| `stripe` | object | Stripe usage-based billing (authorization is sensitive in this spec). |
| `spend_limits` | object | Budget rules -- every rule needs its own unique `id` (spec-enforced). |
| `dynamic_routes[]` | list | Named routing graphs; a graph edit recreates that route object. |

### Outputs

| Field | Description |
|-------|-------------|
| `gateway_id` | The gateway's URL slug |
| `dynamic_route_ids` | Each managed route's id, keyed by route name |

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareAiGateway
metadata:
  name: prod-llm-gateway
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  gateway_id: prod-llm-gateway
  cache_invalidate_on_update: true
  cache_ttl: 300
  collect_logs: true
  rate_limiting_interval: 60
  rate_limiting_limit: 1000
  authentication: true
  spend_limits:
    enabled: true
    rules:
      - id: daily-cap
        limit: 50
        limit_type: cost
        window: 86400
```

## Destroy Semantics

Destroy is a real delete of the gateway and its routes. Clients calling the endpoint URL start failing immediately.

## Related Resources

- **CloudflareSecretsStore** / **CloudflareSecretsStoreSecret** -- the vault behind BYO provider keys
- **CloudflareWorker** -- Workers calling models through the gateway

## Further Reading

For operational judgment -- the slug-is-the-URL trap, route replacement semantics, the spend-limit id rule -- see GUIDE.md.

## References

- [Cloudflare AI Gateway](https://developers.cloudflare.com/ai-gateway/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
