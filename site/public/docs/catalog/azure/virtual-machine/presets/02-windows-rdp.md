---
title: "Windows Server 2022 with RDP Access"
description: "This preset deploys a Windows Server 2022 Datacenter Gen2 VM with password authentication, a public IP for RDP access, boot diagnostics enabled, and a 128 GB Premium SSD OS disk. This configuration..."
type: "preset"
rank: "02"
presetSlug: "02-windows-rdp"
componentSlug: "virtual-machine"
componentTitle: "Virtual Machine"
provider: "azure"
icon: "package"
order: 2
---

# Windows Server 2022 with RDP Access

This preset deploys a Windows Server 2022 Datacenter Gen2 VM with password authentication, a public IP for RDP access, boot diagnostics enabled, and a 128 GB Premium SSD OS disk. This configuration is suitable for development, testing, and workloads that require Windows-specific software.

## When to Use

- Windows-based application servers (.NET, IIS, SQL Server workloads)
- Development or test environments requiring RDP access from the internet
- Teams running Windows-specific tooling that cannot run on Linux
- Quick-start Windows VMs for evaluation or prototyping

## Key Configuration Choices

- **Windows Server 2022 Datacenter Gen2** (`image.sku: 2022-datacenter-g2`) -- Latest LTS Windows Server with Gen2 VM performance
- **Password authentication** (`adminPassword`) -- Simpler setup for RDP access; consider switching to Azure AD login for production
- **Public IP enabled** (`network.enablePublicIp: true`) -- Allows direct RDP access; restrict with an NSG for production use
- **Standard_D2s_v3** (`vmSize`) -- 2 vCPUs, 8 GiB RAM; adequate for typical Windows workloads
- **128 GB Premium SSD** (`osDisk.sizeGb: 128`) -- Windows needs more disk space than Linux; Premium SSD for performance
- **Static public IP** (`network.publicIpAllocation: public_static`) -- IP address persists across VM reboots

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., `eastus`, `westeurope`) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<subnet-id>` | ARM resource ID of the subnet for the VM's NIC | Azure portal or `AzureVpc` / `AzureSubnet` status outputs |
| `<your-admin-password>` | Admin password (12+ chars, uppercase, lowercase, number, special char) | Generate a strong password; store in Key Vault for production |

## Related Presets

- **01-ubuntu-ssh** -- Use instead for Linux workloads with SSH key authentication and no public IP
