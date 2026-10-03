# Cloudflare Zero Trust Gateway Policy

## Overview

`CloudflareZeroTrustGatewayPolicy` is one Gateway rule: a filter expression over employee traffic (DNS queries, HTTP requests, or network connections) plus the action to take when it matches -- block, allow, isolate, redirect, override, resolve, and kin.

Two Cloudflare behaviors deserve loud warnings. `enabled` defaults to **false**: a policy authored without `enabled: true` deploys disabled and filters nothing. Wirefilter expressions (`traffic`, `identity`, `device_posture`) are reformatted by the API before storing -- an expression that round-trips differently shows as permanent plan drift; adopt the API-formatted form when that happens.

The module always sends `rule_settings` (an empty object when you configure nothing). That is the provider's own workaround for API drift. Settings that carry `add_headers` or `override_ips` still drift on first apply at provider v5.23.0 -- a known upstream defect.

## Key Features

- **Sixteen actions** (`allow`, `block`, `isolate`, `redirect`, `override`, `l4_override`, `egress`, `resolve`, `quarantine`, `scan`, and kin)
- **Singular `filter`** -- `dns`, `http`, `l4`, `egress`, or `dns_resolver`; the module wraps it to the provider's one-element list
- **Full `rule_settings` tree** -- block page, session check, DNS resolvers (with a virtual-network foreign key), browser-isolation controls, quarantine file types, redirect, headers, and more
- **Explicit precedence** -- lower runs earlier; omit it and Cloudflare assigns one
- **Schedule and expiration** -- day-of-week windows and an RFC3339 expiry

## Use Cases

**Ideal for:**

- Blocking DNS categories or specific domains for every WARP-enrolled device
- HTTP allow/block with a custom block page and a session-duration check
- Overriding DNS answers or forcing SafeSearch

**Not ideal for:**

- Website WAF / firewall rules -- that is `CloudflareRuleset`
- Reusable domain/IP sets -- put those in `CloudflareZeroTrustList` and reference the list from `traffic`
- Browser Isolation or dedicated egress on an account that lacks those add-ons -- the apply fails rather than upgrading the plan

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | 32-character hex. Gateway policies are account-scoped. |
| `name` | string | Yes | Display name. |
| `action` | string | Yes | One of `on`, `off`, `allow`, `block`, `scan`, `noscan`, `safesearch`, `ytrestricted`, `isolate`, `noisolate`, `override`, `l4_override`, `egress`, `resolve`, `quarantine`, `redirect`. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `filter` | string | `http`, `dns`, `l4`, `egress`, or `dns_resolver`. Empty lets Cloudflare infer from the action. |
| `enabled` | optional bool | **Cloudflare defaults this to false.** Set `true` or the policy is inert. |
| `precedence` | optional int64 | Lower evaluates earlier. Unset = Cloudflare assigns one. |
| `traffic` / `identity` / `device_posture` | string | Wirefilter expressions. Absent means match-all for that dimension. |
| `description` | string | Purpose. |
| `expiration` | message | `expires_at` (RFC3339, required in-message) and optional `duration` (≥5). |
| `schedule` | message | `mon`..`sun` time-interval strings plus `time_zone`. |
| `rule_settings` | message | Action-specific settings. Always emitted (empty object when unset). `dns_resolvers.ipv4[].vnet_id` / `ipv6[].vnet_id` reference a `CloudflareZeroTrustTunnelVirtualNetwork`. |

### Outputs

| Field | Description |
|-------|-------------|
| `policy_id` | UUID of the created policy |
| `precedence` | Evaluation order (Cloudflare-assigned when the spec left it unset) |

## Example Manifests

DNS block:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustGatewayPolicy
metadata:
  name: block-gambling
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: block-gambling-domains
  action: block
  filter: dns
  enabled: true
  precedence: 1000
  traffic: any(dns.domains[*] == "gambling.example.com")
  rule_settings:
    block_page_enabled: true
    block_reason: Blocked by company policy
```

## Destroy Semantics

Destroy is a real delete. The schema carries a computed `deleted_at` the provider never reads; after destroy the API may 404 or return a tombstone. Disable (`enabled: false`) when you need a reversible off-switch.

## Related Resources

- **CloudflareZeroTrustList** -- reusable domain/IP/email sets referenced from `traffic`
- **CloudflareZeroTrustTunnelVirtualNetwork** -- `rule_settings.dns_resolvers.*.vnet_id`
- **CloudflareZeroTrustAccessIdentityProvider** -- identities the `identity` expression matches

## Further Reading

For operational judgment -- the enabled-defaults-false trap, `rule_settings` drift, and entitlement-gated actions -- see GUIDE.md.

## References

- [Cloudflare Gateway policies](https://developers.cloudflare.com/cloudflare-one/policies/gateway/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
