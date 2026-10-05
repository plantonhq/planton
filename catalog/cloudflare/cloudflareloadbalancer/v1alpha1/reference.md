# CloudflareLoadBalancer

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareLoadBalancerSpec defines a zone-scoped Cloudflare Load Balancer. The
load balancer attaches a DNS hostname to a set of account-scoped pools
(CloudflareLoadBalancerPool) and steers traffic across them with the chosen
steering policy, session affinity, and optional geo-routing. Pools and their
monitors are independent, reusable resources referenced here by ID or reference.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareLoadBalancer
metadata:
  name: lb-hack
spec:
  hostname: lb.planton-example.com
  zoneId:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  defaultPools:
    - value: "f1e2d3c4b5a6978869584a3b2c1d0e0f"
  fallbackPool:
    value: "f1e2d3c4b5a6978869584a3b2c1d0e0f"
  proxied: true
  steeringPolicy: random
  sessionAffinity: cookie
  sessionAffinityTtl: 1800
  sessionAffinityAttributes:
    samesite: Lax
    secure: Always
    zeroDowntimeFailover: temporary
  adaptiveRouting:
    failoverAcrossPools: true
  locationStrategy:
    mode: resolver_ip
    preferEcs: proximity
  randomSteering:
    defaultWeight: 0.5
    poolWeights:
      f1e2d3c4b5a6978869584a3b2c1d0e0f: 0.8
  regionPools:
    - code: WNAM
      poolIds:
        - value: "f1e2d3c4b5a6978869584a3b2c1d0e0f"
  networks:
    - "network-hack"
  rules:
    - name: maintenance-page
      condition: http.request.uri.path contains "/maintenance"
      priority: 1
      terminates: true
      fixedResponse:
        contentType: text/html
        location: https://status.planton-example.com
        messageBody: "<h1>Down for maintenance</h1>"
        statusCode: 503
    - name: api-steering
      condition: http.request.uri.path contains "/api"
      disabled: false
      overrides:
        steeringPolicy: "off"
        sessionAffinity: none
        sessionAffinityTtl: 1800
        ttl: 30
        defaultPools:
          - value: "f1e2d3c4b5a6978869584a3b2c1d0e0f"
        fallbackPool:
          value: "f1e2d3c4b5a6978869584a3b2c1d0e0f"
        adaptiveRouting:
          failoverAcrossPools: true
        locationStrategy:
          mode: pop
          preferEcs: never
        randomSteering:
          defaultWeight: 0.3
        regionPools:
          - code: ENAM
            poolIds:
              - value: "f1e2d3c4b5a6978869584a3b2c1d0e0f"
        countryPools:
          - code: US
            poolIds:
              - value: "f1e2d3c4b5a6978869584a3b2c1d0e0f"
        popPools:
          - code: LAX
            poolIds:
              - value: "f1e2d3c4b5a6978869584a3b2c1d0e0f"
        sessionAffinityAttributes:
          samesite: Auto
          secure: Auto
          zeroDowntimeFailover: sticky
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.hostname` | `string` | yes |  |  |
| `spec.zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.proxied` | `bool` |  | `true` |  |
| `spec.sessionAffinity` | `enum` |  |  |  |
| `spec.steeringPolicy` | `enum` |  |  |  |
| `spec.defaultPools` | `[]string \| valueFrom` | yes |  | CloudflareLoadBalancerPool (`status.outputs.pool_id`) |
| `spec.fallbackPool` | `string \| valueFrom` | yes |  | CloudflareLoadBalancerPool (`status.outputs.pool_id`) |
| `spec.description` | `string` |  |  |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.ttl` | `double` |  |  |  |
| `spec.sessionAffinityTtl` | `double` |  |  |  |
| `spec.sessionAffinityAttributes` | `CloudflareLoadBalancerSessionAffinityAttributes` |  |  |  |
| `spec.sessionAffinityAttributes.drainDuration` | `double` |  |  |  |
| `spec.sessionAffinityAttributes.headers` | `[]string` |  |  |  |
| `spec.sessionAffinityAttributes.requireAllHeaders` | `bool` |  |  |  |
| `spec.sessionAffinityAttributes.samesite` | `string` |  |  |  |
| `spec.sessionAffinityAttributes.secure` | `string` |  |  |  |
| `spec.sessionAffinityAttributes.zeroDowntimeFailover` | `string` |  |  |  |
| `spec.regionPools` | `[]CloudflareLoadBalancerGeoPools` |  |  |  |
| `spec.regionPools[].code` | `string` | yes |  |  |
| `spec.regionPools[].poolIds` | `[]string \| valueFrom` | yes |  | CloudflareLoadBalancerPool (`status.outputs.pool_id`) |
| `spec.countryPools` | `[]CloudflareLoadBalancerGeoPools` |  |  |  |
| `spec.countryPools[].code` | `string` | yes |  |  |
| `spec.countryPools[].poolIds` | `[]string \| valueFrom` | yes |  | CloudflareLoadBalancerPool (`status.outputs.pool_id`) |
| `spec.popPools` | `[]CloudflareLoadBalancerGeoPools` |  |  |  |
| `spec.popPools[].code` | `string` | yes |  |  |
| `spec.popPools[].poolIds` | `[]string \| valueFrom` | yes |  | CloudflareLoadBalancerPool (`status.outputs.pool_id`) |
| `spec.adaptiveRouting` | `CloudflareLoadBalancerAdaptiveRouting` |  |  |  |
| `spec.adaptiveRouting.failoverAcrossPools` | `bool` |  |  |  |
| `spec.locationStrategy` | `CloudflareLoadBalancerLocationStrategy` |  |  |  |
| `spec.locationStrategy.mode` | `string` |  |  |  |
| `spec.locationStrategy.preferEcs` | `string` |  |  |  |
| `spec.randomSteering` | `CloudflareLoadBalancerRandomSteering` |  |  |  |
| `spec.randomSteering.defaultWeight` | `double` |  |  |  |
| `spec.randomSteering.poolWeights` | `map<string, double>` |  |  |  |
| `spec.rules` | `[]CloudflareLoadBalancerRule` |  |  |  |
| `spec.rules[].name` | `string` |  |  |  |
| `spec.rules[].condition` | `string` |  |  |  |
| `spec.rules[].priority` | `int32` |  |  |  |
| `spec.rules[].disabled` | `bool` |  |  |  |
| `spec.rules[].terminates` | `bool` |  |  |  |
| `spec.rules[].fixedResponse` | `CloudflareLoadBalancerRuleFixedResponse` |  |  |  |
| `spec.rules[].fixedResponse.contentType` | `string` |  |  |  |
| `spec.rules[].fixedResponse.location` | `string` |  |  |  |
| `spec.rules[].fixedResponse.messageBody` | `string` |  |  |  |
| `spec.rules[].fixedResponse.statusCode` | `int32` |  |  |  |
| `spec.rules[].overrides` | `CloudflareLoadBalancerRuleOverrides` |  |  |  |
| `spec.rules[].overrides.adaptiveRouting` | `CloudflareLoadBalancerAdaptiveRouting` |  |  |  |
| `spec.rules[].overrides.adaptiveRouting.failoverAcrossPools` | `bool` |  |  |  |
| `spec.rules[].overrides.countryPools` | `[]CloudflareLoadBalancerGeoPools` |  |  |  |
| `spec.rules[].overrides.countryPools[].code` | `string` | yes |  |  |
| `spec.rules[].overrides.countryPools[].poolIds` | `[]string \| valueFrom` | yes |  | CloudflareLoadBalancerPool (`status.outputs.pool_id`) |
| `spec.rules[].overrides.defaultPools` | `[]string \| valueFrom` |  |  | CloudflareLoadBalancerPool (`status.outputs.pool_id`) |
| `spec.rules[].overrides.fallbackPool` | `string \| valueFrom` |  |  | CloudflareLoadBalancerPool (`status.outputs.pool_id`) |
| `spec.rules[].overrides.locationStrategy` | `CloudflareLoadBalancerLocationStrategy` |  |  |  |
| `spec.rules[].overrides.locationStrategy.mode` | `string` |  |  |  |
| `spec.rules[].overrides.locationStrategy.preferEcs` | `string` |  |  |  |
| `spec.rules[].overrides.popPools` | `[]CloudflareLoadBalancerGeoPools` |  |  |  |
| `spec.rules[].overrides.popPools[].code` | `string` | yes |  |  |
| `spec.rules[].overrides.popPools[].poolIds` | `[]string \| valueFrom` | yes |  | CloudflareLoadBalancerPool (`status.outputs.pool_id`) |
| `spec.rules[].overrides.randomSteering` | `CloudflareLoadBalancerRandomSteering` |  |  |  |
| `spec.rules[].overrides.randomSteering.defaultWeight` | `double` |  |  |  |
| `spec.rules[].overrides.randomSteering.poolWeights` | `map<string, double>` |  |  |  |
| `spec.rules[].overrides.regionPools` | `[]CloudflareLoadBalancerGeoPools` |  |  |  |
| `spec.rules[].overrides.regionPools[].code` | `string` | yes |  |  |
| `spec.rules[].overrides.regionPools[].poolIds` | `[]string \| valueFrom` | yes |  | CloudflareLoadBalancerPool (`status.outputs.pool_id`) |
| `spec.rules[].overrides.sessionAffinity` | `enum` |  |  |  |
| `spec.rules[].overrides.sessionAffinityAttributes` | `CloudflareLoadBalancerSessionAffinityAttributes` |  |  |  |
| `spec.rules[].overrides.sessionAffinityAttributes.drainDuration` | `double` |  |  |  |
| `spec.rules[].overrides.sessionAffinityAttributes.headers` | `[]string` |  |  |  |
| `spec.rules[].overrides.sessionAffinityAttributes.requireAllHeaders` | `bool` |  |  |  |
| `spec.rules[].overrides.sessionAffinityAttributes.samesite` | `string` |  |  |  |
| `spec.rules[].overrides.sessionAffinityAttributes.secure` | `string` |  |  |  |
| `spec.rules[].overrides.sessionAffinityAttributes.zeroDowntimeFailover` | `string` |  |  |  |
| `spec.rules[].overrides.sessionAffinityTtl` | `double` |  |  |  |
| `spec.rules[].overrides.steeringPolicy` | `enum` |  |  |  |
| `spec.rules[].overrides.ttl` | `double` |  |  |  |
| `spec.networks` | `[]string` |  |  |  |

