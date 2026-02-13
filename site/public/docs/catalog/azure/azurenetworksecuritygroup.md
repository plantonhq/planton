---
title: "Network Security  Group"
description: "Network Security  Group deployment documentation"
icon: "package"
order: 100
componentName: "azurenetworksecuritygroup"
---

# Azure Network Security Group

Deploys an Azure Network Security Group (NSG) with priority-ordered security rules that control inbound and outbound traffic for Azure resources. The component bundles the NSG with its security rules because an NSG without rules relies entirely on Azure defaults, making the rules the substance of the resource.

## What Gets Created

When you deploy an AzureNetworkSecurityGroup resource, OpenMCF provisions:

- **Network Security Group** — a `network.NetworkSecurityGroup` resource in the specified region and resource group, acting as a stateful firewall for associated subnets or NICs
- **Security Rules** — a separate `network.NetworkSecurityRule` resource for each entry in `securityRules`, providing per-rule lifecycle management and explicit state tracking
- **Azure Tags** — resource metadata tags applied to the NSG for tracking and governance

The component does not create subnet-to-NSG associations. Association is handled separately via `azurerm_subnet_network_security_group_association`, keeping the NSG lifecycle independent of any particular subnet or NIC.

Azure automatically creates implicit default rules in every NSG (priorities 65000-65500) that allow VNet-to-VNet traffic, allow Azure Load Balancer probes, and deny all other inbound traffic. User-defined rules (priorities 100-4096) are evaluated before these defaults.

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the NSG will be created (can reference an AzureResourceGroup resource)
- **Network planning** — understand the traffic flows to allow or deny before defining security rules

## Quick Start

Create a file `nsg.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNetworkSecurityGroup
metadata:
  name: my-nsg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureNetworkSecurityGroup.my-nsg
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-nsg
  securityRules:
    - name: allow-https-inbound
      priority: 100
      direction: Inbound
      access: Allow
      protocol: Tcp
      destinationPortRange: "443"
```

Deploy:

```shell
openmcf apply -f nsg.yaml
```

This creates an NSG with a single rule allowing inbound HTTPS traffic from any source. All other inbound traffic is handled by Azure's implicit default rules (VNet-to-VNet allowed, everything else denied).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the NSG (e.g., `eastus`, `westeurope`). Must match the region of resources it will be associated with. | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `name` | `string` | Name of the Network Security Group. Must be unique within the resource group. Allowed characters: alphanumeric, underscores, hyphens, periods. Must start with alphanumeric. | Required, 1-80 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `securityRules` | `AzureSecurityRule[]` | `[]` | Security rules defining allowed or denied traffic flows. Rules are evaluated in priority order (lowest number first). An NSG with no rules relies on Azure defaults: allow VNet-to-VNet, allow load balancer probes, deny all other inbound, allow all outbound. |

### Security Rule Fields

Each entry in `securityRules` supports the following fields:

| Field | Type | Default | Required | Description |
|-------|------|---------|----------|-------------|
| `name` | `string` | — | Yes | Unique name within the NSG. Use descriptive names like `allow-https-inbound`. 1-80 characters. |
| `description` | `string` | — | No | Human-readable description of the rule's purpose. Maximum 140 characters. |
| `priority` | `int` | — | Yes | Evaluation priority. Lower numbers are evaluated first. Range: 100-4096. Use increments of 10 or 100 to leave room for future rules. |
| `direction` | `string` | — | Yes | Traffic direction. Values: `Inbound`, `Outbound`. |
| `access` | `string` | — | Yes | Access decision when the rule matches. Values: `Allow`, `Deny`. |
| `protocol` | `string` | — | Yes | Network protocol. Values: `Tcp`, `Udp`, `Icmp`, `*` (any). |
| `sourcePortRange` | `string` | `*` | No | Source port, range (`1024-65535`), or `*` for any. Most rules use `*` since source ports are typically ephemeral. |
| `destinationPortRange` | `string` | — | Yes | Destination port, range (`1024-65535`), or `*` for any. Examples: `22` (SSH), `80` (HTTP), `443` (HTTPS). |
| `sourceAddressPrefix` | `string` | `*` | No | Source CIDR, IP, Azure service tag (`VirtualNetwork`, `Internet`), or `*`. Ignored if `sourceAddressPrefixes` is set. |
| `destinationAddressPrefix` | `string` | `*` | No | Destination CIDR, IP, Azure service tag, or `*`. Ignored if `destinationAddressPrefixes` is set. |
| `sourceAddressPrefixes` | `string[]` | `[]` | No | Multiple source CIDRs or IPs. Takes precedence over `sourceAddressPrefix` when non-empty. Service tags are not supported in this field. |
| `destinationAddressPrefixes` | `string[]` | `[]` | No | Multiple destination CIDRs or IPs. Takes precedence over `destinationAddressPrefix` when non-empty. Service tags are not supported in this field. |

