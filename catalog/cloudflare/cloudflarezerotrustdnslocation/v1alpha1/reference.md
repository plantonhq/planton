# CloudflareZeroTrustDnsLocation

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareZeroTrustDnsLocationSpec creates a Gateway DNS location: a named
entry point (office, site, network) whose DNS traffic Gateway filters.
Cloudflare assigns the location its resolver endpoints -- a DoH subdomain
and destination IPs -- and Gateway policies can then match on the
location.

A location is a plain CRUD object (real create, update, delete). Two
behaviors worth knowing:
  - Update is a full replace: omitting max_ttl on update resets the TTL
    behavior to inherit.
  - dns_destination_ips_id left unset lets Cloudflare auto-assign the
    shared IPv4 destination pair -- never pin the shared pool's UUID
    yourself.

## Example

```yaml
# Complete example manifest for CloudflareZeroTrustDnsLocation. Creates a
# Gateway DNS location with all four endpoint types declared, source
# networks for the IPv4 endpoint, and a TTL override.
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustDnsLocation
metadata:
  name: hq-office
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: hq-office
  ecs_support: false
  endpoints:
    doh:
      enabled: true
      require_token: true
    dot:
      enabled: false
    ipv4:
      enabled: true
    ipv6:
      enabled: false
  networks:
    - network: 203.0.113.0/24
  max_ttl:
    mode: override
    ttl_secs: 300
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.clientDefault` | `bool` |  |  |  |
| `spec.ecsSupport` | `bool` |  |  |  |
| `spec.dnsDestinationIpsId` | `string` |  |  |  |
| `spec.endpoints` | `CloudflareZeroTrustDnsLocationEndpoints` |  |  |  |
| `spec.endpoints.doh` | `CloudflareZeroTrustDnsLocationDohEndpoint` | yes |  |  |
| `spec.endpoints.doh.enabled` | `bool` |  |  |  |
| `spec.endpoints.doh.requireToken` | `bool` |  |  |  |
| `spec.endpoints.doh.networks` | `[]CloudflareZeroTrustDnsLocationNetwork` |  |  |  |
| `spec.endpoints.doh.networks[].network` | `string` | yes |  |  |
| `spec.endpoints.dot` | `CloudflareZeroTrustDnsLocationNetworkEndpoint` | yes |  |  |
| `spec.endpoints.dot.enabled` | `bool` |  |  |  |
| `spec.endpoints.dot.networks` | `[]CloudflareZeroTrustDnsLocationNetwork` |  |  |  |
| `spec.endpoints.dot.networks[].network` | `string` | yes |  |  |
| `spec.endpoints.ipv4` | `CloudflareZeroTrustDnsLocationIpv4Endpoint` | yes |  |  |
| `spec.endpoints.ipv4.enabled` | `bool` |  |  |  |
| `spec.endpoints.ipv6` | `CloudflareZeroTrustDnsLocationNetworkEndpoint` | yes |  |  |
| `spec.endpoints.ipv6.enabled` | `bool` |  |  |  |
| `spec.endpoints.ipv6.networks` | `[]CloudflareZeroTrustDnsLocationNetwork` |  |  |  |
| `spec.endpoints.ipv6.networks[].network` | `string` | yes |  |  |
| `spec.networks` | `[]CloudflareZeroTrustDnsLocationNetwork` |  |  |  |
| `spec.networks[].network` | `string` | yes |  |  |
| `spec.maxTtl` | `CloudflareZeroTrustDnsLocationMaxTtl` |  |  |  |
| `spec.maxTtl.mode` | `string` | yes |  |  |
| `spec.maxTtl.ttlSecs` | `int64` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account the location belongs to.

- rule: account_id must be a 32-character hex string
- rule: {"required":true}

### spec.name

`string` · required

The location's name, shown in the dashboard and matched by policies.

- rule: {"required":true}

### spec.clientDefault

`bool` · optional (explicit presence)

When true, this location is the account's default: traffic from
unregistered sources is attributed to it.

Two Cloudflare walls, both live-measured: the FIRST location created on
an account is auto-promoted to default even when this field is unset;
and the current default can neither be deleted (API error 1217) nor
demoted in place (error 1216) -- the only way to move the default is
setting client_default true on ANOTHER location. Consequence: an
account's first DNS location is effectively permanent until a
replacement default exists, and destroying a resource that holds the
default FAILS until the default is transferred.

### spec.ecsSupport

`bool` · optional (explicit presence)

When true, the location's resolver honors EDNS Client Subnet, letting
upstream resolvers see the querying network for better geo-answers.

### spec.dnsDestinationIpsId

`string`

The DNS destination IPs to assign, by Cloudflare mapping UUID. Leave
unset to auto-assign the shared IPv4 destination pair (recommended);
set only when the account has a dedicated (BYOIP / dedicated resolver
IP) mapping to pin.

### spec.endpoints

`CloudflareZeroTrustDnsLocationEndpoints`

The resolver endpoint types this location accepts queries on. Setting
this declares ALL FOUR endpoint types (Cloudflare's API takes the whole
tree at once); unset keeps Cloudflare's endpoint defaults.

