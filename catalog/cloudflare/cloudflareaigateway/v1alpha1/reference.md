# CloudflareAiGateway

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareAiGatewaySpec manages one AI Gateway: the control plane
Cloudflare puts in front of AI model traffic. Requests to model providers
route through the gateway's endpoint, which then applies caching,
rate limiting, retries, logging, guardrails, data-loss prevention, spend
limits, and dynamic request routing -- all configured here.

The gateway's id is its URL slug and is create-only: renaming replaces
the gateway (and its endpoint URL). Dynamic routes are managed as part of
this resource; note their elements list is create-only at the provider,
so editing a route's graph recreates that route object (never the
gateway itself).

## Example

```yaml
# Complete example manifest for CloudflareAiGateway. One gateway with
# caching, rate limiting, retries, log management, gateway authentication,
# prompt/response guardrails, a daily spend cap, and a dynamic route that
# sends free-tier traffic to a cheap model and everything else to a smarter
# one. The two dynamic-route objects below the gateway are what the offline
# plan proves as the designed 2-to-add (gateway + route).
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
  rate_limiting_technique: sliding
  retry:
    backoff: exponential
    delay: 1000
    max_attempts: 3
  log_management:
    max_records: 100000
    strategy: DELETE_OLDEST
  authentication: true
  guardrails:
    prompt:
      p1: BLOCK
      s1: FLAG
    response: {}
  spend_limits:
    enabled: true
    rules:
      - id: daily-cap
        limit: 50
        limit_type: cost
        window: 86400
        technique: sliding
  dynamic_routes:
    - name: cheap-first
      elements:
        - id: start
          type: start
          outputs:
            next:
              element_id: check
        - id: check
          type: conditional
          outputs:
            on_true:
              element_id: cheap
            on_false:
              element_id: smart
          properties:
            conditions: '{"metadata.tier": {"$eq": "free"}}'
        # Model nodes REQUIRE a fallback edge plus timeout and retries --
        # Cloudflare rejects the create with 400 code 7001 "Required"
        # without them (live-measured 2026-08-27; the provider schema
        # wrongly calls all three optional).
        - id: cheap
          type: model
          outputs:
            success:
              element_id: done
            fallback:
              element_id: smart
          properties:
            model: "@cf/meta/llama-3.1-8b-instruct"
            provider: workers-ai
            retries: 1
            timeout: 30000
        - id: smart
          type: model
          outputs:
            success:
              element_id: done
            fallback:
              element_id: done
          properties:
            model: gpt-4o
            provider: openai
            retries: 2
            timeout: 60000
        - id: done
          type: end
          outputs: {}
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.gatewayId` | `string` | yes |  |  |
| `spec.cacheInvalidateOnUpdate` | `bool` | yes |  |  |
| `spec.cacheTtl` | `int64` | yes |  |  |
| `spec.collectLogs` | `bool` | yes |  |  |
| `spec.rateLimitingInterval` | `int64` | yes |  |  |
| `spec.rateLimitingLimit` | `int64` | yes |  |  |
| `spec.rateLimitingTechnique` | `string` |  |  |  |
| `spec.retry` | `CloudflareAiGatewayRetry` |  |  |  |
| `spec.retry.backoff` | `string` |  |  |  |
| `spec.retry.delay` | `int64` |  |  |  |
| `spec.retry.maxAttempts` | `int64` |  |  |  |
| `spec.logManagement` | `CloudflareAiGatewayLogManagement` |  |  |  |
| `spec.logManagement.maxRecords` | `int64` |  |  |  |
| `spec.logManagement.strategy` | `string` |  |  |  |
| `spec.authentication` | `bool` |  |  |  |
| `spec.logpush` | `bool` |  |  |  |
| `spec.logpushPublicKey` | `string` |  |  |  |
| `spec.zdr` | `bool` |  |  |  |
| `spec.workersAiBillingMode` | `string` |  |  |  |
| `spec.storeId` | `string \| valueFrom` |  |  | CloudflareSecretsStore (`status.outputs.store_id`) |
| `spec.dlp` | `CloudflareAiGatewayDlp` |  |  |  |
| `spec.dlp.enabled` | `bool` | yes |  |  |
| `spec.dlp.action` | `string` |  |  |  |
| `spec.dlp.profiles` | `[]string` |  |  |  |
| `spec.dlp.policies` | `[]CloudflareAiGatewayDlpPolicy` |  |  |  |
| `spec.dlp.policies[].id` | `string` | yes |  |  |
| `spec.dlp.policies[].enabled` | `bool` | yes |  |  |
| `spec.dlp.policies[].action` | `string` | yes |  |  |
| `spec.dlp.policies[].check` | `[]string` | yes |  |  |
| `spec.dlp.policies[].profiles` | `[]string` | yes |  |  |
| `spec.guardrails` | `CloudflareAiGatewayGuardrails` |  |  |  |
| `spec.guardrails.prompt` | `CloudflareAiGatewayGuardrailsControls` | yes |  |  |
| `spec.guardrails.prompt.p1` | `string` |  |  |  |
| `spec.guardrails.prompt.s1` | `string` |  |  |  |
| `spec.guardrails.prompt.s2` | `string` |  |  |  |
| `spec.guardrails.prompt.s3` | `string` |  |  |  |
| `spec.guardrails.prompt.s4` | `string` |  |  |  |
| `spec.guardrails.prompt.s5` | `string` |  |  |  |
| `spec.guardrails.prompt.s6` | `string` |  |  |  |
| `spec.guardrails.prompt.s7` | `string` |  |  |  |
| `spec.guardrails.prompt.s8` | `string` |  |  |  |
| `spec.guardrails.prompt.s9` | `string` |  |  |  |
| `spec.guardrails.prompt.s10` | `string` |  |  |  |
| `spec.guardrails.prompt.s11` | `string` |  |  |  |
| `spec.guardrails.prompt.s12` | `string` |  |  |  |
| `spec.guardrails.prompt.s13` | `string` |  |  |  |
| `spec.guardrails.response` | `CloudflareAiGatewayGuardrailsControls` | yes |  |  |
| `spec.guardrails.response.p1` | `string` |  |  |  |
| `spec.guardrails.response.s1` | `string` |  |  |  |
| `spec.guardrails.response.s2` | `string` |  |  |  |
| `spec.guardrails.response.s3` | `string` |  |  |  |
| `spec.guardrails.response.s4` | `string` |  |  |  |
| `spec.guardrails.response.s5` | `string` |  |  |  |
| `spec.guardrails.response.s6` | `string` |  |  |  |
| `spec.guardrails.response.s7` | `string` |  |  |  |
| `spec.guardrails.response.s8` | `string` |  |  |  |
| `spec.guardrails.response.s9` | `string` |  |  |  |
| `spec.guardrails.response.s10` | `string` |  |  |  |
| `spec.guardrails.response.s11` | `string` |  |  |  |
| `spec.guardrails.response.s12` | `string` |  |  |  |
| `spec.guardrails.response.s13` | `string` |  |  |  |
| `spec.otel` | `[]CloudflareAiGatewayOtel` |  |  |  |
| `spec.otel[].url` | `string` | yes |  |  |
| `spec.otel[].headers` | `map<string, string>` |  |  |  |
| `spec.otel[].authorization` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.otel[].contentType` | `string` |  |  |  |
| `spec.stripe` | `CloudflareAiGatewayStripe` |  |  |  |
| `spec.stripe.authorization` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.stripe.usageEvents` | `[]CloudflareAiGatewayStripeUsageEvent` | yes |  |  |
| `spec.stripe.usageEvents[].payload` | `string` | yes |  |  |
| `spec.spendLimits` | `CloudflareAiGatewaySpendLimits` |  |  |  |
| `spec.spendLimits.enabled` | `bool` |  |  |  |
| `spec.spendLimits.rules` | `[]CloudflareAiGatewaySpendLimitsRule` |  |  |  |
| `spec.spendLimits.rules[].id` | `string` | yes |  |  |
| `spec.spendLimits.rules[].enabled` | `bool` |  |  |  |
| `spec.spendLimits.rules[].limit` | `double` | yes |  |  |
| `spec.spendLimits.rules[].limitType` | `string` | yes |  |  |
| `spec.spendLimits.rules[].window` | `int64` | yes |  |  |
| `spec.spendLimits.rules[].technique` | `string` |  |  |  |
| `spec.spendLimits.rules[].metadata` | `map<string, CloudflareAiGatewaySpendLimitsMetadataFilter>` |  |  |  |
| `spec.spendLimits.rules[].metadata.*.mode` | `string` | yes |  |  |
| `spec.spendLimits.rules[].metadata.*.values` | `[]string` |  |  |  |
| `spec.spendLimits.rules[].model` | `CloudflareAiGatewaySpendLimitsFilter` |  |  |  |
| `spec.spendLimits.rules[].model.mode` | `string` | yes |  |  |
| `spec.spendLimits.rules[].model.values` | `[]string` | yes |  |  |
| `spec.spendLimits.rules[].provider` | `CloudflareAiGatewaySpendLimitsFilter` |  |  |  |
| `spec.spendLimits.rules[].provider.mode` | `string` | yes |  |  |
| `spec.spendLimits.rules[].provider.values` | `[]string` | yes |  |  |
| `spec.dynamicRoutes` | `[]CloudflareAiGatewayDynamicRoute` |  |  |  |
| `spec.dynamicRoutes[].name` | `string` | yes |  |  |
| `spec.dynamicRoutes[].elements` | `[]CloudflareAiGatewayRouteElement` | yes |  |  |
| `spec.dynamicRoutes[].elements[].id` | `string` | yes |  |  |
| `spec.dynamicRoutes[].elements[].type` | `string` | yes |  |  |
| `spec.dynamicRoutes[].elements[].outputs` | `CloudflareAiGatewayRouteElementOutputs` | yes |  |  |
| `spec.dynamicRoutes[].elements[].outputs.next` | `CloudflareAiGatewayRouteElementBranch` |  |  |  |
| `spec.dynamicRoutes[].elements[].outputs.next.elementId` | `string` | yes |  |  |
| `spec.dynamicRoutes[].elements[].outputs.onTrue` | `CloudflareAiGatewayRouteElementBranch` |  |  |  |
| `spec.dynamicRoutes[].elements[].outputs.onTrue.elementId` | `string` | yes |  |  |
| `spec.dynamicRoutes[].elements[].outputs.onFalse` | `CloudflareAiGatewayRouteElementBranch` |  |  |  |
| `spec.dynamicRoutes[].elements[].outputs.onFalse.elementId` | `string` | yes |  |  |
| `spec.dynamicRoutes[].elements[].outputs.success` | `CloudflareAiGatewayRouteElementBranch` |  |  |  |
| `spec.dynamicRoutes[].elements[].outputs.success.elementId` | `string` | yes |  |  |
| `spec.dynamicRoutes[].elements[].outputs.fallback` | `CloudflareAiGatewayRouteElementBranch` |  |  |  |
| `spec.dynamicRoutes[].elements[].outputs.fallback.elementId` | `string` | yes |  |  |
| `spec.dynamicRoutes[].elements[].outputs.elementId` | `string` |  |  |  |
| `spec.dynamicRoutes[].elements[].properties` | `CloudflareAiGatewayRouteElementProperties` |  |  |  |
| `spec.dynamicRoutes[].elements[].properties.conditions` | `string` |  |  |  |
| `spec.dynamicRoutes[].elements[].properties.key` | `string` |  |  |  |
| `spec.dynamicRoutes[].elements[].properties.limit` | `double` |  |  |  |
| `spec.dynamicRoutes[].elements[].properties.limitType` | `string` |  |  |  |
| `spec.dynamicRoutes[].elements[].properties.window` | `double` |  |  |  |
| `spec.dynamicRoutes[].elements[].properties.model` | `string` |  |  |  |
| `spec.dynamicRoutes[].elements[].properties.provider` | `string` |  |  |  |
| `spec.dynamicRoutes[].elements[].properties.retries` | `double` |  |  |  |
| `spec.dynamicRoutes[].elements[].properties.timeout` | `double` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account the gateway lives in.

