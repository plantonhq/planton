# CloudflareRuleset Pulumi Module — Architecture Overview

## Resource Flow

```
CloudflareRulesetStackInput
  ├── target: CloudflareRuleset (KRM resource)
  │     ├── metadata.name → Pulumi resource name
  │     └── spec → RulesetArgs
  │           ├── zone_id/account_id → scope
  │           ├── ruleset_kind → kind
  │           ├── phase → phase
  │           ├── name → name
  │           └── rules[] → RulesetRuleArray
  │                 ├── expression, action, ref, enabled
  │                 └── action_parameters → RulesetRuleActionParametersArgs
  │                       ├── origin, host_header, sni (route)
  │                       ├── response (block)
  │                       ├── uri, headers (rewrite)
  │                       ├── from_value (redirect)
  │                       ├── phases, products, ruleset (skip)
  │                       ├── id, overrides (execute)
  │                       └── cache, edge_ttl, browser_ttl, serve_stale (cache)
  └── provider_config → cloudflare.Provider
```

## Design Decisions

### Enum-to-String Mapping

Proto enums (`Phase`, `RulesetKind`, `Action`) are mapped to their string representations via helper functions. The Cloudflare API and Pulumi SDK both use plain strings for these values. Proto's `.String()` method on generated enums returns the enum value name, which matches Cloudflare's expected strings (e.g., `http_request_origin`, `zone`, `route`).

### Flat Action Parameters

The `buildActionParameters()` function maps a single proto `CloudflareRulesetActionParameters` message to the Pulumi `RulesetRuleActionParametersArgs`. Both structures are flat — fields from all action types coexist. The function only sets fields that are non-zero, relying on the Cloudflare API to accept sparse parameter objects.

### Optional Fields

The `enabled` field on rules is `optional bool` with a proto-level default of `true`. The `GetEnabled()` getter returns `true` when the field is unset, which is the correct Cloudflare behavior.

## Stack Outputs

| Constant | Pulumi Export | Source |
|----------|--------------|--------|
| `OpRulesetId` | `ruleset_id` | `created.ID()` |
| `OpVersion` | `version` | `created.Version` |
| `OpZoneId` | `zone_id` | `spec.ZoneId.GetValue()` (pass-through) |
| `OpPhase` | `phase` | `phaseString(spec.Phase)` (pass-through) |
