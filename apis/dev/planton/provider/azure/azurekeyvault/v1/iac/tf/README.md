# Azure Key Vault - Terraform Module

This Terraform module provisions an Azure Key Vault with comprehensive security configurations including RBAC, network controls, purge protection, and secret management.

## Overview

This module creates:
- An Azure Key Vault with configurable SKU (Standard or Premium)
- Network access controls (firewall rules, VNet integration)
- Placeholder secrets (values must be set separately for security)
- RBAC authorization (recommended over legacy access policies)
- Purge protection and soft delete for production safety

## Prerequisites

- Terraform >= 1.0
- Azure CLI configured with appropriate credentials
- An existing Azure Resource Group
- Appropriate Azure RBAC permissions (Key Vault Contributor or Owner)

## Usage

### Minimal Configuration

```hcl
module "key_vault" {
  source = "./path/to/module"

  metadata = {
    name = "myapp-prod-kv"
    org  = "myorg"
    env  = "production"
  }

  spec = {
    region         = "eastus"
    resource_group = "prod-security-rg"
    secret_names   = ["database-password", "api-key"]
  }
}
```

### Production Configuration with Network Security

```hcl
module "key_vault_prod" {
  source = "./path/to/module"

  metadata = {
    name = "enterprise-prod-kv"
    id   = "kv-prod-001"
    org  = "enterprise"
    env  = "production"
  }

  spec = {
    region         = "eastus"
    resource_group = "prod-security-rg"
    sku            = "premium" # HSM-backed for compliance

    # Security settings
    enable_rbac_authorization  = true
    enable_purge_protection    = true
    soft_delete_retention_days = 90

    # Network security
    network_acls = {
      default_action        = "Deny"
      bypass_azure_services = true
      ip_rules              = ["203.0.113.0/24", "198.51.100.42"]
      virtual_network_subnet_ids = [
        "/subscriptions/{sub-id}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{vnet}/subnets/{subnet}"
      ]
    }

    # Secrets to create (values set separately)
    secret_names = [
      "database-connection-string",
      "api-key",
      "jwt-secret",
      "encryption-key"
    ]
  }
}
```

### Development Configuration

```hcl
module "key_vault_dev" {
  source = "./path/to/module"

  metadata = {
    name = "myapp-dev-kv"
    env  = "development"
  }

  spec = {
    region         = "eastus"
    resource_group = "dev-rg"
    sku            = "standard"

    # Relaxed security for development
    enable_rbac_authorization  = true
    enable_purge_protection    = false # Allow deletion for cleanup
    soft_delete_retention_days = 7     # Minimum retention

    # Allow access from office IP
    network_acls = {
      default_action        = "Deny"
      bypass_azure_services = true
      ip_rules              = ["203.0.113.0/24"] # Office IP range
    }

    secret_names = ["dev-api-key"]
  }
}
```

## Inputs

### metadata

Object containing resource metadata:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes | - | Resource name (used to generate vault name) |
| `id` | string | No | - | Unique identifier for the resource |
| `org` | string | No | - | Organization name for tagging |
| `env` | string | No | - | Environment (dev, staging, production) |
| `labels` | map(string) | No | {} | Additional labels |
| `tags` | list(string) | No | [] | Additional tags |
| `version` | object | No | - | Version information |

### spec

Object containing Key Vault specification:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | - | Azure region (e.g., "eastus", "westeurope") |
| `resource_group` | string | Yes | - | Azure Resource Group name |
| `sku` | string | No | "standard" | SKU tier: "standard" or "premium" |
| `enable_rbac_authorization` | bool | No | true | Use Azure RBAC instead of access policies |
| `enable_purge_protection` | bool | No | true | Prevent permanent deletion |
| `soft_delete_retention_days` | number | No | 90 | Soft delete retention (7-90 days) |
| `network_acls` | object | No | (see below) | Network access controls |
| `secret_names` | list(string) | No | [] | List of secret names to create |

