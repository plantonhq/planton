# Azure Private DNS Zone

Deploys an Azure Private DNS Zone: name resolution inside virtual networks without running a DNS server. The zone is deliberately just the zone — a global record container. Which networks can resolve it is declared through separate **AzurePrivateDnsZoneVirtualNetworkLink** resources referencing this zone's `zone_id` output: one link per network, added and removed without touching the zone.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Private DNS Zone** -- a global Azure DNS zone (no region) for private name resolution, named for the domain it answers
- **SOA Record customization** (optional) -- when the spec carries an `soaRecord` block, the zone's Start of Authority record is created with your contact email and timers instead of Azure's defaults; this is a create-time-only decision
- **Azure Tags** -- your governance tags merged over the Planton-derived resource tags (organization, environment, resource id); a user tag with the same key wins

A zone with no network links answers nobody: pair every deployment with at least one **AzurePrivateDnsZoneVirtualNetworkLink** referencing the zone's output.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the Private DNS Zone will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef. Zones are global resources, but ARM still homes them in a group -- usually the shared network one.
- **The correct zone name** for your use case. For Private Link scenarios, use the exact Azure-defined zone name for the target service (e.g., `privatelink.postgres.database.azure.com`) -- the FQDN must match precisely or private endpoints cannot register where clients look. For custom internal DNS, use any valid lowercase domain name (e.g., `corp.internal`).

## Deploy

### Console

Open the deployment store, find **Azure Private DNS Zone**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the zone's three steps: zone identity (a purpose-driven picker fills exact privatelink FQDNs so long domain names cannot be mistyped), optional SOA customization, and governance tags. Start from the **Private Link Zone** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePrivateDnsZone
metadata:
  name: pg-dns
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "acme-network-rg"
  name: privatelink.postgres.database.azure.com
```

```shell
planton apply -f private-dns-zone.yaml
```

This creates a Private DNS Zone for PostgreSQL Private Link resolution with Azure's standard SOA record. An Infra Job tracks the provisioning in real time. To make the zone resolvable, follow with an AzurePrivateDnsZoneVirtualNetworkLink per network.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Private DNS Zone to a resource group deployed in the same InfraPipeline -- and wire the link and private-endpoint satellites to this zone:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the zone; downstream links and private endpoints reference `status.outputs.zone_id`.

## Key Configuration

These are the most important decisions when configuring a Private DNS Zone. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Zone name** -- The name IS the domain the zone answers for, and changing it replaces the zone and every record in it. For Private Link scenarios, use the exact Azure-defined privatelink zone name for the target service: `privatelink.postgres.database.azure.com` for PostgreSQL Flexible Server, `privatelink.mysql.database.azure.com` for MySQL Flexible Server, `privatelink.database.windows.net` for Azure SQL, `privatelink.blob.core.windows.net` for Blob Storage, `privatelink.vaultcore.azure.net` for Key Vault, `privatelink.documents.azure.com` for Cosmos DB, `privatelink.redis.cache.windows.net` for Azure Cache for Redis. For custom internal DNS, use any valid lowercase domain. One zone per service type per network topology is the norm -- every PostgreSQL private endpoint in the same topology registers into the SAME zone.

**SOA record** (`soaRecord`) -- Nearly every deployment leaves this unset and takes Azure's defaults. Set it only when operational tooling requires a specific contact email or negative-caching behavior (`minimumTtl` -- how long resolvers cache "name does not exist", the one timer with day-to-day impact). The SOA is created with the zone and cannot be customized afterwards.

**Network reach is NOT configured here** -- the zone has no VNet fields. Deploy one **AzurePrivateDnsZoneVirtualNetworkLink** per network that should resolve the zone (that satellite also owns VM auto-registration for custom internal zones).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `zone_id` | Azure resource ID of the Private DNS Zone | AzurePrivateDnsZoneVirtualNetworkLink (`privateDnsZoneId`), AzurePrivateEndpoint DNS zone group registration (`privateDnsZoneIds`), PostgreSQL/MySQL Flexible Server VNet integration (`privateDnsZoneId`) |
| `zone_name` | Name of the Private DNS Zone (e.g., `privatelink.postgres.database.azure.com`) | DNS record creation, tooling that addresses records by zone name |
| `resource_group_name` | The resource group the zone lives in | Tooling that addresses records or links by zone name + resource group rather than parsing the ARM ID |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Private Link DNS** -- A zone named for the Azure-defined privatelink FQDN of the target service, paired with a VNet link per network and referenced from each private endpoint's `privateDnsZoneIds`. The endpoint registers its private IP here automatically, so the service's public FQDN resolves privately inside linked networks. Start from the **Private Link Zone** preset.

**Custom internal DNS** -- A zone like `corp.internal` for internal service discovery, with a real DNS-admin contact and a low negative-caching TTL pinned in the SOA. Pair with links whose `registrationEnabled: true` auto-registers VM A-records. Start from the **Custom Internal DNS Zone** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the Private DNS Zone is created
- [**Azure Private DNS Zone Virtual Network Link**](/infra-catalog/azure-private-dns-zone-virtual-network-link) -- gives the zone reach: one link per network that should resolve it
- [**Azure Private Endpoint**](/infra-catalog/azure-private-endpoint) -- registers its private IP into this zone via `privateDnsZoneIds` so service FQDNs resolve privately
- [**Azure PostgreSQL Flexible Server**](/infra-catalog/azure-postgresql-flexible-server) / [**Azure MySQL Flexible Server**](/infra-catalog/azure-mysql-flexible-server) -- reference this zone for VNet-integrated deployment
