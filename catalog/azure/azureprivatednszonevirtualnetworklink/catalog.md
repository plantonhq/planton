# Azure Private DNS Zone Virtual Network Link

Deploys an Azure Private DNS Zone Virtual Network Link — the attachment that makes a private DNS zone's records resolvable from inside one virtual network. A zone without links answers nobody; each link adds exactly one network to its audience. The link is a first-class resource because it is many-per-zone with its own lifecycle: a hub-and-spoke topology links one zone (say, `privatelink.postgres.database.azure.com`) to the hub and every spoke network, and networks join and leave the topology without touching the zone or each other. **One link resource per zone-network pair.**

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Virtual Network Link** -- the attachment written on the zone, with your registration and resolution-policy dials
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically and merged with the user tags

The link is an ARM **child of the zone** — it carries no region and no resource group of its own; both derive from the zone's ARM ID and can never contradict it. Peered networks do NOT inherit a link: each network in a topology needs its own.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A Private DNS Zone** to link — reference an AzurePrivateDnsZone Infra Component's `zone_id` output, or provide the full ARM ID of an existing zone.
- **A Virtual Network** to make the zone resolvable from — reference an AzureVirtualNetwork's `virtual_network_id` output, or provide the full ARM ID.

## Deploy

### Console

Open the deployment store, find **Azure Private DNS Zone Virtual Network Link**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Private Link Zone Attachment** preset in the [Presets](#presets) tab for the flagship Private Link attachment.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePrivateDnsZoneVirtualNetworkLink
metadata:
  name: pg-privatelink-hub-link
  org: acme-corp
  env: prod
spec:
  name: hub-vnet
  privateDnsZoneId:
    valueFrom:
      kind: AzurePrivateDnsZone
      name: pg-privatelink
      fieldPath: status.outputs.zone_id
  virtualNetworkId:
    valueFrom:
      kind: AzureVirtualNetwork
      name: hub-vnet
      fieldPath: status.outputs.virtual_network_id
```

```shell
planton apply -f link.yaml
```

This writes the link on the referenced zone: workloads in `hub-vnet` can now resolve the zone's records — and only that network; each spoke gets its own link. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the link to its dependencies:

```yaml
spec:
  privateDnsZoneId:
    valueFrom:
      kind: AzurePrivateDnsZone
      name: pg-privatelink
      fieldPath: status.outputs.zone_id
  virtualNetworkId:
    valueFrom:
      kind: AzureVirtualNetwork
      name: hub-vnet
      fieldPath: status.outputs.virtual_network_id
```

The InfraPipeline resolves the dependency graph, deploys the zone and network first, then writes the link on the zone.

## Key Configuration

These are the most important decisions when configuring a Virtual Network Link. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Registration** -- `registrationEnabled` decides whether Azure auto-registers VM records: each VM in the linked network gets an A record at boot, removed when it goes away. Right for custom internal zones ("corp.internal"); **keep it off (the default) for privatelink.\* zones** — their records are managed by private endpoints — and note Azure allows only ONE registration-enabled link per network.

**Resolution policy** -- what happens when the private zone cannot answer: `DEFAULT` answers strictly from the private zone (unresolvable names fail); `NX_DOMAIN_REDIRECT` retries against public DNS — the fallback pattern for Private Link zones shared across environments where some records exist only publicly. Unset lets Azure choose its per-zone-type default.

**Name** -- unique per zone, up to 80 characters; since each link attaches exactly one network, name it after that network ("hub-vnet", "spoke-payments") so the zone's link list reads as its audience. Renaming replaces the link (a brief resolution gap for the affected network, nothing else).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzurePrivateDnsZone** | `privateDnsZoneId` | `status.outputs.zone_id` |
| **AzureVirtualNetwork** | `virtualNetworkId` | `status.outputs.virtual_network_id` |

### What This Kind Provides

`status.outputs` carries `link_id` (the child path under the zone) and `link_name`, but the link is a leaf — no downstream Infra Component references it. Its effect is entirely side-band: names in the zone start resolving from inside the linked network.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Privatelink zone link** -- attach a shared `privatelink.*` zone to a network so its private endpoints resolve; registration off, records owned by the endpoints. Start from the **Private Link Zone Attachment** preset.

**Internal zone with auto-registration** -- a custom zone ("corp.internal") where every VM registers its hostname at boot — remember: one registration link per network. Start from the **Internal Zone with VM Auto-Registration** preset.

**Public fallback** -- the NX_DOMAIN_REDIRECT policy for shared zones where some records exist only publicly. Start from the **Shared Zone with Public DNS Fallback** preset.

## Works With

- [**Azure Private DNS Zone**](/infra-catalog/azure-private-dns-zone) -- the zone this link is written on, referenced by its `zone_id` output
- [**Azure Virtual Network**](/infra-catalog/azure-virtual-network) -- the network that gains resolution, referenced by its `virtual_network_id` output
- [**Azure Private Endpoint**](/infra-catalog/azure-private-endpoint) -- writes the privatelink records this link makes resolvable
- [**Azure Virtual Network Peering**](/infra-catalog/azure-virtual-network-peering) -- connects networks, but does NOT propagate DNS links — each peered network still needs its own link
