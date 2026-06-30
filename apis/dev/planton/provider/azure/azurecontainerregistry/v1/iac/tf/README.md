# Azure Container Registry - Terraform Module

This Terraform module deploys an Azure Container Registry (ACR) instance using the AzureRM provider.

## Features

- **Multiple SKU Support**: Basic, Standard, and Premium tiers
- **Geo-Replication**: Automatic multi-region replication for Premium SKU
- **Secure by Default**: Admin user disabled unless explicitly enabled
- **Zone Redundancy**: Enabled for geo-replications
- **Network Security**: Azure services bypass enabled

## Prerequisites

- Terraform >= 1.0
- Azure CLI authenticated
- Azure subscription with appropriate permissions

## Usage

### Basic Example

```hcl
module "acr_basic" {
  source = "./iac/tf"

  metadata = {
    name = "basic-acr"
    env  = "production"
  }

  spec = {
    region        = "eastus"
    registry_name = "mycompanyacr"
    sku           = "STANDARD"
  }
}
```

### Premium with Geo-Replication

```hcl
module "acr_global" {
  source = "./iac/tf"

  metadata = {
    name = "global-acr"
    labels = {
      environment = "production"
      scope       = "global"
    }
  }

  spec = {
    region               = "eastus"
    registry_name        = "globalacr"
    sku                  = "PREMIUM"
    admin_user_enabled   = false
    geo_replication_regions = [
      "westeurope",
      "southeastasia",
      "australiaeast"
    ]
  }
}
```

### Development with Admin User

```hcl
module "acr_dev" {
  source = "./iac/tf"

  metadata = {
    name = "dev-acr"
    env  = "development"
  }

  spec = {
    region             = "eastus"
    registry_name      = "devacr123"
    sku                = "BASIC"
    admin_user_enabled = true  # Convenience for local dev
  }
}
```

## Input Variables

### metadata

Object containing resource metadata:

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| name | string | yes | - | Resource name |
| env | string | no | "default" | Environment (dev, staging, prod) |
| labels | map(string) | no | {} | Additional labels/tags |

### spec

Object containing ACR specification:

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| region | string | yes | - | Azure region (e.g., "eastus") |
| registry_name | string | yes | - | Registry name (5-50 lowercase alphanumeric) |
| sku | string | no | "STANDARD" | SKU: "BASIC", "STANDARD", or "PREMIUM" |
| admin_user_enabled | bool | no | false | Enable admin user |
| geo_replication_regions | list(string) | no | [] | Replica regions (Premium only) |

## Outputs

| Name | Description |
|------|-------------|
| registry_login_server | Registry login server URL (e.g., "myacr.azurecr.io") |
| registry_resource_id | Azure Resource Manager ID |
| resource_group_name | Resource group name |
| registry_name | Registry name |
| admin_username | Admin username (if enabled, sensitive) |
| admin_password | Admin password (if enabled, sensitive) |

## SKU Selection Guide

### Basic (~$5/month)
- **Storage**: 10 GB
- **Throughput**: ~1,000 pulls/min
- **Use for**: Dev/test only
- **Limitations**: No geo-replication, no private link

### Standard (~$20/month)
- **Storage**: 100 GB
- **Throughput**: ~3,000 pulls/min
- **Use for**: Single-region production
- **Limitations**: No geo-replication

### Premium (~$20/month + $50/replica)
- **Storage**: 500 GB
- **Throughput**: ~10,000 pulls/min
- **Use for**: Global production, enterprise
- **Features**: Geo-replication, private link, content trust

## Geo-Replication

Only available with Premium SKU. Specify additional regions:

```hcl
spec = {
  sku = "PREMIUM"
  geo_replication_regions = [
    "westeurope",
    "southeastasia"
  ]
}
```

**Benefits:**
- Reduced latency for global deployments
- Lower cross-region bandwidth costs
- Regional redundancy

## Authentication

### From Azure CLI
```shell
az acr login --name <registry-name>
```

### From AKS
```shell
az aks update \
  --name <cluster> \
  --resource-group <rg> \
  --attach-acr <registry-name>
```

### From CI/CD
Use service principal with AcrPush/AcrPull roles.

## Common Operations

### Push Image
```shell
docker tag myapp:v1.0 myacr.azurecr.io/myapp:v1.0
docker push myacr.azurecr.io/myapp:v1.0
```

### Pull Image
```shell
docker pull myacr.azurecr.io/myapp:v1.0
```

### List Images
```shell
az acr repository list --name myacr
```

### Show Tags
```shell
az acr repository show-tags --name myacr --repository myapp
```

## Security Best Practices

✅ **Disable admin user in production** - Use Azure AD authentication  
✅ **Use managed identities for AKS** - No secrets to manage  
✅ **Enable Defender for Cloud** - Automatic vulnerability scanning  
✅ **Implement retention policies** - Auto-delete old/untagged images  
✅ **Use Premium for sensitive workloads** - Private Link and content trust  

## Cost Optimization

- **Use Basic for dev/test** - Save ~75% vs Standard
- **Clean up old images** - Storage costs add up
- **Evaluate geo-replication needs** - Each replica adds ~$50/month
- **Consider Azure Reservations** - Save up to 30% with 1-year commitment

## Limitations

- **Registry name is globally unique** - Must be unique across all Azure
- **Name is immutable** - Cannot rename after creation
- **SKU upgrades only** - Can upgrade Basic→Standard→Premium, not downgrade
- **Geo-replication requires Premium** - Cannot add replicas to Basic/Standard

## Troubleshooting

### Name Conflict
```
Error: Registry name already exists
```
**Solution**: Choose a different registry name (must be globally unique)

### Geo-Replication on Non-Premium
```
Error: Geo-replication requires Premium SKU
```
**Solution**: Set `sku = "PREMIUM"`

### Admin Credentials Not Found
```
Error: Admin user is disabled
```
**Solution**: Set `admin_user_enabled = true` (dev/test only)

## References

- [Main Component Documentation](../../README.md)
- [Research Documentation](../../docs/README.md)
- [Pulumi Module Overview](./overview.md)
- [Azure ACR Documentation](https://docs.microsoft.com/en-us/azure/container-registry/)

