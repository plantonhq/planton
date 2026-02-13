# Azure Subnet

Deploys an Azure Subnet within an existing Virtual Network, with configurable address prefix, service endpoints, service delegation, and private endpoint network policies. Subnets partition a VNet's address space into segments for different workloads, tiers, or service delegations.

## What Gets Created

When you deploy an AzureSubnet resource, OpenMCF provisions:

- **Subnet** — a `network.Subnet` resource inside the specified Virtual Network, configured with the given address prefix, service endpoints, delegation, and network policies
- **Service Endpoints** — optimized routes over the Azure backbone to specified Azure services, bypassing the public internet (when `serviceEndpoints` is provided)
- **Service Delegation** — grants an Azure PaaS service permission to inject service-specific resources and network rules into the subnet (when `delegation` is provided)
- **Azure Tags** — resource metadata tags applied to the subnet for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the parent VNet exists (can reference an AzureResourceGroup resource)
- **An Azure Virtual Network** with an address space that contains the desired subnet CIDR block (can reference an AzureVpc resource)
- **Network planning** — the subnet address prefix must be a subset of the parent VNet's address space and must not overlap with other subnets in the same VNet. Azure reserves 5 IPs per subnet (first 4 + last) for internal use.

## Quick Start

Create a file `subnet.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureSubnet
metadata:
  name: my-subnet
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureSubnet.my-subnet
spec:
  resourceGroup: my-rg
  vnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet
  name: my-subnet
  addressPrefix: "10.0.1.0/24"
```

Deploy:

```shell
openmcf apply -f subnet.yaml
```

This creates a /24 subnet (254 usable IPs) with private endpoint network policies disabled, private link service network policies enabled, and no service endpoints or delegations.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group where the parent VNet exists. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `vnetId` | `StringValueOrRef` | Azure Resource Manager ID of the parent Virtual Network. The subnet is created inside this VNet and must use an address prefix within the VNet's address space. Can reference an AzureVpc resource via `valueFrom`. | Required |
| `name` | `string` | Name of the subnet. Must be unique within the VNet. Allowed characters: alphanumeric, underscores, hyphens, periods. Must start with alphanumeric. | Required, 1–80 characters |
| `addressPrefix` | `string` | IPv4 CIDR block for the subnet (e.g., `10.0.1.0/24`). Must be a subset of the parent VNet's address space and must not overlap with other subnets. | Required, minimum length 1 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `serviceEndpoints` | `string[]` | `[]` | Azure service endpoints to enable. Creates optimized routes over the Azure backbone. Common values: `Microsoft.Storage`, `Microsoft.Sql`, `Microsoft.KeyVault`, `Microsoft.AzureCosmosDB`, `Microsoft.ServiceBus`, `Microsoft.EventHub`, `Microsoft.Web`, `Microsoft.ContainerRegistry`. |
| `delegation` | `object` | none | Service delegation granting an Azure PaaS service permission to inject resources into the subnet. A subnet can have at most one delegation. See delegation fields below. |
| `delegation.name` | `string` | — | A user-chosen label for the delegation (e.g., `postgresql`, `container-apps`). Required when `delegation` is set. |
| `delegation.serviceName` | `string` | — | The Azure service to delegate to. Required when `delegation` is set. Common values: `Microsoft.DBforPostgreSQL/flexibleServers`, `Microsoft.DBforMySQL/flexibleServers`, `Microsoft.App/environments`, `Microsoft.Web/serverFarms`, `Microsoft.ContainerInstance/containerGroups`, `Microsoft.Netapp/volumes`. |
| `delegation.actions` | `string[]` | `[]` | Actions the delegated service is permitted to perform. If omitted, Azure uses the default actions. Common action: `Microsoft.Network/virtualNetworks/subnets/action`. |
| `privateEndpointNetworkPolicies` | `string` | `Disabled` | Controls whether network policies apply to private endpoints. Values: `Disabled` (no policies on private endpoints), `Enabled` (both NSG and route table), `NetworkSecurityGroupEnabled` (NSG only), `RouteTableEnabled` (route table only). |
| `privateLinkServiceNetworkPoliciesEnabled` | `bool` | `true` | Controls whether network policies apply to Private Link Service resources. Set to `false` only when creating a Private Link Service that needs to bypass network policies. |

## Examples

### General-Purpose Workload Subnet

