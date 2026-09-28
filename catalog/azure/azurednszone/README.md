# AzureDnsZone

Create an Azure public DNS zone: an internet-facing, authoritative DNS zone hosted on Azure's global anycast name-server fleet.

## Overview

The zone is deliberately just the zone -- an empty record container plus its Start of Authority settings. Records are declared through standalone `AzureDnsRecord` resources referencing the zone's `zone_name` output, one resource per record set, added and removed without touching the zone.

Creating a zone does not make it authoritative on the internet: Azure assigns four name servers (the `nameservers` output), and the domain only resolves through the zone once those name servers are configured at the domain's registrar -- or as NS records in the parent zone, for subdomain delegation. For name resolution inside virtual networks, use `AzurePrivateDnsZone` instead.

## Key Features

- **Global anycast hosting**: Azure serves the zone from its worldwide name-server fleet -- no DNS servers to run
- **SOA customization**: contact email, refresh/retry/expire timers, and the negative-caching TTL, with Azure's defaults when omitted
- **Delegation handoff**: the `nameservers` output is exactly what the registrar (or the parent zone's NS records) needs
- **Governance tags**: user tags merged over the platform's metadata-derived tags

## When to Use

- Hosting an internet-facing domain (example.com) or a delegated subdomain zone (team.example.com)
- The DNS backbone for TLS/domain flows -- Front Door custom domains and Container Apps custom domains publish their validation records into the zone
- Automated record management: AKS web-app routing manages records in zones referenced by `zone_id`

## Spec Highlights

| Field | Notes |
| --- | --- |
| `zone_name` | The domain, no trailing dot. ForceNew -- renaming replaces the zone, its records, AND the name-server set |
| `resource_group` | ARM lifecycle scope only; no effect on resolution. Defaults to an `AzureResourceGroup` reference |
| `soa_record` | Optional SOA customization; the SOA host name is always Azure's own |
| `tags` | User tags, merged over metadata-derived tags. Updatable in place |

## Outputs

| Output | Purpose |
| --- | --- |
| `zone_id` | The ARM ID -- consumed by kinds that watch or manage the zone (Front Door custom domains, AKS web-app routing) |
| `zone_name` | The join key `AzureDnsRecord` resources address record sets through |
| `resource_group_name` | Pairs with `zone_name` for management-plane addressing |
| `nameservers` | The four assigned name servers -- configure these at the registrar |
| `max_number_of_record_sets` | The zone's record-set capacity limit |

## Quick Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDnsZone
metadata:
  name: example-com
spec:
  zoneName: example.com
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      fieldPath: status.outputs.resource_group_name
  tags:
    team: platform
```

After deployment, configure the four `nameservers` outputs at the domain's registrar to make the zone authoritative.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
