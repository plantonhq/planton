# AzurePrivateEndpoint Terraform Module

## Overview

This Terraform module creates an Azure Private Endpoint that connects to a Private Link-enabled resource. The module optionally associates a Private DNS Zone for automatic DNS resolution.

## Resources Created

- `azurerm_private_endpoint.endpoint` -- The private endpoint resource
  - Includes a `private_service_connection` block for connecting to the Private Link resource
  - Optionally includes a `private_dns_zone_group` block when `private_dns_zone_id` is provided

## Variables

| Variable | Type | Required | Description |
|----------|------|----------|-------------|
| `metadata` | object | Yes | Resource metadata (name, id, org, env, labels, tags, version) |
| `spec` | object | Yes | Private endpoint spec (region, resource_group, name, subnet_id, private_connection_resource_id, subresource_names, private_dns_zone_id) |

### spec Object Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `region` | string | Yes | Azure region where the private endpoint will be created |
| `resource_group` | string | Yes | Name of the Azure resource group |
| `name` | string | Yes | Name of the private endpoint |
| `subnet_id` | string | Yes | Azure Resource Manager ID of the subnet where the endpoint will be created |
| `private_connection_resource_id` | string | Yes | Azure Resource Manager ID of the Private Link-enabled resource to connect to |
| `subresource_names` | list(string) | No | List of subresource names (e.g., `["blob"]` for Storage Account). Defaults to empty list |
| `private_dns_zone_id` | string | No | Azure Resource Manager ID of a Private DNS Zone to associate. If not provided, no DNS zone group is created |

## Outputs

| Output | Description |
|--------|-------------|
| `private_endpoint_id` | Azure Resource Manager ID of the Private Endpoint |
| `private_ip_address` | The private IP address allocated to the Private Endpoint |
| `network_interface_id` | Azure Resource Manager ID of the network interface created for the endpoint |

## Usage

```hcl
module "pg_private_endpoint" {
  source = "./iac/tf"

  metadata = {
    name = "pg-private-endpoint"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region                         = "eastus"
    resource_group                 = "prod-network-rg"
    name                           = "pg-private-endpoint"
    subnet_id                      = "/subscriptions/.../subnets/private-endpoints-subnet"
    private_connection_resource_id = "/subscriptions/.../providers/Microsoft.DBforPostgreSQL/servers/pg-server"
    subresource_names              = ["postgresqlServer"]
    private_dns_zone_id            = "/subscriptions/.../privateDnsZones/privatelink.postgres.database.azure.com"
  }
}
```

## Provider

Requires `hashicorp/azurerm` ~> 4.0.