### spec.endpoints.doh

`CloudflareZeroTrustDnsLocationDohEndpoint` · required

DNS over HTTPS -- the endpoint WARP and browsers use.

- rule: {"required":true}

### spec.endpoints.doh.enabled

`bool` · optional (explicit presence)

Whether the endpoint accepts queries.

### spec.endpoints.doh.requireToken

`bool` · optional (explicit presence)

Require a per-location authentication token on DoH queries (rejects
anonymous resolvers pointed at the subdomain).

### spec.endpoints.doh.networks

`[]CloudflareZeroTrustDnsLocationNetwork`

The source networks allowed to query this endpoint, as IPs or CIDRs.
Empty allows any source -- but note that at Terraform provider
v5.23.0/v5.24.0 a DECLARED endpoint with an empty network list re-plans
a cosmetic in-place update forever (Cloudflare drops empty lists on
read; live-measured): declare real networks for a drift-free plan, or
accept the no-op diff.

### spec.endpoints.doh.networks[].network

`string` · required

The IP address or CIDR (e.g. "203.0.113.0/24").

- rule: {"required":true}

### spec.endpoints.dot

`CloudflareZeroTrustDnsLocationNetworkEndpoint` · required

DNS over TLS.

- rule: {"required":true}

### spec.endpoints.dot.enabled

`bool` · optional (explicit presence)

Whether the endpoint accepts queries.

### spec.endpoints.dot.networks

`[]CloudflareZeroTrustDnsLocationNetwork`

The source networks allowed to query this endpoint, as IPs or CIDRs.
Empty allows any source -- but note that at Terraform provider
v5.23.0/v5.24.0 a DECLARED endpoint with an empty network list re-plans
a cosmetic in-place update forever (Cloudflare drops empty lists on
read; live-measured): declare real networks for a drift-free plan, or
accept the no-op diff.

### spec.endpoints.dot.networks[].network

`string` · required

The IP address or CIDR (e.g. "203.0.113.0/24").

- rule: {"required":true}

### spec.endpoints.ipv4

`CloudflareZeroTrustDnsLocationIpv4Endpoint` · required

Plain IPv4 DNS. Source networks for IPv4 are declared at the spec's
top-level networks field, not here -- the one endpoint type without an
inline network list.

- rule: {"required":true}

### spec.endpoints.ipv4.enabled

`bool` · optional (explicit presence)

Whether the endpoint accepts queries.

### spec.endpoints.ipv6

`CloudflareZeroTrustDnsLocationNetworkEndpoint` · required

Plain IPv6 DNS.

- rule: {"required":true}

### spec.endpoints.ipv6.enabled

`bool` · optional (explicit presence)

Whether the endpoint accepts queries.

### spec.endpoints.ipv6.networks

`[]CloudflareZeroTrustDnsLocationNetwork`

The source networks allowed to query this endpoint, as IPs or CIDRs.
Empty allows any source -- but note that at Terraform provider
v5.23.0/v5.24.0 a DECLARED endpoint with an empty network list re-plans
a cosmetic in-place update forever (Cloudflare drops empty lists on
read; live-measured): declare real networks for a drift-free plan, or
accept the no-op diff.

### spec.endpoints.ipv6.networks[].network

`string` · required

The IP address or CIDR (e.g. "203.0.113.0/24").

- rule: {"required":true}

### spec.networks

`[]CloudflareZeroTrustDnsLocationNetwork`

The source networks allowed to use this location's IPv4 endpoint, as
IPv4 CIDRs. Cloudflare caps IPv4 CIDRs at /24 (nothing broader).
Effective only while the IPv4 endpoint is enabled.

### spec.networks[].network

`string` · required

The IP address or CIDR (e.g. "203.0.113.0/24").

- rule: {"required":true}

### spec.maxTtl

`CloudflareZeroTrustDnsLocationMaxTtl`

How DNS response TTLs are capped for this location relative to the
account setting. OMITTING this on an update RESETS the behavior to
inherit -- a managed location should keep it declared.

- rule: ttl_secs is required with mode override and must be omitted with inherit or disabled

### spec.maxTtl.mode

`string` · required

inherit follows the account's max-TTL setting; override caps at this
location's ttl_secs; disabled returns TTLs unchanged.

- rule: mode must be one of inherit, override, disabled
- rule: {"required":true}

### spec.maxTtl.ttlSecs

`int64` · optional (explicit presence)

The TTL cap in seconds (60-36000). Only with mode override.

- rule: {"int64":{"lte":"36000","gte":"60"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZeroTrustDnsLocation, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.location_id` | `string` | The Cloudflare-assigned UUID of the location. |
| `status.outputs.doh_subdomain` | `string` | The location's unique DNS-over-HTTPS subdomain (the value clients embed in https://<doh_subdomain>.cloudflare-gateway.com/dns-query). |
| `status.outputs.ip` | `string` | The IPv4 destination assigned to the location's plain-DNS endpoint. |

## See Also

- [Overview](../README.md)
