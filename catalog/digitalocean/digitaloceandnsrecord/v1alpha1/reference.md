# DigitalOceanDnsRecord

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanDnsRecordSpec models the full `digitalocean_record` surface: a
single DNS record inside an existing DigitalOcean DNS zone (domain). Every
record type and per-type argument the provider accepts is representable,
with the provider's own type-conditional requirements enforced up front.

## Example

```yaml
# Example DigitalOceanDnsRecord manifests. Deploy with:
#   planton apply -f manifest.yaml
#
# Document 1 -- the smallest real record: an A record pointing a subdomain
# at an IPv4 address, with every optional field left to its default.
#
# Document 2 -- an SRV record: the type with the largest argument surface
# (priority, weight, port) plus a custom TTL.
#
# Document 3 -- a CAA record: certificate authority pinning with the
# flags/tag pair (flags 0 = non-critical, the standard value).
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDnsRecord
metadata:
  name: example-dodnsrec-a
spec:
  domain:
    value: example.com
  name: app
  type: A
  value:
    value: 203.0.113.5
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDnsRecord
metadata:
  name: example-dodnsrec-srv
spec:
  domain:
    value: example.com
  name: _sip._tcp
  type: SRV
  value:
    value: sip.example.com.
  ttlSeconds: 300
  priority: 10
  weight: 60
  port: 5060
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDnsRecord
metadata:
  name: example-dodnsrec-caa
spec:
  domain:
    value: example.com
  name: "@"
  type: CAA
  value:
    value: letsencrypt.org
  flags: 0
  tag: issue
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.domain` | `string \| valueFrom` | yes |  | DigitalOceanDnsZone (`status.outputs.zone_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.type` | `enum` | yes |  |  |
| `spec.value` | `string \| valueFrom` | yes |  |  |
| `spec.ttlSeconds` | `int32` |  | `1800` |  |
| `spec.priority` | `int32` |  |  |  |
| `spec.weight` | `int32` |  |  |  |
| `spec.port` | `int32` |  |  |  |
| `spec.flags` | `int32` |  |  |  |
| `spec.tag` | `string` |  |  |  |

## Field Details

### spec.domain

`string | valueFrom` · required

The DigitalOcean domain name (DNS zone) this record is created in, as a
literal domain or a reference to a DigitalOceanDnsZone resource.
Changing it recreates the record. Example: "example.com"

- references: DigitalOceanDnsZone (`status.outputs.zone_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_name}} -- a bare string does not parse

### spec.name

`string` · required

The record's host name, relative to the zone. Use "@" for the zone apex.
A fully-qualified form ("www.example.com") is also accepted: the
provider suppresses the diff between the short name in state and the
qualified name in configuration.
Examples: "@", "www", "api.v1"

- rule: {"required":true}

### spec.type

`enum` · required

The type of DNS record to create. Changing it recreates the record.

- rule: type must be specified (cannot be record_type_unspecified)
- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `record_type_unspecified` -- Unspecified record type (invalid).
- `A` -- IPv4 address record.
- `AAAA` -- IPv6 address record.
- `CNAME` -- Canonical name (alias) record.
- `MX` -- Mail exchange record.
- `TXT` -- Text record (SPF, DKIM, verification, etc.).
- `SRV` -- Service locator record.
- `NS` -- Nameserver record.
- `CAA` -- Certificate Authority Authorization record.
- `SOA` -- Start of authority record.

### spec.value

`string | valueFrom` · required

The value/target of the DNS record, as a literal or a reference to
another resource's output. Format depends on record type:
  - A: IPv4 address (e.g., "192.0.2.1")
  - AAAA: IPv6 address (e.g., "2001:db8::1")
  - CNAME/MX/NS/SRV: target hostname (e.g., "target.example.com.")
  - TXT: text value (e.g., "v=spf1 include:_spf.google.com ~all")
  - CAA: CA domain (e.g., "letsencrypt.org.")
Hostname values (CNAME, MX, NS, SRV, and CAA except tag=iodef) must be
written either fully qualified WITH a trailing dot
("mail.example.com.", "letsencrypt.org.") or relative to the zone
("mail"). DigitalOcean reports every such value back fully qualified
with a trailing dot, and the provider forgives only those two spellings
— a bare fully-qualified name without the dot ("letsencrypt.org") is
re-applied on every run, forever.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.ttlSeconds

`int32` · optional (explicit presence)

Time to live for the record, in seconds. When unset, the DigitalOcean
API applies its default (1800 seconds). DigitalOcean harmonizes TTLs
across records sharing a fully-qualified name (RFC 2181 §5.2), so the
live TTL can drift from this value when a sibling record changes it —
the provider surfaces that as a warning, not an error.

- default: `1800`
- rule: {"int32":{"gte":1}}

### spec.priority

`int32` · optional (explicit presence)

Priority for MX and SRV records; lower values are preferred. Required
for MX and SRV, ignored for other types. Range: 0-65535. Provider
quirk: an explicit 0 passes validation but is dropped from the create
request, so the API's default applies — use a positive value when the
priority must be exact.

- rule: {"int32":{"lte":65535,"gte":0}}

### spec.weight

`int32` · optional (explicit presence)

Weight for SRV records: the relative share of traffic among records
with equal priority. Required for SRV, ignored for other types.
Range: 0-65535. The same explicit-zero provider quirk as priority
applies.

- rule: {"int32":{"lte":65535,"gte":0}}

### spec.port

`int32` · optional (explicit presence)

Port for SRV records: where the service is available. Required for SRV,
ignored for other types. Range: 0-65535. The same explicit-zero
provider quirk as priority applies.

- rule: {"int32":{"lte":65535,"gte":0}}

### spec.flags

`int32` · optional (explicit presence)

Flags for CAA records:
  - 0: non-critical (the common value) — a CA may ignore unknown tags
  - 128: critical — a CA must refuse issuance if it does not understand
    the tag
Required for CAA, ignored for other types. Range: 0-255. An explicit 0
is dropped from the create request (provider quirk), which is harmless:
the API defaults flags to 0.

- rule: {"int32":{"lte":255,"gte":0}}

### spec.tag

`string`

Tag for CAA records — the property being authorized:
  - "issue": the CA may issue certificates for the domain
  - "issuewild": the CA may issue wildcard certificates
  - "iodef": a URL to report policy violations to
Required for CAA, ignored for other types.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["issue","issuewild","iodef"]}}

## Validation Rules

- `spec.priority_required_for_mx`: priority must be specified for MX records
- `spec.priority_weight_port_required_for_srv`: priority, weight, and port must be specified for SRV records
- `spec.flags_and_tag_required_for_caa`: flags and tag must be specified for CAA records

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanDnsRecord, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.record_id` | `string` | The unique identifier of the created DNS record. DigitalOcean assigns a NUMERIC id (exported here as its string form); together with the domain it addresses the record in the API and in imports ("{domain},{record_id}"). |
| `status.outputs.hostname` | `string` | The fully-qualified hostname of the record (the provider's computed fqdn), e.g. "www.example.com", or "example.com" for apex records. |
| `status.outputs.record_type` | `string` | The DNS record type that was created. One of: A, AAAA, CNAME, MX, TXT, SRV, NS, CAA, SOA. |
| `status.outputs.domain` | `string` | The domain name (DNS zone) the record was created in. |
| `status.outputs.ttl_seconds` | `int32` | The TTL (time to live) in seconds applied to the record. When the spec left ttl_seconds unset, this carries the API's applied default. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.domain` | DigitalOceanDnsZone | `status.outputs.zone_name` |

## See Also

- [Overview](../README.md)
