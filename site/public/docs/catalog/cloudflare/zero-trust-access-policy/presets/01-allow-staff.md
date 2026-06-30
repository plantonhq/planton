---
title: "Preset: Allow staff"
description: "A simple `allow` policy that grants access to anyone with a corporate email domain, connecting from an allowed country, with a 24-hour session."
type: "preset"
rank: "01"
presetSlug: "01-allow-staff"
componentSlug: "zero-trust-access-policy"
componentTitle: "Zero Trust Access Policy"
provider: "cloudflare"
icon: "package"
order: 1
---

# Preset: Allow staff

A simple `allow` policy that grants access to anyone with a corporate email domain,
connecting from an allowed country, with a 24-hour session.

## When to use

- A baseline "staff can access this" policy attached to one or more applications.

## Key choices

- `decision: allow` with `include` (corporate domain) and `require` (country).
- `sessionDuration`: how long before re-authentication.

## Placeholders

| Placeholder | Description |
|---|---|
| `REPLACE_WITH_ACCOUNT_ID` | 32-character Cloudflare account ID |

## Referencing it from an application

```yaml
policies:
  - policy:
      valueFrom:
        kind: CloudflareZeroTrustAccessPolicy
        name: allow-staff
        fieldPath: status.outputs.policy_id
    precedence: 1
```
