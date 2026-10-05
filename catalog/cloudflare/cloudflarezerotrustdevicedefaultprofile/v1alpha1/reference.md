# CloudflareZeroTrustDeviceDefaultProfile

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareZeroTrustDeviceDefaultProfileSpec manages the account's DEFAULT
WARP device profile: the settings every enrolled device receives unless a
custom profile matches it first. The profile always exists on the account
-- applying this spec EDITS it in place (PATCH), and destroy is a NO-OP
that leaves the last-applied values standing. Treat it like a settings
panel, not an object with a lifecycle.

Two companion surfaces fold into this kind because Cloudflare exposes them
as separate write paths on the same profile:
  - fallback_domains: the profile's local-DNS fallback list. The profile's
    own API reports it read-only; writes go through a dedicated
    full-replacement endpoint, so the list here is declarative -- what you
    declare is exactly what exists after apply. Its destroy is also a
    no-op.
  - zone_certificates: the per-zone WARP client certificate provisioning
    toggle (the one ZONE-scoped surface on this account-scoped kind).
    Cloudflare offers no delete for it and no import.

The provider's global_acceleration block (China network endpoints) is
deliberately not modeled: Cloudflare enables that feature per account
through an account representative, so no self-service caller can use it.

## Example

```yaml
# Complete example manifest for CloudflareZeroTrustDeviceDefaultProfile.
# Applies the account's default WARP profile with the toggle body, an
# exclude-mode split tunnel, proxy service mode, DNS search suffixes, the
# folded local-DNS fallback list, and the folded per-zone certificate
# provisioning toggle (3 resources to add).
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustDeviceDefaultProfile
metadata:
  name: default-device-profile
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  allow_mode_switch: false
  allow_updates: true
  allowed_to_leave: false
  auto_connect: 600
  captive_portal: 180
  disable_auto_fallback: false
  exclude_office_ips: true
  register_interface_ip_with_dns: true
  sccm_vpn_boundary_support: false
  support_url: https://support.example.com/warp
  switch_locked: true
  tunnel_protocol: wireguard
  lan_allow_minutes: 30
  lan_allow_subnet_size: 24
  exclude:
    - address: 192.0.2.0/24
      description: lab network stays local
    - host: printers.example.com
      description: local printer discovery
  service_mode_v2:
    mode: warp
  dns_search_suffixes:
    - suffix: corp.example.com
      description: short hostnames resolve in the corp domain
  fallback_domains:
    - suffix: corp.internal
      description: resolved by the on-prem resolver
      dns_server:
        - 10.0.0.53
    - suffix: localdomain
  zone_certificates:
    zone_id:
      value: "REPLACE_WITH_ZONE_ID"
    enabled: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.allowModeSwitch` | `bool` |  |  |  |
