# Azure DNS Zone

Deploys a public Azure DNS Zone: an internet-facing, authoritative DNS zone hosted on Azure's global anycast name-server fleet. The zone is deliberately just the zone — an empty record container plus its Start of Authority settings. Records are declared through separate **AzureDnsRecord** resources referencing this zone's `zone_name` output: one resource per record set, added and removed without touching the zone. Creating the zone does not make it authoritative — the domain resolves through it only after the Azure-assigned name servers are configured at the registrar.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Public DNS Zone** -- a global Azure DNS zone (no region) for the specified domain, with four Azure-assigned name servers
- **SOA Record customization** (optional) -- when the spec carries an `soaRecord` block, the zone's Start of Authority record carries your contact email and timers instead of Azure's defaults; unlike a private zone's SOA, every field updates in place
- **Azure Tags** -- your governance tags merged over the Planton-derived resource tags (organization, environment, resource id); a user tag with the same key wins

Creating the zone does NOT make it authoritative on the internet: the domain resolves through this zone only once the name servers from `status.outputs.nameservers` are configured at the domain's registrar (or as NS records in the parent zone, for a subdomain delegation).

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the DNS zone will be created. Provide the name directly or reference an AzureResourceGroup Cloud Resource via ValueFromRef. Zones are global resources, but ARM still homes them in a group -- the group governs who can manage the zone, not how it resolves.
- **Domain ownership at a registrar** -- Azure hosts whatever zone name you give it (the same name can exist in many subscriptions, each with a different name-server set), but only the copy delegated at the registrar answers the internet.

## Deploy

### Console

Open the deployment store, find **Azure DNS Zone**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the zone's three steps: zone identity (the domain and its resource group, with the delegation handoff explained inline), optional SOA customization, and governance tags.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDnsZone
metadata:
  name: example-zone
  org: acme-corp
  env: prod
spec:
  zoneName: example.com
  resourceGroup:
    value: "acme-prod-rg"
```

```shell
planton apply -f dns-zone.yaml
```

This creates an empty public zone for `example.com` with Azure's standard SOA record. A Stack Job tracks the provisioning in real time. Capture `status.outputs.nameservers` after the first deployment and configure those four servers at the registrar to make the zone live; then declare records as AzureDnsRecord resources.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the DNS zone to a resource group deployed in the same InfraPipeline -- and wire AzureDnsRecord satellites to this zone:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the zone; downstream AzureDnsRecord resources reference `status.outputs.zone_name`.

## Key Configuration

These are the most important decisions when configuring a DNS zone. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Zone name** -- The name IS the internet domain the zone answers for: an apex domain (`example.com`) delegated at the registrar, or a subdomain (`team.example.com`) delegated with NS records in the parent zone. At least two dot-separated lowercase labels, no trailing dot. Changing the name replaces the zone -- destroying every record in it and the assigned name-server set, which breaks the registrar delegation until it is updated.

**SOA record** (`soaRecord`) -- Nearly every deployment leaves this unset and takes Azure's defaults. Set it only when operational tooling requires a specific contact email or negative-caching behavior (`minimumTtl` -- how long resolvers cache "name does not exist"; Azure's default is 300 seconds, worth lowering on zones that back automated certificate flows). All SOA fields update in place, so skipping this at creation costs nothing.

**Records are NOT configured here** -- the zone has no record fields. Declare each record set as an **AzureDnsRecord** resource referencing the zone's `zone_name` output, so records gain independent lifecycle and change tracking.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `nameservers` | The four Azure-assigned name servers | Registrar NS delegation -- the handoff that makes the zone authoritative |
| `zone_name` | The DNS zone name (e.g., `example.com`) | AzureDnsRecord `zoneName` field via ValueFromRef -- the record join seam |
| `resource_group_name` | The resource group the zone lives in | AzureDnsRecord `resourceGroup` field -- the other half of the record join seam |
| `zone_id` | Azure Resource Manager ID of the DNS zone | Kinds that manage the zone as a whole — alias records targeting the zone, diagnostic settings, Azure Policy assignments |

`max_number_of_record_sets` is also exported — a per-zone capacity fact (10000 by default), not a live count; it has no ValueFromRef consumers.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Public zone** -- An empty, internet-facing zone with Azure's SOA defaults, ready for records and registrar delegation. The standard production pattern: records live as independent AzureDnsRecord resources. Start from the **Public Zone** preset.

**Delegation-ready zone** -- The same zone tuned for active operations: the SOA contact points at your DNS team and the negative-caching TTL is lowered so newly created records (certificate validation records especially) become visible fast. Start from the **Delegation-Ready Zone** preset.

## Works With

- [**Azure Resource Group**](/cloud-catalog/azure-resource-group) -- provides the resource group where the DNS zone is created
- [**Azure DNS Record**](/cloud-catalog/azure-dns-record) -- declares each record set in the zone, one resource per record, referencing `zone_name` and `resource_group_name`
