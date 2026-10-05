# CloudflareZeroTrustDeviceCustomProfile

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareZeroTrustDeviceCustomProfileSpec creates a TARGETED WARP device
profile: the same settings body as the account's default profile, applied
only to the devices matched by a wirefilter expression. An account can
carry many custom profiles; the lowest precedence value wins when several
match. Unlike the default profile, a custom profile is a real object --
create, update, and delete all do what they say.

The profile's local-DNS fallback list folds into this kind: Cloudflare
reports it read-only on the profile's own API and writes it through a
dedicated per-profile full-replacement endpoint, so the list here is
declarative -- what you declare is exactly what exists after apply.

The provider's global_acceleration block (China network endpoints) is
deliberately not modeled: Cloudflare enables that feature per account
through an account representative, so no self-service caller can use it.

## Example

```yaml
# Complete example manifest for CloudflareZeroTrustDeviceCustomProfile.
# Creates a targeted WARP profile for one identity group with an
# exclude-mode split tunnel and the folded per-profile local-DNS fallback
# list (2 resources to add).
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustDeviceCustomProfile
metadata:
  name: contractor-devices
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: contractor-devices
  match: identity.groups.name == "contractors"
  precedence: 100
  enabled: true
  description: settings for contractor laptops
  allowed_to_leave: false
  captive_portal: 180
  switch_locked: true
  exclude:
    - address: 192.0.2.0/24
      description: lab network stays local
  service_mode_v2:
    mode: warp
  dns_search_suffixes:
    - suffix: corp.example.com
  fallback_domains:
    - suffix: corp.internal
      description: resolved by the on-prem resolver
      dns_server:
        - 10.0.0.53
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.match` | `string` | yes |  |  |
| `spec.precedence` | `int64` | yes |  |  |
| `spec.enabled` | `bool` |  |  |  |
| `spec.description` | `string` |  |  |  |
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
| `spec.exclude` | `[]CloudflareZeroTrustDeviceCustomProfileSplitTunnelEntry` |  |  |  |
| `spec.exclude[].address` | `string` |  |  |  |
| `spec.exclude[].host` | `string` |  |  |  |
| `spec.exclude[].description` | `string` |  |  |  |
| `spec.include` | `[]CloudflareZeroTrustDeviceCustomProfileSplitTunnelEntry` |  |  |  |
| `spec.include[].address` | `string` |  |  |  |
| `spec.include[].host` | `string` |  |  |  |
| `spec.include[].description` | `string` |  |  |  |
| `spec.serviceModeV2` | `CloudflareZeroTrustDeviceCustomProfileServiceMode` |  |  |  |
| `spec.serviceModeV2.mode` | `string` | yes |  |  |
| `spec.serviceModeV2.port` | `int64` |  |  |  |
| `spec.virtualNetworks` | `CloudflareZeroTrustDeviceCustomProfileVirtualNetworks` |  |  |  |
| `spec.virtualNetworks.allowed` | `[]string \| valueFrom` | yes |  | CloudflareZeroTrustTunnelVirtualNetwork (`status.outputs.virtual_network_id`) |
| `spec.virtualNetworks.defaultVirtualNetworkId` | `string \| valueFrom` | yes |  | CloudflareZeroTrustTunnelVirtualNetwork (`status.outputs.virtual_network_id`) |
| `spec.dnsSearchSuffixes` | `[]CloudflareZeroTrustDeviceCustomProfileDnsSearchSuffix` |  |  |  |
| `spec.dnsSearchSuffixes[].suffix` | `string` | yes |  |  |
| `spec.dnsSearchSuffixes[].description` | `string` |  |  |  |
| `spec.fallbackDomains` | `[]CloudflareZeroTrustDeviceCustomProfileFallbackDomain` |  |  |  |
| `spec.fallbackDomains[].suffix` | `string` | yes |  |  |
| `spec.fallbackDomains[].description` | `string` |  |  |  |
| `spec.fallbackDomains[].dnsServer` | `[]string` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account the profile belongs to.

- rule: account_id must be a 32-character hex string
- rule: {"required":true}

### spec.name

`string` · required

The profile's name, shown in the dashboard.

- rule: {"required":true}

### spec.match

`string` · required

The wirefilter expression selecting which devices get this profile.
Available selectors at the provider: identity.email,
identity.groups.id, identity.groups.name, identity.groups.email,
identity.service_token_uuid, identity.saml_attributes, network,
os.name, os.version. Example:
  identity.email == "dev@example.com" or os.name == "windows"

- rule: {"required":true}

### spec.precedence

`int64` · required

The profile's evaluation order: LOWER values win when several profiles
match a device. Required here although the provider marks it optional
-- Cloudflare's API rejects a create without it (every provider test
sets one), and an explicit ordering is what makes multi-profile
accounts predictable.

- rule: {"required":true,"int64":{"gte":"1"}}

### spec.enabled

`bool` · optional (explicit presence)

Whether the profile is applied to matching devices. Unset keeps
Cloudflare's default (ON).

### spec.description

`string`

A description of the profile.

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

The URL the client's "Send feedback" button opens.

### spec.switchLocked

`bool` · optional (explicit presence)

Whether users are BLOCKED from turning WARP off. Unset keeps
Cloudflare's default (off -- users may disconnect).

### spec.tunnelProtocol

`string`

The tunnel transport protocol: wireguard or masque. Cloudflare's API
accepts the value it currently supports for the account; empty keeps
the account default. Deliberately not CEL-walled -- the provider schema
carries no value list at v5.23.0 and Cloudflare has grown this set
before.

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

`[]CloudflareZeroTrustDeviceCustomProfileSplitTunnelEntry`

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

`[]CloudflareZeroTrustDeviceCustomProfileSplitTunnelEntry`

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

`CloudflareZeroTrustDeviceCustomProfileServiceMode`

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

`CloudflareZeroTrustDeviceCustomProfileVirtualNetworks`

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

`[]CloudflareZeroTrustDeviceCustomProfileDnsSearchSuffix`

DNS search suffixes appended when devices resolve short hostnames,
evaluated in order.

### spec.dnsSearchSuffixes[].suffix

`string` · required

The suffix appended when resolving short hostnames (e.g.
"corp.example.com").

- rule: {"required":true}

### spec.dnsSearchSuffixes[].description

`string`

A note describing the suffix.

### spec.fallbackDomains

`[]CloudflareZeroTrustDeviceCustomProfileFallbackDomain`

This profile's local-DNS fallback list: domain suffixes resolved by the
declared resolvers instead of Gateway. FULL-REPLACEMENT semantics --
this list, exactly, is what exists after apply. Rows ride the profile:
deleting the profile retires them with it.

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

## Validation Rules

- `spec.exclude_xor_include`: exclude and include cannot both be set -- a profile runs in exclude mode or include mode, never both

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZeroTrustDeviceCustomProfile, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.policy_id` | `string` | The Cloudflare-assigned identifier of the profile (its policy ID). |
| `status.outputs.gateway_unique_id` | `string` | The Gateway-side identifier Cloudflare assigns the profile (used when Gateway policies reference device profiles). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.virtualNetworks.allowed` | CloudflareZeroTrustTunnelVirtualNetwork | `status.outputs.virtual_network_id` |
| `spec.virtualNetworks.defaultVirtualNetworkId` | CloudflareZeroTrustTunnelVirtualNetwork | `status.outputs.virtual_network_id` |

## See Also

- [Overview](../README.md)
