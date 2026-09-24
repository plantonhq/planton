# DigitalOceanDnsZone

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanDnsZoneSpec models the full `digitalocean_domain` surface plus
an inline record list (each record is a `digitalocean_record` the module
manages inside the zone). Adding a domain to DigitalOcean DNS does not
require owning it, but the domain only resolves once its registrar
delegates to DigitalOcean's name servers (ns1/ns2/ns3.digitalocean.com).

## Example

```yaml
# Example DigitalOceanDnsZone manifests. Deploy with:
#   planton apply -f manifest.yaml
#
# Document 1 -- the smallest real zone: just the domain name. DigitalOcean
# hosts the zone immediately; it resolves publicly once the registrar
# delegates to ns1/ns2/ns3.digitalocean.com.
#
# Document 2 -- a production-shaped zone: apex A records (two values fan out
# to two provider records), a www alias, mail routing (MX with priority),
# an SPF policy, a service locator (SRV), certificate authority pinning
# (CAA), and the create-only ip_address convenience that seeds an untracked
# apex A record.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDnsZone
metadata:
  name: example-dodns-minimal
spec:
  domainName: example.com
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDnsZone
metadata:
  name: example-dodns-full
spec:
  domainName: example.org
  ipAddress: 203.0.113.10
  records:
    - name: "@"
      type: A
      values:
        - value: 203.0.113.1
        - value: 203.0.113.2
      ttlSeconds: 300
    - name: www
      type: CNAME
      values:
        - value: example.org.
    - name: "@"
      type: MX
      values:
        - value: mail.example.org.
      priority: 10
    - name: "@"
      type: TXT
      values:
        - value: v=spf1 include:_spf.google.com ~all
    - name: _sip._tcp
      type: SRV
      values:
        - value: sip.example.org.
      priority: 10
      weight: 60
      port: 5060
    - name: "@"
      type: CAA
      values:
        - value: letsencrypt.org
      flags: 0
      tag: issue
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.domainName` | `string` | yes |  |  |
| `spec.records` | `[]DigitalOceanDnsZoneRecord` |  |  |  |
| `spec.records[].name` | `string` | yes |  |  |
| `spec.records[].values` | `[]string \| valueFrom` | yes |  |  |
| `spec.records[].ttlSeconds` | `uint32` |  |  |  |
| `spec.records[].type` | `enum` | yes |  |  |
| `spec.records[].priority` | `uint32` |  |  |  |
| `spec.records[].weight` | `uint32` |  |  |  |
| `spec.records[].port` | `uint32` |  |  |  |
| `spec.records[].flags` | `uint32` |  |  |  |
| `spec.records[].tag` | `string` |  |  |  |
| `spec.ipAddress` | `string` |  |  |  |

## Field Details

### spec.domainName

`string` · required

The domain name for the DNS zone. Must be a fully-qualified domain name
(e.g. "example.com"). Domain names are unique across ALL DigitalOcean
accounts — adding a domain another account already holds fails. Changing
it recreates the zone.

- rule: {"required":true,"string":{"pattern":"^(?:[A-Za-z0-9-]+\\.)+[A-Za-z]{2,}$"}}

### spec.records

`[]DigitalOceanDnsZoneRecord`

DNS records to create within the zone (optional).

- rule: type must be one of A, AAAA, CNAME, MX, TXT, SRV, NS, CAA, SOA (DigitalOcean does not support ALIAS or PTR records)
- rule: priority must be specified for MX records
- rule: priority, weight, and port must be specified for SRV records
- rule: flags and tag must be specified for CAA records

### spec.records[].name

`string` · required

The record's host name, relative to the zone. Use "@" for the zone apex.

- rule: {"required":true}

### spec.records[].values

`[]string | valueFrom` · required

The value or values for the record; each value creates its own record
row of the same name and type (e.g. two A values make two A records).
- A/AAAA: one or more IP addresses
- CNAME/MX/NS/SRV: the target hostname
- TXT: the text data
- CAA: the certificate authority domain
Each value can be a literal or a reference to another resource's output.
Hostname values (CNAME, MX, NS, SRV, and CAA except tag=iodef) must be
written either fully qualified WITH a trailing dot
("mail.example.com.", "letsencrypt.org.") or relative to this zone
("mail"). DigitalOcean reports every such value back fully qualified
with a trailing dot, and the provider forgives only those two spellings
— a bare fully-qualified name without the dot ("letsencrypt.org") is
re-applied on every run, forever.

- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.records[].ttlSeconds

`uint32`

Time to live for the record, in seconds. When unset (0), the DigitalOcean
API applies its default (1800 seconds). DigitalOcean harmonizes TTLs
across records sharing a fully-qualified name (RFC 2181 §5.2) and
rewrites the stragglers server-side, so give every record on one name
the same ttl_seconds (or leave them all unset) — a lone custom TTL on a
shared name is overwritten and shows up as a change on every run. The
apex A record seeded by `ip_address` counts as one of those records.

### spec.records[].type

`enum` · required

The type of DNS record. Required; see the message rules for the accepted
subset of the shared enum.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `unspecified`
- `A` -- host address
- `AAAA` -- ipv6 host address
- `ALIAS` -- auto resolved alias
- `CNAME` -- canonical name for an alias
- `MX` -- mail exchange
- `NS` -- name server
- `PTR` -- pointer
- `SOA` -- start of authority
- `SRV` -- location of service
- `TXT` -- descriptive text
- `CAA` -- certificate authority authorization

### spec.records[].priority

`uint32` · optional (explicit presence)

Priority for MX and SRV records; lower values are preferred. Required
for MX and SRV, ignored for other types. Range: 0-65535. Provider
quirk: an explicit 0 passes validation but is dropped from the create
request, so the API's default applies — use a positive value when the
priority must be exact.

- rule: {"uint32":{"lte":65535}}

### spec.records[].weight

`uint32` · optional (explicit presence)

Weight for SRV records: the relative share of traffic among records
with equal priority. Required for SRV, ignored for other types.
Range: 0-65535. The same explicit-zero provider quirk as priority
applies.

- rule: {"uint32":{"lte":65535}}

### spec.records[].port

`uint32` · optional (explicit presence)

Port for SRV records: where the service is available. Required for SRV,
ignored for other types. Range: 0-65535. The same explicit-zero
provider quirk as priority applies.

- rule: {"uint32":{"lte":65535}}

### spec.records[].flags

`uint32` · optional (explicit presence)

Flags for CAA records: 0 (non-critical, the common value) or 128
(critical). Required for CAA, ignored for other types. Range: 0-255.
An explicit 0 is dropped from the create request (provider quirk),
which is harmless: the API defaults flags to 0.

- rule: {"uint32":{"lte":255}}

### spec.records[].tag

`string`

Tag for CAA records — the property being authorized: "issue",
"issuewild", or "iodef". Required for CAA, ignored for other types.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["issue","issuewild","iodef"]}}

### spec.ipAddress

`string`

(Optional) An IPv4 address that seeds an initial A record at the zone
apex when the zone is created. Applied at creation ONLY; later edits are
ignored (both provisioners skip changes to it, because the only
alternative the provider offers is recreating the whole zone). The
DigitalOcean API never returns it, and the A record it creates is NOT
tracked — later edits to `records` will not see or manage it, yet it
shares the apex A record set (and therefore the TTL) with any apex A
records you declare. Prefer declaring an apex A record in `records`,
which is tracked and updatable; use this only when migrating a
configuration that already relies on it.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanDnsZone, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.zone_name` | `string` | The domain name of the DNS zone (e.g. "example.com"). |
| `status.outputs.zone_id` | `string` | The zone's resource identifier. DigitalOcean addresses domains by NAME — this is the domain name itself, not a UUID. |
| `status.outputs.name_servers` | `[]string` | DigitalOcean's authoritative name servers for every hosted zone (ns1/ns2/ns3.digitalocean.com — a fixed platform-wide set the API does not return per zone). Set these at the domain's registrar to delegate. |
| `status.outputs.urn` | `string` | The uniform resource name of the domain (e.g. "do:domain:example.com"). |
| `status.outputs.record_ids` | `map<string, string>` | Numeric ids of the inline `records`, one entry per record value, keyed by "<record name>-<record index>-<value index>" (the index positions in the manifest's `records` list and that entry's `values` list, both from 0 — e.g. "@-0-0", "www-1-0"). DigitalOcean addresses a record as /v2/domains/{domain}/records/{id}, and state import takes "{domain},{record_id}"; this map is where the second half comes from. Empty for a zone with no inline records. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| DigitalOceanApp | `spec.domains[].zone` | `status.outputs.zone_name` |
| DigitalOceanDnsRecord | `spec.domain` | `status.outputs.zone_name` |

## See Also

- [Overview](../README.md)
