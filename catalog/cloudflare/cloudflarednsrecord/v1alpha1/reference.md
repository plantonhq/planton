# CloudflareDnsRecord

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareDnsRecordSpec defines a single DNS record in a Cloudflare zone.

A record is either a "simple" record whose value is a presentation-format
string in `content` (A, AAAA, CNAME, MX, NS, PTR, TXT, OPENPGPKEY) or a
"structured" record whose fields are supplied through the typed `data` oneof
(CAA, CERT, DNSKEY, DS, HTTPS, LOC, NAPTR, SMIMEA, SRV, SSHFP, SVCB, TLSA,
URI). Exactly one of `content` or a `data` case is set, and the chosen
representation must match `type`.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareDnsRecord
metadata:
  name: test-dns-record
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: "www"
  type: A
  content:
    value: "192.0.2.1"
  proxied: true
  ttl: 1
  comment: "Test DNS record for local development"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.type` | `enum` | yes |  |  |
| `spec.content` | `string \| valueFrom` |  |  |  |
| `spec.proxied` | `bool` |  |  |  |
| `spec.ttl` | `int32` |  |  |  |
| `spec.priority` | `int32` |  |  |  |
| `spec.comment` | `string` |  |  |  |
| `spec.caa` | `CaaData` |  |  |  |
| `spec.caa.flags` | `uint32` |  |  |  |
| `spec.caa.tag` | `string` | yes |  |  |
| `spec.caa.value` | `string` | yes |  |  |
| `spec.cert` | `CertData` |  |  |  |
| `spec.cert.type` | `uint32` |  |  |  |
| `spec.cert.keyTag` | `uint32` |  |  |  |
| `spec.cert.algorithm` | `uint32` |  |  |  |
| `spec.cert.certificate` | `string` | yes |  |  |
| `spec.dnskey` | `DnskeyData` |  |  |  |
| `spec.dnskey.flags` | `uint32` |  |  |  |
| `spec.dnskey.protocol` | `uint32` |  |  |  |
| `spec.dnskey.algorithm` | `uint32` |  |  |  |
| `spec.dnskey.publicKey` | `string` | yes |  |  |
| `spec.ds` | `DsData` |  |  |  |
| `spec.ds.keyTag` | `uint32` |  |  |  |
| `spec.ds.algorithm` | `uint32` |  |  |  |
| `spec.ds.digestType` | `uint32` |  |  |  |
| `spec.ds.digest` | `string` | yes |  |  |
| `spec.https` | `HttpsData` |  |  |  |
| `spec.https.priority` | `uint32` |  |  |  |
| `spec.https.target` | `string \| valueFrom` | yes |  |  |
| `spec.https.value` | `string` |  |  |  |
| `spec.loc` | `LocData` |  |  |  |
| `spec.loc.latDirection` | `string` |  |  |  |
| `spec.loc.latDegrees` | `uint32` |  |  |  |
| `spec.loc.latMinutes` | `uint32` |  |  |  |
| `spec.loc.latSeconds` | `double` |  |  |  |
| `spec.loc.longDirection` | `string` |  |  |  |
| `spec.loc.longDegrees` | `uint32` |  |  |  |
| `spec.loc.longMinutes` | `uint32` |  |  |  |
| `spec.loc.longSeconds` | `double` |  |  |  |
| `spec.loc.altitude` | `double` |  |  |  |
| `spec.loc.size` | `double` |  |  |  |
| `spec.loc.precisionHorz` | `double` |  |  |  |
| `spec.loc.precisionVert` | `double` |  |  |  |
| `spec.naptr` | `NaptrData` |  |  |  |
| `spec.naptr.order` | `uint32` |  |  |  |
| `spec.naptr.preference` | `uint32` |  |  |  |
| `spec.naptr.flags` | `string` |  |  |  |
| `spec.naptr.service` | `string` |  |  |  |
| `spec.naptr.regex` | `string` |  |  |  |
| `spec.naptr.replacement` | `string` |  |  |  |
| `spec.smimea` | `SmimeaData` |  |  |  |
| `spec.smimea.usage` | `uint32` |  |  |  |
| `spec.smimea.selector` | `uint32` |  |  |  |
| `spec.smimea.matchingType` | `uint32` |  |  |  |
| `spec.smimea.certificate` | `string` | yes |  |  |
| `spec.srv` | `SrvData` |  |  |  |
| `spec.srv.priority` | `uint32` |  |  |  |
| `spec.srv.weight` | `uint32` |  |  |  |
| `spec.srv.port` | `uint32` |  |  |  |
| `spec.srv.target` | `string \| valueFrom` | yes |  |  |
| `spec.sshfp` | `SshfpData` |  |  |  |
| `spec.sshfp.algorithm` | `uint32` |  |  |  |
| `spec.sshfp.type` | `uint32` |  |  |  |
| `spec.sshfp.fingerprint` | `string` | yes |  |  |
| `spec.svcb` | `SvcbData` |  |  |  |
| `spec.svcb.priority` | `uint32` |  |  |  |
| `spec.svcb.target` | `string \| valueFrom` | yes |  |  |
| `spec.svcb.value` | `string` |  |  |  |
| `spec.tlsa` | `TlsaData` |  |  |  |
| `spec.tlsa.usage` | `uint32` |  |  |  |
| `spec.tlsa.selector` | `uint32` |  |  |  |
| `spec.tlsa.matchingType` | `uint32` |  |  |  |
| `spec.tlsa.certificate` | `string` | yes |  |  |
| `spec.uri` | `UriData` |  |  |  |
| `spec.uri.priority` | `uint32` |  |  |  |
| `spec.uri.weight` | `uint32` |  |  |  |
| `spec.uri.target` | `string \| valueFrom` | yes |  |  |
| `spec.tags` | `[]string` |  |  |  |
| `spec.settings` | `CloudflareDnsRecordSettings` |  |  |  |
| `spec.settings.ipv4Only` | `bool` |  |  |  |
| `spec.settings.ipv6Only` | `bool` |  |  |  |
| `spec.settings.flattenCname` | `bool` |  |  |  |
| `spec.privateRouting` | `bool` |  |  |  |

## Field Details

### spec.zoneId

`string | valueFrom` · required

The Cloudflare Zone ID where this DNS record will be created. Can be a
literal value or a reference to a CloudflareDnsZone resource; when referenced,
it defaults to that zone's status.outputs.zone_id.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.name

`string` · required

The record name (or "@" for the zone apex). May be a subdomain label
("www", "api") or a fully qualified name within the zone.

- rule: {"required":true}

### spec.type

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

### spec.content

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

### spec.proxied

`bool`

Whether the record is proxied through Cloudflare (orange cloud). When true,
traffic flows through Cloudflare's CDN/WAF and the origin IP is hidden; when
false, the record is DNS-only (grey cloud). Only valid for A, AAAA, CNAME.

### spec.ttl

`int32`

Time to live (TTL) in seconds. Leave 0 or set 1 for "automatic" (recommended
for proxied records). Otherwise 30-86400 (the 30s floor applies to Enterprise
zones; most zones use 60s and up). Defaults to automatic.

- rule: ttl must be 0 or 1 (automatic) or between 30 and 86400 seconds

### spec.priority

`int32`

Priority for MX records (lower is preferred). Required for MX; never set it
for other types (SRV/URI/HTTPS/SVCB carry their own priority inside their
typed data). Cloudflare's API mirrors an SRV/URI data priority into the
record's top-level priority field on its own; the IaC modules send that
mirror so the deployed record never drifts -- authors declare priority in
exactly one place.

- rule: priority must be between 0 and 65535

### spec.comment

`string`

Optional comment/note for the record. Has no effect on DNS responses; used
to document the record's purpose.

### spec.caa

`CaaData`

CAA record data.

### spec.caa.flags

`uint32`

Flags for the CAA record (0-255). 128 marks the property critical.

- rule: {"uint32":{"lte":255}}

### spec.caa.tag

`string` · required

The property controlled by this record: "issue", "issuewild", or "iodef".

- rule: {"required":true}

### spec.caa.value

`string` · required

The value for the tag (e.g. a CA domain like "letsencrypt.org", or an
"iodef" reporting URL).

- rule: {"required":true}

### spec.cert

`CertData`

CERT record data.

### spec.cert.type

`uint32`

Certificate type (e.g. 1 = PKIX).

- rule: {"uint32":{"lte":65535}}

### spec.cert.keyTag

`uint32`

Key tag identifying the certificate's key (0-65535).

- rule: {"uint32":{"lte":65535}}

### spec.cert.algorithm

`uint32`

Algorithm code (0-255).

- rule: {"uint32":{"lte":255}}

### spec.cert.certificate

`string` · required

The base64-encoded certificate or CRL.

- rule: {"required":true}

### spec.dnskey

`DnskeyData`

DNSKEY record data.

### spec.dnskey.flags

`uint32`

Flags for the DNSKEY record (e.g. 256 = ZSK, 257 = KSK).

- rule: {"uint32":{"lte":65535}}

### spec.dnskey.protocol

`uint32`

Protocol field; must be 3 for DNSSEC.

- rule: {"uint32":{"lte":255}}

### spec.dnskey.algorithm

`uint32`

Algorithm code (0-255).

- rule: {"uint32":{"lte":255}}

### spec.dnskey.publicKey

`string` · required

The base64-encoded public key.

- rule: {"required":true}

### spec.ds

`DsData`

DS record data.

### spec.ds.keyTag

`uint32`

Key tag of the referenced DNSKEY (0-65535).

- rule: {"uint32":{"lte":65535}}

### spec.ds.algorithm

`uint32`

Algorithm code of the referenced DNSKEY (0-255).

- rule: {"uint32":{"lte":255}}

### spec.ds.digestType

`uint32`

Digest algorithm type (0-255, e.g. 2 = SHA-256).

- rule: {"uint32":{"lte":255}}

### spec.ds.digest

`string` · required

Hex-encoded digest of the referenced DNSKEY.

- rule: {"required":true}

### spec.https

`HttpsData`

HTTPS record data.

### spec.https.priority

`uint32`

Priority (0-65535). 0 selects AliasMode; higher values are ServiceMode
records evaluated in ascending priority order.

- rule: {"uint32":{"lte":65535}}

### spec.https.target

`string | valueFrom` · required

Target hostname (".": the owner name; or a specific endpoint).
A literal or a reference to another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.https.value

`string`

SvcParams string (e.g. "alpn=\"h2,h3\" port=8443").

### spec.loc

`LocData`

LOC record data.

### spec.loc.latDirection

`string`

Latitude direction: "N" or "S".

- rule: {"string":{"in":["N","S"]}}

### spec.loc.latDegrees

`uint32`

Degrees of latitude (0-90).

- rule: {"uint32":{"lte":90}}

### spec.loc.latMinutes

`uint32`

Minutes of latitude (0-59).

- rule: {"uint32":{"lte":59}}

### spec.loc.latSeconds

`double`

Seconds of latitude (0-59.999).

- rule: {"double":{"lte":59.999,"gte":0}}

### spec.loc.longDirection

`string`

Longitude direction: "E" or "W".

- rule: {"string":{"in":["E","W"]}}

### spec.loc.longDegrees

`uint32`

Degrees of longitude (0-180).

- rule: {"uint32":{"lte":180}}

### spec.loc.longMinutes

`uint32`

Minutes of longitude (0-59).

- rule: {"uint32":{"lte":59}}

### spec.loc.longSeconds

`double`

Seconds of longitude (0-59.999).

- rule: {"double":{"lte":59.999,"gte":0}}

### spec.loc.altitude

`double`

Altitude in meters (-100000 to 42849672.95).

- rule: {"double":{"lte":42849672.95,"gte":-100000}}

### spec.loc.size

`double`

Diameter of a sphere enclosing the location, in meters (0-90000000).

- rule: {"double":{"lte":90000000,"gte":0}}

### spec.loc.precisionHorz

`double`

Horizontal precision in meters (0-90000000).

- rule: {"double":{"lte":90000000,"gte":0}}

### spec.loc.precisionVert

`double`

Vertical precision in meters (0-90000000).

- rule: {"double":{"lte":90000000,"gte":0}}

### spec.naptr

`NaptrData`

NAPTR record data.

### spec.naptr.order

`uint32`

Order in which records are processed (0-65535).

- rule: {"uint32":{"lte":65535}}

### spec.naptr.preference

`uint32`

Preference among records with the same order (0-65535).

- rule: {"uint32":{"lte":65535}}

### spec.naptr.flags

`string`

Flags controlling rewrite behavior (e.g. "U", "S", "A", "P").

### spec.naptr.service

`string`

The service parameters (e.g. "E2U+sip").

### spec.naptr.regex

`string`

The substitution regular expression.

### spec.naptr.replacement

`string`

The replacement domain name (or "." when regex is used).

### spec.smimea

`SmimeaData`

SMIMEA record data.

### spec.smimea.usage

`uint32`

Certificate usage (0-255).

- rule: {"uint32":{"lte":255}}

### spec.smimea.selector

`uint32`

Selector specifying which part of the certificate is matched (0-255).

- rule: {"uint32":{"lte":255}}

### spec.smimea.matchingType

`uint32`

Matching type for the association data (0-255).

- rule: {"uint32":{"lte":255}}

### spec.smimea.certificate

`string` · required

Hex-encoded certificate association data.

- rule: {"required":true}

### spec.srv

`SrvData`

SRV record data.

### spec.srv.priority

`uint32`

Priority of the target host (0-65535, lower preferred).

- rule: {"uint32":{"lte":65535}}

### spec.srv.weight

`uint32`

Relative weight among targets with the same priority (0-65535).

- rule: {"uint32":{"lte":65535}}

### spec.srv.port

`uint32`

TCP/UDP port of the service (0-65535).

- rule: {"uint32":{"lte":65535}}

### spec.srv.target

`string | valueFrom` · required

Hostname of the machine providing the service.
A literal or a reference to another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.sshfp

`SshfpData`

SSHFP record data.

### spec.sshfp.algorithm

`uint32`

Algorithm of the SSH public key (0-255, e.g. 1 = RSA, 4 = Ed25519).

- rule: {"uint32":{"lte":255}}

### spec.sshfp.type

`uint32`

Fingerprint hash type (0-255, e.g. 2 = SHA-256).

- rule: {"uint32":{"lte":255}}

### spec.sshfp.fingerprint

`string` · required

Hex-encoded fingerprint of the SSH public key.

- rule: {"required":true}

### spec.svcb

`SvcbData`

SVCB record data.

### spec.svcb.priority

`uint32`

Priority (0-65535). 0 selects AliasMode; higher values are ServiceMode.

- rule: {"uint32":{"lte":65535}}

### spec.svcb.target

`string | valueFrom` · required

Target hostname (".": the owner name; or a specific endpoint).
A literal or a reference to another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.svcb.value

`string`

SvcParams string (e.g. "alpn=\"h2\" ipv4hint=\"192.0.2.1\"").

### spec.tlsa

`TlsaData`

TLSA record data.

### spec.tlsa.usage

`uint32`

Certificate usage (0-255).

- rule: {"uint32":{"lte":255}}

### spec.tlsa.selector

`uint32`

Selector specifying which part of the certificate is matched (0-255).

- rule: {"uint32":{"lte":255}}

### spec.tlsa.matchingType

`uint32`

Matching type for the association data (0-255).

- rule: {"uint32":{"lte":255}}

### spec.tlsa.certificate

`string` · required

Hex-encoded certificate association data.

- rule: {"required":true}

### spec.uri

`UriData`

URI record data.

### spec.uri.priority

`uint32`

Priority of the target URI (0-65535, lower preferred).

- rule: {"uint32":{"lte":65535}}

### spec.uri.weight

`uint32`

Relative weight among URIs with the same priority (0-65535).

- rule: {"uint32":{"lte":65535}}

### spec.uri.target

`string | valueFrom` · required

The target URI (e.g. "https://example.com/path").
A literal or a reference to another resource's output.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.tags

`[]string`

Custom tags for the record. Have no effect on DNS responses; useful for
organizing and filtering records.

### spec.settings

`CloudflareDnsRecordSettings`

Optional record-level settings controlling how proxied records are served.

### spec.settings.ipv4Only

`bool`

When enabled, only A records are generated and AAAA records are suppressed.
For exceptional cases; applies only to proxied records.

### spec.settings.ipv6Only

`bool`

When enabled, only AAAA records are generated and A records are suppressed.
For exceptional cases; applies only to proxied records.

### spec.settings.flattenCname

`bool`

When enabled, a CNAME is resolved externally and the resulting address
records are returned instead of the CNAME. Unavailable for proxied records
(which are always flattened).

### spec.privateRouting

`bool`

Whether the record is restricted to Cloudflare's internal (private) routing
and not served over the public internet — used for internal DNS / Magic WAN
scenarios. Defaults to false (public).

## Validation Rules

- `spec.proxied_only_for_supported_types`: proxied can only be true for A, AAAA, or CNAME records
- `spec.priority_required_for_mx`: priority is required for MX records
- `spec.content_xor_data`: set exactly one of content (simple records) or a data block (structured records)
- `spec.structured_types_require_data`: this record type requires its structured data block (e.g. SRV needs data.srv, CAA needs data.caa)
- `spec.data_block_matches_type`: the supplied data block does not match the record type

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareDnsRecord, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.record_id` | `string` | The unique identifier of the created DNS record in Cloudflare. |
| `status.outputs.record_name` | `string` | The DNS record name as declared in the spec (relative to the zone, e.g. "www" or "@" for the apex). Deliberately NOT the API's echo: Cloudflare normalizes names to the full FQDN on read, so echoing the refreshed value would flip this output after the first refresh. |
| `status.outputs.record_type` | `string` | The DNS record type that was created. |
| `status.outputs.proxied` | `bool` | Whether the record is proxied through Cloudflare (orange cloud). |
| `status.outputs.zone_id` | `string` | The Cloudflare Zone ID the record lives in. A record's API identity is (zone_id, record_id), so downstream consumers -- verification tooling, imports, chart blocks composing on the record -- need the zone alongside the record's own id. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
