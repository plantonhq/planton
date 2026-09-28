# AzureDnsRecord

Declare one record set in an Azure public DNS zone -- every value the zone answers for one (name, type) pair.

## Overview

The record type is declared by which typed payload is present: set exactly one of `a`, `aaaa`, `cname`, `mx`, `srv`, `caa`, `txt`, `ns`, or `ptr`. Each payload carries the value shape DNS actually defines for that type -- MX entries are preference+exchange pairs, SRV entries are priority/weight/port/target, CAA entries are flags/tag/value -- so a record can never be declared with a shape its type cannot hold.

Azure stores all values for a (name, type) pair as one record set: declare all of them in one resource. A second AzureDnsRecord with the same name and type in the same zone conflicts rather than merging.

## Key Features

- **Typed payloads**: the real DNS value shapes per type, validated up front (addresses are real IPs, MX preferences are per-entry, CAA tags are the four Azure accepts)
- **Alias records**: A, AAAA, and CNAME can track an Azure resource (`target_resource_id`) instead of literal values -- Azure keeps the answer in sync when the resource's address changes, and aliases work at the zone apex where DNS forbids CNAME
- **Long TXT values**: up to 4096 characters (DKIM keys); Azure transparently splits them into the 254-character strings DNS requires
- **Composable**: the zone by `zone_name` reference; alias targets by any resource's ARM-id output

## When to Use

- Application records: A/AAAA/CNAME for web apps, APIs, and CDN endpoints
- Email infrastructure: MX, SPF/DKIM/DMARC TXT records
- Domain-verification records: the `asuid.` TXT records Container Apps custom domains require, Front Door's `_dnsauth.` tokens
- Certificate authority control (CAA), service discovery (SRV), subdomain delegation (NS)

## Spec Highlights

| Field | Notes |
| --- | --- |
| `zone_name` + `resource_group` | Address the zone (Azure's management plane has no ARM-id mode for record sets). Both ForceNew |
| `name` | `@` apex, `www`, `*.app` wildcards, `_dmarc`/`_sip._tcp` underscore services. ForceNew |
| `ttl_seconds` | Defaults to 300 |
| one typed payload | `a`/`aaaa`/`cname`/`mx`/`srv`/`caa`/`txt`/`ns`/`ptr` -- exactly one |
| `tags` | Record-set metadata tags. Updatable in place |

## Outputs

| Output | Purpose |
| --- | --- |
| `record_id` | The record set's ARM ID (embeds the type as a path segment) |
| `fqdn` | The record's fully qualified name, with DNS's trailing dot |

## Quick Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDnsRecord
metadata:
  name: www-a-record
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      fieldPath: status.outputs.resource_group_name
  zoneName:
    valueFrom:
      kind: AzureDnsZone
      name: example-com
      fieldPath: status.outputs.zone_name
  name: www
  a:
    addresses:
      - value: 203.0.113.10
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
