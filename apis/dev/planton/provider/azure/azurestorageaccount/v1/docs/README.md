# Azure Storage Account Deployment Methods

## Introduction

"We'll just use blob storage for now — it's basically infinite disk space in the cloud." Sound familiar? Azure Storage Account is the foundational storage service that powers everything from simple blob containers to enterprise data lakes, yet properly deploying and configuring it remains surprisingly nuanced.

Azure Storage Account offers multiple storage types (Blob, Queue, Table, File, Data Lake), various replication strategies (from single-datacenter to cross-region), multiple access tiers (Hot, Cool, Archive), and complex networking options. The promise is simple: scalable, durable storage for any workload. The reality involves careful consideration of performance requirements, durability needs, network security, and cost optimization.

This document explores the full spectrum of Azure Storage Account deployment approaches — from quick portal setups to production-hardened Infrastructure-as-Code patterns that satisfy enterprise security requirements. We'll examine what works, what doesn't, and why Planton defaults to certain choices for its Azure Storage Account implementation.

## The Maturity Spectrum: From Manual Clicks to Production IaC

### Level 0: The Portal Deployment (Quick Start, Not Production)

The Azure Portal provides the fastest path to a Storage Account: navigate to Storage Accounts, click "Create," fill out a few fields, and you have storage in minutes. Azure even provides helpful defaults and explanatory tooltips.

This approach works for learning, prototyping, or ad-hoc storage needs. But it reveals limitations quickly:

- **Configuration Drift**: Manual deployments across environments inevitably differ. Your dev storage account might use LRS while prod accidentally got GRS — discovered only when the billing statement arrives.

- **Security Gaps**: The portal makes it easy to leave storage accounts with public access enabled or without proper network restrictions. "Allow access from all networks" is often the default path of least resistance.

- **Naming Challenges**: Storage account names must be globally unique across all of Azure (3-24 characters, lowercase letters and numbers only). Manual creation means trial-and-error naming.

- **No Change History**: When access tier changes unexpectedly affect costs, there's no Git history showing who changed what and when.

**Verdict**: Portal deployment is excellent for exploration and quick demos. Treat it as training wheels before moving to automation.

### Level 1: CLI and PowerShell Scripts (Repeatable, But Imperative)

Azure CLI and PowerShell bring repeatability through scripting:

```bash
az storage account create \
  --name myappprodstorage \
  --resource-group myapp-rg \
  --location eastus \
  --sku Standard_GRS \
  --kind StorageV2 \
  --access-tier Hot \
  --https-only true \
  --min-tls-version TLS1_2 \
  --allow-blob-public-access false

az storage container create \
  --account-name myappprodstorage \
  --name data \
  --public-access off
```

Scripts can be version-controlled, reviewed, and executed in CI/CD pipelines. This is a meaningful step forward from manual operations.

Challenges include:

- **Imperative nature**: Scripts describe steps, not desired state. Running twice might fail or cause unexpected changes without careful idempotency handling.

- **Dependency orchestration**: Create resource group, then storage account, then containers, then network rules. Miss a dependency and deployment fails.

- **State ignorance**: Scripts don't inherently know what exists. Conditional logic ("if not exists") adds complexity.

- **Limited validation**: Many errors surface only at runtime, potentially mid-deployment.

**Verdict**: Scripts work for simple automation but require discipline to make production-ready. Better tools exist.

### Level 2: Azure Resource Manager Templates and Bicep (Azure-Native IaC)

