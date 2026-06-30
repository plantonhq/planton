---
title: "Preset: Engineering team group"
description: "A reusable account-scoped group that matches your engineering staff by email domain, requires an allowed country, and excludes a known contractor account."
type: "preset"
rank: "01"
presetSlug: "01-engineering-team"
componentSlug: "zero-trust-access-group"
componentTitle: "Zero Trust Access Group"
provider: "cloudflare"
icon: "package"
order: 1
---

# Preset: Engineering team group

A reusable account-scoped group that matches your engineering staff by email domain,
requires an allowed country, and excludes a known contractor account.

## When to use

- You repeatedly grant the same team access across multiple applications.

## Key choices

- `include`: any matching rule adds the user (here, corporate email domains).
- `require`: every rule must also hold (here, an allowed country).
- `exclude`: any matching rule removes the user.

## Placeholders

| Placeholder | Description |
|---|---|
| `REPLACE_WITH_ACCOUNT_ID` | 32-character Cloudflare account ID |

## Referencing it from a policy

```yaml
include:
  - group:
      id:
        valueFrom:
          kind: CloudflareZeroTrustAccessGroup
          name: engineering-team
          fieldPath: status.outputs.group_id
```
