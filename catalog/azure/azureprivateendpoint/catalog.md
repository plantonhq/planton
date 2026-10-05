# Azure Private Endpoint

Deploys an Azure Private Endpoint: a network interface that gives a Private Link-powered service — an Azure PaaS resource (SQL, PostgreSQL, Storage, Key Vault, Cosmos DB, ...) or a partner's Private Link Service — a private IP inside your virtual network. Traffic to the service rides the Microsoft backbone, never the public internet; each endpoint maps to ONE sub-resource of the target (the data-exfiltration boundary), and the private DNS zone group is part of this resource so DNS registration stays atomic with the endpoint.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Private Endpoint** -- a network interface in the specified subnet with a private IP (dynamic from the subnet, or pinned via `ipConfigurations`), connected to the target through its private service connection
- **Private Service Connection** -- the connection block from the spec: either a Private Link-enabled Azure resource plus its sub-resource (group ID), or a partner's Private Link Service alias; auto-approved by default, or a manual request carrying your message to the target's owner
- **Private DNS Zone Group** -- created when `privateDnsZoneIds` lists zones; registers the endpoint's IP as an A-record in each referenced Private DNS zone so the service FQDN resolves to the private IP within linked networks
- **ASG Associations** -- when `applicationSecurityGroupIds` lists groups, the endpoint's network interface joins them so NSG rules can govern its traffic by workload group
- **Azure Tags** -- your governance tags merged over the Planton-derived resource tags (organization, environment, resource id); a user tag with the same key wins

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the Private Endpoint will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **A subnet** in the same region as the private endpoint, permitting private endpoints (its private-endpoint network policies configured accordingly). Reference an AzureSubnet Infra Component or provide the ARM ID.
- **A target**: a Private Link-enabled Azure resource (reference its Infra Component or provide its ARM ID) with the correct sub-resource name for the facet you need -- or a partner's Private Link Service ALIAS (always ends in `.azure.privatelinkservice`) for cross-tenant connections, typically with manual approval.
- **A Private DNS Zone** (strongly recommended) named for the target service's privatelink domain (e.g., `privatelink.postgres.database.azure.com`), already linked to the client networks. Without registration, the service FQDN resolves to its PUBLIC IP inside the VNet and clients silently bypass the endpoint.

## Deploy

### Console

Open the deployment store, find **Azure Private Endpoint**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the endpoint's five steps: placement (with a live VNet-to-subnet cascade from your connected subscription), the private link target (resource or alias, with a curated sub-resource picker and a live connection diagram), DNS integration (the wizard names the exact privatelink zone your chosen sub-resource needs), network details, and governance tags. Start from the **Private Endpoint for Azure SQL Database** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePrivateEndpoint
metadata:
  name: db-endpoint
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-network-rg"
  name: pg-orders-pe
  subnetId:
    valueFrom:
      kind: AzureSubnet
      name: endpoints-subnet
      fieldPath: status.outputs.subnet_id
  privateServiceConnection:
    privateConnectionResourceId:
      valueFrom:
        kind: AzurePostgresqlFlexibleServer
        name: orders-db
        fieldPath: status.outputs.server_id
    subresourceNames:
      - postgresqlServer
  privateDnsZoneIds:
    - valueFrom:
        kind: AzurePrivateDnsZone
        name: pg-privatelink-zone
        fieldPath: status.outputs.zone_id
