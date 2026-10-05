# Azure Private DNS Record

Deploys one DNS record set (A, AAAA, CNAME, MX, PTR, SRV, or TXT) in an Azure Private DNS zone -- name resolution visible only inside the virtual networks linked to the zone. The record type is declared by which typed payload the spec carries, so a record can never be declared with a shape its type cannot hold, and Azure stores all values for a (name, type) pair as one record set.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **One record set** of the type your spec's payload declares -- all values for the (name, type) pair in one resource. Exactly one of the seven typed record resources materializes, addressed by the zone's ARM ID plus the record name.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An Azure Private DNS Zone** -- the record is created inside a referenced zone; `privateDnsZoneId` defaults to referencing an AzurePrivateDnsZone's `zone_id` output, or takes a literal ARM ID for a zone managed outside Planton.

### Azure Subscription

- **One record set per (name, type)** -- a second record with the same name and type in the same zone conflicts rather than merges; declare all of a name's values in one resource.
- **Resolution requires zone links** -- records answer only inside virtual networks linked to the zone; an unlinked zone's records are inert.
- **Free at rest** -- private DNS bills per zone and per million queries, never per record.

## Deploy

### Console

Open the deployment store, find **Azure Private DNS Record**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the zone reference, the record name and TTL, and the typed payload. Start from the **A Record** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePrivateDnsRecord
metadata:
  name: db-record
  org: acme-corp
  env: prod
spec:
  privateDnsZoneId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network/providers/Microsoft.Network/privateDnsZones/internal.acme.com
  name: db
  ttlSeconds: 300
  a:
    - value: 10.0.4.10
```

```shell
planton apply -f record.yaml
```

This creates an A record answering `db.internal.acme.com` with one private IPv4 address, resolvable from every virtual network linked to the zone. An Infra Job tracks the provisioning in real time.

### InfraChart

When the zone (and the record's target) are Infra Components in the same chart, wire them by reference:

```yaml
spec:
  privateDnsZoneId:
    valueFrom:
      name: internal-zone
  name: db
  a:
    - value: 10.0.4.10
```

The InfraPipeline resolves the dependency graph, provisioning the zone before the records inside it.

## Key Configuration

These are the most important decisions when configuring an Azure Private DNS Record. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The payload IS the type -- set exactly one.** Private DNS supports exactly seven types: `a`, `aaaa`, `cname`, `mx`, `ptr`, `srv`, `txt`. There is no CAA (private names never face certificate authorities), no NS (private zones cannot delegate subdomains -- create a separate zone per suffix), and no alias arm (unlike the public Azure DNS Record, the private service has no alias concept). Validation rejects a spec with zero or two payloads.

**Declare all of a name's values in one record, always.** Azure stores every value for a (name, type) pair as one record set. Two resources declaring the same name and type do not merge -- the second conflicts with the first. When a service gains a second address, edit the existing record's list; multiple A values round-robin, capped at 20.

**Name, zone, and type are fixed -- plan renames as add-then-remove.** Changing a record's name, zone, or payload type replaces the record set, a resolution gap for that name while the apply runs. For a rename, add the new record first, remove the old one in a second apply; both names answer during the transition and caches drain on their own schedule.

**Size the TTL to the record's change cadence.** TTL is how long resolvers cache an answer -- and therefore how long clients keep using an address after you change it. Failover-sensitive names want 60 or lower; stable infrastructure names are fine at 3600+; the platform default is 300. The cache drains from the moment of the change: a TTL lowered in the same apply as an address change does not speed that change up.

**A record that "does not resolve" is usually a missing link.** Records live in the zone; resolution lives in the zone's virtual-network links. Check the Azure Private DNS Zone Virtual Network Link for the querying network first, the record second.

**Stay out of the auto-registration namespace.** Zone links with auto-registration enabled write A records for every VM in the linked network, managed by the service. A declared record with the same name as a registered VM hostname fights that lifecycle.

**Reference outputs where public DNS would use an alias.** The equivalent of alias records here is valueFrom: point `cname` at another component's hostname output, or `a` values at address outputs, so redeployments flow the new value through the reference at the next apply. DNS forbids CNAME at the zone apex -- apex names use A records with referenced addresses.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|---|---|---|
| Azure Private DNS Zone | `privateDnsZoneId` | `status.outputs.zone_id` |
| Any component with a hostname output (CNAME target) | `cname` | declared explicitly per kind |
| Any component with a string output (TXT values) | `txt` | declared explicitly per kind |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `fqdn` | The record's fully qualified name, with DNS's trailing dot (e.g. `db.internal.acme.com.`) | CNAME targets in other records, application configuration |

`record_id` (the record set's ARM ID) is also exported; nothing downstream consumes a record set by reference, so it exists for identification and import.

## Common Patterns

**Everyday private name** -- An A record answering `db.<zone>` with a private IPv4 address: databases, internal APIs, anything on an address you manage. Start from the **A Record** preset.

**Stable name over a moving target** -- A CNAME so consumers keep one name while the infrastructure behind it changes: a failover pair's active member, a private endpoint's generated FQDN. Pair it with a low TTL -- the point of an alias is agility. Start from the **CNAME Alias** preset.

**Service discovery in the zone** -- SRV records with underscore-led names (`_sip._tcp`) locate services by priority, weight, port, and target; apex MX records (`name: "@"`) route internal mail. Both payloads carry the exact shape DNS defines, so multi-server setups express directly.

## Works With

- [**Azure Private DNS Zone**](/infra-catalog/azure-private-dns-zone) -- the zone the record lives in; reference its `zone_id` output.
- [**Azure Private DNS Zone Virtual Network Link**](/infra-catalog/azure-private-dns-zone-virtual-network-link) -- what makes the record resolvable: links the zone to the virtual networks that query it.
