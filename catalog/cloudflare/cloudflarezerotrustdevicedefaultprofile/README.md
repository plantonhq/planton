# Cloudflare Zero Trust Device Default Profile

## Overview

`CloudflareZeroTrustDeviceDefaultProfile` manages the account's default WARP device profile: the settings every enrolled device receives unless a custom profile matches it first. The profile always exists on the account -- applying this spec edits it in place, and destroy is a no-op that leaves the last-applied values standing. Two companion surfaces fold in: the local-DNS fallback list (a full-replacement write path of its own) and the per-zone WARP client certificate provisioning toggle.

## Key Features

- **The full WARP settings body** -- mode switching, update notifications, lock-in (`allowed_to_leave`, `switch_locked`), reconnection and captive-portal timers, LAN access windows, tunnel protocol
- **Split tunnel** -- exclude-mode or include-mode route lists (CIDRs and hostnames), mutually exclusive by validation
- **Virtual network scoping** -- which Zero Trust virtual networks devices may reach, and the default landing network (references `CloudflareZeroTrustTunnelVirtualNetwork`)
- **Local-DNS fallback list** -- the folded full-replacement companion; what you declare is exactly what exists after apply
- **Per-zone client certificates** -- the folded zone-scoped toggle letting origins verify traffic came through WARP

## Use Cases

**Ideal for:**

- The company-wide WARP baseline: lock the switch, pin the tunnel protocol, route internal domains to on-prem resolvers
- Split-tunnel hygiene: keep lab networks and printer discovery local while everything else tunnels

**Not ideal for:**

- Group-specific settings -- those are `CloudflareZeroTrustDeviceCustomProfile` (lower precedence wins)
- Device health checks -- those are `CloudflareZeroTrustDevicePostureRule`

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | The Cloudflare account (32-hex). |

### Key Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `exclude` / `include` | list | Split-tunnel routes (address XOR host per entry); the two lists are mutually exclusive. |
| `service_mode_v2` | object | The client mode (warp, proxy + `port`). |
| `virtual_networks` | object | `allowed` (vnet references, min 1) + `default_virtual_network_id`. |
| `dns_search_suffixes` | list | Suffixes appended when resolving short hostnames. |
| `fallback_domains` | list | The folded FULL-REPLACEMENT local-DNS fallback list. |
| `zone_certificates` | object | The folded per-zone certificate provisioning toggle (`zone_id` reference + explicit `enabled`). |
| `switch_locked`, `allowed_to_leave`, ... | bool | The toggle body; unset never sends, keeping Cloudflare's default. |
| `lan_allow_minutes`, `lan_allow_subnet_size` | int | The LAN access window and subnet size. |

### Outputs

| Field | Description |
|-------|-------------|
| `account_id` | The account the profile was applied to (the singleton's identity) |
| `gateway_unique_id` | The Gateway-side identifier of the profile |
| `policy_id` | The profile's policy identifier |

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustDeviceDefaultProfile
metadata:
  name: default-device-profile
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  switch_locked: true
  allowed_to_leave: false
  exclude:
    - address: 192.0.2.0/24
      description: lab network stays local
  fallback_domains:
    - suffix: corp.internal
      dns_server:
        - 10.0.0.53
```

## Destroy Semantics

Destroy is a NO-OP on all three surfaces: the profile, the fallback list, and the certificate toggle keep their last-applied values. Removing the resource abandons management; it never reverts anything.

## Related Resources

- **CloudflareZeroTrustDeviceCustomProfile** -- targeted profiles that override this default for matched devices
- **CloudflareZeroTrustDevicePostureRule** -- the health checks policies demand from these devices
- **CloudflareZeroTrustTunnelVirtualNetwork** -- the networks `virtual_networks` scopes access to

## Further Reading

For operational judgment -- the account-wide blast radius, the fallback list's full-replacement semantics, the certificate toggle's no-delete reality -- see GUIDE.md.

## References

- [Cloudflare WARP device profiles](https://developers.cloudflare.com/cloudflare-one/connections/connect-devices/warp/configure-warp/device-profiles/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
