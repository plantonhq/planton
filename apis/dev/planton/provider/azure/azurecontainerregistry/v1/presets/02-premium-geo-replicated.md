# Premium Container Registry with Geo-Replication

This preset creates an Azure Container Registry with Premium SKU and geo-replication to a secondary region. Premium tier provides 500 GB storage, geo-replication for multi-region image distribution, private endpoint support, content trust (image signing), and IP-based firewall rules. Geo-replicated images are pulled from the nearest replica, providing low-latency pulls for globally distributed AKS clusters and services.

## When to Use

- Multi-region deployments where AKS clusters or services in different regions need fast image pulls
- Enterprise environments requiring private endpoint access to the container registry
- Production workloads needing content trust (Docker Content Trust / Notary) for image signing
- High-availability requirements where the registry must survive a regional outage

## Key Configuration Choices

- **Premium SKU** (`sku: PREMIUM`) -- Required for geo-replication, private endpoints, content trust, and firewall rules. 500 GB storage included
- **Geo-replication** (`geoReplicationRegions`) -- Images are automatically replicated to the listed regions. Pulls are served from the nearest replica. Add multiple regions as needed
- **Admin user disabled** (`adminUserEnabled: false`) -- Use Azure AD service principals or managed identities. Enable only for quick prototyping
- **No private endpoint** -- The registry is publicly accessible by default. Create an `AzurePrivateEndpoint` targeting the registry for private-only access
- **No firewall rules** -- All public IPs can pull/push. Add IP-based firewall rules post-deployment for network restrictions

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Primary Azure region (e.g., "eastus") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<youruniquename>` | Globally unique registry name (5-50 chars, lowercase letters and numbers only, no hyphens) | Choose a name; becomes `{name}.azurecr.io` |
| `<secondary-azure-region>` | Secondary region for geo-replication (e.g., "westeurope") | Regions where you run workloads that pull images |

## Related Presets

- **01-standard** -- Use instead for single-region workloads where geo-replication and private endpoints are not needed
