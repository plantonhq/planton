---
title: "Virtualmachine"
description: "Virtualmachine deployment documentation"
icon: "package"
order: 100
componentName: "azurevirtualmachine"
---

# Azure Virtual Machine

Deploys an Azure Virtual Machine with configurable size, OS image, network interface, authentication, and optional features including public IP, managed identities, boot diagnostics, spot pricing, and availability zone placement. The component creates the VM along with its network interface and, when enabled, a public IP address.

## What Gets Created

When you deploy an AzureVirtualMachine resource, OpenMCF provisions:

- **Network Interface** — a `network.NetworkInterface` attached to the specified subnet with configurable accelerated networking, private IP allocation, and optional NSG association
- **Public IP Address** — a `network.PublicIPAddress` with configurable SKU and allocation method, created only when `network.enablePublicIp` is `true`
- **Virtual Machine** — a `compute.VirtualMachine` in the specified region and resource group, configured with the chosen VM size, OS image, authentication method (SSH key or password), storage profile, and optional features such as boot diagnostics, managed identity, spot pricing, and availability zone placement
- **Azure Tags** — resource metadata tags applied to the VM for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the VM will be created (can reference an AzureResourceGroup resource)
- **A subnet** within an existing Virtual Network where the VM's network interface will be attached (can reference an AzureVpc resource)
- **Authentication material** — either an SSH public key (Linux) or an admin password (Windows or Linux with password auth)

## Quick Start

Create a file `vm.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureVirtualMachine
metadata:
  name: my-vm
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureVirtualMachine.my-vm
spec:
  region: eastus
  resourceGroup: my-rg
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/default
  image:
    publisher: Canonical
    offer: 0001-com-ubuntu-server-jammy
    sku: 22_04-lts-gen2
  sshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E... user@host"
```

Deploy:

```shell
openmcf apply -f vm.yaml
```

