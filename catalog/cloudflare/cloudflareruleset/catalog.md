# Cloudflare Ruleset

Deploys a Cloudflare Ruleset: an ordered collection of rules evaluated during one phase of Cloudflare's HTTP request pipeline — the unified engine behind WAF custom and managed rules, rate limiting, cache rules, origin rules, redirects, transforms, and configuration rules, so one kind covers every phase. Each rule pairs a wirefilter match expression with an action (block, challenge, execute, redirect, rewrite, route, set cache settings, skip, and more) and that action's parameters. Rulesets live at zone scope (most common) or account scope — exactly one — and Cloudflare allows one custom ruleset per scope-and-phase pair.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Ruleset** -- one `cloudflare_ruleset` at zone or account scope, bound to one processing phase (for example `http_request_firewall_custom` for WAF custom rules, `http_ratelimit` for rate limiting, `http_request_dynamic_redirect` for redirects) and carrying the entire ordered rule list — rate-limiting rules with their counting characteristics, WAF rules with optional exposed-credential checks and per-rule logging overrides

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that can edit the relevant zone or account rulesets. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **A zone or an account scope** -- a Cloudflare zone (wire `zoneId` to a Cloudflare DNS Zone resource or pass a literal) for zone-level rulesets, or `accountId` for account-level `custom`/`root` rulesets. Exactly one of the two.
- **A plan carrying the phase** (only for paid phases) -- custom firewall rules (`http_request_firewall_custom`) work on the free plan; executing managed WAF rulesets (`http_request_firewall_managed`) and rate limiting (`http_ratelimit`) are paid-plan features. A deploy into a gated phase on a free-plan zone fails at the API, not at validation.
- **The backing List** (only for `fromList` redirects) -- a Bulk Redirect rule that resolves targets from a reusable list needs the Cloudflare List resource to exist first, and rules whose expressions use `$list_name` references need those lists too.

## Deploy

### Console

Open the deployment store, find **Cloudflare Ruleset**, and click **Deploy**. The creation wizard walks you through the scope, the phase, and an ordered rule builder — expression, action, and the action's parameters per rule. Start from the **Managed WAF — Cloudflare + OWASP Rulesets** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareRuleset
metadata:
  name: waf-custom-rules
  org: acme-corp
  env: prod
spec:
  zoneId:
    value: "023e105f4ecef8ad9ca31a8372d0c353"
  rulesetKind: zone
  phase: http_request_firewall_custom
  name: WAF custom rules
  rules:
    - ref: block-admin-outside-office
      expression: 'http.request.uri.path starts_with "/admin" and not ip.src in {203.0.113.0/24}'
      action: block
      description: Block the admin panel outside the office range
    - ref: challenge-no-user-agent
      expression: 'http.user_agent eq ""'
      action: managed_challenge
      description: Challenge requests with no user agent
```

```shell
planton apply -f ruleset.yaml
```

This creates a zone-level WAF custom ruleset (a free-plan phase) with two ordered rules: a block on the admin panel for traffic outside the office range, then a managed challenge for user-agent-less requests. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the zone — and, for Bulk Redirects, the backing list — to resources managed in the same InfraPipeline:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-com
      fieldPath: status.outputs.zone_id
  phase: http_request_firewall_custom
  name: WAF custom rules
```

The InfraPipeline resolves the dependency graph, deploys the zone (and any referenced lists) first, then provisions the ruleset against them.

## Key Configuration

These are the most important decisions when configuring a ruleset. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One ruleset per scope per phase.** Cloudflare allows exactly one custom ruleset per `{scope, phase}` pair. Deploying a second resource for the same zone and phase does not add rules — it fights the first for the same underlying ruleset. Keep each phase's rules in one manifest and let list order express precedence.

**The phase is the ruleset's identity.** `phase` decides when in the pipeline the rules run and which actions make sense there — `route` belongs to `http_request_origin`, `set_config` to `http_config_settings`, `set_cache_settings` to the cache phases. Not every action is valid in every phase; Cloudflare enforces the pairing at the API.

