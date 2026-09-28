# AzurePrivateDnsRecord

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzurePrivateDnsRecordSpec** defines one record set in an Azure
PRIVATE DNS zone -- every value the zone answers for one (name, type)
pair, visible only inside the virtual networks linked to the zone.

The record type is declared by which typed payload is present: set
exactly one of `a`, `aaaa`, `cname`, `mx`, `ptr`, `srv`, or `txt`.
Each payload carries the value shape DNS actually defines for that
type (MX entries are preference+exchange pairs, SRV entries are
priority/weight/port/target), so a record can never be declared with
a shape its type cannot hold. Private DNS supports exactly these
seven types -- there is no CAA (private names never face public
certificate authorities) and no NS (private zones cannot delegate
subdomains; create a separate zone per suffix instead). Unlike the
public AzureDnsRecord, there is no alias arm either -- the private
DNS service has no alias record concept; auto-registered VM records
are the service's own lifecycle mechanism, not a record type.

**One record set per (name, type)**: Azure stores all values for a
(name, type) pair as one record set, so declare all of them in one
resource. A second AzurePrivateDnsRecord with the same name and type
in the same zone conflicts with this one rather than merging into it.

## Example

```yaml
# Offline-plan test manifest. Exercises a structured payload at depth:
# an apex MX record set (name "@") with two entries at different
# preferences, an explicit non-default TTL, and user tags merged over
# the derived ones. The six other payload types are pure ARM property
# writes on the same create path -- the offline plans and the live lane
# cover them per the profile's record.
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePrivateDnsRecord
metadata:
  name: test-private-dns-record
  org: test-org
  env: dev
spec:
  privateDnsZoneId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/privateDnsZones/internal.contoso.com
  name: "@"
  ttlSeconds: 3600
  mx:
    - preference: 10
      exchange:
        value: mail1.internal.contoso.com
    - preference: 20
      exchange:
        value: mail2.internal.contoso.com
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.privateDnsZoneId` | `string \| valueFrom` | yes |  | AzurePrivateDnsZone (`status.outputs.zone_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.ttlSeconds` | `int32` |  | `300` |  |
| `spec.tags` | `map<string, string>` |  |  |  |
| `spec.a` | `[]string \| valueFrom` |  |  |  |
| `spec.aaaa` | `[]string \| valueFrom` |  |  |  |
| `spec.cname` | `string \| valueFrom` |  |  |  |
| `spec.mx` | `[]AzurePrivateDnsMxEntry` |  |  |  |
| `spec.mx[].preference` | `int32` | yes |  |  |
| `spec.mx[].exchange` | `string \| valueFrom` | yes |  |  |
| `spec.ptr` | `[]string \| valueFrom` |  |  |  |
| `spec.srv` | `[]AzurePrivateDnsSrvEntry` |  |  |  |
| `spec.srv[].priority` | `int32` | yes |  |  |
| `spec.srv[].weight` | `int32` | yes |  |  |
| `spec.srv[].port` | `int32` | yes |  |  |
| `spec.srv[].target` | `string \| valueFrom` | yes |  |  |
| `spec.txt` | `[]string \| valueFrom` |  |  |  |

## Field Details

### spec.privateDnsZoneId

`string | valueFrom` · required

The private DNS zone this record is created in, by ARM resource ID
(.../providers/Microsoft.Network/privateDnsZones/{zone-name}) --
defaults to referencing an AzurePrivateDnsZone's zone_id output;
pass the full ID as a literal for a zone managed outside Planton.
Changing it replaces the record.

- references: AzurePrivateDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzurePrivateDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.name

`string` · required

The record name, relative to the zone. Changing it replaces the
record.
  - "@" for the zone apex (the zone name itself -- the required
    form for apex MX records)
  - "db" for db.<zone-name>
  - "api.v1" for api.v1.<zone-name>
  - "*" or "*.app" for wildcards
  - underscore-led service names: "_sip._tcp" (SRV service names)

- rule: Record name must be '@' for the zone apex, '*' (optionally '*.<name>') for wildcards, or dot-separated labels of lowercase letters, digits, hyphens, and underscores -- e.g. db, api.v1, _sip._tcp
- rule: {"required":true}

### spec.ttlSeconds

`int32` · optional (explicit presence)

Time to live in seconds: how long resolvers may cache this record.
Low TTLs (60-300) make changes visible quickly at the cost of more
queries; long TTLs (3600+) suit stable records. Unspecified applies
300 (5 minutes) -- the modules always send the effective value
explicitly.

- default: `300`
- rule: {"int32":{"lte":2147483647,"gte":0}}

### spec.tags

`map<string, string>`

Free-form tags applied to the record set (stored as ARM record-set
metadata -- record sets carry no ARM tags proper), merged over the
Planton-derived resource tags; a user tag with the same key wins.
Updatable in place.

### spec.a

`[]string | valueFrom`

IPv4 addresses this name answers with; multiple addresses
round-robin. Azure caps an A record set at 20 addresses. Set
exactly one payload field on this spec. Each entry can be a literal or
a reference to another resource's address output (a private endpoint,
an internal load balancer's frontend).

- rule: each literal address is an IPv4 address
- rule: {"repeated":{"maxItems":"20"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.aaaa

`[]string | valueFrom`

IPv6 addresses this name answers with (e.g. "2001:db8::1" -- Azure
normalizes the compressed form); multiple addresses round-robin.
Set exactly one payload field on this spec. Each entry can be a literal
or a reference to another resource's address output.

- rule: each literal address is an IPv6 address
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.cname

`string | valueFrom`

Canonical-name record: the hostname this name is an alias for (e.g.
"db.internal.contoso.com"). A reference or a literal: reference
another resource's hostname output when the target is minted at
deploy time, pass a literal for known names. No kind dominates
CNAME targets, so references declare their kind explicitly. DNS
forbids CNAME at the zone apex. Set exactly one payload field on
this spec.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.mx

`[]AzurePrivateDnsMxEntry`

Mail-exchange entries: one per mail server, each with its own
preference. Private-zone MX records live at the zone apex in
virtually every deployment -- set name to "@". Set exactly one
payload field on this spec.

### spec.mx[].preference

`int32` · required · optional (explicit presence)

Delivery preference: mail servers are tried lowest-preference
first, equal preferences load-balance. Common convention: 10
primary, 20 secondary. 0 is legal (and used by the "null MX"
no-mail convention).

- rule: {"required":true,"int32":{"lte":65535,"gte":0}}

### spec.mx[].exchange

`string | valueFrom` · required

The mail server hostname (e.g. "mail.internal.contoso.com"), a literal
or a reference to another resource's hostname output.

- rule: a literal hostname is at most 253 characters
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.ptr

`[]string | valueFrom`

Pointer hostnames for reverse DNS (IP-to-name, in private
in-addr.arpa / ip6.arpa zones). Set exactly one payload field on
this spec. Each entry can be a literal or a reference to another
resource's hostname output.

- rule: each literal hostname is at most 253 characters
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.srv

`[]AzurePrivateDnsSrvEntry`

Service-locator entries (SIP, LDAP, Kerberos, ...). The record NAME
carries the service and protocol ("_sip._tcp"). Set exactly one
payload field on this spec.

### spec.srv[].priority

`int32` · required · optional (explicit presence)

Endpoint precedence: clients try lowest priority first. 0 is the
conventional primary.

- rule: {"required":true,"int32":{"lte":65535,"gte":0}}

### spec.srv[].weight

`int32` · required · optional (explicit presence)

Relative weight for load-balancing among endpoints of equal
priority. 0 means "no preference" and is legal when only one
endpoint exists at a priority.

- rule: {"required":true,"int32":{"lte":65535,"gte":0}}

### spec.srv[].port

`int32` · required · optional (explicit presence)

The TCP/UDP port the service listens on (e.g. 5060 for SIP).

- rule: {"required":true,"int32":{"lte":65535,"gte":1}}

### spec.srv[].target

`string | valueFrom` · required

The hostname providing the service (e.g. "sip.internal.contoso.com").
Must be a hostname with its own A/AAAA record, never an IP address. A
literal or a reference to another resource's hostname output.

- rule: a literal hostname is at most 253 characters
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.txt

`[]string | valueFrom`

Text values (service discovery metadata, verification strings).
Azure caps each value at 1,024 characters. Values are references or
literals and may mix freely in one record set: reference another
resource's output when the value is minted at deploy time, pass
literals for everything hand-authored. No kind dominates TXT
values, so references declare their kind explicitly. Set exactly
one payload field on this spec.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

## Validation Rules

- `azure_private_dns_record_exactly_one_payload`: Set exactly one record payload -- a, aaaa, cname, mx, ptr, srv, or txt -- the payload determines the record type

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzurePrivateDnsRecord, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.record_id` | `string` | The Azure Resource Manager ID of the record set. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/privateDnsZones/{zone}/{TYPE}/{name} where {TYPE} is the record type (A, AAAA, CNAME, MX, PTR, SRV, TXT). |
| `status.outputs.fqdn` | `string` | The fully qualified domain name of the record set, with a trailing dot as DNS writes it (e.g. "db.internal.contoso.com." -- the zone apex is the zone name itself). Resolvable only from virtual networks linked to the zone. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.privateDnsZoneId` | AzurePrivateDnsZone | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