## Field Details

### spec.hostname

`string` · required

The DNS hostname to associate with this load balancer (e.g.
"app.example.com"). Must be the FULLY QUALIFIED name -- Cloudflare rejects
a bare label like "app" with 400 code 1002 "Invalid load balancer name:
invalid hostname" (measured live). If a DNS record with this name already
exists, the load balancer takes precedence.

- rule: {"required":true}

### spec.zoneId

`string | valueFrom` · required

The Cloudflare zone that owns the hostname, as a literal zone ID or a
reference to a CloudflareDnsZone.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.proxied

`bool` · optional (explicit presence)

Whether the hostname is proxied through Cloudflare (orange cloud). Defaults
to false (gray cloud) when unset; most HTTP load balancers set this true.

- default: `true`

### spec.sessionAffinity

`enum`

Session-affinity mode. Defaults to none.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `none` -- No session affinity (the default).
- `cookie` -- Cookie-based affinity.
- `ip_cookie` -- Cookie-based affinity with stable initial selection by client IP.
- `header` -- Header-based affinity (see session_affinity_attributes.headers).

### spec.steeringPolicy

`enum`

Traffic-steering policy. Defaults to off (static failover over default_pools).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `off` -- Use default_pools as a static failover priority list (also the default).
- `geo` -- Use region_pools / country_pools / pop_pools (geo-routing).
- `random` -- Select a pool at random, weighted by random_steering.
- `dynamic_latency` -- Use round-trip time to pick the closest healthy pool in default_pools.
- `proximity` -- Use pool latitude/longitude to pick the closest pool.
- `least_outstanding_requests` -- Pick the pool with the fewest outstanding requests (weighted).
- `least_connections` -- Pick the pool with the fewest open connections (weighted).

