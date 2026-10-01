# CloudflareDnsZone

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**CloudflareDnsZoneSpec** defines a Cloudflare DNS zone: the domain, its
account, common zone-level options, optional zone-wide DNS settings and DNSSEC,
and an optional set of DNS records managed alongside the zone. Records may also
be managed independently as first-class CloudflareDnsRecord resources; the
embedded list is a convenience for records whose lifecycle tracks the zone.

## Example

```yaml
# Offline development manifest exercising the full spec surface: zone core,
# simple and structured inline records (typed data cases at the record level),
# record tags/settings, zone-wide DNS settings incl. SOA tuning, DNSSEC, zone
# hold, and the plan subscription. Live-blocked surfaces stay OUT of the live
# scenarios and this manifest is their offline tofu-plan proof: hold and
# subscription (billing token scope; unverified hold semantics on a throwaway
# fixture zone), plus four free-plan/pending-zone walls measured live
# 2026-08-26 -- record tags (400 code 9300, tag quota 0), record-level
# ipv4_only/ipv6_only settings (400 code 9227), custom SOA tuning (400 code
# 1003), and DNSSEC on a PENDING zone (400 code 1017 "Invalid zone plan for
# action"; the same PATCH succeeds on an ACTIVE free-plan zone, so the wall is
# delegation state, not plan tier).
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareDnsZone
metadata:
  name: dns-zone-hack
spec:
  zoneName: planton-example.com
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  paused: false
  records:
    - name: www
      type: A
      content:
        value: "192.0.2.1"
      proxied: true
      ttl: 1
      settings:
        ipv4Only: true
      tags:
        - team:web
    - name: "@"
      type: MX
      content:
        value: mail.planton-example.com
      ttl: 1
      priority: 10
    - name: _sip._tcp
      type: SRV
      srv:
        priority: 10
        weight: 5
        port: 5060
        target:
          value: sip.planton-example.com
    - name: "@"
      type: CAA
      caa:
        tag: issue
        value: letsencrypt.org
    - name: "@"
      type: HTTPS
      https:
        priority: 1
        target:
          value: "."
        value: alpn="h2"
  dnsSettings:
    flattenAllCnames: true
    zoneMode: standard
    nsTtl: 3600
    soa:
      minTtl: 1800
      ttl: 3600
  dnssec:
    enabled: true
  hold:
    enabled: true
    includeSubdomains: true
  subscription:
    ratePlan: free
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.zoneName` | `string` | yes |  |  |
| `spec.accountId` | `string` | yes |  |  |
| `spec.paused` | `bool` |  |  |  |
| `spec.records` | `[]CloudflareDnsZoneRecord` |  |  |  |
| `spec.records[].name` | `string` | yes |  |  |
| `spec.records[].type` | `enum` | yes |  |  |
| `spec.records[].content` | `string \| valueFrom` |  |  |  |
| `spec.records[].proxied` | `bool` |  |  |  |
| `spec.records[].ttl` | `int32` |  |  |  |
| `spec.records[].priority` | `int32` |  |  |  |
| `spec.records[].comment` | `string` |  |  |  |
| `spec.records[].caa` | `CaaData` |  |  |  |
| `spec.records[].caa.flags` | `uint32` |  |  |  |
| `spec.records[].caa.tag` | `string` | yes |  |  |
| `spec.records[].caa.value` | `string` | yes |  |  |
| `spec.records[].cert` | `CertData` |  |  |  |
| `spec.records[].cert.type` | `uint32` |  |  |  |
| `spec.records[].cert.keyTag` | `uint32` |  |  |  |
| `spec.records[].cert.algorithm` | `uint32` |  |  |  |
| `spec.records[].cert.certificate` | `string` | yes |  |  |
| `spec.records[].dnskey` | `DnskeyData` |  |  |  |
| `spec.records[].dnskey.flags` | `uint32` |  |  |  |
| `spec.records[].dnskey.protocol` | `uint32` |  |  |  |
| `spec.records[].dnskey.algorithm` | `uint32` |  |  |  |
| `spec.records[].dnskey.publicKey` | `string` | yes |  |  |
| `spec.records[].ds` | `DsData` |  |  |  |
| `spec.records[].ds.keyTag` | `uint32` |  |  |  |
| `spec.records[].ds.algorithm` | `uint32` |  |  |  |
| `spec.records[].ds.digestType` | `uint32` |  |  |  |
| `spec.records[].ds.digest` | `string` | yes |  |  |
| `spec.records[].https` | `HttpsData` |  |  |  |
| `spec.records[].https.priority` | `uint32` |  |  |  |
| `spec.records[].https.target` | `string \| valueFrom` | yes |  |  |
| `spec.records[].https.value` | `string` |  |  |  |
| `spec.records[].loc` | `LocData` |  |  |  |
| `spec.records[].loc.latDirection` | `string` |  |  |  |
| `spec.records[].loc.latDegrees` | `uint32` |  |  |  |
| `spec.records[].loc.latMinutes` | `uint32` |  |  |  |
| `spec.records[].loc.latSeconds` | `double` |  |  |  |
| `spec.records[].loc.longDirection` | `string` |  |  |  |
| `spec.records[].loc.longDegrees` | `uint32` |  |  |  |
| `spec.records[].loc.longMinutes` | `uint32` |  |  |  |
| `spec.records[].loc.longSeconds` | `double` |  |  |  |
| `spec.records[].loc.altitude` | `double` |  |  |  |
| `spec.records[].loc.size` | `double` |  |  |  |
| `spec.records[].loc.precisionHorz` | `double` |  |  |  |
| `spec.records[].loc.precisionVert` | `double` |  |  |  |
| `spec.records[].naptr` | `NaptrData` |  |  |  |
| `spec.records[].naptr.order` | `uint32` |  |  |  |
| `spec.records[].naptr.preference` | `uint32` |  |  |  |
| `spec.records[].naptr.flags` | `string` |  |  |  |
| `spec.records[].naptr.service` | `string` |  |  |  |
| `spec.records[].naptr.regex` | `string` |  |  |  |
| `spec.records[].naptr.replacement` | `string` |  |  |  |
| `spec.records[].smimea` | `SmimeaData` |  |  |  |
| `spec.records[].smimea.usage` | `uint32` |  |  |  |
| `spec.records[].smimea.selector` | `uint32` |  |  |  |
| `spec.records[].smimea.matchingType` | `uint32` |  |  |  |
| `spec.records[].smimea.certificate` | `string` | yes |  |  |
| `spec.records[].srv` | `SrvData` |  |  |  |
| `spec.records[].srv.priority` | `uint32` |  |  |  |
| `spec.records[].srv.weight` | `uint32` |  |  |  |
| `spec.records[].srv.port` | `uint32` |  |  |  |
| `spec.records[].srv.target` | `string \| valueFrom` | yes |  |  |
| `spec.records[].sshfp` | `SshfpData` |  |  |  |
| `spec.records[].sshfp.algorithm` | `uint32` |  |  |  |
| `spec.records[].sshfp.type` | `uint32` |  |  |  |
| `spec.records[].sshfp.fingerprint` | `string` | yes |  |  |
| `spec.records[].svcb` | `SvcbData` |  |  |  |
| `spec.records[].svcb.priority` | `uint32` |  |  |  |
| `spec.records[].svcb.target` | `string \| valueFrom` | yes |  |  |
| `spec.records[].svcb.value` | `string` |  |  |  |
| `spec.records[].tlsa` | `TlsaData` |  |  |  |
| `spec.records[].tlsa.usage` | `uint32` |  |  |  |
| `spec.records[].tlsa.selector` | `uint32` |  |  |  |
| `spec.records[].tlsa.matchingType` | `uint32` |  |  |  |
| `spec.records[].tlsa.certificate` | `string` | yes |  |  |
| `spec.records[].uri` | `UriData` |  |  |  |
| `spec.records[].uri.priority` | `uint32` |  |  |  |
| `spec.records[].uri.weight` | `uint32` |  |  |  |
| `spec.records[].uri.target` | `string \| valueFrom` | yes |  |  |
| `spec.records[].tags` | `[]string` |  |  |  |
| `spec.records[].settings` | `CloudflareDnsZoneRecordSettings` |  |  |  |
| `spec.records[].settings.ipv4Only` | `bool` |  |  |  |
| `spec.records[].settings.ipv6Only` | `bool` |  |  |  |
| `spec.records[].settings.flattenCname` | `bool` |  |  |  |
| `spec.records[].privateRouting` | `bool` |  |  |  |
| `spec.type` | `enum` |  |  |  |
| `spec.vanityNameServers` | `[]string` |  |  |  |
| `spec.dnsSettings` | `CloudflareDnsZoneDnsSettings` |  |  |  |
| `spec.dnsSettings.flattenAllCnames` | `bool` |  |  |  |
| `spec.dnsSettings.foundationDns` | `bool` |  |  |  |
| `spec.dnsSettings.multiProvider` | `bool` |  |  |  |
| `spec.dnsSettings.secondaryOverrides` | `bool` |  |  |  |
| `spec.dnsSettings.nsTtl` | `uint32` |  |  |  |
| `spec.dnsSettings.zoneMode` | `enum` |  |  |  |
| `spec.dnsSettings.soa` | `CloudflareDnsZoneSoa` |  |  |  |
| `spec.dnsSettings.soa.expire` | `uint32` |  |  |  |
| `spec.dnsSettings.soa.minTtl` | `uint32` |  |  |  |
| `spec.dnsSettings.soa.mname` | `string` |  |  |  |
| `spec.dnsSettings.soa.refresh` | `uint32` |  |  |  |
| `spec.dnsSettings.soa.retry` | `uint32` |  |  |  |
| `spec.dnsSettings.soa.rname` | `string` |  |  |  |
| `spec.dnsSettings.soa.ttl` | `uint32` |  |  |  |
| `spec.dnsSettings.nameservers` | `CloudflareDnsZoneNameservers` |  |  |  |
| `spec.dnsSettings.nameservers.nsSet` | `uint32` |  |  |  |
| `spec.dnsSettings.nameservers.type` | `string` |  |  |  |
| `spec.dnsSettings.internalDns` | `CloudflareDnsZoneInternalDns` |  |  |  |
| `spec.dnsSettings.internalDns.referenceZoneId` | `string \| valueFrom` |  |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.dnssec` | `CloudflareDnsZoneDnssec` |  |  |  |
| `spec.dnssec.enabled` | `bool` |  |  |  |
| `spec.dnssec.multiSigner` | `bool` |  |  |  |
| `spec.dnssec.presigned` | `bool` |  |  |  |
| `spec.dnssec.useNsec3` | `bool` |  |  |  |
| `spec.hold` | `CloudflareDnsZoneHold` |  |  |  |
| `spec.hold.enabled` | `bool` |  |  |  |
| `spec.hold.includeSubdomains` | `bool` |  |  |  |
| `spec.hold.holdAfter` | `string` |  |  |  |
| `spec.subscription` | `CloudflareDnsZoneSubscription` |  |  |  |
| `spec.subscription.ratePlan` | `string` |  |  |  |
| `spec.subscription.frequency` | `string` |  |  |  |
| `spec.subscription.scope` | `string` |  |  |  |

## Field Details

### spec.zoneName

`string` · required

The fully qualified domain name of the DNS zone (e.g., "example.com").

- rule: zone_name must be a valid fully qualified domain name
- rule: {"required":true}

### spec.accountId

`string` · required

The Cloudflare account identifier under which to create the zone.

- rule: {"required":true}

### spec.paused

`bool`

Whether the zone is created paused. A paused zone uses Cloudflare DNS only
and receives no security or performance (proxy/CDN/WAF) benefits.

### spec.records

`[]CloudflareDnsZoneRecord`

DNS records managed alongside this zone. For records with independent
lifecycles (or that need the full record feature set), prefer standalone
CloudflareDnsRecord resources instead.

- rule: proxied can only be true for A, AAAA, or CNAME records
- rule: priority is required for MX records
- rule: set exactly one of content (simple records) or a data block (structured records)
- rule: this record type requires its structured data block (e.g. SRV needs data.srv, CAA needs data.caa)
- rule: the supplied data block does not match the record type

### spec.records[].name

`string` · required

The record name (or "@" for the zone apex).

- rule: {"required":true}

### spec.records[].type

`enum` · required

The DNS record type, which determines whether the value comes from `content`
or from the matching `data` block.

- rule: type must be specified (cannot be record_type_unspecified)
- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `record_type_unspecified` -- Unspecified record type (invalid).
- `A` -- IPv4 address record. Uses `content`.
- `AAAA` -- IPv6 address record. Uses `content`.
- `CNAME` -- Canonical name (alias) record. Uses `content`.
- `MX` -- Mail exchange record. Uses `content` (mail host) plus `priority`.
- `TXT` -- Text record (SPF, DKIM, verification, etc.). Uses `content`.
- `SRV` -- Service locator record. Uses `data.srv`.
- `NS` -- Nameserver record. Uses `content`.
- `CAA` -- Certification Authority Authorization record. Uses `data.caa`.
- `PTR` -- Pointer record (reverse DNS). Uses `content`.
- `OPENPGPKEY` -- OpenPGP public-key record. Uses `content` (base64 OpenPGP key).
- `CERT` -- Certificate record. Uses `data.cert`.
- `DNSKEY` -- DNSSEC public key record. Uses `data.dnskey`.
- `DS` -- Delegation Signer record. Uses `data.ds`.
- `HTTPS` -- HTTPS service binding record. Uses `data.https`.
- `LOC` -- Location record. Uses `data.loc`.
- `NAPTR` -- Naming Authority Pointer record. Uses `data.naptr`.
- `SMIMEA` -- S/MIME certificate association record. Uses `data.smimea`.
- `SSHFP` -- SSH public key fingerprint record. Uses `data.sshfp`.
- `SVCB` -- Service binding record. Uses `data.svcb`.
- `TLSA` -- TLSA (DANE) certificate association record. Uses `data.tlsa`.
- `URI` -- Uniform Resource Identifier record. Uses `data.uri`.

### spec.records[].content

`string | valueFrom`

Presentation-format value for simple record types. Set this for A/AAAA/
CNAME/MX/NS/PTR/TXT/OPENPGPKEY; leave it unset for structured types (use
`data`). For A: IPv4 (e.g. "192.0.2.1"). For AAAA: IPv6 (e.g.
"2001:db8::1"). For CNAME/MX/NS/PTR: a hostname. For TXT: the text value.
For OPENPGPKEY: the base64-encoded key.

A literal or a reference to another resource's output, so a record can
publish a value that only exists once that resource does: one of a zone's
name servers for a subdomain delegation (a CloudflareDnsZone's, an
AwsRoute53Zone's or a GcpDnsZone's status.outputs.nameservers.0, .1, ...),
or a validation target (an Auth0CustomDomain's status.outputs.dns_record_value,
a GcpCertManagerDnsAuthorization's status.outputs.dns_record_data). A
referenced record is created after the resource it reads.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.records[].proxied

`bool`

Whether the record is proxied through Cloudflare (orange cloud). Only valid
for A, AAAA, and CNAME records.

### spec.records[].ttl

`int32`

Time to live (TTL) in seconds. Leave 0 or set 1 for automatic; otherwise
30-86400.

- rule: ttl must be 0 or 1 (automatic) or between 30 and 86400 seconds

### spec.records[].priority

`int32`

Priority for MX records (lower is preferred). Required for MX; ignored for
other types (SRV/URI/HTTPS/SVCB carry their own priority inside `data`).
Declare priority in exactly one place: for SRV/URI the modules mirror the
data-block priority into the provider's top-level field themselves,
because Cloudflare reflects it there on read (live-measured; omitting the
mirror re-plans forever).

- rule: priority must be between 0 and 65535

### spec.records[].comment

`string`

Optional comment/note for the record. Has no effect on DNS responses.

### spec.records[].caa

`CaaData`

CAA record data.

### spec.records[].caa.flags

`uint32`

Flags for the CAA record (0-255). 128 marks the property critical.

- rule: {"uint32":{"lte":255}}

### spec.records[].caa.tag

`string` · required

The property controlled by this record: "issue", "issuewild", or "iodef".

- rule: {"required":true}

### spec.records[].caa.value

`string` · required

The value for the tag (e.g. a CA domain like "letsencrypt.org", or an
"iodef" reporting URL).

- rule: {"required":true}

### spec.records[].cert

`CertData`

CERT record data.

### spec.records[].cert.type

`uint32`

Certificate type (e.g. 1 = PKIX).

- rule: {"uint32":{"lte":65535}}

### spec.records[].cert.keyTag

`uint32`

Key tag identifying the certificate's key (0-65535).

- rule: {"uint32":{"lte":65535}}

### spec.records[].cert.algorithm

`uint32`

Algorithm code (0-255).

- rule: {"uint32":{"lte":255}}

### spec.records[].cert.certificate

`string` · required

The base64-encoded certificate or CRL.

- rule: {"required":true}

### spec.records[].dnskey

`DnskeyData`

DNSKEY record data.

### spec.records[].dnskey.flags

`uint32`

Flags for the DNSKEY record (e.g. 256 = ZSK, 257 = KSK).

- rule: {"uint32":{"lte":65535}}

### spec.records[].dnskey.protocol

`uint32`

Protocol field; must be 3 for DNSSEC.

- rule: {"uint32":{"lte":255}}

### spec.records[].dnskey.algorithm

`uint32`

Algorithm code (0-255).

- rule: {"uint32":{"lte":255}}

### spec.records[].dnskey.publicKey

`string` · required

The base64-encoded public key.

- rule: {"required":true}

### spec.records[].ds

`DsData`

DS record data.

### spec.records[].ds.keyTag

`uint32`

Key tag of the referenced DNSKEY (0-65535).

- rule: {"uint32":{"lte":65535}}

### spec.records[].ds.algorithm

`uint32`

Algorithm code of the referenced DNSKEY (0-255).

- rule: {"uint32":{"lte":255}}

### spec.records[].ds.digestType

`uint32`

Digest algorithm type (0-255, e.g. 2 = SHA-256).

- rule: {"uint32":{"lte":255}}

### spec.records[].ds.digest

`string` · required

Hex-encoded digest of the referenced DNSKEY.

- rule: {"required":true}

### spec.records[].https

`HttpsData`

HTTPS record data.

### spec.records[].https.priority

`uint32`

Priority (0-65535). 0 selects AliasMode; higher values are ServiceMode
records evaluated in ascending priority order.

- rule: {"uint32":{"lte":65535}}

### spec.records[].https.target

`string | valueFrom` · required

Target hostname (".": the owner name; or a specific endpoint).
A literal or a reference to another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.records[].https.value

`string`

SvcParams string (e.g. "alpn=\"h2,h3\" port=8443").

### spec.records[].loc

`LocData`

LOC record data.

### spec.records[].loc.latDirection

`string`

Latitude direction: "N" or "S".

- rule: {"string":{"in":["N","S"]}}

### spec.records[].loc.latDegrees

`uint32`

Degrees of latitude (0-90).

- rule: {"uint32":{"lte":90}}

### spec.records[].loc.latMinutes

`uint32`

Minutes of latitude (0-59).

- rule: {"uint32":{"lte":59}}

### spec.records[].loc.latSeconds

`double`

Seconds of latitude (0-59.999).

- rule: {"double":{"lte":59.999,"gte":0}}

### spec.records[].loc.longDirection

`string`

Longitude direction: "E" or "W".

- rule: {"string":{"in":["E","W"]}}

### spec.records[].loc.longDegrees

`uint32`

Degrees of longitude (0-180).

- rule: {"uint32":{"lte":180}}

### spec.records[].loc.longMinutes

`uint32`

Minutes of longitude (0-59).

- rule: {"uint32":{"lte":59}}

### spec.records[].loc.longSeconds

`double`

Seconds of longitude (0-59.999).

- rule: {"double":{"lte":59.999,"gte":0}}

### spec.records[].loc.altitude

`double`

Altitude in meters (-100000 to 42849672.95).

- rule: {"double":{"lte":42849672.95,"gte":-100000}}

### spec.records[].loc.size

`double`

Diameter of a sphere enclosing the location, in meters (0-90000000).

- rule: {"double":{"lte":90000000,"gte":0}}

### spec.records[].loc.precisionHorz

`double`

Horizontal precision in meters (0-90000000).

- rule: {"double":{"lte":90000000,"gte":0}}

### spec.records[].loc.precisionVert

`double`

Vertical precision in meters (0-90000000).

- rule: {"double":{"lte":90000000,"gte":0}}

### spec.records[].naptr

`NaptrData`

NAPTR record data.

### spec.records[].naptr.order

`uint32`

Order in which records are processed (0-65535).

- rule: {"uint32":{"lte":65535}}

### spec.records[].naptr.preference

`uint32`

Preference among records with the same order (0-65535).

- rule: {"uint32":{"lte":65535}}

### spec.records[].naptr.flags

`string`

Flags controlling rewrite behavior (e.g. "U", "S", "A", "P").

### spec.records[].naptr.service

`string`

The service parameters (e.g. "E2U+sip").

### spec.records[].naptr.regex

`string`

The substitution regular expression.

### spec.records[].naptr.replacement

`string`

The replacement domain name (or "." when regex is used).

### spec.records[].smimea

`SmimeaData`

SMIMEA record data.

### spec.records[].smimea.usage

`uint32`

Certificate usage (0-255).

- rule: {"uint32":{"lte":255}}

### spec.records[].smimea.selector

`uint32`

Selector specifying which part of the certificate is matched (0-255).

- rule: {"uint32":{"lte":255}}

### spec.records[].smimea.matchingType

`uint32`

Matching type for the association data (0-255).

- rule: {"uint32":{"lte":255}}

### spec.records[].smimea.certificate

`string` · required

Hex-encoded certificate association data.

- rule: {"required":true}

### spec.records[].srv

`SrvData`

SRV record data.

### spec.records[].srv.priority

`uint32`

Priority of the target host (0-65535, lower preferred).

- rule: {"uint32":{"lte":65535}}

### spec.records[].srv.weight

`uint32`

Relative weight among targets with the same priority (0-65535).

- rule: {"uint32":{"lte":65535}}

### spec.records[].srv.port

`uint32`

TCP/UDP port of the service (0-65535).

- rule: {"uint32":{"lte":65535}}

### spec.records[].srv.target

`string | valueFrom` · required

Hostname of the machine providing the service.
A literal or a reference to another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.records[].sshfp

`SshfpData`

SSHFP record data.

### spec.records[].sshfp.algorithm

`uint32`

Algorithm of the SSH public key (0-255, e.g. 1 = RSA, 4 = Ed25519).

- rule: {"uint32":{"lte":255}}

### spec.records[].sshfp.type

`uint32`

Fingerprint hash type (0-255, e.g. 2 = SHA-256).

- rule: {"uint32":{"lte":255}}

### spec.records[].sshfp.fingerprint

`string` · required

Hex-encoded fingerprint of the SSH public key.

- rule: {"required":true}

### spec.records[].svcb

`SvcbData`

SVCB record data.

### spec.records[].svcb.priority

`uint32`

Priority (0-65535). 0 selects AliasMode; higher values are ServiceMode.

- rule: {"uint32":{"lte":65535}}

### spec.records[].svcb.target

`string | valueFrom` · required

Target hostname (".": the owner name; or a specific endpoint).
A literal or a reference to another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.records[].svcb.value

`string`

SvcParams string (e.g. "alpn=\"h2\" ipv4hint=\"192.0.2.1\"").

### spec.records[].tlsa

`TlsaData`

TLSA record data.

### spec.records[].tlsa.usage

`uint32`

Certificate usage (0-255).

- rule: {"uint32":{"lte":255}}

### spec.records[].tlsa.selector

`uint32`

Selector specifying which part of the certificate is matched (0-255).

- rule: {"uint32":{"lte":255}}

### spec.records[].tlsa.matchingType

`uint32`

Matching type for the association data (0-255).

- rule: {"uint32":{"lte":255}}

### spec.records[].tlsa.certificate

`string` · required

Hex-encoded certificate association data.

- rule: {"required":true}

### spec.records[].uri

`UriData`

URI record data.

### spec.records[].uri.priority

`uint32`

Priority of the target URI (0-65535, lower preferred).

- rule: {"uint32":{"lte":65535}}

### spec.records[].uri.weight

`uint32`

Relative weight among URIs with the same priority (0-65535).

- rule: {"uint32":{"lte":65535}}

### spec.records[].uri.target

`string | valueFrom` · required

The target URI (e.g. "https://example.com/path").
A literal or a reference to another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.records[].tags

`[]string`

Custom tags for the record. Have no effect on DNS responses; useful for
organizing and filtering records. Entitlement-gated: accounts without the
record-tags feature have a tag quota of 0 and the record create fails
with 400 code 9300 (measured live on a free-plan account).

### spec.records[].settings

`CloudflareDnsZoneRecordSettings`

Optional record-level settings controlling how proxied records are served.
Entitlement-gated: on zones without the feature, setting ipv4_only or
ipv6_only fails the record create with 400 code 9227 "not available to
this zone" (measured live on a free-plan zone).

### spec.records[].settings.ipv4Only

`bool`

When enabled, only A records are generated and AAAA records are suppressed.
For exceptional cases; applies only to proxied records.

### spec.records[].settings.ipv6Only

`bool`

When enabled, only AAAA records are generated and A records are suppressed.
For exceptional cases; applies only to proxied records.

### spec.records[].settings.flattenCname

`bool`

When enabled, a CNAME is resolved externally and the resulting address
records are returned instead of the CNAME. Unavailable for proxied records
(which are always flattened).

### spec.records[].privateRouting

`bool`

Whether the record is restricted to Cloudflare's internal (private) routing
and not served over the public internet — used for internal DNS / Magic WAN
scenarios. Defaults to false (public).

### spec.type

`enum`

The zone's deployment type. Defaults to "full".

Allowed values (use exactly as shown):

- `zone_type_unspecified`
- `full`
- `partial`
- `secondary`
- `internal`

### spec.vanityNameServers

`[]string`

Custom (vanity) name servers for the zone. Only available on Business and
Enterprise plans; leave empty to use Cloudflare's assigned name servers.

### spec.dnsSettings

`CloudflareDnsZoneDnsSettings`

Optional zone-wide DNS settings (CNAME flattening, zone mode, SOA, etc.).
Omit to leave Cloudflare's defaults in place. The non-default values are
individually entitlement-gated: on an account/zone without the matching
feature, Cloudflare rejects the write with 400 code 1003 "not available
to this account or zone" (measured live for SOA tuning, ns_ttl, and
flatten_all_cnames). On the terraform engine also note the v5.23.0
provider echoes server defaults into unset optional fields on refresh,
so a partial dns_settings block re-plans forever -- declare every field
the server echoes, or omit the block.

### spec.dnsSettings.flattenAllCnames

`bool`

Flatten all CNAME records in the zone (a CNAME at the apex is always flattened).

### spec.dnsSettings.foundationDns

`bool`

Enable Foundation DNS Advanced Nameservers on the zone.

### spec.dnsSettings.multiProvider

`bool`

Enable multi-provider DNS (activates the zone even with non-Cloudflare NS
records and respects apex NS records during outbound transfers).

### spec.dnsSettings.secondaryOverrides

`bool`

Allow a secondary zone to use proxied override records and apex CNAME
flattening.

### spec.dnsSettings.nsTtl

`uint32`

Time to live (TTL), in seconds, for the zone's nameserver (NS) records.
Leave 0 to use the default; otherwise 30-86400.

- rule: ns_ttl must be 0 (default) or between 30 and 86400 seconds

### spec.dnsSettings.zoneMode

`enum`

Whether the zone is a standard zone or a CDN/DNS-only zone.

Allowed values (use exactly as shown):

- `zone_mode_unspecified`
- `standard`
- `cdn_only`
- `dns_only`

### spec.dnsSettings.soa

`CloudflareDnsZoneSoa`

Components of the zone's SOA record.

### spec.dnsSettings.soa.expire

`uint32`

Seconds a secondary will keep serving the zone after losing contact with the
primary (86400-2419200).

- rule: expire must be 0 (default) or between 86400 and 2419200 seconds

### spec.dnsSettings.soa.minTtl

`uint32`

TTL for negative caching of records within the zone (60-86400).

- rule: min_ttl must be 0 (default) or between 60 and 86400 seconds

### spec.dnsSettings.soa.mname

`string`

The primary nameserver (MNAME). Leave empty for a Cloudflare-assigned value.

### spec.dnsSettings.soa.refresh

`uint32`

Seconds after which a secondary re-checks the SOA for updates (600-86400).

- rule: refresh must be 0 (default) or between 600 and 86400 seconds

### spec.dnsSettings.soa.retry

`uint32`

Seconds a secondary waits before retrying after an unresponsive primary
(600-86400).

- rule: retry must be 0 (default) or between 600 and 86400 seconds

### spec.dnsSettings.soa.rname

`string`

The zone administrator's email (RNAME), first label being the local part.

### spec.dnsSettings.soa.ttl

`uint32`

TTL of the SOA record itself (300-86400).

- rule: ttl must be 0 (default) or between 300 and 86400 seconds

### spec.dnsSettings.nameservers

`CloudflareDnsZoneNameservers`

Settings determining the nameservers through which the zone is available.

### spec.dnsSettings.nameservers.nsSet

`uint32`

Configured nameserver set to use for this zone (1-5).

- rule: ns_set must be 0 (default) or between 1 and 5

### spec.dnsSettings.nameservers.type

`string`

Nameserver type: one of "cloudflare.standard", "custom.account",
"custom.tenant", "custom.zone".

- rule: type must be one of "cloudflare.standard", "custom.account", "custom.tenant", "custom.zone"

### spec.dnsSettings.internalDns

`CloudflareDnsZoneInternalDns`

Settings for an internal zone (only relevant when type is "internal").

### spec.dnsSettings.internalDns.referenceZoneId

`string | valueFrom`

The zone to fall back to for resolution. Can be a literal zone ID or a
reference to another CloudflareDnsZone.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.dnssec

`CloudflareDnsZoneDnssec`

Optional DNSSEC configuration. Enable to have Cloudflare sign the zone; the
DS record material to hand to your registrar is published as stack outputs.
DNSSEC activates only on an ACTIVE (registrar-delegated) zone: on a
PENDING zone Cloudflare rejects the enable with 400 code 1017 "Invalid
zone plan for action" (measured live; the same call succeeds on an
active free-plan zone -- the wall is delegation state, not plan tier).

### spec.dnssec.enabled

`bool`

Whether DNSSEC is active for the zone.

### spec.dnssec.multiSigner

`bool`

Enable multi-signer DNSSEC, allowing multiple providers to serve a
DNSSEC-signed zone simultaneously (required to add external DNSKEY records).

### spec.dnssec.presigned

`bool`

Allow transferring in an externally-signed (presigned) DNSSEC zone without
Cloudflare signing records on the fly (for Cloudflare-as-secondary setups).

### spec.dnssec.useNsec3

`bool`

Use NSEC3 rather than NSEC for authenticated denial of existence.

### spec.hold

`CloudflareDnsZoneHold`

Optional zone hold. While enabled, Cloudflare blocks this zone's hostname
(and optionally its subdomains) from being created as a zone in any other
Cloudflare account — protection against hostname takeover during
migrations or account consolidation.

### spec.hold.enabled

`bool`

Whether the hold is active. Removing the hold (or this whole block) lifts
the protection.

### spec.hold.includeSubdomains

`bool`

Extend the hold to all subdomains of the zone's hostname, not just the
apex hostname itself.

### spec.hold.holdAfter

`string`

RFC3339 timestamp (e.g. "2026-01-31T00:00:00Z"). When future-dated, the
hold is temporarily disabled and automatically re-enabled by Cloudflare at
this time — a planned migration window. A past-dated value has no effect
on an enabled hold. Leave empty for an immediate, indefinite hold.

- rule: hold_after must be an RFC3339 timestamp (e.g. 2026-01-31T00:00:00Z) or empty

### spec.subscription

`CloudflareDnsZoneSubscription`

Optional zone plan subscription. Omit to stay on the account's default
(free) plan. Setting a paid rate plan creates a real Cloudflare
subscription with real billing; the API token needs Billing Write scope.

- rule: set rate_plan and/or frequency — an empty subscription block has no effect

### spec.subscription.ratePlan

`string`

The rate plan to subscribe the zone to. Cloudflare's plan identifiers:
"free", "lite", "pro", "pro_plus", "business", "enterprise", plus the
partner variants ("partners_free", "partners_pro", "partners_business",
"partners_enterprise", "partners_ent").

- rule: rate_plan must be one of Cloudflare's plan identifiers (free, lite, pro, pro_plus, business, enterprise, partners_free, partners_pro, partners_business, partners_enterprise, partners_ent)

### spec.subscription.frequency

`string`

How often the subscription renews automatically. Leave empty for
Cloudflare's default; some plans ignore frequency entirely.

- rule: frequency must be one of "weekly", "monthly", "quarterly", "yearly"

### spec.subscription.scope

`string`

The scope the rate plan applies to. Rarely needed; leave empty for
Cloudflare's default scope for the chosen plan.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareDnsZone, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.zone_id` | `string` | The Cloudflare Zone ID of the created DNS zone. |
| `status.outputs.nameservers` | `[]string` | The list of nameserver addresses assigned to this DNS zone. |
| `status.outputs.status` | `string` | The zone status on Cloudflare (e.g., "initializing", "pending", "active", "moved"). |
| `status.outputs.dnssec_status` | `string` | DNSSEC status (e.g., "active", "disabled"). Empty when DNSSEC is not enabled. |
| `status.outputs.dnssec_ds` | `string` | The full DS record to enter at your registrar. Empty unless DNSSEC is enabled. |
| `status.outputs.dnssec_digest` | `string` | The DS record digest. Empty unless DNSSEC is enabled. |
| `status.outputs.dnssec_digest_type` | `string` | The DS digest type code (e.g., "2" for SHA-256). Empty unless DNSSEC is enabled. |
| `status.outputs.dnssec_digest_algorithm` | `string` | The DS digest algorithm name. Empty unless DNSSEC is enabled. |
| `status.outputs.dnssec_algorithm` | `string` | The DNSKEY algorithm code. Empty unless DNSSEC is enabled. |
| `status.outputs.dnssec_key_tag` | `string` | The DNSKEY key tag. Empty unless DNSSEC is enabled. |
| `status.outputs.dnssec_public_key` | `string` | The DNSKEY public key. Empty unless DNSSEC is enabled. |
| `status.outputs.dnssec_flags` | `string` | The DNSKEY flags. Empty unless DNSSEC is enabled. |
| `status.outputs.record_ids` | `map<string, string>` | Cloudflare-assigned ids of the zone's inline DNS records, keyed by the records' name-type-index key (both engines key record instances the same way). Import recipes derive each record's {zone_id}/{dns_record_id} import ID from this map; empty when the spec declares no inline records. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.dnsSettings.internalDns.referenceZoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflareAuthenticatedOriginPulls | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareAuthenticatedOriginPullsCertificate | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareBotManagement | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareCacheSettings | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareCertificatePack | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareCustomHostname | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareCustomHostnameFallbackOrigin | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareCustomSslCertificate | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareDnsRecord | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareDnsZone | `spec.dnsSettings.internalDns.referenceZoneId` | `status.outputs.zone_id` |
| CloudflareEmailRoutingRule | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareEmailRoutingZone | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareHealthcheck | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareIpAccessRule | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareLoadBalancer | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareLogpushJob | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareR2Bucket | `spec.customDomains[].zoneId` | `status.outputs.zone_id` |
| CloudflareRuleset | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareSnippet | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareSnippetRules | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareWaitingRoom | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareWaitingRoomEvent | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareWebAnalyticsSite | `spec.zoneTag` | `status.outputs.zone_id` |
| CloudflareWorker | `spec.customDomains[].zoneId` | `status.outputs.zone_id` |
| CloudflareWorker | `spec.routes[].zoneId` | `status.outputs.zone_id` |
| CloudflareZeroTrustAccessApplication | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareZeroTrustAccessGroup | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareZeroTrustAccessIdentityProvider | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareZeroTrustAccessServiceToken | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareZeroTrustDeviceDefaultProfile | `spec.zoneCertificates.zoneId` | `status.outputs.zone_id` |
| CloudflareZeroTrustOrganization | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareZoneSettings | `spec.zoneId` | `status.outputs.zone_id` |
| CloudflareZoneTlsSettings | `spec.zoneId` | `status.outputs.zone_id` |
| KubernetesExternalDns | `spec.cloudflare.zoneIdFilters` | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