- rule: account_id must be a 32-character hex string
- rule: {"required":true}

### spec.gatewayId

`string` · required

The gateway's id -- the user-chosen slug that appears in the gateway's
endpoint URL (https://gateway.ai.cloudflare.com/v1/<account>/<id>/...).
Create-only: renaming replaces the gateway and changes the URL every
client calls.

- rule: {"string":{"minLen":"1"}}

### spec.cacheInvalidateOnUpdate

`bool` · required · optional (explicit presence)

Whether a configuration update invalidates previously cached responses.
Required by Cloudflare -- state the choice explicitly.

- rule: {"required":true}

### spec.cacheTtl

`int64` · required · optional (explicit presence)

How long cached model responses are served, in seconds. 0 disables
caching. Required by Cloudflare.

- rule: {"required":true,"int64":{"gte":"0"}}

### spec.collectLogs

`bool` · required · optional (explicit presence)

Whether request/response logs are collected for this gateway. Required
by Cloudflare. Logs are what the dashboard analytics, log management
cap, and logpush all feed from.

- rule: {"required":true}

### spec.rateLimitingInterval

`int64` · required · optional (explicit presence)

The rate-limiting window, in seconds. 0 disables rate limiting.
Required by Cloudflare.

- rule: {"required":true,"int64":{"gte":"0"}}

### spec.rateLimitingLimit

`int64` · required · optional (explicit presence)

The number of requests allowed per rate-limiting window. 0 disables
rate limiting. Required by Cloudflare.

- rule: {"required":true,"int64":{"gte":"0"}}

### spec.rateLimitingTechnique

`string`

How the rate-limiting window advances: "fixed" resets the counter at
interval boundaries; "sliding" evaluates a rolling window. Omit for
Cloudflare's default.

- rule: rate_limiting_technique must be fixed or sliding

### spec.retry

`CloudflareAiGatewayRetry`

Automatic retries toward the model provider on failures. Omit to leave
retries at Cloudflare's defaults. The provider models these as three
flat arguments (retry_backoff / retry_delay / retry_max_attempts); this
spec groups them for authoring clarity and the modules fan them out.

### spec.retry.backoff

`string`

The backoff shape between attempts: constant, linear, or exponential.

- rule: backoff must be constant, linear, or exponential

### spec.retry.delay

`int64` · optional (explicit presence)

The base delay between attempts, in milliseconds (up to 5000).

- rule: {"int64":{"lte":"5000","gte":"0"}}

### spec.retry.maxAttempts

`int64` · optional (explicit presence)

The maximum number of attempts (1-5).

- rule: {"int64":{"lte":"5","gte":"1"}}

### spec.logManagement

`CloudflareAiGatewayLogManagement`

Log-storage management for this gateway: a cap on stored log records
and what happens when the cap is reached. The provider models these as
two flat arguments (log_management / log_management_strategy); this spec
groups them and the modules fan them out. Omit to keep Cloudflare's
defaults (100000 records, DELETE_OLDEST) -- Cloudflare echoes those
defaults on every read, so the modules send them explicitly when unset
to stay refresh-stable (live-measured).

### spec.logManagement.maxRecords

`int64` · optional (explicit presence)

The maximum number of log records retained (10,000 to 10,000,000).

- rule: {"int64":{"lte":"10000000","gte":"10000"}}

### spec.logManagement.strategy

`string`

What happens when the cap is reached: STOP_INSERTING drops new logs,
DELETE_OLDEST evicts the oldest to make room.

- rule: strategy must be STOP_INSERTING or DELETE_OLDEST

### spec.authentication

`bool` · optional (explicit presence)

Require callers to present a gateway authentication token (the
cf-aig-authorization header) on every request. Without it the gateway
endpoint is callable by anyone who knows the URL slug. Cloudflare
echoes this toggle as false when never set, so the modules always send
it (unset means false, Cloudflare's own default).

### spec.logpush

`bool` · optional (explicit presence)

Push this gateway's logs to the account's Logpush destination.
Echoed as false when never set; the modules always send it.

### spec.logpushPublicKey

`string`

The public key used to encrypt logpushed log bodies, when logpush
encryption is desired.

### spec.zdr

`bool` · optional (explicit presence)

Zero Data Retention: prevent Cloudflare from storing request and
response bodies for this gateway. Mutually constraining with logging
features at the API (bodies cannot be both unretained and logged) --
Cloudflare enforces the combination server-side. Echoed as false when
never set; the modules always send it.

### spec.workersAiBillingMode

`string`

How Workers AI inference calls routed through this gateway are billed.
Cloudflare currently supports only "postpaid" (also the default) -- the
field exists so future billing modes are representable.

- rule: workers_ai_billing_mode must be postpaid (the only mode Cloudflare currently supports)

### spec.storeId

`string | valueFrom`

The Secrets Store holding provider API keys the gateway uses (Bring
Your Own Keys): a literal store ID, or a reference to a
CloudflareSecretsStore resource's store_id output.

- references: CloudflareSecretsStore (`status.outputs.store_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareSecretsStore, name: <that resource's name>, fieldPath: status.outputs.store_id}} -- a bare string does not parse

### spec.dlp

`CloudflareAiGatewayDlp`

Data-loss prevention over prompts and responses, driven by the
account's DLP profiles.

### spec.dlp.enabled

`bool` · required · optional (explicit presence)

Whether DLP screening is active for this gateway. Required inside dlp.

- rule: {"required":true}

### spec.dlp.action

`string`

The default action when a DLP profile matches: FLAG logs the match,
BLOCK refuses the request.

- rule: action must be FLAG or BLOCK

### spec.dlp.profiles

`[]string`

The DLP profile IDs applied gateway-wide.

### spec.dlp.policies

`[]CloudflareAiGatewayDlpPolicy`

Fine-grained DLP policies: per-policy profiles, direction, and action.

### spec.dlp.policies[].id

`string` · required

The policy's ID (user-chosen identifier for this row).

- rule: {"required":true}

### spec.dlp.policies[].enabled

`bool` · required · optional (explicit presence)

Whether this policy row is active.

- rule: {"required":true}

### spec.dlp.policies[].action

`string` · required

The action when the policy matches: FLAG logs, BLOCK refuses.

- rule: {"required":true,"string":{"in":["FLAG","BLOCK"]}}

### spec.dlp.policies[].check

`[]string` · required

Which traffic direction the policy screens: the REQUEST (prompt), the
RESPONSE (model output), or both.

- rule: {"repeated":{"minItems":"1","items":{"string":{"in":["REQUEST","RESPONSE"]}}}}

### spec.dlp.policies[].profiles

`[]string` · required

The DLP profile IDs this policy evaluates.

- rule: {"repeated":{"minItems":"1"}}

### spec.guardrails

`CloudflareAiGatewayGuardrails`

Content guardrails: per-hazard-category FLAG/BLOCK controls evaluated
on prompts and on model responses. WRITE-ONLY AT THE API: Cloudflare
accepts guardrails on create/update but no read ever returns them
(live-measured), so at terraform-provider-cloudflare v5.23.0-v5.24.0 a
Terraform configuration carrying guardrails re-plans an in-place update
forever (every apply re-delivers the same values; the settings DO take
effect at Cloudflare). Pulumi is unaffected.

### spec.guardrails.prompt

`CloudflareAiGatewayGuardrailsControls` · required

Controls applied to incoming prompts. Required when guardrails is set
(send an empty object to evaluate nothing on prompts).

- rule: {"required":true}
- rule: every guardrail control must be FLAG or BLOCK

### spec.guardrails.prompt.p1

`string`

Hazard category P1.

### spec.guardrails.prompt.s1

`string`

Hazard category S1.

### spec.guardrails.prompt.s2

`string`

Hazard category S2.

### spec.guardrails.prompt.s3

`string`

Hazard category S3.

### spec.guardrails.prompt.s4

`string`

Hazard category S4.

### spec.guardrails.prompt.s5

`string`

Hazard category S5.

### spec.guardrails.prompt.s6

`string`

Hazard category S6.

### spec.guardrails.prompt.s7

`string`

Hazard category S7.

### spec.guardrails.prompt.s8

`string`

Hazard category S8.

### spec.guardrails.prompt.s9

`string`

Hazard category S9.

### spec.guardrails.prompt.s10

`string`

Hazard category S10.

### spec.guardrails.prompt.s11

`string`

Hazard category S11.

### spec.guardrails.prompt.s12

`string`

Hazard category S12.

### spec.guardrails.prompt.s13

`string`

Hazard category S13.

### spec.guardrails.response

`CloudflareAiGatewayGuardrailsControls` · required

Controls applied to model responses. Required when guardrails is set
(send an empty object to evaluate nothing on responses).

- rule: {"required":true}
- rule: every guardrail control must be FLAG or BLOCK

### spec.guardrails.response.p1

`string`

Hazard category P1.

### spec.guardrails.response.s1

`string`

Hazard category S1.

### spec.guardrails.response.s2

`string`

Hazard category S2.

### spec.guardrails.response.s3

`string`

Hazard category S3.

### spec.guardrails.response.s4

`string`

Hazard category S4.

### spec.guardrails.response.s5

`string`

Hazard category S5.

### spec.guardrails.response.s6

`string`

Hazard category S6.

### spec.guardrails.response.s7

`string`

Hazard category S7.

### spec.guardrails.response.s8

`string`

Hazard category S8.

### spec.guardrails.response.s9

`string`

Hazard category S9.

### spec.guardrails.response.s10

`string`

Hazard category S10.

### spec.guardrails.response.s11

`string`

Hazard category S11.

### spec.guardrails.response.s12

`string`

Hazard category S12.

### spec.guardrails.response.s13

`string`

Hazard category S13.

### spec.otel

`[]CloudflareAiGatewayOtel`

OpenTelemetry export destinations for gateway telemetry. Write-only at
the API like guardrails (no read returns it) -- see that field's
Terraform re-plan caveat.

### spec.otel[].url

`string` · required

The OTLP endpoint URL telemetry is exported to.

- rule: {"required":true}

### spec.otel[].headers

`map<string, string>`

Headers sent with every export request (header name -> value).
Required by Cloudflare -- use an empty map when none are needed.

### spec.otel[].authorization

`string | valueFrom` · sensitive

The Authorization header value for the export endpoint (e.g. a bearer
token). The provider leaves this unmarked, but it is a credential --
this spec treats it as sensitive: provide a managed-secret reference
and the platform resolves it just-in-time at deploy.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.otel[].contentType

`string`

The export payload encoding: json or protobuf. Omit for Cloudflare's
default (json).

- rule: content_type must be json or protobuf

### spec.stripe

`CloudflareAiGatewayStripe`

Stripe usage-based billing integration: report gateway usage events to
Stripe.

### spec.stripe.authorization

`string | valueFrom` · required · sensitive

The Stripe API credential used to report usage. The provider leaves
this unmarked, but it is a credential -- this spec treats it as
sensitive: provide a managed-secret reference and the platform resolves
it just-in-time at deploy.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.stripe.usageEvents

`[]CloudflareAiGatewayStripeUsageEvent` · required

The usage events reported to Stripe.

- rule: {"repeated":{"minItems":"1"}}

### spec.stripe.usageEvents[].payload

`string` · required

The event payload template.

- rule: {"required":true}

### spec.spendLimits

`CloudflareAiGatewaySpendLimits`

Spend limits: cost budgets over configurable windows, optionally
filtered by model, provider, or request metadata. Write-only at the API
like guardrails (no read returns it) -- see that field's Terraform
re-plan caveat.

- rule: every spend-limit rule needs its own unique id -- Cloudflare's API defaults omitted ids to one shared placeholder value, which silently collapses multiple rules into one

### spec.spendLimits.enabled

`bool` · optional (explicit presence)

Whether spend limits are enforced. Cloudflare defaults this to false.

### spec.spendLimits.rules

`[]CloudflareAiGatewaySpendLimitsRule`

The budget rules. Each rule caps spend (in the account's billing
currency) over a window, optionally scoped by model, provider, or
request metadata.

### spec.spendLimits.rules[].id

`string` · required

This rule's identifier. REQUIRED HERE though the provider marks it
optional: the provider's schema ships a leaked example value as the
default, so every rule authored without an id would share one identity
and silently collapse. Choose short stable slugs (e.g. "daily-cap").

- rule: {"required":true}

### spec.spendLimits.rules[].enabled

`bool` · optional (explicit presence)

Whether this rule is enforced. Cloudflare defaults it to true.

### spec.spendLimits.rules[].limit

`double` · required · optional (explicit presence)

The spend cap for the window.

- rule: {"required":true,"double":{"gte":0}}

### spec.spendLimits.rules[].limitType

`string` · required

What the limit measures. Cloudflare currently supports only "cost".

- rule: {"required":true,"string":{"in":["cost"]}}

### spec.spendLimits.rules[].window

`int64` · required · optional (explicit presence)

The budget window, in seconds.

- rule: {"required":true,"int64":{"gte":"0"}}

### spec.spendLimits.rules[].technique

`string`

How the window advances: fixed or sliding. Cloudflare defaults to
sliding.

- rule: technique must be fixed or sliding

### spec.spendLimits.rules[].metadata

`map<string, CloudflareAiGatewaySpendLimitsMetadataFilter>`

Scope the budget by request metadata: metadata key -> partition/filter
over its values. "partition" tracks a separate budget per value;
"filter" counts only requests matching the listed values.

### spec.spendLimits.rules[].metadata.*.mode

`string` · required

partition tracks a separate budget per observed value; filter counts
only requests whose value is in `values`.

- rule: {"required":true,"string":{"in":["partition","filter"]}}

### spec.spendLimits.rules[].metadata.*.values

`[]string`

The values filtered on (used by filter mode).

### spec.spendLimits.rules[].model

`CloudflareAiGatewaySpendLimitsFilter`

Scope the budget to specific models.

### spec.spendLimits.rules[].model.mode

`string` · required

The only mode Cloudflare supports here is "filter".

- rule: {"required":true,"string":{"in":["filter"]}}

### spec.spendLimits.rules[].model.values

`[]string` · required

The model or provider names counted against this budget.

- rule: {"repeated":{"minItems":"1"}}

### spec.spendLimits.rules[].provider

`CloudflareAiGatewaySpendLimitsFilter`

Scope the budget to specific model providers (wire name "provider";
the Terraform provider calls it ai_gateway_provider to dodge the
reserved word).

### spec.spendLimits.rules[].provider.mode

`string` · required

The only mode Cloudflare supports here is "filter".

- rule: {"required":true,"string":{"in":["filter"]}}

### spec.spendLimits.rules[].provider.values

`[]string` · required

The model or provider names counted against this budget.

- rule: {"repeated":{"minItems":"1"}}

### spec.dynamicRoutes

`[]CloudflareAiGatewayDynamicRoute`

Dynamic request routes: named routing graphs (condition, percentage,
rate, and model nodes) that requests can address by route name. Each
route is its own provider object keyed by name; a route's elements list
is CREATE-ONLY at the provider, so editing a graph recreates that route
(in-flight requests re-resolve on the next call). At
terraform-provider-cloudflare v5.23.0-v5.24.0 the route resource's Read
cannot restore elements (the API returns the graph only under
version.data), so a Terraform-managed route is DESTROYED AND RECREATED
on every apply -- prefer Pulumi for dynamic routes until a provider
release fixes the Read (live-measured; the kind's e2e profile carries
the full evidence).

### spec.dynamicRoutes[].name

`string` · required

The route's name -- what requests address. Renaming is the only
in-place update the provider supports; any change to the elements
recreates the route object.

- rule: {"required":true}

### spec.dynamicRoutes[].elements

`[]CloudflareAiGatewayRouteElement` · required

The routing graph's nodes. CREATE-ONLY at the provider: any edit
replaces the whole route object.

- rule: {"repeated":{"minItems":"1"}}

### spec.dynamicRoutes[].elements[].id

`string` · required

This node's identifier within the graph (what outputs edges point at).

- rule: {"required":true}

### spec.dynamicRoutes[].elements[].type

`string` · required

The node type: start (the entry), conditional (branch on conditions),
percentage (split traffic), rate (rate-limit branch), model (send to a
model), end (terminate).

- rule: {"required":true,"string":{"in":["start","conditional","percentage","rate","model","end"]}}

### spec.dynamicRoutes[].elements[].outputs

`CloudflareAiGatewayRouteElementOutputs` · required

Where the flow goes next. Which edges apply depends on the node type
(e.g. conditional uses on_true/on_false, rate uses next/fallback,
model uses success/fallback).

- rule: {"required":true}

### spec.dynamicRoutes[].elements[].outputs.next

`CloudflareAiGatewayRouteElementBranch`

The unconditional next node.

### spec.dynamicRoutes[].elements[].outputs.next.elementId

`string` · required

The id of the element this edge leads to.

- rule: {"required":true}

### spec.dynamicRoutes[].elements[].outputs.onTrue

`CloudflareAiGatewayRouteElementBranch`

The branch taken when a conditional node's conditions match (wire
name: true).

### spec.dynamicRoutes[].elements[].outputs.onTrue.elementId

`string` · required

The id of the element this edge leads to.

- rule: {"required":true}

### spec.dynamicRoutes[].elements[].outputs.onFalse

`CloudflareAiGatewayRouteElementBranch`

The branch taken when a conditional node's conditions do not match
(wire name: false).

### spec.dynamicRoutes[].elements[].outputs.onFalse.elementId

`string` · required

The id of the element this edge leads to.

- rule: {"required":true}

### spec.dynamicRoutes[].elements[].outputs.success

`CloudflareAiGatewayRouteElementBranch`

The branch taken on success (model nodes).

### spec.dynamicRoutes[].elements[].outputs.success.elementId

`string` · required

The id of the element this edge leads to.

- rule: {"required":true}

### spec.dynamicRoutes[].elements[].outputs.fallback

`CloudflareAiGatewayRouteElementBranch`

The branch taken on failure or when a limit trips (model and rate
nodes).
REQUIRED on model nodes: Cloudflare rejects a model element without a
fallback edge with 400 code 7001 "Required" (live-measured; the
provider schema wrongly calls it optional). Point it at the element to
try when the model call fails -- the end node when there is no backup.

### spec.dynamicRoutes[].elements[].outputs.fallback.elementId

`string` · required

The id of the element this edge leads to.

- rule: {"required":true}

### spec.dynamicRoutes[].elements[].outputs.elementId

`string`

A bare next-element id, for node types that take the edge directly
rather than as a named branch.

### spec.dynamicRoutes[].elements[].properties

`CloudflareAiGatewayRouteElementProperties`

The node's type-specific settings.

### spec.dynamicRoutes[].elements[].properties.conditions

`string`

Conditional nodes: the condition expression, as a JSON document.
Whitespace and key order are normalized by the API (the provider
diffs it semantically).

### spec.dynamicRoutes[].elements[].properties.key

`string`

Rate nodes: the key the rate is tracked by.

### spec.dynamicRoutes[].elements[].properties.limit

`double` · optional (explicit presence)

Rate nodes: the limit value.

### spec.dynamicRoutes[].elements[].properties.limitType

`string`

Rate nodes: what the limit counts -- requests (count) or spend (cost).

- rule: limit_type must be count or cost

### spec.dynamicRoutes[].elements[].properties.window

`double` · optional (explicit presence)

Rate nodes: the window the limit applies over, in seconds.

### spec.dynamicRoutes[].elements[].properties.model

`string`

Model nodes: the model requests are sent to.

### spec.dynamicRoutes[].elements[].properties.provider

`string`

Model nodes: the model provider (wire name "provider").

### spec.dynamicRoutes[].elements[].properties.retries

`double` · optional (explicit presence)

Model nodes: retry attempts toward this model. REQUIRED on model nodes
(400 code 7001 without it -- live-measured; the provider schema wrongly
calls it optional).

### spec.dynamicRoutes[].elements[].properties.timeout

`double` · optional (explicit presence)

Model nodes: the request timeout, in milliseconds. REQUIRED on model
nodes (400 code 7001 without it -- live-measured; the provider schema
wrongly calls it optional).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareAiGateway, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.gateway_id` | `string` | The gateway's id (URL slug) -- the segment clients put in the gateway endpoint URL, and what dynamic-routing objects attach to. |
| `status.outputs.dynamic_route_ids` | `map<string, string>` | The id of each managed dynamic route, keyed by route name. What the per-route import identity ({account_id}/{gateway_id}/{route_id}) derives from. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.storeId` | CloudflareSecretsStore | `status.outputs.store_id` |

## See Also

- [Overview](../README.md)