```

```shell
planton apply -f private-endpoint.yaml
```

This creates a Private Endpoint connected to a PostgreSQL Flexible Server with a dynamically allocated private IP, registered into the PostgreSQL privatelink zone so the server's FQDN resolves privately. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Private Endpoint to a subnet, target service, and Private DNS zone deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  subnetId:
    valueFrom:
      kind: AzureSubnet
      name: endpoints-subnet
      fieldPath: status.outputs.subnet_id
  privateServiceConnection:
    privateConnectionResourceId:
      valueFrom:
        kind: AzurePostgresqlFlexibleServer
        name: orders-db
        fieldPath: status.outputs.server_id
    subresourceNames:
      - postgresqlServer
  privateDnsZoneIds:
    - valueFrom:
        kind: AzurePrivateDnsZone
        name: pg-privatelink-zone
        fieldPath: status.outputs.zone_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group, subnet, target service, and DNS zone first, then provisions the Private Endpoint with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Private Endpoint. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Target and sub-resource** (`privateServiceConnection`) -- Exactly one of `privateConnectionResourceId` (an Azure resource, the common case) or `connectionAlias` (a partner's Private Link Service). With a resource target, `subresourceNames` names the facet this endpoint reaches -- `postgresqlServer` for PostgreSQL Flexible Server, `sqlServer` for Azure SQL, `vault` for Key Vault, `blob`/`file`/`queue`/`table` for Storage, `Sql`/`MongoDB` for Cosmos DB, `redisCache` for Redis, `registry` for Container Registry. Most endpoints name exactly one; deploy one endpoint per facet. With an alias, the alias pins the target and no sub-resource is named. The entire connection is fixed at creation.

**Manual approval** (`isManualConnection` + `requestMessage`) -- Auto-approval (the default) completes immediately whenever your credentials have access to the target. Set manual for cross-tenant or cross-subscription targets: the connection sits pending until the owner approves, and your `requestMessage` (1-140 characters, required exactly when manual) is what they see.

**DNS zone group** (`privateDnsZoneIds`) -- Reference the AzurePrivateDnsZone whose name matches the target's privatelink domain; the endpoint's IP registers there automatically. Leave empty ONLY when DNS is managed outside Azure -- without registration the FQDN resolves to the public IP and the private link is silently defeated. Updatable in place.

**Subnet and IPs** -- The subnet must share the endpoint's region and permit private endpoints; every endpoint consumes one IP from it (a dedicated "endpoints" subnet is the common pattern). Dynamic allocation is right for nearly every endpoint; pin `ipConfigurations` (one entry per sub-resource) only when something outside DNS depends on the address, such as on-premises firewall allowlists.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureSubnet** | `subnetId` | `status.outputs.subnet_id` |
| **The target service** (kind varies) | `privateServiceConnection.privateConnectionResourceId` | the service's own ID output (e.g. `status.outputs.server_id`) |
| **AzurePrivateDnsZone** (recommended) | `privateDnsZoneIds` | `status.outputs.zone_id` |
| **AzureApplicationSecurityGroup** (optional) | `applicationSecurityGroupIds` | `status.outputs.application_security_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `private_ip_address` | The private IP allocated from the subnet -- what the service FQDN resolves to inside linked networks | Network troubleshooting, firewall rules, externally-managed DNS |
| `network_interface_id` | ARM ID of the network interface Azure created for the endpoint | Effective-routes and NSG flow-log diagnostics |

`status.outputs` also carries `private_endpoint_id` and `private_endpoint_name`, but no catalog kind consumes the endpoint by reference — clients reach the target through DNS, not through the endpoint's own ID.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Database private connectivity** -- An endpoint targeting Azure SQL (`sqlServer`) with DNS registration in `privatelink.database.windows.net`, so the server's FQDN resolves privately and its public endpoint can be disabled. Start from the **Private Endpoint for Azure SQL Database** preset.

**Storage account private access** -- An endpoint targeting Blob Storage (`blob`) registered in `privatelink.blob.core.windows.net`. Storage exposes one sub-resource per service -- an application needing blob AND file access deploys two endpoints. Start from the **Private Endpoint for Azure Blob Storage** preset.

**Key Vault private access** -- An endpoint targeting Key Vault (`vault`) registered in `privatelink.vaultcore.azure.net`, for zero-trust architectures where secret access must stay on the backbone. Start from the **Key Vault Private Endpoint** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the Private Endpoint is created
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- provides the subnet lending the private IP
- [**Azure Private DNS Zone**](/infra-catalog/azure-private-dns-zone) -- receives the endpoint's A-record so the service FQDN resolves privately
- [**Azure Application Security Group**](/infra-catalog/azure-application-security-group) -- lets NSG rules govern endpoint traffic by workload group
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) / [**Azure Event Hub Namespace**](/infra-catalog/azure-event-hub-namespace) / [**Azure Key Vault**](/infra-catalog/azure-key-vault) -- typical locked-down targets whose public access is disabled once the endpoint carries the traffic
