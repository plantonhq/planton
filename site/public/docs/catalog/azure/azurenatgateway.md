---
title: "NAT Gateway"
description: "NAT Gateway deployment documentation"
icon: "package"
order: 100
componentName: "azurenatgateway"
---

# Azure NAT Gateway

Deploys an Azure NAT Gateway with an automatically provisioned public IP address or public IP prefix, associated with a specified subnet and resource group. The component handles the full lifecycle including IP allocation, gateway creation, IP-to-gateway association, and subnet-to-gateway association.

## What Gets Created

When you deploy an AzureNatGateway resource, OpenMCF provisions:

- **NAT Gateway** — a Standard SKU `network.NatGateway` in the specified region and resource group, configured with an idle timeout for TCP connections
- **Public IP** — a single Standard SKU static `network.PublicIp` allocated and associated with the gateway (when no prefix length is specified)
- **Public IP Prefix** — a Standard SKU `network.PublicIpPrefix` of the requested prefix length, associated with the gateway instead of a single IP (when `publicIpPrefixLength` is set)
- **Subnet Association** — a `network.SubnetNatGatewayAssociation` binding the NAT Gateway to the target subnet so all outbound traffic from that subnet routes through the gateway
- **Azure Tags** — resource metadata tags applied to the gateway, public IP, and prefix for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the NAT Gateway will be created (can reference an AzureResourceGroup resource)
- **A subnet** to attach the NAT Gateway to (can reference an AzureVpc resource's nodes subnet output)

## Quick Start

Create a file `natgateway.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNatGateway
metadata:
  name: my-natgw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureNatGateway.my-natgw
spec:
  region: eastus
  resourceGroup: my-rg
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/nodes
```

Deploy:

```shell
openmcf apply -f natgateway.yaml
```

This creates a Standard SKU NAT Gateway with a single static public IP, a 4-minute idle timeout, and associates it with the specified subnet.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the NAT Gateway (e.g., `eastus`, `westeurope`). | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `subnetId` | `StringValueOrRef` | Subnet resource ID to attach the NAT Gateway to. Can reference an AzureVpc resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `idleTimeoutMinutes` | `int32` | `4` | Idle timeout in minutes for TCP connections through the NAT Gateway. Range: 4--120. |
| `publicIpPrefixLength` | `int32` | unset | CIDR prefix length for a Public IP Prefix. When set (allowed values 28--31), a Public IP Prefix of the given length is created instead of a single public IP. Use this when workloads require multiple outbound IPs. |
| `tags` | `map<string, string>` | `{}` | Additional tags to assign to the NAT Gateway and associated IP resources. Merged with automatic metadata tags. |

## Examples

### Single Public IP with Literal Subnet ID

A NAT Gateway with default settings and a directly specified subnet:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNatGateway
metadata:
  name: basic-natgw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureNatGateway.basic-natgw
spec:
  region: eastus
  resourceGroup: dev-rg
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/dev-rg/providers/Microsoft.Network/virtualNetworks/dev-vnet/subnets/nodes
```

### Custom Idle Timeout with Tags

A NAT Gateway with a longer idle timeout for workloads that hold long-lived outbound TCP connections:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNatGateway
metadata:
  name: long-lived-natgw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureNatGateway.long-lived-natgw
spec:
  region: westeurope
  resourceGroup: prod-rg
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/app
  idleTimeoutMinutes: 30
  tags:
    team: platform
    cost-center: infra
```

### Public IP Prefix for Scale

A NAT Gateway backed by a /28 Public IP Prefix (16 addresses) for high-throughput subnets that need multiple outbound IPs to avoid SNAT port exhaustion:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNatGateway
metadata:
  name: scale-natgw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureNatGateway.scale-natgw
spec:
  region: eastus
  resourceGroup: prod-rg
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/nodes
  publicIpPrefixLength: 28
  idleTimeoutMinutes: 10
  tags:
    workload: high-throughput
```

### Using Foreign Key References

Reference OpenMCF-managed resources for the resource group and subnet instead of hardcoding IDs:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNatGateway
metadata:
  name: ref-natgw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureNatGateway.ref-natgw
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  subnetId:
    valueFrom:
      kind: AzureVpc
      name: my-vpc
      field: status.outputs.nodes_subnet_id
  idleTimeoutMinutes: 15
```

### Production AKS Cluster Egress

A NAT Gateway designed to serve as the outbound egress for an AKS cluster node subnet, with a /31 prefix (2 addresses) and a 120-minute idle timeout for long-running batch jobs:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNatGateway
metadata:
  name: aks-egress-natgw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureNatGateway.aks-egress-natgw
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: aks-rg
      field: status.outputs.resource_group_name
  subnetId:
    valueFrom:
      kind: AzureVpc
      name: aks-vpc
      field: status.outputs.nodes_subnet_id
  publicIpPrefixLength: 31
  idleTimeoutMinutes: 120
  tags:
    purpose: aks-egress
    environment: production
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `natGatewayId` | `string` | Azure Resource Manager ID of the NAT Gateway |
| `publicIpAddresses` | `string[]` | Public IP addresses allocated to the NAT Gateway. Populated when a single public IP is created; empty when a public IP prefix is used. |
| `publicIpPrefixId` | `string` | Azure Resource Manager ID of the Public IP Prefix. Populated when `publicIpPrefixLength` is set; empty otherwise. |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) -- provides the resource group for gateway placement
- [AzureVpc](/docs/catalog/azure/azurevpc) -- provides the VNet and subnets that the NAT Gateway attaches to
- [AzurePublicIp](/docs/catalog/azure/azurepublicip) -- standalone public IP resources, if managing IPs separately from the gateway
- [AzureAksCluster](/docs/catalog/azure/azureakscluster) -- AKS clusters commonly use a NAT Gateway for predictable outbound IPs and SNAT port scaling
