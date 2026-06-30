# Azure Virtual Machine

Deploy Azure Virtual Machines using Planton's declarative approach with cross-references to other Azure resources.

## Overview

The AzureVirtualMachine component provides a streamlined way to deploy and manage Azure Virtual Machines with best-practice defaults. It supports both Linux and Windows VMs, with options for SSH key or password authentication, managed identities, data disks, and spot instances.

## Purpose

This component simplifies Azure VM deployment by:
- Providing sensible defaults for common configurations
- Supporting cross-references to other Azure resources (VNets, Key Vaults)
- Enabling both marketplace and custom images
- Integrating managed identities for secure Azure service access
- Supporting cost optimization through spot instances

## Key Features

- **Network Integration**: Reference subnets from AzureVpc resources using cross-references
- **Flexible Authentication**: SSH keys for Linux, passwords for Windows (or both)
- **Managed Identities**: System-assigned and user-assigned identity support
- **Disk Options**: Configure OS disk and attach additional data disks
- **Availability Zones**: Deploy to specific zones for high availability
- **Spot Instances**: Reduce costs with spot pricing for fault-tolerant workloads
- **Boot Diagnostics**: Debug boot issues with serial console access

## Example Usage

### Basic Linux VM with SSH

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureVirtualMachine
metadata:
  name: web-server
spec:
  region: eastus
  resource_group: my-resource-group
  vm_size: Standard_D2s_v3
  subnet_id:
    value_from:
      kind: AzureVpc
      name: my-vpc
      field_path: status.outputs.nodes_subnet_id
  image:
    publisher: Canonical
    offer: 0001-com-ubuntu-server-jammy
    sku: 22_04-lts-gen2
  ssh_public_key: "ssh-rsa AAAAB3NzaC1yc2E... user@host"
  enable_boot_diagnostics: true
```

### Windows VM with Password

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureVirtualMachine
metadata:
  name: windows-server
spec:
  region: westus2
  resource_group: my-resource-group
  vm_size: Standard_D4s_v5
  subnet_id:
    value: "/subscriptions/.../subnets/default"
  image:
    publisher: MicrosoftWindowsServer
    offer: WindowsServer
    sku: 2022-datacenter-g2
  admin_password:
    value_from:
      kind: AzureKeyVault
      name: my-vault
      field_path: status.outputs.vault_uri
  availability_zone: "1"
```

## Deploy

```bash
planton pulumi up --manifest vm.yaml
```

## Best Practices

- Use SSH keys for Linux VMs (more secure than passwords)
- Enable managed identities instead of storing credentials
- Deploy to availability zones for production workloads
- Use spot instances for dev/test and fault-tolerant workloads
- Configure network security groups to restrict access