### spec.network_acls

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `default_action` | string | No | "Deny" | Default network action: "Allow" or "Deny" |
| `bypass_azure_services` | bool | No | true | Allow trusted Azure services |
| `ip_rules` | list(string) | No | [] | Allowed IP addresses or CIDR ranges |
| `virtual_network_subnet_ids` | list(string) | No | [] | Allowed VNet subnet resource IDs |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `vault_id` | string | Azure Resource Manager ID of the Key Vault |
| `vault_name` | string | Name of the Key Vault |
| `vault_uri` | string | URI for accessing the vault (https://{name}.vault.azure.net/) |
| `secret_id_map` | map(string) | Map of secret names to their full secret IDs |
| `region` | string | Azure region where vault was deployed |
| `resource_group` | string | Resource group name |

## Important Notes

### Secret Value Management

**CRITICAL**: This module creates placeholder secrets with empty values. The actual secret values **must be set separately** using one of these methods:

1. **Azure CLI**:
   ```bash
   az keyvault secret set --vault-name myapp-prod-kv --name database-password --value "actual-password"
   ```

2. **Azure Portal**: Navigate to Key Vault → Secrets → Select secret → New Version

3. **Azure SDK**: Use Key Vault SDK in your application code

4. **Terraform** (not recommended for secrets):
   If you must manage values in Terraform, use:
   ```hcl
   resource "azurerm_key_vault_secret" "password" {
     name         = "password"
     value        = var.secret_value # From secure variable
     key_vault_id = module.key_vault.vault_id
   }
   ```

### Vault Naming Constraints

Azure Key Vault names must be:
- 3-24 characters long
- Alphanumeric and hyphens only
- Globally unique across all of Azure
- This module automatically sanitizes `metadata.name` to meet these requirements

### SKU Selection

- **Standard**: Software-protected keys, suitable for most applications ($0.03/10k operations)
- **Premium**: HSM-backed keys (FIPS 140-2 Level 3), required for:
  - PCI-DSS compliance
  - HIPAA/HITRUST
  - FedRAMP
  - Financial services regulations

### RBAC vs Access Policies

This module defaults to **Azure RBAC** (recommended). To grant access:

```bash
# Grant secret read access
az role assignment create \
  --role "Key Vault Secrets User" \
  --assignee <user-or-sp-object-id> \
  --scope /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.KeyVault/vaults/{vault}
```

Common roles:
- `Key Vault Administrator`: Full management (not recommended for apps)
- `Key Vault Secrets Officer`: Manage secrets (for admins)
- `Key Vault Secrets User`: Read secrets (for applications)
- `Key Vault Crypto User`: Use keys for crypto operations

## Post-Deployment Steps

### 1. Set Secret Values

```bash
# Set individual secrets
az keyvault secret set \
  --vault-name $(terraform output -raw vault_name) \
  --name database-password \
  --value "YourSecurePassword123!"

# Bulk import from file
for secret in $(cat secrets.txt); do
  name=$(echo $secret | cut -d'=' -f1)
  value=$(echo $secret | cut -d'=' -f2)
  az keyvault secret set --vault-name myapp-prod-kv --name $name --value "$value"
done
```

### 2. Grant Application Access

```bash
# For a managed identity
APP_IDENTITY=$(az webapp identity show --name myapp --resource-group rg --query principalId -o tsv)

az role assignment create \
  --role "Key Vault Secrets User" \
  --assignee $APP_IDENTITY \
  --scope $(terraform output -raw vault_id)
```

### 3. Configure Application

For Azure App Service, use Key Vault references:

```bash
az webapp config appsettings set \
  --name myapp \
  --resource-group rg \
  --settings ConnectionString="@Microsoft.KeyVault(SecretUri=$(terraform output -json secret_id_map | jq -r '.["database-password"]'))"
```

## Best Practices

### Security

1. **Always enable purge protection in production** to prevent accidental permanent deletion
2. **Use RBAC authorization** for better integration with Azure AD and PIM
3. **Restrict network access** - never leave vaults open to the public internet
4. **Enable diagnostic logging** to Azure Monitor or Log Analytics
5. **Rotate secrets regularly** - Key Vault versions secrets automatically
6. **Use managed identities** for application authentication (no credentials in code)

### Network Security Levels

From most secure to least:

1. **Private Endpoint** (best): Vault accessible only via private IP in VNet
2. **VNet Integration + IP Allowlist**: Restrict to specific subnets and IPs
3. **IP Allowlist Only**: Permit only known public IPs
4. **Public Access**: Not recommended for production

### Tagging Strategy

This module automatically tags resources with:
- `resource`: "true"
- `resource_id`: From metadata
- `resource_kind`: "azure_key_vault"
- `resource_name`: From metadata
- `organization`: If provided
- `environment`: If provided

Add custom tags via Azure policy or external tools.

### Secret Management Patterns

#### Development Environment
- Lower security, faster iteration
- Public access with IP restrictions
- Purge protection disabled for easy cleanup
- Short retention (7 days)

#### Production Environment
- Maximum security
- Private endpoint or VNet restrictions only
- Purge protection enabled
- Maximum retention (90 days)
- Audit logging mandatory

#### Multi-Tenant
- One vault per tenant for isolation
- RBAC with tenant-specific service principals
- Network isolation per tenant if needed

## Troubleshooting

### "Forbidden" or "Insufficient Permissions"

**Issue**: Cannot access vault or secrets

**Solution**:
1. Verify you have appropriate RBAC role assignment
2. Check network ACLs permit your source IP
3. Ensure soft-deleted vault isn't blocking name reuse

```bash
# Check your access
az keyvault list --query "[?name=='myapp-prod-kv']"

# List soft-deleted vaults
az keyvault list-deleted

# Recover or purge if needed
az keyvault recover --name myapp-prod-kv
az keyvault purge --name myapp-prod-kv # Only if purge protection disabled
```

### Vault Name Collision

**Issue**: "vault name already exists" or "name not available"

**Solution**: Vault names are globally unique. Use a unique suffix:
```hcl
metadata = {
  name = "myapp-${random_string.suffix.result}-kv"
}
```

### Network Access Denied

**Issue**: Applications can't access vault even with RBAC permissions

**Solution**:
1. Check network ACLs allow source IPs or VNets
2. Verify `bypass_azure_services = true` if calling from Azure services
3. Add calling service's outbound IP to allowlist

### Soft Delete Conflicts

**Issue**: Cannot create vault with same name after deletion

**Solution**: Wait for soft delete retention or recover the deleted vault:
```bash
az keyvault recover --name myapp-prod-kv
```

## Advanced Topics

### Private Endpoint Configuration

For maximum security, deploy with private endpoint (requires additional resources not in this module):

```hcl
# After creating vault with this module
resource "azurerm_private_endpoint" "kv" {
  name                = "${module.key_vault.vault_name}-pe"
  location            = var.region
  resource_group_name = var.resource_group
  subnet_id           = azurerm_subnet.private.id

  private_service_connection {
    name                           = "${module.key_vault.vault_name}-psc"
    private_connection_resource_id = module.key_vault.vault_id
    is_manual_connection           = false
    subresource_names              = ["vault"]
  }
}
```

### Diagnostic Logging

Enable comprehensive logging:

```hcl
resource "azurerm_monitor_diagnostic_setting" "kv" {
  name               = "kv-diagnostics"
  target_resource_id = module.key_vault.vault_id
  log_analytics_workspace_id = azurerm_log_analytics_workspace.main.id

  log {
    category = "AuditEvent"
    enabled  = true
  }

  metric {
    category = "AllMetrics"
    enabled  = true
  }
}
```

### Secret Rotation Automation

Use Azure Functions or Logic Apps to automate secret rotation:

```python
# Azure Function example
from azure.identity import DefaultAzureCredential
from azure.keyvault.secrets import SecretClient

credential = DefaultAzureCredential()
client = SecretClient(vault_url="https://myapp-prod-kv.vault.azure.net/", credential=credential)

# Generate new password
new_password = generate_secure_password()

# Update secret (creates new version)
client.set_secret("database-password", new_password)

# Update database with new password
update_database_password(new_password)
```

## Resources Created

This module creates:
- `azurerm_key_vault.main` - The Key Vault itself
- `azurerm_key_vault_secret.secrets[*]` - Placeholder secrets (one per secret_name)

## Compliance

Azure Key Vault is certified for:
- **FIPS 140-2**: Level 1 (Standard), Level 3 (Premium)
- **SOC 1/2/3, ISO/IEC 27001**
- **PCI-DSS**: Premium SKU supports compliance
- **HIPAA/HITRUST**: Suitable for PHI encryption keys
- **FedRAMP**: Available in Azure Government regions

This module helps achieve compliance by:
- Enforcing RBAC (least privilege)
- Requiring network restrictions
- Enabling audit logging capabilities
- Preventing accidental deletion (purge protection)

## License

This module is part of the Planton infrastructure framework.

## Support

For issues and feature requests, see the main Planton documentation.

