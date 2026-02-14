# Standard Container Registry

This preset creates an Azure Container Registry with Standard SKU and admin user disabled. Standard tier provides 100 GB storage, enhanced throughput for image pulls, and webhook support -- sufficient for most team and production container image hosting needs. Authentication is handled via Azure AD (service principals, managed identities) rather than the admin user account, following security best practices.

## When to Use

- Hosting Docker container images for AKS clusters, Azure App Service, or Azure Container Apps
- CI/CD pipelines that build and push container images using service principal or managed identity authentication
- Teams needing a private container registry with more storage and throughput than Basic tier
- Single-region workloads where geo-replication is not required

## Key Configuration Choices

- **Standard SKU** (`sku: STANDARD`) -- 100 GB storage, higher throughput than Basic. Upgrade to Premium for geo-replication, private endpoints, and content trust
- **Admin user disabled** (`adminUserEnabled: false`) -- Recommended for production. Use Azure AD service principals or managed identities for authentication. Enable only for quick prototyping with `docker login`
- **No geo-replication** -- Standard SKU does not support geo-replication. Use the Premium preset for multi-region image distribution
- **No network restrictions** -- The registry is publicly accessible. Upgrade to Premium to add private endpoint access or IP-based firewall rules

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., "eastus", "westeurope") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<youruniquename>` | Globally unique registry name (5-50 chars, lowercase letters and numbers only, no hyphens) | Choose a name; becomes `{name}.azurecr.io` |

## Related Presets

- **02-premium-geo-replicated** -- Use instead for multi-region image distribution, private endpoints, or content trust