**Give rules a stable `ref`.** The per-rule `ref` is what stops a rule from being destroyed and recreated when its position in the list changes. Set it on every rule from the start — retrofitting refs after reorders is how rules churn.

**The skip grains, narrowest first.** The `skip` action has five surfaces: the `rules` map (ruleset ID to rule IDs) surgically exempts individual rules inside another ruleset — the right tool for managed-WAF false positives, because the rest of the ruleset keeps protecting; `ruleset` skips one whole ruleset (`"current"` skips the remainder of this one); `rulesets` skips several; `phases` and `products` are the bluntest grain and easy to over-exempt. Both IDs in the `rules` map are 32-character hex from the API and environment-specific — parameterize them per environment rather than hardcoding one environment's IDs into a shared manifest.

**Expressions fail at deploy, not at validation.** Rule expressions are Cloudflare's wirefilter language, parsed server-side — a typo deploys nothing, because the API rejects the whole ruleset. Test complicated expressions in the dashboard's expression editor first, and prefer several narrow rules over one giant expression: per-rule `enabled` flags then give you a kill switch per behavior.

**`rulesetKind` is almost always the default.** `zone` (the default) covers zone-phase rulesets; `custom` builds a reusable collection invoked from another ruleset via `execute`; `root` is the account-level entry point for a phase; `managed` is Cloudflare-maintained. Reach past `zone` only when you are deliberately building the account-level composition.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |
| **CloudflareList** | `rules[].actionParameters.fromList.name` | `status.outputs.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `ruleset_id` | The Cloudflare-assigned ruleset ID | `skip` parameters in other rulesets; `execute` invocations of a `custom` ruleset |
| `version` | The ruleset version, incrementing on each update | Change auditing and rollback investigation |
| `phase` | The phase the ruleset executes in | InfraChart conditionals that branch on the phase |
| `last_updated` | RFC3339 timestamp of the last change | Change auditing |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Managed WAF** -- Two `execute` rules in `http_request_firewall_managed`: the Cloudflare Managed Ruleset, then the OWASP Core Ruleset with a sensitivity override. Start from the **Managed WAF — Cloudflare + OWASP Rulesets** preset.

**Origin routing** -- `route` rules in `http_request_origin` that send application paths to one origin while the DNS-configured default serves everything else. Start from the **Origin Rule — Split Traffic Between Origins** preset.

**Cache policy** -- `set_cache_settings` rules with aggressive TTLs for static assets and explicit bypass for dynamic API endpoints; the advanced variant takes full control of the cache key and Cache Reserve. Start from the **Cache Settings — Static Assets + API Bypass** preset, or **Advanced cache key and Cache Reserve** for the cache-key work.

**Rate limiting** -- `http_ratelimit` rules counting requests per characteristics (buckets) over a period, applying the action when the threshold trips. Start from the **Rate limiting rules** preset.

**Bulk redirects from a list** -- An account-level redirect ruleset consulting a reusable Bulk Redirect list, with the source-to-target entries managed independently as list items. Start from the **Bulk Redirect — Redirect From a List** preset.

**Surgical WAF exemption** -- A `skip` rule whose `rules` map exempts one path from the specific managed-WAF rules that false-positive on its payloads, leaving every other protection in place. Start from the **Skip Specific Rules in Another Ruleset** preset.

## Works With

- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) -- the zone scope for zone-level rulesets; wire `zoneId` via ValueFromRef.
- [**Cloudflare List**](/infra-catalog/cloudflare-list) -- backs Bulk Redirect rules via `fromList` and `$list` references in expressions.
- [**Cloudflare List Item**](/infra-catalog/cloudflare-list-item) -- the individual source-to-target entries inside a Bulk Redirect list.
- [**Cloudflare Snippet Rules**](/infra-catalog/cloudflare-snippet-rules) -- the other expression table on the zone: same Rules language, routing requests to snippets instead of taking ruleset actions.
