# Cloudflare Zero Trust Device Custom Profile

## Overview

`CloudflareZeroTrustDeviceCustomProfile` creates a targeted WARP device profile: the same settings body as the account's default profile, applied only to the devices matched by a wirefilter expression. An account can carry many custom profiles; the lowest precedence value wins when several match. Unlike the default profile, a custom profile is a real object -- create, update, and delete all do what they say. The per-profile local-DNS fallback list folds in as its own full-replacement write path.

## Key Features

- **Wirefilter targeting** -- match on identity (email, groups, service token, SAML attributes), network, or OS name/version
- **Explicit ordering** -- required `precedence` (lower wins) keeps multi-profile accounts predictable
- **The full WARP settings body** -- split tunnel, service mode, virtual network scoping, DNS search suffixes, lock-in toggles, LAN access windows
- **Per-profile fallback list** -- the folded full-replacement companion; rows ride the profile and retire with it

## Use Cases

**Ideal for:**

- Group-specific overrides: contractors on a tighter split tunnel, developers with LAN access, executives on masque
- OS-specific handling: a Windows-only SCCM boundary profile, an older-macOS grace profile

**Not ideal for:**

- The account-wide baseline -- that is `CloudflareZeroTrustDeviceDefaultProfile`
- Device health checks -- those are `CloudflareZeroTrustDevicePostureRule`

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | The Cloudflare account (32-hex). |
| `name` | string | Yes | The profile's name. |
| `match` | string | Yes | The wirefilter expression selecting devices. |
| `precedence` | int | Yes | Evaluation order; LOWER wins. Required here although the provider marks it optional -- the API rejects a create without it. |

### Key Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `enabled` | bool | Whether the profile applies (default on). |
| `exclude` / `include` | list | Split-tunnel routes (address XOR host per entry); mutually exclusive lists. |
| `service_mode_v2` | object | The client mode (warp, proxy + `port`). |
| `virtual_networks` | object | `allowed` (vnet references, min 1) + `default_virtual_network_id`. |
| `fallback_domains` | list | The folded FULL-REPLACEMENT per-profile fallback list. |
| `lan_allow_minutes`, `lan_allow_subnet_size` | int | The LAN access window. |

### Outputs

| Field | Description |
|-------|-------------|
| `policy_id` | The Cloudflare-assigned profile identifier |
| `gateway_unique_id` | The Gateway-side identifier of the profile |

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustDeviceCustomProfile
metadata:
  name: contractor-devices
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: contractor-devices
  match: identity.groups.name == "contractors"
  precedence: 100
  switch_locked: true
  exclude:
    - address: 192.0.2.0/24
```

## Destroy Semantics

Destroy is a real delete: matched devices fall back to the default profile, and the per-profile fallback rows retire with the profile.

## Related Resources

- **CloudflareZeroTrustDeviceDefaultProfile** -- the baseline this profile overrides for matched devices
- **CloudflareZeroTrustDevicePostureRule** -- the health checks policies demand from these devices
- **CloudflareZeroTrustTunnelVirtualNetwork** -- the networks `virtual_networks` scopes access to

## Further Reading

For operational judgment -- precedence discipline, match-expression pitfalls, the fallback list's replacement semantics -- see GUIDE.md.

## References

- [Cloudflare WARP device profiles](https://developers.cloudflare.com/cloudflare-one/connections/connect-devices/warp/configure-warp/device-profiles/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
