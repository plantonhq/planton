# Cloudflare AI Gateway

Deploys a Cloudflare AI Gateway: the control plane in front of AI model traffic, applying caching, rate limiting, retries, logging, guardrails, data-loss prevention, spend limits, and dynamic request routing to every call that flows through its endpoint. The gateway is configuration, not compute — it is available on free plans. Clients call the gateway's endpoint URL, whose slug is the `gatewayId` you choose here, and each named dynamic route is managed alongside the gateway as part of this one resource.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **AI Gateway** — one `cloudflare_ai_gateway` carrying the endpoint slug, the five required traffic scalars (caching, log collection, rate limiting), and the optional retry, log-management, guardrails, DLP, OTel, Stripe, and spend-limit configuration
- **Dynamic Routes** — one `cloudflare_ai_gateway_dynamic_routing` per `dynamicRoutes` entry, created after the gateway and attached to it. A route's element graph is create-only at the provider: any graph edit replaces that route object (requests re-resolve by name on the next call), never the gateway itself

Destroy is a real delete of the gateway and every dynamic route — clients calling the endpoint URL fail immediately, so drain traffic first.

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module whose API token carries **Account → AI Gateway → Edit**. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Cloudflare Account

- **The account ID** (32-character hex) for `accountId` — free plans included; the gateway itself has no plan-tier gate.
- **A Secrets Store with provider keys** (only for Bring Your Own Keys) — the store referenced by `storeId` must already hold the model-provider API keys, scoped for AI Gateway use.
- **DLP profiles** (only for `dlp`) — the profile IDs you list must already exist in the account's DLP configuration.

## Deploy

### Console

Open the deployment store, find **Cloudflare AI Gateway**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the endpoint slug, the five required traffic scalars (cache, logs, rate limiting), and the optional protection surfaces — guardrails, DLP, spend limits, and dynamic routes. Start from the **Guarded production gateway** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareAiGateway
metadata:
  name: prod-llm-gateway
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  gatewayId: prod-llm-gateway
  cacheInvalidateOnUpdate: true
  cacheTtl: 300
  collectLogs: true
  rateLimitingInterval: 60
  rateLimitingLimit: 1000
```

```shell
planton apply -f ai-gateway.yaml
```

This creates a gateway that caches model responses for five minutes, collects request logs, and allows 1000 requests per minute — the endpoint URL ends in the `prod-llm-gateway` slug. An Infra Job tracks the provisioning in real time.

### InfraChart

When bringing your own provider keys, reference a Secrets Store deployed in the same InfraPipeline:

```yaml
spec:
  storeId:
    valueFrom:
      kind: CloudflareSecretsStore
      name: ai-provider-keys
      fieldPath: status.outputs.store_id
```

The InfraPipeline resolves the dependency graph, deploys the Secrets Store first, then provisions the gateway with the resolved store ID.

## Key Configuration

These are the most important decisions when configuring an AI Gateway. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The slug is the URL** — `gatewayId` appears verbatim in the endpoint every client calls and is create-only: renaming replaces the gateway and breaks every client mid-flight. Choose it like a domain name — boring, stable, forever.

**The five required scalars are required on purpose** — Cloudflare demands an explicit choice for `cacheInvalidateOnUpdate`, `cacheTtl`, `collectLogs`, `rateLimitingInterval`, and `rateLimitingLimit`; there are no server defaults. `cacheTtl: 0` and `rateLimitingInterval: 0` are the explicit OFF switches, so even a minimal gateway states all five.

**Lock the endpoint** — without `authentication: true`, the gateway endpoint is callable by anyone who knows the URL slug. Enabling it requires callers to present a `cf-aig-authorization` token on every request — the right posture for anything beyond an experiment.

**ZDR versus logging** — `zdr: true` (Zero Data Retention) tells Cloudflare never to store request/response bodies; `collectLogs`, `logpush`, and `logManagement` exist to store or ship exactly that data. Cloudflare enforces the legal combinations server-side, so contradictions surface as API rejections at deploy, not plan errors. Decide the data posture first, then configure logging around it.

**Spend-limit rule IDs are identities** — every rule in `spendLimits.rules` requires its own unique `id`, and the spec enforces it because the provider's schema ships a leaked example value as the default: rules authored without IDs would share one identity and silently collapse into a single budget. Pick short stable slugs like `daily-cap`.

**Route graphs replace, never update** — a dynamic route's `elements` list is create-only at the provider: any graph edit — one condition, one model swap — replaces that route object. This is safe (requests re-resolve the route by name on the next call), so a replace on a route in the plan is designed behavior, not a mistake. Renaming a route is its only in-place update. Model elements carry a wiring contract of their own: every `model` node needs a `fallback` edge plus `timeout` and `retries` properties — Cloudflare rejects the create (400 code 7001) without them.

**Guardrails take both sides** — `guardrails` requires both `prompt` and `response` objects; an empty object means "evaluate nothing on this side". Each control names a hazard-category code (`p1`, `s1`–`s13`) and takes FLAG (log the match) or BLOCK (refuse); an absent code leaves that category unevaluated.

**Two credentials the provider forgot to mark** — `otel[].authorization` and `stripe.authorization` are credentials the upstream provider leaves unmarked as sensitive. The spec treats both as sensitive: provide managed-secret references, never paste a bearer token or Stripe key as a literal.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareSecretsStore** (optional) | `storeId` | `status.outputs.store_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `gateway_id` | The gateway's URL slug | Composing the endpoint URL in Workers and application configuration |
| `dynamic_route_ids` | Each managed route's ID, keyed by route name | Route-level import identity and audit tooling |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Guarded production gateway** — response caching, a sliding rate limit, exponential retries, gateway authentication on every call, a prompt guardrail, and a daily spend cap: cost control and abuse protection from day one. Start from the **Guarded production gateway** preset.

**Cheap-first routing** — a named route graph that sends free-tier requests to a Workers AI model and everything else to a premium model, with the cheap arm falling back to the smart one on failure. Model cost follows customer tier. Start from the **Routed gateway (cheap-first)** preset.

**Budget partitioning by metadata** — one spend rule with `metadata` in `partition` mode tracks a separate budget per observed value (per team, per customer tier), instead of authoring one rule per value.

## Works With

- [**Cloudflare Secrets Store**](/infra-catalog/cloudflare-secrets-store) — the vault behind `storeId` for Bring Your Own provider Keys
- [**Cloudflare Secrets Store Secret**](/infra-catalog/cloudflare-secrets-store-secret) — the provider API keys themselves, scoped for AI Gateway use
- [**Cloudflare Worker**](/infra-catalog/cloudflare-worker) — Workers calling models through the gateway endpoint