A /24 subnet for general workloads with no special endpoints or delegations:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureSubnet
metadata:
  name: workload-subnet
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureSubnet.workload-subnet
spec:
  resourceGroup: dev-rg
  vnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/dev-rg/providers/Microsoft.Network/virtualNetworks/dev-vnet
  name: workload-subnet
  addressPrefix: "10.0.1.0/24"
```

### Database Subnet with Delegation and Service Endpoints

A subnet delegated to PostgreSQL Flexible Server with service endpoints for secure access to Storage and Key Vault:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureSubnet
metadata:
  name: postgres-subnet
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureSubnet.postgres-subnet
spec:
  resourceGroup: prod-rg
  vnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet
  name: postgres-subnet
  addressPrefix: "10.0.10.0/24"
  serviceEndpoints:
    - Microsoft.Storage
    - Microsoft.KeyVault
  delegation:
    name: postgresql
    serviceName: Microsoft.DBforPostgreSQL/flexibleServers
    actions:
      - Microsoft.Network/virtualNetworks/subnets/action
```

### Private Endpoint Subnet with Network Policies

A subnet for private endpoints with NSG policies enabled for zero-trust architectures:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureSubnet
metadata:
  name: pe-subnet
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureSubnet.pe-subnet
spec:
  resourceGroup: prod-rg
  vnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet
  name: pe-subnet
  addressPrefix: "10.0.20.0/28"
  privateEndpointNetworkPolicies: NetworkSecurityGroupEnabled
  privateLinkServiceNetworkPoliciesEnabled: true
```

### Using Foreign Key References

Reference OpenMCF-managed resources instead of hardcoding resource group name and VNet ID:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureSubnet
metadata:
  name: app-subnet
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureSubnet.app-subnet
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  vnetId:
    valueFrom:
      kind: AzureVpc
      name: my-vnet
      field: status.outputs.vnet_id
  name: app-subnet
  addressPrefix: "10.0.2.0/24"
  serviceEndpoints:
    - Microsoft.Sql
    - Microsoft.Storage
    - Microsoft.KeyVault
    - Microsoft.Web
```

### Container App Environment Subnet

A subnet delegated to Azure Container App Environments with the minimum /23 sizing recommended by Azure:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureSubnet
metadata:
  name: cae-subnet
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureSubnet.cae-subnet
spec:
  resourceGroup: prod-rg
  vnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet
  name: cae-subnet
  addressPrefix: "10.0.32.0/23"
  delegation:
    name: container-apps
    serviceName: Microsoft.App/environments
    actions:
      - Microsoft.Network/virtualNetworks/subnets/action
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `subnet_id` | `string` | Azure Resource Manager ID of the subnet. This is the most referenced Azure output in OpenMCF, consumed by AzureAksCluster, AzureContainerAppEnvironment, AzurePostgresqlFlexibleServer, AzureMysqlFlexibleServer, AzureRedisCache, AzurePrivateEndpoint, AzureApplicationGateway, AzureLoadBalancer, AzureVirtualMachine, AzureFunctionApp, and AzureLinuxWebApp. |
| `subnet_name` | `string` | Name of the subnet within the VNet |
| `address_prefix` | `string` | IPv4 CIDR block assigned to this subnet. Useful in NSG rules, firewall rules, and network planning where downstream resources need to know the subnet's address range. |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) — provides the resource group where the parent VNet exists
- [AzureVpc](/docs/catalog/azure/azurevpc) — provides the parent Virtual Network that contains this subnet
- [AzureAksCluster](/docs/catalog/azure/azureakscluster) — references `subnet_id` for node pool placement
- [AzureContainerAppEnvironment](/docs/catalog/azure/azurecontainerappenvironment) — requires a delegated subnet for VNet integration
- [AzurePostgresqlFlexibleServer](/docs/catalog/azure/azurepostgresqlflexibleserver) — requires a delegated subnet for VNet integration
- [AzureMysqlFlexibleServer](/docs/catalog/azure/azuremysqlflexibleserver) — requires a delegated subnet for VNet integration
- [AzurePrivateEndpoint](/docs/catalog/azure/azureprivateendpoint) — deployed into a subnet for private connectivity to Azure PaaS services
- [AzureApplicationGateway](/docs/catalog/azure/azureapplicationgateway) — requires a dedicated subnet (minimum /27)
- [AzureKeyVault](/docs/catalog/azure/azurekeyvault) — can restrict access to specific subnet IDs via network ACLs
