# AzurePrivateDnsZone: Research & Deployment Guide

## What is Azure Private DNS?

Azure Private DNS provides a reliable, secure DNS service for virtual networks. It manages and resolves domain names within a virtual network without the need for a custom DNS solution. Private DNS zones are the Azure counterpart to internal DNS zones in traditional networking -- they enable name resolution that is scoped to one or more VNets and invisible from the public internet.

### Two Deployment Models

**1. Private Link DNS (Primary Use Case)**

Azure Private Link enables access to Azure PaaS services (Azure SQL, PostgreSQL, Storage, Key Vault, etc.) over a private IP address in the VNet. Each Azure service that supports Private Link has a predefined privatelink DNS zone name (e.g., `privatelink.postgres.database.azure.com`). When a Private Endpoint is created:

1. Azure allocates a private IP from the endpoint's subnet
2. A DNS A-record is created (or should be created) mapping the service's FQDN to this private IP
3. Clients in the VNet resolve the service FQDN to the private IP and communicate entirely over the private network

Without a properly configured private DNS zone and VNet link, the FQDN resolves to the public IP, completely bypassing the Private Endpoint.

**2. Custom Internal DNS**

For non-Azure-service use cases, private DNS zones enable custom internal name resolution. A zone like `contoso.internal` can host A, AAAA, CNAME, MX, SRV, TXT, and PTR records accessible only from linked VNets. With auto-registration enabled, VMs in the linked VNet automatically get A-records created and removed as they are created and deleted.

## Deployment Landscape

### Manual (Azure Portal / CLI)

```bash
# Create zone
az network private-dns zone create \
  --resource-group myRG \
  --name privatelink.postgres.database.azure.com

# Create VNet link
az network private-dns link vnet create \
  --resource-group myRG \
  --zone-name privatelink.postgres.database.azure.com \
  --name myVnetLink \
  --virtual-network /subscriptions/.../virtualNetworks/myVnet \
  --registration-enabled false
```

### Terraform

```hcl
resource "azurerm_private_dns_zone" "postgres" {
  name                = "privatelink.postgres.database.azure.com"
  resource_group_name = azurerm_resource_group.example.name
}

resource "azurerm_private_dns_zone_virtual_network_link" "postgres" {
  name                  = "postgres-vnet-link"
  resource_group_name   = azurerm_resource_group.example.name
  private_dns_zone_name = azurerm_private_dns_zone.postgres.name
  virtual_network_id    = azurerm_virtual_network.example.id
  registration_enabled  = false
}
```

### Pulumi (Go)

```go
zone, _ := privatedns.NewZone(ctx, "postgres", &privatedns.ZoneArgs{
    Name:              pulumi.String("privatelink.postgres.database.azure.com"),
    ResourceGroupName: rg.Name,
})

privatedns.NewZoneVirtualNetworkLink(ctx, "postgres-link", &privatedns.ZoneVirtualNetworkLinkArgs{
    Name:               pulumi.String("postgres-vnet-link"),
    ResourceGroupName:  rg.Name,
    PrivateDnsZoneName: zone.Name,
    VirtualNetworkId:   vnet.ID(),
    RegistrationEnabled: pulumi.Bool(false),
})
```

## Why Planton Bundles Zone + VNet Link

Per DD03 (Composite Bundling Rules), a private DNS zone without a VNet link is unreachable from any VNet. The zone exists but serves no purpose -- no resources can resolve records in it. This makes the VNet link a structural dependency, not an independent resource.

The bundling follows the same reasoning as:
- AzureNetworkSecurityGroup (NSG + rules -- rules are the substance)
- AzureUserAssignedIdentity (identity + role assignments -- assignments are the substance)

## 80/20 Scoping Rationale

### What's Included

| Feature | Rationale |
|---------|-----------|
| Zone creation | Core resource |
| Single VNet link | Minimum viable DNS resolution |
| `registration_enabled` toggle | Enables both privatelink and custom DNS use cases |
| Tags | Standard Azure resource management |

### What's Excluded

| Feature | Rationale |
|---------|-----------|
| Multiple VNet links | Advanced hub-spoke scenario; 80/20 covers single-VNet |
| SOA record customization | Azure defaults are correct for 99% of use cases |
| DNS record creation | Records are managed by Private Endpoints or other resources |
| Resolution policy | `NxDomainRedirect` is a niche DNS resolver feature |

### Design Decisions

**No `region` field**: Private DNS zones are global Azure resources. Unlike most Azure resources that are deployed to a specific region, private DNS zones exist at the subscription level and are accessible from any VNet in the subscription.

**Required `resource_group`**: Added during research (not in original T02 spec). Every Azure resource requires a resource group, and the Terraform/Pulumi providers require `resource_group_name` as a mandatory parameter.

**Required `vnet_id`**: The VNet link is mandatory because a zone without a link is useless. Making it required ensures users don't accidentally create orphaned zones.

**Optional `registration_enabled`**: Defaults to `false` because the primary use case (privatelink zones) should never have auto-registration. Exposing it as optional enables the secondary use case (custom internal DNS) without adding complexity.

## Best Practices

1. **One zone per service type** -- Create a separate privatelink zone for each Azure service (PostgreSQL, MySQL, Key Vault, etc.). Don't try to share zones across service types.

2. **Same resource group as networking** -- Place private DNS zones in the same resource group as VNet and networking resources for lifecycle alignment.

3. **Registration for custom zones only** -- Only enable `registration_enabled` for custom internal DNS zones (e.g., `contoso.internal`). Never enable it for privatelink zones.

4. **Zone names are exact** -- For Private Link, the zone name must exactly match Azure's predefined name for the service. A typo means DNS resolution fails silently.

5. **VNet link naming** -- The IaC module auto-generates the link name from the resource metadata. This avoids naming conflicts when multiple zones link to the same VNet.

## Downstream Consumers

```
AzurePrivateDnsZone
├── AzurePrivateEndpoint (private_dns_zone_id → zone group)
├── AzurePostgresqlFlexibleServer (private_dns_zone_id → VNet integration)
└── AzureMysqlFlexibleServer (private_dns_zone_id → VNet integration)
```

## Infra Chart Integration

### Database Stack Pattern

The database-stack infra chart creates one privatelink zone per database type:

```
VPC → Subnet → PrivateDnsZone (per DB type) → Database Server → PrivateEndpoint
```

Each database server references its corresponding zone for DNS resolution. The private endpoint then registers its IP in the zone via DNS zone group.

### Enterprise Network Foundation

Optional component for organizations that pre-create private DNS zones as part of their networking foundation, before any databases or services are deployed.

---

**Status**: Production Ready
**Azure Provider Version**: ~> 4.0
**Pulumi Provider Version**: v6
