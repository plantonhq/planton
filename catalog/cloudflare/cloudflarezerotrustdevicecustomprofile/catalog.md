# Cloudflare Zero Trust Device Custom Profile

Deploys a targeted Cloudflare WARP device profile: the full device-settings body — split tunnel, service mode, lock-in toggles, LAN access windows, virtual network scoping, DNS search suffixes — applied only to the devices matched by a wirefilter expression. An account can carry many custom profiles; the lowest `precedence` value wins when several match, and a device matching none falls back to the account's default profile. The profile's local-DNS fallback list folds into this kind as a declarative, full-replacement companion.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Custom Device Profile** — a `cloudflare_zero_trust_device_custom_profile` carrying the wirefilter `match`, the required `precedence`, and every declared settings toggle. A real object: create, update, and delete all do what they say, and deleting it returns matched devices to the default profile
- **Per-Profile Fallback Domain List** — created only when `fallbackDomains` is declared. The profile resource reports its fallback list read-only; this dedicated companion is the only write path, and it replaces the whole list on every apply. Rows ride the profile and retire with it

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with a Cloudflare API token carrying Account → Zero Trust → Edit. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Zero Trust enabled** — the account must have completed Cloudflare Zero Trust onboarding (the team-name step).
- **A precedence plan** — every custom profile on the account competes in one precedence space. Know where this profile ranks before deploying it next to existing ones.
- **Virtual networks provisioned first** (only for `virtualNetworks`) — the Zero Trust virtual networks the profile scopes device access to must already exist.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zero Trust Device Custom Profile**, and click **Deploy**. The creation wizard walks you through the owning account, device targeting (the wirefilter `match` and the `precedence` rank), the WARP settings body (lock-in toggles, service mode, LAN access), split-tunnel routes, and the per-profile fallback domain list. Start from the **Contractor lockdown** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustDeviceCustomProfile
metadata:
  name: contractor-lockdown
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: contractor-lockdown
  match: identity.groups.name == "contractors"
  precedence: 100
  description: contractor laptops run locked-down full-tunnel WARP
  switchLocked: true
  allowedToLeave: false
  allowModeSwitch: false
```

```shell
planton apply -f custom-profile.yaml
```

This creates a profile that puts everyone in the contractors identity group on full-tunnel WARP they cannot switch off, switch modes on, or unenroll from. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire virtual network scoping to networks managed by other Infra Components:

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
```

The InfraPipeline deploys the referenced virtual networks first, then provisions the profile against their real IDs.

## Key Configuration

These are the most important decisions when configuring a custom device profile. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The match expression fails open to the default profile** — a device matching no custom profile gets the account default, so a typo in `match` does not strand devices; it silently baselines them. After changing a match, verify a known device actually landed on this profile in the Zero Trust dashboard rather than trusting the apply. Available selectors: `identity.email`, `identity.groups.id/name/email`, `identity.service_token_uuid`, `identity.saml_attributes`, `network`, `os.name`, `os.version`.

**Precedence is a fleet-wide ordering** — the LOWEST value wins, and every profile on the account competes in one space. The field is required here even though the provider marks it optional, because Cloudflare's API rejects a create without it. Leave gaps (100, 200, 300…) so a future profile can slot between two existing ones without renumbering; two profiles with adjacent values and overlapping matches is how a device lands in the wrong split tunnel.

**Split tunnel runs in one direction** — `exclude` (everything else tunnels) and `include` (everything else bypasses) are mutually exclusive at the API, and the spec rejects a manifest declaring both. Each route is a CIDR `address` or a domain `host`, never both.

**The fallback list is full-replacement** — the declared `fallbackDomains` list is exactly what exists after apply; there is no additive mode. Rows apply only to this profile's matched devices and retire with the profile, so there is no separate cleanup. A row without `dnsServer` entries falls back to the system resolvers unless `disableAutoFallback` fails it closed.

**Unset toggles keep Cloudflare's defaults** — every optional boolean (`switchLocked`, `allowedToLeave`, `allowModeSwitch`, `autoConnect`, …) is only sent when set, so an unset field never fights the dashboard. Lock-in postures are explicit choices: `switchLocked: true` blocks users from turning WARP off, `allowedToLeave: false` prevents unenrollment.

**Virtual network scoping is a containment rule** — `virtualNetworks.defaultVirtualNetworkId` must be one of the `allowed` entries; Cloudflare enforces the containment at the API. Unset leaves routing to the account's default virtual network.

**Deleting a profile is a routing change, not just a removal** — matched devices immediately fall back to the default profile or the next-lowest-precedence match. If this profile carved lab networks out of the tunnel, deleting it sends that traffic through the tunnel on every matched device at once.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| CloudflareZeroTrustTunnelVirtualNetwork | `spec.virtualNetworks.allowed[]` | `status.outputs.virtual_network_id` |
| CloudflareZeroTrustTunnelVirtualNetwork | `spec.virtualNetworks.defaultVirtualNetworkId` | `status.outputs.virtual_network_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `policy_id` | The Cloudflare-assigned profile identifier | Targeting the profile in API automation and imports |
| `gateway_unique_id` | The Gateway-side identifier of the profile | Gateway policies that reference device profiles |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Contractor lockdown** — third-party devices in an identity group run full-tunnel WARP they cannot switch off, switch modes on, or unenroll from. Precedence 100 leaves room below for narrower emergency overrides. Start from the **Contractor lockdown** preset.

**Developer LAN access** — engineers get a 30-minute LAN window on a /24, a lab network excluded from the tunnel, and an internal suffix resolving against the lab resolver through the profile's own fallback list. Precedence 200 ranks after the contractor profile, so an engineer flagged as both gets the tighter profile. Start from the **Developer LAN access** preset.

**OS-specific handling** — a Windows-only profile (`match: os.name == "windows"`) carrying `sccmVpnBoundarySupport`, or a grace profile for an older macOS version pinned by `os.version`.

## Works With

- [**Cloudflare Zero Trust Device Default Profile**](/infra-catalog/cloudflare-zero-trust-device-default-profile) — the account baseline this profile overrides; devices matching no custom profile land there.
- [**Cloudflare Zero Trust Device Posture Rule**](/infra-catalog/cloudflare-zero-trust-device-posture-rule) — health checks that can gate what this profile's devices reach.
- [**Cloudflare Zero Trust Tunnel Virtual Network**](/infra-catalog/cloudflare-zero-trust-tunnel-virtual-network) — the routing segments `virtualNetworks` scopes device access to.