This creates a Standard_D2s_v3 Ubuntu 22.04 VM with SSH key authentication, the default admin username `azureuser`, a Premium SSD OS disk, boot diagnostics enabled, and accelerated networking on the network interface.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region where the VM will be deployed (e.g., `eastus`, `westus2`, `westeurope`). | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `subnetId` | `StringValueOrRef` | Azure resource ID of the subnet for the VM's network interface. Can reference an AzureVpc resource via `valueFrom`. | Required |
| `image` | `object` | Operating system image configuration. Must specify either marketplace image fields (`publisher`, `offer`, `sku`) or `customImageId`. | Required |
| `image.publisher` | `string` | Marketplace image publisher (e.g., `Canonical`, `MicrosoftWindowsServer`, `RedHat`). Required when using marketplace images. | — |
| `image.offer` | `string` | Marketplace image offer (e.g., `0001-com-ubuntu-server-jammy`, `WindowsServer`). Required when using marketplace images. | — |
| `image.sku` | `string` | Marketplace image SKU (e.g., `22_04-lts-gen2`, `2022-datacenter-g2`). Required when using marketplace images. | — |
| Authentication | — | Either `sshPublicKey` or `adminPassword` must be provided. | CEL validation enforced |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `vmSize` | `string` | `Standard_D2s_v3` | Azure VM size determining vCPU count, memory, and capabilities (e.g., `Standard_D4s_v5`). |
| `adminUsername` | `string` | `azureuser` | Admin username for the VM. Linux: SSH user. Windows: administrator name. Max 64 characters. |
| `sshPublicKey` | `string` | — | SSH public key for Linux VMs. Disables password authentication when set. Format: `ssh-rsa AAAAB3... user@host`. |
| `adminPassword` | `StringValueOrRef` | — | Admin password for Windows VMs or Linux VMs with password auth. Can reference an AzureKeyVault secret via `valueFrom`. |
| `image.version` | `string` | `latest` | Image version. Use `latest` for auto-updates or a specific version string for stability. |
| `image.customImageId` | `string` | — | Azure resource ID of a custom or shared image. When set, `publisher`/`offer`/`sku`/`version` are ignored. |
| `osDisk.sizeGb` | `int` | image default | OS disk size in GB. Range: 0–32767. |
| `osDisk.storageType` | `enum` | `premium_lrs` | Storage account type. Values: `standard_lrs` (HDD), `standard_ssd_lrs`, `premium_lrs` (SSD), `premium_zrs` (zone-redundant SSD). |
| `osDisk.caching` | `enum` | `read_write` | Caching mode. Values: `none`, `read_only`, `read_write`. |
| `osDisk.deleteWithVm` | `bool` | `true` | Whether the OS disk is deleted when the VM is deleted. |
| `osDisk.diskEncryptionSetId` | `StringValueOrRef` | — | Disk encryption set ID for customer-managed key encryption. Can reference an AzureKeyVault resource via `valueFrom`. |
| `dataDisks` | `object[]` | `[]` | Additional data disks. Each entry requires `name` (max 80 chars), `sizeGb` (1–32767), and `lun` (0–63). Optional: `storageType` (default `premium_lrs`), `caching` (default `read_only`), `deleteWithVm` (default `true`). |
| `network.enablePublicIp` | `bool` | `false` | Creates a public IP address for the VM. |
| `network.publicIpSku` | `enum` | `standard` | Public IP SKU. Values: `basic`, `standard`. Standard is required for availability zones. |
| `network.publicIpAllocation` | `enum` | `public_static` | Public IP allocation method. Values: `public_dynamic`, `public_static`. |
| `network.networkSecurityGroupId` | `StringValueOrRef` | — | NSG resource ID to associate with the network interface. |
| `network.enableAcceleratedNetworking` | `bool` | `true` | Enables accelerated networking for improved performance. Requires a compatible VM size. |
| `network.privateIpAllocation` | `enum` | `private_dynamic` | Private IP allocation method. Values: `private_dynamic`, `private_static`. |
| `network.privateIpAddress` | `string` | — | Static private IP address. Required when `privateIpAllocation` is `private_static`. Must be within the subnet's address range. |
| `availabilityZone` | `string` | — | Availability zone for zonal placement. Values: `1`, `2`, `3`, or empty for regional. |
| `enableBootDiagnostics` | `bool` | `true` | Enables boot diagnostics (serial console output and boot screenshots). |
| `enableSystemAssignedIdentity` | `bool` | `false` | Enables a system-assigned managed identity for authenticating to Azure services without credentials. |
| `userAssignedIdentityIds` | `string[]` | `[]` | Pre-created user-assigned managed identity resource IDs to attach to the VM. |
| `customData` | `string` | — | Cloud-init script (Linux) or PowerShell script (Windows) executed on first boot. Maximum 64 KB. |
| `tags` | `map<string, string>` | `{}` | Key-value pairs for Azure resource organization and cost tracking. |
| `isSpotInstance` | `bool` | `false` | Enables spot pricing. Spot VMs have significantly reduced cost but can be evicted. |
| `spotMaxPrice` | `double` | `0` | Maximum price per hour in USD for spot VMs. Set to `-1` for on-demand price cap. Only applicable when `isSpotInstance` is `true`. |

## Examples

### Linux VM with SSH Key Authentication

A basic Ubuntu VM with SSH access for development:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureVirtualMachine
metadata:
  name: dev-linux
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureVirtualMachine.dev-linux
spec:
  region: eastus
  resourceGroup: dev-rg
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/dev-rg/providers/Microsoft.Network/virtualNetworks/dev-vnet/subnets/default
  vmSize: Standard_D2s_v3
  image:
    publisher: Canonical
    offer: 0001-com-ubuntu-server-jammy
    sku: 22_04-lts-gen2
  sshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E... user@host"
  network:
    enablePublicIp: true
  tags:
    environment: dev
    team: platform
```

### Windows VM with Password Authentication

A Windows Server VM with password-based admin access and a data disk:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureVirtualMachine
metadata:
  name: win-server
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AzureVirtualMachine.win-server
spec:
  region: westus2
  resourceGroup: staging-rg
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/staging-rg/providers/Microsoft.Network/virtualNetworks/staging-vnet/subnets/app
  vmSize: Standard_D4s_v5
  adminUsername: winadmin
  adminPassword: "P@ssw0rd!Secure2026"
  image:
    publisher: MicrosoftWindowsServer
    offer: WindowsServer
    sku: 2022-datacenter-g2
  osDisk:
    sizeGb: 256
    storageType: premium_lrs
  dataDisks:
    - name: data-01
      sizeGb: 512
      lun: 0
      storageType: premium_lrs
      caching: read_only
  network:
    enablePublicIp: false
```

