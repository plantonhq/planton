---
title: "CloudflareRuleset — Technical Deep Dive"
description: "CloudflareRuleset — Technical Deep Dive deployment documentation"
icon: "package"
order: 100
componentName: "cloudflareruleset"
---

# CloudflareRuleset — Technical Deep Dive

## What Are Cloudflare Rulesets?

Rulesets are the unified rule engine powering Cloudflare's security, performance, and traffic management features. Every Cloudflare feature that evaluates rules — WAF, Cache Rules, Origin Rules, Redirect Rules, Transform Rules — uses the same underlying Ruleset API. Each ruleset executes during a specific **phase** of Cloudflare's HTTP request/response pipeline.

A ruleset contains an ordered list of **rules**. Each rule has:
- A **wirefilter expression** that determines when the rule matches
- An **action** to perform when the expression evaluates to true
- Optional **action parameters** that configure the action's behavior

## Cloudflare's Request Processing Pipeline

```
Client Request
    │
    ├── ddos_l4 (L4 DDoS protection)
    ├── ddos_l7 (L7 DDoS protection)
    ├── http_request_sanitize (request sanitization)
    ├── http_request_transform (URL rewrites, header modifications)
    ├── http_request_origin (origin override — Origin Rules)
    ├── http_request_cache_settings (cache configuration)
    ├── http_config_settings (config: HTTPS rewrites, BIC, etc.)
    ├── http_request_firewall_custom (WAF custom rules)
    ├── http_request_sbfm (Super Bot Fight Mode)
    ├── http_request_firewall_managed (WAF managed rulesets — OWASP, etc.)
    ├── http_ratelimit (rate limiting)
    ├── http_request_redirect / http_request_dynamic_redirect (redirects)
    ├── http_request_late_transform (late-stage transforms)
    │
    │   ← Origin fetch happens here →
    │
    ├── http_response_firewall_managed (response-phase WAF)
    ├── http_response_headers_transform (response header modification)
    ├── http_response_compression (compression rules)
    ├── http_response_cache_settings (response cache settings)
    ├── http_custom_errors (custom error pages)
    ├── http_log_custom_fields (custom log fields)
    │
Client Response
```

## Phase → Action Compatibility

Not all actions are valid in all phases. Cloudflare enforces this at the API level.

| Phase | Valid Actions |
|-------|-------------|
| `http_request_origin` | `route` |
| `http_request_firewall_custom` | `block`, `challenge`, `js_challenge`, `managed_challenge`, `skip`, `log`, `score` |
| `http_request_firewall_managed` | `execute`, `skip` |
| `http_request_cache_settings` | `set_cache_settings` |
| `http_request_dynamic_redirect` | `redirect` |
| `http_request_redirect` | `redirect` |
| `http_request_transform` | `rewrite` |
| `http_request_late_transform` | `rewrite` |
| `http_response_headers_transform` | `rewrite` |
| `http_ratelimit` | `block`, `challenge`, `js_challenge`, `managed_challenge`, `log` |
| `http_config_settings` | `set_config` |
| `http_custom_errors` | `serve_error` |
| `http_response_compression` | `compress_response` |

## Planton Design Decisions

### 80/20 Field Selection

The `cloudflare_ruleset` Terraform resource has over 60 action parameter fields. Many are niche Cloudflare features (autominify, BIC, polish, rocket loader, mirage, etc.) that apply only to specific phases and are rarely configured via IaC — they're more commonly toggled in the dashboard.

This component models the fields needed for the most common IaC-managed ruleset types:

| Category | Included | Excluded |
|----------|----------|----------|
| **Origin Rules** | host_header, origin, sni | — |
| **Security** | block response, skip phases/products, execute overrides | exposed_credential_check, ratelimit (separate from phase) |
| **Cache** | cache, edge_ttl, browser_ttl, serve_stale | cache_key, cache_reserve, origin_cache_control |
| **Transforms** | uri rewrite, headers | — |
| **Redirects** | from_value with target_url; from_list (Bulk Redirect, by list name → CloudflareList) | — |
| **Config** | — | autominify, bic, email_obfuscation, fonts, hotlink_protection, mirage, opportunistic_encryption, polish, rocket_loader, security_level, server_side_excludes, ssl, sxg |

### Flat Action Parameters (Matching Provider API)

The Cloudflare API uses a single flat `action_parameters` object where all fields coexist. Different actions use different subsets of fields. We mirror this in the proto rather than using a `oneof` per action type because:

1. **Direct mapping** — IaC modules (Pulumi/Terraform) can map fields 1:1 without switch statements
2. **Forward compatibility** — New action types can reuse existing parameter fields
3. **Simplicity** — No complex dispatching logic in serialization/deserialization

