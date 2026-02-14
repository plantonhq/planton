# Azure Front Door Profile Terraform Module

This Terraform module deploys Azure Front Door profiles with endpoints, origin groups, origins, and routes for global CDN and application delivery.

## Usage

### Standalone Module Usage

```hcl
module "frontdoor_profile" {
  source = "./path/to/module"

  metadata = {
    name = "myapp-cdn"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    resource_group           = "prod-networking-rg"
    name                     = "myapp-cdn-fd"
    sku                      = "Standard_AzureFrontDoor"
    response_timeout_seconds = 120

    endpoints = [
      { name = "main-endpoint", enabled = true }
    ]

    origin_groups = [
      {
        name = "web-backend"
        health_probe = {
          protocol            = "Https"
          path                = "/health"
          request_type        = "HEAD"
          interval_in_seconds = 30
        }
        origins = [
          {
            name               = "app-service"
            host_name          = "myapp.azurewebsites.net"
            origin_host_header = "myapp.azurewebsites.net"
          }
        ]
      }
    ]

    routes = [
      {
        name              = "catch-all"
        endpoint_name     = "main-endpoint"
        origin_group_name = "web-backend"
        patterns_to_match = ["/*"]
        supported_protocols = ["Http", "Https"]
      }
    ]
  }
}
```

### Via OpenMCF CLI

```bash
# Deploy using OpenTofu/Terraform
openmcf tofu apply \
  --manifest frontdoor.yaml \
  --auto-approve
```

## Required Azure Credentials

The module requires Azure authentication. Set these environment variables:

```bash
export ARM_CLIENT_ID="your-client-id"
export ARM_CLIENT_SECRET="your-client-secret"
export ARM_TENANT_ID="your-tenant-id"
export ARM_SUBSCRIPTION_ID="your-subscription-id"
```

Or use Azure CLI authentication:

```bash
az login
```

## Variables

### metadata (required)

| Attribute | Type | Description |
|-----------|------|-------------|
| name | string | Resource name |
| id | string | Optional resource ID |
| org | string | Organization name (for tagging) |
| env | string | Environment name (for tagging) |

### spec (required)

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| resource_group | string | - | Azure Resource Group name |
| name | string | - | Front Door profile name (globally unique) |
| sku | string | Standard_AzureFrontDoor | SKU tier |
| response_timeout_seconds | number | 120 | Origin response timeout (16-240s) |
| endpoints | list | - | Front Door endpoints |
| origin_groups | list | - | Origin groups with origins |
| routes | list | - | Routes connecting endpoints to origin groups |

## Outputs

| Output | Description |
|--------|-------------|
| profile_id | Azure Resource Manager ID of the profile |
| profile_name | The profile name |
| resource_guid | Front Door service GUID |
| endpoint_ids | Map of endpoint names to resource IDs |
| endpoint_hostnames | Map of endpoint names to generated hostnames (*.azurefd.net) |

## Resources Created

- `azurerm_cdn_frontdoor_profile` - The Front Door profile
- `azurerm_cdn_frontdoor_endpoint` - Endpoints (public entry points)
- `azurerm_cdn_frontdoor_origin_group` - Origin groups (backend pools)
- `azurerm_cdn_frontdoor_origin` - Origins (individual backends)
- `azurerm_cdn_frontdoor_route` - Routes (URL-to-backend mappings)

## Notes

- Profile names must be globally unique (2-46 characters, alphanumeric and hyphens)
- Front Door is a global resource -- it does not have a region setting
- Private link to origins requires the `Premium_AzureFrontDoor` SKU
- After provisioning private link origins, the connection must be approved on the target resource
- TLS 1.2 is the minimum supported version
