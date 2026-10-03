# Cloudflare IP Access Rule

## Overview

`CloudflareIpAccessRule` is one allow, block, or challenge decision applied to traffic matching an IP address, IP range, ASN, or country -- before the request reaches the zone's other security products. Rules live either on the whole account (every zone) or on a single zone. Exactly one scope must be set.

Only `mode` and `notes` can change in place. Cloudflare's API does not honor editing the matched target or value of an existing rule even though it accepts the request shape. To change what a rule matches, delete and recreate it.

## Key Features

- **Dual scope** -- account-wide (`account_id`) or single-zone (`zone_id`); exactly one, never both (the provider would silently prefer the account)
- **Five targets** -- `ip`, `ip6` (fully-expanded long form), `ip_range` (IPv4 `/16` or `/24`; IPv6 `/32`, `/48`, or `/64`), `asn` (`AS13335`), `country` (ISO 3166-1 alpha-2, or `T1` for Tor)
- **Five modes** -- `block`, `challenge`, `whitelist`, `js_challenge`, `managed_challenge`
- **In-place updates are narrow** -- only `mode` and `notes` stick; a configuration change plans as an update that does not persist

## Use Cases

**Ideal for:**

- Blocking a known-bad address across every zone on the account
- Challenging a country (or Tor exit nodes) on one zone
- Allowlisting an office egress IP so it bypasses other security features

**Not ideal for:**

- Expression-based WAF or custom rules -- that is `CloudflareRuleset`
- Bot scoring and fight-mode toggles -- that is `CloudflareBotManagement`

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` XOR `zone_id` | string / StringValueOrRef | Yes | Exactly one. `account_id` is 32-character hex. `zone_id` can reference a `CloudflareDnsZone` via `value_from` (defaults to `status.outputs.zone_id`). |
| `mode` | string | Yes | `block`, `challenge`, `whitelist`, `js_challenge`, or `managed_challenge`. |
| `configuration` | object | Yes | `{target, value}` -- what the rule matches. Changing this on an existing rule does not stick. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `notes` | string | Why the rule exists (shown in the dashboard). Updates in place. |

### Outputs

| Field | Description |
|-------|-------------|
| `rule_id` | The created rule's ID |
| `zone_id` | The zone the rule was created on (empty for account-wide rules) |
| `account_id` | The account the rule was created on (empty for zone-scoped rules) |

## Example Manifests

Account-wide block of a documentation-range IPv4 address:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareIpAccessRule
metadata:
  name: block-scanner
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  mode: block
  configuration:
    target: ip
    value: "192.0.2.10"
  notes: Block a known scanner
```

## Destroy Semantics

Destroy is a real delete. The rule stops matching immediately. Changing `configuration.target` or `configuration.value` is not a replace -- it plans an in-place update that Cloudflare ignores, so the old match keeps serving. Recreate the rule (new identity) to change what it matches.

## Related Resources

- **CloudflareBotManagement** -- zone-wide bot scoring; this kind is a static allow/block/challenge on a selector
- **CloudflareRuleset** -- expression-based WAF and custom rules, a different API family
- **CloudflareDnsZone** -- `zone_id` foreign key for a zone-scoped rule

## Further Reading

For operational judgment -- dual scope, the configuration-does-not-stick trap, and IPv6/CIDR shape -- see GUIDE.md.

## References

- [Cloudflare IP Access rules](https://developers.cloudflare.com/waf/tools/ip-access-rules/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