| `spec.allowUpdates` | `bool` |  |  |  |
| `spec.allowedToLeave` | `bool` |  |  |  |
| `spec.autoConnect` | `int64` |  |  |  |
| `spec.captivePortal` | `int64` |  |  |  |
| `spec.disableAutoFallback` | `bool` |  |  |  |
| `spec.excludeOfficeIps` | `bool` |  |  |  |
| `spec.registerInterfaceIpWithDns` | `bool` |  |  |  |
| `spec.sccmVpnBoundarySupport` | `bool` |  |  |  |
| `spec.supportUrl` | `string` |  |  |  |
| `spec.switchLocked` | `bool` |  |  |  |
| `spec.tunnelProtocol` | `string` |  |  |  |
| `spec.lanAllowMinutes` | `int64` |  |  |  |
| `spec.lanAllowSubnetSize` | `int64` |  |  |  |
| `spec.exclude` | `[]CloudflareZeroTrustDeviceDefaultProfileSplitTunnelEntry` |  |  |  |
| `spec.exclude[].address` | `string` |  |  |  |
| `spec.exclude[].host` | `string` |  |  |  |
| `spec.exclude[].description` | `string` |  |  |  |
| `spec.include` | `[]CloudflareZeroTrustDeviceDefaultProfileSplitTunnelEntry` |  |  |  |
| `spec.include[].address` | `string` |  |  |  |
| `spec.include[].host` | `string` |  |  |  |
| `spec.include[].description` | `string` |  |  |  |
| `spec.serviceModeV2` | `CloudflareZeroTrustDeviceDefaultProfileServiceMode` |  |  |  |
| `spec.serviceModeV2.mode` | `string` | yes |  |  |
| `spec.serviceModeV2.port` | `int64` |  |  |  |
| `spec.virtualNetworks` | `CloudflareZeroTrustDeviceDefaultProfileVirtualNetworks` |  |  |  |
| `spec.virtualNetworks.allowed` | `[]string \| valueFrom` | yes |  | CloudflareZeroTrustTunnelVirtualNetwork (`status.outputs.virtual_network_id`) |
| `spec.virtualNetworks.defaultVirtualNetworkId` | `string \| valueFrom` | yes |  | CloudflareZeroTrustTunnelVirtualNetwork (`status.outputs.virtual_network_id`) |
| `spec.dnsSearchSuffixes` | `[]CloudflareZeroTrustDeviceDefaultProfileDnsSearchSuffix` |  |  |  |
| `spec.dnsSearchSuffixes[].suffix` | `string` | yes |  |  |
| `spec.dnsSearchSuffixes[].description` | `string` |  |  |  |
| `spec.fallbackDomains` | `[]CloudflareZeroTrustDeviceDefaultProfileFallbackDomain` |  |  |  |
| `spec.fallbackDomains[].suffix` | `string` | yes |  |  |
| `spec.fallbackDomains[].description` | `string` |  |  |  |
| `spec.fallbackDomains[].dnsServer` | `[]string` |  |  |  |
| `spec.zoneCertificates` | `CloudflareZeroTrustDeviceDefaultProfileZoneCertificates` |  |  |  |
| `spec.zoneCertificates.zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.zoneCertificates.enabled` | `bool` | yes |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account whose default device profile this manages.

- rule: account_id must be a 32-character hex string
- rule: {"required":true}

### spec.allowModeSwitch

`bool` · optional (explicit presence)

Whether users may switch the WARP client between modes (e.g. from full
tunnel to DNS-only). Unset keeps Cloudflare's default (off).

### spec.allowUpdates

`bool` · optional (explicit presence)

Whether devices show update notifications when a new client version is
available. Unset keeps Cloudflare's default (off).

### spec.allowedToLeave

`bool` · optional (explicit presence)

Whether users may unenroll their device from the organization. Unset
keeps Cloudflare's default (ON) -- set false to lock devices in.

### spec.autoConnect

`int64` · optional (explicit presence)

Seconds after which a user-disabled WARP client reconnects on its own.
0 means stay disabled until the user re-enables. Unset keeps
Cloudflare's default (0).

- rule: {"int64":{"gte":"0"}}

### spec.captivePortal

`int64` · optional (explicit presence)

Seconds WARP waits behind a captive portal (hotel/airport Wi-Fi login
page) before re-engaging. Unset keeps Cloudflare's default (180).

- rule: {"int64":{"gte":"0"}}

### spec.disableAutoFallback

`bool` · optional (explicit presence)

When true, a fallback domain row without dns_server entries fails
closed instead of guessing the system resolvers. Unset keeps
Cloudflare's default (off).

### spec.excludeOfficeIps

`bool` · optional (explicit presence)

Whether Microsoft 365 IPs are automatically added to the split-tunnel
exclude list. Unset keeps Cloudflare's default (off).

### spec.registerInterfaceIpWithDns

`bool` · optional (explicit presence)

Whether the OS registers WARP's local interface IP with the
organization's on-premises DNS servers. Unset keeps Cloudflare's
default (ON).

### spec.sccmVpnBoundarySupport

`bool` · optional (explicit presence)

