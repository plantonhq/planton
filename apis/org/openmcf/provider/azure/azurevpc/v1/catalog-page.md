# Azure VPC (Virtual Network)

Deploys an Azure Virtual Network with a configurable address space, a dedicated AKS nodes subnet, optional NAT Gateway for outbound internet access, and Private DNS zone links for name resolution. This component serves as the networking foundation for AKS clusters and other Azure workloads that require isolated VNet connectivity.

## What Gets Created

When you deploy an AzureVpc resource, OpenMCF provisions:

- **Virtual Network** — a `network.VirtualNetwork` resource in the specified region and resource group with the configured address space CIDR
- **Nodes Subnet** — a `network.Subnet` carved from the VNet address space for AKS cluster nodes
- **NAT Gateway** (optional) — a `network.NatGateway` with a Standard SKU, a static `network.PublicIp`, a `network.NatGatewayPublicIpAssociation`, and a `network.SubnetNatGatewayAssociation` linking the gateway to the nodes subnet; created only when `isNatGatewayEnabled` is `true`
- **Private DNS Zone Links** (optional) — a `privatedns.ZoneVirtualNetworkLink` for each entry in `dnsPrivateZoneLinks`, enabling private DNS resolution within the VNet
- **Azure Tags** — resource metadata tags applied to the VNet, NAT Gateway, Public IP, and DNS zone links for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the VNet and related resources will be created (can reference an AzureResourceGroup resource)
- **Network planning** — determine the address space CIDR and nodes subnet CIDR before deployment; the nodes subnet must be a subset of the address space

## Quick Start

Create a file `vpc.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureVpc
metadata:
  name: my-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureVpc.my-vpc
spec:
  region: eastus
  resourceGroup: my-rg
  addressSpaceCidr: "10.0.0.0/16"
  nodesSubnetCidr: "10.0.0.0/18"
```

Deploy:

```shell
openmcf apply -f vpc.yaml
```

This creates a Virtual Network with a `/16` address space and a `/18` nodes subnet in the specified resource group. No NAT Gateway or DNS zone links are created by default.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the Virtual Network (e.g., `eastus`, `westeurope`). | Required |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `addressSpaceCidr` | `string` | CIDR block defining the Virtual Network address space (e.g., `10.0.0.0/16`). | Required |
| `nodesSubnetCidr` | `string` | CIDR block for the AKS nodes subnet. Must be a subset of `addressSpaceCidr` (e.g., `10.0.0.0/18`). | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `isNatGatewayEnabled` | `bool` | `false` | Creates a NAT Gateway with a static public IP and associates it with the nodes subnet for outbound internet connectivity. |
| `dnsPrivateZoneLinks` | `string[]` | `[]` | List of Azure Private DNS zone names to link to this Virtual Network for private name resolution. |
| `tags` | `map<string, string>` | `{}` | Arbitrary key-value tags applied to the Virtual Network and related resources. Merged with auto-generated metadata tags. |

## Examples

### Basic VNet for Development

A minimal Virtual Network for non-production workloads with no NAT Gateway or DNS links:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureVpc
metadata:
  name: dev-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureVpc.dev-vpc
spec:
  region: eastus
  resourceGroup: dev-rg
  addressSpaceCidr: "10.10.0.0/16"
  nodesSubnetCidr: "10.10.0.0/18"
```

### VNet with NAT Gateway

A Virtual Network with a NAT Gateway for outbound internet access, suitable for AKS clusters that require a stable egress IP:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureVpc
metadata:
  name: staging-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AzureVpc.staging-vpc
spec:
  region: westeurope
  resourceGroup: staging-rg
  addressSpaceCidr: "10.20.0.0/16"
  nodesSubnetCidr: "10.20.0.0/18"
  isNatGatewayEnabled: true
  tags:
    environment: staging
    team: platform
```

### Full-Featured VNet with DNS Links and Foreign Key Reference

A production Virtual Network with NAT Gateway, Private DNS zone links for internal service resolution, custom tags, and a foreign key reference to an OpenMCF-managed resource group:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureVpc
metadata:
  name: prod-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureVpc.prod-vpc
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: prod-rg
      field: status.outputs.resource_group_name
  addressSpaceCidr: "10.0.0.0/16"
  nodesSubnetCidr: "10.0.0.0/18"
  isNatGatewayEnabled: true
  dnsPrivateZoneLinks:
    - privatelink.database.windows.net
    - privatelink.blob.core.windows.net
    - privatelink.vaultcore.azure.net
  tags:
    environment: production
    team: infrastructure
    cost-center: cc-1234
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `vnetId` | `string` | Azure Resource Manager ID of the Virtual Network |
| `nodesSubnetId` | `string` | Azure Resource Manager ID of the AKS nodes subnet within the Virtual Network |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) — provides the resource group for VNet placement
- [AzureAksCluster](/docs/catalog/azure/azureakscluster) — AKS clusters deployed into the nodes subnet of this VNet
- [AzureKeyVault](/docs/catalog/azure/azurekeyvault) — Key Vaults can restrict network access to subnets in this VNet
