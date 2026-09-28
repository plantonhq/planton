# AzureDnsZone

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureDnsZoneSpec** defines the configuration for creating an Azure
public DNS zone: an internet-facing, authoritative DNS zone hosted on
Azure's global anycast name-server fleet.

The zone is deliberately just the zone -- an empty record container plus
its Start of Authority settings. Records are declared through standalone
AzureDnsRecord resources referencing this zone's zone_name output, one
resource per record set, added and removed without touching the zone.

Creating a zone does NOT make it authoritative on the internet: Azure
assigns four name servers (the nameservers output), and the domain only
resolves through this zone once those name servers are configured at the
domain's registrar (or as NS records in the parent zone, for subdomain
delegation). The same zone name can exist in many subscriptions at once --
each copy gets a different name-server set, and only the delegated one
answers the internet.

**Note:** Public DNS zones are global Azure resources -- they have no
region. This is why this spec omits the `region` field that other Azure
resources include. For name resolution inside virtual networks, use
AzurePrivateDnsZone instead.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDnsZone
metadata:
  name: test-dns-zone
spec:
  zone_name: test-zone.example.com
  resource_group:
    value: test-rg
  soa_record:
    email: dnsadmin.example.com
    minimum_ttl: 60
    serial_number: 2
  tags:
    team: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.zoneName` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.soaRecord` | `AzureDnsZoneSoaRecord` |  |  |  |
| `spec.soaRecord.email` | `string` | yes |  |  |
| `spec.soaRecord.expireTime` | `int64` |  | `2419200` |  |
| `spec.soaRecord.minimumTtl` | `int64` |  | `300` |  |
| `spec.soaRecord.refreshTime` | `int64` |  | `3600` |  |
| `spec.soaRecord.retryTime` | `int64` |  | `300` |  |
| `spec.soaRecord.serialNumber` | `int64` |  | `1` |  |
| `spec.soaRecord.ttl` | `int64` |  | `3600` |  |
| `spec.soaRecord.tags` | `map<string, string>` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.zoneName

`string` · required

The DNS zone name, e.g. "example.com" or "team.example.com" for a
delegated subdomain zone. Do not include a trailing dot. Changing the
name replaces the zone -- and with it every record it contains and the
assigned name-server set, so a rename breaks the registrar delegation
until it is updated.

- rule: Zone name must be a DNS domain name of at least two dot-separated labels, e.g. example.com -- lowercase letters, digits, and hyphens, with no trailing dot
- rule: {"required":true}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the DNS zone will be created in. Can be a
literal resource-group name or a reference to an AzureResourceGroup's
name output. The resource group only governs the zone's ARM lifecycle
(who can manage it); it has no effect on DNS resolution.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.soaRecord

`AzureDnsZoneSoaRecord`

Customize the zone's Start of Authority record. Nearly every
deployment leaves this unset and takes Azure's defaults; set it only
when operational tooling requires a specific contact email or
negative-caching behavior. Azure creates the SOA record with the zone
either way; this block edits the fields Azure allows to change (the
SOA host name is always Azure's own and is not configurable).

### spec.soaRecord.email

`string` · required

The email contact for the zone, in SOA host format: dots instead of
"@" (e.g. "dnsadmin.example.com" for dnsadmin@example.com). Azure's
default is "azuredns-hostmaster.microsoft.com".

- rule: SOA email must use dots instead of @ (e.g. dnsadmin.example.com for dnsadmin@example.com), contain only letters, digits, dots, underscores, and hyphens, and have 2-34 dot-separated segments of at most 63 characters each
- rule: {"required":true}

### spec.soaRecord.expireTime

`int64` · optional (explicit presence)

Seconds a secondary name server considers the zone valid without a
successful refresh. Azure's default is 2419200 (28 days).

- default: `2419200`
- rule: {"int64":{"gte":"0"}}

### spec.soaRecord.minimumTtl

`int64` · optional (explicit presence)

Negative-caching TTL: how long resolvers cache a "name does not
exist" answer, in seconds. Azure's default is 300. The one field with
real operational impact -- raising it slows how quickly newly-created
records become visible to clients that asked before the record
existed.

- default: `300`
- rule: {"int64":{"gte":"0"}}

### spec.soaRecord.refreshTime

`int64` · optional (explicit presence)

Seconds between secondary refresh attempts. Azure's default is 3600.

- default: `3600`
- rule: {"int64":{"gte":"0"}}

### spec.soaRecord.retryTime

`int64` · optional (explicit presence)

Seconds a secondary waits before retrying a failed refresh. Azure's
default is 300.

- default: `300`
- rule: {"int64":{"gte":"0"}}

### spec.soaRecord.serialNumber

`int64` · optional (explicit presence)

The zone's serial number. Azure does not auto-increment serials on
record changes (its name servers replicate internally), so this is
cosmetic metadata for tooling that inspects the SOA. Azure's default
is 1.

- default: `1`
- rule: {"int64":{"gte":"0"}}

### spec.soaRecord.ttl

`int64` · optional (explicit presence)

TTL of the SOA record itself, in seconds. Azure's default is 3600.

- default: `3600`
- rule: {"int64":{"lte":"2147483647","gte":"0"}}

### spec.soaRecord.tags

`map<string, string>`

Tags applied to the SOA record set (distinct from the zone's own
tags). Rarely used; exists because ARM models record-set metadata.

### spec.tags

`map<string, string>`

Free-form tags applied to the zone, merged over the Planton-derived
resource tags (organization, environment, resource id); a user tag
with the same key wins. Tags are Azure's governance surface -- Azure
Policy enforces them and Microsoft Cost Management groups by them --
so carry your org's ownership/cost-center conventions here. Updatable
in place.

## Validation Rules

- `azure_dns_zone_name_plus_soa_email_length`: The zone name and the SOA email together cannot exceed 253 characters (Azure combines them into the SOA record) -- shorten the soa_record.email

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureDnsZone, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.zone_id` | `string` | The Azure Resource Manager ID of the DNS zone. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/dnsZones/{name} |
| `status.outputs.zone_name` | `string` | The DNS zone name (e.g. "example.com"). Echoed from the spec -- AzureDnsRecord resources reference it to address record sets. |
| `status.outputs.resource_group_name` | `string` | The resource group the zone lives in. Echoed for downstream tooling that addresses records by zone name + resource group rather than parsing the ARM ID. |
| `status.outputs.nameservers` | `[]string` | The four name servers Azure assigned to this zone (e.g. "ns1-05.azure-dns.com."). The zone only answers the internet once these are configured at the domain's registrar, or as NS records in the parent zone for subdomain delegation. |
| `status.outputs.max_number_of_record_sets` | `int64` | The maximum number of record sets this zone can hold -- Azure's per-zone capacity limit (10000 by default; higher by support request). A capacity fact for planning, not a live count. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureAksCluster | `spec.webAppRouting.dnsZoneIds` | `status.outputs.zone_id` |
| AzureDnsRecord | `spec.zoneName` | `status.outputs.zone_name` |
| AzureFrontDoorCustomDomain | `spec.dnsZoneId` | `status.outputs.zone_id` |
| KubernetesExternalDns | `spec.azureDns.zoneIdFilters` | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
