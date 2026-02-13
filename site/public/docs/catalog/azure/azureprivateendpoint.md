---
title: "Privateendpoint"
description: "Privateendpoint deployment documentation"
icon: "package"
order: 100
componentName: "azureprivateendpoint"
---

# AzurePrivateEndpoint: Research & Deployment Guide

## What is Azure Private Endpoint?

Azure Private Endpoint is a network interface that connects you privately and securely to a service powered by Azure Private Link. It uses a private IP address from your Virtual Network (VNet), effectively bringing the service into your VNet. The service could be an Azure PaaS service (Azure SQL, PostgreSQL, Storage, Key Vault, etc.) or a custom Private Link Service.

Private Endpoints enable three critical capabilities:

1. **Private connectivity** -- Access Azure PaaS services over a private IP instead of the public internet. Traffic stays entirely on the Microsoft backbone network, reducing latency and improving security.

2. **Data exfiltration protection** -- Each private endpoint maps to a specific sub-resource (e.g., "postgresqlServer", "vault", "blob"), not the entire service. Clients can only connect to the specific resource, preventing lateral data access to other resources in the same service account.

3. **Simplified network architecture** -- No need for service endpoints, NAT devices, or public IP addresses to reach Azure services from the VNet. Private endpoints eliminate the complexity of managing public endpoints and firewall rules.

## How Private Link Works

The Private Link flow involves several Azure resources working together:

```
VNet → Subnet → Private Endpoint → Network Interface → Private IP → Target Service
```

1. **VNet** -- The virtual network where your workloads reside
2. **Subnet** -- A dedicated subnet (typically named "pe-subnet" or "private-endpoints") where private endpoints are deployed
3. **Private Endpoint** -- The Azure resource that creates the connection
4. **Network Interface** -- Azure automatically creates a network interface (NIC) for each private endpoint
5. **Private IP** -- A private IP address allocated from the subnet's address space
6. **Target Service** -- The Azure PaaS service (PostgreSQL, Key Vault, Storage, etc.) being accessed privately

When a client in the VNet connects to the service's FQDN (e.g., `myserver.postgres.database.azure.com`), DNS resolution should point to the private IP address. This is where Private DNS Zones come in -- they ensure the FQDN resolves to the private IP instead of the public one.

## Deployment Landscape

### Manual (Azure Portal / CLI)

```bash
# Create private endpoint
az network private-endpoint create \
  --resource-group myRG \
  --name myPE \
  --vnet-name myVnet \
  --subnet pe-subnet \
  --private-connection-resource-id /subscriptions/.../Microsoft.DBforPostgreSQL/flexibleServers/myserver \
  --connection-name myConnection \
  --group-id postgresqlServer

# Create DNS zone group (optional)
az network private-endpoint dns-zone-group create \
  --resource-group myRG \
  --endpoint-name myPE \
  --name myZoneGroup \
  --private-dns-zone /subscriptions/.../privateDnsZones/privatelink.postgres.database.azure.com \
  --zone-name privatelink.postgres.database.azure.com
```

### Terraform

```hcl
resource "azurerm_private_endpoint" "postgres" {
  name                = "pg-private-endpoint"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  subnet_id           = azurerm_subnet.pe.id

  private_service_connection {
    name                           = "pg-connection"
    private_connection_resource_id = azurerm_postgresql_flexible_server.example.id
    subresource_names              = ["postgresqlServer"]
    is_manual_connection           = false
  }

  private_dns_zone_group {
    name                 = "pg-zone-group"
    private_dns_zone_ids = [azurerm_private_dns_zone.postgres.id]
  }
}
```

### Pulumi (Go)

```go
endpoint, _ := network.NewPrivateEndpoint(ctx, "pg-pe", &network.PrivateEndpointArgs{
    Name:                pulumi.String("pg-private-endpoint"),
    ResourceGroupName:   rg.Name,
    Location:            rg.Location,
    SubnetId:            subnet.ID(),
    PrivateServiceConnection: &network.PrivateEndpointPrivateServiceConnectionArgs{
        Name:                           pulumi.String("pg-connection"),
        PrivateConnectionResourceId:     postgresql.ID(),
        SubresourceNames:                pulumi.StringArray{pulumi.String("postgresqlServer")},
        IsManualConnection:              pulumi.Bool(false),
    },
    PrivateDnsZoneGroup: &network.PrivateEndpointPrivateDnsZoneGroupArgs{
        Name:                 pulumi.String("pg-zone-group"),
        PrivateDnsZoneIds:     pulumi.StringArray{zone.ID()},
    },
})
```

## Why OpenMCF Bundles Endpoint + DNS Zone Group

Per DD03 (Composite Bundling Rules), a private endpoint without DNS zone group registration won't resolve correctly in the VNet. The service FQDN will resolve to the public IP instead of the private one, causing clients to bypass the private endpoint entirely.