Windows only: whether WARP tells Microsoft SCCM it is inside a VPN
boundary. Unset keeps Cloudflare's default (off).

### spec.supportUrl

`string`

The URL the client's "Send feedback" button opens. Empty hides nothing
-- Cloudflare shows its own default flow.

### spec.switchLocked

`bool` · optional (explicit presence)

Whether users are BLOCKED from turning WARP off. Unset keeps
Cloudflare's default (off -- users may disconnect).

### spec.tunnelProtocol

`string`

The tunnel transport protocol: wireguard or masque. Deliberately not
CEL-walled -- the provider schema carries no value list at v5.23.0 and
Cloudflare has grown this set before. CAUTION: leaving this empty does
NOT preserve the account's current protocol -- the provider defaults an
unset value to the empty string and sends it, resetting the account to
Cloudflare's default protocol (measured live 2026-08-27: an apply that
omitted this field blanked an account's masque setting). Declare the
protocol explicitly on any account running a non-default one.

### spec.lanAllowMinutes

`int64` · optional (explicit presence)

Minutes a user may access their local network (LAN) after toggling
local-network access. 0 allows LAN access until the next WARP
reconnection (reboot, wake from sleep).

- rule: {"int64":{"gte":"0"}}

### spec.lanAllowSubnetSize

`int64` · optional (explicit presence)

The subnet size WARP carves out for local access network traffic (e.g.
24 for a /24).

- rule: {"int64":{"lte":"128","gte":"0"}}

### spec.exclude

`[]CloudflareZeroTrustDeviceDefaultProfileSplitTunnelEntry`

Routes EXCLUDED from the WARP tunnel (everything else tunnels).
Mutually exclusive with include.

- rule: declare exactly one of address (CIDR) or host (domain name)

### spec.exclude[].address

`string`

The route as a CIDR (e.g. "192.0.2.0/24"). Exactly one of address or
host.

### spec.exclude[].host

`string`

The route as a domain name (e.g. "internal.example.com"). Exactly one
of address or host.

### spec.exclude[].description

`string`

A note shown in the client UI next to the route.

### spec.include

`[]CloudflareZeroTrustDeviceDefaultProfileSplitTunnelEntry`

Routes INCLUDED in the WARP tunnel (everything else bypasses).
Mutually exclusive with exclude.

- rule: declare exactly one of address (CIDR) or host (domain name)

### spec.include[].address

`string`

The route as a CIDR (e.g. "192.0.2.0/24"). Exactly one of address or
host.

### spec.include[].host

`string`

The route as a domain name (e.g. "internal.example.com"). Exactly one
of address or host.

### spec.include[].description

`string`

A note shown in the client UI next to the route.

### spec.serviceModeV2

`CloudflareZeroTrustDeviceDefaultProfileServiceMode`

The client's service mode (e.g. proxy mode with a local port). Unset
keeps the default full-tunnel warp mode.

### spec.serviceModeV2.mode

`string` · required

The client mode. Cloudflare's common values are warp (full tunnel) and
proxy (local proxy on `port`); the API owns the accepted set, so this
is deliberately not CEL-walled.

- rule: {"required":true}

### spec.serviceModeV2.port

`int64` · optional (explicit presence)

The localhost port for proxy mode.

- rule: {"int64":{"lte":"65535","gte":"1"}}

### spec.virtualNetworks

`CloudflareZeroTrustDeviceDefaultProfileVirtualNetworks`

Which Zero Trust virtual networks devices on this profile may reach,
and which one they land in by default. Unset leaves virtual network
routing to the account's default virtual network.

### spec.virtualNetworks.allowed

`[]string | valueFrom` · required

The virtual networks devices may access -- literal virtual network
UUIDs or references to CloudflareZeroTrustTunnelVirtualNetwork
resources. At least one.