## Examples

### Allow HTTPS Only

A minimal NSG that allows inbound HTTPS and denies everything else (via Azure defaults):

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNetworkSecurityGroup
metadata:
  name: web-nsg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureNetworkSecurityGroup.web-nsg
spec:
  region: eastus
  resourceGroup: dev-rg
  name: web-nsg
  securityRules:
    - name: allow-https-inbound
      priority: 100
      direction: Inbound
      access: Allow
      protocol: Tcp
      destinationPortRange: "443"
```

### Web Tier with HTTP and HTTPS

An NSG for a web tier that allows both HTTP and HTTPS inbound from the internet:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNetworkSecurityGroup
metadata:
  name: web-tier-nsg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureNetworkSecurityGroup.web-tier-nsg
spec:
  region: eastus
  resourceGroup: prod-rg
  name: web-tier-nsg
  securityRules:
    - name: allow-https-inbound
      priority: 100
      direction: Inbound
      access: Allow
      protocol: Tcp
      destinationPortRange: "443"
      sourceAddressPrefix: Internet
    - name: allow-http-inbound
      priority: 200
      direction: Inbound
      access: Allow
      protocol: Tcp
      destinationPortRange: "80"
      sourceAddressPrefix: Internet
    - name: deny-all-inbound
      priority: 4096
      direction: Inbound
      access: Deny
      protocol: "*"
      destinationPortRange: "*"
      description: Explicit deny-all as a safety net
```

### Application Tier with Restricted Sources

An NSG for an application tier that only accepts traffic from the web tier subnet and allows SSH from a bastion host:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNetworkSecurityGroup
metadata:
  name: app-tier-nsg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureNetworkSecurityGroup.app-tier-nsg
spec:
  region: eastus
  resourceGroup: prod-rg
  name: app-tier-nsg
  securityRules:
    - name: allow-web-to-app
      priority: 100
      direction: Inbound
      access: Allow
      protocol: Tcp
      destinationPortRange: "8080"
      sourceAddressPrefix: "10.0.1.0/24"
      description: Allow traffic from web tier subnet
    - name: allow-ssh-from-bastion
      priority: 200
      direction: Inbound
      access: Allow
      protocol: Tcp
      destinationPortRange: "22"
      sourceAddressPrefix: "10.0.255.4"
      description: Allow SSH from bastion host
    - name: deny-all-inbound
      priority: 4096
      direction: Inbound
      access: Deny
      protocol: "*"
      destinationPortRange: "*"
```

### Data Tier with Multiple Source Ranges

An NSG for a data tier that allows database traffic from multiple application subnets using plural address prefixes:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNetworkSecurityGroup
metadata:
  name: data-tier-nsg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureNetworkSecurityGroup.data-tier-nsg
spec:
  region: westeurope
  resourceGroup: prod-rg
  name: data-tier-nsg
  securityRules:
    - name: allow-postgres-from-app-subnets
      priority: 100
      direction: Inbound
      access: Allow
      protocol: Tcp
      destinationPortRange: "5432"
      sourceAddressPrefixes:
        - "10.0.2.0/24"
        - "10.0.3.0/24"
        - "10.0.4.0/24"
      description: Allow PostgreSQL from all app subnets
    - name: deny-all-inbound
      priority: 4096
      direction: Inbound
      access: Deny
      protocol: "*"
      destinationPortRange: "*"
    - name: deny-internet-outbound
      priority: 4096
      direction: Outbound
      access: Deny
      protocol: "*"
      destinationPortRange: "*"
      destinationAddressPrefix: Internet
      description: Prevent data tier from reaching the internet
```

### Using Foreign Key References

Reference an OpenMCF-managed resource group instead of hardcoding the name:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureNetworkSecurityGroup
metadata:
  name: ref-nsg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureNetworkSecurityGroup.ref-nsg
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-nsg
  securityRules:
    - name: allow-https-inbound
      priority: 100
      direction: Inbound
      access: Allow
      protocol: Tcp
      destinationPortRange: "443"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `nsg_id` | `string` | Azure Resource Manager ID of the Network Security Group. Format: `/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/networkSecurityGroups/{name}`. Used by infra charts for subnet-NSG association. |
| `nsg_name` | `string` | Name of the Network Security Group |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) — provides the resource group for NSG placement
- [AzureVpc](/docs/catalog/azure/azurevpc) — provides the virtual network and subnets that NSGs are associated with
- [AzureSubnet](/docs/catalog/azure/azuresubnet) — NSGs are associated with subnets to filter traffic at the subnet level
- [AzureAksCluster](/docs/catalog/azure/azureakscluster) — AKS node pool subnets often require NSGs for controlling cluster traffic