### `ruleset_kind` Naming

Cloudflare's ruleset `kind` field (`zone`, `custom`, `managed`, `root`) conflicts with the KRM envelope's `kind` field (`CloudflareRuleset`). We use `ruleset_kind` in the spec to avoid ambiguity.

### Zone vs Account Scope

Most rulesets are zone-scoped (apply to a single domain). Account-scoped rulesets (`root` kind) apply across all zones and are used for organization-wide policies. The CEL validation enforces that exactly one of `zone_id` or `account_id` is set.

### No Phase-Action Validation in Proto

We intentionally do not enforce phase → action compatibility in CEL validations. The matrix is complex, Cloudflare occasionally adds new phase-action combinations, and the Cloudflare API already provides clear error messages when invalid combinations are used. The docs and presets guide users toward valid combinations.

## Deployment Methods

### Planton CLI

```bash
# Preview changes
planton preview -f ruleset.yaml

# Apply
planton apply -f ruleset.yaml

# Destroy
planton destroy -f ruleset.yaml
```

### Pulumi (via Planton)

The Pulumi module at `iac/pulumi/` maps `CloudflareRulesetStackInput` to `cloudflare.NewRuleset()` from the `pulumi-cloudflare/sdk/v6` package. The module:

1. Loads the stack input via `stackinput.LoadStackInput()`
2. Creates a Cloudflare provider via `pulumicloudflareprovider.Get()`
3. Maps proto rules to `cloudflare.RulesetRuleArray`
4. Creates the ruleset and exports outputs

### Terraform (via Planton)

The Terraform module at `iac/tf/` uses the `cloudflare_ruleset` resource. The `variables.tf` mirrors the spec structure, and `main.tf` uses dynamic blocks for rules.

## Wirefilter Expression Language

Cloudflare's expression language is based on Wireshark's display filter syntax.

### Field Reference

| Field | Type | Description |
|-------|------|-------------|
| `http.host` | String | Request hostname |
| `http.request.uri` | String | Full URI including query string |
| `http.request.uri.path` | String | URI path only |
| `http.request.uri.query` | String | Query string only |
| `http.request.method` | String | HTTP method (GET, POST, etc.) |
| `http.request.version` | String | HTTP version |
| `http.referer` | String | Referer header |
| `http.user_agent` | String | User-Agent header |
| `http.cookie` | String | Cookie header |
| `ip.src` | IP | Client IP address |
| `ip.geoip.country` | String | Two-letter country code |
| `ip.geoip.continent` | String | Continent code |
| `cf.threat_score` | Number | Cloudflare threat score (0-100) |
| `cf.bot_management.score` | Number | Bot score (0-99, lower = more likely bot) |
| `ssl` | Boolean | Whether request uses HTTPS |

### Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `eq` | Equals | `http.host eq "example.com"` |
| `ne` | Not equals | `ip.src ne 192.0.2.1` |
| `starts_with` | String prefix | `http.request.uri.path starts_with "/api"` |
| `ends_with` | String suffix | `http.request.uri.path ends_with ".json"` |
| `contains` | String contains | `http.user_agent contains "bot"` |
| `in` | Set membership | `ip.src in {10.0.0.0/8 172.16.0.0/12}` |
| `gt`, `ge`, `lt`, `le` | Numeric comparison | `cf.threat_score gt 50` |
| `not` | Negation | `not http.request.uri.path starts_with "/static"` |
| `and`, `or` | Logical operators | `http.host eq "a.com" and ip.src ne 1.1.1.1` |

### Functions

| Function | Description |
|----------|-------------|
| `concat(a, b)` | Concatenate strings |
| `regex_replace(field, pattern, replacement)` | Regex replacement |
| `lower(field)` | Lowercase string |
| `upper(field)` | Uppercase string |
| `len(field)` | String length |
| `lookup_json_string(field, key)` | Extract JSON value |

## Important Constraints

1. **One custom ruleset per phase per zone**: Cloudflare allows only one zone-level custom ruleset per phase. All rules for a phase must be in a single ruleset resource.

2. **Rule ordering**: Rules are evaluated in the order they appear in the `rules` list. The first matching rule's action is applied (for most phases).

3. **Expression limit**: Expressions have a maximum length of 4096 characters.

4. **Rule limit**: Zone-level rulesets can contain up to 200 rules (varies by Cloudflare plan).

5. **Proxy required**: Origin Rules and most request-phase features require the DNS record to be proxied through Cloudflare (orange cloud enabled).

6. **`ref` stability**: The `ref` field provides a stable identity for rules. Without it, Terraform may destroy and recreate rules when their position in the list changes, causing brief downtime.