The bundling follows the same reasoning as:
- AzureNetworkSecurityGroup (NSG + rules -- rules are the substance)
- AzureUserAssignedIdentity (identity + role assignments -- assignments are the substance)
- AzurePrivateDnsZone (zone + VNet link -- link is the substance)

However, the DNS zone group is **optional** in this component because:
1. Some organizations manage DNS externally (e.g., custom DNS servers)
2. Some scenarios don't require DNS resolution (direct IP access)
3. Flexibility is needed for advanced networking patterns

## 80/20 Scoping Rationale

### What's Included

| Feature | Rationale |
|---------|-----------|
| Private endpoint creation | Core resource |
| Optional DNS zone group | Enables seamless DNS resolution (80/20 case) |
| Auto-approved connections | Manual connections require approval workflows (edge case) |
| Dynamic IP allocation | Standard pattern; static IP is niche |
| Polymorphic `private_connection_resource_id` | Supports all Azure PaaS services via StringValueOrRef |
| Sub-resource names | Required for proper service targeting |
| Tags | Standard Azure resource management |

### What's Excluded

| Feature | Rationale |
|---------|-----------|
| Static IP assignment | Dynamic allocation is the standard pattern; static IP is niche |
| Manual connection approval | Requires request messages and owner approval; edge case |
| Multiple DNS zone groups | Single zone group covers 99% of use cases |
| Custom connection names | Auto-derived from metadata.name follows established patterns |
| Application security groups | Advanced networking feature, very niche |
| Request message | Only needed for manual connections (excluded) |

## Design Decisions

**Polymorphic `private_connection_resource_id`**: This field accepts any Azure resource ID that supports Private Link. No `default_kind` annotation is used because the target can be PostgreSQL, MySQL, Key Vault, Storage, Cosmos DB, Redis, or any other Private Link-enabled service. The flexibility enables the component to work across all Azure PaaS services.

**Hardcoded auto-approval**: `is_manual_connection` is hardcoded to `false` in IaC modules. Auto-approved connections are the 80/20 case; manual connections require request messages and owner approval workflows, adding spec complexity for an edge case.

**No static IP**: The `ip_configuration` field (static IP assignment) is omitted per 80/20 scoping. Dynamic allocation from the subnet is the standard pattern, and static IP assignment is a niche requirement.

**Auto-derived names**: Private service connection name and DNS zone group name are auto-derived from `metadata.name` in IaC modules, following the pattern established by AzurePrivateDnsZone (VNet link name auto-derived). This reduces spec complexity and avoids naming conflicts.

**Optional DNS zone group**: The DNS zone group is optional to support flexible DNS management patterns. When provided, it automatically registers the private IP as an A-record in the specified zone. When omitted, DNS is managed externally or via custom configuration.

## Best Practices

1. **Dedicated subnet for private endpoints** -- Create a dedicated subnet (e.g., "pe-subnet") for all private endpoints. This simplifies network policy management and IP address planning.

2. **One endpoint per service instance** -- Each database, Key Vault, or Storage Account instance should have its own private endpoint. Don't try to share endpoints across instances.

3. **Always use DNS zone groups** -- Unless you have a specific reason to manage DNS externally, always provide `private_dns_zone_id` to enable seamless DNS resolution.

4. **Match zone names exactly** -- For Private Link, the DNS zone name must exactly match Azure's predefined name for the service (e.g., `privatelink.postgres.database.azure.com`). A typo means DNS resolution fails silently.

5. **Sub-resource names are service-specific** -- Each Azure service defines its own sub-resource names. Use the correct name for your service type (see Common Sub-Resource Names table in README.md).

6. **Region must match subnet** -- The private endpoint's region must match the subnet's region. Subnets inherit their region from the parent VNet.

## Downstream Consumers

```
AzurePrivateEndpoint
└── (leaf resource -- no current downstream consumers)
```

AzurePrivateEndpoint is currently a leaf resource in the infra chart DAG. No other OpenMCF resources reference its outputs. However, the outputs (`private_endpoint_id`, `private_ip_address`, `network_interface_id`) are essential for operational visibility and potential future consumers.

## Infra Chart Integration

### Database Stack Pattern

The database-stack infra chart creates one private endpoint per database instance:

```
VPC → Subnet → PrivateDnsZone → Database Server → PrivateEndpoint → DNS Zone Group
```

Each database server (PostgreSQL, MySQL, MSSQL, Redis) gets its own private endpoint. The endpoint is wired to:
1. The subnet (for private IP allocation)
2. The database server (via `private_connection_resource_id`)
3. The corresponding private DNS zone (via `private_dns_zone_id`)

The DNS zone group automatically registers the private IP as an A-record in the zone, ensuring the database FQDN resolves to the private IP within the VNet.

### Enterprise Network Foundation

Optional component for organizations that pre-create private endpoints as part of their networking foundation, before any databases or services are deployed. However, private endpoints are typically created alongside their target services, not as standalone networking infrastructure.

---

**Status**: Production Ready
**Azure Provider Version**: ~> 4.0
**Pulumi Provider Version**: v6
