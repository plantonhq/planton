# Cloudflare Zero Trust List

## Overview

`CloudflareZeroTrustList` is a reusable named set of values -- domains, IPs, URLs, emails, serial numbers, and kin -- that Gateway policies and device-posture rules reference by ID instead of repeating the values inline. Centralizing the values lets many policies share one source of truth that evolves in one place.

Lists are account-scoped. The list **type is immutable** at Cloudflare: changing it replaces the list (new ID) and breaks every policy that referenced the old one. Items are a set: order is not significant and is not preserved.

## Key Features

- **Nine types** -- `SERIAL`, `URL`, `DOMAIN`, `EMAIL`, `IP`, `CATEGORY`, `LOCATION`, `DEVICE`, `AAGUID` (use the uppercase form; lowercase would round-trip as permanent drift)
- **Set semantics** -- two lists with the same values in a different order are the same list
- **Required item values** -- the API tolerates value-less items; this spec rejects them (an entry without a value matches nothing)
- **URL-type caution** -- the API normalizes URLs and the provider does not, so URL lists produce a perpetual plan diff at provider v5.23.0

## Use Cases

**Ideal for:**

- A shared blocklist of domains many Gateway DNS policies reference
- An allowlist of corporate CIDRs for HTTP or L4 policies
- Device serials or AAGUIDs that posture rules match

**Not ideal for:**

- The policy that acts on the list -- that is `CloudflareZeroTrustGatewayPolicy`
- Cloudflare's older account-level Lists used by Rulesets (`CloudflareList` / `CloudflareListItem`) -- a different API family

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | 32-character hex. |
| `name` | string | Yes | Display name. |
| `type` | string | Yes | Immutable. One of `SERIAL`, `URL`, `DOMAIN`, `EMAIL`, `IP`, `CATEGORY`, `LOCATION`, `DEVICE`, `AAGUID`. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Purpose. |
| `items` | list of `{value, description}` | Set of entries. `value` is required. |

### Outputs

| Field | Description |
|-------|-------------|
| `list_id` | UUID referenced by Gateway policies and posture rules |

## Example Manifests

Domain blocklist:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustList
metadata:
  name: blocked-domains
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: blocked-domains
  type: DOMAIN
  description: Domains Gateway policies block for all users
  items:
    - value: gambling.example.com
    - value: casino.example.net
```

## Destroy Semantics

Destroy is a real delete. Policies that referenced `list_id` start failing their list lookup -- update or delete those policies first. Changing `type` is also a replace.

## Related Resources

- **CloudflareZeroTrustGatewayPolicy** -- references this list from `traffic` / `identity` expressions
- **CloudflareList** -- the older Ruleset-family list; do not mix the two

## Further Reading

For operational judgment -- type immutability, set semantics, and the URL-normalization drift -- see GUIDE.md.

## References

- [Cloudflare Zero Trust lists](https://developers.cloudflare.com/cloudflare-one/policies/gateway/lists/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