ARM templates (JSON) and Bicep (Microsoft's domain-specific language) provide declarative Infrastructure-as-Code:

```bicep
resource storageAccount 'Microsoft.Storage/storageAccounts@2023-01-01' = {
  name: 'myappstorage${uniqueString(resourceGroup().id)}'
  location: resourceGroup().location
  kind: 'StorageV2'
  sku: {
    name: 'Standard_GRS'
  }
  properties: {
    accessTier: 'Hot'
    supportsHttpsTrafficOnly: true
    minimumTlsVersion: 'TLS1_2'
    allowBlobPublicAccess: false
    networkAcls: {
      defaultAction: 'Deny'
      bypass: 'AzureServices'
    }
  }
}

resource container 'Microsoft.Storage/storageAccounts/blobServices/containers@2023-01-01' = {
  name: '${storageAccount.name}/default/data'
  properties: {
    publicAccess: 'None'
  }
}
```

Key advantages:

- **Declarative**: Describe what you want, Azure figures out how to achieve it
- **Idempotent**: Re-running produces consistent results
- **Validation**: Templates are validated before deployment
- **Parameter files**: Separate environment-specific values from infrastructure definitions

The limitation is Azure lock-in — ARM/Bicep is Azure-only. Multi-cloud organizations often prefer cross-platform tools.

**Verdict**: ARM/Bicep is production-ready for Azure-centric teams. Use Bicep over raw ARM JSON for better readability and developer experience.

### Level 3: Cross-Platform IaC (Terraform, Pulumi)

Terraform and Pulumi provide cloud-agnostic Infrastructure-as-Code with rich ecosystems:

**Terraform Example:**

```hcl
resource "azurerm_storage_account" "main" {
  name                     = "myappprodstorage"
  resource_group_name      = azurerm_resource_group.main.name
  location                 = azurerm_resource_group.main.location
  account_tier             = "Standard"
  account_replication_type = "GRS"
  account_kind             = "StorageV2"
  access_tier              = "Hot"
  
  enable_https_traffic_only       = true
  min_tls_version                 = "TLS1_2"
  allow_nested_items_to_be_public = false
  
  network_rules {
    default_action = "Deny"
    bypass         = ["AzureServices"]
    ip_rules       = ["203.0.113.0/24"]
  }
  
  blob_properties {
    versioning_enabled = true
    delete_retention_policy {
      days = 30
    }
    container_delete_retention_policy {
      days = 30
    }
  }
}

resource "azurerm_storage_container" "data" {
  name                  = "data"
  storage_account_name  = azurerm_storage_account.main.name
  container_access_type = "private"
}
```

**Pulumi Example (Go):**

```go
storageAccount, err := storage.NewStorageAccount(ctx, "main", &storage.StorageAccountArgs{
    ResourceGroupName: resourceGroup.Name,
    Location:          pulumi.String("eastus"),
    AccountTier:       pulumi.String("Standard"),
    AccountReplicationType: pulumi.String("GRS"),
    AccountKind:       pulumi.String("StorageV2"),
    AccessTier:        pulumi.String("Hot"),
    EnableHttpsTrafficOnly: pulumi.Bool(true),
    MinTlsVersion:     pulumi.String("TLS1_2"),
    AllowNestedItemsToBePublic: pulumi.Bool(false),
    NetworkRules: &storage.StorageAccountNetworkRulesTypeArgs{
        DefaultAction: pulumi.String("Deny"),
        Bypass:        pulumi.StringArray{pulumi.String("AzureServices")},
    },
})
```

Key advantages over ARM/Bicep:

- **Cross-cloud consistency**: Same tooling patterns for AWS, GCP, Azure
- **Rich state management**: Terraform and Pulumi track infrastructure state, enabling drift detection
- **Ecosystem**: Large communities, extensive module libraries
- **Programming languages**: Pulumi enables real programming constructs (loops, conditionals, functions)

**Verdict**: Terraform and Pulumi are the enterprise standard for multi-cloud organizations. Choose based on preference for HCL (Terraform) vs. general-purpose languages (Pulumi).

## The 80/20 of Azure Storage Account Configuration

After analyzing hundreds of production Storage Account deployments, certain patterns emerge. Here's what matters most:

### Always Configure

1. **Replication Type**: Choose based on durability requirements
   - LRS: Development, non-critical data
   - ZRS: Single-region high availability
   - GRS/GZRS: Production, disaster recovery required

2. **Network Restrictions**: Never leave storage open to the internet
   - Default action: Deny
   - Bypass Azure Services: Usually yes
   - IP rules for CI/CD and administrative access
   - VNet rules for application access

3. **Security Settings**:
   - HTTPS only: Always true
   - TLS 1.2 minimum: Always true
   - Disable public blob access: Usually true

4. **Data Protection**:
   - Soft delete for blobs: 7-365 days based on recovery needs
   - Soft delete for containers: Matching retention
   - Versioning: Enable for critical data

### Often Overlooked

1. **Access Tier Selection**: Hot vs Cool has significant cost implications
   - Hot: Frequently accessed data (higher storage cost, lower access cost)
   - Cool: Infrequently accessed (lower storage cost, higher access cost, 30-day minimum)

2. **Account Kind**: StorageV2 is almost always the right choice
   - BlockBlobStorage and FileStorage only for specific premium scenarios
   - Legacy Storage (v1) should be avoided

3. **Private Endpoints**: Required for true network isolation
   - Storage firewall still allows Azure backbone access
   - Private endpoints provide full private network integration

## Why Planton Chose These Defaults

Planton's Azure Storage Account implementation makes specific choices to balance security, usability, and operational simplicity:

### Security-First Defaults

- **HTTPS-only traffic**: Non-negotiable for production storage
- **TLS 1.2 minimum**: Older TLS versions have known vulnerabilities
- **Network rules default to Deny**: Secure by default, explicitly allow what's needed
- **Private container access**: No accidental public blob exposure

### Sensible Storage Defaults

- **StorageV2**: The modern, feature-complete account type
- **Standard tier**: Premium only when performance justifies cost
- **LRS replication**: Start simple, upgrade when needed
- **Hot access tier**: Most application data is accessed regularly

### Data Protection Defaults

- **7-day soft delete**: Balances recovery capability with storage costs
- **Versioning disabled by default**: Enable explicitly when needed
- **Container soft delete**: Matches blob retention for consistency

### Why Not More?

We explicitly don't default:

- **GRS/GZRS replication**: Doubles storage costs; not all workloads need it
- **Premium tier**: Significantly more expensive; Standard handles most workloads
- **Hierarchical namespace**: Requires specific account configuration, not reversible
- **Static website hosting**: Requires additional configuration post-deployment

## Integration Patterns

### With Azure Kubernetes Service (AKS)

Storage accounts commonly serve AKS workloads via:

1. **CSI Driver**: Azure Blob or File CSI driver for persistent volumes
2. **Workload Identity**: Pod identity accessing storage via managed identity
3. **Network Integration**: Storage VNet rules allowing AKS subnet

### With Azure Functions / App Service

1. **Managed Identity**: Applications access storage without credentials
2. **Key Vault integration**: Storage connection strings stored in Key Vault
3. **VNet Integration**: Private endpoint access from app VNet

### With Data Analytics

1. **Data Lake Gen2**: Enable hierarchical namespace for analytics workloads
2. **Synapse/Databricks**: VNet service endpoints for secure access
3. **Event Grid**: Blob events triggering data pipelines

## Common Pitfalls and Solutions

### 1. Name Conflicts

**Problem**: Storage account names must be globally unique, 3-24 characters, lowercase alphanumeric only.

**Solution**: Use deterministic naming with organization prefix and unique suffix:
```
{org}{env}{purpose}{random4}
# Example: mycomproddata1234
```

### 2. Network Rule Lockout

**Problem**: Enabling "Deny" default action without proper IP rules locks out administrators.

**Solution**: Always include:
- Azure Services bypass
- CI/CD system IPs
- VPN/bastion host IPs for emergency access

### 3. Replication Cost Surprises

**Problem**: GRS/GZRS doubles storage costs, surprising teams at billing time.

**Solution**: 
- Default to LRS for development
- Explicitly choose replication based on documented requirements
- Use Azure Cost Management alerts

### 4. Access Tier Cost Optimization

**Problem**: Storing infrequently accessed data in Hot tier wastes money; storing frequently accessed data in Cool tier incurs access charges.

**Solution**:
- Lifecycle management policies for automatic tier transitions
- Monitor access patterns before choosing tier
- Use Hot for applications, Cool for backups/archives

## Conclusion

Azure Storage Account is deceptively simple at first glance but requires careful configuration for production workloads. The key decisions — replication strategy, network security, access tiers, and data protection — significantly impact both security and costs.

Planton's implementation provides secure, sensible defaults while exposing the 20% of configuration options that matter for 80% of use cases. By abstracting the complexity of storage account deployment into a declarative manifest, teams can deploy consistent, secure storage across environments without deep Azure storage expertise.

For organizations requiring advanced features (hierarchical namespace, premium performance, specific compliance configurations), the manifest format provides clear extension points while maintaining the security-first philosophy that production storage demands.

## References

- [Azure Storage Account Overview](https://learn.microsoft.com/en-us/azure/storage/common/storage-account-overview)
- [Storage Redundancy Options](https://learn.microsoft.com/en-us/azure/storage/common/storage-redundancy)
- [Storage Access Tiers](https://learn.microsoft.com/en-us/azure/storage/blobs/access-tiers-overview)
- [Storage Network Security](https://learn.microsoft.com/en-us/azure/storage/common/storage-network-security)
- [Storage Security Best Practices](https://learn.microsoft.com/en-us/azure/storage/blobs/security-recommendations)
- [Terraform Azure Storage Provider](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/storage_account)
- [Pulumi Azure Storage](https://www.pulumi.com/registry/packages/azure-native/api-docs/storage/)
