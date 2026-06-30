# Azure Storage Account Terraform Module

This Terraform module deploys Azure Storage Accounts with blob containers, network access controls, and data protection features.

## Usage

### Standalone Module Usage

```hcl
module "storage_account" {
  source = "./path/to/module"

  metadata = {
    name = "myapp-storage"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region         = "eastus"
    resource_group = "myapp-rg"
    account_kind   = "StorageV2"
    account_tier   = "Standard"
    replication_type = "GRS"
    access_tier    = "Hot"

    network_rules = {
      default_action        = "Deny"
      bypass_azure_services = true
      ip_rules              = ["203.0.113.0/24"]
    }

    blob_properties = {
      enable_versioning          = true
      soft_delete_retention_days = 30
    }

    containers = [
      { name = "data", access_type = "private" },
      { name = "logs", access_type = "private" }
    ]
  }
}
```

### Via Planton CLI

```bash
# Deploy using OpenTofu/Terraform
planton tofu apply \
  --manifest storage.yaml \
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
| name | string | Resource name (used for storage account naming) |
| id | string | Optional resource ID |
| org | string | Organization name (for tagging) |
| env | string | Environment name (for tagging) |

### spec (required)

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| region | string | - | Azure region |
| resource_group | string | - | Resource group name |
| account_kind | string | StorageV2 | Storage account kind |
| account_tier | string | Standard | Performance tier |
| replication_type | string | LRS | Replication strategy |
| access_tier | string | Hot | Default blob access tier |
| enable_https_traffic_only | bool | true | Require HTTPS |
| min_tls_version | string | TLS1_2 | Minimum TLS version |
| network_rules | object | - | Network ACL configuration |
| blob_properties | object | - | Blob service properties |
| containers | list | [] | Blob containers to create |

## Outputs

| Output | Description |
|--------|-------------|
| storage_account_id | Azure Resource Manager ID |
| storage_account_name | Storage account name |
| primary_blob_endpoint | Primary blob service URL |
| primary_queue_endpoint | Primary queue service URL |
| primary_table_endpoint | Primary table service URL |
| primary_file_endpoint | Primary file service URL |
| primary_dfs_endpoint | Primary DFS (Data Lake) URL |
| primary_web_endpoint | Primary static website URL |
| container_url_map | Map of container names to URLs |
| region | Azure region |
| resource_group | Resource group name |

## Notes

- Storage account names must be globally unique (3-24 lowercase alphanumeric characters)
- Network rules default to "Deny" for security
- Blob soft delete is enabled by default (7 days)
- TLS 1.2 is enforced by default
