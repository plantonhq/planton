# Cloudflare Zero Trust Device Default Profile

Manages the account's default Cloudflare WARP device profile: the settings every enrolled device receives unless a custom profile matches it first. The profile always exists on the account — applying this resource edits the singleton in place, and destroy is a no-op that leaves the last-applied values standing. Two companion write paths fold in: the account's local-DNS fallback list (full-replacement) and the per-zone WARP client certificate toggle. Treat it like a settings panel, not an object with a lifecycle.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Default Device Profile** — a `cloudflare_zero_trust_device_default_profile`, a PATCH upsert of the account singleton. Unset spec fields are never sent, leaving the live value (or Cloudflare's default) untouched
- **Local-DNS Fallback List** — created only when `fallbackDomains` is declared, a `cloudflare_zero_trust_device_default_profile_local_domain_fallback` that replaces the account's whole list on every apply
- **Zone Certificate Toggle** — created only when `zoneCertificates` is declared, a `cloudflare_zero_trust_device_default_profile_certificates` enabling or disabling WARP client certificate provisioning on one zone

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with a Cloudflare API token carrying Account → Zero Trust → Edit. The zone-certificates fold additionally needs Zone → SSL and Certificates → Edit. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Zero Trust enabled** — the account must have completed Cloudflare Zero Trust onboarding (the team-name step).
- **The current fallback list fetched** (only for `fallbackDomains`) — Cloudflare seeds every account with a default list (localhost, home.arpa, and kin). Declaring the field replaces the whole list, so fold the seeded rows you want to keep into the manifest before the first apply.
- **The target zone in the same account** (only for `zoneCertificates`) — the zone whose client certificate devices should receive.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zero Trust Device Default Profile**, and click **Deploy**. The creation wizard walks you through the owning account, the WARP settings body (lock-in toggles, timers, service mode, LAN access), split-tunnel routes, the fallback domain list, and the per-zone certificate toggle. Start from the **Locked fleet baseline** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustDeviceDefaultProfile
metadata:
  name: default-device-profile
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  switchLocked: true
  allowedToLeave: false
  autoConnect: 600
```

```shell
planton apply -f default-profile.yaml
```

This patches the account's default profile so users cannot turn WARP off or unenroll, and a user-disabled client reconnects on its own after ten minutes. Every other setting keeps its live value. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire virtual network scoping and the certificate zone to resources managed by other Infra Components:

```yaml
spec:
  virtualNetworks:
    allowed:
      - valueFrom:
          kind: CloudflareZeroTrustTunnelVirtualNetwork
          name: prod-overlay
          fieldPath: status.outputs.virtual_network_id
    defaultVirtualNetworkId:
      valueFrom:
        kind: CloudflareZeroTrustTunnelVirtualNetwork
        name: prod-overlay
        fieldPath: status.outputs.virtual_network_id
  zoneCertificates:
    zoneId:
      valueFrom:
        kind: CloudflareDnsZone
        name: acme-com
        fieldPath: status.outputs.zone_id
    enabled: true
```

The InfraPipeline deploys the referenced virtual networks and zone first, then patches the profile against their real IDs.

## Key Configuration

These are the most important decisions when configuring the default device profile. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Destroy reverts nothing, on any surface** — the profile is a settings singleton (PATCH upsert, no delete), the fallback list's destroy is a no-op, and the certificate toggle has no delete at all. Removing this resource abandons the last-applied state; devices keep behaving as configured. To actually revert, apply the values you want.

**The fallback list is full-replacement — and Cloudflare pre-seeded it** — the moment `fallbackDomains` is declared, your list replaces the whole account list, including the seeded rows. Declaring one row does not add a row; it deletes everything else. Rows without `dnsServer` entries fall back to the system resolvers unless `disableAutoFallback` fails them closed. `dnsSearchSuffixes` is managed the same way — the declared list is exactly what exists, and an empty declaration clears the account's list.

**`switchLocked` and `allowedToLeave` are lock-in levers** — the first removes users' ability to turn WARP off, the second their ability to unenroll; together they hard-lock the fleet. Roll them out after split-tunnel and fallback routing are proven — a bad route with a locked switch is a helpdesk incident on every device at once.

**Split tunnel runs in one direction** — `exclude` (everything else tunnels) and `include` (everything else bypasses) are mutually exclusive modes, not filters to combine; the spec rejects a manifest declaring both. When switching modes, the semantics of every existing entry invert. `excludeOfficeIps` adds Microsoft 365 IPs to the exclude side automatically.

**The certificate toggle is zone-scoped and permanent** — `zoneCertificates` is the one zone-scoped surface on this account-scoped kind, and Cloudflare offers no delete or import for it. `enabled` is required explicitly (true or false) because turning it off IS the managed off state; removing the block leaves it however it last was. The endpoint additionally requires a user-actor credential — account-owned tokens are refused with 401 code 1039 (measured live).

**A blank `tunnelProtocol` is a reset, not a no-op** — an empty value resets the account to Cloudflare's default protocol. On accounts running a non-default protocol, declare it explicitly or an apply silently flips the fleet back.

**Unset fields never fight the dashboard** — every optional toggle and timer is only sent when set, so this resource can manage exactly the settings you declare and leave the rest to Cloudflare's defaults or manual configuration.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| CloudflareZeroTrustTunnelVirtualNetwork | `spec.virtualNetworks.allowed[]` | `status.outputs.virtual_network_id` |
| CloudflareZeroTrustTunnelVirtualNetwork | `spec.virtualNetworks.defaultVirtualNetworkId` | `status.outputs.virtual_network_id` |
| CloudflareDnsZone | `spec.zoneCertificates.zoneId` | `status.outputs.zone_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `account_id` | The account the profile was applied to — the singleton's identity | Correlating the profile with other account-scoped resources |
| `gateway_unique_id` | The Gateway-side identifier of the profile | Gateway policies that reference device profiles |
| `policy_id` | The profile's policy identifier from the device policy API | API automation against the device policy endpoints |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Locked fleet baseline** — the managed-fleet shape: users cannot turn WARP off, unenroll, or switch modes; a disabled client self-reconnects after ten minutes with the standard captive-portal grace for hotel and airport Wi-Fi. Start from the **Locked fleet baseline** preset.

**Split tunnel with on-prem DNS** — the hybrid-office shape: private LAN ranges and printer discovery stay off the tunnel, Microsoft 365 IPs are excluded automatically, short hostnames complete in the corp domain, and internal suffixes resolve against the on-prem resolver through the fallback list (with the seeded `localhost` and `home.arpa` rows re-declared deliberately). Start from the **Split tunnel with on-prem DNS** preset.

**Baseline plus overrides** — keep the default profile permissive and layer stricter custom profiles on specific identity groups, or invert it: a locked baseline with looser custom profiles for engineering. The default is what every device falls back to when no custom profile matches.

## Works With

- [**Cloudflare Zero Trust Device Custom Profile**](/infra-catalog/cloudflare-zero-trust-device-custom-profile) — group-specific overrides of this baseline; lowest precedence wins.
- [**Cloudflare Zero Trust Device Posture Rule**](/infra-catalog/cloudflare-zero-trust-device-posture-rule) — the health checks Access and Gateway policies demand from these devices.
- [**Cloudflare Zero Trust Tunnel Virtual Network**](/infra-catalog/cloudflare-zero-trust-tunnel-virtual-network) — the routing segments `virtualNetworks` scopes device access to.
- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) — the zone `zoneCertificates` provisions WARP client certificates from.
