# AzureDnsRecord

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureDnsRecordSpec** defines one record set in an Azure public DNS
zone -- every value the zone answers for one (name, type) pair.

The record type is declared by which typed payload is present: set
exactly one of `a`, `aaaa`, `cname`, `mx`, `srv`, `caa`, `txt`, `ns`,
or `ptr`. Each payload carries the value shape DNS actually defines for
that type (MX entries are preference+exchange pairs, SRV entries are
priority/weight/port/target, CAA entries are flags/tag/value), so a
record can never be declared with a shape its type cannot hold.

**Alias records**: A, AAAA, and CNAME payloads can point at an Azure
resource instead of carrying literal values (`target_resource_id`).
Azure then keeps the answer in sync with the resource -- when a Public
IP's address changes, the alias A record follows it automatically, with
no drift window and no stale-IP outage. Alias records also work at the
zone apex where CNAME is forbidden by DNS itself.

**One record set per (name, type)**: Azure stores all values for a
(name, type) pair as one record set, so declare all of them in one
resource. A second AzureDnsRecord with the same name and type in the
same zone conflicts with this one rather than merging into it.

## Example

```yaml
# The CAA payload deliberately exercises the record kind's riskiest
# translation seam: the tag enum arrives as the proto value name (ISSUE)
# and must map to Azure's lowercase wire vocabulary in both modules.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDnsRecord
metadata:
  name: test-caa-record
spec:
  resource_group:
    value: test-rg
  zone_name:
    value: test-zone.example.com
  name: "@"
  ttl_seconds: 300
  caa:
    - flags: 0
      tag: ISSUE
      value: letsencrypt.org
    - flags: 0
      tag: IODEF
      value: mailto:security@example.com
  tags:
    team: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.zoneName` | `string \| valueFrom` | yes |  | AzureDnsZone (`status.outputs.zone_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.ttlSeconds` | `int32` |  | `300` |  |
| `spec.tags` | `map<string, string>` |  |  |  |
| `spec.a` | `AzureDnsARecord` |  |  |  |
| `spec.a.addresses` | `[]string \| valueFrom` |  |  |  |
| `spec.a.targetResourceId` | `string \| valueFrom` |  |  |  |
| `spec.aaaa` | `AzureDnsAaaaRecord` |  |  |  |
| `spec.aaaa.addresses` | `[]string \| valueFrom` |  |  |  |
| `spec.aaaa.targetResourceId` | `string \| valueFrom` |  |  |  |
| `spec.cname` | `AzureDnsCnameRecord` |  |  |  |
| `spec.cname.value` | `string \| valueFrom` |  |  |  |
| `spec.cname.targetResourceId` | `string \| valueFrom` |  |  |  |
| `spec.mx` | `[]AzureDnsMxEntry` |  |  |  |
| `spec.mx[].preference` | `int32` | yes |  |  |
| `spec.mx[].exchange` | `string \| valueFrom` | yes |  |  |
| `spec.srv` | `[]AzureDnsSrvEntry` |  |  |  |
| `spec.srv[].priority` | `int32` | yes |  |  |
| `spec.srv[].weight` | `int32` | yes |  |  |
| `spec.srv[].port` | `int32` | yes |  |  |
| `spec.srv[].target` | `string \| valueFrom` | yes |  |  |
| `spec.caa` | `[]AzureDnsCaaEntry` |  |  |  |
| `spec.caa[].flags` | `int32` | yes |  |  |
| `spec.caa[].tag` | `enum` | yes |  |  |
| `spec.caa[].value` | `string` | yes |  |  |
| `spec.txt` | `[]string \| valueFrom` |  |  |  |
| `spec.ns` | `[]string \| valueFrom` |  |  |  |
| `spec.ptr` | `[]string \| valueFrom` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the record's DNS zone lives in. Must address
the SAME zone as zone_name -- Azure's management plane addresses
record sets by (resource group, zone name, type, record name).

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.zoneName

`string | valueFrom` · required

The name of the DNS zone this record is created in (e.g.
"example.com"). Reference an AzureDnsZone's zone_name output, or pass
the zone name of a zone managed outside Planton as a literal. Changing
it replaces the record.

- references: AzureDnsZone (`status.outputs.zone_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_name}} -- a bare string does not parse

### spec.name

`string` · required

The record name, relative to the zone. Changing it replaces the
record.
  - "@" for the zone apex (example.com itself)
  - "www" for www.example.com
  - "api.v1" for api.v1.example.com
  - "*" or "*.app" for wildcards
  - underscore-led service names: "_dmarc", "_sip._tcp",
    "asuid.myapp" (domain-verification records)

- rule: Record name must be '@' for the zone apex, '*' (optionally '*.<name>') for wildcards, or dot-separated labels of lowercase letters, digits, hyphens, and underscores -- e.g. www, api.v1, _dmarc, _sip._tcp
- rule: {"required":true}

### spec.ttlSeconds

`int32` · optional (explicit presence)

Time to live in seconds: how long resolvers may cache this record.
Low TTLs (60-300) make changes visible quickly at the cost of more
queries; long TTLs (3600+) suit stable records like MX. Azure has no
server-side default -- Planton applies 300 (5 minutes) when unset.

- default: `300`
- rule: {"int32":{"lte":2147483647,"gte":0}}

### spec.tags

`map<string, string>`

Free-form tags applied to the record set (stored as ARM record-set
metadata), merged over the Planton-derived resource tags; a user tag
with the same key wins. Updatable in place.

### spec.a

`AzureDnsARecord`

IPv4 address record. Set exactly one payload field on this spec.

- rule: Provide either literal IPv4 addresses or an alias target_resource_id, not both and not neither -- an alias record delegates its answer to the referenced Azure resource

### spec.a.addresses

`[]string | valueFrom`

The IPv4 addresses this name answers with. Multiple addresses
round-robin. Mutually exclusive with target_resource_id. Each entry can
be a literal or a reference to another resource's address output (a
public IP, a load balancer's frontend).

- rule: each literal address is an IPv4 address
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.a.targetResourceId

`string | valueFrom`

Alias target: the ARM ID of an Azure resource whose IPv4 address
this record should track automatically (a Public IP, a Traffic
Manager profile, another record set). Reference the resource's ARM-id
output with an explicit valueFrom (e.g. an AzurePublicIp's
public_ip_id) -- no kind dominates alias targets, so there is no
default. Mutually exclusive with addresses.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.aaaa

`AzureDnsAaaaRecord`

IPv6 address record. Set exactly one payload field on this spec.

- rule: Provide either literal IPv6 addresses or an alias target_resource_id, not both and not neither -- an alias record delegates its answer to the referenced Azure resource

### spec.aaaa.addresses

`[]string | valueFrom`

The IPv6 addresses this name answers with (e.g. "2001:db8::1" --
Azure normalizes the compressed form). Multiple addresses
round-robin. Mutually exclusive with target_resource_id. Each entry can
be a literal or a reference to another resource's address output.

- rule: each literal address is an IPv6 address
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.aaaa.targetResourceId

`string | valueFrom`

Alias target: the ARM ID of an Azure resource whose IPv6 address this
record should track automatically. Reference the resource's ARM-id
output with an explicit valueFrom -- no kind dominates alias targets,
so there is no default. Mutually exclusive with addresses.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.cname

`AzureDnsCnameRecord`

Canonical-name (alias) record. Set exactly one payload field on this
spec. DNS forbids CNAME at the zone apex -- use an alias A/AAAA
record (target_resource_id) to point the apex at an Azure resource.

- rule: Provide either the target hostname in value or an alias target_resource_id, not both and not neither -- a CNAME answers with exactly one canonical name

### spec.cname.value

`string | valueFrom`

The hostname this name is an alias for, at most 253 characters (e.g.
"myapp.azurefd.net"). A trailing dot is optional -- Azure treats the
value as fully qualified either way. A reference or a literal:
reference another resource's hostname output when the target is
minted at deploy time (an AzureFrontDoorEndpoint's hash-suffixed
host_name), pass a literal for externally-known hostnames. No kind
dominates CNAME targets, so references declare their kind explicitly.
Mutually exclusive with target_resource_id.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.cname.targetResourceId

`string | valueFrom`

Alias target: the ARM ID of an Azure resource this CNAME should
track (a Traffic Manager profile, CDN endpoint, or Front Door
endpoint). Reference the resource's ARM-id output with an explicit
valueFrom -- no kind dominates alias targets, so there is no default.
Mutually exclusive with value.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.mx

`[]AzureDnsMxEntry`

Mail-exchange entries: one per mail server, each with its own
preference. Set exactly one payload field on this spec.

### spec.mx[].preference

`int32` · required · optional (explicit presence)

Delivery preference: mail servers are tried lowest-preference first,
equal preferences load-balance. Common convention: 10 primary, 20
secondary. 0 is legal (and used by the "null MX" no-mail convention).

- rule: {"required":true,"int32":{"lte":65535,"gte":0}}

### spec.mx[].exchange

`string | valueFrom` · required

The mail server hostname (e.g. "mail.example.com" or
"aspmx.l.google.com"), a literal or a reference to another resource's
hostname output.

- rule: a literal hostname is at most 253 characters
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.srv

`[]AzureDnsSrvEntry`

Service-locator entries (SIP, XMPP, LDAP, ...). The record NAME
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

- rule: {"required":true,"int32":{"lte":65535,"gte":0}}

### spec.srv[].target

`string | valueFrom` · required

The hostname providing the service (e.g. "sip.example.com"). Must be
a hostname with its own A/AAAA record, never an IP address. A literal or
a reference to another resource's hostname output.

- rule: a literal hostname is at most 253 characters
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.caa

`[]AzureDnsCaaEntry`

Certificate-authority-authorization entries: which CAs may issue
certificates for this name. Set exactly one payload field on this
spec.

### spec.caa[].flags

`int32` · required · optional (explicit presence)

The CAA critical flag: 0 (the near-universal value) lets CAs ignore
unrecognized tags; 128 tells CAs to refuse issuance if they do not
understand the tag.

- rule: {"required":true,"int32":{"lte":255,"gte":0}}

### spec.caa[].tag

`enum` · required

What this entry authorizes or configures.

- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_dns_caa_tag_unspecified`
- `ISSUE` -- Authorize a CA to issue single-name certificates for this name.
- `ISSUEWILD` -- Authorize a CA to issue wildcard certificates for this name.
- `IODEF` -- Where CAs report policy violations (a "mailto:" or https URL).
- `CONTACTEMAIL` -- A contact address CAs may use to reach the domain holder.

### spec.caa[].value

`string` · required

The tag's value: the CA domain for ISSUE/ISSUEWILD (e.g.
"letsencrypt.org", or ";" to forbid issuance), a "mailto:" or https
URL for IODEF, an email address for CONTACTEMAIL.

- rule: {"required":true}

### spec.txt

`[]string | valueFrom`

Text values (SPF, DKIM, DMARC, domain verification). Each value may
be up to 4096 characters -- Azure transparently splits long values
into the 254-character strings DNS requires and reassembles them on
read. Values are references or literals and may mix freely in one
record set: reference another resource's output when the value is
minted at deploy time (an AzureFrontDoorCustomDomain's
validation_token published at `_dnsauth.<host>`), pass literals for
everything hand-authored (SPF policies, DKIM keys). No kind dominates
TXT values, so references declare their kind explicitly. Set exactly
one payload field on this spec.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.ns

`[]string | valueFrom`

Name-server hostnames, for delegating a CHILD subdomain to another
zone's name servers (e.g. "team" NS records pointing at the
team.example.com zone's assigned servers). The zone's own apex NS
records are Azure-managed -- do not declare them. Set exactly one
payload field on this spec. Each entry can be a literal or a reference
to the child zone's name-server output, so the delegation follows the
zone through a recreate.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.ptr

`[]string | valueFrom`

Pointer hostnames for reverse DNS (IP-to-name, in in-addr.arpa /
ip6.arpa zones). Set exactly one payload field on this spec. Each entry
can be a literal or a reference to another resource's hostname output.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

## Validation Rules

- `azure_dns_record_exactly_one_payload`: Set exactly one record payload -- a, aaaa, cname, mx, srv, caa, txt, ns, or ptr -- the payload determines the record type

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureDnsRecord, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.record_id` | `string` | The Azure Resource Manager ID of the record set. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/dnsZones/{zone}/{TYPE}/{name} where {TYPE} is the record type (A, AAAA, CNAME, MX, SRV, CAA, TXT, NS, PTR). |
| `status.outputs.fqdn` | `string` | The fully qualified domain name of the record set, with a trailing dot as DNS writes it (e.g. "www.example.com." -- the zone apex is the zone name itself). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.zoneName` | AzureDnsZone | `status.outputs.zone_name` |

## See Also

- [Overview](../README.md)