### Production VM with Managed Identity and Availability Zone

A production VM with system-assigned managed identity, zonal placement, and cloud-init:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureVirtualMachine
metadata:
  name: prod-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureVirtualMachine.prod-api
spec:
  region: eastus
  resourceGroup: prod-rg
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/app
  vmSize: Standard_D4s_v5
  image:
    publisher: Canonical
    offer: 0001-com-ubuntu-server-jammy
    sku: 22_04-lts-gen2
  sshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E... deploy@ci"
  availabilityZone: "1"
  enableSystemAssignedIdentity: true
  enableBootDiagnostics: true
  osDisk:
    storageType: premium_lrs
    caching: read_write
  network:
    enablePublicIp: false
    enableAcceleratedNetworking: true
    networkSecurityGroupId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/networkSecurityGroups/prod-nsg
  customData: |
    #!/bin/bash
    apt-get update && apt-get install -y docker.io
    systemctl enable docker
    systemctl start docker
  tags:
    environment: production
    service: api
```

### Spot Instance for Batch Workloads

A cost-optimized spot VM for fault-tolerant batch processing:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureVirtualMachine
metadata:
  name: batch-worker
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureVirtualMachine.batch-worker
spec:
  region: westeurope
  resourceGroup: batch-rg
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/batch-rg/providers/Microsoft.Network/virtualNetworks/batch-vnet/subnets/compute
  vmSize: Standard_D8s_v5
  image:
    publisher: Canonical
    offer: 0001-com-ubuntu-server-jammy
    sku: 22_04-lts-gen2
  sshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E... ops@ci"
  isSpotInstance: true
  spotMaxPrice: -1
  osDisk:
    storageType: standard_ssd_lrs
    deleteWithVm: true
  tags:
    workload: batch
```

### Using Foreign Key References

Reference OpenMCF-managed resources instead of hardcoding IDs:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureVirtualMachine
metadata:
  name: ref-vm
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureVirtualMachine.ref-vm
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
      name: my-vnet
      field: status.outputs.nodes_subnet_id
  image:
    publisher: Canonical
    offer: 0001-com-ubuntu-server-jammy
    sku: 22_04-lts-gen2
  sshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E... deploy@ci"
  adminPassword:
    valueFrom:
      kind: AzureKeyVault
      name: my-vault
      field: status.outputs.vault_uri
  enableSystemAssignedIdentity: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `vm_id` | `string` | Azure Resource Manager ID of the Virtual Machine |
| `vm_name` | `string` | Name of the Virtual Machine |
| `private_ip_address` | `string` | Private IP address assigned to the VM's primary network interface |
| `public_ip_address` | `string` | Public IP address assigned to the VM (only populated when `network.enablePublicIp` is `true`) |
| `public_ip_fqdn` | `string` | FQDN of the public IP (only populated when a DNS label is configured on the public IP) |
| `computer_name` | `string` | Hostname of the Virtual Machine |
| `system_assigned_identity_principal_id` | `string` | Principal ID of the system-assigned managed identity (only populated when `enableSystemAssignedIdentity` is `true`) |
| `network_interface_id` | `string` | Azure resource ID of the primary network interface |
| `availability_zone` | `string` | Availability zone where the VM is deployed (only populated for zonal deployments) |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) -- provides the resource group for VM placement
- [AzureVpc](/docs/catalog/azure/azurevpc) -- provides the VNet and subnet where the VM's network interface is attached
- [AzureNetworkSecurityGroup](/docs/catalog/azure/azurenetworksecuritygroup) -- controls inbound and outbound traffic rules for the VM's network interface
- [AzureKeyVault](/docs/catalog/azure/azurekeyvault) -- stores secrets such as admin passwords and disk encryption keys
- [AzurePublicIp](/docs/catalog/azure/azurepublicip) -- standalone public IP management (the VM component creates its own when `network.enablePublicIp` is set)