- references: CloudflareZeroTrustTunnelVirtualNetwork (`status.outputs.virtual_network_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareZeroTrustTunnelVirtualNetwork, name: <that resource's name>, fieldPath: status.outputs.virtual_network_id}} -- a bare string does not parse

### spec.virtualNetworks.defaultVirtualNetworkId

`string | valueFrom` · required

The virtual network devices land in by default. Must be one of the
allowed entries (Cloudflare enforces the containment at the API).

- references: CloudflareZeroTrustTunnelVirtualNetwork (`status.outputs.virtual_network_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareZeroTrustTunnelVirtualNetwork, name: <that resource's name>, fieldPath: status.outputs.virtual_network_id}} -- a bare string does not parse

### spec.dnsSearchSuffixes

`[]CloudflareZeroTrustDeviceDefaultProfileDnsSearchSuffix`

DNS search suffixes appended when devices resolve short hostnames,
evaluated in order. This list is FULLY MANAGED by this field: leaving
it empty clears the account's list (unset and empty are the same
declaration, and the modules always send the list -- the provider's
attribute re-plans forever on an omitted send, measured at v5.23.0).

### spec.dnsSearchSuffixes[].suffix

`string` · required

The suffix appended when resolving short hostnames (e.g.
"corp.example.com").

- rule: {"required":true}

### spec.dnsSearchSuffixes[].description

`string`

A note describing the suffix.

### spec.fallbackDomains

`[]CloudflareZeroTrustDeviceDefaultProfileFallbackDomain`

The profile's local-DNS fallback list: domain suffixes resolved by the
declared resolvers instead of Gateway. FULL-REPLACEMENT semantics --
this list, exactly, is what exists after apply (rows added outside IaC
are removed). Cloudflare seeds every account with a default list; once
this field is declared, the declaration owns the whole list.

### spec.fallbackDomains[].suffix

`string` · required

The domain suffix to resolve locally (e.g. "corp.internal").

- rule: {"required":true}

### spec.fallbackDomains[].description

`string`

A note shown in the client UI next to the row.

### spec.fallbackDomains[].dnsServer

`[]string`

The resolver IPs handling this suffix. Empty falls back to the
system's resolvers unless the profile sets disable_auto_fallback.

### spec.zoneCertificates

`CloudflareZeroTrustDeviceDefaultProfileZoneCertificates`

Per-zone WARP client certificate provisioning: devices get a client
certificate from the zone, letting origins verify traffic really came
through WARP. The one zone-scoped surface on this account-scoped kind.
Cloudflare offers no delete and no import for it -- unset means "not
managed here", never "disabled". CREDENTIAL WALL (measured live
2026-08-27): the endpoint refuses ACCOUNT-OWNED API tokens on both
reads and writes with 401 code 1039 "malformed actor email claim" --
managing this fold needs a user-actor credential (a user-owned token
or API key + email).

### spec.zoneCertificates.zoneId

`string | valueFrom` · required

The zone whose client certificate devices should receive -- a literal
zone ID or a reference to a CloudflareDnsZone.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.zoneCertificates.enabled

`bool` · required · optional (explicit presence)

Whether certificate provisioning is on for the zone. Declared
explicitly (true or false) because the surface has no delete --
turning it off IS the managed off state.

- rule: {"required":true}

## Validation Rules

- `spec.exclude_xor_include`: exclude and include cannot both be set -- a profile runs in exclude mode or include mode, never both

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZeroTrustDeviceDefaultProfile, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.account_id` | `string` | The Cloudflare account the profile belongs to -- the profile is an account singleton, so the account IS its identity. |
| `status.outputs.gateway_unique_id` | `string` | The Gateway-side identifier Cloudflare assigns the profile (used when Gateway policies reference device profiles). |
| `status.outputs.policy_id` | `string` | The profile's policy identifier as reported by the device policy API. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.virtualNetworks.allowed` | CloudflareZeroTrustTunnelVirtualNetwork | `status.outputs.virtual_network_id` |
| `spec.virtualNetworks.defaultVirtualNetworkId` | CloudflareZeroTrustTunnelVirtualNetwork | `status.outputs.virtual_network_id` |
| `spec.zoneCertificates.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