### spec.defaultPools

`[]string | valueFrom` · required

Ordered list of pools used by default (and when no geo mapping matches),
by failover priority. Each is a literal pool ID or a reference to a
CloudflareLoadBalancerPool. At least one is required.

- references: CloudflareLoadBalancerPool (`status.outputs.pool_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareLoadBalancerPool, name: <that resource's name>, fieldPath: status.outputs.pool_id}} -- a bare string does not parse

### spec.fallbackPool

`string | valueFrom` · required

The pool of last resort, used when every other pool is unhealthy. A literal
pool ID or a reference to a CloudflareLoadBalancerPool.

- references: CloudflareLoadBalancerPool (`status.outputs.pool_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareLoadBalancerPool, name: <that resource's name>, fieldPath: status.outputs.pool_id}} -- a bare string does not parse

### spec.description

`string`

Human-readable description shown in the Cloudflare dashboard.

### spec.enabled

`bool` · optional (explicit presence)

Whether the load balancer is enabled. Defaults to true when unset.

- default: `true`

### spec.ttl

`double`

TTL (seconds) of the DNS entry returned by the load balancer. Applies only to
gray-clouded (unproxied) load balancers. Leave 0 to use Cloudflare's default.

- rule: ttl must be 0 (default) or a positive number of seconds

### spec.sessionAffinityTtl

`double`

Seconds until a client's affinity session expires. Leave 0 for the
Cloudflare default for the chosen session_affinity mode.

- rule: session_affinity_ttl must be 0 (default) or a positive number of seconds

### spec.sessionAffinityAttributes

`CloudflareLoadBalancerSessionAffinityAttributes`

Fine-grained session-affinity attributes (drain, headers, cookie flags,
zero-downtime failover). Omit to use defaults.

- rule: samesite "None" cannot be combined with secure "Never"

### spec.sessionAffinityAttributes.drainDuration

`double`

Drain duration (seconds) applied when affinity is enabled.

- rule: drain_duration must be 0 or a positive number of seconds

### spec.sessionAffinityAttributes.headers

`[]string`

HTTP header names that header-based affinity is keyed on (required when
session_affinity is header). Use "cookie:<name>,<name>" to scope to cookies.

### spec.sessionAffinityAttributes.requireAllHeaders

`bool`

When header affinity is enabled, require ALL listed headers (true) or AT
LEAST ONE (false) for a session to be created.

### spec.sessionAffinityAttributes.samesite

`string`

SameSite attribute on the affinity cookie: one of "Auto", "Lax", "None",
"Strict". Empty uses Cloudflare's default ("Auto").

- rule: samesite must be one of "Auto", "Lax", "None", "Strict"

### spec.sessionAffinityAttributes.secure

`string`

Secure attribute on the affinity cookie: one of "Auto", "Always", "Never".
Empty uses Cloudflare's default ("Auto").

- rule: secure must be one of "Auto", "Always", "Never"

### spec.sessionAffinityAttributes.zeroDowntimeFailover

`string`

Zero-downtime failover behavior for pinned sessions: one of "none",
"temporary", "sticky". Empty uses Cloudflare's default ("none").

- rule: zero_downtime_failover must be one of "none", "temporary", "sticky"

### spec.regionPools

`[]CloudflareLoadBalancerGeoPools`

Region-code -> ordered pool list for geo steering. Used when
steering_policy is geo (or auto-geo). Regions not listed fall back to
default_pools.

### spec.regionPools[].code

`string` · required

The geo code this mapping applies to: a region code (e.g. "WNAM"), an ISO
country code (e.g. "US"), or a Cloudflare PoP code (e.g. "LAX"), depending on
which map this entry belongs to.

- rule: {"required":true}

### spec.regionPools[].poolIds

`[]string | valueFrom` · required

Ordered pools (by failover priority) for this code. Each is a literal pool ID
or a reference to a CloudflareLoadBalancerPool.

- references: CloudflareLoadBalancerPool (`status.outputs.pool_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareLoadBalancerPool, name: <that resource's name>, fieldPath: status.outputs.pool_id}} -- a bare string does not parse

### spec.countryPools

`[]CloudflareLoadBalancerGeoPools`

Country-code (ISO 3166-1 alpha-2) -> ordered pool list. Countries not listed
fall back to the matching region_pools entry, else default_pools.

### spec.countryPools[].code

`string` · required

The geo code this mapping applies to: a region code (e.g. "WNAM"), an ISO
country code (e.g. "US"), or a Cloudflare PoP code (e.g. "LAX"), depending on
which map this entry belongs to.

- rule: {"required":true}

### spec.countryPools[].poolIds

`[]string | valueFrom` · required

Ordered pools (by failover priority) for this code. Each is a literal pool ID
or a reference to a CloudflareLoadBalancerPool.

- references: CloudflareLoadBalancerPool (`status.outputs.pool_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareLoadBalancerPool, name: <that resource's name>, fieldPath: status.outputs.pool_id}} -- a bare string does not parse

### spec.popPools

`[]CloudflareLoadBalancerGeoPools`

Cloudflare PoP code -> ordered pool list (Enterprise only). PoPs not listed
fall back to country_pools, then region_pools, then default_pools.

### spec.popPools[].code

`string` · required

The geo code this mapping applies to: a region code (e.g. "WNAM"), an ISO
country code (e.g. "US"), or a Cloudflare PoP code (e.g. "LAX"), depending on
which map this entry belongs to.

- rule: {"required":true}

### spec.popPools[].poolIds

`[]string | valueFrom` · required

Ordered pools (by failover priority) for this code. Each is a literal pool ID
or a reference to a CloudflareLoadBalancerPool.

- references: CloudflareLoadBalancerPool (`status.outputs.pool_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareLoadBalancerPool, name: <that resource's name>, fieldPath: status.outputs.pool_id}} -- a bare string does not parse

### spec.adaptiveRouting

`CloudflareLoadBalancerAdaptiveRouting`

Adaptive-routing controls (zero-downtime failover across pools).

### spec.adaptiveRouting.failoverAcrossPools

`bool`

Extend zero-downtime failover to healthy origins in alternate pools when no
healthy origin remains in the same pool. Defaults to false.

### spec.locationStrategy

`CloudflareLoadBalancerLocationStrategy`

Location-steering controls for non-proxied requests.

### spec.locationStrategy.mode

`string`

Authoritative location when ECS is not used: "pop" (Cloudflare PoP) or
"resolver_ip" (DNS resolver GeoIP). Empty uses Cloudflare's default ("pop").

- rule: mode must be "pop" or "resolver_ip"

### spec.locationStrategy.preferEcs

`string`

Whether to prefer the ECS GeoIP location: one of "always", "never",
"proximity", "geo". Empty uses Cloudflare's default ("proximity").

- rule: prefer_ecs must be one of "always", "never", "proximity", "geo"

### spec.randomSteering

`CloudflareLoadBalancerRandomSteering`

Pool weighting for random / least-* steering policies.

### spec.randomSteering.defaultWeight

`double`

Default weight for pools not present in pool_weights (0.0–1.0). Leave 0 to
use the Cloudflare default (1).

- rule: default_weight must be between 0 and 1

### spec.randomSteering.poolWeights

`map<string, double>`

Per-pool weight overrides, keyed by pool ID. Each weight is relative to the
other pools in this load balancer.

### spec.rules

`[]CloudflareLoadBalancerRule`

Ordered list of traffic rules evaluated per request (the dashboard's
"Custom Rules"). Each rule has a condition expression; when it matches, the
rule either overrides this load balancer's steering for that request or
answers directly with a fixed response. The rule COUNT is capped by the
account's Load Balancing subscription tier -- Basic allows exactly 1 rule
per load balancer, and exceeding the cap fails the write with 400 code
1002 "rule count N exceeds limit M" (measured live). Note: the provider
schema labels this field "BETA Field Not General Access", but
load-balancing rules are a long-standing GA Cloudflare product feature in
production use.

### spec.rules[].name

`string`

Human-readable rule name, used only for dashboard readability.

### spec.rules[].condition

`string`

The condition expression to evaluate (e.g. `http.request.uri.path contains
"/api"`). An empty condition always matches.

### spec.rules[].priority

`int32` · optional (explicit presence)

Execution order relative to other rules: lower values run first; values
need not be sequential. LEAVE UNSET to let the array order of `rules`
assign priorities (Cloudflare's behavior when no rule provides one) — an
explicit 0 is a real priority, not "unset".

- rule: {"int32":{"gte":0}}

### spec.rules[].disabled

`bool`

Disable this rule without deleting it; a disabled rule is not evaluated.

### spec.rules[].terminates

`bool`

Stop evaluating later rules when this rule's condition matches. A rule
with a fixed_response is always terminating regardless of this flag.

### spec.rules[].fixedResponse

`CloudflareLoadBalancerRuleFixedResponse`

Respond to the client directly instead of routing to a pool. Supplying a
fixed response marks the rule as terminating.

### spec.rules[].fixedResponse.contentType

`string`

The HTTP Content-Type header to include in the response (e.g.
"application/json").

### spec.rules[].fixedResponse.location

`string`

The HTTP Location header to include in the response (for redirects).

### spec.rules[].fixedResponse.messageBody

`string`

Text to include as the HTTP response body.

### spec.rules[].fixedResponse.statusCode

`int32`

The HTTP status code to respond with (e.g. 200, 301, 503). Leave 0 to use
Cloudflare's default.

- rule: status_code must be 0 (default) or a valid HTTP status code (100-599)

### spec.rules[].overrides

`CloudflareLoadBalancerRuleOverrides`

Steering overrides applied to this load balancer while the rule's
condition is true. All override fields are optional; unset fields inherit
the load balancer's own configuration.

### spec.rules[].overrides.adaptiveRouting

`CloudflareLoadBalancerAdaptiveRouting`

Adaptive-routing override (zero-downtime failover across pools).

### spec.rules[].overrides.adaptiveRouting.failoverAcrossPools

`bool`

Extend zero-downtime failover to healthy origins in alternate pools when no
healthy origin remains in the same pool. Defaults to false.

### spec.rules[].overrides.countryPools

`[]CloudflareLoadBalancerGeoPools`

Country-code (ISO 3166-1 alpha-2) -> ordered pool list override.

### spec.rules[].overrides.countryPools[].code

`string` · required

The geo code this mapping applies to: a region code (e.g. "WNAM"), an ISO
country code (e.g. "US"), or a Cloudflare PoP code (e.g. "LAX"), depending on
which map this entry belongs to.

- rule: {"required":true}

### spec.rules[].overrides.countryPools[].poolIds

`[]string | valueFrom` · required

Ordered pools (by failover priority) for this code. Each is a literal pool ID
or a reference to a CloudflareLoadBalancerPool.

- references: CloudflareLoadBalancerPool (`status.outputs.pool_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareLoadBalancerPool, name: <that resource's name>, fieldPath: status.outputs.pool_id}} -- a bare string does not parse

### spec.rules[].overrides.defaultPools

`[]string | valueFrom`

Ordered default pool list override (by failover priority). Each is a
literal pool ID or a reference to a CloudflareLoadBalancerPool.

- references: CloudflareLoadBalancerPool (`status.outputs.pool_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareLoadBalancerPool, name: <that resource's name>, fieldPath: status.outputs.pool_id}} -- a bare string does not parse

### spec.rules[].overrides.fallbackPool

`string | valueFrom`

Fallback (pool-of-last-resort) override. A literal pool ID or a reference
to a CloudflareLoadBalancerPool.

- references: CloudflareLoadBalancerPool (`status.outputs.pool_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareLoadBalancerPool, name: <that resource's name>, fieldPath: status.outputs.pool_id}} -- a bare string does not parse

### spec.rules[].overrides.locationStrategy

`CloudflareLoadBalancerLocationStrategy`

Location-steering override for non-proxied requests.

### spec.rules[].overrides.locationStrategy.mode

`string`

Authoritative location when ECS is not used: "pop" (Cloudflare PoP) or
"resolver_ip" (DNS resolver GeoIP). Empty uses Cloudflare's default ("pop").

- rule: mode must be "pop" or "resolver_ip"

### spec.rules[].overrides.locationStrategy.preferEcs

`string`

Whether to prefer the ECS GeoIP location: one of "always", "never",
"proximity", "geo". Empty uses Cloudflare's default ("proximity").

- rule: prefer_ecs must be one of "always", "never", "proximity", "geo"

### spec.rules[].overrides.popPools

`[]CloudflareLoadBalancerGeoPools`

Cloudflare PoP code -> ordered pool list override (Enterprise only).

### spec.rules[].overrides.popPools[].code

`string` · required

The geo code this mapping applies to: a region code (e.g. "WNAM"), an ISO
country code (e.g. "US"), or a Cloudflare PoP code (e.g. "LAX"), depending on
which map this entry belongs to.

- rule: {"required":true}

### spec.rules[].overrides.popPools[].poolIds

`[]string | valueFrom` · required

Ordered pools (by failover priority) for this code. Each is a literal pool ID
or a reference to a CloudflareLoadBalancerPool.

- references: CloudflareLoadBalancerPool (`status.outputs.pool_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareLoadBalancerPool, name: <that resource's name>, fieldPath: status.outputs.pool_id}} -- a bare string does not parse

### spec.rules[].overrides.randomSteering

`CloudflareLoadBalancerRandomSteering`

Pool-weighting override for random / least-* steering policies.

### spec.rules[].overrides.randomSteering.defaultWeight

`double`

Default weight for pools not present in pool_weights (0.0–1.0). Leave 0 to
use the Cloudflare default (1).

- rule: default_weight must be between 0 and 1

### spec.rules[].overrides.randomSteering.poolWeights

`map<string, double>`

Per-pool weight overrides, keyed by pool ID. Each weight is relative to the
other pools in this load balancer.

### spec.rules[].overrides.regionPools

`[]CloudflareLoadBalancerGeoPools`

Region-code -> ordered pool list override.

### spec.rules[].overrides.regionPools[].code

`string` · required

The geo code this mapping applies to: a region code (e.g. "WNAM"), an ISO
country code (e.g. "US"), or a Cloudflare PoP code (e.g. "LAX"), depending on
which map this entry belongs to.

- rule: {"required":true}

### spec.rules[].overrides.regionPools[].poolIds

`[]string | valueFrom` · required

Ordered pools (by failover priority) for this code. Each is a literal pool ID
or a reference to a CloudflareLoadBalancerPool.

- references: CloudflareLoadBalancerPool (`status.outputs.pool_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareLoadBalancerPool, name: <that resource's name>, fieldPath: status.outputs.pool_id}} -- a bare string does not parse

### spec.rules[].overrides.sessionAffinity

`enum` · optional (explicit presence)

Session-affinity mode override. UNSET inherits the load balancer's mode;
an explicit `none` switches affinity OFF for matched traffic — the two
differ, which is why this field carries explicit presence.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `none` -- No session affinity (the default).
- `cookie` -- Cookie-based affinity.
- `ip_cookie` -- Cookie-based affinity with stable initial selection by client IP.
- `header` -- Header-based affinity (see session_affinity_attributes.headers).

### spec.rules[].overrides.sessionAffinityAttributes

`CloudflareLoadBalancerSessionAffinityAttributes`

Session-affinity attribute overrides (drain, headers, cookie flags,
zero-downtime failover).

- rule: samesite "None" cannot be combined with secure "Never"

### spec.rules[].overrides.sessionAffinityAttributes.drainDuration

`double`

Drain duration (seconds) applied when affinity is enabled.

- rule: drain_duration must be 0 or a positive number of seconds

### spec.rules[].overrides.sessionAffinityAttributes.headers

`[]string`

HTTP header names that header-based affinity is keyed on (required when
session_affinity is header). Use "cookie:<name>,<name>" to scope to cookies.

### spec.rules[].overrides.sessionAffinityAttributes.requireAllHeaders

`bool`

When header affinity is enabled, require ALL listed headers (true) or AT
LEAST ONE (false) for a session to be created.

### spec.rules[].overrides.sessionAffinityAttributes.samesite

`string`

SameSite attribute on the affinity cookie: one of "Auto", "Lax", "None",
"Strict". Empty uses Cloudflare's default ("Auto").

- rule: samesite must be one of "Auto", "Lax", "None", "Strict"

### spec.rules[].overrides.sessionAffinityAttributes.secure

`string`

Secure attribute on the affinity cookie: one of "Auto", "Always", "Never".
Empty uses Cloudflare's default ("Auto").

- rule: secure must be one of "Auto", "Always", "Never"

### spec.rules[].overrides.sessionAffinityAttributes.zeroDowntimeFailover

`string`

Zero-downtime failover behavior for pinned sessions: one of "none",
"temporary", "sticky". Empty uses Cloudflare's default ("none").

- rule: zero_downtime_failover must be one of "none", "temporary", "sticky"

### spec.rules[].overrides.sessionAffinityTtl

`double`

Session-affinity TTL override (seconds). Leave 0 to inherit.

- rule: session_affinity_ttl must be 0 (inherit) or a positive number of seconds

### spec.rules[].overrides.steeringPolicy

`enum` · optional (explicit presence)

Steering-policy override. UNSET inherits the load balancer's policy; an
explicit `off` forces static failover over default_pools for matched
traffic — the two differ, which is why this field carries explicit
presence.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `off` -- Use default_pools as a static failover priority list (also the default).
- `geo` -- Use region_pools / country_pools / pop_pools (geo-routing).
- `random` -- Select a pool at random, weighted by random_steering.
- `dynamic_latency` -- Use round-trip time to pick the closest healthy pool in default_pools.
- `proximity` -- Use pool latitude/longitude to pick the closest pool.
- `least_outstanding_requests` -- Pick the pool with the fewest outstanding requests (weighted).
- `least_connections` -- Pick the pool with the fewest open connections (weighted).

### spec.rules[].overrides.ttl

`double`

DNS TTL override (seconds; gray-clouded load balancers only). Leave 0 to
inherit.

- rule: ttl must be 0 (inherit) or a positive number of seconds

### spec.networks

`[]string`

Networks where this load balancer is enabled. Used with Cloudflare private
networking (e.g. Magic WAN / WARP connector networks); leave empty for
ordinary public load balancers.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareLoadBalancer, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.load_balancer_id` | `string` | Unique identifier of the Cloudflare load balancer. |
| `status.outputs.load_balancer_dns_record_name` | `string` | The hostname DNS record associated with the load balancer. |
| `status.outputs.load_balancer_cname_target` | `string` | The canonical CNAME target that the hostname resolves to (Cloudflare endpoint). |
| `status.outputs.zone_id` | `string` | The Cloudflare zone that owns the load balancer. A zone-scoped resource publishes its parent scope: the load balancer's API identity (and its Terraform import ID) is compound -- zones/{zone_id}/load_balancers/{id}. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |
| `spec.defaultPools` | CloudflareLoadBalancerPool | `status.outputs.pool_id` |
| `spec.fallbackPool` | CloudflareLoadBalancerPool | `status.outputs.pool_id` |
| `spec.regionPools[].poolIds` | CloudflareLoadBalancerPool | `status.outputs.pool_id` |
| `spec.countryPools[].poolIds` | CloudflareLoadBalancerPool | `status.outputs.pool_id` |
| `spec.popPools[].poolIds` | CloudflareLoadBalancerPool | `status.outputs.pool_id` |
| `spec.rules[].overrides.countryPools[].poolIds` | CloudflareLoadBalancerPool | `status.outputs.pool_id` |
| `spec.rules[].overrides.defaultPools` | CloudflareLoadBalancerPool | `status.outputs.pool_id` |
| `spec.rules[].overrides.fallbackPool` | CloudflareLoadBalancerPool | `status.outputs.pool_id` |
| `spec.rules[].overrides.popPools[].poolIds` | CloudflareLoadBalancerPool | `status.outputs.pool_id` |
| `spec.rules[].overrides.regionPools[].poolIds` | CloudflareLoadBalancerPool | `status.outputs.pool_id` |

## See Also

- [Overview](../README.md)
